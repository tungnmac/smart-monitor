# Hướng dẫn triển khai Smart Monitor

## 1. Tổng quan

Tài liệu này mô tả quy trình triển khai hệ thống Smart Monitor trong các môi trường khác nhau: Development, Staging và Production.

## 2. Yêu cầu hệ thống

### 2.1. Minimum Requirements

**Backend Server:**
- CPU: 2 cores
- RAM: 2GB
- Disk: 20GB SSD
- Network: 100Mbps
- OS: Linux (Ubuntu 20.04+, CentOS 8+)

**Agent:**
- CPU: 0.5 core
- RAM: 256MB
- Disk: 100MB
- OS: Linux, macOS, Windows

### 2.2. Recommended Production Requirements

**Backend Server:**
- CPU: 4+ cores
- RAM: 8GB+
- Disk: 100GB SSD
- Network: 1Gbps
- OS: Ubuntu 22.04 LTS

**Database (nếu sử dụng):**
- CPU: 4 cores
- RAM: 8GB
- Disk: 200GB SSD
- IOPS: 3000+

## 3. Deployment Architecture

### 3.1. Single Server (Development/Small Scale)

```
┌─────────────────────────────────┐
│      Single Server              │
│                                 │
│  ┌─────────┐  ┌──────────┐    │
│  │ Backend │  │ Database │    │
│  │  :50051 │  │  :5432   │    │
│  │  :8080  │  │          │    │
│  └─────────┘  └──────────┘    │
│                                 │
└─────────────────────────────────┘
         ▲
         │
    ┌────┴────┐
    │ Agents  │
    └─────────┘
```

### 3.2. Production (High Availability)

```
                    ┌──────────────┐
                    │ Load Balancer│
                    │   (Nginx)    │
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  Backend 1    │  │  Backend 2    │  │  Backend 3    │
│   :50051      │  │   :50051      │  │   :50051      │
└───────┬───────┘  └───────┬───────┘  └───────┬───────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
                  ┌────────────────┐
                  │   Database     │
                  │   (Primary)    │
                  └────────┬───────┘
                           │
                  ┌────────▼───────┐
                  │   Database     │
                  │   (Replica)    │
                  └────────────────┘
```

## 4. Build Process

### 4.1. Build Backend

```bash
cd backend

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o smart-monitor-backend main.go

# Build with optimizations
go build -ldflags="-s -w" -o smart-monitor-backend main.go

# Verify binary
./smart-monitor-backend --version
```

### 4.2. Build Agent

```bash
cd agent

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o smart-monitor-agent-linux main.go
GOOS=darwin GOARCH=amd64 go build -o smart-monitor-agent-mac main.go
GOOS=windows GOARCH=amd64 go build -o smart-monitor-agent.exe main.go

# Build with version info
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
go build -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" -o smart-monitor-agent main.go
```

### 4.3. Cross-compilation script

```bash
#!/bin/bash
# build.sh

VERSION=$(git describe --tags --always --dirty)
PLATFORMS=("linux/amd64" "darwin/amd64" "windows/amd64")

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="smart-monitor-agent-${GOOS}-${GOARCH}"
    
    if [ $GOOS = "windows" ]; then
        OUTPUT+='.exe'
    fi
    
    echo "Building $OUTPUT..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.Version=$VERSION" \
        -o build/$OUTPUT \
        main.go
done
```

## 5. Deployment Methods

### 5.1. Manual Deployment

**Backend:**

```bash
# 1. Copy binary to server
scp smart-monitor-backend user@server:/opt/smart-monitor/

# 2. SSH to server
ssh user@server

# 3. Create systemd service
sudo nano /etc/systemd/system/smart-monitor-backend.service
```

**Service file:**
```ini
[Unit]
Description=Smart Monitor Backend Service
After=network.target

[Service]
Type=simple
User=smart-monitor
Group=smart-monitor
WorkingDirectory=/opt/smart-monitor
ExecStart=/opt/smart-monitor/smart-monitor-backend
Restart=always
RestartSec=10

# Security
NoNewPrivileges=true
PrivateTmp=true

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Environment
Environment="GRPC_PORT=50051"
Environment="HTTP_PORT=8080"

[Install]
WantedBy=multi-user.target
```

**Start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable smart-monitor-backend
sudo systemctl start smart-monitor-backend
sudo systemctl status smart-monitor-backend
```

**Agent:**

```bash
# Copy agent to monitored servers
scp smart-monitor-agent user@target-server:/opt/smart-monitor/

# Create service on each monitored server
sudo nano /etc/systemd/system/smart-monitor-agent.service
```

**Agent service file:**
```ini
[Unit]
Description=Smart Monitor Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/smart-monitor/smart-monitor-agent --server=backend-server:50051
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 5.2. Docker Deployment

**Backend Dockerfile:**

```dockerfile
# backend/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backend main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary and static files
COPY --from=builder /app/backend .
COPY --from=builder /app/static ./static

EXPOSE 50051 8080

CMD ["./backend"]
```

**Agent Dockerfile:**

```dockerfile
# agent/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o agent main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/agent .

CMD ["./agent"]
```

**Docker Compose:**

```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: smart-monitor-backend
    ports:
      - "50051:50051"
      - "8080:8080"
    environment:
      - GRPC_PORT=50051
      - HTTP_PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
    depends_on:
      - postgres
    restart: unless-stopped
    networks:
      - smart-monitor

  postgres:
    image: postgres:15-alpine
    container_name: smart-monitor-db
    environment:
      - POSTGRES_DB=smart_monitor
      - POSTGRES_USER=monitor
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    networks:
      - smart-monitor

  agent:
    build:
      context: ./agent
      dockerfile: Dockerfile
    container_name: smart-monitor-agent
    environment:
      - BACKEND_SERVER=backend:50051
      - HOSTNAME=docker-agent
      - INTERVAL=2s
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - smart-monitor

volumes:
  postgres_data:

networks:
  smart-monitor:
    driver: bridge
```

**Deploy with Docker Compose:**

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop
docker-compose down

# Rebuild
docker-compose up -d --build
```

### 5.3. Kubernetes Deployment

**Backend Deployment:**

```yaml
# k8s/backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smart-monitor-backend
  namespace: monitoring
spec:
  replicas: 3
  selector:
    matchLabels:
      app: smart-monitor-backend
  template:
    metadata:
      labels:
        app: smart-monitor-backend
    spec:
      containers:
      - name: backend
        image: smart-monitor/backend:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 8080
          name: http
        env:
        - name: GRPC_PORT
          value: "50051"
        - name: HTTP_PORT
          value: "8080"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: smart-monitor-backend
  namespace: monitoring
spec:
  selector:
    app: smart-monitor-backend
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: http
    port: 8080
    targetPort: 8080
  type: LoadBalancer
```

**Agent DaemonSet:**

```yaml
# k8s/agent-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: smart-monitor-agent
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: smart-monitor-agent
  template:
    metadata:
      labels:
        app: smart-monitor-agent
    spec:
      hostNetwork: true
      hostPID: true
      containers:
      - name: agent
        image: smart-monitor/agent:latest
        env:
        - name: BACKEND_SERVER
          value: "smart-monitor-backend:50051"
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        securityContext:
          privileged: true
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
```

**Deploy to Kubernetes:**

```bash
# Create namespace
kubectl create namespace monitoring

# Apply configurations
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/agent-daemonset.yaml

# Check status
kubectl get pods -n monitoring
kubectl get services -n monitoring

# View logs
kubectl logs -f deployment/smart-monitor-backend -n monitoring
```

## 6. Configuration Management

### 6.1. Environment Variables

**Backend:**

```bash
# Server ports
GRPC_PORT=50051
HTTP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=smart_monitor
DB_USER=monitor
DB_PASSWORD=secure_password

# Security
JWT_SECRET=your-secret-key
TLS_CERT=/path/to/cert.pem
TLS_KEY=/path/to/key.pem

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Performance
MAX_CONNECTIONS=1000
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
```

**Agent:**

```bash
# Backend connection
BACKEND_SERVER=backend-server:50051
BACKEND_TLS=true

# Collection settings
HOSTNAME=server-01
INTERVAL=2s
TIMEOUT=10s

# Metrics to collect
ENABLE_CPU=true
ENABLE_MEMORY=true
ENABLE_DISK=true
ENABLE_NETWORK=true
```

### 6.2. Configuration File

```yaml
# config.yaml
server:
  grpc_port: 50051
  http_port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  name: smart_monitor
  user: monitor
  password: ${DB_PASSWORD}
  max_connections: 100

security:
  tls_enabled: true
  cert_file: /etc/ssl/certs/server.crt
  key_file: /etc/ssl/private/server.key
  jwt_secret: ${JWT_SECRET}

logging:
  level: info
  format: json
  output: /var/log/smart-monitor/backend.log

monitoring:
  enabled: true
  interval: 30s
  retention_days: 30
```

## 7. Security Configuration

### 7.1. TLS/SSL Setup

**Generate certificates:**

```bash
# Generate CA key and certificate
openssl genrsa -out ca.key 4096
openssl req -new -x509 -key ca.key -out ca.crt -days 365

# Generate server key and certificate
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

# Generate client certificates (for agents)
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365
```

**Configure Backend:**

```go
// Load TLS credentials
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
if err != nil {
    log.Fatalf("Failed to load TLS keys: %v", err)
}

// Create gRPC server with TLS
s := grpc.NewServer(grpc.Creds(creds))
```

**Configure Agent:**

```go
// Load client TLS credentials
creds, err := credentials.NewClientTLSFromFile("ca.crt", "backend-server")
if err != nil {
    log.Fatalf("Failed to load TLS certificate: %v", err)
}

// Connect with TLS
conn, err := grpc.Dial("backend-server:50051", grpc.WithTransportCredentials(creds))
```

### 7.2. Firewall Rules

```bash
# Backend server
sudo ufw allow 50051/tcp  # gRPC
sudo ufw allow 8080/tcp   # HTTP Gateway
sudo ufw enable

# Database server (if separate)
sudo ufw allow from backend-ip to any port 5432
```

### 7.3. Authentication

**JWT Token:**

```go
// Generate JWT token
func generateToken(userID string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })
    
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// Middleware for authentication
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing metadata")
    }
    
    // Validate token...
    
    return handler(ctx, req)
}
```

## 8. Monitoring & Logging

### 8.1. Application Logging

**Configure structured logging:**

```go
import "go.uber.org/zap"

func setupLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{
        "/var/log/smart-monitor/app.log",
        "stdout",
    }
    
    logger, _ := config.Build()
    return logger
}
```

### 8.2. System Monitoring

**Prometheus metrics:**

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "smart_monitor_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
}
```

### 8.3. Health Checks

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error": err.Error(),
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}
```

## 9. Backup & Recovery

### 9.1. Database Backup

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/smart-monitor"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_$DATE.sql"

# Create backup
pg_dump -h localhost -U monitor smart_monitor > $BACKUP_FILE

# Compress
gzip $BACKUP_FILE

# Delete old backups (keep 30 days)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_FILE.gz s3://my-backup-bucket/smart-monitor/
```

**Cron job:**
```bash
# Run daily at 2 AM
0 2 * * * /opt/smart-monitor/backup.sh
```

### 9.2. Restore

```bash
# Decompress backup
gunzip backup_20260115_020000.sql.gz

# Restore database
psql -h localhost -U monitor smart_monitor < backup_20260115_020000.sql
```

## 10. Scaling Strategies

### 10.1. Vertical Scaling

- Increase server resources (CPU, RAM)
- Optimize database queries
- Enable caching (Redis)

### 10.2. Horizontal Scaling

**Backend:**
- Deploy multiple instances behind load balancer
- Use stateless design
- Share session state via Redis

**Database:**
- Read replicas for queries
- Write to primary only
- Connection pooling

**Load Balancer (Nginx):**

```nginx
upstream backend_servers {
    least_conn;
    server backend1:50051 max_fails=3 fail_timeout=30s;
    server backend2:50051 max_fails=3 fail_timeout=30s;
    server backend3:50051 max_fails=3 fail_timeout=30s;
}

server {
    listen 50051 http2;
    
    location / {
        grpc_pass grpc://backend_servers;
    }
}
```

## 11. Troubleshooting

### 11.1. Common Issues

**Backend không start:**
```bash
# Check logs
sudo journalctl -u smart-monitor-backend -f

# Check port availability
sudo netstat -tlnp | grep 50051

# Check permissions
ls -l /opt/smart-monitor/
```

**Agent không kết nối:**
```bash
# Test connection
telnet backend-server 50051

# Check firewall
sudo ufw status

# Check DNS resolution
nslookup backend-server
```

### 11.2. Performance Issues

```bash
# Check system resources
top
htop
iotop

# Check network
iftop
netstat -s

# Check database
pg_stat_activity
```

## 12. Rollback Procedure

```bash
# 1. Stop current version
sudo systemctl stop smart-monitor-backend

# 2. Backup current binary
sudo mv /opt/smart-monitor/backend /opt/smart-monitor/backend.backup

# 3. Restore previous version
sudo cp /opt/smart-monitor/backend.previous /opt/smart-monitor/backend

# 4. Start service
sudo systemctl start smart-monitor-backend

# 5. Verify
sudo systemctl status smart-monitor-backend
```

## 13. Maintenance

### 13.1. Updates

```bash
# Build new version
go build -o backend-v2 main.go

# Test new version
./backend-v2 &
# Run tests...

# Deploy with zero-downtime
# 1. Start new version on different port
# 2. Update load balancer
# 3. Stop old version
```

### 13.2. Database Migrations

```bash
# Use migration tool like golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost/db" up
```

## 14. Checklist

### Pre-Deployment

- [ ] Code reviewed and tested
- [ ] Configuration files prepared
- [ ] Certificates generated
- [ ] Firewall rules configured
- [ ] Backup strategy in place
- [ ] Monitoring configured
- [ ] Documentation updated

### Deployment

- [ ] Build binaries
- [ ] Deploy to staging first
- [ ] Run smoke tests
- [ ] Deploy to production
- [ ] Verify health checks
- [ ] Monitor logs

### Post-Deployment

- [ ] Verify all services running
- [ ] Check metrics collection
- [ ] Test API endpoints
- [ ] Review logs for errors
- [ ] Update documentation
- [ ] Notify team

## 15. Support Contacts

- **DevOps Team**: devops@company.com
- **On-call**: +84-xxx-xxx-xxx
- **Slack**: #smart-monitor-alerts
- **Documentation**: https://docs.smart-monitor.com
