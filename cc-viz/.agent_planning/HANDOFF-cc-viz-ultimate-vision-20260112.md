# Handoff: CC-VIZ Ultimate Vision - The Claude Code Data Cockpit

**Created**: 2026-01-12
**For**: Any agent implementing CC-VIZ features
**Status**: Vision complete, ready for planning/implementation

---

## Objective

Transform CC-VIZ from a collection of separate visualizers into a unified **data cockpit** where every piece of ~/.claude data is visible, connected, and navigable - enabling exploration of the complete web of interconnected Claude Code activity.

## Current State

### What's Been Done
- **Conversations page**: Browse session transcripts
- **Configuration page**: View settings.json, CLAUDE.md, MCP servers
- **Projects page**: Activity per project with session list, storage breakdown
- **Session Data page**: Todos and plans visualization (just built)
  - Database-backed ingestion from ~/.claude/todos/ and ~/.claude/plans/
  - Color-coded status indicators
  - Proper Session ID / Agent ID display (verified via investigation)

### What's Missing (The Vision)
The existing pages are **islands** - they don't connect to each other. The data IS interconnected:
- Sessions have UUIDs that appear in todos, projects, history
- Subagents are nested within sessions
- File changes correlate with conversation timestamps
- Plans are created during sessions

But users can't navigate these connections. That's the gap.

## The Ultimate Feature Set

### 1. The Graph Navigator - Universal Data Interconnection
**Make every connection visible and clickable.**

From any entity (conversation, file, plan, todo, subagent), show:
- All other entities it connects to
- Bidirectional navigation (file → conversations that modified it)
- Graph queries: "Show conversations that modified /src/auth/ this week"

**Implementation**: Build a relationship graph during indexing. Each entity stores references to related entities. UI shows collapsible relationship panels.

### 2. The Unified Search Engine
**One search across everything.**

```
Search: "postgres migration"

Results:
- Conversations (3): "Implementing database migration system"...
- Plans (1): peppy-yawning-teapot.md
- History (8): "run the postgres migration"
- File Changes (12): migrations/001_initial.sql
- Todos (2): "Implement postgres migration runner"
```

Advanced syntax: `type:conversation project:cc-viz modified:>2026-01-01`

**Implementation**: FTS5 across all indexed content. Results link back to source entities with context highlighted.

### 3. The Activity Intelligence Dashboard
**Turn stats into understanding.**

- Live status: Active sessions, recent conversations, unfinished todos
- Patterns: Most active hours, productivity trends, tool usage breakdown
- Comparisons: This week vs last week, per-project breakdown
- Click any metric to drill into underlying data

**Implementation**: Parse stats-cache.json, aggregate from indexed sessions. Cache computations for fast loading.

### 4. The Extension Workshop
**All agents, commands, skills, plugins in one place.**

- Browse and manage all extension types
- Usage analytics (from session data correlation)
- Health indicators: unused, outdated, disabled
- Quick actions: enable/disable, edit, create new

**Implementation**: Index ~/.claude/agents/, commands/, skills/, plugins/. Correlate usage from conversation transcripts.

### 5. The Time Machine - Session Replay
**See exactly what happened, step by step.**

- Timeline scrubber for any conversation
- Correlated file diffs at each timestamp
- Tool calls, subagent spawns, file changes - all synchronized
- "Restore" any file version

**Implementation**: Correlate conversation timestamps with file-history/ data. Build timeline data structure linking events chronologically.

### 6. The Subagent Observatory
**Visualize agent hierarchies and flow.**

```
Main Session
  └── code-reviewer (agent-abc123) ✓
        └── analysis-helper (agent-def456) ✓
  └── test-writer (agent-ghi789) ⟳
```

- Tree visualization of spawn hierarchies
- Click any node to view its conversation
- Metrics: tokens, duration, files modified per agent

**Implementation**: Parse subagent directories, build parent-child relationships from sessionId/agentId. Aggregate metrics.

### 7. The Space Manager
**Understand and clean up ~/.claude.**

```
~/.claude Storage: 1.2 GB

[████████████████████████████] projects/     680 MB (57%)
[██████████████]               telemetry/    203 MB (17%) [Clean up?]
[██████████]                   debug/        145 MB (12%) [Compress?]
```

- Breakdown by directory
- Identification of safe-to-delete items
- Bulk actions: archive, compress, clean
- Preview before any destructive action

**Implementation**: Directory scanning with heuristics for stale data. Archive/compress utilities with safety checks.

### 8. The Cross-Session Todo Board
**Kanban aggregation of all todos.**

```
Pending (23)       In Progress (5)    Completed (147)
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│ Implement X  │   │ Add tests    │   │ Fix auth bug │
│ Session: abc │   │ Session: def │   │ Session: ghi │
└──────────────┘   └──────────────┘   └──────────────┘
```

- All todos across all sessions in one view
- Click to navigate to source session
- Filters by session, project, age, keywords

**Implementation**: Already have todos indexed. Add aggregated view with session correlation.

## Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| Database-first | Never load JSON from filesystem for web requests |
| Indexer pattern | Background ingestion with incremental updates |
| UUID-based linking | Use actual UUIDs as foreign keys between entities |
| Session ID / Agent ID | Filename pattern `{session}-agent-{agent}.json` is authoritative |

## Files & Data Sources

### Currently Indexed
- `~/.claude/todos/` → `claude_todos`, `claude_todo_sessions` tables
- `~/.claude/plans/` → `claude_plans` table
- `~/.claude/projects/` → `claude_projects` (directory scan, not full content)
- `~/.claude/settings.json`, `CLAUDE.md`, `.mcp.json` → read on demand

### To Be Indexed
- `~/.claude/projects/{path}/{session}.jsonl` → conversation content, tool calls
- `~/.claude/projects/{path}/{session}/subagents/` → subagent transcripts
- `~/.claude/history.jsonl` → prompt history
- `~/.claude/file-history/` → file versions
- `~/.claude/agents/`, `commands/`, `skills/`, `plugins/` → extensions
- `~/.claude/stats-cache.json` → activity statistics

### Relationships to Build
- Session → Conversations, Todos, Subagents, File Changes
- Project → Sessions
- Subagent → Parent Session, Todos
- File Change → Session(s) that modified it
- Plan → Session that created/referenced it

## Implementation Approach

### Phase 1: Graph Foundation
1. Add conversation indexing (parse JSONL, store messages + tool calls)
2. Add subagent indexing with parent references
3. Build relationship graph structure
4. Add cross-linking in existing UIs (click Session ID → go to conversation)

### Phase 2: Search & Discovery
1. FTS5 index across all content
2. Unified search UI
3. Advanced query syntax
4. Result grouping and context display

### Phase 3: Intelligence & Visualization
1. Activity dashboard from stats-cache.json
2. Subagent tree visualization
3. Time machine / session replay
4. Space manager with cleanup actions

### Phase 4: Extension Management
1. Extension workshop UI
2. Usage analytics correlation
3. Management actions

## Reference Materials

### Planning Documents
- [FEATURE_PROPOSAL_cc-viz-ultimate-vision.md](.agent_planning/FEATURE_PROPOSAL_cc-viz-ultimate-vision.md) - Full vision document
- [PLAN-sprint2-20260112.md](.agent_planning/cc-viz-roadmap/PLAN-sprint2-20260112.md) - Session Data implementation plan

### Codebase References
- `proxy/internal/service/storage_sqlite.go` - Database schema pattern
- `proxy/internal/service/session_data_indexer.go` - Indexer pattern
- `proxy/internal/handler/data_handler.go` - API handler pattern
- `cc-viz/src/pages/SessionData.tsx` - Latest page implementation

### Data Format References
- Todo filename: `{session_uuid}-agent-{agent_uuid}.json`
- Conversation: `.jsonl` with `sessionId`, `agentId`, `parentUuid`, `type`, `message`
- Subagent: `subagents/agent-{short_id}.jsonl`

## Open Questions

- [ ] Data freshness: Auto-refresh or manual?
- [ ] Search index location: In ~/.claude or separate?
- [ ] Performance at scale: How to handle 1000+ sessions?
- [ ] Privacy mode: Should we support content redaction?

## Success Metrics

- Any UUID in the UI is clickable and navigates to context
- Search finds content regardless of source type
- Users can answer "what did I work on last week" in <30 seconds
- Storage cleanup recovers space without data loss risk

---

## Next Steps for Agent

**Immediate actions**:
1. Read FEATURE_PROPOSAL for full vision context
2. Decide which feature to implement next
3. Plan database schema additions for new entity types

**Before starting implementation**:
- [ ] Check existing indexer patterns in session_data_indexer.go
- [ ] Understand JSONL conversation format
- [ ] Verify relationship linkage strategy

**When complete**:
- [ ] Update this handoff with implementation status
- [ ] Add new entity types to relationship graph
