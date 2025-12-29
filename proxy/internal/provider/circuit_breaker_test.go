package provider

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 3,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be Closed, got %v", cb.State())
	}

	if cb.Failures() != 0 {
		t.Errorf("Expected initial failures to be 0, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_SuccessfulCall(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 3,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	callCount := 0
	err := cb.Call(func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed after success, got %v", cb.State())
	}

	if cb.Failures() != 0 {
		t.Errorf("Expected failures to be 0 after success, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_FailuresAccumulate(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 3,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// First failure
	cb.Call(func() error {
		return testErr
	})

	if cb.Failures() != 1 {
		t.Errorf("Expected failures to be 1, got %d", cb.Failures())
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to still be Closed after 1 failure, got %v", cb.State())
	}

	// Second failure
	cb.Call(func() error {
		return testErr
	})

	if cb.Failures() != 2 {
		t.Errorf("Expected failures to be 2, got %d", cb.Failures())
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to still be Closed after 2 failures, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 3,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Trigger max failures
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected circuit to be Open after max failures, got %v", cb.State())
	}

	if cb.Failures() != 3 {
		t.Errorf("Expected failures to be 3, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_BlocksCallsWhenOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 2,
		Timeout:     100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Try to call - should be blocked
	callCount := 0
	err := cb.Call(func() error {
		callCount++
		return nil
	})

	if err == nil {
		t.Error("Expected error when circuit is open, got nil")
	}

	if callCount != 0 {
		t.Errorf("Expected function not to be called when circuit is open, but it was called %d times", callCount)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Try to call - should transition to half-open
	callCount := 0
	cb.Call(func() error {
		callCount++
		return nil
	})

	if callCount != 1 {
		t.Errorf("Expected function to be called in half-open state, got %d calls", callCount)
	}

	// Should be closed now (successful call in half-open)
	if cb.State() != StateClosed {
		t.Errorf("Expected circuit to be Closed after successful half-open call, got %v", cb.State())
	}

	if cb.Failures() != 0 {
		t.Errorf("Expected failures to be reset to 0, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_ReopensOnHalfOpenFailure(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Try to call with failure - should reopen
	cb.Call(func() error {
		return testErr
	})

	if cb.State() != StateOpen {
		t.Errorf("Expected circuit to reopen after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_ResetsFailuresOnSuccess(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 5,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Accumulate some failures
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.Failures() != 3 {
		t.Errorf("Expected 3 failures, got %d", cb.Failures())
	}

	// Successful call should reset
	cb.Call(func() error {
		return nil
	})

	if cb.Failures() != 0 {
		t.Errorf("Expected failures to be reset to 0 after success, got %d", cb.Failures())
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures: 2,
		Timeout:     10 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Manual reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed after reset, got %v", cb.State())
	}

	if cb.Failures() != 0 {
		t.Errorf("Expected failures to be 0 after reset, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_StateStringRepresentation(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("Expected %v.String() to be '%s', got '%s'", tt.state, tt.expected, got)
		}
	}
}
