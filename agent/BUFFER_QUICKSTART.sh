#!/usr/bin/env bash

# Quick Start Guide for Ring Buffer Implementation

cat << 'EOF'

╔══════════════════════════════════════════════════════════════╗
║          Ring Buffer for Smart Monitor Agent                ║
║      Reliable Data Storage During Connection Loss           ║
╚══════════════════════════════════════════════════════════════╝

📋 OVERVIEW
═══════════════════════════════════════════════════════════════

Ring Buffer là cơ chế lưu trữ dữ liệu vòng tròn với các tính năng:

  ✓ Tự động buffer khi mất kết nối backend
  ✓ Persistent storage (dữ liệu lưu trên disk)
  ✓ Tự động khôi phục khi kết nối phục hồi
  ✓ FIFO queue (First In First Out)
  ✓ Thread-safe và không block
  ✓ Batch processing để tối ưu hiệu suất


🚀 QUICK START
═══════════════════════════════════════════════════════════════

1. Initialize Buffer:
   
   rb, err := buffer.NewRingBuffer(1000, ".agent_buffer")
   if err != nil {
       log.Fatal(err)
   }
   rb.StartPersistenceWorker(5 * time.Second)

2. Create Buffer Manager:
   
   bufferMgr := buffer.NewBufferManager(
       rb, client, 
       5,              // maxRetries
       5*time.Second,  // retryInterval
       50,             // batchSize
   )

3. Use Buffer Manager for Sending:
   
   if err := bufferMgr.Send(ctx, stream, stats); err != nil {
       // Error logged, stats buffered automatically
   }

4. On Shutdown:
   
   if err := rb.ShutdownPersist(); err != nil {
       log.Printf("Error persisting: %v", err)
   }


📊 BUFFER LIFECYCLE
═══════════════════════════════════════════════════════════════

Normal Operation:
  ┌─────────┐      ┌──────────┐      ┌──────────┐
  │ Collect │  →   │  Send to │  →   │   Log    │
  │ Metrics │      │ Backend  │      │ Success  │
  └─────────┘      └──────────┘      └──────────┘

Connection Lost:
  ┌─────────┐      ┌──────────┐      ┌─────────────┐
  │ Collect │  →   │  Send    │  →   │   Buffer &  │
  │ Metrics │      │  Failed  │      │   Persist   │
  └─────────┘      └──────────┘      └─────────────┘

Connection Restored:
  ┌──────────┐     ┌────────┐      ┌────────────┐
  │  Buffered│  →  │ Flush  │  →   │   Send to  │
  │  Data    │     │ Batch  │      │  Backend   │
  └──────────┘     └────────┘      └────────────┘


⚙️  CONFIGURATION
═══════════════════════════════════════════════════════════════

Predefined Configs:

  Development (smaller buffer, quick feedback):
    cfg := buffer.DevelopmentConfig()
    // BufferSize: 100
    // PersistInterval: 2s
    // StatusInterval: 10s

  Production (larger buffer, less overhead):
    cfg := buffer.ProductionConfig()
    // BufferSize: 5000
    // PersistInterval: 10s
    // StatusInterval: 60s

  Custom:
    cfg := buffer.Config{
        BufferSize:      2000,
        BufferDataDir:   "/var/lib/agent/buffer",
        PersistInterval: 7 * time.Second,
        BatchSize:       75,
        ...
    }


📈 MONITORING
═══════════════════════════════════════════════════════════════

Buffer Status (printed every 30 seconds):
  📊 Buffer Status: 45/1000 (4.5%) | Connected: true

Interpretation:
  • 45/1000    = 45 metrics in buffer (capacity 1000)
  • 4.5%       = Percentage used
  • Connected  = gRPC connection status

Programmatically:
  status := bufferMgr.GetBufferStatus()
  
  status["count"]       → 45 (items in buffer)
  status["capacity"]    → 1000 (total capacity)
  status["is_full"]     → false (buffer full?)
  status["percentage"]  → 4.5 (% used)
  status["is_connected"] → true (connected to backend?)


💾 PERSISTENT STORAGE
═══════════════════════════════════════════════════════════════

Location: .agent_buffer/buffer_data.json

Format (JSON array of metrics):
  [
    {
      "timestamp": 1707049200,
      "stats": {
        "hostname": "server-01",
        "agent_id": "agent-abc123",
        "cpu": 45.2,
        "ram": 72.5,
        "disk": 65.3,
        ...
      }
    },
    ...
  ]

Features:
  ✓ Survives agent restart
  ✓ Survives system crash
  ✓ Automatically restored on startup
  ✓ Incremental updates every 5 seconds


🔄 RECONNECTION FLOW
═══════════════════════════════════════════════════════════════

1. Connection Lost (gRPC error detected)
   └─ Log: "⚠️  Backend connection lost, buffering metrics..."
   └─ Agent switches to buffer mode

2. Connection Monitor Active
   └─ Every 5 seconds: checks if backend is available
   └─ On reconnect: attempts to restore stream

3. Data Flush
   └─ Sends buffered data in batches of 50
   └─ Waits 100ms between batches
   └─ Continues until buffer is empty

4. Normal Operation Resumes
   └─ New metrics sent directly
   └─ Buffer remains for future disconnections


🧪 TESTING
═══════════════════════════════════════════════════════════════

Run all buffer tests:
  $ go test -v ./internal/buffer

Test coverage:
  $ go test -cover ./internal/buffer

Tests included:
  ✓ Push/Pop operations
  ✓ Batch operations
  ✓ Buffer overflow
  ✓ Persistence (save/load)
  ✓ FIFO order
  ✓ Thread safety


📚 API REFERENCE
═══════════════════════════════════════════════════════════════

RingBuffer:
  Push(stats)              - Add metric to buffer
  Pop()                    - Remove & return oldest metric
  PopBatch(size)           - Remove & return batch
  Peek()                   - View oldest without removing
  Count()                  - Number of items
  IsFull()                 - At capacity?
  Clear()                  - Remove all items
  Persist()                - Save to disk
  ShutdownPersist()        - Final save on shutdown

BufferManager:
  Send(ctx, stream, stats) - Send with buffer fallback
  FlushBuffer(ctx, stream) - Flush all buffered data
  GetBufferStatus()        - Get status as map
  StartConnectionMonitor() - Monitor connection

Config:
  DefaultConfig()          - Balanced settings
  ProductionConfig()       - Optimized for production
  DevelopmentConfig()      - Optimized for development


⚠️  TROUBLESHOOTING
═══════════════════════════════════════════════════════════════

Buffer always full (100%):
  → Check backend connection
  → Verify backend service is running
  → Check network connectivity
  Action: Restart agent once backend is fixed

Data not being flushed:
  → Check "Connected: false" in logs
  → Verify agent can reach backend IP:port
  → Check firewall rules
  Action: `telnet <backend_ip> <port>`

High memory usage:
  → Buffer size too large for system
  Action: Reduce BufferSize in config
  Consider: 100-5000 for most systems

Agent startup slow:
  → Large buffer_data.json file
  Action: Delete .agent_buffer/buffer_data.json
  Side effect: Old buffered data will be lost


🔧 PERFORMANCE TUNING
═══════════════════════════════════════════════════════════════

For high-frequency metrics:
  - Increase BatchSize (50 → 100)
  - Decrease PersistInterval (5s → 3s)
  - Increase BufferSize (1000 → 5000)

For low bandwidth:
  - Decrease BatchSize (50 → 25)
  - Increase PersistInterval (5s → 10s)
  - Decrease BufferSize (1000 → 500)

For reliability:
  - Increase BufferSize (always have backup)
  - Decrease RetryInterval (faster recovery)
  - Increase StatusInterval (less logging)


📝 FILES
═══════════════════════════════════════════════════════════════

Implementation:
  internal/buffer/ringbuffer.go     - Core ring buffer
  internal/buffer/manager.go        - Buffer manager & retry
  internal/buffer/config.go         - Configuration

Testing:
  internal/buffer/ringbuffer_test.go - Unit tests

Documentation:
  RINGBUFFER.md             - Detailed documentation
  RINGBUFFER_DEMO.sh        - Demo script

Integration in main.go:
  - Initialize in main()
  - Use in streamStats()
  - Persist on shutdown


💡 BEST PRACTICES
═══════════════════════════════════════════════════════════════

1. Always call StartPersistenceWorker()
   → Ensures data is saved periodically

2. Always call ShutdownPersist() on exit
   → Persists remaining data

3. Monitor buffer status logs
   → Helps detect connectivity issues early

4. Size buffer appropriately
   → Too small: data loss
   → Too large: memory usage

5. Test with backend offline
   → Verify buffering works
   → Check persistence file

6. Monitor disk space
   → Buffer file can grow large
   → Implement cleanup if needed


═══════════════════════════════════════════════════════════════
                    HAPPY MONITORING! 📊
═══════════════════════════════════════════════════════════════

EOF
