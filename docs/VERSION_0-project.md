# VERSION_0 Project Tasks

This document breaks down the VERSION_0 specification into small, measurable, actionable tasks.

## 1. Project Setup
- [x] Add dependencies in `go.mod`:
  - `go.temporal.io/sdk`
  - `gopkg.in/yaml.v3`
  - `github.com/mattn/go-sqlite3`
- [x] Install dependencies via Make: `make install`
- [x] Build the application: `make build`

## 2. Config Parser
- [x] Define Go struct types in `config.go`:
  - `Config`, `TemporalConfig`, `CacheConfig`, `WorkflowDef`
- [x] Implement `func LoadConfig(path string) (*Config, error)` in `config.go`
- [x] Write unit test `TestLoadConfig` in `config_test.go` using a sample YAML file

## 3. Temporal Client
- [x] Implement `func NewTemporalClient(cfg TemporalConfig) (client.Client, error)` in `temporal.go`
- [x] Write unit test `TestNewTemporalClient` with a stubbed `client.Dial`

## 4. Tool Registry
- [x] Implement workflow tool registration
- [x] Support dynamic tool definitions based on config
- [x] Add default task queue support
- [x] Write unit tests for task queue selection

## 5. MCP Protocol Handler
- [x] Implement MCP server using `mcp-golang` library
- [x] Add workflow tool registration and execution
- [x] Add system prompt registration
- [x] Implement graceful error handling for Temporal connection failures

## 6. Cache Manager
- [x] Implement `CacheClient` with methods `Get`, `Set`, and `Clear`
- [x] Initialize SQLite database with TTL and max size parameters
- [x] Write unit tests for cache functionality

## 7. ClearCache Tool
- [x] Add `ClearCache` tool definition
- [x] Implement handler for ClearCache calling `CacheClient.Clear`
- [x] Write tests for ClearCache functionality

## 8. Example Configuration
- [x] Create configuration examples (`config.yml` and `config.sample.yml`)
- [x] Add MCP configuration examples in `/examples` directory
- [x] Validate `LoadConfig` parses configuration correctly

## 9. Security & Validation
- [x] Add parameter validation for workflow tools
- [x] Implement safe error handling for failed workflows
- [x] Ensure all logging goes to stderr to avoid corrupting the JSON-RPC protocol

## 10. Performance Benchmarking
- [ ] Add benchmark `BenchmarkToolDiscovery` in `benchmarks/tool_discovery_test.go` to verify <100ms discovery
- [ ] Add benchmark `BenchmarkToolInvocation` in `benchmarks/tool_invocation_test.go` to verify <200ms invocation

## 11. Testing & CI
- [x] Add `make test` target to run all unit tests
- [x] Configure a CI workflow (e.g., GitHub Actions) to run tests on push and PR events

## 12. Documentation
- [x] Update `README.md` with project overview
- [x] Create documentation for setup and configuration instructions
- [x] Document how to run the MCP server
- [x] Document example usage with Claude in `examples/README.md`

## 13. Integration with Claude
- [x] Add example configuration for Claude Desktop in `examples/claude_config.json`
- [x] Provide comprehensive examples for temporal-mcp usage
- [x] Fix logging to ensure proper JSON-RPC communication
- [x] Update build system to use `./bin` directory for binaries