package temporal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mocksi/temporal-mcp/internal/config"
	"go.temporal.io/sdk/client"
)

// TestNewTemporalClient tests the client creation with different configurations
func TestNewTemporalClient(t *testing.T) {
	// Test valid local configuration
	t.Run("ValidLocalConfig", func(t *testing.T) {
		// Use a non-standard port to ensure we won't accidentally connect to a real server
		cfg := config.TemporalConfig{
			HostPort:    "localhost:12345", // Use a port that's unlikely to have a Temporal server
			Namespace:   "default",
			Environment: "local",
			Timeout:     "5s",
		}

		// Attempt to create client - we expect a connection error, not a config error
		client, err := NewTemporalClient(cfg)

		// Check that either:
		// 1. We got a connection error (most likely case)
		// 2. Or somehow we got a valid client (unlikely, but possible if a test server is running)
		if err != nil {
			// Verify this is a connection error, not a config validation error
			if !strings.Contains(err.Error(), "failed to create Temporal client") {
				t.Errorf("Expected connection error, got: %v", err)
			}
		} else {
			// If we got a client, make sure to close it
			defer client.Close()
		}
	})

	// Test invalid environment
	t.Run("InvalidEnvironment", func(t *testing.T) {
		cfg := config.TemporalConfig{
			HostPort:    "localhost:7233",
			Namespace:   "default",
			Environment: "invalid",
			Timeout:     "5s",
		}

		_, err := NewTemporalClient(cfg)
		if err == nil {
			t.Error("Expected error for invalid environment, got nil")
		}
	})

	// Test invalid timeout
	t.Run("InvalidTimeout", func(t *testing.T) {
		cfg := config.TemporalConfig{
			HostPort:    "localhost:7233",
			Namespace:   "default",
			Environment: "local",
			Timeout:     "invalid",
		}

		_, err := NewTemporalClient(cfg)
		if err == nil {
			t.Error("Expected error for invalid timeout, got nil")
		}
	})

	// Test remote environment (which is not implemented yet)
	t.Run("RemoteEnvironment", func(t *testing.T) {
		cfg := config.TemporalConfig{
			HostPort:    "test.tmprl.cloud:7233",
			Namespace:   "test-namespace",
			Environment: "remote",
		}

		_, err := NewTemporalClient(cfg)
		if err == nil {
			t.Error("Expected error for unimplemented remote environment, got nil")
		}
	})
}

// MockWorkflowClient is a mock implementation of the Temporal client for testing
type MockWorkflowClient struct {
	lastWorkflowName string
	lastParams       interface{}
	lastOptions      client.StartWorkflowOptions
}

// ExecuteWorkflow mocks the ExecuteWorkflow method for testing
func (m *MockWorkflowClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	m.lastWorkflowName = workflow.(string)
	m.lastOptions = options
	if len(args) > 0 {
		m.lastParams = args[0]
	}
	// Return a mock workflow run
	return &MockWorkflowRun{}, nil
}

// Close is a no-op for the mock client
func (m *MockWorkflowClient) Close() {}

// MockWorkflowRun is a mock implementation of WorkflowRun for testing
type MockWorkflowRun struct{}

// GetID returns a mock workflow ID
func (m *MockWorkflowRun) GetID() string {
	return "mock-workflow-id"
}

// GetRunID returns a mock run ID
func (m *MockWorkflowRun) GetRunID() string {
	return "mock-run-id"
}

// Get is a mock implementation that returns no error
func (m *MockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}

// GetWithOptions is a mock implementation of the WorkflowRun interface method
func (m *MockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, opts client.WorkflowRunGetOptions) error {
	return nil
}

// TestWorkflowExecution tests workflow execution with different types of input parameters
func TestWorkflowExecution(t *testing.T) {
	// Define test structs
	type TestRequest struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}

	type ComplexRequest struct {
		ClientID  string                 `json:"client_id"`
		Command   string                 `json:"command"`
		Data      map[string]interface{} `json:"data"`
		Timestamp time.Time              `json:"timestamp"`
	}

	// Test cases with different input types
	testCases := []struct {
		name           string
		workflowName   string
		taskQueue      string
		params         interface{}
		expectedParams interface{}
	}{
		{
			name:           "String Parameter",
			workflowName:   "string-workflow",
			taskQueue:      "default-queue",
			params:         "simple-string-input",
			expectedParams: "simple-string-input",
		},
		{
			name:         "Struct Parameter",
			workflowName: "struct-workflow",
			taskQueue:    "test-queue",
			params: TestRequest{
				ID:    "req-123",
				Value: "test-value",
			},
			expectedParams: TestRequest{
				ID:    "req-123",
				Value: "test-value",
			},
		},
		{
			name:         "Complex Parameter",
			workflowName: "complex-workflow",
			taskQueue:    "complex-queue",
			params: ComplexRequest{
				ClientID:  "client-456",
				Command:   "analyze",
				Data:      map[string]interface{}{"key": "value"},
				Timestamp: time.Now(),
			},
			expectedParams: ComplexRequest{
				ClientID: "client-456",
				Command:  "analyze",
				Data:     map[string]interface{}{"key": "value"},
				// Time will be different but type should match
			},
		},
		{
			name:         "Map Parameter",
			workflowName: "map-workflow",
			taskQueue:    "map-queue",
			params: map[string]interface{}{
				"id":     "map-789",
				"count":  42,
				"active": true,
			},
			expectedParams: map[string]interface{}{
				"id":     "map-789",
				"count":  42,
				"active": true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock client
			mockClient := &MockWorkflowClient{}

			// Execute the workflow with the test parameters
			ctx := context.Background()
			options := client.StartWorkflowOptions{
				ID:        "test-" + tc.workflowName,
				TaskQueue: tc.taskQueue,
			}

			// Call ExecuteWorkflow on the mock client
			_, err := mockClient.ExecuteWorkflow(ctx, options, tc.workflowName, tc.params)
			if err != nil {
				t.Fatalf("ExecuteWorkflow failed: %v", err)
			}

			// Verify workflow name
			if mockClient.lastWorkflowName != tc.workflowName {
				t.Errorf("Expected workflow name %s, got %s", tc.workflowName, mockClient.lastWorkflowName)
			}

			// Verify task queue
			if mockClient.lastOptions.TaskQueue != tc.taskQueue {
				t.Errorf("Expected task queue %s, got %s", tc.taskQueue, mockClient.lastOptions.TaskQueue)
			}

			// Verify parameters were passed correctly
			switch params := mockClient.lastParams.(type) {
			case string:
				expectedStr, ok := tc.expectedParams.(string)
				if !ok || params != expectedStr {
					t.Errorf("Expected string param %v, got %v", tc.expectedParams, params)
				}
			case TestRequest:
				expected, ok := tc.expectedParams.(TestRequest)
				if !ok || params.ID != expected.ID || params.Value != expected.Value {
					t.Errorf("Expected struct param %v, got %v", tc.expectedParams, params)
				}
			case ComplexRequest:
				expected, ok := tc.expectedParams.(ComplexRequest)
				if !ok || params.ClientID != expected.ClientID || params.Command != expected.Command {
					t.Errorf("Expected complex param %v, got %v", tc.expectedParams, params)
				}
			case map[string]interface{}:
				expected, ok := tc.expectedParams.(map[string]interface{})
				if !ok {
					t.Errorf("Expected map param %v, got %v", tc.expectedParams, params)
				}
				// Check key values
				for k, v := range expected {
					if params[k] != v {
						t.Errorf("Expected map[%s]=%v, got %v", k, v, params[k])
					}
				}
			default:
				t.Errorf("Unexpected parameter type: %T", params)
			}
		})
	}
}
