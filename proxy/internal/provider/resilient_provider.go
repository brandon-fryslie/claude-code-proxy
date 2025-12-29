package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/metrics"
)

// ResilientProvider wraps a provider with circuit breaker, retry, and fallback logic
type ResilientProvider struct {
	name             string
	primaryProvider  Provider
	fallbackProvider Provider
	circuitBreaker   *CircuitBreaker
	retryConfig      RetryConfig
	config           *config.ProviderConfig
}

// NewResilientProvider creates a provider with resilience features
// If config has a fallback_provider set and circuit breaker enabled, those features are activated
func NewResilientProvider(
	name string,
	primaryProvider Provider,
	fallbackProvider Provider, // can be nil if no fallback configured
	cfg *config.ProviderConfig,
) Provider {
	rp := &ResilientProvider{
		name:             name,
		primaryProvider:  primaryProvider,
		fallbackProvider: fallbackProvider,
		config:           cfg,
	}

	// Initialize circuit breaker if enabled
	if cfg.CircuitBreaker.Enabled {
		cbConfig := CircuitBreakerConfig{
			MaxFailures: cfg.CircuitBreaker.MaxFailures,
			Timeout:     cfg.CircuitBreaker.TimeoutDuration,
		}
		rp.circuitBreaker = NewCircuitBreaker(cbConfig)

		// Set up circuit breaker state change callback to update metrics and log
		rp.circuitBreaker.SetStateChangeCallback(func(oldState, newState CircuitState) {
			// Update metrics
			metrics.UpdateCircuitBreakerState(name, int(newState))
			metrics.RecordCircuitBreakerStateChange(name, oldState.String(), newState.String())

			// Structured logging for circuit breaker state changes
			logEvent := map[string]interface{}{
				"event":      "circuit_breaker_state_change",
				"provider":   name,
				"old_state":  oldState.String(),
				"new_state":  newState.String(),
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}
			logJSON, _ := json.Marshal(logEvent)
			log.Printf("%s", logJSON)
		})

		// Initialize circuit breaker state metric
		metrics.UpdateCircuitBreakerState(name, int(StateClosed))
	}

	// Initialize retry config
	rp.retryConfig = RetryConfig{
		MaxRetries:        cfg.MaxRetries,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	return rp
}

// Name returns the provider name
func (rp *ResilientProvider) Name() string {
	return rp.name
}

// ForwardRequest forwards a request with circuit breaker, retry, and fallback logic
func (rp *ResilientProvider) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Try primary provider with circuit breaker and retry
	resp, err := rp.tryPrimaryProvider(ctx, req)

	// Record request metrics
	status := "success"
	if err != nil {
		status = "error"
	} else if resp != nil && resp.StatusCode >= 400 {
		status = fmt.Sprintf("http_%d", resp.StatusCode)
	}

	duration := time.Since(startTime).Seconds()
	model := "unknown"
	// Try to extract model from request context or headers
	if modelVal := req.Header.Get("X-Model"); modelVal != "" {
		model = modelVal
	}
	metrics.RecordRequest(rp.name, model, status, duration)

	// If primary succeeded or we don't have a fallback, return the result
	if err == nil || rp.fallbackProvider == nil {
		return resp, err
	}

	// If circuit breaker is open or primary failed after retries, try fallback
	log.Printf("⚠️ Provider '%s' failed, attempting fallback to '%s': %v",
		rp.name, rp.fallbackProvider.Name(), err)

	// Structured logging for fallback activation
	logEvent := map[string]interface{}{
		"event":           "fallback_activated",
		"from_provider":   rp.name,
		"to_provider":     rp.fallbackProvider.Name(),
		"reason":          err.Error(),
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}
	logJSON, _ := json.Marshal(logEvent)
	log.Printf("%s", logJSON)

	// Record fallback metric
	metrics.RecordFallback(rp.name, rp.fallbackProvider.Name())

	// Try fallback provider (without circuit breaker to avoid cascading failures)
	return rp.tryFallbackProvider(ctx, req)
}

// tryPrimaryProvider attempts to forward the request through the primary provider
// with circuit breaker protection and retry logic
func (rp *ResilientProvider) tryPrimaryProvider(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Function to execute through circuit breaker
	executeRequest := func() error {
		// Retry with exponential backoff
		var attempts int
		resp, err, attempts = RetryWithBackoff(ctx, rp.retryConfig, func() (*http.Response, error) {
			return rp.primaryProvider.ForwardRequest(ctx, req)
		})

		if attempts > 1 {
			log.Printf("📊 Provider '%s' request completed after %d attempts", rp.name, attempts)

			// Record retry attempts
			for i := 1; i < attempts; i++ {
				metrics.RecordRetry(rp.name)
			}

			// Structured logging for retry attempts
			logEvent := map[string]interface{}{
				"event":      "retry_attempts",
				"provider":   rp.name,
				"attempts":   attempts,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}
			logJSON, _ := json.Marshal(logEvent)
			log.Printf("%s", logJSON)
		}

		return err
	}

	// Execute through circuit breaker if enabled
	if rp.circuitBreaker != nil {
		cbErr := rp.circuitBreaker.Call(executeRequest)
		if cbErr != nil {
			// Circuit breaker error (circuit is open)
			if cbErr.Error() == "circuit breaker is open" {
				log.Printf("🔴 Circuit breaker OPEN for provider '%s' (too many failures)", rp.name)
				return nil, fmt.Errorf("circuit breaker is open for provider '%s': recent failures detected", rp.name)
			}
			// Other circuit breaker error
			return resp, cbErr
		}
	} else {
		// No circuit breaker - just execute with retry
		if err := executeRequest(); err != nil {
			return resp, err
		}
	}

	return resp, err
}

// tryFallbackProvider attempts to forward the request through the fallback provider
// This is called when the primary provider fails
func (rp *ResilientProvider) tryFallbackProvider(ctx context.Context, req *http.Request) (*http.Response, error) {
	log.Printf("🔄 Routing to fallback provider '%s'", rp.fallbackProvider.Name())

	// Forward to fallback provider
	// Note: Fallback provider may itself be a ResilientProvider with its own fallback chain
	resp, err := rp.fallbackProvider.ForwardRequest(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("fallback provider '%s' also failed: %w", rp.fallbackProvider.Name(), err)
	}

	return resp, nil
}

// GetCircuitBreakerState returns the current circuit breaker state
// Returns nil if circuit breaker is not enabled
func (rp *ResilientProvider) GetCircuitBreakerState() *CircuitState {
	if rp.circuitBreaker == nil {
		return nil
	}
	state := rp.circuitBreaker.State()
	return &state
}
