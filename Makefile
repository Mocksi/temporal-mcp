.PHONY: all build clean test fmt vet lint run install help

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

# Binary name
BINARY_NAME=temporal-mcp

# Binary path
BIN_DIR=./bin
BINARY=$(BIN_DIR)/$(BINARY_NAME)



all: clean fmt vet test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BINARY) ./cmd/temporal-mcp
	@echo "Binary built at $(BINARY)"
	@chmod +x $(BINARY)



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

run: build
	@echo "Running $(BINARY_NAME)..."
	$(BINARY)

install:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

help:
	@echo "Makefile commands:"
	@echo "  make build           - Build the temporal-mcp binary"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make test            - Run tests"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Vet code"
	@echo "  make lint            - Lint code"
	@echo "  make run             - Build and run temporal-mcp"
	@echo "  make install         - Install dependencies"
	@echo "  make all             - Clean, format, vet, test, and build"
	@echo "  make help            - Show this help message"
