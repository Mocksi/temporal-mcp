package tool

// Definition represents an MCP tool definition
type Definition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  Schema `json:"parameters"`
	Internal    bool   `json:"-"` // Flag for internal tools like ClearCache
}

// Schema represents a JSON Schema for tool parameters
type Schema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required"`
}

// SchemaProperty represents a property in a JSON Schema
type SchemaProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
