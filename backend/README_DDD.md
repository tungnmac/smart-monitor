# Backend - DDD Architecture

## Cấu trúc thư mục

```
backend/
├── cmd/                          # Application entry points
│   └── server/
│       └── main.go              # Main server application
├── internal/                     # Private application code
│   ├── domain/                  # Domain Layer (Business Logic)
│   │   ├── entity/              # Business entities
│   │   │   └── stats.go         # Stats and Host entities
│   │   ├── repository/          # Repository interfaces
│   │   │   └── stats_repository.go
│   │   └── service/             # Domain services
│   │       └── stats_service.go # Business logic for stats
│   ├── application/             # Application Layer (Use Cases)
│   │   ├── dto/                 # Data Transfer Objects
│   │   │   └── stats_dto.go
│   │   └── usecase/             # Application use cases
│   │       └── monitor_usecase.go
│   └── infrastructure/          # Infrastructure Layer (External)
│       ├── grpc/                # gRPC handlers
│       │   └── monitor_handler.go
│       ├── http/                # HTTP handlers
│       │   └── handlers.go
│       └── persistence/         # Data persistence
│           └── memory_repository.go
├── pkg/                         # Public shared packages
│   ├── config/                  # Configuration
│   │   └── config.go
│   └── logger/                  # Logging utilities
│       └── logger.go
├── static/                      # Swagger UI static files
└── main.go                      # Legacy main file (can be removed)
```

## Domain-Driven Design Layers

### 1. Domain Layer (`internal/domain/`)
**Trách nhiệm**: Chứa business logic thuần túy, không phụ thuộc vào framework hay infrastructure.

#### Entities (`entity/`)
- `Stats`: Core business entity representing system metrics
- `Host`: Entity representing monitored hosts
- Business rules và validation

#### Repository Interfaces (`repository/`)
- Định nghĩa contracts cho data access
- Không implement, chỉ định nghĩa interface

#### Domain Services (`service/`)
- Business logic phức tạp không thuộc về entity
- Orchestrate operations giữa nhiều entities

**Ví dụ**:
```go
// Entity with business rules
type Stats struct {
    Hostname string
    CPU      float64
    // ...
}

func (s *Stats) IsValid() bool {
    return s.CPU >= 0 && s.CPU <= 100
}
```

### 2. Application Layer (`internal/application/`)
**Trách nhiệm**: Điều phối use cases của ứng dụng.

#### Use Cases (`usecase/`)
- Implement business workflows
- Gọi domain services
- Convert giữa DTOs và entities

#### DTOs (`dto/`)
- Data Transfer Objects
- Dùng để transfer data giữa layers
- Không chứa business logic

**Ví dụ**:
```go
func (uc *MonitorUseCase) RecordStats(ctx context.Context, req *dto.StatsRequest) error {
    // Convert DTO to Entity
    stats := entity.NewStats(req.Hostname, req.CPU, req.RAM, req.Disk)
    
    // Call domain service
    return uc.statsService.ProcessStats(ctx, stats)
}
```

### 3. Infrastructure Layer (`internal/infrastructure/`)
**Trách nhiệm**: Implement technical details và external communications.

#### gRPC (`grpc/`)
- gRPC service handlers
- Convert protobuf messages to DTOs

#### HTTP (`http/`)
- HTTP handlers (health checks, metrics)
- REST endpoints

#### Persistence (`persistence/`)
- Implement repository interfaces
- Data access implementations (in-memory, database, etc.)

**Ví dụ**:
```go
// Implement repository interface
type InMemoryStatsRepository struct {
    stats map[string]*entity.Stats
}

func (r *InMemoryStatsRepository) Save(ctx context.Context, stats *entity.Stats) error {
    r.stats[stats.Hostname] = stats
    return nil
}
```

### 4. Package Layer (`pkg/`)
**Trách nhiệm**: Reusable utilities có thể dùng ở nhiều nơi.

- `config`: Configuration management
- `logger`: Logging utilities

## Dependency Flow

```
┌─────────────────────────────────────────┐
│         cmd/server (main.go)            │
│    (Wiring dependencies together)       │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Infrastructure Layer               │
│  (gRPC, HTTP, Persistence)              │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Application Layer                  │
│  (Use Cases, DTOs)                      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Domain Layer                    │
│  (Entities, Services, Repositories)     │
│     (Core Business Logic)               │
└─────────────────────────────────────────┘
```

**Quan trọng**: Dependencies chỉ flow từ ngoài vào trong. Domain layer không biết gì về Infrastructure.

## Ưu điểm của DDD Architecture

### 1. **Separation of Concerns**
- Mỗi layer có trách nhiệm rõ ràng
- Dễ maintain và test

### 2. **Testability**
```go
// Test domain logic without infrastructure
func TestStatsValidation(t *testing.T) {
    stats := entity.NewStats("server-1", 45.5, 60.2, 75.0)
    assert.True(t, stats.IsValid())
}

// Test use case with mock repository
func TestRecordStats(t *testing.T) {
    mockRepo := &MockStatsRepository{}
    service := service.NewStatsService(mockRepo, nil)
    useCase := usecase.NewMonitorUseCase(service)
    // Test use case...
}
```

### 3. **Flexibility**
- Dễ thay đổi infrastructure (in-memory → database)
- Không ảnh hưởng đến business logic

### 4. **Scalability**
- Dễ thêm features mới
- Dễ refactor code

### 5. **Clean Code**
- Code rõ ràng, dễ đọc
- Follow SOLID principles

## Chạy ứng dụng

### Development
```bash
# Từ thư mục backend/
cd cmd/server
go run main.go

# Hoặc từ root
cd backend
go run cmd/server/main.go
```

### Build
```bash
cd backend
go build -o smart-monitor-backend cmd/server/main.go
./smart-monitor-backend
```

### With Docker
```bash
# Build image
docker build -t smart-monitor-backend .

# Run container
docker run -p 50051:50051 -p 8080:8080 smart-monitor-backend
```

## Testing

### Unit Tests
```bash
# Test domain layer
go test ./internal/domain/...

# Test application layer
go test ./internal/application/...

# Test all
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

## Migration từ cấu trúc cũ

1. **main.go cũ** → Được refactor thành:
   - `cmd/server/main.go`: Application entry point
   - `internal/infrastructure/grpc/monitor_handler.go`: gRPC handlers
   - `internal/infrastructure/http/handlers.go`: HTTP handlers

2. **Business logic** → Moved to:
   - `internal/domain/entity/`: Entities
   - `internal/domain/service/`: Domain services
   - `internal/application/usecase/`: Use cases

3. **Data access** → Organized in:
   - `internal/domain/repository/`: Interfaces
   - `internal/infrastructure/persistence/`: Implementations

## Thêm features mới

### Ví dụ: Thêm Alert Service

1. **Domain Layer**: Tạo entity và service
```go
// internal/domain/entity/alert.go
type Alert struct {
    ID        string
    Hostname  string
    Metric    string
    Threshold float64
    // ...
}

// internal/domain/service/alert_service.go
type AlertService struct {
    // ...
}
```

2. **Application Layer**: Tạo use case
```go
// internal/application/usecase/alert_usecase.go
type AlertUseCase struct {
    alertService *service.AlertService
}
```

3. **Infrastructure Layer**: Tạo handlers
```go
// internal/infrastructure/grpc/alert_handler.go
type AlertServiceServer struct {
    alertUseCase *usecase.AlertUseCase
}
```

## Best Practices

1. **Domain First**: Bắt đầu với domain entities và business rules
2. **Interface Segregation**: Tách interfaces nhỏ, specific
3. **Dependency Injection**: Inject dependencies qua constructor
4. **Immutability**: Prefer immutable objects khi có thể
5. **Error Handling**: Return errors, don't panic
6. **Testing**: Write tests cho mỗi layer

## Resources

- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

**Version**: 1.0.0  
**Architecture**: Domain-Driven Design (DDD)  
**Last Updated**: January 15, 2026
