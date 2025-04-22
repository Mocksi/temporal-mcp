# Temporal MCP

A Golang repository for Temporal Mission Control Panel (MCP).

## Overview

Temporal MCP is a management and control panel for [Temporal](https://temporal.io/) workflows. It provides enhanced monitoring, management, and control capabilities for Temporal workflow executions.

## Features

- Workflow monitoring and visualization
- Advanced workflow control operations
- Customizable dashboards
- Metrics and analytics
- User management and access control

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Temporal server running locally or accessible endpoint

### Installation

Clone the repository:

```bash
git clone https://github.com/yourusername/temporal-mcp.git
cd temporal-mcp
```

Install dependencies:

```bash
go mod tidy
```

### Running the Application

```bash
go run cmd/main.go
```

## Development

### Project Structure

```
temporal-mcp/
├── cmd/                # Command-line applications
├── internal/           # Private application code
│   ├── api/            # API handlers
│   ├── config/         # Configuration
│   ├── models/         # Data models
│   └── service/        # Business logic
├── pkg/                # Public libraries
├── scripts/            # Build and deployment scripts
└── test/               # Test files
```

### Building

```bash
go build -o bin/temporal-mcp cmd/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Using with Claude

To use the Pig Latin MCP with Claude:

1. Build the piglatin-mcp binary:
```bash
make piglatin-mcp
```

2. Configure Claude to use the MCP server by setting up `examples/claude_config.json`:
```json
{
  "mcpServers": {
    "piglatin-mcp": {
      "command": "/Users/YOUR_USERNAME/Code/mocksi/temporal-mcp/bin/piglatin-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

Replace `YOUR_USERNAME` with your actual username in the path.

3. Start Claude with this configuration.

## Configuration Best Practices

### Writing Effective Purpose Fields

The `purpose` field in configuration files is **critically important** for Claude and other AI assistants to understand and correctly interact with workflows and activities. A well-written purpose field should:

1. **Be comprehensive and detailed** - Provide a thorough explanation of what the workflow/activity does, not just a brief label
2. **Include implementation details** - Mention how the component handles errors, retries, validation, etc.
3. **Explain context and relationships** - Describe how the component fits into the larger system
4. **Use precise technical language** - Be specific about data transformations, storage mechanisms, etc.

For example, instead of:
```yaml
get_file_info:
  purpose: "Gets information about a file."
```

Use:
```yaml
get_file_info:
  purpose: "Retrieve detailed metadata about a file or directory. Returns comprehensive information including size, creation time, last modified time, permissions, and type. This tool is perfect for understanding file characteristics without reading the actual content. Only works within allowed directories."
```

### Why Good Purpose Fields Matter

1. **Enhanced AI Understanding** - Claude and other AI tools can provide much more accurate and helpful responses when they fully understand the capabilities and limitations of each component
2. **Fewer Errors** - Detailed descriptions reduce the chances of AI systems using components incorrectly
3. **Improved Debugging** - Clear descriptions help identify issues when workflows don't behave as expected
4. **Better Developer Experience** - New team members can understand your system more quickly
5. **Documentation As Code** - Purpose fields serve as living documentation that stays in sync with the codebase
