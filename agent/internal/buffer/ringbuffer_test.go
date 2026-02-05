package buffer

import (
	"os"
	"testing"
	"time"

	pb "smart-monitor/pbtypes/monitor"
)

func TestRingBuffer_Push(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	stats := &pb.StatsRequest{
		Hostname: "test-host",
		AgentId:  "agent-123",
		Cpu:      45.5,
		Ram:      72.3,
		Disk:     65.1,
	}

	rb.Push(stats)

	if rb.Count() != 1 {
		t.Errorf("Expected count 1, got %d", rb.Count())
	}
}

func TestRingBuffer_Pop(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	stats1 := &pb.StatsRequest{
		Hostname: "test-host-1",
		AgentId:  "agent-1",
		Cpu:      45.5,
	}
	stats2 := &pb.StatsRequest{
		Hostname: "test-host-2",
		AgentId:  "agent-2",
		Cpu:      55.5,
	}

	rb.Push(stats1)
	rb.Push(stats2)

	if rb.Count() != 2 {
		t.Errorf("Expected count 2, got %d", rb.Count())
	}

	data := rb.Pop()
	if data == nil {
		t.Fatal("Expected data, got nil")
	}

	if data.Stats.Hostname != "test-host-1" {
		t.Errorf("Expected hostname test-host-1, got %s", data.Stats.Hostname)
	}

	if rb.Count() != 1 {
		t.Errorf("Expected count 1 after pop, got %d", rb.Count())
	}
}

func TestRingBuffer_PopBatch(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	// Add 5 items
	for i := 1; i <= 5; i++ {
		stats := &pb.StatsRequest{
			Hostname: "host",
			AgentId:  "agent",
		}
		rb.Push(stats)
	}

	batch := rb.PopBatch(3)
	if len(batch) != 3 {
		t.Errorf("Expected batch size 3, got %d", len(batch))
	}

	if rb.Count() != 2 {
		t.Errorf("Expected count 2 after batch pop, got %d", rb.Count())
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb, err := NewRingBuffer(3, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	// Add 5 items (exceeds capacity of 3)
	for i := 1; i <= 5; i++ {
		stats := &pb.StatsRequest{
			Hostname: "host",
			AgentId:  "agent",
			Cpu:      float64(i * 10),
		}
		rb.Push(stats)
	}

	// Should have 3 items (oldest 2 were replaced)
	if rb.Count() != 3 {
		t.Errorf("Expected count 3, got %d", rb.Count())
	}

	// First item should be the 3rd one we pushed (40.0)
	first := rb.Peek()
	if first.Stats.Cpu != 30.0 {
		t.Errorf("Expected CPU 30.0, got %f", first.Stats.Cpu)
	}
}

func TestRingBuffer_Persist(t *testing.T) {
	dir := t.TempDir()
	rb, err := NewRingBuffer(5, dir)
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	stats := &pb.StatsRequest{
		Hostname: "test-host",
		AgentId:  "agent-123",
		Cpu:      45.5,
		Ram:      72.3,
	}

	rb.Push(stats)

	// Persist to disk
	if err := rb.Persist(); err != nil {
		t.Errorf("Failed to persist: %v", err)
	}

	// Check file exists
	filePath := rb.GetFilePath()
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Buffer file not found: %v", err)
	}

	// Create new buffer and restore
	rb2, err := NewRingBuffer(5, dir)
	if err != nil {
		t.Fatalf("Failed to create new buffer: %v", err)
	}

	if rb2.Count() != 1 {
		t.Errorf("Expected restored count 1, got %d", rb2.Count())
	}

	restored := rb2.Pop()
	if restored.Stats.Hostname != "test-host" {
		t.Errorf("Expected hostname test-host, got %s", restored.Stats.Hostname)
	}
}

func TestRingBuffer_IsFull(t *testing.T) {
	rb, err := NewRingBuffer(2, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	if rb.IsFull() {
		t.Error("Buffer should not be full initially")
	}

	rb.Push(&pb.StatsRequest{})
	if rb.IsFull() {
		t.Error("Buffer should not be full with 1 item")
	}

	rb.Push(&pb.StatsRequest{})
	if !rb.IsFull() {
		t.Error("Buffer should be full with 2 items")
	}

	rb.Push(&pb.StatsRequest{})
	if !rb.IsFull() {
		t.Error("Buffer should remain full after overflow")
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	for i := 0; i < 3; i++ {
		rb.Push(&pb.StatsRequest{})
	}

	if rb.Count() != 3 {
		t.Errorf("Expected count 3, got %d", rb.Count())
	}

	rb.Clear()

	if rb.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", rb.Count())
	}

	if rb.IsFull() {
		t.Error("Buffer should not be full after clear")
	}
}

func TestRingBuffer_Peek(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	stats := &pb.StatsRequest{
		Hostname: "test-host",
		AgentId:  "agent-123",
	}

	rb.Push(stats)

	// Peek should not remove the item
	peeked := rb.Peek()
	if peeked == nil {
		t.Fatal("Expected data from peek, got nil")
	}

	if rb.Count() != 1 {
		t.Errorf("Count should remain 1 after peek, got %d", rb.Count())
	}

	if peeked.Stats.Hostname != "test-host" {
		t.Errorf("Expected hostname test-host, got %s", peeked.Stats.Hostname)
	}
}

func TestRingBuffer_Timestamp(t *testing.T) {
	rb, err := NewRingBuffer(5, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create ring buffer: %v", err)
	}

	before := time.Now().Unix()
	rb.Push(&pb.StatsRequest{})
	after := time.Now().Unix()

	data := rb.Peek()
	if data.Timestamp < before || data.Timestamp > after {
		t.Errorf("Timestamp should be between %d and %d, got %d", before, after, data.Timestamp)
	}
}
