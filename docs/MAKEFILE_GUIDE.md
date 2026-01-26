# Makefile Quick Reference

## 📖 Overview
Makefile tổng hợp cho Smart Monitor project - quản lý build, run, test và generate code.

## 🚀 Quick Start Commands

```bash
# 1. Cài đặt dependencies
make install-deps

# 2. Generate proto & swagger
make gen-all

# 3. Build tất cả
make build

# 4. Chạy backend
make run-backend

# 5. Chạy agent (terminal khác)
make run-agent
```

## 📋 All Available Commands

### General
- `make help` - Hiển thị tất cả commands
- `make quick-start` - Hướng dẫn bắt đầu nhanh
- `make status` - Kiểm tra trạng thái services
- `make version` - Hiển thị version info

### Build
- `make build` - Build tất cả services
- `make build-backend` - Build backend only
- `make build-agent` - Build agent only
- `make build-monitor-test` - Build monitor-test

### Run (Foreground)
- `make run-backend` - Chạy backend (Ctrl+C để dừng)
- `make run-agent` - Chạy agent (Ctrl+C để dừng)
- `make run-monitor-test` - Chạy monitor test

### Run (Background)
- `make run-all` - Chạy tất cả services ở background
- `make run-backend-bg` - Chạy backend ở background
- `make run-agent-bg` - Chạy agent ở background

### Stop
- `make stop-all` - Dừng tất cả services
- `make stop-backend` - Dừng backend
- `make stop-agent` - Dừng agent

### Generate
- `make gen-all` - Generate proto + swagger
- `make gen-proto` - Generate protobuf files
- `make gen-swagger` - Generate Swagger docs

### Test
- `make test` - Test tất cả
- `make test-backend` - Test backend only
- `make test-agent` - Test agent only
- `make test-integration` - Integration tests

### Code Quality
- `make fmt` - Format code (gofmt)
- `make vet` - Vet code (go vet)
- `make lint` - Lint code (golangci-lint)
- `make check` - Format + vet

### Clean
- `make clean` - Clean build + logs
- `make clean-build` - Xóa build artifacts
- `make clean-logs` - Xóa log files
- `make clean-proto` - Xóa generated proto files
- `make clean-all` - Xóa tất cả

### Logs
- `make logs` - Xem recent logs
- `make logs-backend` - Follow backend logs
- `make logs-agent` - Follow agent logs

### Docker
- `make docker-build` - Build tất cả Docker images
- `make docker-build-backend` - Build backend image
- `make docker-build-agent` - Build agent image
- `make docker-up` - Start với Docker Compose
- `make docker-down` - Stop Docker Compose
- `make docker-logs` - Xem Docker logs

### CI/CD
- `make ci` - CI pipeline: check + test + build
- `make ci-full` - Full CI: clean + deps + gen + check + test + build

### Development
- `make dev` - Development workflow: clean + build + run
- `make install-deps` - Cài đặt dependencies

## 💡 Usage Examples

### Daily Development

```bash
# Bắt đầu ngày mới
make clean build

# Chạy backend để test
make run-backend

# Hoặc chạy tất cả ở background
make run-all

# Kiểm tra status
make status

# Xem logs
make logs

# Dừng khi xong
make stop-all
```

### Adding New Features

```bash
# 1. Sửa proto files
vim pbtypes/monitor/monitor.proto

# 2. Generate lại
make gen-proto

# 3. Implement code

# 4. Format & check
make check

# 5. Test
make test

# 6. Build
make build

# 7. Run để test
make run-backend
```

### Before Commit

```bash
# Format code
make fmt

# Check code quality
make check

# Run tests
make test

# Ensure everything builds
make build

# Or run all at once
make ci
```

### Updating Swagger

```bash
# Edit swagger generator
vim scripts/generate-swagger.sh

# Regenerate
make gen-swagger

# Restart backend để xem
make stop-backend
make run-backend

# Access: http://localhost:8080/swagger/
```

### Troubleshooting

```bash
# Clean everything and rebuild
make clean-all
make install-deps
make gen-all
make build

# Check what's running
make status

# View logs
make logs

# View real-time logs
make logs-backend  # or logs-agent
```

## 📊 Service Endpoints

After running services:

- **Backend HTTP**: http://localhost:8080
- **Backend gRPC**: localhost:50051
- **Swagger UI**: http://localhost:8080/swagger/
- **Health Check**: http://localhost:8080/health
- **Ready Check**: http://localhost:8080/ready

## 🔧 Dependencies

Required:
- Go 1.22+
- protoc (Protocol Buffers compiler)
- make

Optional:
- golangci-lint (for linting)
- Docker (for containerization)
- tree (for project structure)

Install all:
```bash
make install-deps
```

## 📁 Log Files

Logs are stored in `logs/` directory:
- `logs/backend.log` - Backend logs
- `logs/agent.log` - Agent logs
- `logs/backend.pid` - Backend process ID
- `logs/agent.pid` - Agent process ID

## 🎯 Common Workflows

### Full Setup (First Time)
```bash
make install-deps
make gen-all
make build
make run-all
make status
```

### Quick Development
```bash
make dev
```

### Test & Build
```bash
make check test build
```

### Production Build
```bash
make clean-all
make gen-all
make ci-full
```

## ⚡ Tips

1. **Use tab completion**: Type `make` and press Tab twice
2. **Run in background**: Use `-bg` variants for background execution
3. **Check status**: Use `make status` to see what's running
4. **View logs**: Use `make logs` for recent logs
5. **Clean start**: Use `make clean-all` before fresh builds

## 🆘 Help

```bash
# Full help with all commands
make help

# Quick start guide
make quick-start

# Check service status
make status

# View version info
make version
```

## 📚 Related Documentation

- [AGENT_SETUP.md](docs/AGENT_SETUP.md) - Agent installation guide
- [SWAGGER_GUIDE.md](docs/SWAGGER_GUIDE.md) - Swagger documentation
- [README_SWAGGER.md](README_SWAGGER.md) - Swagger overview
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture

---

**Happy coding! 🚀**

For issues or suggestions, run `make help` or check the documentation.
