# Hướng Dẫn Cài Đặt và Quản Lý Agent

## Mục Lục
- [Tổng Quan Hệ Thống](#tổng-quan-hệ-thống)
- [Kiến Trúc Hệ Thống](#kiến-trúc-hệ-thống)
- [Cài Đặt Agent](#cài-đặt-agent)
- [Quy Trình Registration](#quy-trình-registration)
- [Luồng Hoạt Động](#luồng-hoạt-động)
- [Xử Lý Dữ Liệu Trên Backend](#xử-lý-dữ-liệu-trên-backend)
- [Giao Tiếp Agent-Backend](#giao-tiếp-agent-backend)
- [Quản Lý Agent](#quản-lý-agent)
- [Bảo Mật](#bảo-mật)
- [Troubleshooting](#troubleshooting)

---

## Tổng Quan Hệ Thống

Smart Monitor là hệ thống giám sát phân tán với kiến trúc agent-based:
- **Backend Center**: Hệ thống quản lý trung tâm nhận và xử lý dữ liệu từ các agents
- **Agents**: Các chương trình chạy trên máy cần giám sát, thu thập metrics và gửi về backend
- **Authentication**: Cơ chế xác thực đảm bảo chỉ agents được phép mới có thể kết nối

---

## Kiến Trúc Hệ Thống

```
┌─────────────────────────────────────────────────────────────┐
│                    BACKEND CENTER                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   gRPC       │  │    HTTP      │  │   Swagger    │     │
│  │   :50051     │  │    :8080     │  │   UI         │     │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘     │
│         │                  │                                │
│  ┌──────┴──────────────────┴───────┐                       │
│  │      gRPC/HTTP Handlers         │                       │
│  └──────────────┬───────────────────┘                       │
│                 │                                           │
│  ┌──────────────┴──────────────┐                           │
│  │     Use Cases Layer          │                           │
│  │  • MonitorUseCase            │                           │
│  └──────────────┬───────────────┘                           │
│                 │                                           │
│  ┌──────────────┴──────────────┐                           │
│  │    Domain Services           │                           │
│  │  • StatsService              │                           │
│  │  • AuthService    ◄─────────┼──── Authentication        │
│  └──────────────┬───────────────┘                           │
│                 │                                           │
│  ┌──────────────┴──────────────┐                           │
│  │      Repositories            │                           │
│  │  • StatsRepository           │                           │
│  │  • HostRepository            │                           │
│  │  • AgentRegistryRepository   │                           │
│  └──────────────┬───────────────┘                           │
│                 │                                           │
│  ┌──────────────┴──────────────┐                           │
│  │      Data Storage            │                           │
│  │  • In-Memory / Database      │                           │
│  └──────────────────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
           ▲                    ▲                    ▲
           │                    │                    │
     gRPC Stream           gRPC Stream          gRPC Stream
           │                    │                    │
┌──────────┴─────┐   ┌──────────┴─────┐   ┌─────────┴──────┐
│  Agent 1       │   │  Agent 2       │   │  Agent N       │
│  Server-01     │   │  Server-02     │   │  Server-N      │
│  192.168.1.10  │   │  192.168.1.20  │   │  192.168.1.N   │
└────────────────┘   └────────────────┘   └────────────────┘
```

---

## Cài Đặt Agent

### Bước 1: Chuẩn Bị

**Yêu cầu hệ thống:**
- Go 1.22+ (để build từ source)
- Linux/Windows/MacOS
- Network kết nối đến Backend Center
- Quyền đọc system metrics

**Download Agent:**
```bash
# Option 1: Build từ source
git clone <repository-url>
cd smart-monitor/agent
go build -o agent

# Option 2: Download binary compiled
wget https://releases.example.com/smart-monitor-agent-v1.0.0-linux-amd64
chmod +x smart-monitor-agent-v1.0.0-linux-amd64
```

### Bước 2: Cấu Hình

**Chỉnh sửa file `main.go` (nếu build từ source):**
```go
const (
    backendAddr   = "backend.example.com:50051"  // Địa chỉ Backend Center
    interval      = 2 * time.Second              // Tần suất gửi metrics
    agentVersion  = "1.0.0"                      // Version của agent
)

// Metadata - thông tin mô tả agent
metadata := map[string]string{
    "location":    "datacenter-01",              // Vị trí data center
    "environment": "production",                 // Môi trường (prod/staging/dev)
    "os":          "linux",                      // Hệ điều hành
    "team":        "infrastructure",             // Team quản lý
    "tier":        "web",                        // Tier (web/app/db)
}
```

**Rebuild nếu thay đổi config:**
```bash
go build -o agent
```

### Bước 3: Khởi Chạy Agent

**Chạy manual:**
```bash
./agent
```

**Chạy dưới dạng systemd service (Linux):**

Tạo file `/etc/systemd/system/smart-monitor-agent.service`:
```ini
[Unit]
Description=Smart Monitor Agent
After=network.target

[Service]
Type=simple
User=monitoring
WorkingDirectory=/opt/smart-monitor
ExecStart=/opt/smart-monitor/agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable và start service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable smart-monitor-agent
sudo systemctl start smart-monitor-agent
sudo systemctl status smart-monitor-agent
```

**Chạy dưới dạng Windows Service:**
```powershell
# Sử dụng NSSM (Non-Sucking Service Manager)
nssm install SmartMonitorAgent "C:\Program Files\SmartMonitor\agent.exe"
nssm start SmartMonitorAgent
```

---

## Quy Trình Registration

### Luồng Registration Chi Tiết

```
┌─────────┐                                     ┌──────────────┐
│  Agent  │                                     │   Backend    │
└────┬────┘                                     └──────┬───────┘
     │                                                 │
     │  1. Connect gRPC                                │
     ├────────────────────────────────────────────────►│
     │                                                 │
     │  2. RegisterAgent RPC                           │
     │     {                                           │
     │       hostname: "server-01",                    │
     │       ip_address: "192.168.1.10",               │
     │       agent_version: "1.0.0",                   │
     │       metadata: {...}                           │
     │     }                                           │
     ├────────────────────────────────────────────────►│
     │                                                 │
     │                              3. Validate Request│
     │                                                 ├──┐
     │                                                 │  │
     │                              4. Generate AgentID│  │
     │                                agent-a3f5c2d1   │  │
     │                                                 │  │
     │                              5. Generate Token  │  │
     │                                64-char hex      │  │
     │                                                 │  │
     │                              6. Save to Registry│  │
     │                                                 │◄─┘
     │                                                 │
     │  7. RegisterResponse                            │
     │     {                                           │
     │       success: true,                            │
     │       agent_id: "agent-a3f5c2d1",               │
     │       access_token: "3f4a8b...",                │
     │       expires_at: 1737849600                    │
     │     }                                           │
     │◄────────────────────────────────────────────────┤
     │                                                 │
     │  8. Save credentials to .agent_token            │
     ├──┐                                              │
     │  │                                              │
     │◄─┘                                              │
     │                                                 │
     │  9. Ready to stream metrics                     │
     │                                                 │
```

### Chi Tiết Các Bước

#### **Bước 1-2: Agent Khởi Tạo Kết Nối**
```go
// Agent code
conn, err := grpc.Dial(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewMonitorServiceClient(conn)

req := &pb.RegisterRequest{
    Hostname:     hostname,
    IpAddress:    ipAddress,
    AgentVersion: agentVersion,
    Metadata:     metadata,
}

resp, err := client.RegisterAgent(ctx, req)
```

#### **Bước 3-6: Backend Xử Lý Registration**
```go
// Backend AuthService
func (s *AuthService) RegisterAgent(ctx context.Context, hostname, ipAddress, agentVersion string, metadata map[string]string) (*entity.AgentRegistry, error) {
    // Generate unique AgentID
    agentID := generateAgentID(hostname, ipAddress)
    
    // Check if agent already registered
    existingAgent, err := s.agentRepo.GetByAgentID(ctx, agentID)
    if err == nil && existingAgent != nil {
        if existingAgent.IsValid() {
            // Return existing credentials
            return existingAgent, nil
        }
        // Renew if expired
        existingAgent.RenewToken()
        return existingAgent, nil
    }
    
    // Create new registry entry
    agent := entity.NewAgentRegistry(agentID, hostname, ipAddress, agentVersion, metadata)
    s.agentRepo.Register(ctx, agent)
    
    return agent, nil
}
```

#### **Bước 7-8: Agent Lưu Credentials**
```go
// Save to file .agent_token
credentials := &AgentCredentials{
    AgentID:     resp.AgentId,
    AccessToken: resp.AccessToken,
    ExpiresAt:   resp.ExpiresAt,
}

data, _ := json.MarshalIndent(credentials, "", "  ")
os.WriteFile(".agent_token", data, 0600)
```

**Nội dung file `.agent_token`:**
```json
{
  "AgentID": "agent-a3f5c2d1",
  "AccessToken": "3f4a8b2c1d9e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2",
  "ExpiresAt": 1737849600
}
```

---

## Luồng Hoạt Động

### Streaming Metrics

```
┌─────────┐                                     ┌──────────────┐
│  Agent  │                                     │   Backend    │
└────┬────┘                                     └──────┬───────┘
     │                                                 │
     │  1. Open StreamStats                            │
     ├────────────────────────────────────────────────►│
     │                                                 │
     │  ┌──────── Every 2 seconds ────────┐           │
     │  │                                  │           │
     │  │  2. Collect System Metrics       │           │
     │  │     • CPU Usage                  │           │
     │  │     • RAM Usage                  │           │
     │  │     • Disk Usage                 │           │
     │  └──────────────────────────────────┘           │
     │                                                 │
     │  3. Send StatsRequest                           │
     │     {                                           │
     │       agent_id: "agent-a3f5c2d1",               │
     │       access_token: "3f4a...",                  │
     │       hostname: "server-01",                    │
     │       ip_address: "192.168.1.10",               │
     │       cpu: 45.2,                                │
     │       ram: 68.5,                                │
     │       disk: 72.3                                │
     │     }                                           │
     ├────────────────────────────────────────────────►│
     │                                                 │
     │                              4. Validate Token  │
     │                                 Check AgentID   │
     │                                 Check Expiry    │
     │                                                 ├──┐
     │                                                 │  │
     │                              5. Process Stats   │  │
     │                                 • Validate Data │  │
     │                                 • Save to DB    │  │
     │                                 • Update Host   │  │
     │                                                 │◄─┘
     │                                                 │
     │  6. Continue streaming...                       │
     │                                                 │
     │  (Loop continues until connection closed)       │
     │                                                 │
```

### Code Agent - Thu Thập và Gửi Metrics

```go
// Collect system metrics
func collectStats(hostname, agentID, ipAddress, accessToken string, metadata map[string]string) (*pb.StatsRequest, error) {
    // CPU usage
    cpuPercent, _ := cpu.Percent(time.Second, false)
    cpuUsage := cpuPercent[0]
    
    // RAM usage
    memInfo, _ := mem.VirtualMemory()
    
    // Disk usage
    diskInfo, _ := disk.Usage("/")
    
    return &pb.StatsRequest{
        Hostname:     hostname,
        AgentId:      agentID,
        IpAddress:    ipAddress,
        AgentVersion: agentVersion,
        AccessToken:  accessToken,
        Cpu:          cpuUsage,
        Ram:          memInfo.UsedPercent,
        Disk:         diskInfo.UsedPercent,
        Metadata:     metadata,
    }, nil
}

// Stream to backend
stream, _ := client.StreamStats(ctx)
ticker := time.NewTicker(2 * time.Second)

for {
    select {
    case <-ticker.C:
        stats, _ := collectStats(...)
        stream.Send(stats)
    }
}
```

---

## Xử Lý Dữ Liệu Trên Backend

### Kiến Trúc DDD (Domain-Driven Design)

Backend sử dụng kiến trúc DDD với các layer:

```
┌───────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                    │
│  • gRPC Handlers      • HTTP Handlers                     │
│  • Repositories       • External Services                 │
└──────────────────┬────────────────────────────────────────┘
                   │
┌──────────────────┴────────────────────────────────────────┐
│                   Application Layer                        │
│  • Use Cases      • DTOs                                  │
│  • Orchestration  • Application Logic                     │
└──────────────────┬────────────────────────────────────────┘
                   │
┌──────────────────┴────────────────────────────────────────┐
│                     Domain Layer                           │
│  • Entities       • Value Objects                         │
│  • Services       • Business Rules                        │
│  • Repositories   • Domain Events                         │
└───────────────────────────────────────────────────────────┘
```

### Luồng Xử Lý Request

```
gRPC Request
     │
     ▼
┌─────────────────────┐
│  gRPC Handler       │  1. Nhận request từ agent
│  monitor_handler.go │  2. Validate access token
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Use Case           │  3. Convert DTO
│  MonitorUseCase     │  4. Call domain service
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Domain Service     │  5. Business logic
│  StatsService       │  6. Validate stats
│                     │  7. Make decisions:
│                     │     • Lưu vào DB?
│                     │     • Skip duplicate?
│                     │     • Alert threshold?
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Repository         │  8. Persist data
│  StatsRepository    │  9. Update host info
└─────────────────────┘
```

### Code Backend - Xử Lý Stats

#### **1. gRPC Handler (Infrastructure Layer)**
```go
func (s *MonitorServiceServer) StreamStats(stream pb.MonitorService_StreamStatsServer) error {
    for {
        req, err := stream.Recv()
        
        // Authenticate
        if err := s.authService.ValidateToken(ctx, req.AgentId, req.AccessToken); err != nil {
            log.Printf("Authentication failed: %v", err)
            continue
        }
        
        // Convert to DTO
        statsReq := &dto.StatsRequest{
            Hostname:  req.Hostname,
            AgentID:   req.AgentId,
            CPU:       req.Cpu,
            RAM:       req.Ram,
            Disk:      req.Disk,
        }
        
        // Process through use case
        s.monitorUseCase.RecordStats(ctx, statsReq)
    }
}
```

#### **2. Use Case (Application Layer)**
```go
func (uc *MonitorUseCase) RecordStats(ctx context.Context, req *dto.StatsRequest) error {
    // Convert DTO to domain entity
    stats := entity.NewStats(req.Hostname, req.AgentID, req.IPAddress, req.CPU, req.RAM, req.Disk)
    
    // Delegate to domain service
    return uc.statsService.ProcessStats(ctx, stats, req.AgentVersion)
}
```

#### **3. Domain Service (Domain Layer)**
```go
func (s *StatsService) ProcessStats(ctx context.Context, stats *entity.Stats, agentVersion string) error {
    // Business Rule 1: Validate data
    if !stats.IsValid() {
        return fmt.Errorf("invalid stats")
    }
    
    // Business Rule 2: Check thresholds
    if stats.CPU > 90 || stats.RAM > 90 {
        // Trigger alert (future feature)
        log.Printf("⚠ High resource usage on %s", stats.Hostname)
    }
    
    // Business Rule 3: Skip duplicate (nếu data không thay đổi)
    existingStats, err := s.statsRepo.Get(ctx, stats.Hostname)
    if err == nil {
        if isDataUnchanged(existingStats, stats) {
            log.Printf("Skipping duplicate data from %s", stats.Hostname)
            return nil
        }
    }
    
    // Save stats
    if err := s.statsRepo.Save(ctx, stats); err != nil {
        return err
    }
    
    // Update host information
    host, _ := s.hostRepo.Get(ctx, stats.AgentID)
    if host == nil {
        host = entity.NewHost(stats.Hostname, stats.IPAddress, stats.AgentID)
        host.AgentVersion = agentVersion
        s.hostRepo.Create(ctx, host)
    } else {
        host.MarkSeen()
        host.AgentVersion = agentVersion
        s.hostRepo.Update(ctx, host)
    }
    
    return nil
}
```

### Quyết Định Xử Lý

Backend có các logic quyết định:

**1. Lưu trữ (Save):**
- ✅ Data hợp lệ (CPU/RAM/Disk 0-100%)
- ✅ Agent authenticated
- ✅ Data có thay đổi so với lần trước
- ❌ Data không hợp lệ → reject
- ❌ Token expired/invalid → reject
- ❌ Data duplicate → skip

**2. Cập nhật (Update):**
- Host status: online/offline
- Last seen timestamp
- Agent version
- IP address (nếu thay đổi)

**3. Xử lý đặc biệt (Future):**
- Alert khi vượt threshold
- Aggregate data theo time window
- Anomaly detection
- Trend analysis

---

## Giao Tiếp Agent-Backend

### Protocol: gRPC

**Lợi ích của gRPC:**
- ✅ Binary protocol (nhanh hơn JSON)
- ✅ Bi-directional streaming
- ✅ Strong typing với Protocol Buffers
- ✅ Built-in authentication support
- ✅ Connection multiplexing

### Protocol Buffers Definition

```protobuf
service MonitorService {
  // Registration RPC
  rpc RegisterAgent (RegisterRequest) returns (RegisterResponse);
  
  // Streaming metrics RPC
  rpc StreamStats (stream StatsRequest) returns (StatsResponse);
  
  // Query stats RPC
  rpc GetStats (StatsRequest) returns (StatsResponse);
}

message RegisterRequest {
  string hostname = 1;
  string ip_address = 2;
  string agent_version = 3;
  map<string, string> metadata = 4;
}

message RegisterResponse {
  bool success = 1;
  string message = 2;
  string agent_id = 3;
  string access_token = 4;
  int64 expires_at = 5;
}

message StatsRequest {
  string hostname = 1;
  double cpu = 2;
  double ram = 3;
  double disk = 4;
  string agent_id = 5;
  string ip_address = 6;
  string agent_version = 7;
  map<string, string> metadata = 8;
  string access_token = 9;  // For authentication
}
```

### Network Flow

```
Agent ──────► Backend Center
     gRPC
     Port: 50051
     Protocol: HTTP/2
     Format: Protobuf Binary

     Connection: Persistent
     Stream: Bi-directional
     Compression: gzip
```

### Xử Lý Lỗi & Retry

**Agent-side:**
```go
func streamStats(ctx context.Context, client pb.MonitorServiceClient, ...) error {
    stream, err := client.StreamStats(ctx)
    if err != nil {
        // Retry with exponential backoff
        return retryWithBackoff(err)
    }
    
    for {
        err := stream.Send(stats)
        if err != nil {
            log.Printf("Send error: %v", err)
            // Reconnect
            return err
        }
    }
}
```

**Backend-side:**
```go
func (s *MonitorServiceServer) StreamStats(stream pb.MonitorService_StreamStatsServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            // Client closed connection gracefully
            return stream.SendAndClose(&pb.StatsResponse{...})
        }
        if err != nil {
            // Connection error
            log.Printf("Stream error: %v", err)
            return err
        }
        
        // Process request...
    }
}
```

---

## Quản Lý Agent

### Các Thao Tác Quản Lý

#### **1. Liệt Kê Tất Cả Agents**
```bash
# API endpoint (future)
curl http://backend:8080/v1/agents

# Response
{
  "agents": [
    {
      "agent_id": "agent-a3f5c2d1",
      "hostname": "server-01",
      "ip_address": "192.168.1.10",
      "status": "active",
      "last_seen": "2026-01-26T10:30:00Z",
      "version": "1.0.0"
    },
    ...
  ]
}
```

#### **2. Suspend Agent**
```go
// Backend service
authService.SuspendAgent(ctx, agentID)

// Agent sẽ không thể gửi metrics
// Authentication sẽ fail với status "suspended"
```

#### **3. Revoke Agent**
```go
// Backend service
authService.RevokeAgent(ctx, agentID)

// Agent bị thu hồi quyền vĩnh viễn
// Cần register lại để có token mới
```

#### **4. Renew Token**
```go
// Backend service
agent, _ := authService.GetAgentByID(ctx, agentID)
agent.RenewToken()
authService.UpdateAgent(ctx, agent)

// Agent cần re-register để lấy token mới
```

### Agent Status

```
┌──────────┐  Register   ┌──────────┐
│   New    ├────────────►│  Active  │
└──────────┘             └─────┬────┘
                               │
                    ┌──────────┼──────────┐
                    │                     │
              Suspend                  Revoke
                    │                     │
                    ▼                     ▼
              ┌──────────┐         ┌──────────┐
              │Suspended │         │ Revoked  │
              └─────┬────┘         └──────────┘
                    │
                Activate
                    │
                    ▼
              ┌──────────┐
              │  Active  │
              └──────────┘
```

---

## Bảo Mật

### 1. Token-Based Authentication

**Token Generation:**
```go
func generateAccessToken() string {
    bytes := make([]byte, 32)  // 256 bits
    rand.Read(bytes)
    return hex.EncodeToString(bytes)  // 64 characters hex
}
```

**Token Storage:**
- Agent: File `.agent_token` với permission `0600` (owner only)
- Backend: In-memory hoặc encrypted database

**Token Validation:**
```go
func (a *AgentRegistry) IsTokenValid(token string) bool {
    return a.AccessToken == token && 
           a.Status == AgentStatusActive && 
           time.Now().Before(a.TokenExpiry)
}
```

### 2. Access Control

- ✅ Agent phải register trước khi streaming
- ✅ Mỗi request phải kèm valid token
- ✅ Token có expiry time (default: 1 năm)
- ✅ Backend có thể suspend/revoke bất kỳ lúc nào

### 3. Network Security

**Recommendations:**
```
☑ Sử dụng TLS/SSL cho gRPC connection
☑ Firewall rules: chỉ cho phép agents connect đến backend
☑ VPN hoặc private network cho agent-backend communication
☑ Rate limiting để chống DoS
☑ IP whitelist cho registered agents
```

**Enable TLS (Production):**
```go
// Backend
creds, _ := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
grpcServer := grpc.NewServer(grpc.Creds(creds))

// Agent
creds, _ := credentials.NewClientTLSFromFile("ca-cert.pem", "")
conn, _ := grpc.Dial(backendAddr, grpc.WithTransportCredentials(creds))
```

### 4. Data Validation

Backend validate tất cả dữ liệu:
```go
func (s *Stats) IsValid() bool {
    if s.Hostname == "" || s.AgentID == "" {
        return false
    }
    if s.CPU < 0 || s.CPU > 100 {
        return false
    }
    if s.RAM < 0 || s.RAM > 100 {
        return false
    }
    if s.Disk < 0 || s.Disk > 100 {
        return false
    }
    return true
}
```

---

## Troubleshooting

### Lỗi Thường Gặp

#### **1. Agent không kết nối được Backend**

**Triệu chứng:**
```
Failed to connect to backend: connection refused
```

**Giải pháp:**
- ☑ Kiểm tra backend có đang chạy: `netstat -an | grep 50051`
- ☑ Kiểm tra firewall rules
- ☑ Kiểm tra địa chỉ backend trong config
- ☑ Test connectivity: `telnet backend-host 50051`

#### **2. Registration Failed**

**Triệu chứng:**
```
Failed to register agent: registration failed: hostname is required
```

**Giải pháp:**
- ☑ Kiểm tra hostname có được set đúng
- ☑ Kiểm tra network connectivity
- ☑ Xem logs backend để biết lý do cụ thể

#### **3. Authentication Failed**

**Triệu chứng:**
```
Authentication failed for agent agent-xxx: invalid or expired token
```

**Giải pháp:**
- ☑ Xóa file `.agent_token` và restart agent để re-register
- ☑ Kiểm tra token expiry time
- ☑ Kiểm tra agent status trên backend (có bị suspend/revoke?)

#### **4. Metrics Không Được Ghi Nhận**

**Triệu chứng:**
- Agent gửi metrics nhưng không thấy trên backend

**Giải pháp:**
- ☑ Kiểm tra backend logs có thấy metrics không
- ☑ Kiểm tra validation: data có hợp lệ không (0-100%)
- ☑ Kiểm tra duplicate skip logic
- ☑ Query backend API: `curl http://backend:8080/v1/stats`

#### **5. High Memory/CPU Usage**

**Triệu chứng:**
- Agent hoặc backend tiêu tốn quá nhiều resources

**Giải pháp Backend:**
- ☑ Giảm số lượng stats lưu trữ (implement retention policy)
- ☑ Sử dụng database thay vì in-memory
- ☑ Enable compression cho gRPC

**Giải pháp Agent:**
- ☑ Tăng interval time (từ 2s → 5s, 10s)
- ☑ Giảm số metrics thu thập

### Debug Mode

**Enable verbose logging:**

Agent:
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
// Thêm debug logs
log.Printf("DEBUG: Collecting stats...")
```

Backend:
```bash
# Set log level
export LOG_LEVEL=debug
./backend
```

### Health Check

**Backend health endpoint:**
```bash
curl http://backend:8080/health

# Response
{
  "status": "healthy",
  "timestamp": 1737849600,
  "service": "smart-monitor-backend"
}
```

**Agent health check:**
```bash
# Kiểm tra process
ps aux | grep agent

# Kiểm tra logs
tail -f /var/log/smart-monitor/agent.log

# Kiểm tra token file
cat .agent_token
```

---

## Best Practices

### 1. Deployment

- ✅ Sử dụng systemd/supervisor để quản lý agent process
- ✅ Enable auto-restart on failure
- ✅ Rotate logs để tránh đầy disk
- ✅ Monitor agent health

### 2. Monitoring

- ✅ Monitor connection status
- ✅ Track metrics delivery rate
- ✅ Alert khi agent offline > 5 phút
- ✅ Monitor token expiry

### 3. Maintenance

- ✅ Backup agent credentials
- ✅ Document agent locations và metadata
- ✅ Regular security audits
- ✅ Update agents to latest version

### 4. Scaling

**Horizontal Scaling:**
- Backend có thể chạy multiple instances với load balancer
- Agents kết nối đến load balancer endpoint

**Vertical Scaling:**
- Tăng resources cho backend server
- Optimize database queries
- Enable caching

---

## Tổng Kết

Hệ thống Smart Monitor cung cấp:

✅ **Agent tự động hóa**: Install → Register → Monitor  
✅ **Bảo mật**: Token-based authentication  
✅ **Mở rộng**: Hỗ trợ hàng ngàn agents  
✅ **Tin cậy**: Persistent connections, auto-retry  
✅ **Quản lý tập trung**: Backend xử lý tất cả logic  

**Workflow tổng quan:**
```
Install Agent → Register với Backend → Nhận Token → 
Stream Metrics → Backend Validate → Backend Process → 
Backend Store → Dashboard Display
```

Để có hỗ trợ hoặc báo lỗi, vui lòng tạo issue trên GitHub repository.
