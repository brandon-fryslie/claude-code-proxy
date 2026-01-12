# Handoff: cc-viz as Separate Project

**Created**: 2026-01-12T19:30:00Z
**For**: Agent continuing cc-viz extraction and development
**Status**: Ready to Start - Planning Phase Complete

---

## Objective

Extract the Conversations tab from the proxy dashboard into a standalone cc-viz (Claude Code Visualizations) project. cc-viz will be a dedicated webapp for exploring and visualizing Claude conversation history, separate from the main proxy dashboard but consuming the same proxy-data API.

---

## Current State

### What's Been Done
- ✅ Evaluated current monorepo structure
- ✅ Planned proxy-core extraction (Phase 1 of restructuring)
- ✅ Identified Conversations UI components in proxy dashboard
- ✅ Confirmed proxy-data API is ready to serve conversation data
- ✅ Created conversation indexing backend (database tables, FTS5 search)
- ✅ Created message pagination endpoints (`/api/v2/conversations/{id}/messages`)
- ✅ Implemented conversation browser component in dashboard

### What's In Progress
- None yet - ready for Phase 1 (proxy-core extraction)

### What Remains - Before cc-viz Extraction
1. **Phase 1: Extract proxy-core** (separate sprint)
   - Move proxy-core into standalone module with independent go.mod
   - Update build system and Docker for split modules
   - Establish monorepo pattern for future phases

2. **Phase 2: Extract cc-viz** (this sprint, after Phase 1)
   - Create `cc-viz/` project structure
   - Move Conversations UI from proxy dashboard → cc-viz
   - Update proxy dashboard to remove Conversations tab
   - Configure cc-viz to call proxy-data API
   - Update Justfile and docker-compose for new service

3. **Future: Expand cc-viz**
   - Add conversation search with filters (unified search across all conversations)
   - Add message visualization features
   - Add analytics/insights about conversations
   - Add export/sharing capabilities

---

## Context & Background

### Why We're Doing This

**Problem**: The proxy dashboard mixes request monitoring (core proxy functionality) with conversation browsing (Claude Code feature). These are separate concerns with different users and evolution paths.

**Solution**: Extract Conversations into cc-viz, a dedicated visualization tool for Claude conversation history. This allows:
- Proxy dashboard stays minimal and stable (focused on API proxy monitoring)
- cc-viz evolves as a specialized conversation exploration/analysis tool
- Clear separation of concerns and independent deployment
- Foundation for future visualization features (analytics, patterns, etc.)

**Architecture Principle**: Keep proxy-core stable for production use, allow proxy-data and cc-viz to evolve with new features.

### Key Decisions Made

| Decision | Rationale | Date |
|----------|-----------|------|
| Extract to separate project vs. component | Allows independent versioning, deployment, and feature development | 2026-01-12 |
| cc-viz consumes proxy-data API (not direct DB) | Maintains clean architecture, API can change storage without affecting UI | 2026-01-12 |
| Extract after proxy-core split | proxy-core extraction establishes monorepo pattern, reduces variables | 2026-01-12 |
| Start with Conversations tab only | Minimum viable extraction, can expand cc-viz features later | 2026-01-12 |

### Important Constraints

- **Proxy API stability**: cc-viz depends on proxy-data API endpoints. All endpoints must remain stable during extraction.
- **No breaking changes to proxy**: Existing dashboard must continue working without Conversations tab.
- **Monorepo structure**: Both proxy and cc-viz live in same repo. Build system must handle both.
- **Docker orchestration**: docker-compose must start proxy-core, proxy-data, and cc-viz services correctly.
- **Authentication**: Use same auth patterns as current dashboard (if any).
- **Configuration**: cc-viz must know proxy-data base URL (can be localhost:8000 in dev, configure for prod).

---

## Scope

### Files to Extract from proxy dashboard → cc-viz

**Components**:
- `dashboard/src/pages/Conversations.tsx` - Main page component
- `dashboard/src/components/features/GlobalSearchResults.tsx` - Search results display (new)
- `dashboard/src/components/features/ConversationSearch.tsx` - Search input with filters
- `dashboard/src/components/features/ConversationThread.tsx` - Message thread display
- `dashboard/src/components/features/ConversationList.tsx` - Conversation list
- `dashboard/src/lib/search.ts` - Search utilities
- `dashboard/src/lib/types.ts` - Types for conversations (SearchMessage, etc.)

**What to Keep in proxy dashboard**:
- Everything EXCEPT Conversations tab/routes
- Dashboard, Requests, Token Usage, Performance, Provider Routing, Settings tabs
- Sidebar navigation (remove Conversations link)
- Shared UI components that dashboard still uses

### Related API Endpoints (in proxy-data)

These are already implemented and ready for cc-viz to consume:

- `GET /api/v2/conversations` - List conversations
- `GET /api/v2/conversations/{id}` - Get conversation details
- `GET /api/v2/conversations/{id}/messages?limit=N&offset=M&include_subagents=true` - Paginated messages
- `GET /api/v2/conversations/search?q=<query>&types=<filter>` - Search conversations
- `POST /api/v2/conversations/reindex` - Trigger re-indexing

### Out of Scope (This Sprint)

- Proxy dashboard modifications (will be follow-up after cc-viz extraction)
- Unified search with filters (planned for future cc-viz expansion)
- Advanced visualization features (planned for future)
- Analytics/insights (planned for future)

---

## Current Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ claude-code-proxy monorepo                                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  proxy-core (PLANNED Phase 1)        proxy-data            │
│  ├── cmd/proxy-core                  ├── cmd/proxy-data    │
│  ├── internal/handler/              ├── internal/handler/  │
│  │   └── core_handler.go             │   └── data_handler  │
│  ├── internal/provider/              ├── internal/service/  │
│  │   ├── anthropic.go                │   ├── storage       │
│  │   └── openai.go                   │   ├── indexer       │
│  └── go.mod                          ├── go.mod (with      │
│                                       │   replacement)      │
│  Lightweight API proxy                Data APIs, indexing   │
│  Stable, minimal changes              Evolving features    │
│                                                              │
├──────────────────────────────────────────────────────────────│
│                                                              │
│  dashboard/                           cc-viz/ (PLANNED)    │
│  ├── src/pages/                      ├── src/pages/        │
│  │   ├── Dashboard.tsx                │   └── Conversations │
│  │   ├── Requests.tsx                ├── src/components/   │
│  │   ├── Token Usage.tsx             ├── src/lib/          │
│  │   └── Conversations.tsx (EXTRACT) └── package.json      │
│  ├── src/lib/api.ts                                         │
│  └── package.json                    Conversation browser  │
│                                       New functionality     │
│  Main proxy dashboard                                       │
│  All monitoring features                                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
        ↑                        ↑         ↑
        Caddy reverse proxy (port 8000)
        Routes /v1/* → proxy-core (8001)
        Routes /api/* → proxy-data (8002)
        Routes / → dashboard (8173)
                     cc-viz (8174)
```

---

## Target Structure After Phase 2 (cc-viz extraction)

```
project-root/
├── proxy-core/              # Phase 1 deliverable
│   ├── go.mod
│   └── cmd/proxy-core/
├── proxy/                   # Remaining proxy code (proxy-data)
│   ├── go.mod
│   └── cmd/proxy-data/
├── dashboard/               # Updated: Conversations tab removed
│   ├── src/pages/
│   │   ├── Dashboard.tsx
│   │   ├── Requests.tsx
│   │   ├── TokenUsage.tsx
│   │   ├── Performance.tsx
│   │   ├── Routing.tsx
│   │   ├── Settings.tsx
│   │   └── ToolsDemo.tsx
│   └── src/lib/api.ts
├── cc-viz/                  # Phase 2 deliverable - NEW
│   ├── package.json
│   ├── src/pages/
│   │   └── Conversations.tsx    # Moved from dashboard/
│   ├── src/components/
│   │   ├── GlobalSearchResults.tsx
│   │   ├── ConversationSearch.tsx
│   │   ├── ConversationThread.tsx
│   │   └── ConversationList.tsx
│   ├── src/lib/
│   │   ├── api.ts               # Conversation API calls
│   │   ├── types.ts             # Conversation types
│   │   ├── search.ts            # Search utilities
│   │   └── formatters.ts        # Display formatting
│   └── src/main.tsx
├── Justfile                 # Updated build targets
├── docker-compose.yml       # Updated with cc-viz service
└── docker/                  # New Dockerfile for cc-viz
    └── Dockerfile.cc-viz
```

---

## Acceptance Criteria

How we'll know Phase 2 (cc-viz extraction) is complete:

- [ ] `cc-viz/` directory exists with proper React/Vite structure
- [ ] All Conversations-related components moved to cc-viz
- [ ] cc-viz can build independently: `cd cc-viz && npm run build`
- [ ] cc-viz app runs locally: `cd cc-viz && npm run dev` on port 8174
- [ ] cc-viz calls proxy-data API correctly (all endpoints work)
- [ ] Dashboard builds and runs without Conversations tab
- [ ] Sidebar in dashboard no longer has Conversations link
- [ ] `just docker` starts all three services (proxy-core, proxy-data, cc-viz)
- [ ] Docker-compose file properly configured
- [ ] All existing tests pass
- [ ] Curl tests pass for all proxy-data conversation endpoints
- [ ] E2E: Can browse conversations through cc-viz webapp
- [ ] Zero-downtime: Can deploy cc-viz independently without affecting proxy

---

## Implementation Approach

### Phase Sequence

1. **Phase 1: Proxy-Core Extraction** (separate sprint, ~5 days)
   - Prerequisite for monorepo pattern
   - Establishes go.mod replacement approach
   - Makes proxy-core stable and sealed off
   - Plan: `.agent_planning/monorepo-restructure/PLAN-20260112.md`

2. **Phase 2: cc-viz Extraction** (THIS SPRINT, ~3-4 days after Phase 1)
   - Create cc-viz project structure
   - Extract Conversations components from dashboard
   - Configure API calls to proxy-data
   - Update dashboard to remove Conversations
   - Update Justfile and docker-compose
   - Verification and E2E testing

3. **Phase 3 & Beyond: Expand cc-viz** (future)
   - Unified search with filters
   - Analytics/insights
   - Export capabilities
   - Real-time conversation updates

### Recommended Steps for Phase 2

1. **Create cc-viz structure** (30 min)
   - `mkdir -p cc-viz/src/{pages,components,lib}`
   - Create `cc-viz/package.json` based on dashboard
   - Create `cc-viz/vite.config.ts`
   - Create `cc-viz/tsconfig.json`

2. **Extract components** (1 hour)
   - Copy Conversations page and supporting components
   - Copy conversation-related types and utilities
   - Update import paths to be relative

3. **Configure API calls** (30 min)
   - Create `cc-viz/src/lib/api.ts` with conversation endpoints
   - Point to proxy-data API (configurable base URL)
   - Handle loading states and errors

4. **Update dashboard** (30 min)
   - Remove Conversations page
   - Remove Conversations link from sidebar
   - Remove conversation-related imports and types from main app
   - Keep other pages and functionality intact

5. **Update build system** (1 hour)
   - Add cc-viz build targets to Justfile
   - Create `docker/Dockerfile.cc-viz`
   - Update docker-compose.yml with cc-viz service
   - Update `docker-compose.yml` port mappings

6. **Testing & Verification** (1-2 hours)
   - `just build` works for both projects
   - `just test` passes in both
   - `just docker` starts all services
   - `curl localhost:8000/api/v2/conversations` returns data
   - cc-viz app loads and displays conversations
   - E2E: Can click conversation and see details

### Patterns to Follow

- **API calls**: Use React Query (already in dashboard) for state management
- **Types**: All conversation types in `cc-viz/src/lib/types.ts`
- **Styling**: Keep using current Tailwind/CSS setup from dashboard
- **Error handling**: Match existing error handling patterns from dashboard
- **Loading states**: Use existing spinner/loading components

### Known Gotchas

- **API base URL**: cc-viz needs to know where proxy-data is (env var or config)
- **CORS**: If cc-viz runs on different port, proxy-data must allow CORS
- **Auth**: If proxy uses auth, cc-viz must handle auth tokens the same way
- **Shared components**: Some UI components may be used by both dashboard and cc-viz
  - Decision: Copy for now (can refactor to shared package later)
- **Types duplication**: conversation types will be in cc-viz only (was in shared dashboard before)

---

## Reference Materials

### Planning Documents

- **Monorepo Restructure Plan**: `.agent_planning/monorepo-restructure/PLAN-20260112.md`
  - Full details on proxy-core extraction (Phase 1)
  - Monorepo structure and rationale
  - Justfile and Docker changes needed

- **Current Evaluation**: `.agent_planning/monorepo-restructure/EVALUATION-20260112.md`
  - Current architecture analysis
  - Target structure overview

### Codebase References

**Files to Extract**:
- `dashboard/src/pages/Conversations.tsx` - Main page
- `dashboard/src/components/features/ConversationThread.tsx` - Message display
- `dashboard/src/components/features/ConversationSearch.tsx` - Search input
- `dashboard/src/components/features/ConversationList.tsx` - List of conversations
- `dashboard/src/lib/search.ts` - Search utilities
- `dashboard/src/lib/types.ts` - Conversation types

**API to Consume**:
- `proxy/internal/handler/data_handler.go:GetConversationsV2()` (line 658)
- `proxy/internal/handler/data_handler.go:GetConversationMessagesV2()` (line 726)
- `proxy/internal/handler/data_handler.go:SearchConversations()` (line 536)

**Config/Build References**:
- `Justfile` - Build orchestration
- `docker-compose.yml` - Service orchestration
- `dashboard/vite.config.ts` - Vite config (copy as template)
- `dashboard/package.json` - Dependencies (copy and adjust)

### Related Issues

- None yet - create beads issue after planning phase

---

## Questions & Blockers

### Open Questions

- [ ] Should cc-viz be a separate npm package/repository later, or stay in monorepo?
- [ ] What's the deployment strategy for cc-viz? (Same docker-compose, separate container registry, etc.)
- [ ] Should cc-viz have its own authentication/session, or inherit from proxy?
- [ ] What's the domain/port for cc-viz in production?

### Current Blockers

None identified. Phase 1 (proxy-core extraction) must complete first, but that's planned, not blocked.

### Need User Input On

- Should cc-viz be a long-term separate project or eventually merge back? (Affects naming, structure decisions)

---

## Testing Strategy

### Existing Tests

- `dashboard/` has test setup with Vitest
- Can copy test infrastructure to cc-viz

### New Tests Needed

- [ ] Test conversation list loads from API
- [ ] Test message pagination works
- [ ] Test search filters work correctly
- [ ] Test error states display properly

### Manual Testing

1. Start all services: `just docker`
2. Navigate to cc-viz: `http://localhost:8174`
3. Verify conversations load: See list of conversations
4. Click conversation: See messages displayed
5. Search: Type in search box, see filtered results
6. Verify API calls: Check DevTools network tab shows `/api/v2/conversations` calls

---

## Success Metrics

How to validate Phase 2 completion:

1. ✅ `just build` succeeds for all projects
2. ✅ `just test` passes (if tests exist)
3. ✅ `just docker` starts all services without errors
4. ✅ cc-viz accessible at `http://localhost:8174` (dev) or configured URL (prod)
5. ✅ Can list conversations and view details
6. ✅ All conversation API endpoints working (curl tests pass)
7. ✅ Dashboard still works without Conversations tab
8. ✅ Services can be deployed/redeployed independently
9. ✅ Zero errors in browser console
10. ✅ All component state management working correctly

---

## Next Steps for Agent

### Immediate Actions (After Phase 1 Complete)

1. Review this handoff document
2. Create `.agent_planning/cc-viz/PLAN-<date>.md` for Phase 2 implementation
3. Start cc-viz extraction:
   - Create project structure
   - Move components
   - Update build system
4. Verify with tests and docker-compose

### Before Starting Implementation

- [ ] Phase 1 (proxy-core extraction) is complete and verified
- [ ] All proxy-data conversation endpoints working (tested via curl)
- [ ] Understand current dashboard structure
- [ ] Review Justfile and docker-compose patterns

### When Complete

- [ ] Update this handoff document with completion notes
- [ ] Create beads issue to track future cc-viz expansion
- [ ] Update `.agent_planning/ROADMAP.md` with completed work
- [ ] Update `CLAUDE.md` with new project structure

---

## Architecture Decisions Documented

**Why extract Conversations separately?**
- Clean separation of concerns: proxy focuses on request monitoring, cc-viz focuses on conversation exploration
- Independent evolution: cc-viz can add visualization features without affecting proxy stability
- Deployment flexibility: Can deploy cc-viz updates without touching stable proxy-core

**Why keep proxy monorepo vs. separate git repo?**
- Shared API (proxy-data) that both consume
- Single docker-compose for full stack
- Easier initial development and deployment
- Can be split to separate repos later if needed

**Why consume via API vs. direct DB?**
- Loose coupling: cc-viz doesn't depend on database schema
- Security: API provides access control point
- Flexibility: proxy-data can change storage layer without affecting cc-viz
- Scalability: API can be cached, rate-limited, etc.

---

## Success looks like:

```
$ just docker
[proxy-core]  ✓ Started on :8001
[proxy-data]  ✓ Started on :8002
[cc-viz]      ✓ Started on :8174
[caddy]       ✓ Reverse proxy ready

$ curl http://localhost:8174
[HTML response from cc-viz app]

$ curl http://localhost:8000/api/v2/conversations
[JSON array of conversations]

✅ All services healthy and working together
```

---
