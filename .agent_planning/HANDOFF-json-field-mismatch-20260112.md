# Handoff: JSON Field Name Mismatch Bug & Prevention

**Created**: 2026-01-12T03:50:00
**For**: Any agent implementing API fixes or new endpoints
**Status**: COMPLETED

---

## Objective

Fix the JSON field name mismatch between Go backend (PascalCase) and TypeScript frontend (snake_case), AND establish automated safeguards to prevent this class of bug from ever happening again.

## Current State

### What's Been Done
- ✅ Root cause identified: Go structs serialize to PascalCase, TypeScript expects snake_case
- ✅ Verified API returns data correctly (all fields populated)
- ✅ Verified frontend code is correct (accessing `config.base_url`, `config.format`, etc.)
- ✅ Confirmed the mismatch is in JSON serialization only

### What's In Progress
- 🔄 Need to add JSON tags to Go structs OR update TypeScript types

### What Remains
- [ ] Fix the immediate issue (add JSON tags or update TS types)
- [ ] Add automated contract testing to prevent regression
- [ ] Document the API contract pattern for future development

## Context & Background

### Why We're Doing This
The Provider Routing page shows provider names but no details (base URL, format, version, retries). Users reported "can't see any details" - the data is being returned by the API but the frontend can't read it due to field name case mismatch.

### Key Decisions Made
| Decision | Rationale | Date |
|----------|-----------|------|
| Go returns PascalCase by default | Go's json package uses struct field names as-is unless tagged | - |
| TypeScript types use snake_case | Following common JS/REST API conventions | - |
| **Need to decide**: Fix Go or TS? | See options below | 2026-01-12 |

### The Bug in Detail

**API Response (actual)**:
```json
{
  "anthropic": {
    "Format": "anthropic",
    "BaseURL": "https://api.anthropic.com",
    "APIKey": "",
    "Version": "2023-06-01",
    "MaxRetries": 3
  }
}
```

**TypeScript expects**:
```typescript
interface ProviderConfig {
  format: string;
  base_url: string;
  api_key?: string;
  version?: string;
  max_retries?: number;
}
```

**Frontend code accessing** (in `Routing.tsx`):
```typescript
<div>{config.base_url}</div>  // undefined! Should be config.BaseURL
<div>{config.format}</div>     // undefined! Should be config.Format
```

### Important Constraints
- Must maintain backward compatibility with any existing API consumers
- Go's default JSON behavior is PascalCase unless explicitly tagged
- TypeScript types are already defined with snake_case
- The fix should be consistent across ALL API endpoints, not just this one

## The Root Problem: No Contract Enforcement

**Why this will keep happening without systemic fix:**

1. **No shared schema** - Go and TypeScript types are defined independently
2. **No contract tests** - Nothing verifies the API response matches TypeScript expectations
3. **Manual coordination** - Developers must remember to match case conventions
4. **Silent failures** - TypeScript doesn't error when fields are undefined (just shows nothing)

## Fix Options

### Option A: Add JSON Tags to Go (RECOMMENDED)

**Pros:**
- Go code explicitly declares JSON field names
- Matches REST API conventions (snake_case)
- One-time fix, then enforced by Go compiler
- Works with existing TypeScript types

**Cons:**
- Must update all Go struct fields
- Slightly more verbose Go code

**Implementation:**
```go
type ProviderConfig struct {
    Format           string `yaml:"format" json:"format"`
    BaseURL          string `yaml:"base_url" json:"base_url"`
    APIKey           string `yaml:"api_key" json:"api_key"`
    Version          string `yaml:"version" json:"version"`
    MaxRetries       int    `yaml:"max_retries" json:"max_retries"`
    FallbackProvider string `yaml:"fallback_provider" json:"fallback_provider"`
    CircuitBreaker   CircuitBreakerConfig `yaml:"circuit_breaker" json:"circuit_breaker"`
}
```

### Option B: Update TypeScript to Match Go

**Pros:**
- No backend changes needed

**Cons:**
- Unusual in JS ecosystem (PascalCase for API fields)
- Must update ALL TypeScript types
- Easy to forget in future

### Option C: Transform in API Handler

**Pros:**
- No struct changes

**Cons:**
- Adds runtime overhead
- Easy to forget for new endpoints
- Hides the real shape of data

## Preventing Future Occurrences

### Systemic Fix 1: API Contract Tests (HIGH PRIORITY)

Create a test that verifies API responses match TypeScript expectations:

```go
// proxy/internal/handler/api_contract_test.go
func TestProviderConfigJSONContract(t *testing.T) {
    // This test ensures our Go structs serialize to the expected JSON format
    cfg := &config.ProviderConfig{
        Format:   "anthropic",
        BaseURL:  "https://api.anthropic.com",
        Version:  "2023-06-01",
    }

    data, _ := json.Marshal(cfg)
    var result map[string]interface{}
    json.Unmarshal(data, &result)

    // Assert snake_case field names
    require.Contains(t, result, "format", "Expected snake_case field 'format'")
    require.Contains(t, result, "base_url", "Expected snake_case field 'base_url'")
    require.NotContains(t, result, "Format", "Should NOT have PascalCase 'Format'")
    require.NotContains(t, result, "BaseURL", "Should NOT have PascalCase 'BaseURL'")
}
```

### Systemic Fix 2: OpenAPI/Swagger Schema (MEDIUM PRIORITY)

Generate TypeScript types from a shared OpenAPI spec:

1. Define API in `openapi.yaml`
2. Generate Go server stubs with correct JSON tags
3. Generate TypeScript client types
4. Both sides use same source of truth

### Systemic Fix 3: Linting Rule (LOW PRIORITY)

Add a Go linter that warns about missing JSON tags on exported struct fields:

```yaml
# .golangci.yml
linters:
  enable:
    - tagliatelle  # Checks struct tag consistency
```

## Acceptance Criteria

How we'll know this is complete:

- [x] Provider Routing page shows all provider details (format, base_url, version, max_retries)
- [x] API contract test exists and passes
- [x] All config-related Go structs have explicit JSON tags
- [x] CLAUDE.md updated with API contract conventions

## Scope

### Files to Modify
- `proxy/internal/config/config.go` - Add JSON tags to all structs
- `proxy/internal/handler/api_contract_test.go` - New contract test file
- `CLAUDE.md` - Document the convention

### Related Components
- `proxy/internal/handler/data_handler.go` - Uses config structs
- `proxy/internal/handler/handlers_v2.go` - Uses config structs
- `dashboard/src/lib/types.ts` - TypeScript types (should match after fix)
- `dashboard/src/pages/Routing.tsx` - Will work after fix

### Out of Scope
- OpenAPI spec generation (good future improvement)
- Full type generation pipeline (overkill for now)

## Implementation Approach

### Recommended Steps

1. **Add JSON tags to Go config structs** (5 min)
   - Edit `proxy/internal/config/config.go`
   - Add `json:"field_name"` to all exported fields
   - Focus on: ProviderConfig, CircuitBreakerConfig, SubagentsConfig

2. **Verify the fix** (2 min)
   - Restart backend: `just restart-data` or `just docker`
   - Check API response: `curl http://localhost:8000/api/v2/config/providers | jq`
   - Confirm fields are now snake_case

3. **Create contract test** (10 min)
   - Add `proxy/internal/handler/api_contract_test.go`
   - Test all config structs serialize with correct field names
   - Run: `go test ./proxy/internal/handler/... -run Contract`

4. **Document the convention** (5 min)
   - Add to CLAUDE.md under a new "API Conventions" section
   - State: "All Go structs exposed via API must have explicit JSON tags using snake_case"

### Patterns to Follow
- Look at `model/models.go` for examples of structs with proper JSON tags
- Use `json:"field_name,omitempty"` for optional fields

### Known Gotchas
- CircuitBreakerConfig has a `TimeoutDuration` field marked `yaml:"-"` - should also be `json:"-"`
- Some fields may need `omitempty` to avoid sending zero values

## Reference Materials

### Planning Documents
- [STATUS-20260112.md](.agent_planning/web-routing-configuration/STATUS-20260112.md) - Investigation status

### Codebase References
- `proxy/internal/config/config.go:38-56` - ProviderConfig struct (needs JSON tags)
- `proxy/internal/model/models.go` - Example of proper JSON tagging
- `dashboard/src/lib/types.ts:286-297` - TypeScript ProviderConfig interface
- `dashboard/src/pages/Routing.tsx:17-61` - ProviderCard component

## Questions & Blockers

### Open Questions
- [x] Root cause identified? **YES** - JSON field name case mismatch
- [ ] Should we also add OpenAPI spec? **Defer to later**

### Current Blockers
- None - ready to implement

### Need User Input On
- None - clear path forward

## Testing Strategy

### Existing Tests
- `proxy/internal/handler/data_handler_test.go` - Handler tests exist
- Coverage: Unknown for config serialization

### New Tests Needed
- [ ] Contract test: ProviderConfig serializes with snake_case fields
- [ ] Contract test: SubagentsConfig serializes with snake_case fields
- [ ] Contract test: CircuitBreakerConfig serializes with snake_case fields

### Manual Testing
- [ ] Visit http://localhost:8173/routing
- [ ] Verify provider cards show: format badge, base URL, version, max retries
- [ ] Verify subagent mappings still display correctly

## Success Metrics

How to validate implementation:

- All existing tests pass (`just test`)
- New contract tests pass
- Provider Routing page shows all details (visual verification)
- Future API changes will fail tests if JSON tags are missing

---

## Next Steps for Agent

**Immediate actions**:
1. Edit `proxy/internal/config/config.go` - add JSON tags to ProviderConfig, CircuitBreakerConfig
2. Test: `curl http://localhost:8000/api/v2/config/providers | jq` - verify snake_case
3. Create `proxy/internal/handler/api_contract_test.go` with contract tests

**Before starting implementation**:
- [x] Review this handoff document
- [ ] Verify backend is running (`just docker`)

**When complete**:
- [ ] Update CLAUDE.md with API conventions
- [ ] Mark todo items complete
- [ ] Take screenshot showing provider details displayed
