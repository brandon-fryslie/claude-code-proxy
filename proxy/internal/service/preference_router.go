package service

import (
	"log"

	"github.com/seifghazi/claude-code-monitor/internal/provider"
)

// Preference represents routing optimization goals
type Preference string

const (
	PreferenceCost     Preference = "cost"
	PreferenceSpeed    Preference = "speed"
	PreferenceQuality  Preference = "quality"
	PreferenceBalanced Preference = "balanced"
)

// ProviderProfile describes characteristics of a provider for routing decisions
type ProviderProfile struct {
	Speed   int // 1-10 scale, higher is faster
	Cost    int // 1-10 scale, higher is more cost-effective
	Quality int // 1-10 scale, higher is better quality
}

// TaskPreference defines routing preference for a specific task type
type TaskPreference struct {
	Preference Preference
	Providers  []string // Ordered list of preferred providers
}

// RoutingConfig holds preference-based routing configuration
type RoutingConfig struct {
	DefaultPreference Preference                  // Default routing preference
	Tasks             map[string]TaskPreference   // Per-task preferences
	ProviderProfiles  map[string]ProviderProfile  // Provider characteristics
}

// PreferenceRouter selects providers based on routing preferences
type PreferenceRouter struct {
	config       *RoutingConfig
	modelRouter  *ModelRouter
	providers    map[string]provider.Provider
	loadBalancer *LoadBalancer
	logger       *log.Logger
}

// NewPreferenceRouter creates a new preference-based router
func NewPreferenceRouter(
	cfg *RoutingConfig,
	modelRouter *ModelRouter,
	providers map[string]provider.Provider,
	logger *log.Logger,
) *PreferenceRouter {
	// Initialize load balancer with provider weights based on profiles
	weights := make(map[string]int)
	for name, profile := range cfg.ProviderProfiles {
		// Default weight based on balanced preference
		weights[name] = (profile.Speed + profile.Cost + profile.Quality) / 3
	}

	// For providers without profiles, assign default weight
	for name := range providers {
		if _, exists := weights[name]; !exists {
			weights[name] = 5 // Default middle weight
		}
	}

	return &PreferenceRouter{
		config:       cfg,
		modelRouter:  modelRouter,
		providers:    providers,
		loadBalancer: NewLoadBalancer(weights),
		logger:       logger,
	}
}

// SelectProvider chooses the best provider based on preference
// Returns provider name and model name
func (r *PreferenceRouter) SelectProvider(task string, preference Preference, model string) (string, string) {
	// Get task-specific preference if available
	taskPref := r.GetTaskPreference(task)
	if taskPref.Preference != "" {
		preference = taskPref.Preference
	}

	// Get available providers based on task preference
	candidateProviders := r.getCandidateProviders(taskPref)
	if len(candidateProviders) == 0 {
		// Fallback to all available providers
		candidateProviders = r.getAllHealthyProviders()
	}

	// Filter out unhealthy providers
	healthyProviders := r.filterHealthyProviders(candidateProviders)
	if len(healthyProviders) == 0 {
		r.logger.Printf("⚠️ No healthy providers available for preference '%s'", preference)
		return "", ""
	}

	// Rank providers by preference
	rankedProviders := r.rankProviders(healthyProviders, preference)

	// Use load balancer to select from top-ranked providers
	// Take top 3 providers or all if less than 3
	topN := 3
	if len(rankedProviders) < topN {
		topN = len(rankedProviders)
	}
	topProviders := rankedProviders[:topN]

	// Load balance across top providers
	selectedProvider := r.loadBalancer.SelectProvider(topProviders)

	return selectedProvider, model
}

// GetTaskPreference returns the preference for a given task type
func (r *PreferenceRouter) GetTaskPreference(task string) TaskPreference {
	if pref, exists := r.config.Tasks[task]; exists {
		return pref
	}
	return TaskPreference{
		Preference: r.config.DefaultPreference,
	}
}

// getCandidateProviders returns providers configured for a task
func (r *PreferenceRouter) getCandidateProviders(taskPref TaskPreference) []string {
	if len(taskPref.Providers) > 0 {
		return taskPref.Providers
	}
	return nil
}

// getAllHealthyProviders returns all available providers
func (r *PreferenceRouter) getAllHealthyProviders() []string {
	providers := make([]string, 0, len(r.providers))
	for name := range r.providers {
		providers = append(providers, name)
	}
	return providers
}

// filterHealthyProviders excludes providers with open circuit breakers
func (r *PreferenceRouter) filterHealthyProviders(candidates []string) []string {
	healthy := make([]string, 0, len(candidates))

	for _, name := range candidates {
		prov, exists := r.providers[name]
		if !exists {
			continue
		}

		// Check if this is a ResilientProvider with circuit breaker
		if resilient, ok := prov.(*provider.ResilientProvider); ok {
			if state := resilient.GetCircuitBreakerState(); state != nil {
				// Exclude providers with open circuit breakers
				if *state == provider.StateOpen {
					r.logger.Printf("⚠️ Excluding provider '%s' (circuit breaker is open)", name)
					continue
				}
			}
		}

		healthy = append(healthy, name)
	}

	return healthy
}

// rankProviders orders providers by preference score
func (r *PreferenceRouter) rankProviders(providers []string, preference Preference) []string {
	// Calculate scores for each provider
	type scoredProvider struct {
		name  string
		score int
	}

	scored := make([]scoredProvider, 0, len(providers))

	for _, name := range providers {
		profile, exists := r.config.ProviderProfiles[name]
		if !exists {
			// Default profile for providers without configuration
			profile = ProviderProfile{
				Speed:   5,
				Cost:    5,
				Quality: 5,
			}
		}

		score := r.calculateScore(profile, preference)
		scored = append(scored, scoredProvider{name: name, score: score})
	}

	// Sort by score (descending)
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Extract provider names
	result := make([]string, len(scored))
	for i, s := range scored {
		result[i] = s.name
	}

	return result
}

// calculateScore computes a provider's score for a given preference
func (r *PreferenceRouter) calculateScore(profile ProviderProfile, preference Preference) int {
	switch preference {
	case PreferenceCost:
		// Prioritize cost, with quality as secondary factor
		return profile.Cost*2 + profile.Quality
	case PreferenceSpeed:
		// Prioritize speed, with quality as secondary factor
		return profile.Speed*2 + profile.Quality
	case PreferenceQuality:
		// Prioritize quality above all
		return profile.Quality * 3
	case PreferenceBalanced:
		// Equal weighting of all factors
		return profile.Speed + profile.Cost + profile.Quality
	default:
		// Default to balanced
		return profile.Speed + profile.Cost + profile.Quality
	}
}
