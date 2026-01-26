# Monitor-Test - Process Alert System

## 🎯 Tính năng mới

Monitor-test đã được nâng cấp với khả năng phát hiện và cảnh báo processes chiếm tài nguyên cao.

## 📊 Features

### 1. System Monitoring
- **CPU Usage**: Theo dõi và hiển thị với color-coded status
- **RAM Usage**: Giám sát memory với cảnh báo ngưỡng
- **Disk Usage**: Kiểm tra dung lượng ổ đĩa

### 2. Process Monitoring
- **Top CPU Consumers**: Liệt kê processes chiếm CPU cao nhất
- **Top RAM Consumers**: Liệt kê processes chiếm RAM nhiều nhất
- **Smart Filtering**: Chỉ hiển thị processes có tác động đáng kể

### 3. Alert System
- **Critical Alerts**: 🔴 Cảnh báo khi vượt ngưỡng (mặc định: 80%)
- **Warning Alerts**: 🟡 Cảnh báo sớm khi đạt 75% ngưỡng
- **OK Status**: 🟢 Hiển thị khi hoạt động bình thường

## 🚀 Sử dụng

### Build
```bash
cd monitor-test
go build -o monitor-test main.go
```

Hoặc dùng Makefile:
```bash
make build-monitor-test
```

### Run
```bash
./monitor-test
```

Hoặc:
```bash
make run-monitor-test
```

## 📋 Output Format

```
╔════════════════════════════════════════════════════════════════╗
║              System Monitor with Process Alerts                ║
╚════════════════════════════════════════════════════════════════╝

📊 CPU Usage:    15.86% 🟢 OK
💾 RAM Usage:    83.65% (Used: 13068MB / Total: 15623MB) 🔴 CRITICAL
💿 Disk Usage:   12.78% (Free: 362GB) 🟢 OK

⚠️  WARNING: High RAM usage detected (83.65%)!
🔥 Top processes by RAM:
   PID      NAME                      RAM(MB)      RAM%       USER
   ──────────────────────────────────────────────────────────────────────
   6735     code                      2937.3       18.80      tungnm2@...
   1053     clamd                     902.1        5.77       clamav
   217279   java                      829.0        5.31       it

📈 Top Resource Consumers:

  CPU Top 3:
   PID      NAME                      CPU%       RAM(MB)      USER
   ──────────────────────────────────────────────────────────────────────
   7168     code                      62.90      427.9        tungnm2@...
   204573   java                      45.09      777.5        it
   295103   code                      29.81      497.3        tungnm2@...

  RAM Top 3:
   PID      NAME                      RAM(MB)      RAM%       USER
   ──────────────────────────────────────────────────────────────────────
   6735     code                      2937.3       18.80      tungnm2@...
   1053     clamd                     902.1        5.77       clamav
   217279   java                      829.0        5.31       it
```

## ⚙️ Configuration

Có thể tùy chỉnh ngưỡng cảnh báo trong code:

```go
// Thresholds for alerts
cpuThreshold := 80.0     // CPU warning threshold (%)
ramThreshold := 80.0     // RAM warning threshold (%)
processCount := 5        // Top N processes to show in alerts
```

## 🎯 Alert Levels

| Level | Symbol | Điều kiện |
|-------|--------|-----------|
| OK | 🟢 | < 75% ngưỡng |
| WARNING | 🟡 | 75% - 99% ngưỡng |
| CRITICAL | 🔴 | ≥ ngưỡng |

## 📊 Process Information

Mỗi process hiển thị:
- **PID**: Process ID
- **NAME**: Tên process
- **CPU%**: Phần trăm CPU sử dụng
- **RAM(MB)**: Memory sử dụng (MB)
- **RAM%**: Phần trăm RAM sử dụng
- **USER**: User chạy process

## 🔍 Filtering Rules

### CPU Monitoring
- Chỉ hiển thị processes với CPU ≥ 0.1%
- Sort theo CPU usage giảm dần
- Hiển thị top N processes

### RAM Monitoring
- Chỉ hiển thị processes với RAM ≥ 10MB
- Sort theo RAM usage giảm dần
- Hiển thị top N processes

## 💡 Use Cases

### 1. Phát hiện Memory Leaks
```bash
# Chạy monitor và quan sát RAM usage tăng dần
./monitor-test

# Nếu thấy một process chiếm RAM tăng liên tục:
# → Có thể là memory leak
```

### 2. Tìm CPU Bottlenecks
```bash
# Khi hệ thống chậm, check CPU consumers
./monitor-test

# Processes chiếm CPU cao → Cần optimize hoặc restart
```

### 3. Early Warning System
```bash
# Setup để chạy background và log alerts
./monitor-test >> system-alerts.log 2>&1 &

# Review logs định kỳ
tail -f system-alerts.log | grep "WARNING"
```

## 🔧 Integration với Backend

Có thể mở rộng để gửi alerts tới backend:

```go
// Future enhancement
if cpuUsage > cpuThreshold || ramUsage > ramThreshold {
    sendAlertToBackend(AlertData{
        Type: "HIGH_RESOURCE_USAGE",
        CPU: cpuUsage,
        RAM: ramUsage,
        TopProcesses: getTopProcesses(),
    })
}
```

## 📈 Performance

- **Update Interval**: 5 giây (có thể điều chỉnh)
- **Memory Impact**: Minimal (~10-20MB)
- **CPU Impact**: Negligible (<1%)

## 🆘 Troubleshooting

### Process list empty
```bash
# Cần quyền đọc process info
sudo ./monitor-test
```

### High memory usage by monitor-test
```bash
# Giảm update frequency
# Sửa time.Sleep(5 * time.Second) thành 10 hoặc 30 giây
```

### Alerts không xuất hiện
```bash
# Check thresholds
# Có thể system usage < threshold
# Giảm threshold để test: cpuThreshold := 10.0
```

## 🚀 Future Enhancements

1. **Alert History**: Lưu lịch sử alerts
2. **Process Trends**: Theo dõi trends theo thời gian
3. **Kill Process**: Tự động kill processes vượt ngưỡng
4. **Network Monitoring**: Thêm network I/O tracking
5. **Web Dashboard**: Hiển thị real-time qua web interface
6. **Notification**: Email/Slack alerts
7. **Database Logging**: Lưu metrics vào DB

## 📚 Related

- [Makefile](../Makefile) - Build và run commands
- [Agent](../agent/main.go) - Production monitoring agent
- [Backend](../backend/) - Central monitoring server

---

**Tip**: Combine với `watch` command để monitoring liên tục:
```bash
watch -n 5 './monitor-test | head -50'
```
