package temporal

import (
	"strings"
	"testing"

	"github.com/mocksi/temporal-mcp/internal/config"
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
