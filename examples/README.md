# Temporal MCP Examples

This directory contains examples and configuration for the Temporal Model Context Protocol (MCP) server that exposes Temporal workflows as tools for AI assistants.

## What is MCP?

The Model Context Protocol (MCP) is a protocol that allows AI models like Claude to interact with external tools and services. It provides a standardized way for AI models to access functionality outside of their training data.

## What is Temporal?

[Temporal](https://temporal.io/) is a workflow orchestration platform that simplifies the development of reliable applications. The Temporal MCP server allows Claude to execute and interact with Temporal workflows, enabling complex task automation, data processing, and service orchestration.

## Features

The Temporal MCP server provides access to workflows configured in `config.yml`, such as:

1. **Dynamic workflow execution** - Run any workflow defined in the configuration
2. **Cached results** - Optionally cache workflow results for improved performance
3. **Task queue management** - Configure specific or default task queues for workflow execution

## Using with Claude Desktop

### Automatic Configuration (Recommended)

The easiest way to configure Claude Desktop is to use the provided script:

1. Build the MCP server using the Makefile from the root directory:
   ```bash
   cd .. && make build
   ```

2. Run the configuration script from the examples directory:
   ```bash
   ./generate_claude_config.sh
   ```
   This will:
   - Generate a `claude_config.json` file with correct paths for your system
   - Add the file to .gitignore to prevent committing personal paths
   - Show instructions for deploying the config file

3. Copy the generated config to Claude's configuration directory:
   ```bash
   cp claude_config.json ~/Library/Application\ Support/Claude/claude_desktop_config.json
   ```

4. Restart Claude Desktop

### Manual Configuration

Alternatively, you can manually create a configuration file at `~/Library/Application Support/Claude/claude_desktop_config.json` with the following content:
   ```json
   {
     "mcpServers": {
       "temporal-mcp": {
         "command": "/full/path/to/your/bin/temporal-mcp",
         "args": ["--config", "/full/path/to/your/config.yml"],
         "env": {}
       }
     }
   }
   ```
   
   Remember to replace the paths with the actual full paths to your binaries and config file.

4. When chatting with Claude, you can ask it to use the Pig Latin conversion tools.

## Example Prompts for Claude



### Temporal MCP

Once connected to the Temporal MCP server, you can ask Claude things like:

- "Can you run the GreetingWorkflow with my name as a parameter?"
- "Please execute the DataProcessingWorkflow with the following parameters..."
- "Clear the cache for all workflows"
- "Run the AnalyticsWorkflow and show me the results"


## How It Works

The Temporal MCP server also uses the [mcp-golang](https://github.com/metoro-io/mcp-golang) library but connects to a Temporal service to execute workflows. When Claude needs to run a workflow:

1. It recognizes the need to execute a Temporal workflow
2. It calls the appropriate workflow tool with the required parameters
3. The MCP server executes the workflow on the Temporal service
4. The workflow result is returned to Claude
5. Claude presents the result to the user

The Temporal MCP server also supports result caching to improve performance for repetitive workflow executions.


