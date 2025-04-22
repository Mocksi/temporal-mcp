package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config holds the top-level configuration
type Config struct {
	Temporal  TemporalConfig         `yaml:"temporal"`
	Workflows map[string]WorkflowDef `yaml:"workflows"`
	Cache     CacheConfig            `yaml:"cache"`
}

// TemporalConfig defines connection settings for Temporal service
type TemporalConfig struct {
	HostPort         string `yaml:"hostPort"`
	Namespace        string `yaml:"namespace"`
	Environment      string `yaml:"environment"`
	Timeout          string `yaml:"timeout,omitempty"`
	DefaultTaskQueue string `yaml:"defaultTaskQueue,omitempty"`
}

// CacheConfig defines SQLite cache settings
type CacheConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DatabasePath    string `yaml:"databasePath"`
	TTL             string `yaml:"ttl"`
	MaxCacheSize    int64  `yaml:"maxCacheSize"`
	CleanupInterval string `yaml:"cleanupInterval"`
}

// WorkflowDef describes a Temporal workflow exposed as a tool
type WorkflowDef struct {
	Purpose   string       `yaml:"purpose"`
	Input     ParameterDef `yaml:"input"`
	Output    ParameterDef `yaml:"output"`
	TaskQueue string       `yaml:"taskQueue"`
}

// ParameterDef defines input/output schema for a workflow
type ParameterDef struct {
	Type        string              `yaml:"type"`
	Fields      []map[string]string `yaml:"fields"`
	Description string              `yaml:"description,omitempty"`
}

// LoadConfig reads and parses YAML config from file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
