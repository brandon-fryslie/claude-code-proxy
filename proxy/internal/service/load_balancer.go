package service

import (
	"sync"
)

// LoadBalancer implements weighted round-robin load balancing
type LoadBalancer struct {
	weights map[string]int // provider -> weight (higher = more requests)
	current map[string]int // provider -> current request count
	mu      sync.Mutex     // Protects current counts
}

// NewLoadBalancer creates a load balancer with provider weights
func NewLoadBalancer(weights map[string]int) *LoadBalancer {
	return &LoadBalancer{
		weights: weights,
		current: make(map[string]int),
	}
}

// SelectProvider returns the next provider using weighted round-robin
// The algorithm distributes requests proportionally to weights
func (lb *LoadBalancer) SelectProvider(available []string) string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(available) == 0 {
		return ""
	}

	// Single provider - no need for balancing
	if len(available) == 1 {
		return available[0]
	}

	// Calculate total weight for available providers
	totalWeight := 0
	for _, provider := range available {
		weight := lb.getWeight(provider)
		totalWeight += weight
	}

	if totalWeight == 0 {
		// All weights are zero, use simple round-robin
		return lb.roundRobin(available)
	}

	// Weighted round-robin: find provider with lowest (current / weight) ratio
	var selectedProvider string
	minRatio := -1.0

	for _, provider := range available {
		weight := lb.getWeight(provider)
		if weight == 0 {
			continue // Skip providers with zero weight
		}

		currentCount := lb.current[provider]
		ratio := float64(currentCount) / float64(weight)

		if minRatio < 0 || ratio < minRatio {
			minRatio = ratio
			selectedProvider = provider
		}
	}

	// If no provider selected (all had zero weight), fallback to simple round-robin
	if selectedProvider == "" {
		return lb.roundRobin(available)
	}

	// Increment count for selected provider
	lb.current[selectedProvider]++

	// Periodically reset counts to prevent overflow
	// Reset when total count exceeds 1000 * number of providers
	totalCount := 0
	for _, count := range lb.current {
		totalCount += count
	}
	if totalCount > 1000*len(lb.current) {
		lb.resetCounts()
	}

	return selectedProvider
}

// getWeight returns the weight for a provider, defaulting to 1 if not set
func (lb *LoadBalancer) getWeight(provider string) int {
	if weight, exists := lb.weights[provider]; exists {
		return weight
	}
	return 1
}

// roundRobin implements simple round-robin selection (no weighting)
func (lb *LoadBalancer) roundRobin(available []string) string {
	// Find provider with minimum count
	minCount := -1
	var selected string

	for _, provider := range available {
		count := lb.current[provider]
		if minCount < 0 || count < minCount {
			minCount = count
			selected = provider
		}
	}

	if selected != "" {
		lb.current[selected]++
	}

	return selected
}

// resetCounts proportionally reduces all counts to prevent overflow
func (lb *LoadBalancer) resetCounts() {
	// Divide all counts by 2
	for provider := range lb.current {
		lb.current[provider] = lb.current[provider] / 2
	}
}

// UpdateWeights updates the weights for providers
// This can be called to adjust load balancing based on performance metrics
func (lb *LoadBalancer) UpdateWeights(weights map[string]int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for provider, weight := range weights {
		lb.weights[provider] = weight
	}
}

// GetStats returns current request counts for each provider
func (lb *LoadBalancer) GetStats() map[string]int {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	stats := make(map[string]int)
	for provider, count := range lb.current {
		stats[provider] = count
	}

	return stats
}
