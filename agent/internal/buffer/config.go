package buffer

import (
	"time"
)

// Config holds configuration for ring buffer and manager
type Config struct {
	// Ring Buffer configuration
	BufferSize      int
	BufferDataDir   string
	PersistInterval time.Duration

	// Manager configuration
	MaxRetries       int
	RetryInterval    time.Duration
	BatchSize        int
	MaxBatchWait     time.Duration
	ConnectionTicker time.Duration

	// Status reporting
	StatusInterval time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		// Ring Buffer
		BufferSize:      1000,
		BufferDataDir:   ".agent_buffer",
		PersistInterval: 5 * time.Second,

		// Manager
		MaxRetries:       5,
		RetryInterval:    5 * time.Second,
		BatchSize:        50,
		MaxBatchWait:     5 * time.Second,
		ConnectionTicker: 5 * time.Second,

		// Status
		StatusInterval: 30 * time.Second,
	}
}

// ProductionConfig returns optimized config for production
func ProductionConfig() Config {
	return Config{
		// Larger buffer for production
		BufferSize:      5000,
		BufferDataDir:   "/var/lib/agent/buffer",
		PersistInterval: 10 * time.Second,

		// More aggressive retry
		MaxRetries:       10,
		RetryInterval:    3 * time.Second,
		BatchSize:        100,
		MaxBatchWait:     10 * time.Second,
		ConnectionTicker: 3 * time.Second,

		// Less frequent status reporting
		StatusInterval: 60 * time.Second,
	}
}

// DevelopmentConfig returns config for development
func DevelopmentConfig() Config {
	return Config{
		// Smaller buffer for development
		BufferSize:      100,
		BufferDataDir:   ".agent_buffer",
		PersistInterval: 2 * time.Second,

		// Quicker feedback in development
		MaxRetries:       3,
		RetryInterval:    2 * time.Second,
		BatchSize:        10,
		MaxBatchWait:     2 * time.Second,
		ConnectionTicker: 2 * time.Second,

		// Frequent status for debugging
		StatusInterval: 10 * time.Second,
	}
}
