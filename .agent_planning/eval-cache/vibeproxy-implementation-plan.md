# VibeProxy Feature Implementation Plan
Status: READY TO IMPLEMENT
Last Updated: 2025-12-27 00:56:46

## Features to Implement

### 1. Gemini Provider Support
**Effort:** 1-2 days
**Approach:** Configuration only (no code changes)

**Key Discovery:**
- Gemini provides OpenAI-compatible endpoint
- Use existing OpenAIProvider with Gemini base URL
- No custom format converter needed

**Config:**
```yaml
providers:
  gemini:
    base_url: "https://generativelanguage.googleapis.com/v1beta/openai/"
    api_key: "..."
    format: "openai"
```

**Implementation Steps:**
1. Add Gemini config to config.yaml.example
2. Test with Gemini API key
3. Verify streaming works
4. Document setup

---

### 2. Multi-Account Round-Robin
**Effort:** 3-4 days
**Approach:** Config changes + provider modifications

**Design:**
- **Scope:** Per-provider (multiple keys per provider)
- **Algorithm:** Sequential round-robin with atomic counter
- **Failover:** Retry on HTTP 429/503, max 2 attempts
- **Thread Safety:** sync/atomic, no locks needed
- **Integration:** Transparent to subagent routing

**Config Schema:**
```yaml
providers:
  openai:
    format: "openai"
    accounts:
      - api_key: "sk-account1..."
        weight: 1
      - api_key: "sk-account2..."
        weight: 1
    failover:
      retry_on: [429, 503]
      max_attempts: 2
```

**Code Changes:**

1. **config/config.go:**
```go
type ProviderConfig struct {
    Format     string            `yaml:"format"`
    BaseURL    string            `yaml:"base_url"`
    APIKey     string            `yaml:"api_key"`     // Legacy: single key
    Accounts   []AccountConfig   `yaml:"accounts"`    // New: multiple keys
    Failover   FailoverConfig    `yaml:"failover"`
}

type AccountConfig struct {
    APIKey string `yaml:"api_key"`
    Weight int    `yaml:"weight"`
}

type FailoverConfig struct {
    RetryOn     []int `yaml:"retry_on"`
    MaxAttempts int   `yaml:"max_attempts"`
}
```

2. **provider/openai.go:**
```go
type RoundRobinSelector struct {
    accounts []AccountConfig
    counter  atomic.Uint64
}

func (s *RoundRobinSelector) NextAccount() AccountConfig {
    idx := s.counter.Add(1) % uint64(len(s.accounts))
    return s.accounts[idx]
}

// In ForwardRequest:
account := p.accountSelector.NextAccount()
proxyReq.Header.Set("Authorization", "Bearer "+account.APIKey)
```

3. **Failover Logic:**
```go
if resp.StatusCode in p.config.Failover.RetryOn {
    for i := 0; i < p.config.Failover.MaxAttempts; i++ {
        account = p.accountSelector.NextAccount()
        // retry with next account
    }
}
```

---

## Use Cases

**Primary:** Rate limit distribution
- Distribute Claude Code requests across multiple API keys
- Increase effective rate limit by 2-3x (with 2-3 accounts)

**Secondary:** Cost tracking per account/project
- Assign different API keys to different projects
- Track usage/cost per account

---

## Integration with Existing Code

**Subagent Routing:**
- No changes needed to subagent mappings
- Account selection happens inside provider.ForwardRequest()
- Transparent to routing layer

**Example:**
```yaml
subagents:
  mappings:
    code-reviewer: "openai:gpt-4o"  # Same format, multi-account works automatically
```

**Request Flow:**
1. ModelRouter extracts "openai" + "gpt-4o" from mapping
2. Handler looks up "openai" provider
3. OpenAIProvider selects account via round-robin
4. Request forwarded with selected account's API key

---

## Testing Requirements

1. **Config validation test:**
   - Single `api_key` format still works
   - Multi-account `accounts` array works
   - Both cannot be specified (error)

2. **Round-robin distribution test:**
   - Equal weights produce equal distribution
   - Atomic counter is thread-safe

3. **Failover integration test:**
   - 429 response triggers retry with next account
   - Max attempts limit is respected
   - Non-retryable errors don't retry

---

## Backward Compatibility

**Requirement:** Support both single `api_key` and multi-account `accounts` formats.

**Strategy:**
- If `accounts` exists, use multi-account
- If `api_key` exists (and no `accounts`), treat as single-account
- Error if both are specified

---

## Open Questions (Minor)

1. **Gemini auth:** Assume standard `Authorization: Bearer <key>` (verify in docs)
2. **Account visibility:** Should logs show which account was used? (debuggability)
3. **Weighted distribution:** Support weights other than 1? (future enhancement)

**Impact:** All low - can be addressed during implementation.
