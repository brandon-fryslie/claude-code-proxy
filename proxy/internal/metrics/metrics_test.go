package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordRequest(t *testing.T) {
	// Reset metrics before test
	RequestsTotal.Reset()

	// Record a request
	RecordRequest("test-provider", "test-model", "success", 1.5)

	// Verify counter was incremented
	count := testutil.ToFloat64(RequestsTotal.WithLabelValues("test-provider", "test-model", "success"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}
}

func TestUpdateCircuitBreakerState(t *testing.T) {
	// Reset metrics before test
	CircuitBreakerState.Reset()

	// Update state to open (1)
	UpdateCircuitBreakerState("test-provider", 1)

	// Verify gauge was set
	state := testutil.ToFloat64(CircuitBreakerState.WithLabelValues("test-provider"))
	if state != 1.0 {
		t.Errorf("Expected state 1.0 (open), got %f", state)
	}

	// Update state to half-open (2)
	UpdateCircuitBreakerState("test-provider", 2)

	state = testutil.ToFloat64(CircuitBreakerState.WithLabelValues("test-provider"))
	if state != 2.0 {
		t.Errorf("Expected state 2.0 (half-open), got %f", state)
	}

	// Update state to closed (0)
	UpdateCircuitBreakerState("test-provider", 0)

	state = testutil.ToFloat64(CircuitBreakerState.WithLabelValues("test-provider"))
	if state != 0.0 {
		t.Errorf("Expected state 0.0 (closed), got %f", state)
	}
}

func TestRecordFallback(t *testing.T) {
	// Reset metrics before test
	FallbackTotal.Reset()

	// Record a fallback
	RecordFallback("primary-provider", "fallback-provider")

	// Verify counter was incremented
	count := testutil.ToFloat64(FallbackTotal.WithLabelValues("primary-provider", "fallback-provider"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}

	// Record another fallback
	RecordFallback("primary-provider", "fallback-provider")

	count = testutil.ToFloat64(FallbackTotal.WithLabelValues("primary-provider", "fallback-provider"))
	if count != 2.0 {
		t.Errorf("Expected count 2.0, got %f", count)
	}
}

func TestRecordRetry(t *testing.T) {
	// Reset metrics before test
	RetryTotal.Reset()

	// Record a retry
	RecordRetry("test-provider")

	// Verify counter was incremented
	count := testutil.ToFloat64(RetryTotal.WithLabelValues("test-provider"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}

	// Record multiple retries
	RecordRetry("test-provider")
	RecordRetry("test-provider")

	count = testutil.ToFloat64(RetryTotal.WithLabelValues("test-provider"))
	if count != 3.0 {
		t.Errorf("Expected count 3.0, got %f", count)
	}
}

func TestRecordCircuitBreakerStateChange(t *testing.T) {
	// Reset metrics before test
	CircuitBreakerStateChanges.Reset()

	// Record a state change from closed to open
	RecordCircuitBreakerStateChange("test-provider", "closed", "open")

	// Verify counter was incremented
	count := testutil.ToFloat64(CircuitBreakerStateChanges.WithLabelValues("test-provider", "closed", "open"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}

	// Record another state change from open to half-open
	RecordCircuitBreakerStateChange("test-provider", "open", "half-open")

	count = testutil.ToFloat64(CircuitBreakerStateChanges.WithLabelValues("test-provider", "open", "half-open"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}

	// Record state change from half-open back to closed
	RecordCircuitBreakerStateChange("test-provider", "half-open", "closed")

	count = testutil.ToFloat64(CircuitBreakerStateChanges.WithLabelValues("test-provider", "half-open", "closed"))
	if count != 1.0 {
		t.Errorf("Expected count 1.0, got %f", count)
	}
}

func TestMultipleProviders(t *testing.T) {
	// Reset metrics before test
	RequestsTotal.Reset()

	// Record requests for multiple providers
	RecordRequest("provider1", "model1", "success", 1.0)
	RecordRequest("provider2", "model2", "success", 2.0)
	RecordRequest("provider1", "model1", "error", 1.5)

	// Verify each provider/model/status combination is tracked separately
	count1 := testutil.ToFloat64(RequestsTotal.WithLabelValues("provider1", "model1", "success"))
	if count1 != 1.0 {
		t.Errorf("Expected count 1.0 for provider1/model1/success, got %f", count1)
	}

	count2 := testutil.ToFloat64(RequestsTotal.WithLabelValues("provider2", "model2", "success"))
	if count2 != 1.0 {
		t.Errorf("Expected count 1.0 for provider2/model2/success, got %f", count2)
	}

	count3 := testutil.ToFloat64(RequestsTotal.WithLabelValues("provider1", "model1", "error"))
	if count3 != 1.0 {
		t.Errorf("Expected count 1.0 for provider1/model1/error, got %f", count3)
	}
}
