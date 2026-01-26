# Smart Monitor System

## Giới thiệu

Smart Monitor là hệ thống giám sát hiệu suất và tài nguyên hệ thống theo thời gian thực, được xây dựng với kiến trúc phân tán sử dụng gRPC và Protocol Buffers.

## Tổng quan kiến trúc

Hệ thống bao gồm 3 thành phần chính:

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Agent     │─────▶│   Backend   │─────▶│  Frontend   │
│  (Client)   │ gRPC │   (Server)  │ HTTP │    (UI)     │
└─────────────┘      └─────────────┘      └─────────────┘
```

### 1. Agent (Monitoring Client)
- Thu thập metrics từ hệ thống (CPU, RAM, Disk, Network, Processes)
- Gửi dữ liệu đến Backend qua gRPC stream
- Chạy như một service trên các máy cần giám sát

### 2. Backend (gRPC Server)
- Nhận dữ liệu từ nhiều agents
- Xử lý và lưu trữ metrics
- Cung cấp REST API thông qua gRPC Gateway
- Quản lý Swagger UI cho API documentation

### 3. Frontend (Web UI)
- Hiển thị dashboard theo thời gian thực
- Visualize metrics và alerts
- Quản lý cấu hình monitoring

## Công nghệ sử dụng

- **Backend**: Go 1.24
- **Protocol**: gRPC + Protocol Buffers
- **API Gateway**: grpc-gateway/v2
- **Monitoring Library**: gopsutil/v3
- **API Documentation**: Swagger/OpenAPI

## Cấu trúc thư mục

```
smart-monitor/
├── agent/              # Monitoring agent code
├── backend/            # gRPC server & API gateway
├── frontend/           # Web UI (React/Vue/Angular)
├── pbtypes/            # Protocol Buffer definitions
│   ├── monitor/        # Core monitoring services
│   ├── system/         # System information
│   ├── process/        # Process monitoring
│   ├── network/        # Network monitoring
│   ├── disk/           # Disk monitoring
│   ├── logs/           # Log collection
│   └── Infrastructure/ # Infrastructure management
├── docs/               # Documentation
└── third_party/        # External dependencies

```

## Quick Start

### Yêu cầu

- Go 1.24+
- Protocol Buffers compiler (protoc)
- Make

### Cài đặt

```bash
# Clone repository
git clone <repository-url>
cd smart-monitor

# Install dependencies
go mod download

# Generate Protocol Buffers
cd pbtypes
./run_makefile.sh
```

### Chạy Backend

```bash
cd backend
go run main.go
```

Backend sẽ khởi động:
- gRPC server: `localhost:50051`
- HTTP Gateway: `localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/`

### Chạy Agent

```bash
cd agent
go run main.go
```

Agent sẽ bắt đầu gửi metrics đến backend mỗi 2 giây.

## Tài liệu chi tiết

- [Kiến trúc hệ thống](ARCHITECTURE.md) - Chi tiết về design patterns và components
- [Cấu trúc hạ tầng](INFRASTRUCTURE.md) - Mô tả các services và protobuf definitions
- [Hướng dẫn phát triển](DEVELOPMENT.md) - Setup môi trường và coding standards
- [Hướng dẫn triển khai](DEPLOYMENT.md) - Deploy production và configuration
- [API Documentation](API.md) - Chi tiết về gRPC services và REST endpoints

## Monitoring Features

### Đã triển khai
- ✅ CPU monitoring
- ✅ Memory (RAM) monitoring
- ✅ gRPC streaming
- ✅ REST API gateway
- ✅ Swagger documentation

### Đang phát triển
- 🔄 Disk monitoring
- 🔄 Network monitoring
- 🔄 Process monitoring
- 🔄 Log collection
- 🔄 Container monitoring
- 🔄 Security monitoring

### Kế hoạch
- 📋 User management
- 📋 Alert system
- 📋 Data persistence (Database)
- 📋 Dashboard visualization
- 📋 Historical data analysis
- 📋 Multi-tenant support

## Contributing

Vui lòng đọc [DEVELOPMENT.md](DEVELOPMENT.md) để biết chi tiết về coding standards và quy trình đóng góp code.

## License

[Specify your license here]

## Liên hệ

[Specify contact information]
