package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/provider"
)

// testProvider implements provider.Provider for testing
type testProvider struct {
	name string
}

func (m *testProvider) Name() string {
	return m.name
}

func (m *testProvider) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestPreferenceRouter_SelectProvider(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	// Create test providers
	providers := map[string]provider.Provider{
		"fast-provider":    &testProvider{name: "fast-provider"},
		"cheap-provider":   &testProvider{name: "cheap-provider"},
		"quality-provider": &testProvider{name: "quality-provider"},
	}

	// Create routing config with provider profiles
	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles: map[string]ProviderProfile{
			"fast-provider": {
				Speed:   9,
				Cost:    5,
				Quality: 6,
			},
			"cheap-provider": {
				Speed:   5,
				Cost:    9,
				Quality: 5,
			},
			"quality-provider": {
				Speed:   6,
				Cost:    5,
				Quality: 9,
			},
		},
		Tasks: make(map[string]TaskPreference),
	}

	router := NewPreferenceRouter(routingCfg, nil, providers, logger)

	// Test that all providers are selected (load balancing works)
	selections := make(map[string]int)

	for i := 0; i < 100; i++ {
		providerName, _ := router.SelectProvider("test-task", PreferenceBalanced, "test-model")
		selections[providerName]++
	}

	// All providers should be selected at least once (load balancing across top 3)
	if selections["fast-provider"] == 0 {
		t.Error("fast-provider should be selected at least once")
	}
	if selections["cheap-provider"] == 0 {
		t.Error("cheap-provider should be selected at least once")
	}
	if selections["quality-provider"] == 0 {
		t.Error("quality-provider should be selected at least once")
	}

	// Distribution should be relatively even (within 20% of each other)
	total := 100
	expectedPerProvider := total / 3
	for provider, count := range selections {
		if count < expectedPerProvider-10 || count > expectedPerProvider+10 {
			t.Errorf("Provider %s has unbalanced selection count: %d (expected around %d)", provider, count, expectedPerProvider)
		}
	}
}

func TestPreferenceRouter_TaskPreference(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	// Create routing config with task-specific preferences
	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles: map[string]ProviderProfile{
			"fast-provider":  {Speed: 9, Cost: 5, Quality: 6},
			"cheap-provider": {Speed: 5, Cost: 9, Quality: 5},
		},
		Tasks: map[string]TaskPreference{
			"code_generation": {
				Preference: PreferenceSpeed,
				Providers:  []string{"fast-provider"},
			},
			"budget_tasks": {
				Preference: PreferenceCost,
				Providers:  []string{"cheap-provider"},
			},
		},
	}

	providers := map[string]provider.Provider{
		"fast-provider":  &testProvider{name: "fast-provider"},
		"cheap-provider": &testProvider{name: "cheap-provider"},
	}

	router := NewPreferenceRouter(routingCfg, nil, providers, logger)

	tests := []struct {
		name         string
		task         string
		wantProvider string
	}{
		{
			name:         "Code generation uses fast provider",
			task:         "code_generation",
			wantProvider: "fast-provider",
		},
		{
			name:         "Budget tasks use cheap provider",
			task:         "budget_tasks",
			wantProvider: "cheap-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerName, _ := router.SelectProvider(tt.task, PreferenceBalanced, "test-model")

			if providerName != tt.wantProvider {
				t.Errorf("Expected %s for task %s, got %s", tt.wantProvider, tt.task, providerName)
			}
		})
	}
}

func TestPreferenceRouter_HealthBasedRouting(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	// Create resilient providers with circuit breakers
	cfg1 := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: "http://test1",
		CircuitBreaker: config.CircuitBreakerConfig{
			Enabled:     true,
			MaxFailures: 2,
		},
	}

	cfg2 := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: "http://test2",
	}

	mockProv1 := &testProvider{name: "provider1"}
	mockProv2 := &testProvider{name: "provider2"}

	resilient1 := provider.NewResilientProvider("provider1", mockProv1, nil, cfg1)
	resilient2 := provider.NewResilientProvider("provider2", mockProv2, nil, cfg2)

	providers := map[string]provider.Provider{
		"provider1": resilient1,
		"provider2": resilient2,
	}

	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles: map[string]ProviderProfile{
			"provider1": {Speed: 9, Cost: 9, Quality: 9},
			"provider2": {Speed: 5, Cost: 5, Quality: 5},
		},
		Tasks: make(map[string]TaskPreference),
	}

	router := NewPreferenceRouter(routingCfg, nil, providers, logger)

	// Test that healthy providers are selected
	providerName, _ := router.SelectProvider("test", PreferenceBalanced, "test-model")
	if providerName == "" {
		t.Error("Expected a provider to be selected")
	}
}

func TestPreferenceRouter_NoHealthyProviders(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	// Empty providers map
	providers := make(map[string]provider.Provider)

	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles:  make(map[string]ProviderProfile),
		Tasks:             make(map[string]TaskPreference),
	}

	router := NewPreferenceRouter(routingCfg, nil, providers, logger)

	providerName, model := router.SelectProvider("test", PreferenceBalanced, "test-model")

	if providerName != "" {
		t.Errorf("Expected empty provider name when no providers available, got %s", providerName)
	}
	if model != "" {
		t.Errorf("Expected empty model when no providers available, got %s", model)
	}
}

func TestPreferenceRouter_CalculateScore(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles:  make(map[string]ProviderProfile),
		Tasks:             make(map[string]TaskPreference),
	}

	router := NewPreferenceRouter(routingCfg, nil, make(map[string]provider.Provider), logger)

	profile := ProviderProfile{
		Speed:   8,
		Cost:    6,
		Quality: 9,
	}

	tests := []struct {
		name       string
		preference Preference
		wantScore  int
	}{
		{
			name:       "Speed preference",
			preference: PreferenceSpeed,
			wantScore:  8*2 + 9, // Speed*2 + Quality
		},
		{
			name:       "Cost preference",
			preference: PreferenceCost,
			wantScore:  6*2 + 9, // Cost*2 + Quality
		},
		{
			name:       "Quality preference",
			preference: PreferenceQuality,
			wantScore:  9 * 3, // Quality*3
		},
		{
			name:       "Balanced preference",
			preference: PreferenceBalanced,
			wantScore:  8 + 6 + 9, // Speed + Cost + Quality
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := router.calculateScore(profile, tt.preference)
			if score != tt.wantScore {
				t.Errorf("Expected score %d for %s preference, got %d", tt.wantScore, tt.preference, score)
			}
		})
	}
}

func TestPreferenceRouter_RankingLogic(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	providers := map[string]provider.Provider{
		"fast":    &testProvider{name: "fast"},
		"cheap":   &testProvider{name: "cheap"},
		"quality": &testProvider{name: "quality"},
	}

	routingCfg := &RoutingConfig{
		DefaultPreference: PreferenceBalanced,
		ProviderProfiles: map[string]ProviderProfile{
			"fast":    {Speed: 10, Cost: 5, Quality: 6},
			"cheap":   {Speed: 5, Cost: 10, Quality: 5},
			"quality": {Speed: 6, Cost: 5, Quality: 10},
		},
		Tasks: make(map[string]TaskPreference),
	}

	router := NewPreferenceRouter(routingCfg, nil, providers, logger)

	tests := []struct {
		name          string
		preference    Preference
		expectedFirst string
	}{
		{
			name:          "Speed ranks fast highest",
			preference:    PreferenceSpeed,
			expectedFirst: "fast", // 10*2 + 6 = 26
		},
		{
			name:          "Cost ranks cheap highest",
			preference:    PreferenceCost,
			expectedFirst: "cheap", // 10*2 + 5 = 25
		},
		{
			name:          "Quality ranks quality highest",
			preference:    PreferenceQuality,
			expectedFirst: "quality", // 10*3 = 30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranked := router.rankProviders([]string{"fast", "cheap", "quality"}, tt.preference)

			if len(ranked) == 0 {
				t.Fatal("Expected non-empty ranking")
			}

			if ranked[0] != tt.expectedFirst {
				t.Errorf("Expected %s to be ranked first for %s preference, got %s\nFull ranking: %v",
					tt.expectedFirst, tt.preference, ranked[0], ranked)
			}
		})
	}
}
