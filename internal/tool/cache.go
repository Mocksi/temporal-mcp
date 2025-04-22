package tool

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mocksi/temporal-mcp/internal/config"
)

// CacheClient handles workflow result caching operations
type CacheClient struct {
	db      *sql.DB
	enabled bool
	ttl     time.Duration
}

// NewCacheClient creates a new cache client instance
func NewCacheClient(cfg config.CacheConfig) (*CacheClient, error) {
	if !cfg.Enabled {
		return &CacheClient{
			enabled: false,
		}, nil
	}

	// Parse TTL
	ttl, err := time.ParseDuration(cfg.TTL)
	if err != nil {
		return nil, fmt.Errorf("invalid TTL format: %w", err)
	}

	// Use /tmp for cache path if the configured path is relative
	databasePath := cfg.DatabasePath
	if !filepath.IsAbs(databasePath) {
		// For relative paths, store in /tmp/temporal-mcp instead
		databasePath = filepath.Join("/tmp/temporal-mcp", filepath.Base(databasePath))
		log.Printf("Using temporary cache path: %s", databasePath)
	}

	// Ensure cache directory exists
	dbDir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_cache (
			workflow_name TEXT NOT NULL,
			params_hash TEXT NOT NULL,
			params TEXT NOT NULL,
			result TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY (workflow_name, params_hash)
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create cache table: %w", err)
	}

	return &CacheClient{
		db:      db,
		enabled: true,
		ttl:     ttl,
	}, nil
}

// Get retrieves a cached workflow result
func (c *CacheClient) Get(workflowName string, params map[string]string) (string, bool) {
	if !c.enabled {
		return "", false
	}

	// Serialize parameters for hashing
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return "", false
	}
	paramsHash := fmt.Sprintf("%x", paramsBytes)

	// Query cache
	row := c.db.QueryRow(
		"SELECT result, created_at FROM workflow_cache WHERE workflow_name = ? AND params_hash = ?",
		workflowName, paramsHash,
	)

	var result string
	var createdAt int64
	if err := row.Scan(&result, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}

	// Check if cache entry has expired
	if time.Since(time.Unix(createdAt, 0)) > c.ttl {
		// Delete expired entry
		c.db.Exec(
			"DELETE FROM workflow_cache WHERE workflow_name = ? AND params_hash = ?",
			workflowName, paramsHash,
		)
		return "", false
	}

	return result, true
}

// Set stores a workflow result in the cache
func (c *CacheClient) Set(workflowName string, params map[string]string, result string) error {
	if !c.enabled {
		return nil
	}

	// Serialize parameters for storage
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to serialize parameters: %w", err)
	}
	paramsHash := fmt.Sprintf("%x", paramsBytes)
	paramsString := string(paramsBytes)

	// Insert or replace cache entry
	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO workflow_cache (workflow_name, params_hash, params, result, created_at) VALUES (?, ?, ?, ?, ?)",
		workflowName, paramsHash, paramsString, result, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert cache entry: %w", err)
	}

	return nil
}

// Clear removes cache entries
func (c *CacheClient) Clear(workflowName string) (int64, error) {
	if !c.enabled {
		return 0, nil
	}

	var result sql.Result
	var err error

	if workflowName == "" {
		// Clear entire cache
		result, err = c.db.Exec("DELETE FROM workflow_cache")
	} else {
		// Clear cache for specific workflow
		result, err = c.db.Exec(
			"DELETE FROM workflow_cache WHERE workflow_name = ?",
			workflowName,
		)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to clear cache: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Close closes the database connection
func (c *CacheClient) Close() error {
	if c.enabled && c.db != nil {
		return c.db.Close()
	}
	return nil
}
