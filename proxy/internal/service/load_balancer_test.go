package service

import (
	"testing"
)

func TestLoadBalancer_WeightedRoundRobin(t *testing.T) {
	weights := map[string]int{
		"provider1": 3, // Should get ~75% of requests
		"provider2": 1, // Should get ~25% of requests
	}

	lb := NewLoadBalancer(weights)

	// Track selections over 100 requests
	selections := make(map[string]int)
	available := []string{"provider1", "provider2"}

	for i := 0; i < 100; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// Verify distribution is roughly proportional to weights
	// provider1 should get ~75 requests, provider2 ~25
	ratio1 := float64(selections["provider1"]) / 100.0
	ratio2 := float64(selections["provider2"]) / 100.0

	// Allow 10% tolerance
	if ratio1 < 0.65 || ratio1 > 0.85 {
		t.Errorf("provider1 got %d%% of requests, expected ~75%%", int(ratio1*100))
	}
	if ratio2 < 0.15 || ratio2 > 0.35 {
		t.Errorf("provider2 got %d%% of requests, expected ~25%%", int(ratio2*100))
	}
}

func TestLoadBalancer_EqualWeights(t *testing.T) {
	weights := map[string]int{
		"provider1": 1,
		"provider2": 1,
		"provider3": 1,
	}

	lb := NewLoadBalancer(weights)

	selections := make(map[string]int)
	available := []string{"provider1", "provider2", "provider3"}

	// Select 90 times (30 per provider expected)
	for i := 0; i < 90; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// Each provider should get roughly 1/3 of requests
	for provider, count := range selections {
		expected := 30
		// Allow ±5 tolerance
		if count < expected-5 || count > expected+5 {
			t.Errorf("provider %s got %d requests, expected ~%d", provider, count, expected)
		}
	}
}

func TestLoadBalancer_SingleProvider(t *testing.T) {
	weights := map[string]int{
		"only-provider": 1,
	}

	lb := NewLoadBalancer(weights)
	available := []string{"only-provider"}

	// Select multiple times
	for i := 0; i < 10; i++ {
		provider := lb.SelectProvider(available)
		if provider != "only-provider" {
			t.Errorf("Expected only-provider, got %s", provider)
		}
	}
}

func TestLoadBalancer_EmptyAvailable(t *testing.T) {
	weights := map[string]int{
		"provider1": 1,
	}

	lb := NewLoadBalancer(weights)
	available := []string{}

	provider := lb.SelectProvider(available)
	if provider != "" {
		t.Errorf("Expected empty string for empty available list, got %s", provider)
	}
}

func TestLoadBalancer_ZeroWeights(t *testing.T) {
	weights := map[string]int{
		"provider1": 0,
		"provider2": 0,
	}

	lb := NewLoadBalancer(weights)
	available := []string{"provider1", "provider2"}

	selections := make(map[string]int)

	// Should fallback to round-robin
	for i := 0; i < 10; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// Both should get roughly equal requests (round-robin fallback)
	if selections["provider1"] != 5 || selections["provider2"] != 5 {
		t.Errorf("Expected equal distribution with zero weights, got %v", selections)
	}
}

func TestLoadBalancer_MixedWeights(t *testing.T) {
	weights := map[string]int{
		"provider1": 5,
		"provider2": 3,
		"provider3": 2,
	}

	lb := NewLoadBalancer(weights)
	available := []string{"provider1", "provider2", "provider3"}

	selections := make(map[string]int)

	// Select 100 times
	for i := 0; i < 100; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// Verify proportions (5:3:2 ratio)
	// provider1: ~50%, provider2: ~30%, provider3: ~20%
	ratio1 := float64(selections["provider1"]) / 100.0
	ratio2 := float64(selections["provider2"]) / 100.0
	ratio3 := float64(selections["provider3"]) / 100.0

	if ratio1 < 0.40 || ratio1 > 0.60 {
		t.Errorf("provider1 got %d%% of requests, expected ~50%%", int(ratio1*100))
	}
	if ratio2 < 0.20 || ratio2 > 0.40 {
		t.Errorf("provider2 got %d%% of requests, expected ~30%%", int(ratio2*100))
	}
	if ratio3 < 0.10 || ratio3 > 0.30 {
		t.Errorf("provider3 got %d%% of requests, expected ~20%%", int(ratio3*100))
	}
}

func TestLoadBalancer_UpdateWeights(t *testing.T) {
	weights := map[string]int{
		"provider1": 1,
		"provider2": 1,
	}

	lb := NewLoadBalancer(weights)
	available := []string{"provider1", "provider2"}

	// Select a few times with equal weights
	for i := 0; i < 10; i++ {
		lb.SelectProvider(available)
	}

	// Update weights to favor provider1
	lb.UpdateWeights(map[string]int{
		"provider1": 9,
		"provider2": 1,
	})

	selections := make(map[string]int)

	// Select 100 more times
	for i := 0; i < 100; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// provider1 should now get ~90% of requests
	ratio1 := float64(selections["provider1"]) / 100.0

	if ratio1 < 0.80 || ratio1 > 0.95 {
		t.Errorf("After weight update, provider1 got %d%% of requests, expected ~90%%", int(ratio1*100))
	}
}

func TestLoadBalancer_GetStats(t *testing.T) {
	weights := map[string]int{
		"provider1": 1,
		"provider2": 1,
	}

	lb := NewLoadBalancer(weights)
	available := []string{"provider1", "provider2"}

	// Select a few times
	for i := 0; i < 6; i++ {
		lb.SelectProvider(available)
	}

	stats := lb.GetStats()

	// Should have counts for both providers
	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 providers, got %d", len(stats))
	}

	total := stats["provider1"] + stats["provider2"]
	if total != 6 {
		t.Errorf("Expected total count of 6, got %d", total)
	}
}

func TestLoadBalancer_SubsetAvailable(t *testing.T) {
	weights := map[string]int{
		"provider1": 3,
		"provider2": 2,
		"provider3": 1,
	}

	lb := NewLoadBalancer(weights)

	// Only make provider1 and provider3 available
	available := []string{"provider1", "provider3"}

	selections := make(map[string]int)

	// Select 40 times
	for i := 0; i < 40; i++ {
		provider := lb.SelectProvider(available)
		selections[provider]++
	}

	// provider2 should not be selected at all
	if selections["provider2"] != 0 {
		t.Errorf("provider2 should not be selected when not in available list, got %d", selections["provider2"])
	}

	// provider1 should get ~75% (weight 3), provider3 ~25% (weight 1)
	ratio1 := float64(selections["provider1"]) / 40.0
	ratio3 := float64(selections["provider3"]) / 40.0

	if ratio1 < 0.65 || ratio1 > 0.85 {
		t.Errorf("provider1 got %d%% of requests, expected ~75%%", int(ratio1*100))
	}
	if ratio3 < 0.15 || ratio3 > 0.35 {
		t.Errorf("provider3 got %d%% of requests, expected ~25%%", int(ratio3*100))
	}
}
