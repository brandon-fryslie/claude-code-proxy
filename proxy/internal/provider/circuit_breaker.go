package provider

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed means the circuit is functioning normally
	StateClosed CircuitState = iota
	// StateOpen means the circuit has detected too many failures and is blocking requests
	StateOpen
	// StateHalfOpen means the circuit is testing if the service has recovered
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening the circuit
	MaxFailures int
	// Timeout is how long to wait before transitioning from Open to HalfOpen
	Timeout time.Duration
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures: 5,
		Timeout:     30 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern for a provider
type CircuitBreaker struct {
	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	lastStateTime time.Time
	config        CircuitBreakerConfig
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:         StateClosed,
		failures:      0,
		lastStateTime: time.Now(),
		config:        config,
	}
}

// Call attempts to execute a function through the circuit breaker
// Returns an error if the circuit is open
func (cb *CircuitBreaker) Call(fn func() error) error {
	if err := cb.beforeCall(); err != nil {
		return err
	}

	err := fn()
	cb.afterCall(err)
	return err
}

// beforeCall checks if the circuit breaker allows the call
func (cb *CircuitBreaker) beforeCall() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Allow the call
		return nil

	case StateOpen:
		// Check if it's time to try again
		if time.Since(cb.lastStateTime) >= cb.config.Timeout {
			// Transition to half-open to test if the service has recovered
			cb.state = StateHalfOpen
			cb.lastStateTime = time.Now()
			return nil
		}
		// Circuit is still open, reject the call
		return fmt.Errorf("circuit breaker is open")

	case StateHalfOpen:
		// Allow the call to test if service has recovered
		return nil

	default:
		return fmt.Errorf("unknown circuit breaker state: %v", cb.state)
	}
}

// afterCall records the result of the call and updates circuit state
func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		// Success - reset failures and close circuit if needed
		cb.onSuccess()
	} else {
		// Failure - increment counter and possibly open circuit
		cb.onFailure()
	}
}

// onSuccess handles a successful call
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0

	if cb.state == StateHalfOpen {
		// Service has recovered, close the circuit
		cb.state = StateClosed
		cb.lastStateTime = time.Now()
	}
}

// onFailure handles a failed call
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == StateHalfOpen {
		// Failed during recovery test, reopen the circuit
		cb.state = StateOpen
		cb.lastStateTime = time.Now()
		return
	}

	if cb.failures >= cb.config.MaxFailures {
		// Too many failures, open the circuit
		cb.state = StateOpen
		cb.lastStateTime = time.Now()
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state
// This should be used sparingly, typically for testing or administrative actions
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.lastStateTime = time.Now()
}
