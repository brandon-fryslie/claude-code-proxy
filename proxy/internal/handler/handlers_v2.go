package handler

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

// ============================================================================
// Request Endpoints V2
// ============================================================================

// GetRequestsSummaryV2 returns array of request summaries directly (not wrapped)
func (h *Handler) GetRequestsSummaryV2(w http.ResponseWriter, r *http.Request) {
	modelFilter := r.URL.Query().Get("model")
	if modelFilter == "" {
		modelFilter = "all"
	}

	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	offset := 0
	limit := 100

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100000 {
			limit = parsed
		}
	}

	summaries, total, err := h.storageService.GetRequestsSummaryPaginated(modelFilter, startTime, endTime, offset, limit)
	if err != nil {
		log.Printf("Error getting request summaries: %v", err)
		writeErrorResponse(w, "Failed to get requests", http.StatusInternalServerError)
		return
	}

	// Set pagination headers
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
	w.Header().Set("X-Limit", strconv.Itoa(limit))

	// Return empty array instead of null
	if summaries == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	writeJSONResponse(w, summaries)
}

// GetRequestByIDV2 returns request directly (not wrapped in an object)
func (h *Handler) GetRequestByIDV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	if requestID == "" {
		writeErrorResponse(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	request, _, err := h.storageService.GetRequestByShortID(requestID)
	if err != nil {
		log.Printf("Error getting request by ID %s: %v", requestID, err)
		writeErrorResponse(w, "Failed to get request", http.StatusInternalServerError)
		return
	}

	if request == nil {
		writeErrorResponse(w, "Request not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, request)
}

// ============================================================================
// Conversation Endpoints V2
// ============================================================================

// GetConversationsV2 returns array of conversations from the database index - fast!
func (h *Handler) GetConversationsV2(w http.ResponseWriter, r *http.Request) {
	// Use the fast database-backed method
	conversations, err := h.storageService.GetIndexedConversations(100)
	if err != nil {
		log.Printf("❌ Error getting indexed conversations: %v", err)
		writeErrorResponse(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	if conversations == nil || len(conversations) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	writeJSONResponse(w, conversations)
}

// GetConversationByIDV2 returns conversation directly using session ID only
func (h *Handler) GetConversationByIDV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, ok := vars["id"]
	if !ok {
		writeErrorResponse(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	conversations, err := h.conversationService.GetConversations()
	if err != nil {
		log.Printf("❌ Error getting conversations: %v", err)
		writeErrorResponse(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	// Search through all conversations to find matching session ID
	for _, convs := range conversations {
		for _, conv := range convs {
			if conv.SessionID == sessionID {
				writeJSONResponse(w, conv)
				return
			}
		}
	}

	writeErrorResponse(w, "Conversation not found", http.StatusNotFound)
}

// ============================================================================
// Stats Endpoints V2
// ============================================================================

// GetHourlyStatsV2 returns hourly stats with consistent format (null arrays become empty)
func (h *Handler) GetHourlyStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		writeErrorResponse(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetHourlyStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting hourly stats: %v", err)
		writeErrorResponse(w, "Failed to get hourly stats", http.StatusInternalServerError)
		return
	}

	// Ensure HourlyStats is never null
	if stats != nil && stats.HourlyStats == nil {
		stats.HourlyStats = []model.HourlyTokens{}
	}

	writeJSONResponse(w, stats)
}

// GetModelStatsV2 returns model stats with null arrays as empty
func (h *Handler) GetModelStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		writeErrorResponse(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetModelStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting model stats: %v", err)
		writeErrorResponse(w, "Failed to get model stats", http.StatusInternalServerError)
		return
	}

	// Ensure ModelStats is never null
	if stats != nil && stats.ModelStats == nil {
		stats.ModelStats = []model.ModelTokens{}
	}

	writeJSONResponse(w, stats)
}

// GetSubagentStatsV2 returns subagent stats with null arrays as empty
func (h *Handler) GetSubagentStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		writeErrorResponse(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetSubagentStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting subagent stats: %v", err)
		writeErrorResponse(w, "Failed to get subagent stats", http.StatusInternalServerError)
		return
	}

	// Ensure Subagents is never null
	if stats != nil && stats.Subagents == nil {
		stats.Subagents = []model.SubagentStats{}
	}

	writeJSONResponse(w, stats)
}

// GetPerformanceStatsV2 returns performance stats with null arrays as empty
func (h *Handler) GetPerformanceStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		writeErrorResponse(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetPerformanceStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting performance stats: %v", err)
		writeErrorResponse(w, "Failed to get performance stats", http.StatusInternalServerError)
		return
	}

	// Ensure Stats is never null
	if stats != nil && stats.Stats == nil {
		stats.Stats = []model.PerformanceStats{}
	}

	writeJSONResponse(w, stats)
}

// GetWeeklyStatsV2 returns weekly stats with null arrays as empty
func (h *Handler) GetWeeklyStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		// Default to last 30 days if not specified
		now := time.Now()
		endTime = now.Format(time.RFC3339)
		startTime = now.AddDate(0, 0, -30).Format(time.RFC3339)
	}

	stats, err := h.storageService.GetStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting weekly stats: %v", err)
		writeErrorResponse(w, "Failed to get weekly stats", http.StatusInternalServerError)
		return
	}

	// Ensure DailyStats is never null
	if stats != nil && stats.DailyStats == nil {
		stats.DailyStats = []model.DailyTokens{}
	}

	writeJSONResponse(w, stats)
}

// ============================================================================
// Configuration API V2
// ============================================================================

// GetConfigV2 returns the full configuration (sanitized)
func (h *Handler) GetConfigV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	// Sanitize the config before returning
	sanitized := sanitizeConfig(h.config)
	writeJSONResponse(w, sanitized)
}

// GetProvidersV2 returns all provider configurations (sanitized)
func (h *Handler) GetProvidersV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	// Create sanitized provider map
	providers := make(map[string]*config.ProviderConfig)
	for name, provider := range h.config.Providers {
		providers[name] = &config.ProviderConfig{
			Format:     provider.Format,
			BaseURL:    provider.BaseURL,
			Version:    provider.Version,
			MaxRetries: provider.MaxRetries,
			APIKey:     redactAPIKey(provider.APIKey),
		}
	}

	// Return empty object if no providers (not null)
	if providers == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}

	writeJSONResponse(w, providers)
}

// GetSubagentConfigV2 returns subagent routing configuration
func (h *Handler) GetSubagentConfigV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	// Create response with subagent config
	subagentConfig := map[string]interface{}{
		"enable":   h.config.Subagents.Enable,
		"mappings": h.config.Subagents.Mappings,
	}

	// Ensure mappings is never null
	if subagentConfig["mappings"] == nil {
		subagentConfig["mappings"] = make(map[string]string)
	}

	writeJSONResponse(w, subagentConfig)
}

// ============================================================================
// Routing API V2 (NEW for Phase 4.1)
// ============================================================================

// GetRoutingConfigV2 returns current routing configuration including:
// - Provider mappings
// - Subagent routing rules
// - Circuit breaker settings
// - Fallback configuration
func (h *Handler) GetRoutingConfigV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	// Build routing configuration response
	routingConfig := map[string]interface{}{
		"providers": make(map[string]interface{}),
		"subagents": map[string]interface{}{
			"enable":   h.config.Subagents.Enable,
			"mappings": h.config.Subagents.Mappings,
		},
	}

	// Add provider routing details
	providers := make(map[string]interface{})
	for name, providerCfg := range h.config.Providers {
		providers[name] = map[string]interface{}{
			"format":            providerCfg.Format,
			"base_url":          providerCfg.BaseURL,
			"max_retries":       providerCfg.MaxRetries,
			"fallback_provider": providerCfg.FallbackProvider,
			"circuit_breaker": map[string]interface{}{
				"enabled":      providerCfg.CircuitBreaker.Enabled,
				"max_failures": providerCfg.CircuitBreaker.MaxFailures,
				"timeout":      providerCfg.CircuitBreaker.TimeoutDuration.String(),
			},
		}
	}
	routingConfig["providers"] = providers

	// Ensure mappings is never null
	if routingConfig["subagents"].(map[string]interface{})["mappings"] == nil {
		subagents := routingConfig["subagents"].(map[string]interface{})
		subagents["mappings"] = make(map[string]string)
	}

	writeJSONResponse(w, routingConfig)
}

// GetProviderStatusV2 returns real-time provider health status including:
// - Circuit breaker state (open/closed/half-open)
// - Fallback provider configuration
// - Health status
func (h *Handler) GetProviderStatusV2(w http.ResponseWriter, r *http.Request) {
	if h.modelRouter == nil {
		writeErrorResponse(w, "Model router not available", http.StatusInternalServerError)
		return
	}

	// Get provider health from model router
	providerHealth := h.modelRouter.GetProviderHealth()

	// Sort by name for consistent ordering
	sort.Slice(providerHealth, func(i, j int) bool {
		return providerHealth[i].Name < providerHealth[j].Name
	})

	writeJSONResponse(w, providerHealth)
}

// GetRoutingStatsV2 returns routing statistics including:
// - Requests per provider
// - Circuit breaker trips
// - Fallback activations
// - Average response times per provider
func (h *Handler) GetRoutingStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		// Default to last 24 hours if not specified
		now := time.Now()
		endTime = now.Format(time.RFC3339)
		startTime = now.AddDate(0, 0, -1).Format(time.RFC3339)
	}

	// Get provider stats from storage service
	providerStats, err := h.storageService.GetProviderStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting routing stats: %v", err)
		writeErrorResponse(w, "Failed to get routing stats", http.StatusInternalServerError)
		return
	}

	// Add subagent stats
	subagentStats, err := h.storageService.GetSubagentStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting subagent stats: %v", err)
		// Continue with just provider stats
		subagentStats = &model.SubagentStatsResponse{
			Subagents: []model.SubagentStats{},
		}
	}

	// Build comprehensive routing stats
	routingStats := map[string]interface{}{
		"providers": providerStats,
		"subagents": subagentStats,
		"timeRange": map[string]string{
			"start": startTime,
			"end":   endTime,
		},
	}

	writeJSONResponse(w, routingStats)
}

// ============================================================================
// Utility Functions
// ============================================================================

// sanitizeConfig creates a deep copy of the config with API keys redacted
func sanitizeConfig(cfg *config.Config) *config.Config {
	// Create a copy of the config
	sanitized := &config.Config{
		Server:  cfg.Server,
		Storage: cfg.Storage,
		Subagents: config.SubagentsConfig{
			Enable:   cfg.Subagents.Enable,
			Mappings: make(map[string]string),
		},
		Providers: make(map[string]*config.ProviderConfig),
	}

	// Copy subagent mappings (no sensitive data)
	for k, v := range cfg.Subagents.Mappings {
		sanitized.Subagents.Mappings[k] = v
	}

	// Copy and sanitize providers
	for name, provider := range cfg.Providers {
		sanitized.Providers[name] = &config.ProviderConfig{
			Format:     provider.Format,
			BaseURL:    provider.BaseURL,
			Version:    provider.Version,
			MaxRetries: provider.MaxRetries,
			// Redact API key if present
			APIKey: redactAPIKey(provider.APIKey),
		}
	}

	return sanitized
}

// redactAPIKey returns a redacted string if the API key is non-empty
func redactAPIKey(apiKey string) string {
	if apiKey != "" {
		return "***REDACTED***"
	}
	return ""
}
