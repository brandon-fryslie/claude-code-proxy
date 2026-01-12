package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
	"github.com/seifghazi/claude-code-monitor/internal/service"
)

// DataHandler handles all data and dashboard API endpoints:
// - /api/requests/* - Request data endpoints
// - /api/stats/* - Statistics endpoints
// - /api/conversations/* - Conversation endpoints
// - /api/v2/* - V2 API endpoints
// - /api/config/* - Configuration endpoints
//
// It has full dependencies: read/write storage, conversation service, logger, config.
// This handler contains all the dashboard and analytics functionality.
type DataHandler struct {
	storageService      service.StorageService
	conversationService service.ConversationService
	indexer             *service.ConversationIndexer
	logger              *log.Logger
	config              *config.Config
}

// NewDataHandler creates a new DataHandler with the required dependencies.
func NewDataHandler(storageService service.StorageService, logger *log.Logger, cfg *config.Config) *DataHandler {
	conversationService := service.NewConversationService()

	return &DataHandler{
		storageService:      storageService,
		conversationService: conversationService,
		logger:              logger,
		config:              cfg,
	}
}

// SetIndexer sets the conversation indexer (for health checks).
func (h *DataHandler) SetIndexer(indexer *service.ConversationIndexer) {
	h.indexer = indexer
}

// Health handles the /health endpoint for proxy-data.
func (h *DataHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	dbStatus := "connected"
	if h.storageService == nil {
		dbStatus = "disconnected"
	}

	// Check indexer status
	indexerStatus := "not_configured"
	if h.indexer != nil {
		indexerStatus = "running"
	}

	response := map[string]interface{}{
		"status":    "ok",
		"service":   "proxy-data",
		"database":  dbStatus,
		"indexer":   indexerStatus,
		"timestamp": time.Now(),
	}

	writeJSONResponse(w, response)
}

// UI serves the UI page.
func (h *DataHandler) UI(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("index.html")
	if err != nil {
		http.Error(w, "UI not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(htmlContent)
}

// NotFound handles 404 responses.
func (h *DataHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, "Not found", http.StatusNotFound)
}

// ============================================================================
// Request Endpoints
// ============================================================================

// GetRequests returns paginated requests.
func (h *DataHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	modelFilter := r.URL.Query().Get("model")
	if modelFilter == "" {
		modelFilter = "all"
	}

	allRequests, err := h.storageService.GetAllRequests(modelFilter)
	if err != nil {
		log.Printf("Error getting requests: %v", err)
		http.Error(w, "Failed to get requests", http.StatusInternalServerError)
		return
	}

	requests := make([]model.RequestLog, len(allRequests))
	for i, req := range allRequests {
		if req != nil {
			requests[i] = *req
		}
	}

	total := len(requests)

	start := (page - 1) * limit
	end := start + limit
	if start >= len(requests) {
		requests = []model.RequestLog{}
	} else {
		if end > len(requests) {
			end = len(requests)
		}
		requests = requests[start:end]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Requests []model.RequestLog `json:"requests"`
		Total    int                `json:"total"`
	}{
		Requests: requests,
		Total:    total,
	})
}

// GetRequestsSummary returns lightweight request data for fast list rendering.
func (h *DataHandler) GetRequestsSummary(w http.ResponseWriter, r *http.Request) {
	modelFilter := r.URL.Query().Get("model")
	if modelFilter == "" {
		modelFilter = "all"
	}

	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	offset := 0
	limit := 0

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
		http.Error(w, "Failed to get requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Requests []*model.RequestSummary `json:"requests"`
		Total    int                     `json:"total"`
		Offset   int                     `json:"offset"`
		Limit    int                     `json:"limit"`
	}{
		Requests: summaries,
		Total:    total,
		Offset:   offset,
		Limit:    limit,
	})
}

// GetRequestByID returns a single request by its ID.
func (h *DataHandler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	if requestID == "" {
		http.Error(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	request, fullID, err := h.storageService.GetRequestByShortID(requestID)
	if err != nil {
		log.Printf("Error getting request by ID %s: %v", requestID, err)
		http.Error(w, "Failed to get request", http.StatusInternalServerError)
		return
	}

	if request == nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Request *model.RequestLog `json:"request"`
		FullID  string            `json:"fullId"`
	}{
		Request: request,
		FullID:  fullID,
	})
}

// GetLatestRequestDate returns the date of the most recent request.
func (h *DataHandler) GetLatestRequestDate(w http.ResponseWriter, r *http.Request) {
	latestDate, err := h.storageService.GetLatestRequestDate()
	if err != nil {
		log.Printf("Error getting latest request date: %v", err)
		http.Error(w, "Failed to get latest request date", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"latestDate": latestDate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteRequests clears all requests.
func (h *DataHandler) DeleteRequests(w http.ResponseWriter, r *http.Request) {
	clearedCount, err := h.storageService.ClearRequests()
	if err != nil {
		log.Printf("Error clearing requests: %v", err)
		writeErrorResponse(w, "Error clearing request history", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Request history cleared",
		"deleted": clearedCount,
	}

	writeJSONResponse(w, response)
}

// ============================================================================
// Stats Endpoints
// ============================================================================

// GetStats returns aggregated dashboard statistics.
func (h *DataHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		now := time.Now().UTC()
		endTime = now.Format(time.RFC3339)
		startTime = now.AddDate(0, 0, -7).Format(time.RFC3339)
	}

	stats, err := h.storageService.GetStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetHourlyStats returns hourly breakdown for a specific date range.
func (h *DataHandler) GetHourlyStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetHourlyStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting hourly stats: %v", err)
		http.Error(w, "Failed to get hourly stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetModelStats returns model breakdown for a specific date range.
func (h *DataHandler) GetModelStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetModelStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting model stats: %v", err)
		http.Error(w, "Failed to get model stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetProviderStats returns analytics broken down by provider.
func (h *DataHandler) GetProviderStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetProviderStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting provider stats: %v", err)
		http.Error(w, "Failed to get provider stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetSubagentStats returns analytics broken down by subagent.
func (h *DataHandler) GetSubagentStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetSubagentStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting subagent stats: %v", err)
		http.Error(w, "Failed to get subagent stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetToolStats returns analytics broken down by tool usage.
func (h *DataHandler) GetToolStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetToolStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting tool stats: %v", err)
		http.Error(w, "Failed to get tool stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetPerformanceStats returns response time analytics with percentiles.
func (h *DataHandler) GetPerformanceStats(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	stats, err := h.storageService.GetPerformanceStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting performance stats: %v", err)
		http.Error(w, "Failed to get performance stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ============================================================================
// Conversation Endpoints
// ============================================================================

// GetConversations returns all conversations.
func (h *DataHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.conversationService.GetConversations()
	if err != nil {
		log.Printf("❌ Error getting conversations: %v", err)
		writeErrorResponse(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	var allConversations []map[string]interface{}
	for _, convs := range conversations {
		for _, conv := range convs {
			var firstMessage string
			for _, msg := range conv.Messages {
				if msg.Type == "user" {
					text := extractTextFromMessage(msg.Message)
					if text != "" {
						firstMessage = text
						if len(firstMessage) > 200 {
							firstMessage = firstMessage[:200] + "..."
						}
						break
					}
				}
			}

			allConversations = append(allConversations, map[string]interface{}{
				"id":           conv.SessionID,
				"requestCount": conv.MessageCount,
				"startTime":    conv.StartTime.Format(time.RFC3339),
				"lastActivity": conv.EndTime.Format(time.RFC3339),
				"duration":     conv.EndTime.Sub(conv.StartTime).Milliseconds(),
				"firstMessage": firstMessage,
				"projectName":  conv.ProjectName,
			})
		}
	}

	sort.Slice(allConversations, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, allConversations[i]["lastActivity"].(string))
		t2, _ := time.Parse(time.RFC3339, allConversations[j]["lastActivity"].(string))
		return t1.After(t2)
	})

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(allConversations) {
		allConversations = []map[string]interface{}{}
	} else {
		if end > len(allConversations) {
			end = len(allConversations)
		}
		allConversations = allConversations[start:end]
	}

	response := map[string]interface{}{
		"conversations": allConversations,
	}

	writeJSONResponse(w, response)
}

// GetConversationByID returns a conversation by its session ID.
func (h *DataHandler) GetConversationByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, ok := vars["id"]
	if !ok {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	projectPath := r.URL.Query().Get("project")
	if projectPath == "" {
		http.Error(w, "Project path is required", http.StatusBadRequest)
		return
	}

	conversation, err := h.conversationService.GetConversation(projectPath, sessionID)
	if err != nil {
		log.Printf("❌ Error getting conversation: %v", err)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, conversation)
}

// GetConversationsByProject returns conversations for a specific project.
func (h *DataHandler) GetConversationsByProject(w http.ResponseWriter, r *http.Request) {
	projectPath := r.URL.Query().Get("project")
	if projectPath == "" {
		http.Error(w, "Project path is required", http.StatusBadRequest)
		return
	}

	conversations, err := h.conversationService.GetConversationsByProject(projectPath)
	if err != nil {
		log.Printf("❌ Error getting project conversations: %v", err)
		writeErrorResponse(w, "Failed to get project conversations", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, conversations)
}

// SearchConversations performs full-text search on conversation content.
func (h *DataHandler) SearchConversations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeErrorResponse(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	projectPath := r.URL.Query().Get("project")

	opts := model.SearchOptions{
		Query:       query,
		ProjectPath: projectPath,
		Limit:       limit,
		Offset:      offset,
	}

	log.Printf("🔍 Searching conversations: query=%q, project=%q, limit=%d, offset=%d", query, projectPath, limit, offset)

	results, err := h.storageService.SearchConversations(opts)
	if err != nil {
		log.Printf("❌ Error searching conversations (query=%q, project=%q): %v", query, projectPath, err)
		writeErrorResponse(w, "Failed to search conversations", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Search completed: found %d results (total: %d)", len(results.Results), results.Total)
	writeJSONResponse(w, results)
}

// ============================================================================
// V2 API Endpoints
// ============================================================================

// GetRequestsSummaryV2 returns array of request summaries directly.
func (h *DataHandler) GetRequestsSummaryV2(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
	w.Header().Set("X-Limit", strconv.Itoa(limit))

	if summaries == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	writeJSONResponse(w, summaries)
}

// GetRequestByIDV2 returns request directly (not wrapped).
func (h *DataHandler) GetRequestByIDV2(w http.ResponseWriter, r *http.Request) {
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

// GetConversationsV2 returns array of conversations from the database index - fast!
func (h *DataHandler) GetConversationsV2(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 GetConversationsV2 called - requesting limit 100")
	// Use the fast database-backed method
	conversations, err := h.storageService.GetIndexedConversations(100)
	if err != nil {
		log.Printf("❌ Error getting indexed conversations: %v", err)
		writeErrorResponse(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Got %d conversations from GetIndexedConversations", len(conversations))

	if conversations == nil || len(conversations) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	writeJSONResponse(w, conversations)
}

// GetConversationByIDV2 returns conversation directly using session ID only.
// Uses indexed database lookup for fast retrieval instead of scanning all files.
func (h *DataHandler) GetConversationByIDV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, ok := vars["id"]
	if !ok {
		writeErrorResponse(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Fast path: look up file path from database index
	filePath, projectPath, err := h.storageService.GetConversationFilePath(sessionID)
	if err != nil {
		log.Printf("⚠️ Conversation %s not in index, falling back to scan: %v", sessionID, err)
		// Fallback to slow scan for conversations not yet indexed
		h.getConversationByIDFallback(w, sessionID)
		return
	}

	// Load the specific conversation file directly
	conversation, err := h.conversationService.GetConversation(projectPath, sessionID)
	if err != nil {
		log.Printf("❌ Error loading conversation from %s: %v", filePath, err)
		writeErrorResponse(w, "Failed to load conversation", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, conversation)
}

// getConversationByIDFallback scans all conversations when index lookup fails
func (h *DataHandler) getConversationByIDFallback(w http.ResponseWriter, sessionID string) {
	conversations, err := h.conversationService.GetConversations()
	if err != nil {
		log.Printf("❌ Error getting conversations: %v", err)
		writeErrorResponse(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

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

// GetConversationMessagesV2 returns conversation messages from the database.
// This is faster than reading from files as messages are pre-indexed.
func (h *DataHandler) GetConversationMessagesV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID, ok := vars["id"]
	if !ok {
		writeErrorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	// Parse pagination params
	limit := 100
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, total, err := h.storageService.GetConversationMessages(conversationID, limit, offset)
	if err != nil {
		log.Printf("❌ Error getting conversation messages: %v", err)
		writeErrorResponse(w, "Failed to get conversation messages", http.StatusInternalServerError)
		return
	}

	response := model.ConversationMessagesResponse{
		ConversationID: conversationID,
		Messages:       messages,
		Total:          total,
		Offset:         offset,
		Limit:          limit,
	}

	writeJSONResponse(w, response)
}

// ReindexConversationsV2 triggers a re-index of all conversations.
func (h *DataHandler) ReindexConversationsV2(w http.ResponseWriter, r *http.Request) {
	if err := h.storageService.ReindexConversations(); err != nil {
		log.Printf("❌ Error triggering re-index: %v", err)
		writeErrorResponse(w, "Failed to trigger re-index", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{
		"status":  "ok",
		"message": "Re-indexing triggered. Conversations will be re-indexed in the background.",
	})
}

// GetHourlyStatsV2 returns hourly stats with consistent format.
func (h *DataHandler) GetHourlyStatsV2(w http.ResponseWriter, r *http.Request) {
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

	if stats != nil && stats.HourlyStats == nil {
		stats.HourlyStats = []model.HourlyTokens{}
	}

	writeJSONResponse(w, stats)
}

// GetModelStatsV2 returns model stats with null arrays as empty.
func (h *DataHandler) GetModelStatsV2(w http.ResponseWriter, r *http.Request) {
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

	if stats != nil && stats.ModelStats == nil {
		stats.ModelStats = []model.ModelTokens{}
	}

	writeJSONResponse(w, stats)
}

// GetProvidersV2 returns all provider configurations (sanitized).
func (h *DataHandler) GetProvidersV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

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

	if providers == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}

	writeJSONResponse(w, providers)
}

// GetSubagentStatsV2 returns subagent stats with null arrays as empty.
func (h *DataHandler) GetSubagentStatsV2(w http.ResponseWriter, r *http.Request) {
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

	if stats != nil && stats.Subagents == nil {
		stats.Subagents = []model.SubagentStats{}
	}

	writeJSONResponse(w, stats)
}

// GetPerformanceStatsV2 returns performance stats with null arrays as empty.
func (h *DataHandler) GetPerformanceStatsV2(w http.ResponseWriter, r *http.Request) {
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

	if stats != nil && stats.Stats == nil {
		stats.Stats = []model.PerformanceStats{}
	}

	writeJSONResponse(w, stats)
}

// GetWeeklyStatsV2 returns weekly stats with null arrays as empty.
func (h *DataHandler) GetWeeklyStatsV2(w http.ResponseWriter, r *http.Request) {
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	if startTime == "" || endTime == "" {
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

	if stats != nil && stats.DailyStats == nil {
		stats.DailyStats = []model.DailyTokens{}
	}

	writeJSONResponse(w, stats)
}

// GetConfigV2 returns the full configuration (sanitized).
func (h *DataHandler) GetConfigV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	sanitized := sanitizeConfig(h.config)
	writeJSONResponse(w, sanitized)
}

// GetSubagentConfigV2 returns subagent routing configuration.
func (h *DataHandler) GetSubagentConfigV2(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeErrorResponse(w, "Configuration not available", http.StatusInternalServerError)
		return
	}

	subagentConfig := map[string]interface{}{
		"enable":   h.config.Subagents.Enable,
		"mappings": h.config.Subagents.Mappings,
	}

	if subagentConfig["mappings"] == nil {
		subagentConfig["mappings"] = make(map[string]string)
	}

	writeJSONResponse(w, subagentConfig)
}
