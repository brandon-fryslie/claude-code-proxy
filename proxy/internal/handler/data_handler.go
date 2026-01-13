package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
// Supports ?include_subagents=true to merge subagent messages with parent conversation.
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

	// Check for include_subagents parameter
	includeSubagents := r.URL.Query().Get("include_subagents") == "true"

	var messages []*model.DBConversationMessage
	var total int
	var err error

	if includeSubagents {
		messages, total, err = h.storageService.GetConversationMessagesWithSubagents(conversationID, limit, offset)
	} else {
		messages, total, err = h.storageService.GetConversationMessages(conversationID, limit, offset)
	}

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

// ============================================================================
// CC-VIZ Claude Directory Endpoints
// ============================================================================

// GetClaudeConfigV2 returns the user's ~/.claude configuration files
func (h *DataHandler) GetClaudeConfigV2(w http.ResponseWriter, r *http.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		writeErrorResponse(w, "Could not determine home directory", http.StatusInternalServerError)
		return
	}
	claudeDir := filepath.Join(homeDir, ".claude")

	response := make(map[string]interface{})

	// Read settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if settingsData, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]interface{}
		if err := json.Unmarshal(settingsData, &settings); err == nil {
			// Parse permissions into groups
			permissions := parsePermissions(settings)
			plugins := parsePlugins(settings)

			response["settings"] = map[string]interface{}{
				"model":        settings["model"],
				"default_mode": getNestedString(settings, "permissions", "defaultMode"),
				"permissions":  permissions,
				"plugins":      plugins,
				"raw":          settings,
			}
		}
	} else {
		response["settings"] = nil
		response["settings_error"] = "File not found or not readable"
	}

	// Read CLAUDE.md (follow symlinks automatically via ReadFile)
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if claudeMdData, err := os.ReadFile(claudeMdPath); err == nil {
		claudeMdContent := string(claudeMdData)
		sections := parseClaudeMdSections(claudeMdContent)
		response["claude_md"] = map[string]interface{}{
			"content":  claudeMdContent,
			"sections": sections,
		}
	} else {
		response["claude_md"] = nil
		response["claude_md_error"] = "File not found or not readable"
	}

	// Read .mcp.json
	mcpPath := filepath.Join(claudeDir, ".mcp.json")
	if mcpData, err := os.ReadFile(mcpPath); err == nil {
		var mcpConfig map[string]interface{}
		if err := json.Unmarshal(mcpData, &mcpConfig); err == nil {
			servers := parseMCPServers(mcpConfig)
			response["mcp_config"] = map[string]interface{}{
				"servers": servers,
				"raw":     mcpConfig,
			}
		}
	} else {
		response["mcp_config"] = nil
		response["mcp_config_error"] = "File not found or not readable"
	}

	writeJSONResponse(w, response)
}

// GetClaudeProjectsV2 returns a list of all projects in ~/.claude/projects/
func (h *DataHandler) GetClaudeProjectsV2(w http.ResponseWriter, r *http.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		writeErrorResponse(w, "Could not determine home directory", http.StatusInternalServerError)
		return
	}
	projectsDir := filepath.Join(homeDir, ".claude", "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		writeErrorResponse(w, "Could not read projects directory", http.StatusInternalServerError)
		return
	}

	var projects []map[string]interface{}
	var totalSize int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Decode path: "-Users-bmf-code-foo" -> "/Users/bmf/code/foo"
		decodedPath := strings.ReplaceAll(entry.Name(), "-", "/")

		// Calculate stats for this project
		fileCount, dirSize, sessionCount, agentCount, lastModified := calculateProjectStats(projectPath)
		totalSize += dirSize

		// Extract short project name from decoded path
		projectName := filepath.Base(decodedPath)

		projects = append(projects, map[string]interface{}{
			"id":            entry.Name(),
			"path":          decodedPath,
			"name":          projectName,
			"file_count":    fileCount,
			"total_size":    dirSize,
			"session_count": sessionCount,
			"agent_count":   agentCount,
			"last_modified": lastModified,
			"created":       info.ModTime(),
		})
	}

	// Sort by last_modified descending
	sort.Slice(projects, func(i, j int) bool {
		ti, _ := projects[i]["last_modified"].(time.Time)
		tj, _ := projects[j]["last_modified"].(time.Time)
		return ti.After(tj)
	})

	response := map[string]interface{}{
		"projects":    projects,
		"total_count": len(projects),
		"total_size":  totalSize,
	}

	writeJSONResponse(w, response)
}

// GetClaudeProjectDetailV2 returns detailed info about a specific project
func (h *DataHandler) GetClaudeProjectDetailV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	if projectID == "" {
		writeErrorResponse(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		writeErrorResponse(w, "Could not determine home directory", http.StatusInternalServerError)
		return
	}
	projectPath := filepath.Join(homeDir, ".claude", "projects", projectID)

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		writeErrorResponse(w, "Project not found", http.StatusNotFound)
		return
	}

	// Decode path
	decodedPath := strings.ReplaceAll(projectID, "-", "/")

	// Get detailed stats
	fileCount, totalSize, sessionCount, agentCount, lastModified := calculateProjectStats(projectPath)

	// Get list of sessions with details
	sessions := getProjectSessions(projectPath)

	// Calculate size breakdown
	var sessionSize, agentSize int64
	for _, session := range sessions {
		if isAgent, _ := session["is_agent"].(bool); isAgent {
			agentSize += session["size"].(int64)
		} else {
			sessionSize += session["size"].(int64)
		}
	}

	response := map[string]interface{}{
		"id":            projectID,
		"path":          decodedPath,
		"name":          filepath.Base(decodedPath),
		"file_count":    fileCount,
		"total_size":    totalSize,
		"session_count": sessionCount,
		"agent_count":   agentCount,
		"last_modified": lastModified,
		"sessions":      sessions,
		"size_breakdown": map[string]interface{}{
			"sessions": sessionSize,
			"agents":   agentSize,
		},
	}

	writeJSONResponse(w, response)
}

// Helper functions for Claude config parsing

func parsePermissions(settings map[string]interface{}) map[string][]string {
	result := map[string][]string{
		"bash":  {},
		"tools": {},
		"mcp":   {},
		"other": {},
	}

	permissions, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		return result
	}

	allow, ok := permissions["allow"].([]interface{})
	if !ok {
		return result
	}

	for _, p := range allow {
		perm, ok := p.(string)
		if !ok {
			continue
		}

		if strings.HasPrefix(perm, "Bash(") {
			// Extract just the command part: "Bash(git:*)" -> "git:*"
			inner := strings.TrimPrefix(perm, "Bash(")
			inner = strings.TrimSuffix(inner, ")")
			result["bash"] = append(result["bash"], inner)
		} else if strings.HasPrefix(perm, "mcp__") || strings.Contains(perm, "mcp") {
			result["mcp"] = append(result["mcp"], perm)
		} else if strings.Contains(perm, "(") {
			// Tool permissions like "Edit(*)", "Read(*)"
			result["tools"] = append(result["tools"], perm)
		} else {
			result["other"] = append(result["other"], perm)
		}
	}

	return result
}

func parsePlugins(settings map[string]interface{}) map[string][]string {
	result := map[string][]string{
		"enabled":  {},
		"disabled": {},
	}

	plugins, ok := settings["enabledPlugins"].(map[string]interface{})
	if !ok {
		return result
	}

	for name, enabled := range plugins {
		if isEnabled, ok := enabled.(bool); ok && isEnabled {
			result["enabled"] = append(result["enabled"], name)
		} else {
			result["disabled"] = append(result["disabled"], name)
		}
	}

	// Sort for consistent output
	sort.Strings(result["enabled"])
	sort.Strings(result["disabled"])

	return result
}

func parseClaudeMdSections(content string) []map[string]interface{} {
	var sections []map[string]interface{}

	// Look for XML-like tags that are commonly used
	tags := []string{"system-reminder", "memory", "personal-note", "universal-laws", "guidelines", "context-specific"}

	for _, tag := range tags {
		openTag := "<" + tag + ">"
		if strings.Contains(content, openTag) {
			// Find approximate position
			idx := strings.Index(content, openTag)
			sections = append(sections, map[string]interface{}{
				"name":     tag,
				"position": idx,
			})
		}
	}

	// Sort by position
	sort.Slice(sections, func(i, j int) bool {
		pi, _ := sections[i]["position"].(int)
		pj, _ := sections[j]["position"].(int)
		return pi < pj
	})

	return sections
}

func parseMCPServers(mcpConfig map[string]interface{}) []map[string]interface{} {
	var servers []map[string]interface{}

	serversMap, ok := mcpConfig["mcpServers"].(map[string]interface{})
	if !ok {
		return servers
	}

	for name, config := range serversMap {
		serverConfig, ok := config.(map[string]interface{})
		if !ok {
			continue
		}

		server := map[string]interface{}{
			"name":    name,
			"command": serverConfig["command"],
			"type":    serverConfig["type"],
		}

		if args, ok := serverConfig["args"].([]interface{}); ok {
			server["args"] = args
		}

		servers = append(servers, server)
	}

	return servers
}

func getNestedString(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(string); ok {
				return val
			}
			return ""
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

func calculateProjectStats(projectPath string) (fileCount int, totalSize int64, sessionCount int, agentCount int, lastModified time.Time) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectories are subagent conversations
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileCount++
		totalSize += info.Size()

		if info.ModTime().After(lastModified) {
			lastModified = info.ModTime()
		}

		name := entry.Name()
		if strings.HasPrefix(name, "agent-") {
			agentCount++
		} else if strings.HasSuffix(name, ".jsonl") {
			sessionCount++
		}
	}

	return
}

func getProjectSessions(projectPath string) []map[string]interface{} {
	var sessions []map[string]interface{}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return sessions
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}

		isAgent := strings.HasPrefix(name, "agent-")
		sessionID := strings.TrimSuffix(name, ".jsonl")

		sessions = append(sessions, map[string]interface{}{
			"id":        sessionID,
			"file":      name,
			"size":      info.Size(),
			"modified":  info.ModTime(),
			"is_agent":  isAgent,
		})
	}

	// Sort by modified time descending
	sort.Slice(sessions, func(i, j int) bool {
		ti, _ := sessions[i]["modified"].(time.Time)
		tj, _ := sessions[j]["modified"].(time.Time)
		return ti.After(tj)
	})

	return sessions
}

// ============================================================================
// CC-VIZ Session Data Endpoints
// ============================================================================

// GetTodosV2 returns aggregated todo stats and session list from database
func (h *DataHandler) GetTodosV2(w http.ResponseWriter, r *http.Request) {
	// Query database for aggregated stats
	query := `
		SELECT 
			COUNT(*) as total_files,
			COALESCE(SUM(CASE WHEN todo_count > 0 THEN 1 ELSE 0 END), 0) as non_empty_files,
			COALESCE(SUM(pending_count), 0) as pending,
			COALESCE(SUM(in_progress_count), 0) as in_progress,
			COALESCE(SUM(completed_count), 0) as completed,
			MAX(indexed_at) as last_indexed
		FROM claude_todo_sessions
	`

	storage, ok := h.storageService.(*service.SQLiteStorageService)
	if !ok {
		writeErrorResponse(w, "Storage service not available", http.StatusInternalServerError)
		return
	}

	var totalFiles, nonEmptyFiles, pending, inProgress, completed int
	var lastIndexed sql.NullString

	err := storage.GetDB().QueryRow(query).Scan(
		&totalFiles,
		&nonEmptyFiles,
		&pending,
		&inProgress,
		&completed,
		&lastIndexed,
	)
	if err != nil {
		log.Printf("Error querying todo stats: %v", err)
		writeErrorResponse(w, "Failed to query todo stats", http.StatusInternalServerError)
		return
	}

	// Query sessions
	sessionsQuery := `
		SELECT session_uuid, agent_uuid, file_path, file_size, todo_count,
		       pending_count, in_progress_count, completed_count, modified_at
		FROM claude_todo_sessions
		ORDER BY modified_at DESC
	`

	rows, err := storage.GetDB().Query(sessionsQuery)
	if err != nil {
		log.Printf("Error querying todo sessions: %v", err)
		writeErrorResponse(w, "Failed to query todo sessions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var sessionUUID, agentUUID, filePath, modifiedAt string
		var fileSize, todoCount, pendingCount, inProgressCount, completedCount int

		err := rows.Scan(
			&sessionUUID,
			&agentUUID,
			&filePath,
			&fileSize,
			&todoCount,
			&pendingCount,
			&inProgressCount,
			&completedCount,
			&modifiedAt,
		)
		if err != nil {
			continue
		}

		sessions = append(sessions, map[string]interface{}{
			"session_uuid":       sessionUUID,
			"agent_uuid":         agentUUID,
			"file_path":          filePath,
			"file_size":          fileSize,
			"todo_count":         todoCount,
			"pending_count":      pendingCount,
			"in_progress_count":  inProgressCount,
			"completed_count":    completedCount,
			"modified_at":        modifiedAt,
		})
	}

	response := map[string]interface{}{
		"total_files":      totalFiles,
		"non_empty_files":  nonEmptyFiles,
		"status_breakdown": map[string]int{
			"pending":     pending,
			"in_progress": inProgress,
			"completed":   completed,
		},
		"sessions":     sessions,
		"last_indexed": lastIndexed.String,
	}

	writeJSONResponse(w, response)
}

// GetTodoDetailV2 returns todos for a specific session from database
func (h *DataHandler) GetTodoDetailV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionUUID := vars["session_uuid"]

	if sessionUUID == "" {
		writeErrorResponse(w, "Session UUID is required", http.StatusBadRequest)
		return
	}

	storage, ok := h.storageService.(*service.SQLiteStorageService)
	if !ok {
		writeErrorResponse(w, "Storage service not available", http.StatusInternalServerError)
		return
	}

	// Query todos for this session
	query := `
		SELECT content, status, active_form
		FROM claude_todos
		WHERE session_uuid = ?
		ORDER BY item_index ASC
	`

	rows, err := storage.GetDB().Query(query, sessionUUID)
	if err != nil {
		log.Printf("Error querying todos for session %s: %v", sessionUUID, err)
		writeErrorResponse(w, "Failed to query todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []map[string]interface{}
	for rows.Next() {
		var content, status, activeForm string
		err := rows.Scan(&content, &status, &activeForm)
		if err != nil {
			continue
		}

		todos = append(todos, map[string]interface{}{
			"content":     content,
			"status":      status,
			"active_form": activeForm,
		})
	}

	if len(todos) == 0 {
		writeErrorResponse(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get session metadata
	var agentUUID, filePath, modifiedAt string
	sessionQuery := `
		SELECT agent_uuid, file_path, modified_at
		FROM claude_todo_sessions
		WHERE session_uuid = ?
	`
	err = storage.GetDB().QueryRow(sessionQuery, sessionUUID).Scan(&agentUUID, &filePath, &modifiedAt)
	if err != nil {
		log.Printf("Error querying session metadata: %v", err)
	}

	response := map[string]interface{}{
		"session_uuid": sessionUUID,
		"agent_uuid":   agentUUID,
		"file_path":    filePath,
		"modified_at":  modifiedAt,
		"todos":        todos,
	}

	writeJSONResponse(w, response)
}

// GetPlansV2 returns all plans from database
func (h *DataHandler) GetPlansV2(w http.ResponseWriter, r *http.Request) {
	storage, ok := h.storageService.(*service.SQLiteStorageService)
	if !ok {
		writeErrorResponse(w, "Storage service not available", http.StatusInternalServerError)
		return
	}

	// Get aggregated stats
	statsQuery := `
		SELECT COUNT(*) as count, COALESCE(SUM(file_size), 0) as total_size, MAX(indexed_at) as last_indexed
		FROM claude_plans
	`

	var count int
	var totalSize int64
	var lastIndexed sql.NullString
	err := storage.GetDB().QueryRow(statsQuery).Scan(&count, &totalSize, &lastIndexed)
	if err != nil {
		log.Printf("Error querying plan stats: %v", err)
		writeErrorResponse(w, "Failed to query plan stats", http.StatusInternalServerError)
		return
	}

	// Query all plans
	plansQuery := `
		SELECT id, file_name, display_name, preview, file_size, modified_at
		FROM claude_plans
		ORDER BY modified_at DESC
	`

	rows, err := storage.GetDB().Query(plansQuery)
	if err != nil {
		log.Printf("Error querying plans: %v", err)
		writeErrorResponse(w, "Failed to query plans", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plans []map[string]interface{}
	for rows.Next() {
		var id int
		var fileName, displayName, preview, modifiedAt string
		var fileSize int64

		err := rows.Scan(&id, &fileName, &displayName, &preview, &fileSize, &modifiedAt)
		if err != nil {
			continue
		}

		plans = append(plans, map[string]interface{}{
			"id":           id,
			"file_name":    fileName,
			"display_name": displayName,
			"preview":      preview,
			"file_size":    fileSize,
			"modified_at":  modifiedAt,
		})
	}

	response := map[string]interface{}{
		"total_count":  count,
		"total_size":   totalSize,
		"plans":        plans,
		"last_indexed": lastIndexed.String,
	}

	writeJSONResponse(w, response)
}

// GetPlanDetailV2 returns a specific plan's content from database
func (h *DataHandler) GetPlanDetailV2(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	if idStr == "" {
		writeErrorResponse(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	storage, ok := h.storageService.(*service.SQLiteStorageService)
	if !ok {
		writeErrorResponse(w, "Storage service not available", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT id, file_name, display_name, content, file_size, modified_at
		FROM claude_plans
		WHERE id = ?
	`

	var fileName, displayName, content, modifiedAt string
	var fileSize int64
	var planID int

	err = storage.GetDB().QueryRow(query, id).Scan(&planID, &fileName, &displayName, &content, &fileSize, &modifiedAt)
	if err == sql.ErrNoRows {
		writeErrorResponse(w, "Plan not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error querying plan %d: %v", id, err)
		writeErrorResponse(w, "Failed to query plan", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":           planID,
		"file_name":    fileName,
		"display_name": displayName,
		"content":      content,
		"file_size":    fileSize,
		"modified_at":  modifiedAt,
	}

	writeJSONResponse(w, response)
}

// ReindexTodosV2 triggers manual reindexing of todos and plans
func (h *DataHandler) ReindexTodosV2(w http.ResponseWriter, r *http.Request) {
	storage, ok := h.storageService.(*service.SQLiteStorageService)
	if !ok {
		writeErrorResponse(w, "Storage service not available", http.StatusInternalServerError)
		return
	}

	// Create session data indexer
	indexer, err := service.NewSessionDataIndexer(storage)
	if err != nil {
		log.Printf("Error creating session data indexer: %v", err)
		writeErrorResponse(w, "Failed to create indexer", http.StatusInternalServerError)
		return
	}

	start := time.Now()

	// Index todos
	filesProcessed, todosIndexed, errors := indexer.IndexTodos()

	// Index plans
	plansIndexed, planErrors := indexer.IndexPlans()
	errors = append(errors, planErrors...)

	duration := time.Since(start)

	response := map[string]interface{}{
		"files_processed": filesProcessed,
		"todos_indexed":   todosIndexed,
		"plans_indexed":   plansIndexed,
		"errors":          errors,
		"duration":        duration.String(),
	}

	writeJSONResponse(w, response)
}
