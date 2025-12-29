package provider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 = no retries)
	MaxRetries int
	// InitialBackoff is the initial backoff duration (default: 1s)
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration (default: 30s)
	MaxBackoff time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// IsRetryableError determines if an error should be retried
// Only transient errors (5xx, timeout, connection errors) are retryable
func IsRetryableError(err error, statusCode int) bool {
	// Network/timeout errors are always retryable
	if err != nil {
		return true
	}

	// 5xx errors are retryable (server errors)
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// 429 (rate limit) is retryable
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	// 408 (request timeout) is retryable
	if statusCode == http.StatusRequestTimeout {
		return true
	}

	// 4xx errors (except above) are not retryable - they indicate client errors
	// 2xx/3xx are success/redirect - no retry needed
	return false
}

// RetryWithBackoff retries a function with exponential backoff
// The function fn should return (response, error, statusCode)
// Returns the response, error, and total number of attempts made
func RetryWithBackoff(
	ctx context.Context,
	config RetryConfig,
	fn func() (*http.Response, error),
) (*http.Response, error, int) {
	var lastErr error
	var lastResp *http.Response
	attempts := 0

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		attempts++

		// Try the function
		resp, err := fn()
		lastResp = resp
		lastErr = err

		// Determine status code for retry decision
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}

		// If successful (2xx) or non-retryable error, return immediately
		if err == nil && statusCode >= 200 && statusCode < 300 {
			return resp, nil, attempts
		}

		// Check if we should retry
		if !IsRetryableError(err, statusCode) {
			// Not retryable - return the error
			return resp, err, attempts
		}

		// Don't retry if this was the last attempt
		if attempt >= config.MaxRetries {
			break
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, config)

		// Check context cancellation before sleeping
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), attempts
		case <-time.After(backoff):
			// Continue to next retry attempt
		}

		// Close the response body if present (we're retrying)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return lastResp, fmt.Errorf("request failed after %d attempts: %w", attempts, lastErr), attempts
	}

	// Return the last response (non-2xx)
	return lastResp, nil, attempts
}

// calculateBackoff calculates the backoff duration for a given attempt
// Uses exponential backoff: initialBackoff * (multiplier ^ attempt)
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffMultiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	return time.Duration(backoff)
}
