# API Documentation - Smart Monitor

## Tổng quan

Smart Monitor cung cấp 2 loại API:
1. **gRPC API** - Cho communication giữa Agent và Backend
2. **REST API** - Được generate tự động từ gRPC qua grpc-gateway

## 1. gRPC Services

### 1.1. Monitor Service

Service chính để thu thập và quản lý metrics từ các agents.

**Package**: `monitor`  
**Proto file**: `pbtypes/monitor/monitor.proto`

#### Methods

##### 1.1.1. StreamStats (Bidirectional Streaming)

Thu thập metrics realtime từ agent qua gRPC streaming.

**Request Stream**: `StatsRequest`
```protobuf
message StatsRequest {
  string hostname = 1;  // Tên server/hostname
  double cpu = 2;       // CPU usage (0-100)
  double ram = 3;       // RAM usage (0-100)
  double disk = 4;      // Disk usage (0-100)
}
```

**Response**: `StatsResponse`
```protobuf
message StatsResponse {
  string message = 1;    // Status message
  int64 timestamp = 2;   // Unix timestamp
}
```

**Example (Go Client):**
```go
stream, err := client.StreamStats(context.Background())
if err != nil {
    log.Fatal(err)
}

// Send stats continuously
ticker := time.NewTicker(2 * time.Second)
for range ticker.C {
    stats := &pb.StatsRequest{
        Hostname: "server-01",
        Cpu:      getCPU(),
        Ram:      getRAM(),
        Disk:     getDisk(),
    }
    
    if err := stream.Send(stats); err != nil {
        log.Printf("Error sending stats: %v", err)
    }
}
```

**Example (grpcurl):**
```bash
grpcurl -plaintext -d @ localhost:50051 monitor.MonitorService/StreamStats <<EOM
{
  "hostname": "test-server",
  "cpu": 45.5,
  "ram": 60.2,
  "disk": 75.0
}
EOM
```

##### 1.1.2. GetStats (Unary)

Lấy stats hiện tại của một server cụ thể.

**Request**: `StatsRequest`
```protobuf
message StatsRequest {
  string hostname = 1;
  double cpu = 2;
  double ram = 3;
  double disk = 4;
}
```

**Response**: `StatsResponse`
```protobuf
message StatsResponse {
  string message = 1;
  int64 timestamp = 2;
}
```

**Example (Go Client):**
```go
req := &pb.StatsRequest{
    Hostname: "server-01",
}

resp, err := client.GetStats(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s (timestamp: %d)\n", resp.Message, resp.Timestamp)
```

**Example (grpcurl):**
```bash
grpcurl -plaintext -d '{"hostname":"server-01"}' \
  localhost:50051 monitor.MonitorService/GetStats
```

---

## 2. REST API (HTTP Gateway)

REST API được tự động generate từ gRPC definitions qua grpc-gateway.

**Base URL**: `http://localhost:8080`

### 2.1. Monitor Endpoints

#### 2.1.1. POST /v1/stats/stream

Stream stats từ agent (HTTP streaming).

**Request Body:**
```json
{
  "hostname": "server-01",
  "cpu": 45.5,
  "ram": 60.2,
  "disk": 75.0
}
```

**Response:**
```json
{
  "message": "Stats received",
  "timestamp": 1705334400
}
```

**Example (curl):**
```bash
curl -X POST http://localhost:8080/v1/stats/stream \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server-01",
    "cpu": 45.5,
    "ram": 60.2,
    "disk": 75.0
  }'
```

**Example (httpie):**
```bash
http POST localhost:8080/v1/stats/stream \
  hostname=server-01 \
  cpu:=45.5 \
  ram:=60.2 \
  disk:=75.0
```

#### 2.1.2. GET /v1/stats

Lấy stats hiện tại.

**Query Parameters:**
- `hostname` (string, optional) - Filter by hostname

**Response:**
```json
{
  "message": "Stats retrieved",
  "timestamp": 1705334400
}
```

**Example (curl):**
```bash
curl http://localhost:8080/v1/stats?hostname=server-01
```

**Example (httpie):**
```bash
http GET localhost:8080/v1/stats hostname==server-01
```

---

## 3. Infrastructure Services

### 3.1. Machine Service

**Proto**: `pbtypes/Infrastructure/machines/machine.proto`

Quản lý thông tin về physical/virtual machines.

**Example Messages:**
```protobuf
message MachineInfo {
  string machine_id = 1;
  string hostname = 2;
  string ip_address = 3;
  string os_type = 4;
  string os_version = 5;
  int32 cpu_cores = 6;
  int64 total_memory = 7;
  int64 total_disk = 8;
}
```

### 3.2. Container Service

**Proto**: `pbtypes/Infrastructure/containers/container.proto`

Giám sát Docker containers và Kubernetes pods.

**Example Messages:**
```protobuf
message ContainerInfo {
  string container_id = 1;
  string name = 2;
  string image = 3;
  string status = 4;
  double cpu_usage = 5;
  int64 memory_usage = 6;
}
```

### 3.3. Server Service

**Proto**: `pbtypes/Infrastructure/servers/server.proto`

Quản lý cấu hình servers.

### 3.4. Resource Service

**Proto**: `pbtypes/Infrastructure/resources/resource.proto`

Quản lý resource allocation và quotas.

---

## 4. System Services

### 4.1. System Service

**Proto**: `pbtypes/system/system.proto`

Thông tin về hệ điều hành và system info.

**Example Methods:**
```protobuf
service SystemService {
  rpc GetSystemInfo(SystemInfoRequest) returns (SystemInfoResponse);
  rpc GetUptime(UptimeRequest) returns (UptimeResponse);
}
```

### 4.2. Process Service

**Proto**: `pbtypes/process/process.proto`

Giám sát processes đang chạy.

**Example Methods:**
```protobuf
service ProcessService {
  rpc ListProcesses(ListProcessesRequest) returns (ListProcessesResponse);
  rpc GetProcessInfo(ProcessInfoRequest) returns (ProcessInfoResponse);
  rpc KillProcess(KillProcessRequest) returns (KillProcessResponse);
}
```

### 4.3. Network Service

**Proto**: `pbtypes/network/network.proto`

Monitor network traffic và connections.

**Example Messages:**
```protobuf
message NetworkStats {
  string interface = 1;
  int64 bytes_sent = 2;
  int64 bytes_recv = 3;
  int64 packets_sent = 4;
  int64 packets_recv = 5;
  double bandwidth_usage = 6;
}
```

### 4.4. Disk Service

**Proto**: `pbtypes/disk/disk.proto`

Monitor disk usage và I/O operations.

**Example Messages:**
```protobuf
message DiskStats {
  string device = 1;
  string mount_point = 2;
  int64 total = 3;
  int64 used = 4;
  int64 free = 5;
  double usage_percent = 6;
  int64 read_bytes = 7;
  int64 write_bytes = 8;
}
```

### 4.5. Security Service

**Proto**: `pbtypes/security/security.proto`

Security monitoring và audit logs.

**Example Methods:**
```protobuf
service SecurityService {
  rpc GetSecurityEvents(SecurityEventsRequest) returns (SecurityEventsResponse);
  rpc CheckVulnerabilities(VulnerabilityRequest) returns (VulnerabilityResponse);
}
```

### 4.6. User Service

**Proto**: `pbtypes/user/user.proto`

User management và authentication.

**Example Methods:**
```protobuf
service UserService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc GetUserInfo(UserInfoRequest) returns (UserInfoResponse);
}
```

### 4.7. Log Service

**Proto**: `pbtypes/logs/log.proto`

Log aggregation và analysis.

**Example Methods:**
```protobuf
service LogService {
  rpc StreamLogs(stream LogRequest) returns (LogResponse);
  rpc QueryLogs(LogQueryRequest) returns (LogQueryResponse);
}
```

---

## 5. Error Handling

### 5.1. gRPC Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 0 | OK | Success |
| 1 | CANCELLED | Operation was cancelled |
| 2 | UNKNOWN | Unknown error |
| 3 | INVALID_ARGUMENT | Invalid request |
| 4 | DEADLINE_EXCEEDED | Timeout |
| 5 | NOT_FOUND | Resource not found |
| 7 | PERMISSION_DENIED | Permission denied |
| 14 | UNAVAILABLE | Service unavailable |
| 16 | UNAUTHENTICATED | Authentication required |

### 5.2. Error Response Format

**gRPC Error:**
```go
return nil, status.Errorf(codes.InvalidArgument, "hostname is required")
```

**REST API Error:**
```json
{
  "error": "hostname is required",
  "code": 3,
  "message": "INVALID_ARGUMENT"
}
```

---

## 6. Authentication & Authorization

### 6.1. gRPC Authentication

**Using Metadata:**
```go
md := metadata.Pairs(
    "authorization", "Bearer YOUR_TOKEN_HERE",
)
ctx := metadata.NewOutgoingContext(context.Background(), md)

resp, err := client.GetStats(ctx, req)
```

### 6.2. REST Authentication

**Using Headers:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8080/v1/stats
```

---

## 7. Rate Limiting

### 7.1. Default Limits

- **gRPC**: 1000 requests/minute per client
- **REST API**: 100 requests/minute per IP

### 7.2. Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705334400
```

---

## 8. Pagination

### 8.1. Request Format

```bash
curl "http://localhost:8080/v1/logs?page=1&page_size=50"
```

### 8.2. Response Format

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_items": 1000,
    "total_pages": 20
  }
}
```

---

## 9. Webhooks (Future)

### 9.1. Event Types

- `server.online` - Server comes online
- `server.offline` - Server goes offline
- `alert.critical` - Critical alert triggered
- `alert.warning` - Warning alert triggered

### 9.2. Webhook Payload

```json
{
  "event": "alert.critical",
  "timestamp": 1705334400,
  "data": {
    "hostname": "server-01",
    "metric": "cpu",
    "value": 95.5,
    "threshold": 90.0
  }
}
```

---

## 10. API Versioning

### 10.1. Version in URL

```
/v1/stats          # Version 1
/v2/stats          # Version 2 (future)
```

### 10.2. Version in Header

```bash
curl -H "API-Version: v1" http://localhost:8080/stats
```

---

## 11. Swagger/OpenAPI Documentation

### 11.1. Access Swagger UI

**URL**: `http://localhost:8080/swagger/`

### 11.2. OpenAPI Spec

**Download**: `http://localhost:8080/swagger/combined.swagger.json`

### 11.3. Import to Postman

```bash
# Download spec
curl http://localhost:8080/swagger/combined.swagger.json > smart-monitor-api.json

# Import to Postman
# File > Import > Upload Files > smart-monitor-api.json
```

---

## 12. Client SDKs

### 12.1. Go Client

```go
import (
    pb "smart-monitor/pbtypes/monitor"
    "google.golang.org/grpc"
)

// Create connection
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Create client
client := pb.NewMonitorServiceClient(conn)

// Make request
resp, err := client.GetStats(context.Background(), &pb.StatsRequest{
    Hostname: "server-01",
})
```

### 12.2. Python Client (Example)

```python
import grpc
import monitor_pb2
import monitor_pb2_grpc

# Create channel
channel = grpc.insecure_channel('localhost:50051')

# Create stub
stub = monitor_pb2_grpc.MonitorServiceStub(channel)

# Make request
request = monitor_pb2.StatsRequest(hostname='server-01')
response = stub.GetStats(request)

print(f"Response: {response.message}")
```

### 12.3. JavaScript Client (Example)

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

// Load proto
const packageDefinition = protoLoader.loadSync('monitor.proto');
const proto = grpc.loadPackageDefinition(packageDefinition);

// Create client
const client = new proto.monitor.MonitorService(
  'localhost:50051',
  grpc.credentials.createInsecure()
);

// Make request
client.getStats({ hostname: 'server-01' }, (err, response) => {
  if (err) {
    console.error(err);
  } else {
    console.log('Response:', response.message);
  }
});
```

---

## 13. Testing APIs

### 13.1. Using grpcurl

```bash
# List all services
grpcurl -plaintext localhost:50051 list

# List methods in a service
grpcurl -plaintext localhost:50051 list monitor.MonitorService

# Describe a method
grpcurl -plaintext localhost:50051 describe monitor.MonitorService.GetStats

# Call a method
grpcurl -plaintext -d '{"hostname":"server-01"}' \
  localhost:50051 monitor.MonitorService/GetStats
```

### 13.2. Using curl for REST API

```bash
# GET request
curl -X GET http://localhost:8080/v1/stats

# POST request
curl -X POST http://localhost:8080/v1/stats/stream \
  -H "Content-Type: application/json" \
  -d '{"hostname":"server-01","cpu":45.5,"ram":60.2,"disk":75.0}'

# With authentication
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/v1/stats
```

### 13.3. Using Postman

1. Import OpenAPI spec from `http://localhost:8080/swagger/combined.swagger.json`
2. Set base URL to `http://localhost:8080`
3. Configure authentication if needed
4. Test endpoints

---

## 14. Performance Considerations

### 14.1. Streaming Best Practices

- Keep connection alive với reasonable timeout
- Implement backpressure mechanism
- Buffer messages khi cần
- Handle reconnection gracefully

### 14.2. Batch Requests

Thay vì gửi nhiều requests riêng lẻ, batch chúng lại:

```protobuf
message BatchStatsRequest {
  repeated StatsRequest stats = 1;
}
```

### 14.3. Compression

Enable gRPC compression:

```go
// Client side
conn, err := grpc.Dial(
    "localhost:50051",
    grpc.WithCompressor(grpc.NewGZIPCompressor()),
)

// Server side
s := grpc.NewServer(
    grpc.RPCCompressor(grpc.NewGZIPCompressor()),
)
```

---

## 15. Monitoring API Usage

### 15.1. Metrics to Track

- Request count per endpoint
- Response times (p50, p95, p99)
- Error rates
- Active connections
- Data transfer volume

### 15.2. Prometheus Metrics

```
# Request count
smart_monitor_requests_total{method="GetStats",status="success"} 1234

# Response time
smart_monitor_request_duration_seconds{method="GetStats",quantile="0.95"} 0.05

# Active connections
smart_monitor_active_connections 42
```

---

## 16. Changelog

### Version 1.0.0 (Current)

- ✅ Basic monitoring (CPU, RAM, Disk)
- ✅ gRPC streaming
- ✅ REST API gateway
- ✅ Swagger documentation

### Version 1.1.0 (Planned)

- 🔄 Full system metrics
- 🔄 Process monitoring
- 🔄 Network monitoring
- 🔄 Authentication & authorization

### Version 2.0.0 (Future)

- 📋 Advanced alerting
- 📋 Historical data queries
- 📋 Custom metrics
- 📋 Webhook notifications

---

## 17. Support & Resources

### 17.1. Documentation

- [Architecture Guide](ARCHITECTURE.md)
- [Development Guide](DEVELOPMENT.md)
- [Deployment Guide](DEPLOYMENT.md)

### 17.2. API Playground

- Swagger UI: `http://localhost:8080/swagger/`
- gRPC Reflection: `grpcurl -plaintext localhost:50051 list`

### 17.3. Getting Help

- GitHub Issues: [Create an issue]
- Email: api-support@smart-monitor.com
- Slack: #api-questions

---

## Appendix: Complete API Reference

### Monitor Service (v1)

| Method | Type | Endpoint | Description |
|--------|------|----------|-------------|
| StreamStats | Streaming | POST /v1/stats/stream | Stream metrics from agent |
| GetStats | Unary | GET /v1/stats | Get current stats |

### System Service (Future)

| Method | Type | Endpoint | Description |
|--------|------|----------|-------------|
| GetSystemInfo | Unary | GET /v1/system/info | Get system information |
| GetUptime | Unary | GET /v1/system/uptime | Get system uptime |

### Process Service (Future)

| Method | Type | Endpoint | Description |
|--------|------|----------|-------------|
| ListProcesses | Unary | GET /v1/processes | List all processes |
| GetProcessInfo | Unary | GET /v1/processes/{pid} | Get process details |
| KillProcess | Unary | DELETE /v1/processes/{pid} | Kill a process |

---

**Last Updated**: January 15, 2026  
**API Version**: v1.0.0
