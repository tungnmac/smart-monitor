package buffer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "smart-monitor/pbtypes/monitor"
)

// MetricData wraps stats with metadata for storage
type MetricData struct {
	Timestamp int64            `json:"timestamp"`
	Stats     *pb.StatsRequest `json:"stats"`
}

// RingBuffer is a circular buffer that stores metrics
type RingBuffer struct {
	mu        sync.RWMutex
	buffer    []*MetricData
	size      int
	head      int // Next write position
	tail      int // Next read position
	count     int // Number of items in buffer
	isFull    bool
	dataDir   string
	persistCh chan bool
}

// NewRingBuffer creates a new ring buffer with specified capacity
func NewRingBuffer(capacity int, dataDir string) (*RingBuffer, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	rb := &RingBuffer{
		buffer:    make([]*MetricData, capacity),
		size:      capacity,
		head:      0,
		tail:      0,
		count:     0,
		isFull:    false,
		dataDir:   dataDir,
		persistCh: make(chan bool, 1),
	}

	// Try to restore from disk
	if err := rb.restore(); err != nil {
		// If restore fails, start fresh
		fmt.Printf("Warning: Failed to restore buffer from disk: %v\n", err)
	}

	return rb, nil
}

// Push adds a new metric to the buffer
func (rb *RingBuffer) Push(stats *pb.StatsRequest) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	data := &MetricData{
		Timestamp: time.Now().Unix(),
		Stats:     stats,
	}

	rb.buffer[rb.head] = data
	rb.head = (rb.head + 1) % rb.size

	if rb.isFull {
		rb.tail = (rb.tail + 1) % rb.size
	} else {
		rb.count++
		if rb.count == rb.size {
			rb.isFull = true
		}
	}

	// Trigger persistence
	select {
	case rb.persistCh <- true:
	default:
	}
}

// Pop retrieves and removes the oldest metric from the buffer
func (rb *RingBuffer) Pop() *MetricData {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	data := rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % rb.size
	rb.count--
	rb.isFull = false

	return data
}

// PopBatch retrieves multiple metrics from the buffer
func (rb *RingBuffer) PopBatch(batchSize int) []*MetricData {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	if batchSize > rb.count {
		batchSize = rb.count
	}

	result := make([]*MetricData, batchSize)
	for i := 0; i < batchSize; i++ {
		result[i] = rb.buffer[rb.tail]
		rb.tail = (rb.tail + 1) % rb.size
	}
	rb.count -= batchSize
	rb.isFull = false

	return result
}

// Peek returns the oldest metric without removing it
func (rb *RingBuffer) Peek() *MetricData {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	return rb.buffer[rb.tail]
}

// Count returns the number of items in the buffer
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// GetSize returns the capacity of the buffer
func (rb *RingBuffer) GetSize() int {
	return rb.size
}

// IsFull checks if the buffer is at capacity
func (rb *RingBuffer) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.isFull
}

// Clear removes all items from the buffer
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.tail = 0
	rb.count = 0
	rb.isFull = false
	rb.buffer = make([]*MetricData, rb.size)
}

// GetFilePath returns the path to the persistence file
func (rb *RingBuffer) GetFilePath() string {
	return filepath.Join(rb.dataDir, "buffer_data.json")
}

// Persist saves all buffer contents to disk
func (rb *RingBuffer) Persist() error {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		// Delete file if buffer is empty
		_ = os.Remove(rb.GetFilePath())
		return nil
	}

	// Collect all data from buffer
	var data []*MetricData
	idx := rb.tail
	for i := 0; i < rb.count; i++ {
		data = append(data, rb.buffer[idx])
		idx = (idx + 1) % rb.size
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal buffer data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(rb.GetFilePath(), jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write buffer file: %w", err)
	}

	return nil
}

// restore loads buffer contents from disk
func (rb *RingBuffer) restore() error {
	filePath := rb.GetFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, this is fine
			return nil
		}
		return fmt.Errorf("failed to read buffer file: %w", err)
	}

	var metrics []*MetricData
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal buffer data: %w", err)
	}

	// Restore data into buffer
	for _, metric := range metrics {
		if rb.count >= rb.size {
			break // Don't exceed capacity
		}
		rb.buffer[rb.head] = metric
		rb.head = (rb.head + 1) % rb.size
		rb.count++
	}

	if rb.count == rb.size {
		rb.isFull = true
	}

	return nil
}

// StartPersistenceWorker starts a goroutine that periodically persists buffer to disk
func (rb *RingBuffer) StartPersistenceWorker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-rb.persistCh:
				// Persist immediately when triggered
				if err := rb.Persist(); err != nil {
					fmt.Printf("Error persisting buffer: %v\n", err)
				}
			case <-ticker.C:
				// Periodic persistence
				if err := rb.Persist(); err != nil {
					fmt.Printf("Error persisting buffer: %v\n", err)
				}
			}
		}
	}()
}

// ShutdownPersist persists all data before shutdown
func (rb *RingBuffer) ShutdownPersist() error {
	return rb.Persist()
}
