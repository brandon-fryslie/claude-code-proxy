package provider

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

// PlanoProvider implements the Provider interface for Plano/ArchGW routing
// Plano is an LLM proxy that routes requests to multiple providers (Gemini, DeepSeek, Qwen, etc.)
// It accepts OpenAI-format requests, so we reuse the Anthropic→OpenAI conversion from OpenAIProvider
type PlanoProvider struct {
	name   string
	client *http.Client
	config *config.ProviderConfig
}

// NewPlanoProvider creates a new Plano provider instance
func NewPlanoProvider(name string, cfg *config.ProviderConfig) Provider {
	return &PlanoProvider{
		name: name,
		client: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes timeout (matches other providers)
		},
		config: cfg,
	}
}

// Name returns the provider name
func (p *PlanoProvider) Name() string {
	return p.name
}

// ForwardRequest forwards a request to Plano, converting between Anthropic and OpenAI formats
// This implementation is nearly identical to OpenAIProvider, with key differences:
// - No Authorization header (Plano handles upstream auth via environment variables)
// - Error messages reference "Plano" instead of "OpenAI"
// - Provider name for logging is "plano"
func (p *PlanoProvider) ForwardRequest(ctx context.Context, originalReq *http.Request) (*http.Response, error) {
	// 1. Read and parse the Anthropic request
	bodyBytes, err := io.ReadAll(originalReq.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	originalReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var anthropicReq model.AnthropicRequest
	if err := json.Unmarshal(bodyBytes, &anthropicReq); err != nil {
		return nil, fmt.Errorf("failed to parse anthropic request: %w", err)
	}

	// 2. Convert Anthropic request to OpenAI format
	// REUSE: convertAnthropicToOpenAI() from openai.go
	// This handles:
	// - System messages concatenation
	// - Message content blocks → text content
	// - Tool definitions (Anthropic → OpenAI format)
	// - Tool results conversion
	// - Stream options
	openAIReq := convertAnthropicToOpenAI(&anthropicReq)
	newBodyBytes, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	// 3. Build request to Plano
	proxyReq := originalReq.Clone(ctx)
	proxyReq.Body = io.NopCloser(bytes.NewReader(newBodyBytes))
	proxyReq.ContentLength = int64(len(newBodyBytes))

	// Parse the configured Plano base URL
	baseURL, err := url.Parse(p.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL '%s': %w", p.config.BaseURL, err)
	}

	// Update the destination URL to point to Plano
	proxyReq.URL.Scheme = baseURL.Scheme
	proxyReq.URL.Host = baseURL.Host
	proxyReq.URL.Path = "/v1/chat/completions" // Plano's OpenAI-compatible endpoint

	// Update request headers
	proxyReq.RequestURI = ""
	proxyReq.Host = baseURL.Host

	// Remove Anthropic-specific headers
	proxyReq.Header.Del("anthropic-version")
	proxyReq.Header.Del("x-api-key")

	// Set headers for Plano
	// NOTE: No Authorization header - Plano reads API keys from environment variables
	// (OPENAI_API_KEY, GEMINI_API_KEY, DEEPSEEK_API_KEY, etc.)
	proxyReq.Header.Set("Content-Type", "application/json")

	// 4. Forward the request to Plano
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("plano request failed: %w", err)
	}

	// 5. Handle error responses from Plano
	if resp.StatusCode >= 400 {
		// Read the error body for debugging
		errorBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Try to parse Plano's error response
		var planoError map[string]interface{}
		errorMessage := string(errorBody)
		if err := json.Unmarshal(errorBody, &planoError); err == nil {
			// Plano returned structured error - include it in the message
			if errMsg, ok := planoError["error"].(map[string]interface{}); ok {
				if msg, ok := errMsg["message"].(string); ok {
					errorMessage = msg
				}
			}
		}

		// Create an error response in Anthropic format
		errorResp := map[string]interface{}{
			"type": "error",
			"error": map[string]interface{}{
				"type":    "api_error",
				"message": fmt.Sprintf("Plano API error (status %d): %s", resp.StatusCode, errorMessage),
			},
		}
		errorJSON, _ := json.Marshal(errorResp)

		// Create a new response with the error
		resp.Body = io.NopCloser(bytes.NewReader(errorJSON))
		resp.Header.Set("Content-Type", "application/json")
		resp.Header.Del("Content-Encoding")
		resp.ContentLength = int64(len(errorJSON))

		return resp, nil
	}

	// 6. Handle gzip-encoded responses
	var bodyReader io.ReadCloser = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		bodyReader = gzReader
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
	}

	// 7. Transform response from OpenAI format back to Anthropic format
	// For streaming responses, we need to transform SSE events in real-time
	if anthropicReq.Stream {
		// Create a pipe to transform the stream
		pr, pw := io.Pipe()

		// Start a goroutine to transform the stream
		// REUSE: transformOpenAIStreamToAnthropic() from openai.go
		// This handles:
		// - SSE event parsing
		// - message_start, content_block_start, content_block_delta events
		// - Tool calls in streaming mode
		// - Usage data in final chunk
		// - message_stop event
		go func() {
			defer pw.Close()
			defer bodyReader.Close()
			transformOpenAIStreamToAnthropic(bodyReader, pw)
		}()

		// Replace the response body with our transformed stream
		resp.Body = pr
	} else {
		// For non-streaming, read and convert the full response
		respBody, err := io.ReadAll(bodyReader)
		bodyReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// REUSE: transformOpenAIResponseToAnthropic() from openai.go
		// This handles:
		// - OpenAI choices → Anthropic content blocks
		// - Tool calls → tool_use blocks
		// - Usage data conversion (prompt_tokens → input_tokens, etc.)
		transformedBody := transformOpenAIResponseToAnthropic(respBody)
		resp.Body = io.NopCloser(bytes.NewReader(transformedBody))
		resp.ContentLength = int64(len(transformedBody))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))
	}

	return resp, nil
}
