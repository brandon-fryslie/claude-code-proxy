# Handoff: CC-VIZ - Claude Directory Visualization

**Created**: 2026-01-12T22:30:00Z
**For**: Agent continuing CC-VIZ feature development and Claude directory integration
**Status**: Ready to Start - Foundation Complete, Feature Expansion Phase

---

## Objective

Extend CC-VIZ beyond current conversation browsing to visualize and explore the entire `~/.claude` directory structure, including agents, projects, configuration, and project metadata. Create a comprehensive visualization tool for Claude Code users to understand and manage their Claude workspace.

---

## Current State

### What's Been Done

**Phase 2: CC-VIZ Extraction (Complete)**
- ✅ Extracted Conversations feature from dashboard into standalone cc-viz project
- ✅ Created cc-viz as independent React+Vite+TypeScript application
- ✅ Set up proper sidebar/layout matching dashboard design
- ✅ Integrated with proxy-data API for conversation data
- ✅ Added cc-viz link to dashboard sidebar (in Overview section)
- ✅ Implemented reverse proxy routing (no hardcoded localhost)
- ✅ Added cc-viz to run-split.sh and docker-compose.yml
- ✅ All builds succeed - dashboard and cc-viz compile cleanly
- ✅ Zero localhost hardcoding - fully scalable

**Foundation Files in Place**
- ✅ `cc-viz/` project structure complete
- ✅ Sidebar with navigation and collapsible menu
- ✅ AppLayout component matching dashboard
- ✅ Conversations page with search, sort, and detail view
- ✅ API layer configured for proxy-data consumption
- ✅ CSS/styling matching dashboard theme

### What's In Progress

- Currently: Conversation browsing only
- Integration with reverse proxy working correctly
- Accessible via `/cc-viz/` path through Caddy

### What Remains

**Phase 3: Claude Directory Visualization (Ready to Start)**

Core Features:
- [ ] Directory browser showing `~/.claude` structure
- [ ] Agents explorer (view `.claude/agents/` YAML definitions)
- [ ] Projects navigator (view `.claude/projects/` structure)
- [ ] Configuration viewer (view `.claude` files and settings)
- [ ] Workspace metadata (profile, settings, bindings)

Advanced Features:
- [ ] Cross-reference explorer (agents → projects they work on)
- [ ] Search/filter across entire directory
- [ ] Real-time directory watching and refresh
- [ ] Conversation history linked to agents/projects
- [ ] Analytics on agent usage, project activity

Documentation & Testing:
- [ ] Unit tests for directory reading
- [ ] E2E tests for directory navigation
- [ ] Documentation on directory structure
- [ ] User guide for CC-VIZ features

---

## Context & Background

### Why We're Doing This

**Problem**: Claude Code users have a complex `~/.claude` directory structure (agents, projects, conversations, config) but no way to visualize or understand it. The Conversations feature is only a small part - users need to see the whole picture.

**Solution**: CC-VIZ becomes a comprehensive workspace visualization tool that lets users:
- Understand their agent ecosystem
- See project organization
- Browse configuration and settings
- View conversation history in context
- Identify unused agents or orphaned conversations
- Export/analyze workspace usage patterns

**Why Now**: The conversation browsing foundation is solid. The reverse proxy is properly configured. The next logical step is expanding the visualization scope to the entire Claude workspace.

### Key Decisions Made

| Decision | Rationale | Date |
|----------|-----------|------|
| Extracted as standalone project | Allows independent evolution; clear separation from dashboard | 2026-01-12 |
| Use reverse proxy routing (/cc-viz/) | Works on any hostname; no env vars; fully scalable | 2026-01-12 |
| Sidebar + header layout | Consistent UX with dashboard; room for navigation expansion | 2026-01-12 |
| Consume proxy-data API | Loose coupling; lets API handle auth/caching | 2026-01-12 |
| Start with conversations | MVP feature; validates architecture before directory features | 2026-01-12 |

### Important Constraints

- **Filesystem Access**: CC-VIZ runs in browser; cannot directly read filesystem. Must use backend API.
- **Architecture**: proxy-data backend must handle directory reading and serving (requires new endpoints)
- **Workspace Privacy**: Users may have sensitive data in agents/projects - respect existing proxy auth
- **Performance**: Directory can contain 100s of agents/projects - must be efficient
- **Real-time**: Directory changes on disk (user edits agents via CLI) should reflect in UI
- **Backward Compatibility**: Conversation browsing must continue working exactly as is

---

## Scope

### Phase 3a: Directory API Backend

**New proxy-data endpoints needed**:

```
GET /api/v2/claude/directory
  Returns: { agents: [], projects: [], config: {} }

GET /api/v2/claude/agents
  Returns: List of agent definitions from .claude/agents/

GET /api/v2/claude/agents/{id}
  Returns: Full agent definition (YAML parsed to JSON)

GET /api/v2/claude/projects
  Returns: List of projects from .claude/projects/

GET /api/v2/claude/projects/{id}/metadata
  Returns: Project metadata and activity

GET /api/v2/claude/config
  Returns: ~/.claude configuration files

POST /api/v2/claude/refresh
  Trigger directory rescan/refresh
```

**Implementation**:
- Add directory reading service to proxy/internal/service/
- Parse YAML agent definitions
- Cache directory structure with TTL
- Watch for directory changes (optional: inotify on Linux)

### Phase 3b: CC-VIZ Frontend Pages

**New cc-viz routes**:
```
/cc-viz/               # Conversations (existing)
/cc-viz/agents         # Agent explorer
/cc-viz/projects       # Project navigator
/cc-viz/config         # Configuration viewer
/cc-viz/search         # Unified search across directory
/cc-viz/usage          # Workspace analytics
```

**Components to build**:
- AgentBrowser - List and detail view of agents
- ProjectNavigator - Tree view of projects
- ConfigViewer - Display config files
- DirectorySearch - Search/filter all items
- WorkspaceAnalytics - Usage stats and visualizations

### Phase 3c: Integration & Polish

- Link conversations to agents/projects they belong to
- Cross-referencing between views
- Export/analyze functionality
- Performance optimization for large workspaces

### Files to Create/Modify

**Backend** (`proxy/internal/`):
- `service/directory.go` - New directory reading service
- `handler/data_handler.go` - Add directory endpoints

**Frontend** (`cc-viz/src/`):
- `lib/api.ts` - Add directory API hooks
- `lib/types.ts` - Add directory types
- `pages/Agents.tsx` - New agent browser page
- `pages/Projects.tsx` - New project navigator page
- `pages/Config.tsx` - New config viewer page
- `components/features/AgentList.tsx`
- `components/features/ProjectTree.tsx`
- `components/features/ConfigDisplay.tsx`

### Out of Scope (This Phase)

- Editing agents/projects/config through UI (read-only first)
- Real-time file watching with websockets (polling/manual refresh)
- Advanced analytics/ML on usage patterns
- Integration with GitHub/remote repos
- Mobile UI (focus on desktop first)

---

## Implementation Approach

### Recommended Sequence

1. **Week 1: Backend Directory API**
   - Create `directory.go` service for reading `~/.claude/`
   - Implement YAML parsing for agent definitions
   - Add endpoints to data_handler.go
   - Test with curl

2. **Week 2: Frontend Agent Browser**
   - Create AgentBrowser page and components
   - Connect to new API endpoints
   - Implement agent detail view
   - Search/filter agents

3. **Week 3: Frontend Project Navigator**
   - Create ProjectNavigator with tree view
   - Implement project detail view
   - Show project structure and metadata

4. **Week 4: Config Viewer & Integration**
   - Create ConfigViewer for `~/.claude` files
   - Link conversations to agents
   - Cross-referencing between views
   - Polish and optimize

### Patterns to Follow

- **API Calls**: Use React Query hooks (already established pattern in cc-viz)
- **Types**: All directory types in `cc-viz/src/lib/types.ts`
- **Styling**: Continue with Tailwind + CSS variables matching dashboard
- **Error Handling**: Match existing error handling patterns from Conversations
- **Loading States**: Use existing spinner/loading patterns
- **Backend**: Follow handler pattern from `proxy/internal/handler/data_handler.go`

### Known Gotchas

- **Path Handling**: `~/.claude` expands differently in Docker vs host - use absolute paths in Go
- **YAML Parsing**: Go YAML parser may have different rules than Python - test thoroughly
- **Large Directories**: Agent lists can be 100s+ items - implement pagination/virtualization early
- **Permission Issues**: May not have read access to some `.claude` subdirs - handle gracefully
- **Circular References**: Agents can reference other agents - avoid infinite loops in visualization
- **File Watching**: Don't implement inotify yet - stick with polling for MVP

---

## Reference Materials

### Planning Documents

- [HANDOFF-cc-viz-separate-project-20260112.md](./HANDOFF-cc-viz-separate-project-20260112.md) - Phase 2 completion and architecture
- [Monorepo Restructure PLAN](./monorepo-restructure/PLAN-20260112.md) - Build system and docker setup

### Codebase References

**Backend Examples**:
- `proxy/internal/handler/data_handler.go:658` - GetConversationsV2 endpoint pattern
- `proxy/internal/service/storage_sqlite.go` - Data storage pattern
- `proxy/internal/service/indexer.go` - File processing pattern

**Frontend Examples**:
- `cc-viz/src/pages/Conversations.tsx` - Page structure and hooks pattern
- `cc-viz/src/components/layout/AppLayout.tsx` - Layout and sidebar pattern
- `dashboard/src/lib/api.ts` - API query hook pattern
- `dashboard/src/lib/search.ts` - Filter/search utilities

**Project Structure**:
- `~/.claude/agents/` - YAML agent definitions
- `~/.claude/projects/` - Project metadata and logs
- `~/.claude/` - Config files, profile, settings

### API Contracts Already Working

- `GET /api/v2/conversations` - List conversations
- `GET /api/v2/conversations/{id}` - Get conversation
- `GET /api/v2/conversations/{id}/messages` - Paginated messages
- **These become reference for building new directory API**

---

## Acceptance Criteria

How to know Phase 3 is complete:

Frontend:
- [ ] Agent browser page loads list of agents from `~/.claude/agents/`
- [ ] Can click agent to see full definition
- [ ] Agent list searchable/filterable
- [ ] Projects page shows project directory structure
- [ ] Config viewer displays `~/.claude` configuration files
- [ ] All pages match dashboard styling
- [ ] Navigation works between all cc-viz pages
- [ ] No hardcoded URLs or localhost references

Backend:
- [ ] `/api/v2/claude/agents` endpoint returns agent list
- [ ] `/api/v2/claude/agents/{id}` returns parsed agent definition
- [ ] `/api/v2/claude/projects` returns project structure
- [ ] `/api/v2/claude/config` returns config files
- [ ] All endpoints handle missing directories gracefully
- [ ] YAML parsing works correctly
- [ ] Paths expand properly in Docker and local dev

Testing:
- [ ] All existing conversation tests still pass
- [ ] New directory endpoints tested
- [ ] E2E test for agent browser workflow
- [ ] Performance acceptable for 100+ agents

---

## Questions & Blockers

### Open Questions

- [ ] Should we show conversation history within agent/project context, or keep it separate?
- [ ] How deep should project tree view go? (Just top level or all subdirs?)
- [ ] Should config viewer be read-only, or eventually support editing?
- [ ] How to handle symlinks in project directories?
- [ ] What's the expected size of largest workspace? (for pagination decisions)

### Current Blockers

None identified. All prerequisites complete:
- ✅ CC-VIZ project structure ready
- ✅ Reverse proxy routing working
- ✅ API infrastructure established
- ✅ Backend can add new endpoints
- ✅ Frontend can add new pages

### Need User Input On

- Expected timeline for Phase 3 vs other priorities
- Which features are must-have vs nice-to-have
- Preference for read-only or eventually editable config

---

## Testing Strategy

### Existing Tests

- CC-VIZ build: ✅ Vite tests pass
- Conversation browsing: ✅ Component tests exist
- API integration: ✅ Patterns established

### New Tests Needed

Backend:
- [ ] Unit tests for directory reading (various scenarios)
- [ ] YAML parsing tests (valid/invalid/edge cases)
- [ ] Path expansion tests (docker vs host)
- [ ] Permission error handling tests
- [ ] Endpoint integration tests

Frontend:
- [ ] Agent list component tests
- [ ] Project navigator tests
- [ ] Search/filter tests
- [ ] API hook tests
- [ ] E2E test: Browse agents → view detail → link to conversation

### Manual Testing

1. Start all services: `just docker`
2. Navigate to `http://localhost:8000/cc-viz/`
3. Verify Conversations page works
4. Test new Agent browser page
5. Test Project navigator
6. Test Config viewer
7. Verify links between views work
8. Test with 50+ agents in `~/.claude/agents/`

---

## Success Metrics

How to validate Phase 3 implementation:

- All acceptance criteria above pass
- Dashboard + CC-VIZ + Proxy all build cleanly
- Zero hardcoded URLs
- Works on any hostname without config changes
- Handles 100+ agents efficiently (< 500ms load)
- All existing conversation features still work
- Code follows established patterns from Phase 2
- No new dependencies added

---

## Next Steps for Agent

**Immediate Actions**:

1. Review Phase 2 completion (HANDOFF-cc-viz-separate-project-20260112.md)
2. Understand current directory structure (`~/.claude/agents/`, etc.)
3. Plan backend directory service in Go
4. Design directory API contracts
5. Create frontend page structure for agent browser

**Before Starting Implementation**:
- [ ] Read all reference materials above
- [ ] Understand reverse proxy routing (no localhost!)
- [ ] Verify local dev setup works (`just run` or `just docker`)
- [ ] Analyze `~/.claude` directory structure in your system
- [ ] Review existing API patterns in proxy-data

**When Complete**:
- [ ] All builds pass: `just build`
- [ ] Tests pass: `just test`
- [ ] Docker works: `just docker`
- [ ] No localhost hardcoding
- [ ] Update this handoff with completion status
- [ ] Create new HANDOFF for Phase 4 (if applicable)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│ User Browser                                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  http://hostname:8000/cc-viz/                         │
│  ↓                                                     │
│  Caddy (port 8000) - Reverse Proxy                    │
│  ├── /cc-viz/* → cc-viz:8174 (port 8174)            │
│  ├── /api/*    → proxy-data:8002                     │
│  └── /v1/*     → proxy-core:8001                     │
│                                                       │
└─────────────────────────────────────────────────────────┘
         ↓                  ↓                    ↓
    ┌─────────┐      ┌──────────────┐    ┌────────────┐
    │ CC-VIZ  │      │ Proxy-Data   │    │ Proxy-Core │
    │ React   │      │ Go Backend   │    │ Go Backend │
    │ Port    │      │ Port 8002    │    │ Port 8001  │
    │ 8174    │      │              │    │            │
    │         │      │ Reads:       │    │ Forwards:  │
    │ Features│      │ - sqlite.db  │    │ - API      │
    │ -Agents │      │ - ~/.claude/ │    │   calls    │
    │ -Config │      │ - File FS    │    │ - Routes   │
    │ -Project│      │              │    │   to LLMs  │
    │ -Conver │      │ Serves:      │    │            │
    │         │      │ - /api/v2/*  │    │ Serves:    │
    │         │      │ - /claude/*  │    │ - /v1/*    │
    └─────────┘      └──────────────┘    └────────────┘
```

## Key Files

**Backend** (Go):
- `proxy/internal/handler/data_handler.go` - HTTP handlers (add directory endpoints here)
- `proxy/internal/service/directory.go` - NEW, directory reading service
- `proxy/go.mod` - Dependencies

**Frontend** (React):
- `cc-viz/src/pages/Conversations.tsx` - Existing (conversions reference)
- `cc-viz/src/pages/Agents.tsx` - NEW
- `cc-viz/src/pages/Projects.tsx` - NEW
- `cc-viz/src/pages/Config.tsx` - NEW
- `cc-viz/src/lib/api.ts` - Add directory hooks
- `cc-viz/src/lib/types.ts` - Add directory types
- `cc-viz/src/components/layout/AppLayout.tsx` - Navigation hub

**Routing**:
- `docker/Caddyfile` - Already routes `/cc-viz/*` to cc-viz:8174
- `docker-compose.yml` - Services configured
- `Justfile` - Build targets

---

## Communication & Handoff

**This document is ready for agent pickup**. An agent can:

1. Read this handoff completely
2. Start with Phase 3a (backend directory API)
3. Reference acceptance criteria above
4. Use testing strategy for validation
5. Update handoff when complete

**For questions or blockers**: Update the "Open Questions" section above and communicate with project lead.

---

## Success Definition

When Phase 3 is complete:
- User navigates to `http://mydomain.com/cc-viz/agents` → sees all agents
- User clicks agent → sees full definition
- User navigates to `/cc-viz/projects` → sees project tree
- User navigates to `/cc-viz/config` → sees configuration
- All works seamlessly on any hostname (no localhost)
- Performance is snappy even with 100+ agents
- Code is clean, tested, and maintainable

---
