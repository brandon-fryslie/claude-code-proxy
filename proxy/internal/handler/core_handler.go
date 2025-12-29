package handler

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
	"github.com/seifghazi/claude-code-monitor/internal/service"
)

// CoreHandler handles the core proxy functionality:
// - /v1/messages - Main Claude API endpoint
// - /v1/models - List available models
// - /health - Health check
//
// It has minimal dependencies: write-only storage, model router, logger, config.
// This handler is designed to be lightweight and stable - changes are rare.
type CoreHandler struct {
	storageService service.StorageService
	modelRouter    *service.ModelRouter
	logger         *log.Logger
	config         *config.Config
}

// NewCoreHandler creates a new CoreHandler with the required dependencies.
func NewCoreHandler(storageService service.StorageService, logger *log.Logger, modelRouter *service.ModelRouter, cfg *config.Config) *CoreHandler {
	return &CoreHandler{
		storageService: storageService,
		modelRouter:    modelRouter,
		logger:         logger,
		config:         cfg,
	}
}

// ChatCompletions handles OpenAI-format requests with an error message.
func (h *CoreHandler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, "This is an Anthropic proxy. Please use the /v1/messages endpoint instead of /v1/chat/completions", http.StatusBadRequest)
}

// Messages handles the main /v1/messages endpoint for proxying Claude API requests.
func (h *CoreHandler) Messages(w http.ResponseWriter, r *http.Request) {
	// Get body bytes from context (set by middleware)
	bodyBytes := getBodyBytes(r)
	if bodyBytes == nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Parse the request
	var req model.AnthropicRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	requestID := generateCoreRequestID()
	startTime := time.Now()

	// Use model router to determine provider and route the request
	decision, err := h.modelRouter.DetermineRoute(&req)
	if err != nil {
		log.Printf("❌ Error routing request: %v", err)
		writeErrorResponse(w, "Failed to route request", http.StatusInternalServerError)
		return
	}

	// Extract tools used from request
	var toolsUsed []string
	for _, tool := range req.Tools {
		toolsUsed = append(toolsUsed, tool.Name)
	}

	// Create request log with routing information
	requestLog := &model.RequestLog{
		RequestID:     requestID,
		Timestamp:     time.Now().Format(time.RFC3339),
		Method:        r.Method,
		Endpoint:      r.URL.Path,
		Headers:       SanitizeHeaders(r.Header),
		Body:          req,
		Model:         decision.OriginalModel,
		OriginalModel: decision.OriginalModel,
		RoutedModel:   decision.TargetModel,
		Provider:      decision.ProviderName,
		SubagentName:  decision.SubagentName,
		ToolsUsed:     toolsUsed,
		UserAgent:     r.Header.Get("User-Agent"),
		ContentType:   r.Header.Get("Content-Type"),
	}

	if _, err := h.storageService.SaveRequest(requestLog); err != nil {
		log.Printf("❌ Error saving request: %v", err)
	}

	// If the model was changed by routing, update the request body
	if decision.TargetModel != decision.OriginalModel {
		req.Model = decision.TargetModel

		// Re-marshal the request with the updated model
		updatedBodyBytes, err := json.Marshal(req)
		if err != nil {
			log.Printf("❌ Error marshaling updated request: %v", err)
			writeErrorResponse(w, "Failed to process request", http.StatusInternalServerError)
			return
		}

		// Update the request body
		r.Body = io.NopCloser(bytes.NewReader(updatedBodyBytes))
		r.ContentLength = int64(len(updatedBodyBytes))
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(updatedBodyBytes)))
	}

	// Forward the request to the selected provider
	resp, err := decision.Provider.ForwardRequest(r.Context(), r)
	if err != nil {
		log.Printf("❌ Error forwarding to %s API: %v", decision.Provider.Name(), err)
		writeErrorResponse(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		h.handleStreamingResponse(w, resp, requestLog, startTime)
		return
	}

	h.handleNonStreamingResponse(w, resp, requestLog, startTime)
}

// Models handles the /v1/models endpoint.
func (h *CoreHandler) Models(w http.ResponseWriter, r *http.Request) {
	response := &model.ModelsResponse{
		Object: "list",
		Data: []model.ModelInfo{
			{
				ID:      "claude-3-sonnet-20240229",
				Object:  "model",
				Created: 1677610602,
				OwnedBy: "anthropic",
			},
			{
				ID:      "claude-3-opus-20240229",
				Object:  "model",
				Created: 1677610602,
				OwnedBy: "anthropic",
			},
			{
				ID:      "claude-3-haiku-20240307",
				Object:  "model",
				Created: 1677610602,
				OwnedBy: "anthropic",
			},
		},
	}

	writeJSONResponse(w, response)
}

// Health handles the /health endpoint.
func (h *CoreHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	dbStatus := "connected"
	if h.storageService == nil {
		dbStatus = "disconnected"
	}

	// Get provider health information including circuit breaker status
	providerHealth := h.modelRouter.GetProviderHealth()

	response := map[string]interface{}{
		"status":          "ok",
		"service":         "proxy-core",
		"database":        dbStatus,
		"provider_health": providerHealth,
		"timestamp":       time.Now(),
	}

	writeJSONResponse(w, response)
}

// NotFound handles 404 responses.
func (h *CoreHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, "Not found", http.StatusNotFound)
}

func (h *CoreHandler) handleStreamingResponse(w http.ResponseWriter, resp *http.Response, requestLog *model.RequestLog, startTime time.Time) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Anthropic API error: %d", resp.StatusCode)
		errorBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error details: %s", string(errorBytes))

		responseLog := &model.ResponseLog{
			StatusCode:   resp.StatusCode,
			Headers:      SanitizeHeaders(resp.Header),
			BodyText:     string(errorBytes),
			ResponseTime: time.Since(startTime).Milliseconds(),
			IsStreaming:  true,
			CompletedAt:  time.Now().Format(time.RFC3339),
		}

		requestLog.Response = responseLog
		if err := h.storageService.UpdateRequestWithResponse(requestLog); err != nil {
			log.Printf("❌ Error updating request with error response: %v", err)
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(errorBytes)
		return
	}

	var fullResponseText strings.Builder
	var toolCalls []model.ContentBlock
	var streamingChunks []string
	var finalUsage *model.AnthropicUsage
	var messageID string
	var modelName string
	var stopReason string
	var firstByteTime int64

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		// Track time to first byte (first actual data)
		if firstByteTime == 0 {
			firstByteTime = time.Since(startTime).Milliseconds()
		}

		streamingChunks = append(streamingChunks, line)
		fmt.Fprintf(w, "%s\n\n", line)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		jsonData := strings.TrimPrefix(line, "data: ")

		// Parse as generic JSON first to capture usage data
		var genericEvent map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &genericEvent); err != nil {
			log.Printf("⚠️ Error unmarshalling streaming event: %v", err)
			continue
		}

		// Capture metadata from message_start event
		if eventType, ok := genericEvent["type"].(string); ok && eventType == "message_start" {
			if message, ok := genericEvent["message"].(map[string]interface{}); ok {
				if id, ok := message["id"].(string); ok {
					messageID = id
				}
				if m, ok := message["model"].(string); ok {
					modelName = m
				}
				if reason, ok := message["stop_reason"].(string); ok {
					stopReason = reason
				}
			}
		}

		// Capture usage data from message_delta event
		if eventType, ok := genericEvent["type"].(string); ok && eventType == "message_delta" {
			if usage, ok := genericEvent["usage"].(map[string]interface{}); ok {
				if finalUsage == nil {
					finalUsage = &model.AnthropicUsage{}
				}

				if inputTokens, ok := usage["input_tokens"].(float64); ok {
					finalUsage.InputTokens = int(inputTokens)
				}
				if outputTokens, ok := usage["output_tokens"].(float64); ok {
					finalUsage.OutputTokens = int(outputTokens)
				}
				if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
					finalUsage.CacheCreationInputTokens = int(cacheCreation)
				}
				if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
					finalUsage.CacheReadInputTokens = int(cacheRead)
				}
			}
		}

		// Parse as structured event for content processing
		var event model.StreamingEvent
		if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
			continue
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta != nil {
				if event.Delta.Type == "text_delta" {
					fullResponseText.WriteString(event.Delta.Text)
				} else if event.Delta.Type == "input_json_delta" {
					if event.Index != nil && *event.Index < len(toolCalls) {
						toolCalls[*event.Index].Input = append(toolCalls[*event.Index].Input, event.Delta.Input...)
					}
				}
			}
		case "content_block_start":
			if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
				toolCalls = append(toolCalls, *event.ContentBlock)
			}
		case "message_stop":
			// End of stream
		}
	}

	responseLog := &model.ResponseLog{
		StatusCode:      resp.StatusCode,
		Headers:         SanitizeHeaders(resp.Header),
		StreamingChunks: streamingChunks,
		ResponseTime:    time.Since(startTime).Milliseconds(),
		FirstByteTime:   firstByteTime,
		IsStreaming:     true,
		CompletedAt:     time.Now().Format(time.RFC3339),
		ToolCallCount:   len(toolCalls),
	}

	// Create a structured response body that matches Anthropic's format
	var contentBlocks []model.AnthropicContentBlock
	if fullResponseText.Len() > 0 {
		contentBlocks = append(contentBlocks, model.AnthropicContentBlock{
			Type: "text",
			Text: fullResponseText.String(),
		})
	}

	responseBody := map[string]interface{}{
		"content":     contentBlocks,
		"id":          messageID,
		"model":       modelName,
		"role":        "assistant",
		"stop_reason": stopReason,
		"type":        "message",
	}

	if finalUsage != nil {
		responseBody["usage"] = finalUsage
	}

	responseBodyBytes, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("❌ Error marshaling streaming response body: %v", err)
		responseBodyBytes = []byte("{}")
	}

	responseLog.Body = json.RawMessage(responseBodyBytes)

	requestLog.Response = responseLog
	if err := h.storageService.UpdateRequestWithResponse(requestLog); err != nil {
		log.Printf("❌ Error updating request with streaming response: %v", err)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Streaming error: %v", err)
	} else {
		log.Println("✅ Streaming response completed")
	}
}

func (h *CoreHandler) handleNonStreamingResponse(w http.ResponseWriter, resp *http.Response, requestLog *model.RequestLog, startTime time.Time) {
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading Anthropic response: %v", err)
		writeErrorResponse(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	responseLog := &model.ResponseLog{
		StatusCode:   resp.StatusCode,
		Headers:      SanitizeHeaders(resp.Header),
		ResponseTime: time.Since(startTime).Milliseconds(),
		IsStreaming:  false,
		CompletedAt:  time.Now().Format(time.RFC3339),
	}

	if resp.StatusCode == http.StatusOK {
		var anthropicResp model.AnthropicResponse
		if err := json.Unmarshal(responseBytes, &anthropicResp); err == nil {
			responseLog.Body = json.RawMessage(responseBytes)

			toolCallCount := 0
			for _, block := range anthropicResp.Content {
				if block.Type == "tool_use" {
					toolCallCount++
				}
			}
			responseLog.ToolCallCount = toolCallCount
		} else {
			log.Printf("⚠️ Failed to parse Anthropic response: %v", err)
			log.Printf("📄 Response body (first 500 chars): %s", string(responseBytes[:minInt(500, len(responseBytes))]))
			responseLog.BodyText = string(responseBytes)
		}
	} else {
		responseLog.BodyText = string(responseBytes)
	}

	requestLog.Response = responseLog
	if err := h.storageService.UpdateRequestWithResponse(requestLog); err != nil {
		log.Printf("❌ Error updating request with response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Anthropic API error: %d %s", resp.StatusCode, string(responseBytes))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(responseBytes)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func generateCoreRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
