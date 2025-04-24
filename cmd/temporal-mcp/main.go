package main

import (
	"context"
	"flag"
	"fmt"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/mocksi/temporal-mcp/internal/config"
	"github.com/mocksi/temporal-mcp/internal/temporal"
	"github.com/mocksi/temporal-mcp/internal/tool"
	temporal_enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"hash/fnv"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
)

func main() {
	// Parse command line arguments
	configFile := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	// CRITICAL: Configure all loggers to write to stderr instead of stdout
	// This is essential as any output to stdout will corrupt the JSON-RPC stream
	log.SetOutput(os.Stderr)
	log.Println("Starting Temporal MCP server...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Loaded configuration with %d workflows", len(cfg.Workflows))

	// Initialize Temporal client
	var temporalClient client.Client
	var temporalError error

	temporalClient, temporalError = temporal.NewTemporalClient(cfg.Temporal)
	if temporalError != nil {
		log.Printf("WARNING: Failed to connect to Temporal service: %v", temporalError)
		log.Printf("MCP will run in degraded mode - workflow executions will return errors")
	} else {
		defer temporalClient.Close()
		log.Printf("Connected to Temporal service at %s", cfg.Temporal.HostPort)
	}

	// Initialize cache if enabled
	var cacheClient *tool.CacheClient
	if cfg.Cache.Enabled {
		cacheClient, err = tool.NewCacheClient(cfg.Cache)
		if err != nil {
			log.Fatalf("Failed to initialize cache: %v", err)
		}
		defer cacheClient.Close()
		log.Printf("Cache initialized with TTL of %s", cfg.Cache.TTL)
	} else {
		log.Println("Cache is disabled")
	}

	// Create a new MCP server with stdio transport for AI model communication
	server := mcp.NewServer(stdio.NewStdioServerTransport())

	// Create tool registry - used in future enhancements
	// registry := tool.NewRegistry(cfg, temporalClient, cacheClient)

	// Register all workflow tools
	log.Println("Registering workflow tools...")
	err = registerWorkflowTools(server, cfg, temporalClient, cacheClient)
	if err != nil {
		log.Fatalf("Failed to register workflow tools: %v", err)
	}

	// Register system prompt
	err = registerSystemPrompt(server, cfg)
	if err != nil {
		log.Fatalf("Failed to register system prompt: %v", err)
	}

	// Start the MCP server in a goroutine
	go func() {
		log.Printf("Temporal MCP server is running. Press Ctrl+C to stop.")
		if err := server.Serve(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down MCP server...", sig)
	log.Printf("Temporal MCP server has been stopped.")
}

// registerWorkflowTools registers all workflow definitions as MCP tools
func registerWorkflowTools(server *mcp.Server, cfg *config.Config, tempClient client.Client, cacheClient *tool.CacheClient) error {
	// Register all workflows as tools
	for name, workflow := range cfg.Workflows {
		err := registerWorkflowTool(server, name, workflow, tempClient, cacheClient, cfg.Cache.Enabled, cfg)
		if err != nil {
			return fmt.Errorf("failed to register workflow tool %s: %w", name, err)
		}
		log.Printf("Registered workflow tool: %s", name)
	}

	// Register cache clear tool if enabled
	if cfg.Cache.Enabled && cacheClient != nil {
		err := registerCacheClearTool(server, cacheClient)
		if err != nil {
			return fmt.Errorf("failed to register cache clear tool: %w", err)
		}
		log.Printf("Registered ClearCache tool")
	}

	return nil
}

// registerWorkflowTool registers a single workflow as an MCP tool
func registerWorkflowTool(server *mcp.Server, name string, workflow config.WorkflowDef, tempClient client.Client, cacheClient *tool.CacheClient, cacheEnabled bool, cfg *config.Config) error {
	// Define the type for workflow parameters based on fields
	type WorkflowParams struct {
		Params     map[string]string `json:"params"`
		ForceRerun bool              `json:"force_rerun"`
	}

	// Register the tool with MCP server
	return server.RegisterTool(name, workflow.Purpose, func(args WorkflowParams) (*mcp.ToolResponse, error) {
		// Check if result is in cache
		if cacheEnabled && cacheClient != nil {
			cachedResult, found := cacheClient.Get(name, args.Params)
			if found {
				log.Printf("Cache hit for workflow %s", name)
				return mcp.NewToolResponse(mcp.NewTextContent(cachedResult)), nil
			}
			log.Printf("Cache miss for workflow %s", name)
		}

		// Check if Temporal client is available
		if tempClient == nil {
			log.Printf("Error: Temporal client is not available for workflow: %s", name)
			return mcp.NewToolResponse(mcp.NewTextContent(
				"Error: Temporal service is currently unavailable. Please try again later.",
			)), nil
		}

		// Execute the workflow
		// Determine which task queue to use (workflow-specific or default)
		taskQueue := workflow.TaskQueue
		if taskQueue == "" && cfg != nil {
			taskQueue = cfg.Temporal.DefaultTaskQueue
			log.Printf("Using default task queue: %s for workflow %s", taskQueue, name)
		}

		workflowID, err := computeWorkflowID(workflow, args.Params)
		if err != nil {
			log.Printf("Error computing workflow ID from arguments: %v", err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error computing workflow ID from arguments: %v", err),
			)), nil
		}

		// This will execute a new workflow when:
		// - there is no workflow with the given id
		// - there is a failed workflow with the given id (e.g. terminated, failed, timed out)
		// and attach to an existing workflow when:
		// - there is a running workflow with the given id
		// - there is a successful workflow with the given id
		//
		// Note that temporal's data retention window (a setting on each namespace) influences the behavior above
		reusePolicy := temporal_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
		conflictPolicy := temporal_enums.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING

		if args.ForceRerun {
			// This will execute a new workflow in all cases. If there is a running workflow with the given id, it will
			// be terminated.
			reusePolicy = temporal_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
			conflictPolicy = temporal_enums.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING
		}

		wfOptions := client.StartWorkflowOptions{
			TaskQueue:                taskQueue,
			ID:                       workflowID,
			WorkflowIDReusePolicy:    reusePolicy,
			WorkflowIDConflictPolicy: conflictPolicy,
		}

		log.Printf("Starting workflow %s on task queue %s", name, taskQueue)

		// Start workflow execution
		run, err := tempClient.ExecuteWorkflow(context.Background(), wfOptions, name, args.Params)
		if err != nil {
			log.Printf("Error starting workflow %s: %v", name, err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error executing workflow: %v", err),
			)), nil
		}

		log.Printf("Workflow started: WorkflowID=%s RunID=%s", run.GetID(), run.GetRunID())

		// Wait for workflow completion
		var result string
		if err := run.Get(context.Background(), &result); err != nil {
			log.Printf("Error in workflow %s execution: %v", name, err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Workflow failed: %v", err),
			)), nil
		}

		log.Printf("Workflow %s completed successfully", name)

		// Cache the result if enabled
		if cacheEnabled && cacheClient != nil {
			if err := cacheClient.Set(name, args.Params, result); err != nil {
				log.Printf("Warning: Failed to cache result for workflow %s: %v", name, err)
			} else {
				log.Printf("Cached result for workflow %s", name)
			}
		}

		return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
	})
}

// hashWorkflowArgs produces a short (suitable for inclusion in workflow id) hash of the given arguments. Args must be
// printf-able (%+v).
func hashWorkflowArgs(allParams map[string]string, paramsToHash ...any) string {
	if len(paramsToHash) == 0 {
		log.Printf("Warning: No hash arguments provided - will hash all arguments. Please replace {{ hash }} with {{ hash . }} in the workflowIDRecipe")
		paramsToHash = []any{allParams}
	}

	hasher := fnv.New32()
	for _, arg := range paramsToHash {
		bytes := []byte(fmt.Sprintf("%+v", arg))
		_, _ = hasher.Write(bytes) // never fails, per docs
	}
	return fmt.Sprintf("%d", hasher.Sum32())
}

func computeWorkflowID(workflow config.WorkflowDef, params map[string]string) (string, error) {
	tmpl := template.New("id_recipe")

	tmpl.Funcs(template.FuncMap{
		"hash": func(paramsToHash ...any) string {
			return hashWorkflowArgs(params, paramsToHash...)
		},
	})
	if _, err := tmpl.Parse(workflow.WorkflowIDRecipe); err != nil {
		return "", err
	}

	writer := strings.Builder{}
	if err := tmpl.Execute(&writer, params); err != nil {
		return "", err
	}

	return writer.String(), nil
}

// registerCacheClearTool registers the tool to clear cache entries
func registerCacheClearTool(server *mcp.Server, cacheClient *tool.CacheClient) error {
	type ClearCacheParams struct {
		WorkflowName string `json:"workflowName,omitempty" jsonschema:"description=Optional. Name of the workflow to clear the cache for. If not provided, all cache entries will be cleared."`
	}

	return server.RegisterTool("ClearCache", "Clears cached workflow results, either by specific workflow or the entire cache.", func(args ClearCacheParams) (*mcp.ToolResponse, error) {
		if cacheClient == nil {
			return mcp.NewToolResponse(mcp.NewTextContent("Cache is not initialized")), nil
		}

		log.Printf("Clearing cache for workflow: %s", args.WorkflowName)
		rowsCleared, err := cacheClient.Clear(args.WorkflowName)
		if err != nil {
			return nil, fmt.Errorf("failed to clear cache: %w", err)
		}

		responseMsg := fmt.Sprintf("Successfully cleared %d cache entries", rowsCleared)
		if args.WorkflowName != "" {
			responseMsg = fmt.Sprintf("Successfully cleared %d cache entries for workflow '%s'", rowsCleared, args.WorkflowName)
		}

		log.Printf("%s", responseMsg)
		return mcp.NewToolResponse(mcp.NewTextContent(responseMsg)), nil
	})
}

// registerSystemPrompt registers the system prompt for the MCP
func registerSystemPrompt(server *mcp.Server, cfg *config.Config) error {
	return server.RegisterPrompt("system_prompt", "System prompt for the Temporal MCP", func(_ struct{}) (*mcp.PromptResponse, error) {
		// Build list of available tools from workflows
		workflowList := ""
		for name, workflow := range cfg.Workflows {
			workflowList += fmt.Sprintf("- %s: %s\n", name, workflow.Purpose)

			// Add parameter information
			workflowList += "  Parameters:\n"
			for _, field := range workflow.Input.Fields {
				for fieldName, description := range field {
					workflowList += fmt.Sprintf("    - %s: %s\n", fieldName, description)
				}
			}
			workflowList += "\n"
		}

		// Add ClearCache tool if enabled
		if cfg.Cache.Enabled {
			workflowList += "- ClearCache: Clears cached workflow results\n"
			workflowList += "  Parameters:\n"
			workflowList += "    - workflowName: (Optional) Name of the workflow to clear the cache for\n\n"
		}

		systemPrompt := fmt.Sprintf(`You are now connected to a Temporal MCP (Model Control Protocol) server that provides access to various Temporal workflows.

This MCP exposes the following workflow tools:

%s
Use these tools to help users interact with Temporal workflows. Each workflow requires a 'params' object containing the necessary parameters listed above.

Set force_rerun to true to force the workflow to run again. When force_rerun is false, temporal will deduplicate workflows
based on their arguments. Only set force_rerun to true if the user explicitly tells you to.

Example usage: 
Call GreetingWorkflow with {"params": {"name": "John"}}`, workflowList)

		return mcp.NewPromptResponse("system_prompt", mcp.NewPromptMessage(mcp.NewTextContent(systemPrompt), mcp.Role("system"))), nil
	})
}
