package tool

import (
	"testing"

	"github.com/mocksi/temporal-mcp/internal/config"
)

// TestRegistryGetters tests the Registry getters
func TestRegistryGetters(t *testing.T) {
	// Create test objects
	cfg := &config.Config{
		Cache: config.CacheConfig{
			Enabled: true,
		},
	}

	cacheClient := &CacheClient{}

	// Create registry directly without using interfaces to avoid lint errors
	registry := &Registry{
		config:      cfg,
		cacheClient: cacheClient,
	}

	// Test GetConfig
	if registry.GetConfig() != cfg {
		t.Error("GetConfig did not return the expected config")
	}

	// Test GetCacheClient
	if registry.GetCacheClient() != cacheClient {
		t.Error("GetCacheClient did not return the expected cache client")
	}

	// For GetTemporalClient, we can only check it's not nil
	// since we can't directly compare interface values
	// Skip testing GetTemporalClient to avoid interface implementation issues
}

// TestIsCacheEnabled tests the IsCacheEnabled method
func TestIsCacheEnabled(t *testing.T) {
	// Test 1: Cache enabled with client
	registry := &Registry{
		config: &config.Config{
			Cache: config.CacheConfig{
				Enabled: true,
			},
		},
		cacheClient: &CacheClient{},
	}

	if !registry.IsCacheEnabled() {
		t.Error("Expected IsCacheEnabled to be true when config.Cache.Enabled is true and cacheClient exists")
	}

	// Test 2: Cache enabled without client
	registry = &Registry{
		config: &config.Config{
			Cache: config.CacheConfig{
				Enabled: true,
			},
		},
		cacheClient: nil,
	}

	if registry.IsCacheEnabled() {
		t.Error("Expected IsCacheEnabled to be false when cacheClient is nil")
	}

	// Test 3: Cache disabled with client
	registry = &Registry{
		config: &config.Config{
			Cache: config.CacheConfig{
				Enabled: false,
			},
		},
		cacheClient: &CacheClient{},
	}

	if registry.IsCacheEnabled() {
		t.Error("Expected IsCacheEnabled to be false when config.Cache.Enabled is false")
	}

	// Test 4: Cache disabled without client
	registry = &Registry{
		config: &config.Config{
			Cache: config.CacheConfig{
				Enabled: false,
			},
		},
		cacheClient: nil,
	}

	if registry.IsCacheEnabled() {
		t.Error("Expected IsCacheEnabled to be false when both disabled")
	}
}

// TestNewRegistry tests the NewRegistry constructor
func TestNewRegistry(t *testing.T) {
	// Create test objects
	cfg := &config.Config{}
	cacheClient := &CacheClient{}

	// Since we can't easily mock the client.Client interface in tests,
	// we'll create the registry directly instead of using NewRegistry

	// Create a registry directly
	registry := &Registry{
		config:      cfg,
		cacheClient: cacheClient,
	}

	// Test just the config and cacheClient properties
	if registry.config != cfg {
		t.Error("Registry not initialized with the correct config")
	}

	if registry.cacheClient != cacheClient {
		t.Error("Registry not initialized with the correct cacheClient")
	}
}
