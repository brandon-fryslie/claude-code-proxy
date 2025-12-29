package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for the proxy
var (
	// RequestsTotal counts total requests by provider, model, and status
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of requests processed by the proxy",
		},
		[]string{"provider", "model", "status"},
	)

	// RequestDuration tracks request duration by provider
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	// CircuitBreakerState tracks circuit breaker state (0=closed, 1=open, 2=half-open)
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"provider"},
	)

	// FallbackTotal counts fallback activations
	FallbackTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_fallback_total",
			Help: "Total number of fallback activations",
		},
		[]string{"from_provider", "to_provider"},
	)

	// RetryTotal counts retry attempts
	RetryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_retry_total",
			Help: "Total number of retry attempts",
		},
		[]string{"provider"},
	)

	// CircuitBreakerStateChanges counts state transitions
	CircuitBreakerStateChanges = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_circuit_breaker_state_changes_total",
			Help: "Total number of circuit breaker state changes",
		},
		[]string{"provider", "from_state", "to_state"},
	)
)

// RecordRequest records a completed request
func RecordRequest(provider, model, status string, duration float64) {
	RequestsTotal.WithLabelValues(provider, model, status).Inc()
	RequestDuration.WithLabelValues(provider).Observe(duration)
}

// UpdateCircuitBreakerState updates the circuit breaker state gauge
func UpdateCircuitBreakerState(provider string, state int) {
	CircuitBreakerState.WithLabelValues(provider).Set(float64(state))
}

// RecordFallback records a fallback activation
func RecordFallback(fromProvider, toProvider string) {
	FallbackTotal.WithLabelValues(fromProvider, toProvider).Inc()
}

// RecordRetry records a retry attempt
func RecordRetry(provider string) {
	RetryTotal.WithLabelValues(provider).Inc()
}

// RecordCircuitBreakerStateChange records a circuit breaker state transition
func RecordCircuitBreakerStateChange(provider, fromState, toState string) {
	CircuitBreakerStateChanges.WithLabelValues(provider, fromState, toState).Inc()
}
