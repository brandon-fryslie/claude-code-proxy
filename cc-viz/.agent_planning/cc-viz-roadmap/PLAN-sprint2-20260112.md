# Sprint Plan: Session Data Visualizer
Generated: 2026-01-12 (Revised)

## Sprint Goal
Implement Session Data visualizer (todos, plans) with proper database ingestion and storage.

---

## Scope

**In scope (this sprint):**
- P0: Session Data Visualizer (todos, plans) with database backend

**Explicitly out of scope (future sprints):**
- Extensibility Browser (agents, commands, skills)
- Plugins visualizer
- History/File-History visualizer
- Telemetry visualizer

---

## Architecture: Database-First Approach

**CRITICAL RULE**: Never load JSON from filesystem to return in web requests. All data must be:
1. Ingested into SQLite database
2. Cleaned and validated during ingestion
3. Served from database via API

### Database Schema

```sql
-- Todos table
CREATE TABLE IF NOT EXISTS claude_todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_uuid TEXT NOT NULL,
    agent_uuid TEXT,
    file_path TEXT UNIQUE NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL,  -- 'pending', 'in_progress', 'completed'
    active_form TEXT,
    item_index INTEGER NOT NULL,  -- Position in original array
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL,
    UNIQUE(file_path, item_index)
);

CREATE INDEX idx_todos_session ON claude_todos(session_uuid);
CREATE INDEX idx_todos_status ON claude_todos(status);
CREATE INDEX idx_todos_modified ON claude_todos(modified_at);

-- Todo sessions (aggregated view)
CREATE TABLE IF NOT EXISTS claude_todo_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_uuid TEXT NOT NULL,
    agent_uuid TEXT,
    file_path TEXT UNIQUE NOT NULL,
    file_size INTEGER NOT NULL,
    todo_count INTEGER NOT NULL,
    pending_count INTEGER NOT NULL,
    in_progress_count INTEGER NOT NULL,
    completed_count INTEGER NOT NULL,
    modified_at DATETIME NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Plans table
CREATE TABLE IF NOT EXISTS claude_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,  -- Human-readable parsed name
    content TEXT NOT NULL,
    preview TEXT NOT NULL,  -- First 200 chars
    file_size INTEGER NOT NULL,
    modified_at DATETIME NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plans_modified ON claude_plans(modified_at);
```

### Indexer Service

Add to existing indexer pattern in `proxy/internal/service/indexer.go`:

```go
// IndexTodos scans ~/.claude/todos/ and ingests into database
func (i *Indexer) IndexTodos() error {
    todosDir := filepath.Join(i.claudeDir, "todos")

    // Walk directory
    // For each .json file:
    //   1. Parse filename to extract session_uuid, agent_uuid
    //   2. Read and parse JSON array
    //   3. Validate each item has required fields
    //   4. Insert/update in database
    //   5. Update session aggregate table

    return nil
}

// IndexPlans scans ~/.claude/plans/ and ingests into database
func (i *Indexer) IndexPlans() error {
    plansDir := filepath.Join(i.claudeDir, "plans")

    // Walk directory
    // For each .md file:
    //   1. Parse filename to create display_name
    //   2. Read content
    //   3. Generate preview (first 200 chars)
    //   4. Insert/update in database

    return nil
}
```

### Data Validation

During ingestion, handle:
- **Empty files**: Skip or mark as empty in DB
- **Invalid JSON**: Log error, skip file
- **Missing fields**: Use defaults (empty string, 'pending')
- **Corrupt data**: Log and skip
- **Duplicate entries**: Upsert based on file_path + item_index

---

## P0: Session Data Visualizer

### API Endpoints (Database-Backed)

```go
// GET /api/v2/claude/todos
// Returns aggregated stats and session list from database
type TodosResponse struct {
    TotalFiles      int           `json:"total_files"`
    NonEmptyFiles   int           `json:"non_empty_files"`
    StatusBreakdown StatusCounts  `json:"status_breakdown"`
    Sessions        []TodoSession `json:"sessions"`
    LastIndexed     time.Time     `json:"last_indexed"`
}

// GET /api/v2/claude/todos/{session_uuid}
// Returns todos for a specific session from database
type TodoDetailResponse struct {
    SessionUUID string     `json:"session_uuid"`
    AgentUUID   string     `json:"agent_uuid"`
    FilePath    string     `json:"file_path"`
    ModifiedAt  time.Time  `json:"modified_at"`
    Todos       []TodoItem `json:"todos"`
}

// GET /api/v2/claude/plans
// Returns all plans from database
type PlansResponse struct {
    TotalCount  int           `json:"total_count"`
    TotalSize   int64         `json:"total_size"`
    Plans       []PlanSummary `json:"plans"`
    LastIndexed time.Time     `json:"last_indexed"`
}

// GET /api/v2/claude/plans/{id}
// Returns single plan content from database
type PlanDetailResponse struct {
    ID         int       `json:"id"`
    FileName   string    `json:"file_name"`
    Name       string    `json:"name"`
    Content    string    `json:"content"`
    Size       int64     `json:"size"`
    ModifiedAt time.Time `json:"modified_at"`
}

// POST /api/v2/claude/todos/reindex
// Triggers re-indexing of todos directory
type ReindexResponse struct {
    FilesProcessed int       `json:"files_processed"`
    TodosIndexed   int       `json:"todos_indexed"`
    Errors         []string  `json:"errors"`
    Duration       string    `json:"duration"`
}
```

### UI Design

```
┌─────────────────────────────────────────────────────────────────────┐
│ Session Data                                                         │
│ Debug logs, todos, plans, and session environment                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ [Todos] [Plans]                                    ← Tab navigation │
│                                                                      │
│ ═══════════════════════════════════════════════════════════════════ │
│                                                                      │
│ TODOS TAB:                                                          │
│                                                                      │
│ ┌─ Summary ─────────────────────────────────────────────────────────┐
│ │  1,038 Files  │  184 Non-empty  │  ● 45 Pending  ● 12 In Progress │
│ │               │                 │  ● 892 Completed                │
│ │                              Last indexed: 2 min ago [Refresh ↻] │
│ └───────────────────────────────────────────────────────────────────┘
│                                                                      │
│ ┌─ Sessions ──────────────────┐  ┌─ Session Detail ────────────────┐│
│ │ [Search...]                 │  │ Session: 016c7c55-1a1e...       ││
│ │                             │  │ Agent: aee4a4e8-a8af...         ││
│ │ ● 016c7c55... (5 todos)  6m│  │                                  ││
│ │   ✓4 ●1 ○0                 │  │ ✓ Build Configuration endpoint   ││
│ │                             │  │ ● Update sidebar navigation      ││
│ │ ○ 00f953f7... (3 todos) 2h │  │ ○ Write documentation            ││
│ │   ✓2 ●0 ○1                 │  │                                  ││
│ └─────────────────────────────┘  └──────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Acceptance Criteria

**Database & Indexing:**
- [ ] SQLite tables created on startup (claude_todos, claude_todo_sessions, claude_plans)
- [ ] Indexer scans ~/.claude/todos/ and ingests valid entries
- [ ] Indexer scans ~/.claude/plans/ and ingests content
- [ ] Invalid/corrupt files logged and skipped
- [ ] Incremental indexing (only process modified files)

**API Endpoints:**
- [ ] `GET /api/v2/claude/todos` returns data from database
- [ ] `GET /api/v2/claude/todos/{session_uuid}` returns session details from database
- [ ] `GET /api/v2/claude/plans` returns plans from database
- [ ] `GET /api/v2/claude/plans/{id}` returns plan content from database
- [ ] `POST /api/v2/claude/todos/reindex` triggers manual reindex

**Frontend:**
- [ ] Route `/cc-viz/session-data` loads without errors
- [ ] Todos tab shows summary stats from database
- [ ] Session list with search capability
- [ ] Clicking session shows todos with status badges
- [ ] Plans tab shows all plans with content viewer
- [ ] "Refresh" button triggers reindex
- [ ] Last indexed timestamp displayed

**Navigation:**
- [ ] Sidebar "Session Data" item enabled
- [ ] Home page "Session Data" card clickable

---

## Implementation Sequence

### Phase 1: Database Schema (Day 1)
1. Add tables to storage_sqlite.go InitSchema()
2. Create indexes for efficient queries

### Phase 2: Indexer (Day 1-2)
1. Add IndexTodos() method with validation
2. Add IndexPlans() method
3. Add incremental indexing (check modified_at)
4. Handle errors gracefully, log issues

### Phase 3: API Endpoints (Day 2)
1. Implement GetTodosV2() - query database
2. Implement GetTodoDetailV2() - query database
3. Implement GetPlansV2() - query database
4. Implement GetPlanDetailV2() - query database
5. Implement ReindexTodosV2() - trigger indexer

### Phase 4: Frontend (Day 3)
1. Create SessionData.tsx with tabs
2. Add TypeScript interfaces
3. Add query hooks
4. Implement TodosPanel
5. Implement PlansPanel

### Phase 5: Integration & Verification (Day 4)
1. Test indexer with actual data
2. Verify API returns correct data
3. Test UI via Chrome DevTools
4. Handle edge cases (empty files, corrupt data)

---

## Dependencies
- Existing SQLite database infrastructure
- Existing indexer patterns

## Risks
- Large todo directory (1,038 files) - mitigate with incremental indexing
- Invalid JSON files - graceful error handling during ingestion
- Performance with 10,000+ todo items - use pagination and indexes
