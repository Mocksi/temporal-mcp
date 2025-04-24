package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/mocksi/temporal-mcp/internal/config"
	"github.com/mocksi/temporal-mcp/internal/temporal"
	temporal_enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
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

	// Create a new MCP server with stdio transport for AI model communication
	server := mcp.NewServer(stdio.NewStdioServerTransport())

	// Create tool registry - used in future enhancements
	// registry := tool.NewRegistry(cfg, temporalClient, cacheClient)

	// Register all workflow tools
	log.Println("Registering workflow tools...")
	err = registerWorkflowTools(server, cfg, temporalClient)
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
func registerWorkflowTools(server *mcp.Server, cfg *config.Config, tempClient client.Client) error {
	// Register all workflows as tools
	for name, workflow := range cfg.Workflows {
		err := registerWorkflowTool(server, name, workflow, tempClient, cfg)
		if err != nil {
			return fmt.Errorf("failed to register workflow tool %s: %w", name, err)
		}
		log.Printf("Registered workflow tool: %s", name)
	}

	return nil
}

// registerWorkflowTool registers a single workflow as an MCP tool
func registerWorkflowTool(server *mcp.Server, name string, workflow config.WorkflowDef, tempClient client.Client, cfg *config.Config) error {
	// Define the type for workflow parameters based on fields
	type WorkflowParams struct {
		Params     map[string]string `json:"params"`
		ForceRerun bool              `json:"force_rerun"`
	}

	// Register the tool with MCP server
	return server.RegisterTool(name, workflow.Purpose, func(args WorkflowParams) (*mcp.ToolResponse, error) {
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

		if workflowID == "" {
			log.Printf("Workflow %q has an empty or missing workflowIDRecipe - using a random workflow id", name)
			workflowID = uuid.NewString()
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

		return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
	})
}

func computeWorkflowID(workflow config.WorkflowDef, params map[string]string) (string, error) {
	tmpl := template.New("id_recipe")

	tmpl.Funcs(template.FuncMap{
		"hash": func(paramsToHash ...any) (string, error) {
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
