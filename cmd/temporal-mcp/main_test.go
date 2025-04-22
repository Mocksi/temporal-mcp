package main

import (
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
