.PHONY: build clean test install run help tidy build-server run-server docker-build

# Binary name
BINARY_NAME=kubehelp
SERVER_NAME=kubehelp-server

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

build: tidy ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-server: tidy ## Build the server binary
	@echo "Building $(SERVER_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_NAME) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(SERVER_NAME)"

tidy: ## Download dependencies and clean up go.mod
	@echo "Running go mod tidy..."
	$(GOMOD) tidy

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(SERVER_NAME)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

install: build ## Install CLI binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Install complete"

run: build ## Build and run CLI with example flags
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) diagnose -n default --verbose

run-server: build-server ## Build and run the server
	@echo "Running $(SERVER_NAME)..."
	./$(SERVER_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t kubehelp-server:latest .
	@echo "Docker build complete"

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run --network kind -p 8080:8080 \
		-v ~/.kube:/root/.kube:ro \
		-e OLLAMA_BASE_URL=http://host.docker.internal:11434 \
		kubehelp-server:latest

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

all: clean build build-server test ## Clean, build CLI and server, and test
