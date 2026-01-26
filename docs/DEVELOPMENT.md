# Hướng dẫn phát triển Smart Monitor

## 1. Yêu cầu hệ thống

### 1.1. Phần mềm cần thiết

- **Go**: Version 1.24 trở lên
  ```bash
  go version
  ```

- **Protocol Buffers Compiler**:
  ```bash
  # Linux
  sudo apt install protobuf-compiler
  
  # macOS
  brew install protobuf
  
  # Verify
  protoc --version
  ```

- **Go Plugins cho protoc**:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
  ```

- **Make**: Build automation
  ```bash
  # Linux
  sudo apt install build-essential
  
  # macOS
  xcode-select --install
  ```

- **Git**: Version control
  ```bash
  git --version
  ```

### 1.2. IDE/Editor đề xuất

- **VS Code** với extensions:
  - Go (golang.go)
  - vscode-proto3 (Protocol Buffers)
  - REST Client (HTTP requests)
  
- **GoLand** (JetBrains)

## 2. Setup môi trường phát triển

### 2.1. Clone và cài đặt

```bash
# Clone repository
git clone <repository-url>
cd smart-monitor

# Install Go dependencies
go mod download

# Verify setup
go mod verify
```

### 2.2. Cấu trúc workspace

```bash
smart-monitor/
├── .git/
├── .gitignore
├── go.mod              # Root module
├── go.sum
├── README.md
├── agent/
│   ├── go.mod          # Agent module
│   ├── go.sum
│   └── main.go
├── backend/
│   ├── main.go
│   └── static/         # Swagger UI files
├── pbtypes/            # Protocol Buffers
│   ├── makefile
│   └── */              # Service definitions
└── docs/               # Documentation
```

### 2.3. Generate Protocol Buffers

```bash
cd pbtypes

# Run makefile to generate all proto files
./run_makefile.sh

# Or use make directly
make all

# Verify generated files
ls -la */
```

Các file sẽ được generate:
- `*.pb.go` - Protocol Buffer messages
- `*_grpc.pb.go` - gRPC service definitions
- `*.pb.gw.go` - Gateway reverse proxy
- `*.swagger.json` - OpenAPI/Swagger definitions

## 3. Development Workflow

### 3.1. Quy trình phát triển feature mới

```
1. Create feature branch
   ↓
2. Define protobuf (if needed)
   ↓
3. Generate code
   ↓
4. Implement service
   ↓
5. Test locally
   ↓
6. Create pull request
   ↓
7. Code review
   ↓
8. Merge to main
```

### 3.2. Branch naming convention

```
feature/<feature-name>   # New feature
bugfix/<bug-name>        # Bug fix
hotfix/<issue>           # Production hotfix
refactor/<component>     # Code refactoring
docs/<topic>             # Documentation
```

**Ví dụ:**
```bash
git checkout -b feature/disk-monitoring
git checkout -b bugfix/memory-leak
git checkout -b docs/api-documentation
```

### 3.3. Commit message convention

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Ví dụ:**
```bash
git commit -m "feat(monitor): add disk monitoring service"
git commit -m "fix(agent): resolve memory leak in stats collection"
git commit -m "docs(readme): update installation instructions"
```

## 4. Coding Standards

### 4.1. Go Code Style

Tuân theo [Effective Go](https://golang.org/doc/effective_go.html) và [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key principles:**

```go
// ✅ Good: Exported names with documentation
// MonitorService handles system monitoring operations
type MonitorService struct {
    // ...
}

// CollectMetrics gathers system metrics from the host
func CollectMetrics() (*Stats, error) {
    // ...
}

// ❌ Bad: No documentation, unclear naming
type MS struct { }
func cm() (*Stats, error) { }
```

### 4.2. Error Handling

```go
// ✅ Good: Proper error handling with context
stats, err := collectStats()
if err != nil {
    return nil, fmt.Errorf("failed to collect stats: %w", err)
}

// ❌ Bad: Ignoring errors
stats, _ := collectStats()
```

### 4.3. Package Organization

```go
// ✅ Good: Logical grouping
package monitor

import (
    "context"
    "fmt"
    
    pb "smart-monitor/pbtypes/monitor"
)

// ❌ Bad: Too many imports, unclear organization
package main
import (
    "context"
    "fmt"
    pb1 "smart-monitor/pbtypes/monitor"
    pb2 "smart-monitor/pbtypes/system"
    // ... 20 more imports
)
```

### 4.4. Naming Conventions

```go
// Variables & Functions: camelCase (unexported), PascalCase (exported)
var maxConnections int
var ServerPort int

func processMetrics() { }
func StartServer() { }

// Constants: PascalCase or UPPER_CASE
const DefaultTimeout = 30
const MAX_RETRIES = 3

// Interfaces: -er suffix
type Reader interface { }
type MetricsCollector interface { }

// Struct fields: PascalCase (exported), camelCase (unexported)
type Config struct {
    Port        int    // Exported
    serverName  string // Unexported
}
```

## 5. Protocol Buffers Best Practices

### 5.1. Message Design

```protobuf
// ✅ Good: Clear, versioned, with comments
syntax = "proto3";

package monitor.v1;

// StatsRequest contains system metrics
message StatsRequest {
  string hostname = 1;      // Server hostname
  double cpu_percent = 2;   // CPU usage (0-100)
  double ram_percent = 3;   // RAM usage (0-100)
  int64 timestamp = 4;      // Unix timestamp
}

// ❌ Bad: No comments, unclear fields
message Req {
  string h = 1;
  double c = 2;
  double r = 3;
}
```

### 5.2. Service Definition

```protobuf
// ✅ Good: RESTful-like, clear operations
service MonitorService {
  // Stream real-time stats from agent
  rpc StreamStats(stream StatsRequest) returns (StatsResponse) {
    option (google.api.http) = {
      post: "/v1/monitor/stats"
      body: "*"
    };
  }
  
  // Get current stats for a host
  rpc GetStats(GetStatsRequest) returns (StatsResponse) {
    option (google.api.http) = {
      get: "/v1/monitor/stats/{hostname}"
    };
  }
}
```

### 5.3. Versioning

```
pbtypes/
├── monitor/
│   └── v1/
│       └── monitor.proto    # Version 1
├── system/
│   └── v1/
│       └── system.proto
```

## 6. Testing

### 6.1. Unit Tests

```go
// File: metrics_test.go
package monitor

import (
    "testing"
)

func TestCollectCPUMetrics(t *testing.T) {
    metrics, err := CollectCPUMetrics()
    if err != nil {
        t.Fatalf("CollectCPUMetrics failed: %v", err)
    }
    
    if metrics.Percent < 0 || metrics.Percent > 100 {
        t.Errorf("Invalid CPU percent: %f", metrics.Percent)
    }
}
```

**Run tests:**
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./agent/...

# Verbose output
go test -v ./...
```

### 6.2. Integration Tests

```go
func TestGRPCConnection(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Stop()
    
    // Create client
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        t.Fatalf("Connection failed: %v", err)
    }
    defer conn.Close()
    
    // Test RPC call
    client := pb.NewMonitorServiceClient(conn)
    resp, err := client.GetStats(context.Background(), &pb.GetStatsRequest{})
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

### 6.3. Benchmark Tests

```go
func BenchmarkCollectMetrics(b *testing.B) {
    for i := 0; i < b.N; i++ {
        CollectMetrics()
    }
}
```

**Run benchmarks:**
```bash
go test -bench=. -benchmem
```

## 7. Local Development

### 7.1. Chạy Backend

```bash
cd backend

# Run directly
go run main.go

# Or build and run
go build -o backend
./backend
```

**Access points:**
- gRPC: `localhost:50051`
- HTTP Gateway: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/`

### 7.2. Chạy Agent

```bash
cd agent

# Run with default config
go run main.go

# Run with custom parameters
go run main.go -server=localhost:50051 -interval=5s
```

### 7.3. Development Tools

**1. gRPC Client Testing:**
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Call method
grpcurl -plaintext -d '{"hostname":"test-server"}' \
  localhost:50051 monitor.MonitorService/GetStats
```

**2. REST API Testing:**
```bash
# Using curl
curl http://localhost:8080/v1/monitor/stats/test-server

# Using httpie
http GET localhost:8080/v1/monitor/stats/test-server
```

**3. Hot Reload:**
```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Create .air.toml config
air init

# Run with hot reload
air
```

## 8. Debugging

### 8.1. Logging

```go
import "log"

// Add contextual logs
log.Printf("[%s] Received stats: CPU=%.2f%%, RAM=%.2f%%", 
    req.Hostname, req.Cpu, req.Ram)

// Error logging
log.Printf("ERROR: Failed to process stats: %v", err)
```

### 8.2. Debugging với Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug backend
cd backend
dlv debug

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

### 8.3. Profiling

```go
import _ "net/http/pprof"

func main() {
    // Enable pprof endpoint
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... rest of code
}
```

**Access profiling:**
```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## 9. Code Quality Tools

### 9.1. Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### 9.2. Formatting

```bash
# Format all files
go fmt ./...

# Or use gofmt
gofmt -w .

# Use goimports (preferred)
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

### 9.3. Static Analysis

```bash
# go vet
go vet ./...

# staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## 10. Documentation

### 10.1. Code Documentation

```go
// Package monitor provides system monitoring capabilities.
//
// This package includes functions for collecting CPU, memory,
// disk, and network metrics from the host system.
package monitor

// Collector defines the interface for metrics collection.
//
// Implementations should be thread-safe and handle errors gracefully.
type Collector interface {
    // Collect gathers current metrics and returns them.
    // Returns an error if collection fails.
    Collect() (*Metrics, error)
}
```

### 10.2. Generate documentation

```bash
# View package documentation
go doc monitor

# Generate HTML documentation
godoc -http=:6060
# Visit http://localhost:6060/pkg/smart-monitor/
```

## 11. Performance Optimization

### 11.1. Optimization tips

- Use `sync.Pool` cho object reuse
- Minimize allocations trong hot paths
- Buffer channels appropriately
- Use `context` for cancellation
- Profile before optimizing

### 11.2. Example optimization

```go
// ❌ Bad: Creates new buffer每次
func processMetrics() {
    buf := make([]byte, 1024)
    // ... use buffer
}

// ✅ Good: Reuse buffers
var bufPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processMetrics() {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // ... use buffer
}
```

## 12. Troubleshooting

### 12.1. Common Issues

**Issue: Port already in use**
```bash
# Find process using port
lsof -i :50051

# Kill process
kill -9 <PID>
```

**Issue: Proto generation fails**
```bash
# Verify protoc installation
which protoc

# Check plugins
which protoc-gen-go
which protoc-gen-go-grpc

# Re-install plugins if needed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

**Issue: Import cycle**
```bash
# Analyze dependencies
go list -f '{{.ImportPath}} {{.Imports}}' ./...

# Visualize with go-callvis
go install github.com/ofabry/go-callvis@latest
go-callvis .
```

## 13. Resources

### 13.1. Documentation
- [Go Documentation](https://golang.org/doc/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [grpc-gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)

### 13.2. Books
- "The Go Programming Language" - Alan Donovan & Brian Kernighan
- "Concurrency in Go" - Katherine Cox-Buday
- "gRPC: Up and Running" - Kasun Indrasiri & Danesh Kuruppu

### 13.3. Community
- [Go Forum](https://forum.golangbridge.org/)
- [Gophers Slack](https://gophers.slack.com/)
- [r/golang](https://reddit.com/r/golang)

## 14. Getting Help

Khi gặp vấn đề:

1. ✅ Check documentation trong `docs/`
2. ✅ Search existing issues
3. ✅ Review error logs carefully
4. ✅ Create minimal reproducible example
5. ✅ Ask team members
6. ✅ Create issue với đầy đủ context

**Issue template:**
```markdown
## Environment
- OS: 
- Go version:
- Component: [agent/backend]

## Expected Behavior

## Actual Behavior

## Steps to Reproduce
1. 
2. 
3. 

## Error Logs
```paste logs here```
```
