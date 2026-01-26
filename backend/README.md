# Smart Monitor Backend

## Architecture

Backend được tổ chức theo **Domain-Driven Design (DDD)** architecture. Xem chi tiết tại [README_DDD.md](README_DDD.md).

## Cấu trúc thư mục

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # ← Application entry point
├── internal/
│   ├── domain/                  # Business logic
│   ├── application/             # Use cases
│   └── infrastructure/          # External services
├── pkg/                         # Shared packages
└── static/                      # Swagger UI
```

## Quick Start

### Chạy Backend

**Từ thư mục backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Từ root project:**
```bash
cd smart-monitor
go run backend/cmd/server/main.go
```

**Build và chạy:**
```bash
cd backend
go build -o smart-monitor-backend cmd/server/main.go
./smart-monitor-backend
```

## Endpoints

Sau khi start, backend sẽ expose các endpoints:

- **gRPC Server**: `localhost:50051`
- **HTTP Gateway**: `http://localhost:8080`
- **API**: `http://localhost:8080/v1/`
- **Swagger UI**: `http://localhost:8080/swagger/`
- **Health Check**: `http://localhost:8080/health`
- **Ready Check**: `http://localhost:8080/ready`
- **Live Check**: `http://localhost:8080/live`
- **Metrics**: `http://localhost:8080/metrics`

## Environment Variables

```bash
# gRPC port (default: 50051)
export GRPC_PORT=50051

# HTTP port (default: 8080)
export HTTP_PORT=8080
```

## Testing

### Test endpoints

```bash
# Health check
curl http://localhost:8080/health

# Ready check
curl http://localhost:8080/ready

# Root endpoint
curl http://localhost:8080/

# Metrics
curl http://localhost:8080/metrics
```

### Test with agent

```bash
# Chạy agent trong terminal khác
cd ../agent
go run main.go
```

## Development

### Run tests
```bash
go test ./...
```

### Format code
```bash
go fmt ./...
```

### Lint code
```bash
golangci-lint run
```

## Docker

### Build image
```bash
docker build -t smart-monitor-backend .
```

### Run container
```bash
docker run -p 50051:50051 -p 8080:8080 smart-monitor-backend
```

## Architecture Details

Xem [README_DDD.md](README_DDD.md) để hiểu chi tiết về:
- Domain-Driven Design layers
- Dependency flow
- Best practices
- How to add new features

## Migration Note

- **main.old.go**: File main.go cũ (monolithic structure) - kept for reference
- **cmd/server/main.go**: File main.go mới (DDD structure) - **USE THIS**

## Support

Xem documentation đầy đủ tại `/docs/`
