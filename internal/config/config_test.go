package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configPath := filepath.Join(t.TempDir(), "test_config.yml")

	// Sample YAML content matching our struct definitions
	configContent := `
temporal:
  hostPort: "localhost:7233"
  namespace: "default"
  environment: "local"

cache:
  enabled: true
  databasePath: "./test_cache.db"
  ttl: "1h"
  maxCacheSize: 10485760
  cleanupInterval: "10m"

workflows:
  TestWorkflow:
    purpose: "Test workflow"
    input:
      type: "TestRequest"
      fields:
        - id: "The test ID"
    output:
      type: "string"
      description: "Test result"
    taskQueue: "test-queue"
`
	// Write the test config
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate the loaded config
	if cfg.Temporal.HostPort != "localhost:7233" {
		t.Errorf("Expected HostPort to be localhost:7233, got %s", cfg.Temporal.HostPort)
	}

	if cfg.Temporal.Namespace != "default" {
		t.Errorf("Expected Namespace to be default, got %s", cfg.Temporal.Namespace)
	}

	if !cfg.Cache.Enabled {
		t.Error("Expected Cache.Enabled to be true")
	}

	workflow, exists := cfg.Workflows["TestWorkflow"]
	if !exists {
		t.Fatal("TestWorkflow not found in config")
	}

	if workflow.Purpose != "Test workflow" {
		t.Errorf("Expected workflow purpose to be 'Test workflow', got '%s'", workflow.Purpose)
	}

	if workflow.TaskQueue != "test-queue" {
		t.Errorf("Expected task queue to be 'test-queue', got '%s'", workflow.TaskQueue)
	}

	if len(workflow.Input.Fields) != 1 {
		t.Fatalf("Expected 1 input field, got %d", len(workflow.Input.Fields))
	}

	if _, ok := workflow.Input.Fields[0]["id"]; !ok {
		t.Error("Expected input field 'id' not found")
	}
}
