package temporal

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mocksi/temporal-mcp/internal/config"
	"go.temporal.io/sdk/client"
)

// NewTemporalClient creates a Temporal client based on the provided configuration
func NewTemporalClient(cfg config.TemporalConfig) (client.Client, error) {
	// Validate timeout format if specified
	if cfg.Timeout != "" {
		_, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		// Note: We're only validating the format, actual timeout handling would be implemented here
	}

	// Configure a logger that uses stderr
	tempLogger := log.New(os.Stderr, "[temporal] ", log.LstdFlags)

	// Create Temporal logger adapter that ensures all logs go to stderr
	temporalLogger := &StderrLogger{logger: tempLogger}

	// Set client options
	options := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    temporalLogger,
	}

	// Handle environment-specific configuration
	switch cfg.Environment {
	case "local":
		// Local Temporal server (default settings)
	case "remote":
		// To be implemented for remote/cloud Temporal connections
		// This would include TLS and authentication setup
		return nil, fmt.Errorf("remote environment configuration not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported environment type: %s", cfg.Environment)
	}

	// Create the client
	temporalClient, err := client.Dial(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return temporalClient, nil
}
