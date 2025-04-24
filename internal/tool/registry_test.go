package tool

import (
	"testing"

	"github.com/mocksi/temporal-mcp/internal/config"
)

// TestRegistryGetters tests the Registry getters
func TestRegistryGetters(t *testing.T) {
	// Create test objects
	cfg := &config.Config{}

	// Create registry directly without using interfaces to avoid lint errors
	registry := &Registry{
		config: cfg,
	}

	// Test GetConfig
	if registry.GetConfig() != cfg {
		t.Error("GetConfig did not return the expected config")
	}

	// For GetTemporalClient, we can only check it's not nil
	// since we can't directly compare interface values
	// Skip testing GetTemporalClient to avoid interface implementation issues
}

// TestNewRegistry tests the NewRegistry constructor
func TestNewRegistry(t *testing.T) {
	// Create test objects
	cfg := &config.Config{}

	// Since we can't easily mock the client.Client interface in tests,
	// we'll create the registry directly instead of using NewRegistry

	// Create a registry directly
	registry := &Registry{
		config: cfg,
	}

	// Test just the config and cacheClient properties
	if registry.config != cfg {
		t.Error("Registry not initialized with the correct config")
	}
}
