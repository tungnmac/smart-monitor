package buffer

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "smart-monitor/pbtypes/monitor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BufferManager manages the ring buffer and handles retry logic
type BufferManager struct {
	buffer           *RingBuffer
	client           pb.MonitorServiceClient
	maxRetries       int
	retryInterval    time.Duration
	batchSize        int
	maxBatchWait     time.Duration
	isConnected      bool
	connectionTicker time.Duration
}

// NewBufferManager creates a new buffer manager
func NewBufferManager(
	buffer *RingBuffer,
	client pb.MonitorServiceClient,
	maxRetries int,
	retryInterval time.Duration,
	batchSize int,
) *BufferManager {
	return &BufferManager{
		buffer:           buffer,
		client:           client,
		maxRetries:       maxRetries,
		retryInterval:    retryInterval,
		batchSize:        batchSize,
		maxBatchWait:     5 * time.Second,
		isConnected:      true,
		connectionTicker: 5 * time.Second,
	}
}

// Send tries to send stats immediately, falls back to buffer if fails
func (bm *BufferManager) Send(ctx context.Context, stream pb.MonitorService_StreamStatsClient, stats *pb.StatsRequest) error {
	// First, try to send any buffered data
	if err := bm.FlushBuffer(ctx, stream); err != nil {
		log.Printf("Warning: Failed to flush buffer: %v", err)
	}

	// Try to send the new stats
	if err := stream.Send(stats); err != nil {
		log.Printf("Failed to send stats directly, buffering: %v", err)

		// Extract connection error details
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded {
				bm.isConnected = false
				log.Println("⚠️  Backend connection lost, buffering metrics...")
			}
		}

		// Buffer the stats for later
		bm.buffer.Push(stats)
		return nil
	}

	bm.isConnected = true
	return nil
}

// FlushBuffer attempts to send all buffered data to backend
func (bm *BufferManager) FlushBuffer(ctx context.Context, stream pb.MonitorService_StreamStatsClient) error {
	for {
		batch := bm.buffer.PopBatch(bm.batchSize)
		if len(batch) == 0 {
			return nil
		}

		log.Printf("📤 Flushing %d buffered metrics...", len(batch))

		for _, data := range batch {
			select {
			case <-ctx.Done():
				// Re-buffer the data if context is cancelled
				bm.buffer.Push(data.Stats)
				return ctx.Err()
			default:
				if err := stream.Send(data.Stats); err != nil {
					// Re-buffer on failure
					bm.buffer.Push(data.Stats)
					return fmt.Errorf("failed to flush buffer: %w", err)
				}
				log.Printf("✓ Sent buffered metric from %s", time.Unix(data.Timestamp, 0).Format(time.RFC3339))
			}
		}

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}
}

// StartConnectionMonitor monitors connection status and attempts reconnection
func (bm *BufferManager) StartConnectionMonitor(ctx context.Context, reconnectFunc func() (pb.MonitorService_StreamStatsClient, error)) {
	go func() {
		ticker := time.NewTicker(bm.connectionTicker)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !bm.isConnected {
					log.Println("🔄 Attempting to reconnect to backend...")
					if stream, err := reconnectFunc(); err == nil {
						log.Println("✓ Reconnected to backend")
						bm.isConnected = true

						// Try to flush buffered data
						if err := bm.FlushBuffer(ctx, stream); err != nil {
							log.Printf("Error flushing buffer after reconnect: %v", err)
						}
					} else {
						log.Printf("⚠️  Reconnection failed: %v", err)
					}
				}
			}
		}
	}()
}

// GetBufferStatus returns current buffer status
func (bm *BufferManager) GetBufferStatus() map[string]interface{} {
	return map[string]interface{}{
		"count":        bm.buffer.Count(),
		"capacity":     bm.buffer.size,
		"is_full":      bm.buffer.IsFull(),
		"is_connected": bm.isConnected,
		"percentage":   float64(bm.buffer.Count()) / float64(bm.buffer.size) * 100,
	}
}

// WaitForConnectionWithBuffer waits for connection, buffering data meanwhile
func (bm *BufferManager) WaitForConnectionWithBuffer(ctx context.Context, dialFunc func() (*grpc.ClientConn, error), maxWait time.Duration) (*grpc.ClientConn, error) {
	startTime := time.Now()
	attempt := 0

	for {
		attempt++
		log.Printf("Attempting to connect to backend (attempt %d)...", attempt)

		conn, err := dialFunc()
		if err == nil {
			log.Println("✓ Connected to backend")
			bm.isConnected = true
			return conn, nil
		}

		log.Printf("⚠️  Connection attempt %d failed: %v", attempt, err)

		if time.Since(startTime) > maxWait {
			return nil, fmt.Errorf("failed to connect within %v", maxWait)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(bm.retryInterval):
			// Continue retry
		}
	}
}
