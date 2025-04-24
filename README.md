# ⏰🧠 Temporal MCP

> **Empowering AI with Workflow Orchestration**

Temporal MCP is a bridge connecting AI assistants like Claude with the powerful Temporal workflow engine. By implementing the Model Context Protocol (MCP), it allows AI assistants to discover, execute, and monitor complex workflow orchestrations—all through natural language conversations.

## ✨ What Can It Do?

Temporal MCP transforms how your AI assistants interact with your backend systems:

- **🔍 Automatic Discovery** — AI assistants can explore available workflows and their capabilities
- **🏃‍♂️ Seamless Execution** — Execute complex workflows with parameters through natural conversations
- **📊 Real-time Monitoring** — Check status of running workflows and get updates
- **⚡ Performance Optimization** — Smart caching of results for faster responses
- **🧠 AI-friendly Descriptions** — Rich metadata helps AI understand workflow purposes and operations

### Why Temporal MCP Exists

AI assistants are powerful for generating content and reasoning, but they lack the ability to execute complex workflows or maintain state across long-running operations. Temporal provides robust workflow orchestration with reliability features like retries, timeouts, and failover mechanisms. By connecting these systems:

- **AI assistants gain workflow superpowers** - Execute complex business processes, data pipelines, and service orchestrations
- **Temporal workflows become conversational** - Trigger and monitor workflows through natural language
- **Enterprise systems become AI-accessible** - Expose existing workflow infrastructure to AI assistants without rebuilding

## 🏁 Getting Started

### Prerequisites

Before you dive in, make sure you have:

- **Go 1.21+** — For building and running the MCP server
- **Temporal Server** — Running locally or remotely (see [Temporal docs](https://docs.temporal.io/docs/server/quick-install/))

### Installation

Let's get you up and running in just a few minutes:

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/temporal-mcp.git
cd temporal-mcp
```

2. **Build the project**
```bash
make build
```

3. **Configure your workflows**
Copy the sample config and customize it for your environment:
```bash
cp config.sample.yml config.yml
# Edit config.yml with your favorite editor
```

4. **Launch the server**
```bash
./bin/temporal-mcp --config ./config.yml
```

> 💡 **Tip:** For development, you can use `make run` to build and run in one step!

## Development

### Project Structure

```
./
├── cmd/            # Entry points for executables
├── internal/       # Internal package code
│   ├── api/        # MCP API implementation
│   ├── cache/      # Caching layer
│   ├── config/     # Configuration management
│   └── temporal/   # Temporal client integration
├── examples/       # Example configurations and scripts
└── docs/           # Documentation
```

### Common Commands

| Command | Description |
|---------|-------------|
| `make build` | Builds the binary in `./bin` |
| `make test` | Runs all unit tests |
| `make fmt` | Formats code according to Go standards |
| `make run` | Builds and runs the server |
| `make clean` | Removes build artifacts |

## 🔍 Troubleshooting

### Common Issues

**Connection Refused**
- ✓ Check if Temporal server is running
- ✓ Verify hostPort is correct in config.yml

**Workflow Not Found**
- ✓ Ensure workflow is registered in Temporal
- ✓ Check namespace settings in config.yml

**Claude Can't See Workflows**
- ✓ Verify claude_config.json is in the correct location
- ✓ Restart Claude after configuration changes

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤖 Using with AI Assistants

One of the most exciting features of Temporal MCP is how it bridges the gap between natural language and complex workflow orchestration using the [Model Context Protocol (MCP)](https://github.com/anthropics/anthropic-cookbook/tree/main/mcp).

### Setting Up Claude

Let's get Claude talking to your workflows in 5 easy steps:

1. **Build the server** (if you haven't already)
```bash
make build
```

2. **Define your workflows** in `config.yml`
Here's an example of a workflow definition that Claude can understand:

```yaml
workflows:
  data-processing-workflow:
    purpose: "Processes data from various sources with configurable parameters."
    input:
      type: "DataProcessingRequest"
      fields:
        - source_type: "The type of data source to process."
        - batch_size: "The number of records to process in each batch."
    output:
      type: "ProcessingResult"
      description: "Results of the data processing operation."
```

3. **Generate Claude's configuration**
```bash
cd examples
./generate_claude_config.sh
```

4. **Install the configuration**
```bash
cp examples/claude_config.json ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

5. **Start Claude** with this configuration

### Conversing with Your Workflows

Now for the magic part! You can talk to your workflows through Claude using natural language:

> 💬 "Claude, can you run the data processing workflow on the customer database with a batch size of 500?"

> 💬 "What workflows are available for me to use?"

> 💬 "Execute the billing workflow for customer ABC123 with the April invoice data"

> 💬 "Has the daily analytics workflow completed yet?"

Behind the scenes, Temporal MCP translates these natural language requests into properly formatted workflow executions—no more complex API calls or parameter formatting!

## ⚙️ Configuration

The heart of Temporal MCP is its configuration file, which connects your AI assistants to your workflow engine. Let's break it down:

### Configuration Architecture

Your `config.yml` consists of three key sections:

1. **🔌 Temporal Connection** — How to connect to your Temporal server
2. **💾 Cache Settings** — Performance optimization for workflow results
3. **🔧 Workflow Definitions** — The workflows your AI can discover and use

### Example Configuration

Here's a complete example with annotations to guide you:

```yaml
# Temporal server connection details
temporal:
  hostPort: "localhost:7233"       # Your Temporal server address
  namespace: "default"             # Temporal namespace
  environment: "local"             # "local" or "remote"
  defaultTaskQueue: "executions-task-queue"

  # Fine-tune connection behavior
  timeout: "5s"                    # Connection timeout
  retryOptions:                     # Robust retry settings
    initialInterval: "100ms"       # Start with quick retries
    maximumInterval: "10s"         # Max wait between retries
    maximumAttempts: 5              # Don't try forever
    backoffCoefficient: 2.0         # Exponential backoff

# Optimize performance with caching
cache:
  enabled: true                     # Turn caching on/off
  databasePath: "./workflow_cache.db"  # Where to store cache
  ttl: "24h"                       # Cache entries expire after 24h
  maxCacheSize: 104857600           # 100MB max cache size
  cleanupInterval: "1h"            # Automatic maintenance

# Define AI-discoverable workflows
workflows:
  data-analysis-workflow:           # Workflow ID in kebab-case
    purpose: "Analyzes a dataset and generates insights with comprehensive error handling and validation. Processes data through extraction, transformation, analysis stages and returns formatted results with visualizations."
    input:                          # What the workflow needs
      type: "AnalysisRequest"      # Input type name
      fields:                       # Required parameters
        - dataset_id: "ID of the dataset to analyze."
        - metrics: "List of metrics to compute."
        - format: "Output format (json, csv, etc)."
    output:                         # What the workflow returns
      type: "AnalysisResult"       # Output type name
      description: "Analysis results with computed metrics and visualizations."
    taskQueue: "analysis-queue"     # Optional custom task queue
```

> 💡 **Pro Tip:** Copy `config.sample.yml` as your starting point and customize from there.

## 💎 Best Practices

### Crafting Perfect Purpose Fields

The `purpose` field is your AI assistant's window into understanding what each workflow does. Make it count!

#### ✅ Do This
- Write clear, detailed descriptions of functionality
- Mention key parameters and how they customize behavior
- Describe expected outputs and their formats
- Note any limitations or constraints

#### ❌ Avoid This
- Vague descriptions ("Processes data")
- Technical jargon without explanation
- Missing important parameters
- Ignoring error cases or limitations

#### Before & After

**Before:** "Gets information about a file."

**After:** "Retrieves detailed metadata about a file or directory including size, creation time, last modified time, permissions, and type. Performs access validation to ensure the requested file is within allowed directories. Returns formatted JSON with all attributes or an appropriate error message."

### Naming Conventions

| Item | Convention | Example |
|------|------------|----------|
| Workflow IDs | kebab-case | `financial-report-generator` |
| Parameter names | snake_case | `account_id`, `start_date` |
| Parameters with units | Include unit | `timeout_seconds`, `batch_size` |

### Security Guidelines

⚠️ **Important Security Notes:**

- Keep credentials out of your configuration files
- Use environment variables for sensitive values
- Consider access controls for workflows with sensitive data
- Validate and sanitize all workflow inputs

> 💡 **Tip:** Create different configurations for development and production environments

### Why Good Purpose Fields Matter

1. **Enhanced AI Understanding** - Claude and other AI tools can provide much more accurate and helpful responses when they fully understand the capabilities and limitations of each component
2. **Fewer Errors** - Detailed descriptions reduce the chances of AI systems using components incorrectly
3. **Improved Debugging** - Clear descriptions help identify issues when workflows don't behave as expected
4. **Better Developer Experience** - New team members can understand your system more quickly
5. **Documentation As Code** - Purpose fields serve as living documentation that stays in sync with the codebase
