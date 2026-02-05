package buffer

import (
	"os"
	"testing"
	"time"

	pb "smart-monitor/pbtypes/monitor"
)

// TestIntegration_BufferWorkflow tests the complete workflow
func TestIntegration_BufferWorkflow(t *testing.T) {
	t.Log("Testing complete buffer workflow...")

	// Step 1: Create buffer
	t.Log("Step 1: Creating ring buffer...")
	rb, err := NewRingBuffer(10, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Step 2: Simulate normal operation
	t.Log("Step 2: Simulating normal operation (adding 5 metrics)...")
	for i := 1; i <= 5; i++ {
		stats := &pb.StatsRequest{
			Hostname: "server-01",
			AgentId:  "agent-001",
			Cpu:      float64(i * 10),
			Ram:      float64(i * 20),
		}
		rb.Push(stats)
	}

	if rb.Count() != 5 {
		t.Fatalf("Expected 5 items, got %d", rb.Count())
	}
	t.Logf("✓ Buffer has %d metrics", rb.Count())

	// Step 3: Persist to disk
	t.Log("Step 3: Persisting buffer to disk...")
	if err := rb.Persist(); err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}

	filePath := rb.GetFilePath()
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("Persistence file not found: %v", err)
	}
	t.Logf("✓ Buffer persisted to %s", filePath)

	// Step 4: Simulate data transmission (flush buffer)
	t.Log("Step 4: Flushing buffer (simulating data transmission)...")
	batch := rb.PopBatch(3)
	if len(batch) != 3 {
		t.Fatalf("Expected batch of 3, got %d", len(batch))
	}
	t.Logf("✓ Flushed %d metrics, %d remaining", len(batch), rb.Count())

	// Step 5: Persist remaining data
	t.Log("Step 5: Persisting remaining buffer...")
	if err := rb.Persist(); err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}

	// Step 6: Restart agent (create new buffer instance)
	t.Log("Step 6: Simulating agent restart (new buffer instance)...")
	rb2, err := NewRingBuffer(10, rb.dataDir)
	if err != nil {
		t.Fatalf("Failed to create new buffer: %v", err)
	}

	expectedRemaining := 2
	if rb2.Count() != expectedRemaining {
		t.Fatalf("Expected %d restored items, got %d", expectedRemaining, rb2.Count())
	}
	t.Logf("✓ Restored %d metrics from disk", rb2.Count())

	// Step 7: Add more metrics
	t.Log("Step 7: Adding new metrics to restored buffer...")
	for i := 0; i < 3; i++ {
		stats := &pb.StatsRequest{
			Hostname: "server-02",
			AgentId:  "agent-002",
			Cpu:      45.5,
		}
		rb2.Push(stats)
	}

	if rb2.Count() != 5 {
		t.Fatalf("Expected 5 items, got %d", rb2.Count())
	}
	t.Logf("✓ Buffer now has %d metrics", rb2.Count())

	t.Log("✓ Integration test PASSED")
}

// TestIntegration_PersistenceWorker tests the persistence worker
func TestIntegration_PersistenceWorker(t *testing.T) {
	t.Log("Testing persistence worker...")

	dir := t.TempDir()
	rb, err := NewRingBuffer(10, dir)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Start persistence worker
	rb.StartPersistenceWorker(100 * time.Millisecond)
	t.Log("✓ Persistence worker started")

	// Add metrics
	for i := 0; i < 5; i++ {
		rb.Push(&pb.StatsRequest{
			Hostname: "server",
			Cpu:      float64(i * 10),
		})
	}

	// Wait for persistence
	time.Sleep(200 * time.Millisecond)

	// Check file was created
	filePath := rb.GetFilePath()
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("Persistence file not created: %v", err)
	}

	t.Log("✓ Persistence worker test PASSED")
}

// TestIntegration_BufferManager tests the buffer manager
func TestIntegration_BufferManager(t *testing.T) {
	t.Log("Testing buffer manager...")

	rb, err := NewRingBuffer(10, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Create mock client (nil for this test)
	bufferMgr := NewBufferManager(rb, nil, 3, 1*time.Second, 3)

	// Check initial status
	status := bufferMgr.GetBufferStatus()
	if status["count"].(int) != 0 {
		t.Error("Expected initial count 0")
	}
	if status["is_connected"].(bool) != true {
		t.Error("Expected initial connected state")
	}
	t.Log("✓ Initial status correct")

	// Simulate adding metrics
	for i := 0; i < 5; i++ {
		rb.Push(&pb.StatsRequest{Hostname: "server", Cpu: float64(i)})
	}

	// Check status after adding metrics
	status = bufferMgr.GetBufferStatus()
	if status["count"].(int) != 5 {
		t.Errorf("Expected count 5, got %d", status["count"].(int))
	}
	t.Logf("✓ Status: %d/10 (%.1f%%) buffered",
		status["count"], status["percentage"])

	t.Log("✓ Buffer manager test PASSED")
}

// TestIntegration_HighVolume tests with high volume of data
func TestIntegration_HighVolume(t *testing.T) {
	t.Log("Testing high volume of metrics...")

	rb, err := NewRingBuffer(1000, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	// Add 1500 metrics (exceeds capacity)
	t.Log("Adding 1500 metrics to buffer with capacity 1000...")
	start := time.Now()
	for i := 0; i < 1500; i++ {
		rb.Push(&pb.StatsRequest{
			Hostname: "server",
			AgentId:  "agent-001",
			Cpu:      float64(i % 100),
			Ram:      float64((i + 50) % 100),
		})
	}
	elapsed := time.Since(start)

	if rb.Count() != 1000 {
		t.Fatalf("Expected full buffer (1000), got %d", rb.Count())
	}
	if !rb.IsFull() {
		t.Fatal("Buffer should be full")
	}
	t.Logf("✓ Added 1500 metrics in %v", elapsed)

	// Pop batch multiple times
	t.Log("Popping data in batches...")
	batchCount := 0
	totalPopped := 0
	for {
		batch := rb.PopBatch(100)
		if len(batch) == 0 {
			break
		}
		batchCount++
		totalPopped += len(batch)
	}

	if totalPopped != 1000 {
		t.Fatalf("Expected to pop 1000, got %d", totalPopped)
	}
	if rb.Count() != 0 {
		t.Fatalf("Expected empty buffer, got %d items", rb.Count())
	}
	t.Logf("✓ Popped %d metrics in %d batches", totalPopped, batchCount)

	t.Log("✓ High volume test PASSED")
}

// TestIntegration_ConcurrentOperations tests thread safety
func TestIntegration_ConcurrentOperations(t *testing.T) {
	t.Log("Testing concurrent operations...")

	rb, err := NewRingBuffer(1000, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	done := make(chan bool)
	errors := make(chan error, 10)

	// Goroutine 1: Pushing data
	go func() {
		for i := 0; i < 100; i++ {
			rb.Push(&pb.StatsRequest{
				Hostname: "server",
				Cpu:      float64(i),
			})
		}
		done <- true
	}()

	// Goroutine 2: Popping data
	go func() {
		time.Sleep(50 * time.Millisecond) // Let some data accumulate
		for i := 0; i < 100; i++ {
			_ = rb.Pop()
		}
		done <- true
	}()

	// Goroutine 3: Peeking
	go func() {
		for i := 0; i < 100; i++ {
			_ = rb.Peek()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	if len(errors) > 0 {
		t.Fatalf("Got errors during concurrent operations: %v", <-errors)
	}

	t.Log("✓ Concurrent operations test PASSED")
}

// TestIntegration_Recovery tests recovery from persistence
func TestIntegration_Recovery(t *testing.T) {
	t.Log("Testing recovery from persistence...")

	dir := t.TempDir()

	// Create first buffer instance and add data
	t.Log("Creating first buffer instance...")
	rb1, err := NewRingBuffer(100, dir)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	for i := 0; i < 10; i++ {
		rb1.Push(&pb.StatsRequest{
			Hostname: "server-01",
			Cpu:      float64(i * 10),
		})
	}

	if err := rb1.ShutdownPersist(); err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}
	t.Logf("✓ Persisted %d metrics", rb1.Count())

	// Create second buffer instance and recover
	t.Log("Creating second buffer instance (recovery)...")
	rb2, err := NewRingBuffer(100, dir)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	if rb2.Count() != 10 {
		t.Fatalf("Expected 10 recovered items, got %d", rb2.Count())
	}
	t.Logf("✓ Recovered %d metrics", rb2.Count())

	// Verify first item
	first := rb2.Peek()
	if first.Stats.Cpu != 0 {
		t.Errorf("Expected first CPU 0, got %.1f", first.Stats.Cpu)
	}
	t.Log("✓ Recovered data integrity verified")

	t.Log("✓ Recovery test PASSED")
}
