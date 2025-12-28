package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

// TestNewPlanoProvider verifies PlanoProvider initialization
func TestNewPlanoProvider(t *testing.T) {
	cfg := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: "http://localhost:8080",
	}

	provider := NewPlanoProvider("plano", cfg)

	if provider == nil {
		t.Fatal("NewPlanoProvider returned nil")
	}

	planoProvider, ok := provider.(*PlanoProvider)
	if !ok {
		t.Fatal("NewPlanoProvider did not return a *PlanoProvider")
	}

	if planoProvider.Name() != "plano" {
		t.Errorf("Expected provider name 'plano', got '%s'", planoProvider.Name())
	}

	if planoProvider.config.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected base URL 'http://localhost:8080', got '%s'", planoProvider.config.BaseURL)
	}

	if planoProvider.client == nil {
		t.Error("PlanoProvider client is nil")
	}
}

// TestPlanoProvider_ForwardRequest_NonStreaming tests non-streaming request forwarding
func TestPlanoProvider_ForwardRequest_NonStreaming(t *testing.T) {
	// Create a mock Plano server that returns OpenAI-format responses
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path '/v1/chat/completions', got '%s'", r.URL.Path)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Verify Authorization header is NOT present (Plano handles auth)
		if r.Header.Get("Authorization") != "" {
			t.Error("Authorization header should not be set for Plano requests")
		}

		// Read and verify request body is in OpenAI format
		bodyBytes, _ := io.ReadAll(r.Body)
		var openAIReq map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &openAIReq); err != nil {
			t.Errorf("Failed to parse OpenAI request: %v", err)
		}

		// Verify it has OpenAI-format fields
		if _, ok := openAIReq["messages"]; !ok {
			t.Error("OpenAI request missing 'messages' field")
		}

		// Return a mock OpenAI response
		response := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677652288,
			"model":   "gemini/gemini-2.0-flash-exp",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello from Plano!",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create PlanoProvider pointing to mock server
	cfg := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: mockServer.URL,
	}
	provider := NewPlanoProvider("plano", cfg)

	// Create an Anthropic-format request
	anthropicReq := model.AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		Messages: []model.AnthropicMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: false,
	}

	reqBody, _ := json.Marshal(anthropicReq)
	httpReq := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Forward the request
	resp, err := provider.ForwardRequest(context.Background(), httpReq)
	if err != nil {
		t.Fatalf("ForwardRequest failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read and parse response body (should be in Anthropic format)
	respBody, _ := io.ReadAll(resp.Body)
	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse Anthropic response: %v", err)
	}

	// Verify Anthropic-format response fields
	if anthropicResp["type"] != "message" {
		t.Errorf("Expected type 'message', got '%v'", anthropicResp["type"])
	}

	if anthropicResp["role"] != "assistant" {
		t.Errorf("Expected role 'assistant', got '%v'", anthropicResp["role"])
	}

	// Verify content blocks
	content, ok := anthropicResp["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("Response missing content blocks")
	}

	contentBlock, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatal("Content block is not a map")
	}

	if contentBlock["type"] != "text" {
		t.Errorf("Expected content type 'text', got '%v'", contentBlock["type"])
	}

	if contentBlock["text"] != "Hello from Plano!" {
		t.Errorf("Expected text 'Hello from Plano!', got '%v'", contentBlock["text"])
	}

	// Verify usage is converted to Anthropic format
	usage, ok := anthropicResp["usage"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing usage field")
	}

	// OpenAI's prompt_tokens → Anthropic's input_tokens
	if usage["input_tokens"] != float64(10) {
		t.Errorf("Expected input_tokens 10, got %v", usage["input_tokens"])
	}

	// OpenAI's completion_tokens → Anthropic's output_tokens
	if usage["output_tokens"] != float64(20) {
		t.Errorf("Expected output_tokens 20, got %v", usage["output_tokens"])
	}
}

// TestPlanoProvider_ForwardRequest_Streaming tests streaming request forwarding
func TestPlanoProvider_ForwardRequest_Streaming(t *testing.T) {
	// Create a mock Plano server that returns OpenAI-format SSE stream
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path '/v1/chat/completions', got '%s'", r.URL.Path)
		}

		// Send OpenAI-format SSE stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected http.ResponseWriter to be an http.Flusher")
		}

		// Send first chunk
		chunk1 := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion.chunk",
			"created": 1677652288,
			"model":   "gemini/gemini-2.0-flash-exp",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"delta": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello",
					},
				},
			},
		}
		chunk1JSON, _ := json.Marshal(chunk1)
		w.Write([]byte("data: "))
		w.Write(chunk1JSON)
		w.Write([]byte("\n\n"))
		flusher.Flush()

		// Send second chunk
		chunk2 := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion.chunk",
			"created": 1677652288,
			"model":   "gemini/gemini-2.0-flash-exp",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"delta": map[string]interface{}{
						"content": " world",
					},
				},
			},
		}
		chunk2JSON, _ := json.Marshal(chunk2)
		w.Write([]byte("data: "))
		w.Write(chunk2JSON)
		w.Write([]byte("\n\n"))
		flusher.Flush()

		// Send [DONE]
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer mockServer.Close()

	// Create PlanoProvider pointing to mock server
	cfg := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: mockServer.URL,
	}
	provider := NewPlanoProvider("plano", cfg)

	// Create an Anthropic-format streaming request
	anthropicReq := model.AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		Messages: []model.AnthropicMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: true,
	}

	reqBody, _ := json.Marshal(anthropicReq)
	httpReq := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Forward the request
	resp, err := provider.ForwardRequest(context.Background(), httpReq)
	if err != nil {
		t.Fatalf("ForwardRequest failed: %v", err)
	}
	defer resp.Body.Close()

	// Read the SSE stream
	respBody, _ := io.ReadAll(resp.Body)
	responseStr := string(respBody)

	// Verify Anthropic-format SSE events are present
	expectedEvents := []string{
		"message_start",
		"content_block_start",
		"content_block_delta",
		"content_block_stop",
		"message_stop",
	}

	for _, event := range expectedEvents {
		if !strings.Contains(responseStr, event) {
			t.Errorf("Expected response to contain '%s' event, but it was missing", event)
		}
	}

	// Verify we got the content deltas
	if !strings.Contains(responseStr, "Hello") {
		t.Error("Expected response to contain 'Hello'")
	}

	if !strings.Contains(responseStr, "world") {
		t.Error("Expected response to contain 'world'")
	}
}

// TestPlanoProvider_ErrorHandling tests error response handling
func TestPlanoProvider_ErrorHandling(t *testing.T) {
	// Create a mock Plano server that returns errors
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a 404 error (invalid model)
		w.WriteHeader(http.StatusNotFound)
		errorResp := map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Model not found: invalid/model",
				"type":    "invalid_request_error",
			},
		}
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer mockServer.Close()

	// Create PlanoProvider pointing to mock server
	cfg := &config.ProviderConfig{
		Format:  "openai",
		BaseURL: mockServer.URL,
	}
	provider := NewPlanoProvider("plano", cfg)

	// Create a request
	anthropicReq := model.AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		Messages: []model.AnthropicMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: false,
	}

	reqBody, _ := json.Marshal(anthropicReq)
	httpReq := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Forward the request
	resp, err := provider.ForwardRequest(context.Background(), httpReq)
	if err != nil {
		t.Fatalf("ForwardRequest failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify error response is in Anthropic format
	respBody, _ := io.ReadAll(resp.Body)
	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	// Verify Anthropic error format
	if anthropicResp["type"] != "error" {
		t.Errorf("Expected type 'error', got '%v'", anthropicResp["type"])
	}

	errorObj, ok := anthropicResp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing error object")
	}

	if errorObj["type"] != "api_error" {
		t.Errorf("Expected error type 'api_error', got '%v'", errorObj["type"])
	}

	// Verify error message mentions Plano and status code
	message, ok := errorObj["message"].(string)
	if !ok {
		t.Fatal("Error message is not a string")
	}

	if !strings.Contains(message, "Plano") {
		t.Error("Error message should mention 'Plano'")
	}

	if !strings.Contains(message, "404") {
		t.Error("Error message should include status code")
	}
}

// TestPlanoProvider_ConversionFunctionsWork verifies OpenAI conversion functions work with Plano
func TestPlanoProvider_ConversionFunctionsWork(t *testing.T) {
	// This test verifies that the existing conversion functions from openai.go
	// work correctly when called by PlanoProvider

	// Create a sample Anthropic request
	anthropicReq := &model.AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		System: []model.AnthropicSystemMessage{
			{Type: "text", Text: "You are a helpful assistant."},
		},
		Messages: []model.AnthropicMessage{
			{
				Role:    "user",
				Content: "What is 2+2?",
			},
		},
		Stream: false,
	}

	// Call the conversion function (from openai.go, reused by plano.go)
	openAIReq := convertAnthropicToOpenAI(anthropicReq)

	// Verify conversion worked
	if openAIReq == nil {
		t.Fatal("convertAnthropicToOpenAI returned nil")
	}

	// Verify OpenAI format fields
	if openAIReq["model"] != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%v'", openAIReq["model"])
	}

	messages, ok := openAIReq["messages"].([]map[string]interface{})
	if !ok {
		t.Fatal("OpenAI request missing messages field")
	}

	// Should have system message + user message
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify system message
	if messages[0]["role"] != "system" {
		t.Errorf("Expected first message role 'system', got '%v'", messages[0]["role"])
	}

	if messages[0]["content"] != "You are a helpful assistant." {
		t.Errorf("Expected system content 'You are a helpful assistant.', got '%v'", messages[0]["content"])
	}

	// Verify user message
	if messages[1]["role"] != "user" {
		t.Errorf("Expected second message role 'user', got '%v'", messages[1]["role"])
	}

	if messages[1]["content"] != "What is 2+2?" {
		t.Errorf("Expected user content 'What is 2+2?', got '%v'", messages[1]["content"])
	}
}
