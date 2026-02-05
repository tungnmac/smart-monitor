package client

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries        int           // Maximum number of retry attempts
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffMultiplier float64       // Exponential backoff multiplier
	JitterFraction    float64       // Jitter fraction (0-1) to add randomness
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        10,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        60 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	}
}

// ConnectionManager manages gRPC connection with retry and circuit breaker
type ConnectionManager struct {
	address        string
	retryConfig    RetryConfig
	circuitBreaker *CircuitBreaker
	connection     *grpc.ClientConn
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(address string, retryConfig RetryConfig) *ConnectionManager {
	return &ConnectionManager{
		address:     address,
		retryConfig: retryConfig,
		circuitBreaker: NewCircuitBreaker(
			"gRPC-Backend",
			5,              // maxFailures
			3,              // successThreshold
			30*time.Second, // timeout
		),
	}
}

// Connect establishes a gRPC connection with retry and circuit breaker
func (cm *ConnectionManager) Connect() (*grpc.ClientConn, error) {
	var lastErr error
	var attempt int

	for attempt = 0; attempt < cm.retryConfig.MaxRetries; attempt++ {
		// Check circuit breaker state
		state := cm.circuitBreaker.GetState()
		if state == StateOpen {
			// Wait before attempting recovery
			waitTime := time.Until(time.Now().Add(30 * time.Second))
			log.Printf("⏳ Circuit breaker OPEN - waiting %v before retry attempt %d/%d",
				waitTime, attempt+1, cm.retryConfig.MaxRetries)
		} else if state == StateHalfOpen {
			log.Printf("🟡 Circuit breaker HALF_OPEN - attempting recovery (attempt %d/%d)",
				attempt+1, cm.retryConfig.MaxRetries)
		} else {
			log.Printf("🔌 Attempting connection to %s (attempt %d/%d)",
				cm.address, attempt+1, cm.retryConfig.MaxRetries)
		}

		// Try to connect through circuit breaker
		err := cm.circuitBreaker.Call(func() error {
			conn, err := grpc.Dial(
				cm.address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock(),
				grpc.WithTimeout(5*time.Second),
			)
			if err == nil {
				cm.connection = conn
			}
			return err
		})

		if err == nil {
			log.Printf("✅ Successfully connected to backend at %s after %d attempt(s)", cm.address, attempt+1)
			return cm.connection, nil
		}

		lastErr = err

		// Calculate backoff duration with jitter
		backoff := cm.calculateBackoff(attempt)
		jitter := time.Duration(float64(backoff) * cm.retryConfig.JitterFraction * (rand.Float64()*2 - 1))
		totalWait := backoff + jitter
		if totalWait < 0 {
			totalWait = backoff
		}

		log.Printf("❌ Connection attempt %d failed: %v", attempt+1, err)
		log.Printf("⏱️  Waiting %v before next retry (exponential backoff)...", totalWait)

		// Wait before retry
		time.Sleep(totalWait)
	}

	// All retries exhausted
	log.Printf("❌ Failed to connect after %d attempts: %v", cm.retryConfig.MaxRetries, lastErr)
	return nil, fmt.Errorf("failed to connect to %s after %d retries: %w",
		cm.address, cm.retryConfig.MaxRetries, lastErr)
}

// calculateBackoff calculates backoff duration with exponential increase
func (cm *ConnectionManager) calculateBackoff(attempt int) time.Duration {
	backoff := time.Duration(
		float64(cm.retryConfig.InitialBackoff) *
			math.Pow(cm.retryConfig.BackoffMultiplier, float64(attempt)),
	)

	if backoff > cm.retryConfig.MaxBackoff {
		backoff = cm.retryConfig.MaxBackoff
	}

	return backoff
}

// GetConnection returns the current connection
func (cm *ConnectionManager) GetConnection() *grpc.ClientConn {
	return cm.connection
}

// IsConnected checks if the connection is active
func (cm *ConnectionManager) IsConnected() bool {
	return cm.connection != nil
}

// Close closes the gRPC connection
func (cm *ConnectionManager) Close() error {
	if cm.connection != nil {
		return cm.connection.Close()
	}
	return nil
}

// Reset resets the connection manager and circuit breaker
func (cm *ConnectionManager) Reset() {
	cm.circuitBreaker.Reset()
	if cm.connection != nil {
		cm.connection.Close()
	}
	cm.connection = nil
}

// GetCircuitBreakerState returns the current circuit breaker state
func (cm *ConnectionManager) GetCircuitBreakerState() CircuitBreakerState {
	return cm.circuitBreaker.GetState()
}

// ReconnectWithBackoff attempts to reconnect with exponential backoff
// This is useful for background recovery attempts
func (cm *ConnectionManager) ReconnectWithBackoff() (*grpc.ClientConn, error) {
	log.Println("🔄 Starting automatic reconnection with exponential backoff...")
	return cm.Connect()
}
