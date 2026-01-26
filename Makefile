# Smart Monitor - Makefile
# Tổng hợp các lệnh quản lý project

.PHONY: help build run clean test gen-proto gen-swagger install-deps

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

##@ General

help: ## Hiển thị danh sách commands
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║          Smart Monitor - Makefile Commands                     ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(YELLOW)<target>$(NC)\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

install-deps: ## Cài đặt dependencies (protoc, gRPC, etc.)
	@echo "$(BLUE)📦 Installing dependencies...$(NC)"
	@command -v protoc >/dev/null 2>&1 || { echo "$(RED)❌ protoc not found. Please install: https://grpc.io/docs/protoc-installation/$(NC)"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go not found. Please install Go 1.22+$(NC)"; exit 1; }
	@echo "$(GREEN)✅ Installing Go dependencies...$(NC)"
	@cd backend && go mod download
	@cd agent && go mod download
	@echo "$(GREEN)✅ Installing protoc plugins...$(NC)"
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "$(GREEN)✅ All dependencies installed!$(NC)"

##@ Build

build: build-backend build-agent ## Build tất cả services
	@echo "$(GREEN)✅ All services built successfully!$(NC)"

build-backend: ## Build backend service
	@echo "$(BLUE)🔨 Building backend...$(NC)"
	@cd backend && go build -o backend cmd/server/main.go
	@echo "$(GREEN)✅ Backend built: backend/backend$(NC)"

build-agent: ## Build agent
	@echo "$(BLUE)🔨 Building agent...$(NC)"
	@cd agent && go build -o agent main.go
	@echo "$(GREEN)✅ Agent built: agent/agent$(NC)"

build-monitor-test: ## Build monitor test tool
	@echo "$(BLUE)🔨 Building monitor-test...$(NC)"
	@cd monitor-test && go build -o monitor-test main.go
	@echo "$(GREEN)✅ Monitor-test built: monitor-test/monitor-test$(NC)"

##@ Run

run-backend: ## Chạy backend service
	@echo "$(BLUE)🚀 Starting backend service...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@cd backend && go run cmd/server/main.go

run-agent: ## Chạy agent (cần backend đang chạy)
	@echo "$(BLUE)🚀 Starting agent...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@cd agent && go run main.go

run-monitor-test: ## Chạy monitor test tool
	@echo "$(BLUE)🧪 Running monitor-test...$(NC)"
	@cd monitor-test && go run main.go

run-all: ## Chạy tất cả services (background mode)
	@echo "$(BLUE)🚀 Starting all services...$(NC)"
	@make run-backend-bg
	@sleep 3
	@make run-agent-bg
	@echo "$(GREEN)✅ All services started!$(NC)"
	@echo "$(YELLOW)Backend:$(NC) http://localhost:8080"
	@echo "$(YELLOW)gRPC:$(NC)    localhost:50051"
	@echo "$(YELLOW)Swagger:$(NC) http://localhost:8080/swagger/"
	@echo ""
	@echo "$(BLUE)To stop: make stop-all$(NC)"

run-backend-bg: ## Chạy backend ở background
	@echo "$(BLUE)🚀 Starting backend in background...$(NC)"
	@cd backend && nohup go run cmd/server/main.go > ../logs/backend.log 2>&1 & echo $$! > ../logs/backend.pid
	@echo "$(GREEN)✅ Backend started (PID: $$(cat logs/backend.pid))$(NC)"

run-agent-bg: ## Chạy agent ở background
	@echo "$(BLUE)🚀 Starting agent in background...$(NC)"
	@cd agent && nohup go run main.go > ../logs/agent.log 2>&1 & echo $$! > ../logs/agent.pid
	@echo "$(GREEN)✅ Agent started (PID: $$(cat logs/agent.pid))$(NC)"

##@ Stop

stop-all: stop-backend stop-agent ## Dừng tất cả services
	@echo "$(GREEN)✅ All services stopped!$(NC)"

stop-backend: ## Dừng backend service
	@if [ -f logs/backend.pid ]; then \
		echo "$(BLUE)🛑 Stopping backend (PID: $$(cat logs/backend.pid))...$(NC)"; \
		kill $$(cat logs/backend.pid) 2>/dev/null || true; \
		rm -f logs/backend.pid; \
		echo "$(GREEN)✅ Backend stopped$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  Backend not running (no PID file)$(NC)"; \
	fi

stop-agent: ## Dừng agent
	@if [ -f logs/agent.pid ]; then \
		echo "$(BLUE)🛑 Stopping agent (PID: $$(cat logs/agent.pid))...$(NC)"; \
		kill $$(cat logs/agent.pid) 2>/dev/null || true; \
		rm -f logs/agent.pid; \
		echo "$(GREEN)✅ Agent stopped$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  Agent not running (no PID file)$(NC)"; \
	fi

##@ Generate

gen-proto: ## Generate protobuf files
	@echo "$(BLUE)🔄 Generating protobuf files...$(NC)"
	@cd pbtypes && chmod +x generate_proto.sh && ./generate_proto.sh
	@echo "$(GREEN)✅ Protobuf files generated!$(NC)"

gen-swagger: ## Generate Swagger documentation
	@echo "$(BLUE)📚 Generating Swagger documentation...$(NC)"
	@chmod +x scripts/generate-swagger.sh && ./scripts/generate-swagger.sh

gen-all: gen-proto gen-swagger ## Generate tất cả (proto + swagger)
	@echo "$(GREEN)✅ All files generated!$(NC)"

##@ Test

test: ## Chạy tests cho tất cả services
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@cd backend && go test -v ./...
	@cd agent && go test -v ./...
	@echo "$(GREEN)✅ All tests passed!$(NC)"

test-backend: ## Test backend only
	@echo "$(BLUE)🧪 Testing backend...$(NC)"
	@cd backend && go test -v ./...

test-agent: ## Test agent only
	@echo "$(BLUE)🧪 Testing agent...$(NC)"
	@cd agent && go test -v ./...

test-integration: build-backend build-agent ## Chạy integration tests
	@echo "$(BLUE)🧪 Running integration tests...$(NC)"
	@make run-backend-bg
	@sleep 3
	@cd agent && go test -v -tags=integration ./...
	@make stop-backend
	@echo "$(GREEN)✅ Integration tests passed!$(NC)"

##@ Lint & Format

fmt: ## Format code (gofmt)
	@echo "$(BLUE)🎨 Formatting code...$(NC)"
	@cd backend && go fmt ./...
	@cd agent && go fmt ./...
	@cd monitor-test && go fmt ./...
	@echo "$(GREEN)✅ Code formatted!$(NC)"

lint: ## Lint code (golangci-lint)
	@echo "$(BLUE)🔍 Linting code...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(YELLOW)⚠️  golangci-lint not found. Install: https://golangci-lint.run/usage/install/$(NC)"; exit 1; }
	@cd backend && golangci-lint run ./...
	@cd agent && golangci-lint run ./...
	@echo "$(GREEN)✅ Lint check passed!$(NC)"

vet: ## Vet code (go vet)
	@echo "$(BLUE)🔍 Vetting code...$(NC)"
	@cd backend && go vet ./...
	@cd agent && go vet ./...
	@cd monitor-test && go vet ./...
	@echo "$(GREEN)✅ Vet check passed!$(NC)"

check: fmt vet ## Format & vet code
	@echo "$(GREEN)✅ Code checks completed!$(NC)"

##@ Clean

clean: clean-build clean-logs ## Clean tất cả
	@echo "$(GREEN)✅ Cleaned all!$(NC)"

clean-build: ## Xóa build artifacts
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(NC)"
	@rm -f backend/backend
	@rm -f agent/agent
	@rm -f monitor-test/monitor-test
	@rm -f agent/.agent_token
	@echo "$(GREEN)✅ Build artifacts cleaned!$(NC)"

clean-logs: ## Xóa log files
	@echo "$(BLUE)🧹 Cleaning logs...$(NC)"
	@rm -f logs/*.log
	@rm -f logs/*.pid
	@mkdir -p logs
	@echo "$(GREEN)✅ Logs cleaned!$(NC)"

clean-proto: ## Xóa generated proto files
	@echo "$(BLUE)🧹 Cleaning generated proto files...$(NC)"
	@find pbtypes -name "*.pb.go" -delete
	@find pbtypes -name "*.pb.gw.go" -delete
	@find pbtypes -name "*.swagger.json" -delete
	@echo "$(GREEN)✅ Proto files cleaned!$(NC)"

clean-all: clean clean-proto ## Xóa tất cả (bao gồm proto)
	@echo "$(GREEN)✅ Complete cleanup done!$(NC)"

##@ Docker

docker-build-backend: ## Build backend Docker image
	@echo "$(BLUE)🐳 Building backend Docker image...$(NC)"
	@docker build -t smart-monitor-backend:latest -f backend/Dockerfile .
	@echo "$(GREEN)✅ Backend image built!$(NC)"

docker-build-agent: ## Build agent Docker image
	@echo "$(BLUE)🐳 Building agent Docker image...$(NC)"
	@docker build -t smart-monitor-agent:latest -f agent/Dockerfile .
	@echo "$(GREEN)✅ Agent image built!$(NC)"

docker-build: docker-build-backend docker-build-agent ## Build tất cả Docker images
	@echo "$(GREEN)✅ All Docker images built!$(NC)"

docker-up: ## Start services với Docker Compose
	@echo "$(BLUE)🐳 Starting services with Docker Compose...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✅ Services started!$(NC)"

docker-down: ## Stop Docker Compose services
	@echo "$(BLUE)🐳 Stopping Docker Compose services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✅ Services stopped!$(NC)"

docker-logs: ## Xem Docker logs
	@docker-compose logs -f

##@ Database (Future)

db-migrate: ## Run database migrations (future)
	@echo "$(YELLOW)⚠️  Database migrations not implemented yet$(NC)"

db-seed: ## Seed database (future)
	@echo "$(YELLOW)⚠️  Database seeding not implemented yet$(NC)"

##@ Info

status: ## Kiểm tra trạng thái services
	@echo "$(BLUE)📊 Service Status:$(NC)"
	@echo ""
	@if [ -f logs/backend.pid ] && kill -0 $$(cat logs/backend.pid) 2>/dev/null; then \
		echo "$(GREEN)✅ Backend:$(NC) Running (PID: $$(cat logs/backend.pid))"; \
	else \
		echo "$(RED)❌ Backend:$(NC) Not running"; \
	fi
	@if [ -f logs/agent.pid ] && kill -0 $$(cat logs/agent.pid) 2>/dev/null; then \
		echo "$(GREEN)✅ Agent:$(NC)   Running (PID: $$(cat logs/agent.pid))"; \
	else \
		echo "$(RED)❌ Agent:$(NC)   Not running"; \
	fi
	@echo ""
	@echo "$(BLUE)📂 Build Artifacts:$(NC)"
	@if [ -f backend/backend ]; then echo "$(GREEN)✅$(NC) backend/backend"; else echo "$(RED)❌$(NC) backend/backend"; fi
	@if [ -f agent/agent ]; then echo "$(GREEN)✅$(NC) agent/agent"; else echo "$(RED)❌$(NC) agent/agent"; fi
	@echo ""
	@echo "$(BLUE)🌐 Endpoints:$(NC)"
	@echo "  Backend HTTP: http://localhost:8080"
	@echo "  Backend gRPC: localhost:50051"
	@echo "  Swagger UI:   http://localhost:8080/swagger/"
	@echo "  Health:       http://localhost:8080/health"

logs: ## Xem logs
	@echo "$(BLUE)📝 Recent logs:$(NC)"
	@echo ""
	@if [ -f logs/backend.log ]; then \
		echo "$(YELLOW)=== Backend Logs (last 20 lines) ===$(NC)"; \
		tail -20 logs/backend.log; \
		echo ""; \
	fi
	@if [ -f logs/agent.log ]; then \
		echo "$(YELLOW)=== Agent Logs (last 20 lines) ===$(NC)"; \
		tail -20 logs/agent.log; \
	fi

logs-backend: ## Xem backend logs
	@tail -f logs/backend.log

logs-agent: ## Xem agent logs
	@tail -f logs/agent.log

version: ## Hiển thị version info
	@echo "$(BLUE)📌 Smart Monitor Version Information:$(NC)"
	@echo ""
	@echo "Go version:     $$(go version | awk '{print $$3}')"
	@echo "Protoc version: $$(protoc --version 2>/dev/null || echo 'not installed')"
	@echo "Project:        Smart Monitor v1.0.0"
	@echo ""

tree: ## Hiển thị cấu trúc project
	@echo "$(BLUE)📁 Project Structure:$(NC)"
	@tree -L 2 -I 'node_modules|vendor|.git' --dirsfirst

##@ Quick Start

dev: clean build run-backend ## Development workflow: clean, build, run
	@echo "$(GREEN)✅ Development environment ready!$(NC)"

quick-start: ## Quick start guide
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║              Smart Monitor - Quick Start                       ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Install dependencies:$(NC)"
	@echo "   make install-deps"
	@echo ""
	@echo "$(YELLOW)2. Generate proto & swagger:$(NC)"
	@echo "   make gen-all"
	@echo ""
	@echo "$(YELLOW)3. Build services:$(NC)"
	@echo "   make build"
	@echo ""
	@echo "$(YELLOW)4. Run backend:$(NC)"
	@echo "   make run-backend"
	@echo "   (in new terminal)"
	@echo ""
	@echo "$(YELLOW)5. Run agent:$(NC)"
	@echo "   make run-agent"
	@echo ""
	@echo "$(YELLOW)6. Access Swagger UI:$(NC)"
	@echo "   http://localhost:8080/swagger/"
	@echo ""
	@echo "$(GREEN)For more commands: make help$(NC)"

##@ CI/CD

ci: check test build ## CI pipeline: format, vet, test, build
	@echo "$(GREEN)✅ CI checks passed!$(NC)"

ci-full: clean install-deps gen-all check test build ## Full CI pipeline
	@echo "$(GREEN)✅ Full CI pipeline completed!$(NC)"
