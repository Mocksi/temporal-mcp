// Package tool provides utilities for working with Temporal workflows as MCP tools
package tool

import (
	"github.com/mocksi/temporal-mcp/internal/config"
	"go.temporal.io/sdk/client"
)

// Registry manages workflow tools metadata and dependencies
type Registry struct {
	config     *config.Config
	tempClient client.Client
}

// NewRegistry creates a new tool registry with required dependencies
func NewRegistry(cfg *config.Config, tempClient client.Client) *Registry {
	return &Registry{
		config:     cfg,
		tempClient: tempClient,
	}
}

// GetConfig returns the configuration used by this registry
func (r *Registry) GetConfig() *config.Config {
	return r.config
}

// GetTemporalClient returns the Temporal client instance
func (r *Registry) GetTemporalClient() client.Client {
	return r.tempClient
}
