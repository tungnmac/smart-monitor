# Smart Monitor Agent

Modern, modular monitoring agent with clean architecture.

## Architecture

```
agent/
├── cmd/
│   └── agent/          # Entry point
│       └── main.go
├── internal/
│   ├── agent/          # Core agent logic
│   │   └── agent.go
│   ├── client/         # Backend communication
│   │   └── client.go
│   ├── collector/      # Metrics collection
│   │   └── collector.go
│   ├── config/         # Configuration management
│   │   └── config.go
│   └── identity/       # Identity & credentials
│       └── identity.go
└── go.mod
```

## Features

- ✅ **Modular Architecture**: Clean separation of concerns
- ✅ **Auto-reconnect**: Automatic reconnection on network failures
- ✅ **Retry Logic**: Configurable retry attempts
- ✅ **Credential Caching**: Stores auth tokens locally
- ✅ **Graceful Shutdown**: Proper cleanup on SIGTERM/SIGINT
- ✅ **Environment Config**: Configure via environment variables
- ✅ **Extended Metrics**: CPU, RAM, Disk, Load, Network, Uptime
- ✅ **Easy to Extend**: Add new collectors or features easily

## Configuration

Configure via environment variables:

```bash
# Backend connection
export BACKEND_ADDR="localhost:50051"
export BACKEND_TLS="false"

# Metrics
export METRICS_INTERVAL="5"        # seconds
export BATCH_SIZE="10"

# Retry settings
export MAX_RETRIES="3"
export RETRY_INTERVAL="5"          # seconds
export RECONNECT_DELAY="10"        # seconds

# Metadata
export ENVIRONMENT="production"
export LOCATION="datacenter-01"
export DATACENTER="dc-01"

# Storage
export TOKEN_FILE=".agent_token"
export CONFIG_FILE="agent.yaml"
export LOG_FILE="agent.log"
```

## Building

```bash
# Build agent
cd agent
go build -o bin/agent ./cmd/agent

# Or use make
make build
```

## Running

```bash
# Run directly
./bin/agent

# With custom config
BACKEND_ADDR="backend.example.com:50051" \
METRICS_INTERVAL="10" \
ENVIRONMENT="staging" \
./bin/agent

# Run in background
nohup ./bin/agent > agent.log 2>&1 &
```

## Process Control CLI (`procctl`)

Call the backend `ProcessService` for process list/detail/restart:

```bash
# List processes on a host (default backend from BACKEND_ADDR or localhost:50051)
go run ./cmd/procctl --hostname host-1 --action list

# Show detail for a PID
go run ./cmd/procctl --hostname host-1 --action detail --pid 1234

# Restart (sends KillProcess; assumes supervisor restarts it)
go run ./cmd/procctl --hostname host-1 --action restart --pid 1234

# Override backend address
go run ./cmd/procctl --hostname host-1 --backend 10.0.0.5:50051 --action list
```

## Development

### Adding New Metrics

1. Add fields to `collector.Metrics` struct:
```go
type Metrics struct {
    CPUPercent float64
    // ... add your metric
    MyMetric float64
}
```

2. Implement collector method:
```go
func (c *Collector) collectMyMetric(metrics *Metrics) error {
    // Collect your metric
    metrics.MyMetric = value
    return nil
}
```

3. Call in `Collect()` method:
```go
if err := c.collectMyMetric(metrics); err != nil {
    // handle error
}
```

### Adding New Features

The modular design makes it easy to add:
- **New collectors**: Add to `internal/collector/`
- **New backends**: Add to `internal/client/`
- **New protocols**: Implement in `internal/client/`
- **New config sources**: Extend `internal/config/`

## Testing

```bash
# Run tests
go test ./...

# Test with coverage
go test -cover ./...

# Test specific package
go test ./internal/collector/
```

## Deployment

### Systemd Service

Create `/etc/systemd/system/smart-agent.service`:

```ini
[Unit]
Description=Smart Monitor Agent
After=network.target

[Service]
Type=simple
User=monitoring
Group=monitoring
WorkingDirectory=/opt/smart-monitor-agent
Environment="BACKEND_ADDR=backend.example.com:50051"
Environment="ENVIRONMENT=production"
ExecStart=/opt/smart-monitor-agent/bin/agent
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable smart-agent
sudo systemctl start smart-agent
sudo systemctl status smart-agent
```

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agent ./cmd/agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
CMD ["./agent"]
```

Build and run:
```bash
docker build -t smart-agent .
docker run -d \
  -e BACKEND_ADDR="backend:50051" \
  -e ENVIRONMENT="production" \
  --name smart-agent \
  smart-agent
```

## Troubleshooting

### Connection Issues

```bash
# Check connectivity
telnet backend.example.com 50051

# Check logs
tail -f agent.log

# Increase retry attempts
export MAX_RETRIES="10"
export RETRY_INTERVAL="10"
```

### Credential Issues

```bash
# Remove cached credentials
rm .agent_token

# Agent will re-register on next start
./bin/agent
```

### High Memory Usage

```bash
# Reduce batch size
export BATCH_SIZE="5"

# Increase metrics interval
export METRICS_INTERVAL="30"
```

## Performance

- **Memory**: ~10-20 MB baseline
- **CPU**: <1% on idle, ~2-3% during collection
- **Network**: ~1-5 KB/s depending on interval

## License

Copyright © 2026 Smart Monitor Team
