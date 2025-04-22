# Temporal MCP Technical Specification

## 1. Introduction

### Purpose and Goals
The primary purpose of this project is to create a Model Control Protocol (MCP) server implementation in Go that:
- Reads a YAML configuration defining Temporal workflows
- Automatically exposes each workflow as an MCP tool
- Handles tool invocation requests from MCP clients
- Executes workflows via Temporal and returns results

### Scope of Implementation
This specification covers the development of an MCP-compliant server that:
- Parses workflow definitions from YAML configuration
- Connects to a Temporal service using the Go SDK
- Dynamically generates MCP tool definitions based on the configured workflows
- Handles tool invocation lifecycle and result retrieval
- Implements caching with cache invalidation capabilities

### Definitions and Acronyms
- **MCP**: Model Control Protocol - A protocol for communication between AI models and external tools
- **Temporal**: A distributed, scalable workflow orchestration engine
- **Workflow**: A durable function that orchestrates activities in Temporal
- **YAML**: YAML Ain't Markup Language - A human-friendly data serialization standard

## 2. System Overview

### Client-Host-Server Architecture
```
+----------------+     +----------------+     +----------------+
|  MCP Client    |<--->|  MCP Server    |<--->|  Temporal      |
| (AI Assistant) |     | (Go Service)   |     |  Service       |
+----------------+     +----------------+     +----------------+
                          ^       ^
                          |       |
                          v       v
                    +----------------+     +----------------+
                    | YAML Workflow  |     | SQLite Cache  |
                    | Configuration  |     | Database      |
                    +----------------+     +----------------+
```

### Server Components
1. **Config Parser**: Loads and parses YAML workflow definitions
2. **MCP Protocol Handler**: Manages MCP message processing
3. **Temporal Client**: Interfaces with Temporal service
4. **Tool Registry**: Dynamically generates tool definitions from workflows
5. **Cache Manager**: Handles SQLite caching for workflow results

## 3. MCP Protocol Implementation

### 3.1 Protocol Definition
- Compliant with MCP Specification v0.1
- JSON-based message exchange
- Request/response communication pattern

### 3.2 Message Format

#### Tool Definitions Format
Tools are dynamically generated from workflow definitions in the YAML configuration, with an additional tool for cache management.

## 4. Server Implementation

### Core Components

#### YAML Configuration Schema
```go
type Config struct {
    Temporal  TemporalConfig           `yaml:"temporal"`
    Workflows map[string]WorkflowDef   `yaml:"workflows"`
    Cache     CacheConfig              `yaml:"cache"`
}

type TemporalConfig struct {
    // Connection configuration
    HostPort  string    `yaml:"hostPort"`
    Namespace string    `yaml:"namespace"`

    // Environment type
    Environment string  `yaml:"environment"` // "local" or "remote"

    // Authentication (for remote)
    Auth      *AuthConfig `yaml:"auth,omitempty"`

    // TLS configuration (for remote)
    TLS       *TLSConfig `yaml:"tls,omitempty"`

    // Connection options
    RetryOptions *RetryConfig `yaml:"retryOptions,omitempty"`
    Timeout      string       `yaml:"timeout,omitempty"`
}

type CacheConfig struct {
    Enabled        bool   `yaml:"enabled"`
    DatabasePath   string `yaml:"databasePath"`
    TTL            string `yaml:"ttl"`             // Time-to-live for cached results
    MaxCacheSize   int64  `yaml:"maxCacheSize"`    // Maximum size in bytes
    CleanupInterval string `yaml:"cleanupInterval"` // How often to clean expired entries
}
```

## 5. Cache Implementation

### Cache Clear Tool
The server implements a special tool for cache management:

```json
{
  "name": "ClearCache",
  "description": "Clears cached workflow results, either by specific workflow or the entire cache.",
  "parameters": {
    "type": "object",
    "properties": {
      "workflowName": {
        "type": "string",
        "description": "Optional. Name of the workflow to clear the cache for. If not provided, all cache entries will be cleared."
      }
    },
    "required": []
  }
}
```

### Cache Manager with Clear Function
```go
func (cm *CacheManager) Clear(workflowName string) (int64, error) {
    if !cm.enabled {
        return 0, nil
    }

    var result sql.Result
    var err error

    if workflowName == "" {
        // Clear entire cache
        result, err = cm.db.Exec("DELETE FROM workflow_cache")
    } else {
        // Clear cache for specific workflow
        result, err = cm.db.Exec(
            "DELETE FROM workflow_cache WHERE workflow_name = ?",
            workflowName,
        )
    }

    if err != nil {
        return 0, fmt.Errorf("failed to clear cache: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("failed to get rows affected: %w", err)
    }

    return rowsAffected, nil
}
```

### Clear Cache Tool Handler
```go
func (s *MCPServer) handleClearCache(params map[string]interface{}) (ToolCallResponse, error) {
    // Extract workflow name if provided
    var workflowName string
    if name, ok := params["workflowName"].(string); ok {
        workflowName = name
    }

    // Clear cache
    rowsAffected, err := s.cacheManager.Clear(workflowName)
    if err != nil {
        return ToolCallResponse{}, fmt.Errorf("failed to clear cache: %w", err)
    }

    // Create response
    response := ToolCallResponse{
        ToolName: "ClearCache",
        Status:   "completed",
        Result: map[string]interface{}{
            "success":      true,
            "entriesCleared": rowsAffected,
            "workflow":     workflowName,
        },
    }

    return response, nil
}
```

### Tool Registration Integration
```go
func (s *MCPServer) generateToolDefinitions() error {
    // Generate tools for workflows
    for name, workflow := range s.config.Workflows {
        // Create tool definition for workflow
        // ...existing code...
    }

    // Add special tool for clearing cache
    s.tools["ClearCache"] = ToolDefinition{
        Name:        "ClearCache",
        Description: "Clears cached workflow results, either by specific workflow or the entire cache.",
        Parameters: JSONSchema{
            Type: "object",
            Properties: map[string]JSONSchemaProperty{
                "workflowName": {
                    Type:        "string",
                    Description: "Optional. Name of the workflow to clear the cache for. If not provided, all cache entries will be cleared.",
                },
            },
            Required: []string{},
        },
        Internal: true,
    }

    return nil
}
```

### Enhanced Tool Invocation Router
```go
func (s *MCPServer) handleToolCall(call ToolCallRequest) (ToolCallResponse, error) {
    toolName := call.Name
    tool, exists := s.tools[toolName]
    if !exists {
        return ToolCallResponse{}, fmt.Errorf("tool %s not found", toolName)
    }

    // Handle special internal tools
    if tool.Internal {
        switch toolName {
        case "ClearCache":
            return s.handleClearCache(call.Parameters)
        default:
            return ToolCallResponse{}, fmt.Errorf("unknown internal tool: %s", toolName)
        }
    }

    // Regular workflow tool handling
    // ...existing workflow execution code...
}
```

## 6. Configuration Example

```yaml
temporal:
  # Connection configuration
  hostPort: "localhost:7233"  # Local Temporal server
  namespace: "default"
  environment: "local"        # "local" or "remote"

  # Connection options
  timeout: "5s"
  retryOptions:
    initialInterval: "100ms"
    maximumInterval: "10s"
    maximumAttempts: 5
    backoffCoefficient: 2.0

  # For remote Temporal server (Temporal Cloud)
  # environment: "remote"
  # hostPort: "your-namespace.tmprl.cloud:7233"
  # namespace: "your-namespace"
  # auth:
  #   clientID: "your-client-id"
  #   clientSecret: "your-client-secret"
  #   audience: "your-audience"
  #   oauth2URL: "https://auth.temporal.io/oauth2/token"
  # tls:
  #   certPath: "/path/to/client.pem"
  #   keyPath: "/path/to/client.key"
  #   caPath: "/path/to/ca.pem"
  #   serverName: "*.tmprl.cloud"
  #   insecureSkipVerify: false

# Cache configuration
cache:
  enabled: true
  databasePath: "./workflow_cache.db"
  ttl: "24h"                # Cache entries expire after 24 hours
  maxCacheSize: 104857600   # 100MB max cache size
  cleanupInterval: "1h"     # Run cleanup every hour

workflows:
  IngestWorkflow:
    purpose: "Ingests documents into the vector store."
    input:
      type: "IngestRequest"
      fields:
        - doc_id: "The document ID to ingest."
    output:
      type: "string"
      description: "ID of the ingested document."
    taskQueue: "ingest-queue"

  UpdateXHRWorkflow:
    purpose: "Updates XHR requests for DOM elements."
    input:
      type: "RAGRequest"
      fields:
        - session_id: "The session ID associated with the request."
        - action: "The action to perform on the DOM element."
    output:
      type: "RAGResponse"
      description: "The result of processing the fetched data."
    taskQueue: "xhr-queue"

  ChatWorkflow:
    purpose: "Processes chat completion requests."
    input:
      type: "ChatRequest"
      fields:
        - id: "The ID of the chat request."
        - prompt_id: "The prompt ID for the chat."
    output:
      type: "string"
      description: "The string completion response."
    taskQueue: "chat-queue"

  NewChatMessageWorkflow:
    purpose: "Processes new chat messages and updates JSON."
    input:
      type: "Message"
      fields:
        - client_id: "The client ID associated with the message."
        - timestamp: "The timestamp of the message."
        - content: "The content of the message."
        - updated_json: "The updated JSON content."
    output:
      type: "Message"
      description: "The message with explanation and updated JSON."
    taskQueue: "message-queue"

  ProxieJSONWorkflow:
    purpose: "Processes JSON payloads from Proxie."
    input:
      type: "ProxieJSONRequest"
      fields:
        - client_id: "The client ID associated with the request."
        - content: "The content of the request."
        - updated_json: "The updated JSON content."
        - request_hash: "The request hash for tracking."
    output:
      type: "string"
      description: "JSON string response for Proxie."
    taskQueue: "json-queue"
```

## 7. Security Considerations

### Data Protection
- No persistent storage of sensitive workflow data in MCP server
- TLS for Temporal Cloud connections
- Secure parameter handling

### Validation
- Input validation against schema before workflow execution
- Configuration validation at startup
- Response validation before returning to client

## 8. Performance Requirements

### Scalability
- Support for multiple concurrent tool invocations
- Efficient type conversion and serialization
- Minimal memory footprint

### Latency
- Tool discovery response < 100ms
- Tool invocation initialization < 200ms
- Cache hits < 10ms

## 9. Testing Strategy

### Unit Testing
- Configuration parsing
- Tool definition generation
- Cache operations
- MCP message handling

### Integration Testing
- End-to-end workflow execution
- Cache hit/miss scenarios
- Cache clearing functionality

## 10. Future Enhancements

### Roadmap
1. **VERSION_0 (Initial Release)**:
   - Basic YAML configuration parsing
   - Dynamic tool definition generation
   - Temporal workflow integration
   - SQLite caching with clear functionality
   - Stdio transport for MCP

2. **VERSION_1**:
   - Enhanced type system with automatic struct generation
   - HTTP/SSE transport support
   - Advanced cache analytics
   - Improved error handling and reporting

3. **VERSION_2**:
   - Advanced workflow querying capabilities
   - Metrics and monitoring integration
   - Hot reloading of configuration
   - Cloud-native deployment options

## Conclusion

This specification provides a comprehensive blueprint for developing a Golang MCP server that dynamically exposes Temporal workflows as tools. The addition of the cache clear functionality provides important operational capabilities for managing the cache system, allowing for targeted clearing of specific workflow results or complete cache resets when necessary.

The design leverages Go's strong typing system and reflection capabilities while providing a clean, standardized interface for AI assistants to discover and invoke Temporal workflows through the MCP protocol.