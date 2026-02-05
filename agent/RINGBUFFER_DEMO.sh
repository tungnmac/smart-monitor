#!/bin/bash

# Ring Buffer Demo Script
# This script demonstrates how the ring buffer handles connection failures

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUFFER_DIR="${SCRIPT_DIR}/.agent_buffer"
LOG_DIR="${SCRIPT_DIR}/logs"

echo "=== Ring Buffer Connection Failure Demo ==="
echo ""

# Create necessary directories
mkdir -p "$BUFFER_DIR" "$LOG_DIR"

echo "📦 Ring Buffer Demo Scenarios"
echo "=============================="
echo ""

echo "Scenario 1: Normal operation (all data sent immediately)"
echo "─────────────────────────────────────────────────────"
echo "• Agent collects metrics every 2 seconds"
echo "• Backend is available"
echo "• All data is sent directly"
echo "• Buffer remains empty (0/1000)"
echo ""

echo "Scenario 2: Backend temporarily unavailable"
echo "──────────────────────────────────────────"
echo "• Agent loses connection to backend"
echo "• Agent detects connection loss (gRPC error)"
echo "• Metrics are automatically buffered locally"
echo "• Buffer fills up: 50/1000 → 100/1000 → 500/1000"
echo ""

echo "Scenario 3: Connection restored"
echo "───────────────────────────────"
echo "• Connection monitor detects backend is back"
echo "• All buffered metrics are sent in batches of 50"
echo "• Buffer is flushed: 500/1000 → 450/1000 → ... → 0/1000"
echo "• Normal operation resumes"
echo ""

echo "Scenario 4: Agent shutdown with pending data"
echo "────────────────────────────────────────────"
echo "• Remaining buffered metrics are flushed first"
echo "• All data is persisted to: $BUFFER_DIR/buffer_data.json"
echo "• On next startup, data is automatically restored"
echo ""

echo "Buffer Configuration:"
echo "───────────────────"
grep -E "bufferSize|bufferDataDir|batchSize" "$SCRIPT_DIR/main.go" | \
  sed 's/const.*= //' | \
  sed 's/ *\/\/.*//' | \
  awk '{print "  " $0}'
echo ""

echo "Key Features:"
echo "─────────────"
echo "✓ Automatic buffering on connection loss"
echo "✓ Persistent storage (survives restarts)"
echo "✓ Batch flushing for efficiency"
echo "✓ Connection monitoring (5-second interval)"
echo "✓ FIFO queue (oldest data sent first)"
echo "✓ Thread-safe operations"
echo "✓ Status reporting every 30 seconds"
echo ""

echo "Testing Ring Buffer:"
echo "───────────────────"
echo "Run: cd $SCRIPT_DIR && go test -v ./internal/buffer"
echo ""

echo "Files Created:"
echo "──────────────"
echo "• $SCRIPT_DIR/internal/buffer/ringbuffer.go     - Ring buffer implementation"
echo "• $SCRIPT_DIR/internal/buffer/manager.go         - Buffer manager & retry logic"
echo "• $SCRIPT_DIR/internal/buffer/ringbuffer_test.go - Unit tests"
echo "• $SCRIPT_DIR/RINGBUFFER.md                      - Documentation"
echo ""

echo "Monitoring Buffer in Production:"
echo "───────────────────────────────"
echo "Check log output for buffer status:"
echo "  📊 Buffer Status: 45/1000 (4.5%) | Connected: true"
echo ""
echo "Or programmatically:"
echo "  status := bufferMgr.GetBufferStatus()"
echo "  count := status[\"count\"]"
echo "  capacity := status[\"capacity\"]"
echo ""

echo "Demo Complete! ✓"
