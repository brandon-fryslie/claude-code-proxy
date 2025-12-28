# ArchGW/Plano Research & Setup Guide

**Date:** 2025-12-28
**Phase:** P0.1 - Research & Local Setup
**Status:** Complete

---

## Executive Summary

**ArchGW has been renamed to "Plano"** (https://planoai.dev, https://github.com/katanemo/archgw)

Plano is a **models-native proxy server and data plane for agents**, built on Envoy by its core contributors. It provides intelligent LLM routing, guardrails, and observability features. For claude-code-proxy, we'll use it as an **LLM router** to access 11+ providers through a single unified interface.

### Key Facts

- **Technology:** Rust/Envoy-based proxy
- **Input Format:** OpenAI `/v1/chat/completions` (also supports `/v1/messages` for Anthropic format)
- **Deployment:** Docker + Python CLI (`planoai`)
- **Port:** Default 8080 (configurable)
- **Configuration:** YAML file (`plano_config.yaml`)
- **Free Tier:** Hosted Arch-Router models available in US-central region

---

## What is Plano/ArchGW?

### Purpose

Plano is **middleware for agentic applications** that handles:

1. **LLM Routing** - Route requests to 11+ providers (OpenAI, Anthropic, DeepSeek, Gemini, Mistral, Groq, etc.)
2. **Intelligent Routing** - Policy-based routing using the Arch-Router 1.5B model
3. **Guardrails** - Jailbreak detection, content moderation, PII detection
4. **Orchestration** - Multi-agent coordination
5. **Observability** - OpenTelemetry traces, metrics, and "agentic signals"

### For claude-code-proxy

We're using **only the LLM routing capability** to:
- Access providers beyond Anthropic/OpenAI (Gemini, DeepSeek, Qwen, Mistral, Groq)
- Implement preference-based routing (cost/speed/quality optimization)
- Provide failover across multiple providers
- Centralize provider API key management

Plano runs as a **sidecar service** alongside our Go proxy.

---

## Installation & Prerequisites

### System Requirements

1. **Docker** (v24+)
2. **Docker Compose** (v2.29+)
3. **Python** (v3.13+) - for CLI tool

### Installing Plano CLI

The `planoai` CLI manages Plano via Docker:

```bash
# Recommended: Install with uv
uv tool install planoai==0.4.0

# Alternative: Install with pip
python3 -m venv venv
source venv/bin/activate
pip install planoai==0.4.0
```

**Installed executable:** `planoai`

### Verifying Installation

```bash
planoai --version
# Output: Shows planoai cli version

planoai --help
# Commands:
#   up    - Starts Plano (pulls Docker images, starts containers)
#   down  - Stops Plano
#   logs  - Stream access logs
#   build - Build from source (for development)
```

---

## Supported Providers

Plano supports **11+ LLM providers** out of the box. Each provider requires:
1. **Provider prefix** in model name (e.g., `openai/`, `gemini/`, `deepseek/`)
2. **API key** (passed as environment variable)
3. **Model ID** (specific model to use)

### Provider Summary Table

| Provider | Prefix | API Key Env Var | Example Model | Get API Key |
|----------|--------|----------------|---------------|-------------|
| **OpenAI** | `openai/` | `OPENAI_API_KEY` | `openai/gpt-4o` | [platform.openai.com/api-keys](https://platform.openai.com/api-keys) |
| **Anthropic** | `anthropic/` | `ANTHROPIC_API_KEY` | `anthropic/claude-sonnet-4-5` | [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys) |
| **Google Gemini** | `gemini/` | `GEMINI_API_KEY` | `gemini/gemini-2.0-flash-exp` | [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey) |
| **DeepSeek** | `deepseek/` | `DEEPSEEK_API_KEY` | `deepseek/deepseek-chat` | [platform.deepseek.com/api_keys](https://platform.deepseek.com/api_keys) |
| **Mistral AI** | `mistral/` | `MISTRAL_API_KEY` | `mistral/mistral-large-latest` | [console.mistral.ai/api-keys](https://console.mistral.ai/api-keys/) |
| **Groq** | `groq/` | `GROQ_API_KEY` | `groq/llama-3.3-70b-versatile` | [console.groq.com/keys](https://console.groq.com/keys) |
| **Alibaba Qwen** | `qwen/` | `DASHSCOPE_API_KEY` | `qwen/qwen-max` | [dashscope.console.aliyun.com](https://dashscope.console.aliyun.com) |
| **Together AI** | `together/` | `TOGETHER_API_KEY` | `together/meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo` | [api.together.xyz/settings/api-keys](https://api.together.xyz/settings/api-keys) |
| **Fireworks AI** | `fireworks/` | `FIREWORKS_API_KEY` | `fireworks/accounts/fireworks/models/llama-v3p3-70b-instruct` | [fireworks.ai/api-keys](https://fireworks.ai/api-keys) |
| **Perplexity** | `perplexity/` | `PERPLEXITY_API_KEY` | `perplexity/llama-3.1-sonar-large-128k-online` | [perplexity.ai/settings/api](https://perplexity.ai/settings/api) |
| **Cerebras** | `cerebras/` | `CEREBRAS_API_KEY` | `cerebras/llama3.3-70b` | [cloud.cerebras.ai/](https://cloud.cerebras.ai/) |

**Note:** All providers support streaming and non-streaming responses. Plano handles format conversion internally.

---

## Configuration File Format

### File Location

Default: `plano_config.yaml` (pass path to `planoai up <file>`)

### Minimal Configuration

```yaml
version: v0.1.0

listeners:
  - name: egress_traffic
    address: 0.0.0.0
    port: 8080
    message_format: openai
    timeout: 30s

llm_providers:
  - model: openai/gpt-4o
    access_key: $OPENAI_API_KEY
```

### Configuration with Multiple Providers

```yaml
version: v0.1.0

listeners:
  - name: egress_traffic
    address: 0.0.0.0
    port: 8080
    message_format: openai  # Accept OpenAI format requests
    timeout: 30s

llm_providers:
  # OpenAI GPT-4o
  - model: openai/gpt-4o
    access_key: $OPENAI_API_KEY

  # Google Gemini 2.0 Flash
  - model: gemini/gemini-2.0-flash-exp
    access_key: $GEMINI_API_KEY

  # DeepSeek Chat
  - model: deepseek/deepseek-chat
    access_key: $DEEPSEEK_API_KEY

  # DeepSeek Coder (specialized for code)
  - model: deepseek/deepseek-coder
    access_key: $DEEPSEEK_API_KEY
```

### Configuration with Routing Preferences (Advanced)

Plano can intelligently route based on **natural language descriptions** using the Arch-Router model:

```yaml
llm_providers:
  - model: openai/gpt-4o
    access_key: $OPENAI_API_KEY
    routing_preferences:
      - name: complex_reasoning
        description: Deep analysis, mathematical problem solving, logical reasoning
      - name: high_quality
        description: High quality responses requiring accuracy

  - model: gemini/gemini-2.0-flash-exp
    access_key: $GEMINI_API_KEY
    routing_preferences:
      - name: fast_generation
        description: Fast code generation, quick responses
      - name: experimental
        description: Testing new model capabilities

  - model: deepseek/deepseek-coder
    access_key: $DEEPSEEK_API_KEY
    routing_preferences:
      - name: code_generation
        description: Generating new code, writing functions, creating scripts
      - name: code_review
        description: Analyzing existing code for bugs, improvements, optimization
```

**How it works:** When you send a request, Plano analyzes the prompt using the 1.5B Arch-Router model and automatically selects the best provider based on your preference descriptions. No manual routing logic needed.

---

## Running Plano Locally

### Step 1: Set API Keys

Export API keys as environment variables:

```bash
# Required for testing (choose at least one)
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export DEEPSEEK_API_KEY="sk-..."

# Optional additional providers
export ANTHROPIC_API_KEY="sk-ant-..."
export MISTRAL_API_KEY="..."
export GROQ_API_KEY="gsk_..."
export DASHSCOPE_API_KEY="..."  # For Qwen
```

### Step 2: Create Configuration File

Create `plano_config.yaml` with your desired providers (see examples above).

### Step 3: Start Plano

```bash
# Start Plano (pulls Docker images if needed, starts containers)
planoai up plano_config.yaml

# Or run in foreground for debugging
planoai up plano_config.yaml --foreground

# Check logs
planoai logs
```

**What happens:**
1. CLI validates `plano_config.yaml`
2. Pulls Plano Docker images (first time only)
3. Starts Plano container with your config
4. Exposes port 8080 (or configured port)
5. Plano is ready to accept requests

### Step 4: Stop Plano

```bash
planoai down
```

---

## Testing Plano

### Test 1: Simple Request (Model-Based Routing)

Send a request directly specifying the model:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o",
    "messages": [{"role": "user", "content": "What is 2+2?"}],
    "max_tokens": 100
  }'
```

**Expected:** Response from OpenAI GPT-4o

### Test 2: Streaming Request

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini/gemini-2.0-flash-exp",
    "messages": [{"role": "user", "content": "Count to 5"}],
    "stream": true
  }'
```

**Expected:** Server-sent events (SSE) with streamed response chunks

### Test 3: Preference-Based Routing (Advanced)

If you configured `routing_preferences`, send a request and let Plano choose:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-Arch-Preference: fast_generation" \
  -d '{
    "messages": [{"role": "user", "content": "Write a Python function to sort a list"}]
  }'
```

**Expected:** Plano routes to the provider with `fast_generation` preference

### Test 4: Error Handling

Test with invalid model:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "invalid/model",
    "messages": [{"role": "user", "content": "test"}]
  }'
```

**Expected:** Clear error message from Plano

---

## Measuring Latency

To measure Plano's overhead:

```bash
# Test direct OpenAI API (baseline)
time curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hi"}], "max_tokens": 10}'

# Test through Plano
time curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "openai/gpt-4o", "messages": [{"role": "user", "content": "Hi"}], "max_tokens": 10}'
```

**Target:** Plano overhead should be <100ms

**Measured Results:** *(To be filled during testing)*

---

## Integration with claude-code-proxy

### Architecture Overview

```
Client (Claude Code)
    |
    v
Go Proxy (port 3001)
    |
    +-- Anthropic requests → Anthropic API (passthrough, unchanged)
    +-- OpenAI requests → OpenAI API (existing flow, unchanged)
    +-- ArchGW requests → Plano (port 8080) → Multi-provider routing
    |
    v
SQLite Logging + React Dashboard
```

### Request Flow for ArchGW Routing

1. **Client** sends request to Go proxy at `localhost:3001/v1/messages`
2. **ModelRouter** detects subagent prompt (via prompt hashing)
3. **Routing decision:** If mapping is `archgw:gemini-2.0-flash-exp`, route to ArchGW provider
4. **ArchGWProvider** converts Anthropic request → OpenAI format
5. **Forward** to Plano at `http://localhost:8080/v1/chat/completions`
6. **Plano** routes to Google Gemini API
7. **Response** flows back: Gemini → Plano → ArchGWProvider (convert to Anthropic) → Client
8. **Logging:** Both original Anthropic request and ArchGW-routed request logged to SQLite

### Configuration in config.yaml

```yaml
providers:
  anthropic:
    base_url: "https://api.anthropic.com"
    format: "anthropic"

  openai:
    api_key: "..."
    base_url: "https://api.openai.com"
    format: "openai"

  archgw:
    base_url: "http://localhost:8080"  # Plano running locally
    format: "openai"  # ArchGW expects OpenAI format

subagents:
  enable: true
  mappings:
    code-reviewer: "archgw:gemini/gemini-2.0-flash-exp"
    planner: "archgw:deepseek/deepseek-chat"
    budget-agent: "archgw:qwen/qwen-max"
```

---

## Docker Deployment (Phase 1 P2)

For production deployment, both Plano and the Go proxy will run in Docker Compose:

```yaml
version: '3.8'

services:
  archgw:
    image: katanemo/plano:latest  # Official Plano image
    volumes:
      - ./plano_config.yaml:/app/plano_config.yaml:ro
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
    networks:
      - proxy-net
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  proxy:
    build: ./proxy
    ports:
      - "3001:3001"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - ARCHGW_BASE_URL=http://archgw:8080  # Internal Docker network
    depends_on:
      archgw:
        condition: service_healthy
    networks:
      - proxy-net

networks:
  proxy-net:
    driver: bridge
```

**Note:** In Docker, the Go proxy references Plano as `http://archgw:8080` (service name, not localhost).

---

## Routing Modes

Plano supports three routing modes:

### 1. Model-Based Routing (Simple)

**Direct model selection** - You specify exactly which model to use:

```json
{
  "model": "gemini/gemini-2.0-flash-exp",
  "messages": [{"role": "user", "content": "Hello"}]
}
```

**Use case:** When you know exactly which model you want (e.g., subagent mappings)

### 2. Alias-Based Routing (Semantic Names)

**Semantic model names** - Decouple application from specific providers:

```yaml
# In plano_config.yaml
model_aliases:
  fast-code-gen: "gemini/gemini-2.0-flash-exp"
  high-quality: "openai/gpt-4o"
  budget-friendly: "deepseek/deepseek-chat"
```

```json
{
  "model": "fast-code-gen",  // Resolves to Gemini
  "messages": [{"role": "user", "content": "Hello"}]
}
```

**Use case:** When you want to change providers without changing application code

### 3. Preference-Based Routing (Intelligent)

**Automatic selection** - Plano analyzes the prompt and chooses the best model:

```yaml
llm_providers:
  - model: openai/gpt-4o
    routing_preferences:
      - name: complex_reasoning
        description: "deep analysis, mathematical problem solving"

  - model: gemini/gemini-2.0-flash-exp
    routing_preferences:
      - name: fast_generation
        description: "fast code generation, quick responses"
```

```bash
# Request with preference hint
curl -H "X-Arch-Preference: fast_generation" ...

# Or let Plano analyze the prompt automatically
curl ... -d '{"messages": [{"role": "user", "content": "Write a Python function quickly"}]}'
```

**Use case:** Optimize for cost, speed, or quality without manual routing logic (Phase 3)

---

## Guardrails (Phase 4)

Plano supports built-in guardrails for safety and moderation:

### Available Guardrails

1. **Jailbreak Detection** - Detect attempts to bypass model safety
2. **Content Moderation** - Filter violence, hate speech, sexual content
3. **PII Detection** - Identify emails, phone numbers, SSNs, credit cards
4. **Custom Filters** - User-defined content filters

### Configuration Example

```yaml
guardrails:
  enabled: true
  policies:
    jailbreak_detection:
      action: "block"  # block | warn | log
      sensitivity: "medium"  # low | medium | high

    content_moderation:
      action: "warn"
      categories:
        - "violence"
        - "hate_speech"
        - "sexual_content"

    pii_detection:
      action: "log"
      redact: false  # Redact PII from logs
      types:
        - "email"
        - "phone"
        - "ssn"
```

### Guardrail Responses

When guardrails trigger:
- **Block:** Returns 403 with guardrail reason in headers
- **Warn:** Allows request but logs warning
- **Log:** Silent logging for analysis

**Note:** Guardrails integration is deferred to Phase 4.

---

## Observability

Plano provides rich observability features:

### OpenTelemetry Integration

- **Traces:** Distributed tracing across Plano → Provider
- **Metrics:** Request latency, error rates, token usage
- **Agentic Signals:** Custom metrics for agentic applications

### Configuration

```yaml
tracing:
  random_sampling: 100  # 100% sampling for development
  endpoint: "http://jaeger:14268/api/traces"  # Optional: Export to Jaeger/OTEL collector
```

### Access Logs

```bash
# Stream access logs
planoai logs

# Example log entry
{
  "timestamp": "2025-12-28T14:45:00Z",
  "request_id": "abc123",
  "model": "gemini/gemini-2.0-flash-exp",
  "provider": "gemini",
  "latency_ms": 850,
  "tokens": {"prompt": 45, "completion": 120},
  "status": 200
}
```

**Note:** For claude-code-proxy, we'll capture Plano routing decisions and log them to SQLite (Phase 2).

---

## Key Findings & Recommendations

### What We Learned

1. **Naming:** ArchGW is now called Plano (repository still `katanemo/archgw`)
2. **Easy Setup:** Python CLI (`planoai`) handles all Docker complexity
3. **Provider Support:** 11+ providers out of the box with consistent interface
4. **Format Conversion:** Plano accepts OpenAI format, handles provider differences internally
5. **Intelligent Routing:** Arch-Router 1.5B model enables preference-based routing without manual logic
6. **Free Tier:** Hosted Arch-Router available (no local model hosting needed)

### Recommended Approach for Integration

1. **Phase 1:** Use **model-based routing** only (simple, predictable)
   - Config: `code-reviewer: "archgw:gemini/gemini-2.0-flash-exp"`
   - Plano routes directly to specified model

2. **Phase 3:** Add **preference-based routing** (intelligent optimization)
   - Config: `code-reviewer: "archgw:preference:fast_generation"`
   - Plano analyzes prompt and selects best model

3. **Phase 4:** Enable **guardrails** (safety, moderation)
   - Protect against jailbreaks and harmful content

### API Key Requirements

**Minimum for testing (Phase 1):**
- At least 1 provider (e.g., OpenAI or Gemini)

**Recommended for comprehensive testing:**
- OpenAI (baseline comparison)
- Gemini (fast, experimental)
- DeepSeek (cost-effective, code-focused)

**Optional (expand later):**
- Qwen, Mistral, Groq, Anthropic, etc.

### Performance Expectations

- **Target Latency:** <100ms overhead for Plano hop
- **Streaming:** Fully supported (SSE pass-through)
- **Reliability:** Circuit breakers needed (Phase 2) for production use

---

## Next Steps

### Immediate (Phase 0.2 - Design)

1. Create architecture diagram showing Go proxy → Plano flow
2. Define ArchGWProvider interface implementation
3. Design config schema for `providers.archgw` section
4. Document format conversion strategy (reuse OpenAI conversion)
5. Specify error handling approach
6. Create design document: `DESIGN-archgw-provider.md`

### Phase 1 (Implementation)

1. Implement `proxy/internal/provider/archgw.go`
2. Update `config.yaml.example` with archgw provider
3. Add subagent mapping support for archgw
4. Write unit tests for ArchGW provider
5. Create Docker Compose setup
6. Verify end-to-end integration

---

## References

- **Website:** https://planoai.dev
- **GitHub:** https://github.com/katanemo/archgw
- **Documentation:** https://docs.planoai.dev
- **Supported Providers:** https://docs.planoai.dev/concepts/llm_providers/supported_providers.html
- **Quickstart Guide:** https://docs.planoai.dev/get_started/quickstart.html
- **Discord Support:** https://discord.gg/pGZf2gcwEc

---

## Appendix: Environment Variables Summary

```bash
# Core Providers (choose at least one for testing)
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export DEEPSEEK_API_KEY="sk-..."

# Additional Providers (optional)
export ANTHROPIC_API_KEY="sk-ant-..."
export MISTRAL_API_KEY="..."
export GROQ_API_KEY="gsk_..."
export DASHSCOPE_API_KEY="..."  # Qwen
export TOGETHER_API_KEY="..."
export FIREWORKS_API_KEY="..."
export PERPLEXITY_API_KEY="..."
export CEREBRAS_API_KEY="..."
```

---

## Appendix: plano_config.yaml Template

```yaml
version: v0.1.0

listeners:
  - name: egress_traffic
    address: 0.0.0.0
    port: 8080
    message_format: openai
    timeout: 30s

llm_providers:
  # OpenAI
  - model: openai/gpt-4o
    access_key: $OPENAI_API_KEY

  # Google Gemini
  - model: gemini/gemini-2.0-flash-exp
    access_key: $GEMINI_API_KEY

  # DeepSeek
  - model: deepseek/deepseek-chat
    access_key: $DEEPSEEK_API_KEY

  - model: deepseek/deepseek-coder
    access_key: $DEEPSEEK_API_KEY

# Optional: Enable tracing
tracing:
  random_sampling: 100

# Optional: Guardrails (Phase 4)
# guardrails:
#   enabled: false
```

---

**Research Complete:** 2025-12-28
**Ready for:** Phase 0.2 (Design ArchGW Provider Integration)
