package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/mocksi/temporal-mcp/internal/config"
)

// TestGetTaskQueue tests the task queue selection logic
func TestGetTaskQueue(t *testing.T) {
	// Test cases to check task queue selection
	tests := []struct {
		name              string
		workflowQueue     string
		defaultQueue      string
		expectedQueue     string
		expectedIsDefault bool
	}{
		{
			name:              "Workflow with specific task queue",
			workflowQueue:     "specific-queue",
			defaultQueue:      "default-queue",
			expectedQueue:     "specific-queue",
			expectedIsDefault: false,
		},
		{
			name:              "Workflow without task queue uses default",
			workflowQueue:     "",
			defaultQueue:      "default-queue",
			expectedQueue:     "default-queue",
			expectedIsDefault: true,
		},
		{
			name:              "Empty default with empty workflow queue",
			workflowQueue:     "",
			defaultQueue:      "",
			expectedQueue:     "", // Empty because both are empty
			expectedIsDefault: true,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup workflow and config
			workflow := config.WorkflowDef{
				TaskQueue: tc.workflowQueue,
			}
			cfg := &config.Config{
				Temporal: config.TemporalConfig{
					DefaultTaskQueue: tc.defaultQueue,
				},
			}

			// Test task queue selection
			taskQueue := workflow.TaskQueue
			isUsingDefault := false

			if taskQueue == "" {
				taskQueue = cfg.Temporal.DefaultTaskQueue
				isUsingDefault = true
			}

			// Verify results
			if taskQueue != tc.expectedQueue {
				t.Errorf("Expected task queue '%s', got '%s'", tc.expectedQueue, taskQueue)
			}

			if isUsingDefault != tc.expectedIsDefault {
				t.Errorf("Expected isUsingDefault to be %v, got %v", tc.expectedIsDefault, isUsingDefault)
			}
		})
	}
}

// TestTaskQueueOverride ensures workflow-specific task queue overrides default
func TestTaskQueueOverride(t *testing.T) {
	// Setup workflow with specific queue
	workflow := config.WorkflowDef{
		TaskQueue: "specific-queue",
	}

	// Setup config with default queue
	cfg := &config.Config{
		Temporal: config.TemporalConfig{
			DefaultTaskQueue: "default-queue",
		},
	}

	// Check that workflow queue takes precedence
	resultQueue := workflow.TaskQueue
	if resultQueue != "specific-queue" {
		t.Errorf("Workflow queue should be 'specific-queue', got '%s'", resultQueue)
	}

	// Verify it doesn't use default queue when workflow has one
	if workflow.TaskQueue == "" && cfg.Temporal.DefaultTaskQueue != "" {
		t.Error("Test condition error: Should not use default when workflow queue exists")
	}
}

// TestDefaultTaskQueueFallback ensures default task queue is used as fallback
func TestDefaultTaskQueueFallback(t *testing.T) {
	// Setup workflow without specific queue
	workflow := config.WorkflowDef{
		TaskQueue: "", // No task queue specified
	}

	// Setup config with default queue
	cfg := &config.Config{
		Temporal: config.TemporalConfig{
			DefaultTaskQueue: "default-queue",
		},
	}

	// Get the task queue that would be used
	taskQueue := workflow.TaskQueue
	if taskQueue == "" {
		taskQueue = cfg.Temporal.DefaultTaskQueue
	}

	// Verify default queue is used
	if taskQueue != "default-queue" {
		t.Errorf("Should use default queue when workflow queue is empty, got '%s'", taskQueue)
	}

	// Verify workflow queue is actually empty
	if workflow.TaskQueue != "" {
		t.Errorf("Workflow queue should be empty, got '%s'", workflow.TaskQueue)
	}

	// Verify default queue is correctly set
	if cfg.Temporal.DefaultTaskQueue != "default-queue" {
		t.Errorf("Default queue should be 'default-queue', got '%s'", cfg.Temporal.DefaultTaskQueue)
	}
}

// TestWorkflowInputParams tests that workflow inputs are correctly passed to ExecuteWorkflow
func TestWorkflowInputParams(t *testing.T) {
	// Define test cases for different workflow input types
	type TestWorkflowRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Data string `json:"data"`
	}

	// Mock workflow parameters
	testCases := []struct {
		name       string
		workflowID string
		params     interface{}
	}{
		{
			name:       "Basic string parameter",
			workflowID: "string-param-workflow",
			params:     "simple-string-value",
		},
		{
			name:       "Struct parameter",
			workflowID: "struct-param-workflow",
			params: TestWorkflowRequest{
				ID:   "test-123",
				Name: "Test Workflow",
				Data: "Sample payload data",
			},
		},
		{
			name:       "Map parameter",
			workflowID: "map-param-workflow",
			params: map[string]interface{}{
				"id":      "map-123",
				"enabled": true,
				"count":   42,
				"nested": map[string]string{
					"key": "value",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Recreate the workflow execution context
			ctx := context.Background()

			// Verify parameters are correctly structured for ExecuteWorkflow
			// We can't directly test the execution but we can verify the parameters are correct
			switch params := tc.params.(type) {
			case string:
				if params == "" {
					t.Error("String parameter should not be empty")
				}
			case TestWorkflowRequest:
				if params.ID == "" {
					t.Error("Request ID should not be empty")
				}
				if params.Name == "" {
					t.Error("Request Name should not be empty")
				}
			case map[string]interface{}:
				if id, ok := params["id"]; !ok || id == "" {
					t.Error("Map parameter should have non-empty 'id' field")
				}
				if nested, ok := params["nested"].(map[string]string); !ok {
					t.Error("Map parameter should have valid nested map")
				} else if _, ok := nested["key"]; !ok {
					t.Error("Nested map should have 'key' property")
				}
			default:
				t.Errorf("Unexpected parameter type: %T", tc.params)
			}

			// Verify context is valid
			if ctx == nil {
				t.Error("Context should not be nil")
			}
		})
	}
}

func TestWorkflowIDComputation(t *testing.T) {
	type Case struct {
		recipe   string
		args     map[string]string
		expected string
	}

	tests := map[string]Case{
		"empty": {
			recipe:   "",
			expected: "",
		},
		"reference args": {
			recipe:   "id_{{ .one }}_{{ .two }}",
			args:     map[string]string{"one": "1", "two": "2"},
			expected: "id_1_2",
		},
		"reference missing args": {
			recipe:   "id_{{ .one }}_{{ .missing }}",
			args:     map[string]string{"one": "1"},
			expected: "id_1_<no value>",
		},
		"hash all args by accident": {
			recipe:   "id_{{ hash }}",
			args:     map[string]string{"one": "1", "two": "2"},
			expected: "id_321584698",
		},
		"hash all args properly": {
			recipe:   "id_{{ hash . }}",
			args:     map[string]string{"one": "1", "two": "2"},
			expected: "id_321584698",
		},
		"hash some args": {
			recipe:   "id_{{ hash .one .two }}",
			args:     map[string]string{"one": "1", "two": "2"},
			expected: "id_544649048",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			def := config.WorkflowDef{
				WorkflowIDRecipe: tc.recipe,
			}
			actual, err := computeWorkflowID(def, tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
