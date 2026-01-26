# Cấu trúc hạ tầng Smart Monitor

## 1. Tổng quan

Tài liệu này mô tả chi tiết cấu trúc hạ tầng của hệ thống Smart Monitor, bao gồm các services, components và cách chúng tương tác với nhau.

## 2. Component Hierarchy

```
Smart Monitor System
│
├── Agent Layer (Data Collection)
│   ├── System Metrics Collector
│   ├── Process Monitor
│   ├── Network Monitor
│   ├── Disk Monitor
│   └── Container Monitor
│
├── Backend Layer (Processing & API)
│   ├── gRPC Server
│   ├── HTTP Gateway
│   ├── Business Logic
│   ├── Data Processor
│   └── Alert Engine
│
├── Data Layer (Persistence)
│   ├── Time-series Database
│   ├── Relational Database
│   ├── Cache (Redis)
│   └── Message Queue
│
└── Frontend Layer (Visualization)
    ├── Web Dashboard
    ├── Real-time Charts
    ├── Alert Dashboard
    └── Admin Console
```

## 3. Protocol Buffers Structure

### 3.1. Directory Layout

```
pbtypes/
├── combined.swagger.json        # Combined OpenAPI spec
├── generate_proto.sh           # Proto generation script
├── makefile                    # Build automation
├── run_makefile.sh            # Makefile runner
│
├── monitor/                    # Core monitoring service
│   ├── monitor.proto          # Service definition
│   ├── monitor.pb.go          # Generated Go code
│   ├── monitor_grpc.pb.go     # gRPC server/client
│   ├── monitor.pb.gw.go       # Gateway reverse proxy
│   └── monitor.swagger.json   # OpenAPI spec
│
├── system/                     # System information
│   ├── system.proto
│   ├── system.pb.go
│   ├── system_grpc.pb.go
│   ├── system.pb.gw.go
│   └── system.swagger.json
│
├── process/                    # Process monitoring
│   ├── process.proto
│   ├── process.pb.go
│   ├── process_grpc.pb.go
│   ├── process.pb.gw.go
│   └── process.swagger.json
│
├── network/                    # Network monitoring
│   ├── network.proto
│   ├── network.pb.go
│   ├── network_grpc.pb.go
│   ├── network.pb.gw.go
│   └── network.swagger.json
│
├── disk/                       # Disk monitoring
│   ├── disk.proto
│   ├── disk.pb.go
│   ├── disk_grpc.pb.go
│   ├── disk.pb.gw.go
│   └── disk.swagger.json
│   └── drivers/
│       └── vga/               # Display drivers (future)
│
├── screen/                     # Screen/Display monitoring
│   ├── screen.proto
│   ├── screen.pb.go
│   ├── screen_grpc.pb.go
│   ├── screen.pb.gw.go
│   └── screen.swagger.json
│
├── logs/                       # Log aggregation
│   ├── log.proto
│   ├── log.pb.go
│   ├── log_grpc.pb.go
│   ├── log.pb.gw.go
│   └── log.swagger.json
│
├── security/                   # Security monitoring
│   ├── security.proto
│   ├── security.pb.go
│   ├── security_grpc.pb.go
│   ├── security.pb.gw.go
│   └── security.swagger.json
│
├── user/                       # User management
│   ├── user.proto
│   ├── user.pb.go
│   ├── user_grpc.pb.go
│   ├── user.pb.gw.go
│   └── user.swagger.json
│
└── Infrastructure/             # Infrastructure management
    ├── machines/              # Machine information
    │   ├── machine.proto
    │   ├── machine.pb.go
    │   ├── machine_grpc.pb.go
    │   ├── machine.pb.gw.go
    │   └── machine.swagger.json
    │
    ├── containers/            # Container monitoring
    │   ├── container.proto
    │   ├── container.pb.go
    │   ├── container_grpc.pb.go
    │   ├── container.pb.gw.go
    │   └── container.swagger.json
    │
    ├── servers/               # Server management
    │   ├── server.proto
    │   ├── server.pb.go
    │   ├── server_grpc.pb.go
    │   ├── server.pb.gw.go
    │   └── server.swagger.json
    │
    ├── resources/             # Resource allocation
    │   ├── resource.proto
    │   ├── resource.pb.go
    │   ├── resource_grpc.pb.go
    │   ├── resource.pb.gw.go
    │   └── resource.swagger.json
    │
    └── storage/               # Storage management
        └── (proto files)
```

## 4. Service Definitions

### 4.1. Core Services

#### Monitor Service
**Purpose**: Thu thập và quản lý metrics realtime từ agents

**Key Features**:
- Bidirectional streaming cho real-time data
- Unary calls cho queries
- Auto-scaling support

**Use Cases**:
- Continuous metrics collection
- Real-time monitoring
- Alert triggering

#### System Service
**Purpose**: Cung cấp thông tin về hệ điều hành

**Key Features**:
- OS information
- System uptime
- Hardware specs
- Kernel version

**Use Cases**:
- System inventory
- Compatibility checking
- Resource planning

#### Process Service
**Purpose**: Giám sát và quản lý processes

**Key Features**:
- List running processes
- Process details (CPU, Memory usage)
- Process lifecycle management
- Parent-child relationships

**Use Cases**:
- Process monitoring
- Resource troubleshooting
- Process management

#### Network Service
**Purpose**: Monitor network metrics và connections

**Key Features**:
- Interface statistics
- Connection tracking
- Bandwidth monitoring
- Packet analysis

**Use Cases**:
- Network performance monitoring
- Traffic analysis
- Bandwidth optimization

#### Disk Service
**Purpose**: Monitor disk usage và I/O

**Key Features**:
- Disk space monitoring
- I/O statistics
- Mount point tracking
- SMART data

**Use Cases**:
- Capacity planning
- Performance optimization
- Disk health monitoring

### 4.2. Infrastructure Services

#### Machine Service
**Purpose**: Quản lý thông tin physical/virtual machines

**Key Features**:
- Machine inventory
- Hardware specifications
- Location tracking
- Status monitoring

**Use Cases**:
- Asset management
- Capacity planning
- Infrastructure overview

#### Container Service
**Purpose**: Monitor Docker containers và K8s pods

**Key Features**:
- Container lifecycle
- Resource usage per container
- Image management
- Network mapping

**Use Cases**:
- Container orchestration
- Resource optimization
- Microservices monitoring

#### Server Service
**Purpose**: Quản lý server configurations

**Key Features**:
- Server profiles
- Configuration management
- Service dependencies
- Health checks

**Use Cases**:
- Server management
- Configuration tracking
- Dependency mapping

#### Resource Service
**Purpose**: Quản lý resource allocation và quotas

**Key Features**:
- Resource pools
- Quota management
- Allocation tracking
- Over-subscription monitoring

**Use Cases**:
- Resource planning
- Cost optimization
- Multi-tenancy

### 4.3. Supporting Services

#### Log Service
**Purpose**: Aggregate và analyze logs

**Key Features**:
- Log streaming
- Log search and filter
- Log correlation
- Log retention

**Use Cases**:
- Troubleshooting
- Audit trails
- Security analysis

#### Security Service
**Purpose**: Security monitoring và compliance

**Key Features**:
- Security events
- Vulnerability scanning
- Compliance checking
- Access control

**Use Cases**:
- Security monitoring
- Compliance reporting
- Threat detection

#### User Service
**Purpose**: User authentication và authorization

**Key Features**:
- User management
- Role-based access control
- Session management
- API key management

**Use Cases**:
- User authentication
- Access control
- Audit logging

## 5. Data Flow Patterns

### 5.1. Metrics Collection Flow

```
Agent                 Backend              Database
  │                      │                     │
  │──[1] Collect────▶    │                     │
  │    Metrics           │                     │
  │                      │                     │
  │──[2] Stream────▶     │                     │
  │    via gRPC          │                     │
  │                      │                     │
  │                      │──[3] Validate──▶    │
  │                      │    & Transform      │
  │                      │                     │
  │                      │──[4] Store─────▶    │
  │                      │                     │
  │                      │──[5] Process───▶    │
  │                      │    Alerts           │
  │                      │                     │
  │◀────[6] ACK──────    │                     │
```

### 5.2. Query Flow

```
Frontend             Gateway              Backend             Database
   │                    │                    │                   │
   │──[1] HTTP GET──▶   │                    │                   │
   │                    │                    │                   │
   │                    │──[2] gRPC Call─▶   │                   │
   │                    │                    │                   │
   │                    │                    │──[3] Query────▶   │
   │                    │                    │                   │
   │                    │                    │◀──[4] Results──   │
   │                    │                    │                   │
   │                    │◀──[5] Response──   │                   │
   │                    │                    │                   │
   │◀────[6] JSON───    │                    │                   │
```

### 5.3. Alert Flow

```
Metrics         Alert Engine      Notification      User
   │                 │                 │              │
   │──[1] Stream─▶   │                 │              │
   │                 │                 │              │
   │                 │──[2] Check──▶   │              │
   │                 │    Rules        │              │
   │                 │                 │              │
   │                 │──[3] Trigger─▶  │              │
   │                 │    Alert        │              │
   │                 │                 │              │
   │                 │                 │──[4] Send─▶  │
   │                 │                 │   (Email/    │
   │                 │                 │    Slack)    │
```

## 6. Network Architecture

### 6.1. Port Allocation

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Backend gRPC | 50051 | gRPC | Agent communication |
| Backend HTTP | 8080 | HTTP | REST API & Swagger |
| Database | 5432 | PostgreSQL | Data persistence |
| Redis | 6379 | Redis | Caching |
| Message Queue | 5672 | AMQP | Async messaging |
| Metrics | 9090 | HTTP | Prometheus metrics |
| Frontend | 3000 | HTTP | Web UI |

### 6.2. Network Segmentation

```
┌─────────────────────────────────────────────────────────┐
│                    Public Network                       │
│  ┌──────────────┐                                       │
│  │  Frontend    │  Port 3000 (HTTPS)                   │
│  │  Load        │                                       │
│  │  Balancer    │                                       │
│  └──────┬───────┘                                       │
└─────────┼─────────────────────────────────────────────┘
          │
┌─────────┼─────────────────────────────────────────────┐
│         │         Application Network                  │
│  ┌──────▼───────┐      ┌──────────────┐              │
│  │  Frontend    │      │  Backend     │              │
│  │  Servers     │─────▶│  Servers     │              │
│  └──────────────┘      └──────┬───────┘              │
└────────────────────────────────┼──────────────────────┘
                                 │
┌────────────────────────────────┼──────────────────────┐
│                                │  Data Network         │
│  ┌──────────────┐      ┌──────▼───────┐              │
│  │  Database    │      │  Redis       │              │
│  │  Primary     │      │  Cache       │              │
│  └──────────────┘      └──────────────┘              │
└─────────────────────────────────────────────────────┘
          ▲
          │ gRPC (TLS)
          │
┌─────────┴─────────────────────────────────────────────┐
│                    Agent Network                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ Agent 1  │  │ Agent 2  │  │ Agent N  │           │
│  └──────────┘  └──────────┘  └──────────┘           │
└─────────────────────────────────────────────────────┘
```

## 7. Scaling Architecture

### 7.1. Horizontal Scaling

```
              ┌────────────────┐
              │ Load Balancer  │
              │  (Nginx/HAProxy│
              └────────┬───────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
   ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
   │Backend 1│    │Backend 2│    │Backend N│
   └────┬────┘    └────┬────┘    └────┬────┘
        │              │              │
        └──────────────┼──────────────┘
                       │
              ┌────────▼───────┐
              │  Shared Cache  │
              │    (Redis)     │
              └────────┬───────┘
                       │
              ┌────────▼───────┐
              │   Database     │
              │   Cluster      │
              └────────────────┘
```

### 7.2. Database Sharding

```
┌─────────────────────────────────────┐
│         Sharding Logic              │
│   (Based on hostname hash)          │
└─────┬──────────┬──────────┬─────────┘
      │          │          │
   ┌──▼──┐    ┌──▼──┐    ┌──▼──┐
   │Shard│    │Shard│    │Shard│
   │  1  │    │  2  │    │  N  │
   └─────┘    └─────┘    └─────┘
   Host A-H   Host I-P   Host Q-Z
```

## 8. High Availability Setup

### 8.1. Backend HA

```
        ┌────────────────┐
        │   Keepalived   │
        │   VIP: x.x.x.1 │
        └───────┬────────┘
                │
        ┌───────┴────────┐
        │                │
   ┌────▼────┐      ┌────▼────┐
   │Backend 1│      │Backend 2│
   │ Active  │      │ Standby │
   └────┬────┘      └────┬────┘
        │                │
        └───────┬────────┘
                │
        ┌───────▼────────┐
        │   Database     │
        │   Primary      │
        └────────────────┘
```

### 8.2. Database HA

```
   ┌─────────────┐
   │  Primary    │
   │  (Write)    │
   └──────┬──────┘
          │
    ┌─────┴─────┐
    │           │
┌───▼───┐   ┌───▼───┐
│Replica│   │Replica│
│  (R1) │   │  (R2) │
│(Read) │   │(Read) │
└───────┘   └───────┘
```

## 9. Security Architecture

### 9.1. Security Layers

```
┌─────────────────────────────────────┐
│   Layer 1: Network Security         │
│   - Firewall rules                  │
│   - VPN/Private network             │
│   - IP whitelisting                 │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│   Layer 2: Transport Security       │
│   - TLS/SSL encryption              │
│   - Certificate validation          │
│   - mTLS for service-to-service     │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│   Layer 3: Application Security     │
│   - JWT authentication              │
│   - Role-based access control       │
│   - API rate limiting               │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│   Layer 4: Data Security            │
│   - Encryption at rest              │
│   - Data masking                    │
│   - Audit logging                   │
└─────────────────────────────────────┘
```

## 10. Monitoring & Observability

### 10.1. Metrics Collection

```
Application
    │
    ├─▶ Prometheus Metrics
    │       │
    │       ├─▶ Request count
    │       ├─▶ Response time
    │       ├─▶ Error rate
    │       └─▶ Resource usage
    │
    ├─▶ Structured Logs
    │       │
    │       └─▶ ELK Stack / Loki
    │
    └─▶ Distributed Tracing
            │
            └─▶ Jaeger / Zipkin
```

### 10.2. Health Check Endpoints

| Endpoint | Purpose | Status Codes |
|----------|---------|--------------|
| /health | Overall health | 200: OK, 503: Unhealthy |
| /ready | Ready to serve | 200: Ready, 503: Not ready |
| /live | Process alive | 200: Alive |
| /metrics | Prometheus metrics | 200: Metrics |

## 11. Disaster Recovery

### 11.1. Backup Strategy

- **Continuous**: Database WAL archiving
- **Daily**: Full database backup
- **Weekly**: Full system snapshot
- **Monthly**: Archive to cold storage

### 11.2. Recovery Time Objectives (RTO)

| Component | RTO | RPO |
|-----------|-----|-----|
| Backend | 5 minutes | 1 minute |
| Database | 15 minutes | 5 minutes |
| Cache | Immediate | N/A (can rebuild) |
| Frontend | 2 minutes | N/A (stateless) |

## 12. Capacity Planning

### 12.1. Resource Estimates

**Per 100 Agents:**
- Backend: 1 vCPU, 2GB RAM
- Database: 2 vCPU, 4GB RAM, 50GB disk
- Cache: 512MB RAM
- Network: 10 Mbps

**Per 1000 Agents:**
- Backend: 4 vCPUs, 8GB RAM (2-3 instances)
- Database: 4 vCPUs, 16GB RAM, 500GB disk
- Cache: 4GB RAM
- Network: 100 Mbps

## 13. Development vs Production

### 13.1. Environment Differences

| Feature | Development | Production |
|---------|------------|------------|
| TLS | Optional | Required |
| Authentication | Optional | Required |
| Load Balancer | No | Yes |
| HA Database | No | Yes |
| Monitoring | Basic | Full stack |
| Backups | Manual | Automated |
| Log Level | Debug | Info/Warn |

## 14. Service Dependencies

```
Frontend
  └─▶ Backend
       ├─▶ Database (Required)
       ├─▶ Redis (Required)
       ├─▶ Message Queue (Optional)
       └─▶ External APIs (Optional)

Agent
  └─▶ Backend (Required)
```

## 15. Future Enhancements

### 15.1. Planned Infrastructure

- **Service Mesh** (Istio/Linkerd)
- **API Gateway** (Kong/Tyk)
- **Event Bus** (Kafka)
- **Distributed Tracing** (Jaeger)
- **Feature Flags** (LaunchDarkly)
- **CDN** for frontend assets
- **Object Storage** (S3/MinIO)

### 15.2. Advanced Features

- Auto-scaling based on load
- Multi-region deployment
- Edge computing support
- Hybrid cloud deployment
- Kubernetes native deployment

---

**Document Version**: 1.0  
**Last Updated**: January 15, 2026  
**Maintained by**: Smart Monitor Team
