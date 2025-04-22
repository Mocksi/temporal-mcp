.PHONY: all build build-temporal build-piglatin clean test fmt vet lint run install help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary names
TEMPORAL_BINARY_NAME=temporal-mcp
PIGLATIN_BINARY_NAME=piglatin-mcp

# Binary paths
BIN_DIR=./bin
TEMPORAL_BINARY=$(BIN_DIR)/$(TEMPORAL_BINARY_NAME)
PIGLATIN_BINARY=$(BIN_DIR)/$(PIGLATIN_BINARY_NAME)



all: clean fmt vet test build

build: build-temporal build-piglatin
	@echo "All binaries built successfully"

build-temporal:
	@echo "Building temporal-mcp..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(TEMPORAL_BINARY) ./cmd/temporal-mcp
	@echo "Binary built at $(TEMPORAL_BINARY)"
	@chmod +x $(TEMPORAL_BINARY)

build-piglatin:
	@echo "Building piglatin-mcp..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(PIGLATIN_BINARY) ./cmd/piglatin-mcp
	@echo "Binary built at $(PIGLATIN_BINARY)"
	@chmod +x $(PIGLATIN_BINARY)

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@echo "Cleaned build artifacts"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

fmt:
	@echo "Formatting code..."
	find ./cmd ./internal ./pkg ./test -type f -name "*.go" | xargs -I{} go fmt {}

vet:
	@echo "Vetting code..."
	$(GOVET) ./...

lint:
	@echo "Linting code..."
	$(GOLINT) run

run: build-temporal
	@echo "Running temporal-mcp..."
	$(TEMPORAL_BINARY)

install:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

help:
	@echo "Makefile commands:"
	@echo "  make build           - Build all binaries"
	@echo "  make build-temporal  - Build only the temporal-mcp binary"
	@echo "  make build-piglatin  - Build only the piglatin-mcp binary"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make test            - Run tests"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Vet code"
	@echo "  make lint            - Lint code"
	@echo "  make run             - Build and run temporal-mcp"
	@echo "  make install         - Install dependencies"
	@echo "  make all             - Clean, format, vet, test, and build"
	@echo "  make help            - Show this help message"
