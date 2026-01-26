# Kiến trúc hệ thống Smart Monitor

## 1. Tổng quan kiến trúc

### 1.1. Kiến trúc tổng thể

```
┌────────────────────────────────────────────────────────────────┐
│                        Frontend Layer                          │
│                    (Web UI / Dashboard)                        │
└────────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTP/REST
                              │
┌────────────────────────────────────────────────────────────────┐
│                      Backend Layer                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │ gRPC Server  │    │   Gateway    │    │   Swagger    │   │
│  │  :50051      │◀──▶│   :8080      │◀──▶│     UI       │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
│         ▲                                                      │
└─────────┼──────────────────────────────────────────────────────┘
          │ gRPC Streaming
          │
┌─────────┴──────────────────────────────────────────────────────┐
│                      Agent Layer                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Agent 1  │  │ Agent 2  │  │ Agent 3  │  │ Agent N  │     │
│  │ Server-1 │  │ Server-2 │  │ Server-3 │  │ Server-N │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
└────────────────────────────────────────────────────────────────┘
```

### 1.2. Luồng dữ liệu (Data Flow)

```
Agent                Backend               Frontend
  │                     │                     │
  │──[1] Collect────▶   │                     │
  │    Metrics          │                     │
  │                     │                     │
  │──[2] gRPC Stream─▶  │                     │
  │    (CPU, RAM...)    │                     │
  │                     │                     │
  │                     │──[3] Process────▶   │
  │                     │    & Store          │
  │                     │                     │
  │                     │◀──[4] HTTP GET───   │
  │                     │    Request          │
  │                     │                     │
  │                     │──[5] JSON Reply──▶  │
  │                     │                     │
```

## 2. Components chi tiết

### 2.1. Agent (Monitoring Client)

**Chức năng:**
- Thu thập system metrics từ host machine
- Gửi dữ liệu realtime qua gRPC streaming
- Tự động retry khi mất kết nối

**Metrics thu thập:**
- CPU usage percentage
- Memory (RAM) usage
- Disk I/O và usage
- Network traffic
- Process information
- System logs
- Container metrics (Docker/K8s)

**Công nghệ:**
- `gopsutil/v3`: Thu thập system metrics
- gRPC client: Giao tiếp với backend
- Ticker mechanism: Định kỳ gửi dữ liệu

### 2.2. Backend (Server)

**Components:**

#### a) gRPC Server (Port 50051)
```go
// Xử lý bidirectional streaming
func (s *server) StreamStats(stream MonitorService_StreamStatsServer) error
```

**Chức năng:**
- Nhận metrics từ nhiều agents đồng thời
- Xử lý bidirectional streaming
- Validate và transform data
- Push notifications cho alerts

#### b) gRPC Gateway (Port 8080)
```go
// Tự động generate REST API từ protobuf
runtime.NewServeMux()
```

**Chức năng:**
- Chuyển đổi gRPC calls thành REST API
- Tự động mapping từ protobuf definitions
- CORS handling
- Request/Response transformation

#### c) Swagger UI
**Chức năng:**
- Interactive API documentation
- API testing interface
- Auto-generated từ .proto files

### 2.3. Frontend (Dashboard)

**Features:**
- Real-time metrics visualization
- Historical data charts
- Alert management
- Multi-server monitoring
- Custom dashboards

**Công nghệ đề xuất:**
- React/Vue/Angular
- WebSocket cho real-time updates
- Chart libraries (Chart.js, D3.js)
- Material UI / Ant Design

## 3. Protocol Buffers Structure

### 3.1. Core Monitoring Services

```protobuf
// pbtypes/monitor/monitor.proto
service MonitorService {
  rpc StreamStats(stream StatsRequest) returns (StatsResponse);
  rpc GetStats(StatsRequest) returns (StatsResponse);
}

message StatsRequest {
  string hostname = 1;
  double cpu = 2;
  double ram = 3;
  int64 timestamp = 4;
}
```

### 3.2. Infrastructure Services

```
pbtypes/
├── Infrastructure/
│   ├── machines/      # Physical/Virtual machine info
│   ├── containers/    # Docker container monitoring
│   ├── servers/       # Server configuration
│   ├── resources/     # Resource allocation
│   └── storage/       # Storage management
```

### 3.3. System Services

```
pbtypes/
├── system/    # OS and system information
├── process/   # Process monitoring
├── network/   # Network metrics
├── disk/      # Disk usage and I/O
├── logs/      # Log aggregation
├── security/  # Security monitoring
└── user/      # User management
```

## 4. Design Patterns

### 4.1. Streaming Pattern

```go
// Agent sends continuous stream
for {
    stats := collectMetrics()
    stream.Send(stats)
    time.Sleep(interval)
}
```

### 4.2. Server-Side Processing

```go
// Backend receives and processes
for {
    req, err := stream.Recv()
    // Process metrics
    processMetrics(req)
    // Store to database
    store.Save(req)
    // Trigger alerts if needed
    checkAlerts(req)
}
```

### 4.3. Gateway Pattern

```
gRPC Proto Definitions
        ↓
Auto-generate REST API
        ↓
Swagger Documentation
```

## 5. Scalability & Performance

### 5.1. Horizontal Scaling

```
┌────────┐     ┌────────┐     ┌────────┐
│Backend │     │Backend │     │Backend │
│   1    │     │   2    │     │   3    │
└────┬───┘     └────┬───┘     └────┬───┘
     │              │              │
     └──────────────┴──────────────┘
                    ▲
             Load Balancer
                    ▲
                    │
              ┌─────┴─────┐
              │  Agents   │
              └───────────┘
```

### 5.2. Data Flow Optimization

- **Batching**: Gộp nhiều metrics trong một request
- **Compression**: Nén dữ liệu trước khi gửi
- **Buffering**: Buffer metrics khi network unstable
- **Sampling**: Thu thập metrics với tần suất hợp lý

### 5.3. Storage Strategy

```
Hot Data (Recent)     → In-Memory Cache (Redis)
Warm Data (24h-7d)    → Fast Database (PostgreSQL)
Cold Data (>7d)       → Time-series DB (InfluxDB/TimescaleDB)
Archive Data (>30d)   → Object Storage (S3/MinIO)
```

## 6. Security Architecture

### 6.1. Authentication & Authorization

```
Agent ──[TLS + Token]──▶ Backend ──[JWT]──▶ Frontend
```

**Layers:**
1. TLS encryption cho gRPC
2. Token-based authentication
3. Role-based access control (RBAC)
4. API rate limiting

### 6.2. Network Security

- Private network cho agent-backend communication
- Public API với authentication
- Firewall rules
- DDoS protection

## 7. Monitoring & Observability

### 7.1. Self-Monitoring

Hệ thống tự monitor chính nó:
- Backend health checks
- Agent connectivity status
- API response times
- Error rates and logs

### 7.2. Metrics

- Request latency (p50, p95, p99)
- Throughput (requests/second)
- Error rate
- Active connections
- Resource usage

### 7.3. Logging

```
Agent   → Structured logs → Log aggregation
Backend → Structured logs → Log aggregation → Analysis
```

## 8. Future Enhancements

### Phase 1 (Current)
- ✅ Basic monitoring (CPU, RAM)
- ✅ gRPC streaming
- ✅ REST API gateway

### Phase 2 (Next)
- 🔄 Full metrics support (Disk, Network, Process)
- 🔄 Database persistence
- 🔄 Basic dashboard

### Phase 3 (Future)
- 📋 Advanced alerting
- 📋 Historical analysis
- 📋 Predictive monitoring
- 📋 Machine learning insights

### Phase 4 (Advanced)
- 📋 Multi-tenant architecture
- 📋 Plugin system
- 📋 Custom metrics
- 📋 Integration với external systems (Prometheus, Grafana)

## 9. Technology Stack Summary

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Language | Go 1.24 | Backend & Agent |
| RPC | gRPC | Service communication |
| Serialization | Protocol Buffers | Data format |
| API Gateway | grpc-gateway | REST API |
| Documentation | Swagger/OpenAPI | API docs |
| Monitoring Library | gopsutil | System metrics |
| Frontend | TBD | Web dashboard |
| Database | TBD | Data persistence |
| Cache | TBD | Performance |
| Message Queue | TBD | Async processing |

## 10. Development Principles

1. **Modularity**: Mỗi service độc lập, dễ maintain
2. **Scalability**: Thiết kế để scale horizontal
3. **Reliability**: Fault tolerance và retry mechanisms
4. **Performance**: Optimize cho low latency và high throughput
5. **Security**: Security-first approach
6. **Observability**: Easy to debug và monitor
7. **Documentation**: Well-documented code và APIs
