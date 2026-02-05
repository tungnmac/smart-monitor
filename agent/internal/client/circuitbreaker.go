package client

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "CLOSED"    // Normal operation
	StateOpen     CircuitBreakerState = "OPEN"      // Failing, reject requests
	StateHalfOpen CircuitBreakerState = "HALF_OPEN" // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// Configuration
	maxFailures      int           // Number of failures before opening
	successThreshold int           // Number of successes before closing (from half-open)
	timeout          time.Duration // Duration before attempting recovery
	name             string        // Circuit breaker name for logging
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureCount:     0,
		successCount:     0,
		maxFailures:      maxFailures,
		successThreshold: successThreshold,
		timeout:          timeout,
		name:             name,
		lastStateChange:  time.Now(),
	}
}

// Call executes a function through the circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	switch state {
	case StateClosed:
		return cb.callClosed(fn)
	case StateOpen:
		return cb.callOpen()
	case StateHalfOpen:
		return cb.callHalfOpen(fn)
	default:
		return fmt.Errorf("unknown circuit breaker state: %s", state)
	}
}

// callClosed handles calls when circuit is closed (normal operation)
func (cb *CircuitBreaker) callClosed(fn func() error) error {
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		log.Printf("[%s] ⚠️  Failure detected (count: %d/%d)", cb.name, cb.failureCount, cb.maxFailures)

		if cb.failureCount >= cb.maxFailures {
			cb.openCircuit()
		}
		return err
	}

	// Reset failure count on success
	cb.failureCount = 0
	return nil
}

// callOpen handles calls when circuit is open (rejecting requests)
func (cb *CircuitBreaker) callOpen() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if enough time has passed to attempt recovery
	if time.Since(cb.lastStateChange) >= cb.timeout {
		cb.transitionToHalfOpen()
		return fmt.Errorf("circuit breaker is HALF_OPEN, attempting recovery")
	}

	return fmt.Errorf("circuit breaker is OPEN, service is unavailable")
}

// callHalfOpen handles calls when circuit is half-open (testing recovery)
func (cb *CircuitBreaker) callHalfOpen(fn func() error) error {
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount = 1
		cb.openCircuit()
		return err
	}

	cb.successCount++
	log.Printf("[%s] ✓ Recovery attempt succeeded (%d/%d)", cb.name, cb.successCount, cb.successThreshold)

	if cb.successCount >= cb.successThreshold {
		cb.closeCircuit()
	}
	return nil
}

// openCircuit transitions to OPEN state
func (cb *CircuitBreaker) openCircuit() {
	cb.state = StateOpen
	cb.lastStateChange = time.Now()
	cb.successCount = 0
	log.Printf("[%s] 🔴 Circuit opened - service is unavailable (will retry in %s)", cb.name, cb.timeout)
}

// transitionToHalfOpen transitions to HALF_OPEN state
func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.state = StateHalfOpen
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	log.Printf("[%s] 🟡 Circuit half-open - attempting recovery", cb.name)
}

// closeCircuit transitions to CLOSED state
func (cb *CircuitBreaker) closeCircuit() {
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	log.Printf("[%s] 🟢 Circuit closed - service recovered", cb.name)
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
	log.Printf("[%s] 🔄 Circuit breaker reset", cb.name)
}
