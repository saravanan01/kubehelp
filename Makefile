.PHONY: build clean test install run help tidy

# Binary name
BINARY_NAME=kubehelp

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: tidy ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/...
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

tidy: ## Download dependencies and clean up go.mod
	@echo "Running go mod tidy..."
	$(GOMOD) tidy

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Install complete"

run: build ## Build and run with example flags
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) diagnose -n default --verbose

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete"

deps: ## Download all dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...
	$(GOMOD) download

all: clean build test ## Clean, build, and test
