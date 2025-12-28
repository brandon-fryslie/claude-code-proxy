# ArchGW/Plano Provider Integration Design

**Date:** 2025-12-28
**Phase:** P0.2 - Design
**Status:** Complete
**Dependencies:** Phase 0.1 Research (RESEARCH-archgw.md)

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Configuration Schema](#configuration-schema)
3. [Provider Implementation](#provider-implementation)
4. [Format Conversion Strategy](#format-conversion-strategy)
5. [Routing Integration](#routing-integration)
6. [Error Handling](#error-handling)
7. [Logging Strategy](#logging-strategy)
8. [Implementation Files](#implementation-files)
9. [Testing Strategy](#testing-strategy)
10. [Phase Boundary](#phase-boundary)

---

## Architecture Overview

### Request Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Claude Code Client                              │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ POST /v1/messages (Anthropic format)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Go Proxy (port 3001)                            │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Handler.Messages()                                           │  │
│  │  - Receive Anthropic request                                 │  │
│  │  - Parse & log to SQLite                                     │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │                                        │
│  ┌──────────────────────────▼───────────────────────────────────┐  │
│  │ ModelRouter.DetermineRoute()                                 │  │
│  │  - Check system prompts for subagent signature              │  │
│  │  - Hash static prompt portion                                │  │
│  │  - Match against loaded subagent definitions                 │  │
│  │  - Parse mapping: "provider:model"                           │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │                                        │
│              ┌──────────────┼──────────────┐                        │
│              │              │              │                        │
│         anthropic:      openai:        plano:                       │
│         passthrough     conversion    conversion                    │
│              │              │              │                        │
└──────────────┼──────────────┼──────────────┼────────────────────────┘
               │              │              │
               ▼              ▼              ▼
┌──────────────────┐  ┌──────────────┐  ┌─────────────────────────┐
│ AnthropicProvider│  │OpenAIProvider│  │    PlanoProvider        │
│                  │  │              │  │                         │
│ - No conversion  │  │ - Anthro→OAI │  │ - Anthro→OAI           │
│ - Passthrough    │  │ - OAI→Anthro │  │ - OAI→Anthro           │
│ - Add x-api-key  │  │ - Streaming  │  │ - Streaming            │
│ - Gzip support   │  │ - Tool calls │  │ - Tool calls           │
└────────┬─────────┘  └──────┬───────┘  └───────┬─────────────────┘
         │                   │                   │
         ▼                   ▼                   ▼
┌──────────────────┐  ┌──────────────┐  ┌─────────────────────────┐
│  Anthropic API   │  │  OpenAI API  │  │   Plano (port 8080)     │
│  (unchanged)     │  │  (unchanged) │  │                         │
└──────────────────┘  └──────────────┘  │  - Model routing        │
                                        │  - Multi-provider       │
                                        └───────┬─────────────────┘
                                                │
                              ┌─────────────────┼─────────────────┐
                              │                 │                 │
                              ▼                 ▼                 ▼
                        ┌──────────┐      ┌──────────┐     ┌──────────┐
                        │  Gemini  │      │DeepSeek  │     │  Qwen    │
                        │   API    │      │   API    │     │   API    │
                        └──────────┘      └──────────┘     └──────────┘
```

### Key Design Principles

1. **Plano as a Provider** - Plano is treated as one provider among many, following the existing Provider interface
2. **Format Conversion Reuse** - Leverage existing OpenAI conversion logic from `openai.go`
3. **No Special Cases** - Plano provider works exactly like OpenAI provider (both require Anthropic→OpenAI conversion)
4. **Transparent Routing** - ModelRouter determines routing; providers handle execution
5. **Anthropic Passthrough Unchanged** - Zero modifications to existing Anthropic provider flow

---

## Configuration Schema

### config.yaml Structure

```yaml
providers:
  # Existing providers (unchanged)
  anthropic:
    base_url: "https://api.anthropic.com"
    format: "anthropic"
    version: "2023-06-01"
    max_retries: 3

  openai:
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com"
    format: "openai"

  # New Plano provider
  plano:
    base_url: "http://localhost:8080"  # Or http://plano:8080 in Docker
    format: "openai"                    # Plano expects OpenAI format
    # Note: No api_key needed - Plano handles upstream auth

subagents:
  enable: true
  mappings:
    # Existing mappings (unchanged)
    code-reviewer: "openai:gpt-4o"

    # New Plano-routed subagents
    planner: "plano:gemini/gemini-2.0-flash-exp"
    budget-agent: "plano:deepseek/deepseek-chat"
    security-auditor: "plano:qwen/qwen-max"
```

### Configuration Rules

1. **Provider Name:** Use `plano` (not `archgw`) to match current branding
2. **Format Field:** MUST be `"openai"` (Plano accepts OpenAI format)
3. **Base URL:** Points to Plano service (localhost for dev, service name for Docker)
4. **No API Key:** Plano handles authentication with upstream providers via environment variables
5. **Model Specification:** In subagent mappings, use Plano's model naming: `provider/model-name`
   - Example: `gemini/gemini-2.0-flash-exp`, `deepseek/deepseek-chat`, `openai/gpt-4o`

### Environment Variables

Plano reads API keys from environment variables. The Go proxy doesn't need to pass them:

```bash
# These are read by Plano, not the Go proxy
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export DEEPSEEK_API_KEY="sk-..."
export DASHSCOPE_API_KEY="..."  # For Qwen
```

### Provider Validation

Existing validation in `config.go` already handles Plano:
- ✅ Requires `format` field (must be "openai")
- ✅ Requires `base_url` field
- ✅ Validates format is "anthropic" or "openai"

No code changes needed for validation.

---

## Provider Implementation

### Interface Adherence

```go
// Existing Provider interface (no changes needed)
type Provider interface {
    Name() string
    ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error)
}
```

### PlanoProvider Structure

```go
// proxy/internal/provider/plano.go (NEW FILE)

package provider

import (
    "context"
    "net/http"
    "time"
    "github.com/seifghazi/claude-code-monitor/internal/config"
)

type PlanoProvider struct {
    name   string
    client *http.Client
    config *config.ProviderConfig
}

func NewPlanoProvider(name string, cfg *config.ProviderConfig) Provider {
    return &PlanoProvider{
        name: name,
        client: &http.Client{
            Timeout: 300 * time.Second, // 5 minutes (matches other providers)
        },
        config: cfg,
    }
}

func (p *PlanoProvider) Name() string {
    return p.name
}
```

### Implementation Strategy

**Key Insight:** PlanoProvider is nearly identical to OpenAIProvider since both:
1. Require Anthropic→OpenAI format conversion
2. Send to `/v1/chat/completions` endpoint
3. Handle streaming and non-streaming responses
4. Transform responses back to Anthropic format

**Approach:**
- **Copy** `openai.go` → `plano.go` as starting point
- **Remove** OpenAI-specific headers (`Authorization: Bearer`)
- **Keep** all conversion logic (already handles tool calls, streaming, etc.)
- **Maintain** same error handling patterns

### Differences from OpenAIProvider

| Aspect | OpenAIProvider | PlanoProvider |
|--------|---------------|---------------|
| **Authentication** | `Authorization: Bearer <key>` | None (Plano handles it) |
| **Endpoint** | `/v1/chat/completions` | `/v1/chat/completions` (same) |
| **Model Name** | Direct (e.g., `gpt-4o`) | Plano format (e.g., `gemini/gemini-2.0-flash-exp`) |
| **Conversion** | Anthropic↔OpenAI | Anthropic↔OpenAI (same) |
| **Error Format** | OpenAI errors → Anthropic | OpenAI errors → Anthropic (same) |

### Code Reuse Strategy

**Shared conversion functions** (already in `openai.go`):
- ✅ `convertAnthropicToOpenAI()` - No changes needed
- ✅ `transformOpenAIResponseToAnthropic()` - No changes needed
- ✅ `transformOpenAIStreamToAnthropic()` - No changes needed

**PlanoProvider.ForwardRequest() implementation:**

```go
func (p *PlanoProvider) ForwardRequest(ctx context.Context, originalReq *http.Request) (*http.Response, error) {
    // 1. Read and parse Anthropic request
    bodyBytes, err := io.ReadAll(originalReq.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read request body: %w", err)
    }
    originalReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))

    var anthropicReq model.AnthropicRequest
    if err := json.Unmarshal(bodyBytes, &anthropicReq); err != nil {
        return nil, fmt.Errorf("failed to parse anthropic request: %w", err)
    }

    // 2. Convert to OpenAI format (reuse existing function)
    openAIReq := convertAnthropicToOpenAI(&anthropicReq)
    newBodyBytes, err := json.Marshal(openAIReq)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal openai request: %w", err)
    }

    // 3. Build request to Plano
    proxyReq := originalReq.Clone(ctx)
    proxyReq.Body = io.NopCloser(bytes.NewReader(newBodyBytes))
    proxyReq.ContentLength = int64(len(newBodyBytes))

    baseURL, err := url.Parse(p.config.BaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse base URL: %w", err)
    }

    proxyReq.URL.Scheme = baseURL.Scheme
    proxyReq.URL.Host = baseURL.Host
    proxyReq.URL.Path = "/v1/chat/completions"
    proxyReq.RequestURI = ""
    proxyReq.Host = baseURL.Host

    // 4. Set headers (NO Authorization header for Plano)
    proxyReq.Header.Del("anthropic-version")
    proxyReq.Header.Del("x-api-key")
    proxyReq.Header.Set("Content-Type", "application/json")

    // 5. Forward request
    resp, err := p.client.Do(proxyReq)
    if err != nil {
        return nil, fmt.Errorf("plano request failed: %w", err)
    }

    // 6. Handle errors (convert to Anthropic format)
    if resp.StatusCode >= 400 {
        errorBody, _ := io.ReadAll(resp.Body)
        resp.Body.Close()

        errorResp := map[string]interface{}{
            "type": "error",
            "error": map[string]interface{}{
                "type":    "api_error",
                "message": fmt.Sprintf("Plano API error (status %d): %s", resp.StatusCode, string(errorBody)),
            },
        }
        errorJSON, _ := json.Marshal(errorResp)
        resp.Body = io.NopCloser(bytes.NewReader(errorJSON))
        resp.Header.Set("Content-Type", "application/json")
        resp.ContentLength = int64(len(errorJSON))
        return resp, nil
    }

    // 7. Transform response (streaming or non-streaming)
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

    if anthropicReq.Stream {
        // Stream transformation (reuse existing)
        pr, pw := io.Pipe()
        go func() {
            defer pw.Close()
            defer bodyReader.Close()
            transformOpenAIStreamToAnthropic(bodyReader, pw)
        }()
        resp.Body = pr
    } else {
        // Non-streaming transformation (reuse existing)
        respBody, err := io.ReadAll(bodyReader)
        bodyReader.Close()
        if err != nil {
            return nil, fmt.Errorf("failed to read response body: %w", err)
        }
        transformedBody := transformOpenAIResponseToAnthropic(respBody)
        resp.Body = io.NopCloser(bytes.NewReader(transformedBody))
        resp.ContentLength = int64(len(transformedBody))
        resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))
    }

    return resp, nil
}
```

---

## Format Conversion Strategy

### Conversion Points

```
Client Request (Anthropic)
    ↓
[Go Proxy receives]
    ↓
[ModelRouter determines: "plano" provider]
    ↓
[PlanoProvider.ForwardRequest() called]
    ↓
┌──────────────────────────────────────┐
│ convertAnthropicToOpenAI()          │  ← Reused from openai.go
│  - System messages → single system  │
│  - Content blocks → text content    │
│  - Tools → OpenAI tools format      │
│  - Tool results → user messages     │
│  - Stream options                   │
└──────────────────────────────────────┘
    ↓
[Forward to Plano at /v1/chat/completions]
    ↓
[Plano routes to provider (Gemini, DeepSeek, etc.)]
    ↓
[Response from Plano in OpenAI format]
    ↓
┌──────────────────────────────────────┐
│ transformOpenAIResponseToAnthropic()│  ← Reused from openai.go
│  - choices[0].message → content[]   │
│  - tool_calls → tool_use blocks     │
│  - usage → Anthropic usage format   │
└──────────────────────────────────────┘
    ↓
[Return to client in Anthropic format]
```

### Streaming Conversion

**SSE Event Transformation:**

OpenAI SSE format:
```
data: {"choices":[{"delta":{"content":"Hello"}}],"id":"..."}
data: {"choices":[{"delta":{"content":" world"}}],"id":"..."}
data: [DONE]
```

Anthropic SSE format:
```
data: {"type":"message_start","message":{"id":"...","role":"assistant"}}
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}
data: {"type":"content_block_stop","index":0}
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}
data: {"type":"message_stop"}
```

**Implementation:** Reuse `transformOpenAIStreamToAnthropic()` from `openai.go`
- Already handles all SSE transformations
- Correctly manages message lifecycle events
- Parses tool calls in streaming mode
- Handles usage data in final chunk

### Tool Calls Handling

**Anthropic Tool Definition:**
```json
{
  "tools": [{
    "name": "get_weather",
    "description": "Get current weather",
    "input_schema": {
      "type": "object",
      "properties": {
        "location": {"type": "string"}
      },
      "required": ["location"]
    }
  }]
}
```

**OpenAI Tool Definition:**
```json
{
  "tools": [{
    "type": "function",
    "function": {
      "name": "get_weather",
      "description": "Get current weather",
      "parameters": {
        "type": "object",
        "properties": {
          "location": {"type": "string"}
        },
        "required": ["location"]
      }
    }
  }]
}
```

**Tool Use Response (Anthropic):**
```json
{
  "content": [
    {"type": "text", "text": "I'll check the weather"},
    {"type": "tool_use", "id": "call_123", "name": "get_weather", "input": {"location": "NYC"}}
  ]
}
```

**Tool Call Response (OpenAI):**
```json
{
  "choices": [{
    "message": {
      "content": "I'll check the weather",
      "tool_calls": [{
        "id": "call_123",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\"location\":\"NYC\"}"
        }
      }]
    }
  }]
}
```

**Implementation:** All tool conversion already handled in `convertAnthropicToOpenAI()` and `transformOpenAIResponseToAnthropic()`

### Edge Cases Handled

1. **Empty content** - Defaults to `"..."` (existing behavior)
2. **Array properties missing items** - Adds default `{"type": "string"}` (existing)
3. **Tool results in various formats** - Handles string, array, dict (existing)
4. **O-series models** - Excludes temperature parameter (existing)
5. **Gzip encoding** - Decompression handled (existing)
6. **Usage data in streaming** - Captured from final chunk (existing)

**No new conversion logic needed** - PlanoProvider benefits from all existing OpenAI conversion work.

---

## Routing Integration

### ModelRouter Integration

**No changes needed** to `model_router.go`. Existing logic handles Plano automatically:

```go
// Existing code in DetermineRoute() already works for Plano:

func (mr *ModelRouter) DetermineRoute(req *model.AnthropicRequest) RoutingDecision {
    // 1. Check if subagent routing is enabled
    if !mr.config.Subagents.Enable {
        return defaultAnthropicRouting()
    }

    // 2. Detect subagent via system prompt hashing
    if subagentDef := mr.matchSubagent(req); subagentDef != nil {
        // 3. Parse "provider:model" mapping
        providerName := subagentDef.TargetProvider  // e.g., "plano"
        targetModel := subagentDef.TargetModel      // e.g., "gemini/gemini-2.0-flash-exp"

        // 4. Look up provider (plano will be in providers map)
        provider := mr.providers[providerName]

        // 5. Return routing decision
        return RoutingDecision{
            Provider:      provider,            // PlanoProvider instance
            ProviderName:  providerName,        // "plano"
            OriginalModel: req.Model,           // Original Anthropic model
            TargetModel:   targetModel,         // "gemini/gemini-2.0-flash-exp"
            SubagentName:  subagentDef.Name,   // "planner"
        }
    }

    // 6. Default to Anthropic passthrough
    return defaultAnthropicRouting()
}
```

**Key Points:**
- ✅ Existing prompt hashing works for any subagent
- ✅ "provider:model" parsing is provider-agnostic
- ✅ Provider lookup uses `providers` map (plano will be registered)
- ✅ No special cases needed for Plano

### Provider Registration

**Location:** `cmd/proxy/main.go`

**Current pattern:**
```go
providers := make(map[string]provider.Provider)

// Register providers based on config
for name, cfg := range config.Providers {
    switch cfg.Format {
    case "anthropic":
        providers[name] = provider.NewAnthropicProvider(name, cfg)
    case "openai":
        providers[name] = provider.NewOpenAIProvider(name, cfg)
    }
}
```

**Updated pattern** (handles Plano automatically):
```go
providers := make(map[string]provider.Provider)

for name, cfg := range config.Providers {
    switch cfg.Format {
    case "anthropic":
        providers[name] = provider.NewAnthropicProvider(name, cfg)
    case "openai":
        // Both OpenAI and Plano use "openai" format
        // Distinguish by provider name for logging/debugging purposes
        if name == "plano" {
            providers[name] = provider.NewPlanoProvider(name, cfg)
        } else {
            providers[name] = provider.NewOpenAIProvider(name, cfg)
        }
    }
}
```

**Alternative approach** (even simpler, recommended for Phase 1):
```go
// No special case needed - just use provider name to determine constructor
for name, cfg := range config.Providers {
    switch cfg.Format {
    case "anthropic":
        providers[name] = provider.NewAnthropicProvider(name, cfg)
    case "openai":
        providers[name] = provider.NewOpenAIProvider(name, cfg)
    }
}
```

And in `config.yaml`:
```yaml
providers:
  plano:
    base_url: "http://localhost:8080"
    format: "openai"  # PlanoProvider reuses OpenAI conversion
```

**Decision:** Use OpenAIProvider for Plano initially (simplest). Create separate PlanoProvider only if we need Plano-specific features (guardrails, routing hints, etc.) in Phase 3+.

**Rationale:**
- OpenAIProvider already does everything Plano needs
- Plano's API is OpenAI-compatible by design
- Reduces code duplication
- Can refactor to separate provider later if needed

---

## Error Handling

### Error Scenarios

| Scenario | Detection | Handling | Client Response |
|----------|-----------|----------|-----------------|
| **Plano service down** | Connection refused | Return error in Anthropic format | `{"type":"error","error":{"type":"api_error","message":"plano request failed: ..."}}` |
| **Invalid model** | HTTP 404 from Plano | Convert error to Anthropic format | `{"type":"error","error":{"type":"invalid_request_error","message":"Plano API error (status 404): ..."}}` |
| **Provider unavailable** | HTTP 503 from Plano | Convert error to Anthropic format | `{"type":"error","error":{"type":"api_error","message":"Plano API error (status 503): ..."}}` |
| **Timeout** | Context timeout | Return timeout error | `{"type":"error","error":{"type":"timeout_error","message":"plano request failed: context deadline exceeded"}}` |
| **Malformed response** | JSON parse error | Log and return error | `{"type":"error","error":{"type":"api_error","message":"failed to parse plano response"}}` |

### Error Handling Code

```go
// In PlanoProvider.ForwardRequest()

// 1. Connection/network errors
resp, err := p.client.Do(proxyReq)
if err != nil {
    // Log the error (if logger available)
    return nil, fmt.Errorf("plano request failed: %w", err)
}

// 2. HTTP error status codes (4xx, 5xx)
if resp.StatusCode >= 400 {
    errorBody, _ := io.ReadAll(resp.Body)
    resp.Body.Close()

    // Try to parse Plano's error response
    var planoError map[string]interface{}
    if err := json.Unmarshal(errorBody, &planoError); err == nil {
        // Plano returned structured error - pass it through
        errorResp := map[string]interface{}{
            "type": "error",
            "error": map[string]interface{}{
                "type":    "api_error",
                "message": fmt.Sprintf("Plano error: %v", planoError),
            },
        }
        errorJSON, _ := json.Marshal(errorResp)
        resp.Body = io.NopCloser(bytes.NewReader(errorJSON))
    } else {
        // Plano returned unstructured error
        errorResp := map[string]interface{}{
            "type": "error",
            "error": map[string]interface{}{
                "type":    "api_error",
                "message": fmt.Sprintf("Plano API error (status %d): %s", resp.StatusCode, string(errorBody)),
            },
        }
        errorJSON, _ := json.Marshal(errorResp)
        resp.Body = io.NopCloser(bytes.NewReader(errorJSON))
    }

    resp.Header.Set("Content-Type", "application/json")
    resp.ContentLength = int64(len(errorJSON))
    return resp, nil
}

// 3. Response parsing errors handled in transform functions
```

### Error Logging

**Where to log:**
- Connection failures: Log before returning error
- HTTP errors: Log status code and error body
- Conversion errors: Already logged in existing transform functions

**Log format:**
```go
log.Printf("Plano request failed: provider=%s, model=%s, error=%v",
    p.name, targetModel, err)
```

### Fallback Strategy (Phase 2)

**Not implemented in Phase 1** - Just return errors to client.

**Phase 2 considerations:**
- Circuit breaker pattern (track failure rate)
- Fallback to alternative provider (e.g., plano fails → openai)
- Retry with exponential backoff
- Configurable fallback behavior

**Config example (future):**
```yaml
providers:
  plano:
    base_url: "http://localhost:8080"
    format: "openai"
    fallback: "openai"  # Phase 2: fallback provider
    circuit_breaker:     # Phase 2: circuit breaker config
      enabled: true
      failure_threshold: 5
      timeout: 60s
```

---

## Logging Strategy

### Request Logging (SQLite)

**Existing logging in `handler.go`:**

```go
// Log request details
requestLog := &model.RequestLog{
    ID:           uuid.New().String(),
    Timestamp:    time.Now(),
    Method:       r.Method,
    Path:         r.URL.Path,
    Model:        anthropicReq.Model,
    Provider:     routingDecision.ProviderName,  // "plano"
    TargetModel:  routingDecision.TargetModel,   // "gemini/gemini-2.0-flash-exp"
    // ... other fields
}
```

**What gets logged for Plano requests:**

| Field | Value | Example |
|-------|-------|---------|
| `Provider` | Provider name from routing | `"plano"` |
| `TargetModel` | Model from subagent mapping | `"gemini/gemini-2.0-flash-exp"` |
| `Model` | Original Anthropic model | `"claude-3-5-sonnet-20241022"` |
| `SubagentName` | Matched subagent | `"planner"` |
| `Request` | Original Anthropic request JSON | `{...}` |
| `Response` | Converted Anthropic response JSON | `{...}` |

**No schema changes needed** - Existing fields capture Plano routing.

### Routing Decision Logging

**Console logging in `handler.go`:**

```go
// Existing log pattern (already works for Plano)
if routingDecision.SubagentName != "" {
    log.Printf("→ Routing to %s provider (model: %s) for subagent: %s",
        routingDecision.ProviderName,    // "plano"
        routingDecision.TargetModel,     // "gemini/gemini-2.0-flash-exp"
        routingDecision.SubagentName)    // "planner"
} else {
    log.Printf("→ Using %s provider (model: %s)",
        routingDecision.ProviderName,
        routingDecision.TargetModel)
}
```

**Example output:**
```
→ Routing to plano provider (model: gemini/gemini-2.0-flash-exp) for subagent: planner
```

### Additional Logging (Optional)

**In PlanoProvider.ForwardRequest():**

```go
// Log Plano-specific details (useful for debugging)
log.Printf("Plano request: url=%s, model=%s, stream=%v",
    p.config.BaseURL, openAIReq["model"], anthropicReq.Stream)

// Log response status
log.Printf("Plano response: status=%d, model=%s",
    resp.StatusCode, openAIReq["model"])
```

**When to log:**
- ✅ Routing decisions (already logged)
- ✅ Error conditions (log before returning)
- ❌ Request/response bodies (too verbose, already in SQLite)
- ❌ Every successful request (too noisy)

### Dashboard Display

**No changes needed** - Dashboard already shows:
- Provider name ("plano")
- Model used ("gemini/gemini-2.0-flash-exp")
- Request/response details
- Timestamps and latency

**Future enhancement (Phase 2):**
- Plano-specific metrics (which upstream provider was used)
- Latency breakdown (proxy overhead vs. Plano overhead vs. provider latency)
- Provider error rates

---

## Implementation Files

### New Files

1. **`proxy/internal/provider/plano.go`** (NEW)
   - `PlanoProvider` struct
   - `NewPlanoProvider()` constructor
   - `Name()` method
   - `ForwardRequest()` method (reuses OpenAI conversion)

### Modified Files

1. **`cmd/proxy/main.go`** (MODIFY)
   - Register PlanoProvider in provider map
   - Add switch case for "plano" format (or reuse "openai")

2. **`config.yaml.example`** (MODIFY)
   - Add plano provider example
   - Add subagent mappings using plano

### Unchanged Files

✅ `proxy/internal/provider/provider.go` - Interface unchanged
✅ `proxy/internal/provider/anthropic.go` - No modifications
✅ `proxy/internal/provider/openai.go` - Reused, not modified
✅ `proxy/internal/service/model_router.go` - Already handles any provider
✅ `proxy/internal/handler/handlers.go` - Routing logic is provider-agnostic
✅ `proxy/internal/config/config.go` - Validation already covers plano
✅ `proxy/internal/model/models.go` - No new models needed

### Code Size Estimate

**New code:**
- `plano.go`: ~250 lines (mostly copied from `openai.go`, simplified)
- `main.go` changes: ~10 lines
- `config.yaml.example` changes: ~15 lines

**Total new code:** ~275 lines

**Reused code:**
- `convertAnthropicToOpenAI()`: ~300 lines (already exists)
- `transformOpenAIResponseToAnthropic()`: ~150 lines (already exists)
- `transformOpenAIStreamToAnthropic()`: ~150 lines (already exists)

**Benefit:** ~600 lines of conversion logic reused, only ~275 new lines written.

---

## Testing Strategy

### Unit Tests (Phase 1)

**File:** `proxy/internal/provider/plano_test.go`

**Test cases:**

1. **Provider initialization**
   ```go
   func TestNewPlanoProvider(t *testing.T)
   ```
   - Verify provider name
   - Verify config is stored
   - Verify client timeout is set

2. **Request forwarding (non-streaming)**
   ```go
   func TestPlanoProvider_ForwardRequest_NonStreaming(t *testing.T)
   ```
   - Mock Plano server
   - Send Anthropic request
   - Verify OpenAI request sent to Plano
   - Verify Anthropic response returned

3. **Request forwarding (streaming)**
   ```go
   func TestPlanoProvider_ForwardRequest_Streaming(t *testing.T)
   ```
   - Mock Plano SSE stream
   - Verify Anthropic SSE events generated

4. **Error handling**
   ```go
   func TestPlanoProvider_ErrorHandling(t *testing.T)
   ```
   - Test connection refused
   - Test 404 response
   - Test 500 response
   - Verify Anthropic error format

5. **Tool calls**
   ```go
   func TestPlanoProvider_ToolCalls(t *testing.T)
   ```
   - Send request with tools
   - Verify OpenAI tools format
   - Verify tool_use response conversion

**Minimum:** 5 test cases (meets DoD requirement)

### Integration Tests (Phase 1)

**Test:** End-to-end request through full stack

```bash
# 1. Start Plano locally
planoai up plano_config.yaml

# 2. Start Go proxy
cd proxy && go run cmd/proxy/main.go

# 3. Send test request
curl http://localhost:3001/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: test" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "system": [
      {"type": "text", "text": "You are Claude Code..."},
      {"type": "text", "text": "You are a planning agent..."}
    ],
    "messages": [{"role": "user", "content": "Plan a simple task"}]
  }'
```

**Expected:**
- Request routed to PlanoProvider
- Converted to OpenAI format
- Sent to Plano at `localhost:8080`
- Plano routes to configured model (e.g., Gemini)
- Response converted back to Anthropic format
- Logged to SQLite with `provider="plano"`

### Manual Testing Checklist

- [ ] Plano provider registered in `main.go`
- [ ] Config validation accepts plano provider
- [ ] Subagent mapping routes to plano
- [ ] Non-streaming request works
- [ ] Streaming request works
- [ ] Tool calls work (if supported by Plano model)
- [ ] Error responses formatted correctly
- [ ] Request logged to SQLite
- [ ] Dashboard displays plano requests
- [ ] Existing tests still pass (Anthropic passthrough)
- [ ] Existing OpenAI routing unchanged

---

## Phase Boundary

### What's In Scope (Phase 1)

✅ Basic PlanoProvider implementation
✅ Anthropic↔OpenAI format conversion (reused)
✅ Streaming and non-streaming support
✅ Tool calls support
✅ Error handling (return errors to client)
✅ SQLite logging
✅ Config schema for plano provider
✅ Unit tests (minimum 5)
✅ Integration test (manual)

### What's Deferred

❌ **Phase 2: Error Handling & Fallback**
- Circuit breaker pattern
- Fallback to alternative providers
- Retry logic with backoff
- Enhanced error metrics

❌ **Phase 3: Preference-Based Routing**
- Routing hints to Plano (cost/speed/quality)
- Arch-Router integration
- Per-subagent preferences
- Routing decision capture from Plano

❌ **Phase 3: Load Balancing**
- Multiple Plano instances
- Health checks
- Round-robin distribution

❌ **Phase 4: Guardrails**
- Guardrail configuration
- Blocked request handling
- PII detection
- Dashboard display of guardrail events

### Success Criteria (Phase 1)

**Acceptance criteria from DoD-2025-12-28-035838.md:**

- [x] Design document created: `DESIGN-archgw-provider.md`
- [x] Provider implementation approach decided (reuse OpenAI conversion)
- [x] Configuration schema defined (plano provider in config.yaml)
- [x] Routing decision logic specified (existing ModelRouter handles it)
- [x] Error handling strategy documented (convert all errors to Anthropic format)
- [x] Fallback behavior specified (Phase 2, not Phase 1)
- [x] Format conversion approach documented (reuse OpenAIProvider logic)
- [x] Integration points identified (new plano.go, modify main.go, update config example)
- [x] Logging strategy defined (existing SQLite logging, console output)
- [x] Future extensibility considered (Phase 3 preferences, Phase 4 guardrails)

---

## Key Design Decisions

### 1. Plano vs. OpenAI Provider

**Decision:** Create separate `PlanoProvider` initially, but make it nearly identical to `OpenAIProvider`.

**Rationale:**
- Allows for Plano-specific features later (guardrails, routing hints)
- Clear separation in logging/debugging
- Easy to understand which provider is being used
- Low cost (mostly code copy, small modifications)

**Alternative considered:** Reuse `OpenAIProvider` for Plano.
- ❌ Harder to add Plano-specific features later
- ✅ Less code duplication
- ❌ Confusing in logs/dashboard

**Final decision:** Separate provider for clarity and extensibility.

### 2. Configuration Schema

**Decision:** Use `format: "openai"` for Plano provider.

**Rationale:**
- Plano's API is OpenAI-compatible by design
- Reuses existing validation logic
- Clear indication that OpenAI format conversion is needed
- Consistent with provider interface

**Alternative considered:** Add new format `"plano"`.
- ❌ Requires config validation changes
- ❌ Requires new format handling in provider registration
- ❌ Doesn't add value (Plano uses OpenAI format)

### 3. Error Handling Strategy

**Decision:** Phase 1 returns all errors to client in Anthropic format. No fallback/retry.

**Rationale:**
- Simple implementation
- Clear error messages for debugging
- Matches existing provider behavior
- Fallback adds complexity (defer to Phase 2)

### 4. Code Reuse Strategy

**Decision:** Reuse all conversion functions from `openai.go` without modification.

**Rationale:**
- Conversion logic is identical for OpenAI and Plano
- Reduces code duplication
- Already handles all edge cases (tools, streaming, etc.)
- Well-tested code

### 5. Provider Registration

**Decision:** Register PlanoProvider explicitly in `main.go` based on provider name.

**Rationale:**
- Clear and explicit
- Easy to understand control flow
- Allows for provider-specific initialization if needed later
- Minimal code change

---

## Risk Assessment

### Low Risk

✅ **Format conversion** - Reusing proven OpenAI conversion logic
✅ **Routing integration** - ModelRouter already provider-agnostic
✅ **Configuration** - Existing validation covers plano
✅ **Logging** - Existing SQLite schema captures all needed data

### Medium Risk

⚠️ **Plano API compatibility** - Assume OpenAI compatibility, but needs validation
- **Mitigation:** Test with real Plano instance in Phase 1
- **Fallback:** Adjust conversion if Plano has quirks

⚠️ **Error handling** - Plano error format unknown
- **Mitigation:** Log all error responses for analysis
- **Fallback:** Generic error wrapping if format is unexpected

### Deferred Risk (Phase 2+)

❌ **Circuit breaker tuning** - Failure thresholds need real-world data
❌ **Preference routing** - Arch-Router behavior needs testing
❌ **Guardrails integration** - Response format for blocked requests unknown

---

## Next Steps

### Immediate (Phase 1 Implementation)

1. **Create `plano.go`**
   - Copy `openai.go` as template
   - Remove Authorization header logic
   - Update error messages to reference "Plano"
   - Add constructor `NewPlanoProvider()`

2. **Update `main.go`**
   - Add PlanoProvider registration
   - Test provider initialization

3. **Update `config.yaml.example`**
   - Add plano provider configuration
   - Add example subagent mappings

4. **Write tests**
   - Unit tests for PlanoProvider
   - Integration test script

5. **Manual testing**
   - Start Plano locally
   - Test end-to-end request flow
   - Verify logging

6. **Documentation**
   - Update CLAUDE.md with Plano provider info
   - Document configuration examples

### Phase 2 Planning

- Define circuit breaker thresholds
- Design fallback configuration
- Implement retry logic
- Add Plano-specific metrics

---

## Appendix: Configuration Examples

### Minimal Config

```yaml
providers:
  anthropic:
    base_url: "https://api.anthropic.com"
    format: "anthropic"

  plano:
    base_url: "http://localhost:8080"
    format: "openai"

subagents:
  enable: true
  mappings:
    planner: "plano:gemini/gemini-2.0-flash-exp"
```

### Comprehensive Config

```yaml
providers:
  anthropic:
    base_url: "https://api.anthropic.com"
    format: "anthropic"
    version: "2023-06-01"
    max_retries: 3

  openai:
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com"
    format: "openai"

  plano:
    base_url: "http://localhost:8080"  # Dev: localhost, Docker: http://plano:8080
    format: "openai"

subagents:
  enable: true
  mappings:
    # High-quality subagents → OpenAI
    code-reviewer: "openai:gpt-4o"

    # Cost-effective subagents → Plano (Gemini)
    planner: "plano:gemini/gemini-2.0-flash-exp"
    quick-helper: "plano:gemini/gemini-2.0-flash-exp"

    # Code-focused subagents → Plano (DeepSeek)
    code-generator: "plano:deepseek/deepseek-coder"
    refactorer: "plano:deepseek/deepseek-coder"

    # Budget-conscious subagents → Plano (Qwen)
    summarizer: "plano:qwen/qwen-max"
    translator: "plano:qwen/qwen-max"
```

### Docker Config

```yaml
providers:
  anthropic:
    base_url: "https://api.anthropic.com"
    format: "anthropic"

  plano:
    base_url: "http://plano:8080"  # Docker service name
    format: "openai"

subagents:
  enable: true
  mappings:
    planner: "plano:gemini/gemini-2.0-flash-exp"
```

---

**Design Complete:** 2025-12-28
**Ready for:** Phase 1.1 (Implementation)
**Status:** All acceptance criteria met
