# CC-VIZ Roadmap: The Claude Code Data Cockpit

**Last updated**: 2026-01-12
**Vision**: Transform CC-VIZ into a unified data cockpit where every piece of ~/.claude data is visible, connected, and navigable.

---

## Phase 1: Graph Foundation 🟢 [ACTIVE]

**Goal**: Build the underlying infrastructure for interconnected data visualization. Establish database schema, indexing patterns, and relationship linking.

**Status**: 1/4 completed | Session Data page launched

### Topics

#### ✅ session-data-indexer [COMPLETED]
- **Epic**: Session Data visualization page
- **Status**: Completed
- **What's Done**:
  - Database-backed ingestion from ~/.claude/todos/ and ~/.claude/plans/
  - Color-coded status indicators
  - Session ID / Agent ID display with verification
- **Reference**: `.agent_planning/cc-viz-roadmap/`

#### 🔄 conversation-indexer [IN_PROGRESS]
- **Epic**: CONVERSATION-INDEXER-1
- **Status**: Planning
- **Description**: Parse JSONL conversations, extract messages, tool calls, timestamps
- **Includes**:
  - Parse ~/.claude/projects/{path}/{session}.jsonl
  - Extract and index tool calls (Read, Write, Bash, etc.)
  - Build timestamp correlations
  - Store in SQLite with FTS5 for search

#### 💡 subagent-graph [PROPOSED]
- **Epic**: SUBAGENT-GRAPH-1
- **Status**: Proposed
- **Description**: Index subagent hierarchies with parent-child relationships
- **Includes**:
  - Parse ~/.claude/projects/{path}/{session}/subagents/
  - Build parent spawn references
  - Correlate with session UUIDs
  - Graph visualization structure

#### 💡 relationship-linker [PROPOSED]
- **Epic**: RELATIONSHIP-LINKER-1
- **Status**: Proposed
- **Description**: Build cross-entity relationship graph
- **Dependencies**: conversation-indexer, subagent-graph
- **Includes**:
  - Session → Conversations, Todos, Subagents, File Changes
  - File Change → Sessions that modified it
  - Plan → Sessions that created/referenced it
  - Bidirectional navigation support

---

## Phase 2: Search & Discovery 🔄 [QUEUED]

**Goal**: Enable users to find anything across all data sources with unified search.

**Status**: 0/3 completed

### Topics

#### 💡 unified-search [PROPOSED]
- **Epic**: UNIFIED-SEARCH-1
- **Status**: Proposed
- **Dependencies**: conversation-indexer, relationship-linker
- **Description**: FTS5 search across all content types
- **Includes**:
  - Index conversations, plans, todos, history, file changes
  - Multi-type result grouping
  - Result ranking by relevance and recency
  - Context snippets with highlights

#### 💡 advanced-query-syntax [PROPOSED]
- **Epic**: ADVANCED-QUERY-SYNTAX-1
- **Status**: Proposed
- **Dependencies**: unified-search
- **Description**: Power user query syntax
- **Examples**:
  - `type:conversation project:cc-viz modified:>2026-01-01`
  - `tool:Bash status:active`
  - `modified:>2026-01-01 project:*`

#### 💡 search-ui [PROPOSED]
- **Epic**: SEARCH-UI-1
- **Status**: Proposed
- **Dependencies**: unified-search, advanced-query-syntax
- **Description**: Unified search interface
- **Includes**:
  - Top-level search bar (all pages)
  - Result categorization and grouping
  - Inline previews and navigation
  - Saved searches

---

## Phase 3: Intelligence & Visualization 📋 [QUEUED]

**Goal**: Transform raw data into understanding and visual exploration.

**Status**: 0/4 completed

### Topics

#### 💡 activity-dashboard [PROPOSED]
- **Epic**: ACTIVITY-DASHBOARD-1
- **Status**: Proposed
- **Dependencies**: conversation-indexer, session-data-indexer
- **Description**: Turn stats into understanding
- **Includes**:
  - Live status (active sessions, recent conversations, unfinished todos)
  - Activity patterns (most active hours, trends, per-project breakdown)
  - Tool usage breakdown
  - Token usage by model over time

#### 💡 time-machine [PROPOSED]
- **Epic**: TIME-MACHINE-1
- **Status**: Proposed
- **Dependencies**: conversation-indexer, relationship-linker
- **Description**: Session replay with correlated file history
- **Includes**:
  - Timeline scrubber for any conversation
  - File diffs at each timestamp
  - Tool calls synchronized with file changes
  - File version restore capability

#### 💡 subagent-observatory [PROPOSED]
- **Epic**: SUBAGENT-OBSERVATORY-1
- **Status**: Proposed
- **Dependencies**: subagent-graph, conversation-indexer
- **Description**: Visualize agent hierarchies and flow
- **Includes**:
  - Tree visualization of spawn hierarchies
  - Click-through to agent conversations
  - Per-agent metrics (tokens, duration, files modified)
  - Aggregate views (most common agents, average chain depth)

#### 💡 space-manager [PROPOSED]
- **Epic**: SPACE-MANAGER-1
- **Status**: Proposed
- **Description**: Understand and clean up ~/.claude
- **Includes**:
  - Storage breakdown by directory
  - Age and staleness indicators
  - Safe-to-delete identification
  - Bulk actions: archive, compress, clean, export
  - Smart suggestions for cleanup

---

## Phase 4: Extension Management 💭 [QUEUED]

**Goal**: Consolidate extension management and show usage analytics.

**Status**: 0/2 completed

### Topics

#### 💡 extension-workshop [PROPOSED]
- **Epic**: EXTENSION-WORKSHOP-1
- **Status**: Proposed
- **Dependencies**: conversation-indexer
- **Description**: Browse and manage agents, commands, skills, plugins
- **Includes**:
  - Unified browser for all extension types
  - Extraction of name, description from markdown
  - Size and usage stats
  - Enable/disable toggles
  - Quick actions: edit, delete, create new

#### 💡 cross-session-todo-board [PROPOSED]
- **Epic**: CROSS-SESSION-TODO-BOARD-1
- **Status**: Proposed
- **Dependencies**: session-data-indexer
- **Description**: Kanban aggregation of all todos
- **Includes**:
  - Aggregated todo view across all sessions
  - Kanban columns: Pending, In Progress, Completed
  - Click-through to source session
  - Filters: session, project, age, keywords
  - Live updates during active sessions

---

## Dependency Graph

```
Session Data Indexer ✅
    ↓
Conversation Indexer → Subagent Graph
    ↓                       ↓
Relationship Linker ←─────┘
    ↓
├── Unified Search
│   ├── Advanced Query Syntax
│   └── Search UI
├── Activity Dashboard
├── Time Machine
├── Subagent Observatory
├── Space Manager
├── Extension Workshop
└── Cross-Session Todo Board
```

---

## Success Metrics

- ✓ Any UUID in the UI is clickable and navigates to context
- ✓ Search finds content regardless of source type
- ✓ Users can answer "what did I work on last week" in <30 seconds
- ✓ Storage cleanup recovers space without data loss risk
- ✓ Subagent hierarchies are visible and navigable
- ✓ File history correlates with conversation events

---

## Reference Materials

- **Full Vision**: [FEATURE_PROPOSAL_cc-viz-ultimate-vision.md](.agent_planning/FEATURE_PROPOSAL_cc-viz-ultimate-vision.md)
- **Handoff**: [HANDOFF-cc-viz-ultimate-vision-20260112.md](.agent_planning/HANDOFF-cc-viz-ultimate-vision-20260112.md)
- **Session Data Plan**: [PLAN-sprint2-20260112.md](.agent_planning/cc-viz-roadmap/PLAN-sprint2-20260112.md)

---

## Next Steps

1. **Select next feature**: Choose from Phase 1 topics (conversation-indexer, subagent-graph, or relationship-linker)
2. **Create implementation plan**: Run `/do:plan <feature-name>` to generate detailed plan
3. **Track progress**: Use beads (`bd create`, `bd update`) for multi-session work
4. **Iterate**: Each phase builds on previous - maintain dependency order
