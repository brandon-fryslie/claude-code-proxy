package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestIsRetryableError_NetworkErrors(t *testing.T) {
	testErr := errors.New("network error")
	if !IsRetryableError(testErr, 0) {
		t.Error("Network errors should be retryable")
	}
}

func TestIsRetryableError_5xxErrors(t *testing.T) {
	tests := []int{500, 502, 503, 504}
	for _, statusCode := range tests {
		if !IsRetryableError(nil, statusCode) {
			t.Errorf("Status %d should be retryable", statusCode)
		}
	}
}

func TestIsRetryableError_RateLimitAndTimeout(t *testing.T) {
	retryableStatuses := []int{
		http.StatusTooManyRequests, // 429
		http.StatusRequestTimeout,  // 408
	}

	for _, status := range retryableStatuses {
		if !IsRetryableError(nil, status) {
			t.Errorf("Status %d should be retryable", status)
		}
	}
}

func TestIsRetryableError_NonRetryable4xxErrors(t *testing.T) {
	nonRetryableStatuses := []int{
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotFound,            // 404
		http.StatusMethodNotAllowed,    // 405
		http.StatusUnprocessableEntity, // 422
	}

	for _, status := range nonRetryableStatuses {
		if IsRetryableError(nil, status) {
			t.Errorf("Status %d should NOT be retryable", status)
		}
	}
}

func TestIsRetryableError_SuccessStatuses(t *testing.T) {
	successStatuses := []int{200, 201, 204}

	for _, status := range successStatuses {
		if IsRetryableError(nil, status) {
			t.Errorf("Success status %d should NOT be retryable", status)
		}
	}
}

func TestRetryWithBackoff_SuccessFirstAttempt(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	resp, err, attempts := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		return &http.Response{StatusCode: 200}, nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil || resp.StatusCode != 200 {
		t.Error("Expected successful response")
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithBackoff_RetriesOnFailure(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	testErr := errors.New("network error")

	start := time.Now()
	_, err, attempts := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		if callCount < 3 {
			return nil, testErr
		}
		return &http.Response{StatusCode: 200}, nil
	})
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Should have some backoff delay (at least 10ms initial + 20ms second retry = 30ms)
	if duration < 30*time.Millisecond {
		t.Errorf("Expected some backoff delay, but got %v", duration)
	}
}

func TestRetryWithBackoff_DoesNotRetryNonRetryableErrors(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	resp, err, attempts := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		// Return 400 Bad Request (not retryable)
		return &http.Response{StatusCode: 400}, nil
	})

	if err != nil {
		t.Errorf("Expected no error (just non-2xx response), got %v", err)
	}

	if resp == nil || resp.StatusCode != 400 {
		t.Error("Expected 400 response")
	}

	if callCount != 1 {
		t.Errorf("Expected only 1 call (no retries for 400), got %d", callCount)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithBackoff_RespectsMaxRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	testErr := errors.New("persistent error")

	_, err, attempts := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		return nil, testErr
	})

	if err == nil {
		t.Error("Expected error after exhausting retries")
	}

	// Should be initial attempt + MaxRetries
	expectedAttempts := 1 + config.MaxRetries
	if callCount != expectedAttempts {
		t.Errorf("Expected %d calls, got %d", expectedAttempts, callCount)
	}

	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxRetries:        5,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	testErr := errors.New("error")

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err, _ := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		return nil, testErr
	})
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	// Should fail quickly due to context cancellation
	if duration > 200*time.Millisecond {
		t.Errorf("Expected quick failure due to context cancellation, but took %v", duration)
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},  // 1 * 2^0 = 1
		{1, 2 * time.Second},  // 1 * 2^1 = 2
		{2, 4 * time.Second},  // 1 * 2^2 = 4
		{3, 8 * time.Second},  // 1 * 2^3 = 8
		{4, 16 * time.Second}, // 1 * 2^4 = 16
		{5, 30 * time.Second}, // 1 * 2^5 = 32, capped at 30
		{6, 30 * time.Second}, // 1 * 2^6 = 64, capped at 30
	}

	for _, tt := range tests {
		result := calculateBackoff(tt.attempt, config)
		if result != tt.expected {
			t.Errorf("Attempt %d: expected backoff %v, got %v", tt.attempt, tt.expected, result)
		}
	}
}

func TestRetryWithBackoff_Retries5xxErrors(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	callCount := 0
	_, err, attempts := RetryWithBackoff(ctx, config, func() (*http.Response, error) {
		callCount++
		if callCount < 3 {
			// Return 503 Service Unavailable (retryable)
			return &http.Response{StatusCode: 503}, nil
		}
		return &http.Response{StatusCode: 200}, nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}
