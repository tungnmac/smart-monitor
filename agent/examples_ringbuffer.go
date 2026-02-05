package main

import (
	"fmt"
	"log"
	"time"

	"smart-agent/internal/buffer"
	pb "smart-monitor/pbtypes/monitor"
)

// ExampleRingBuffer demonstrates how to use the ring buffer
func ExampleRingBuffer() {
	log.Println("=== Ring Buffer Example ===\n")

	// Step 1: Create a ring buffer
	log.Println("Step 1: Creating ring buffer with capacity 10...")
	rb, err := buffer.NewRingBuffer(10, "./example_buffer")
	if err != nil {
		log.Fatalf("Failed to create buffer: %v", err)
	}
	log.Printf("✓ Ring buffer created\n")

	// Step 2: Add some metrics
	log.Println("Step 2: Adding metrics to buffer...")
	metrics := []struct {
		hostname string
		cpu      float64
		ram      float64
	}{
		{"server-01", 45.5, 72.3},
		{"server-02", 52.1, 65.4},
		{"server-03", 38.9, 58.2},
		{"server-04", 61.2, 81.5},
		{"server-05", 33.7, 55.9},
	}

	for i, m := range metrics {
		stats := &pb.StatsRequest{
			Hostname:  m.hostname,
			AgentId:   fmt.Sprintf("agent-%d", i+1),
			Cpu:       m.cpu,
			Ram:       m.ram,
			IpAddress: fmt.Sprintf("192.168.1.%d", 100+i),
		}
		rb.Push(stats)
		log.Printf("  Added metric for %s (CPU: %.1f%%, RAM: %.1f%%)", m.hostname, m.cpu, m.ram)
	}
	log.Printf("✓ Total items in buffer: %d\n", rb.Count())

	// Step 3: Check buffer status
	log.Println("Step 3: Checking buffer status...")
	count := rb.Count()
	capacity := rb.GetSize()
	percentage := float64(count) / float64(capacity) * 100
	log.Printf("  Count: %d/%d (%.1f%% used)\n", count, capacity, percentage)

	// Step 4: Peek at the oldest metric
	log.Println("Step 4: Peeking at the oldest metric...")
	oldest := rb.Peek()
	if oldest != nil {
		log.Printf("  Oldest metric: %s | CPU: %.1f%% | Timestamp: %s\n",
			oldest.Stats.Hostname, oldest.Stats.Cpu,
			time.Unix(oldest.Timestamp, 0).Format("2006-01-02 15:04:05"))
	}

	// Step 5: Pop items one by one
	log.Println("Step 5: Popping metrics (FIFO order)...")
	for i := 0; i < 2; i++ {
		data := rb.Pop()
		if data != nil {
			log.Printf("  Popped: %s | CPU: %.1f%% | Timestamp: %s",
				data.Stats.Hostname, data.Stats.Cpu,
				time.Unix(data.Timestamp, 0).Format("2006-01-02 15:04:05"))
		}
	}
	log.Printf("✓ Remaining items: %d\n", rb.Count())

	// Step 6: Pop batch
	log.Println("Step 6: Popping batch of 2 metrics...")
	batch := rb.PopBatch(2)
	for _, data := range batch {
		log.Printf("  Batch item: %s | CPU: %.1f%%", data.Stats.Hostname, data.Stats.Cpu)
	}
	log.Printf("✓ Remaining items: %d\n", rb.Count())

	// Step 7: Persist to disk
	log.Println("Step 7: Persisting buffer to disk...")
	if err := rb.Persist(); err != nil {
		log.Printf("Error persisting: %v\n", err)
	} else {
		log.Printf("✓ Buffer persisted to: %s\n", rb.GetFilePath())
	}

	// Step 8: Clear and show empty buffer
	log.Println("Step 8: Clearing buffer...")
	rb.Clear()
	log.Printf("✓ Buffer cleared. Items: %d\n", rb.Count())

	// Step 9: Restore from disk
	log.Println("Step 9: Restoring buffer from disk...")
	rb2, err := buffer.NewRingBuffer(10, "./example_buffer")
	if err != nil {
		log.Fatalf("Failed to create buffer: %v", err)
	}
	log.Printf("✓ Buffer restored. Items: %d\n", rb2.Count())

	// Step 10: Show restored data
	log.Println("Step 10: Displaying restored metrics...")
	for i := 1; i <= rb2.Count(); i++ {
		data := rb2.Pop()
		if data != nil {
			log.Printf("  Restored[%d]: %s | CPU: %.1f%% | RAM: %.1f%%",
				i, data.Stats.Hostname, data.Stats.Cpu, data.Stats.Ram)
		}
	}

	log.Println("\n=== Example Complete ===")
}
