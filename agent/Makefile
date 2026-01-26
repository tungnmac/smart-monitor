.PHONY: build run clean test install

BINARY_NAME=agent
BUILD_DIR=bin
MAIN_PATH=./cmd/agent

# Build the agent
build:
	@echo "🔨 Building agent..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the agent
run: build
	@echo "🚀 Starting agent..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with custom config
run-dev:
	@echo "🚀 Starting agent (development mode)..."
	@METRICS_INTERVAL=2 ENVIRONMENT=development $(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f .agent_token
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -cover ./...

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Format code
fmt:
	@echo "💅 Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run ./...

# Build for multiple platforms
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Multi-platform build complete"

# Show help
help:
	@echo "Smart Monitor Agent - Makefile Commands"
	@echo ""
	@echo "Usage: make [command]"
	@echo ""
	@echo "Commands:"
	@echo "  build         Build the agent binary"
	@echo "  run           Build and run the agent"
	@echo "  run-dev       Run in development mode"
	@echo "  clean         Remove build artifacts"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  install       Install dependencies"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  build-all     Build for multiple platforms"
	@echo "  help          Show this help message"
