# Smart Monitor - Quick Start Guide

Chào mừng đến với hệ thống giám sát Smart Monitor! Tài liệu này sẽ giúp bạn nhanh chóng hiểu và sử dụng hệ thống.

## 📚 Tài liệu chi tiết

Hệ thống Smart Monitor bao gồm các tài liệu sau:

### 1. [README.md](README.md) - Tổng quan hệ thống
- Giới thiệu về Smart Monitor
- Kiến trúc tổng thể
- Quick start guide
- Features và roadmap

### 2. [ARCHITECTURE.md](ARCHITECTURE.md) - Kiến trúc chi tiết
- Kiến trúc hệ thống
- Components và luồng dữ liệu
- Design patterns
- Scalability & Performance
- Security architecture

### 3. [INFRASTRUCTURE.md](INFRASTRUCTURE.md) - Cấu trúc hạ tầng
- Cấu trúc Protocol Buffers
- Service definitions chi tiết
- Network architecture
- High availability setup
- Monitoring & observability

### 4. [DEVELOPMENT.md](DEVELOPMENT.md) - Hướng dẫn phát triển
- Setup môi trường
- Coding standards
- Testing practices
- Development workflow
- Debugging và profiling

### 5. [DEPLOYMENT.md](DEPLOYMENT.md) - Hướng dẫn triển khai
- Build process
- Deployment methods (Manual, Docker, K8s)
- Configuration management
- Security setup
- Monitoring và maintenance

### 6. [API.md](API.md) - API Documentation
- gRPC Services
- REST API endpoints
- Authentication
- Code examples
- Testing APIs

## 🚀 Quick Start

### Prerequisites
```bash
# Install Go 1.24+
go version

# Install protoc
protoc --version

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

### Setup
```bash
# Clone repository
git clone <repository-url>
cd smart-monitor

# Install dependencies
go mod download

# Generate protobuf files
cd pbtypes
./run_makefile.sh
cd ..
```

### Run Backend
```bash
cd backend
go run main.go

# Backend started on:
# - gRPC: localhost:50051
# - HTTP: http://localhost:8080
# - Swagger: http://localhost:8080/swagger/
```

### Run Agent
```bash
cd agent
go run main.go

# Agent will start sending metrics every 2 seconds
```

## 📖 Đọc tài liệu theo vai trò

### Nếu bạn là Developer
1. Đọc [README.md](README.md) để hiểu tổng quan
2. Đọc [ARCHITECTURE.md](ARCHITECTURE.md) để hiểu kiến trúc
3. Đọc [DEVELOPMENT.md](DEVELOPMENT.md) để setup và code
4. Đọc [API.md](API.md) để tích hợp APIs

### Nếu bạn là DevOps/SRE
1. Đọc [README.md](README.md) để hiểu hệ thống
2. Đọc [INFRASTRUCTURE.md](INFRASTRUCTURE.md) để hiểu hạ tầng
3. Đọc [DEPLOYMENT.md](DEPLOYMENT.md) để triển khai
4. Đọc [ARCHITECTURE.md](ARCHITECTURE.md) phần Security & Scaling

### Nếu bạn là QA/Tester
1. Đọc [README.md](README.md) để hiểu features
2. Đọc [API.md](API.md) để test APIs
3. Đọc [DEVELOPMENT.md](DEVELOPMENT.md) phần Testing

### Nếu bạn là Product Manager
1. Đọc [README.md](README.md) để hiểu tổng quan và roadmap
2. Đọc [ARCHITECTURE.md](ARCHITECTURE.md) để hiểu khả năng của hệ thống
3. Đọc [INFRASTRUCTURE.md](INFRASTRUCTURE.md) phần Future Enhancements

## 🎯 Use Cases

### 1. Giám sát server realtime
```bash
# Chạy agent trên server cần giám sát
./smart-monitor-agent --server=backend:50051

# Xem metrics trên dashboard
# http://localhost:8080/swagger/
```

### 2. Monitor Docker containers
```bash
# Deploy with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

### 3. Deploy lên Kubernetes
```bash
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/agent-daemonset.yaml
```

## 🔧 Troubleshooting

### Backend không start
```bash
# Check logs
sudo journalctl -u smart-monitor-backend -f

# Check port
sudo netstat -tlnp | grep 50051
```

### Agent không kết nối
```bash
# Test connection
telnet backend-server 50051

# Check firewall
sudo ufw status
```

### Proto generation fails
```bash
# Verify protoc
which protoc

# Re-install plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

## 📞 Support

- **Documentation**: Xem các file trong thư mục `docs/`
- **Issues**: Tạo issue trên GitHub
- **Email**: support@smart-monitor.com

## 🗺️ Project Roadmap

### ✅ Phase 1 (Current)
- Basic monitoring (CPU, RAM, Disk)
- gRPC streaming
- REST API gateway
- Swagger documentation

### 🔄 Phase 2 (In Progress)
- Full metrics support
- Database persistence
- Basic dashboard
- Authentication

### 📋 Phase 3 (Planned)
- Advanced alerting
- Historical analysis
- Container monitoring
- Multi-tenant support

## 📝 License

[Specify your license]

---

**Happy Monitoring! 🎉**
