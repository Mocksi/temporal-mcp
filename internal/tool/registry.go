// Package tool provides utilities for working with Temporal workflows as MCP tools
package tool

import (
	"github.com/mocksi/temporal-mcp/internal/config"
	"go.temporal.io/sdk/client"
)

// Registry manages workflow tools metadata and dependencies
type Registry struct {
	config      *config.Config
	tempClient  client.Client
	cacheClient *CacheClient
}

// NewRegistry creates a new tool registry with required dependencies
func NewRegistry(cfg *config.Config, tempClient client.Client, cacheClient *CacheClient) *Registry {
	return &Registry{
		config:      cfg,
		tempClient:  tempClient,
		cacheClient: cacheClient,
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

// GetCacheClient returns the cache client instance
func (r *Registry) GetCacheClient() *CacheClient {
	return r.cacheClient
}

// IsCacheEnabled returns whether caching is enabled
func (r *Registry) IsCacheEnabled() bool {
	return r.config.Cache.Enabled && r.cacheClient != nil
}
