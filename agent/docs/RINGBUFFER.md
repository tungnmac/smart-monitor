# Ring Buffer Implementation for Agent

## Tổng Quan

Ring Buffer được triển khai để lưu trữ dữ liệu tạm thời khi mất kết nối với backend center. Hệ thống này đảm bảo không mất dữ liệu và tự động khôi phục khi kết nối được phục hồi.

## Thành phần chính

### 1. RingBuffer (`ringbuffer.go`)
- **Chức năng chính**: Cung cấp cơ chế vòng tròn để lưu trữ dữ liệu
- **Tính năng**:
  - Dung lượng cố định (mặc định: 1000 bản ghi)
  - Ghi đè dữ liệu cũ nhất khi buffer đầy (FIFO)
  - Thread-safe với mutex
  - Lưu trữ persistent trên disk

**API chính**:
```go
// Thêm một metric vào buffer
func (rb *RingBuffer) Push(stats *pb.StatsRequest)

// Lấy và xóa metric cũ nhất
func (rb *RingBuffer) Pop() *MetricData

// Lấy batch metrics (tối ưu hóa việc gửi)
func (rb *RingBuffer) PopBatch(batchSize int) []*MetricData

// Xem metric cũ nhất mà không xóa
func (rb *RingBuffer) Peek() *MetricData

// Lấy số lượng items trong buffer
func (rb *RingBuffer) Count() int

// Lưu tất cả dữ liệu ra disk
func (rb *RingBuffer) Persist() error

// Tải dữ liệu từ disk
func (rb *RingBuffer) restore() error
```

### 2. BufferManager (`manager.go`)
- **Chức năng chính**: Quản lý buffer và xử lý retry logic
- **Tính năng**:
  - Tự động buffering khi mất kết nối
  - Flush buffer tự động khi reconnect
  - Theo dõi trạng thái kết nối
  - Gửi dữ liệu theo batch để tối ưu

**API chính**:
```go
// Gửi stats, buffer nếu không gửi được
func (bm *BufferManager) Send(ctx context.Context, stream pb.MonitorService_StreamStatsClient, stats *pb.StatsRequest) error

// Xóa buffer (gửi tất cả dữ liệu đã lưu)
func (bm *BufferManager) flushBuffer(ctx context.Context, stream pb.MonitorService_StreamStatsClient) error

// Theo dõi kết nối và tự động reconnect
func (bm *BufferManager) StartConnectionMonitor(ctx context.Context, reconnectFunc func() (pb.MonitorService_StreamStatsClient, error))

// Lấy trạng thái buffer
func (bm *BufferManager) GetBufferStatus() map[string]interface{}
```

## Luồng hoạt động

### Khi kết nối bình thường:
```
Agent → Collect Metrics → Send directly to Backend → Log success
```

### Khi mất kết nối:
```
Agent → Collect Metrics → Send failed → Buffer locally → Log warning
    ↓
    Buffer persisted to disk (.agent_buffer/buffer_data.json)
    ↓
    Connection monitor detects disconnect
    ↓
    Attempts to reconnect every 5 seconds
    ↓
    On reconnect: Flush all buffered data → Send to Backend
```

### Khi shutdown:
```
Agent receives SIGTERM/SIGINT
    ↓
    Flush remaining buffered data to backend
    ↓
    Persist buffer to disk
    ↓
    Close connection gracefully
```

## Persistent Storage

Buffer data được lưu tại: `.agent_buffer/buffer_data.json`

Định dạng:
```json
[
  {
    "timestamp": 1707049200,
    "stats": {
      "hostname": "server-01",
      "agent_id": "agent-abc123",
      "cpu": 45.2,
      "ram": 72.5,
      "disk": 65.3,
      "access_token": "...",
      ...
    }
  },
  ...
]
```

## Configuration

Các hằng số có thể điều chỉnh trong `main.go`:

```go
bufferSize     = 1000     // Sức chứa ring buffer
bufferDataDir  = ".agent_buffer"  // Thư mục lưu dữ liệu
batchSize      = 50       // Kích thước batch khi flush
interval       = 2 * time.Second  // Khoảng thời gian collect metrics
```

## Status Monitoring

Mỗi 30 giây, agent sẽ in ra trạng thái buffer:

```
📊 Buffer Status: 45/1000 (4.5%) | Connected: true
```

Các thông tin:
- `45/1000`: 45 metrics trong buffer (tổng 1000)
- `4.5%`: Phần trăm sử dụng buffer
- `Connected: true`: Trạng thái kết nối với backend

## Lợi ích

1. **Không mất dữ liệu**: Tất cả metrics đều được lưu, ngay cả khi backend không khả dụng
2. **Tự động khôi phục**: Khi backend khôi phục, dữ liệu cũ sẽ được gửi lại tự động
3. **Persistent**: Dữ liệu được lưu trên disk, kể cả khi agent bị restart
4. **Batch processing**: Dữ liệu được gửi theo batch để giảm overhead
5. **Non-blocking**: Nếu kết nối slow, buffer sẽ xử lý tự động

## Ví dụ sử dụng

```go
// Khởi tạo ring buffer
rb, err := buffer.NewRingBuffer(1000, ".agent_buffer")
if err != nil {
    log.Fatal(err)
}

// Bắt đầu persistence worker
rb.StartPersistenceWorker(5 * time.Second)

// Tạo buffer manager
bufferMgr := buffer.NewBufferManager(rb, client, 5, 5*time.Second, 50)

// Gửi metrics với buffer fallback
if err := bufferMgr.Send(ctx, stream, stats); err != nil {
    // Đã được buffer tự động
}

// Khi shutdown, lưu buffer
if err := rb.ShutdownPersist(); err != nil {
    log.Printf("Error persisting: %v", err)
}
```

## Troubleshooting

### Buffer đầy
Nếu buffer luôn ở trạng thái đầy (100%), điều này có nghĩa:
- Backend không thể nhận dữ liệu
- Kiểm tra kết nối network
- Kiểm tra backend service status

### Dữ liệu không được flush
Kiểm tra:
1. Backend có đang chạy không?
2. Connection status có "Connected: true" không?
3. File `.agent_buffer/buffer_data.json` có tồn tại không?

### Performance impact
- Nếu buffer lớn, việc persist có thể mất 1-2 giây
- Khuyến nghị: Giữ buffer size ≤ 5000 cho hiệu suất tốt
