# Implementation Context: CC-VIZ Sprint 2 (Revised)
Generated: 2026-01-12

**Scope**: Session Data Visualizer only - database-backed

---

## Critical Rule

**NEVER load JSON from filesystem to return in web requests.**

All data must be:
1. Ingested into SQLite database by an indexer
2. Cleaned/validated during ingestion
3. Served from database via API

---

## File Locations

### Backend Files
| File | Purpose |
|------|---------|
| `proxy/internal/service/storage_sqlite.go` | Add tables to InitSchema() |
| `proxy/internal/service/indexer.go` | Add IndexTodos(), IndexPlans() |
| `proxy/internal/handler/data_handler.go` | Add API handlers |
| `proxy/cmd/proxy-data/main.go` | Register routes |

### Frontend Files
| File | Purpose |
|------|---------|
| `cc-viz/src/lib/types.ts` | Add TypeScript interfaces |
| `cc-viz/src/lib/api.ts` | Add query hooks |
| `cc-viz/src/main.tsx` | Add route |
| `cc-viz/src/pages/SessionData.tsx` | NEW - Session Data page |
| `cc-viz/src/components/layout/AppLayout.tsx` | Enable sidebar item |
| `cc-viz/src/pages/Home.tsx` | Add href to card |

### Data Source Paths
| Path | Description |
|------|-------------|
| `~/.claude/todos/` | 1,038 JSON files |
| `~/.claude/plans/` | 12 markdown files |

---

## Database Schema

Add to `storage_sqlite.go` InitSchema():

```go
// Claude todos - individual items
_, err = s.db.Exec(`
    CREATE TABLE IF NOT EXISTS claude_todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_uuid TEXT NOT NULL,
        agent_uuid TEXT,
        file_path TEXT NOT NULL,
        content TEXT NOT NULL,
        status TEXT NOT NULL,
        active_form TEXT,
        item_index INTEGER NOT NULL,
        modified_at DATETIME NOT NULL,
        indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(file_path, item_index)
    )
`)

_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_session ON claude_todos(session_uuid)`)
_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_status ON claude_todos(status)`)
_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_modified ON claude_todos(modified_at)`)

// Claude todo sessions - aggregated
_, err = s.db.Exec(`
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
    )
`)

// Claude plans
_, err = s.db.Exec(`
    CREATE TABLE IF NOT EXISTS claude_plans (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_name TEXT UNIQUE NOT NULL,
        display_name TEXT NOT NULL,
        content TEXT NOT NULL,
        preview TEXT NOT NULL,
        file_size INTEGER NOT NULL,
        modified_at DATETIME NOT NULL,
        indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
`)

_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_plans_modified ON claude_plans(modified_at)`)
```

---

## Indexer Implementation

Add to `indexer.go`:

```go
// IndexTodos scans ~/.claude/todos/ and ingests into database
func (i *Indexer) IndexTodos() (int, int, []string) {
    todosDir := filepath.Join(i.claudeDir, "todos")
    filesProcessed := 0
    todosIndexed := 0
    var errors []string

    err := filepath.Walk(todosDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
            return nil
        }

        filesProcessed++

        // Parse filename: {session_uuid}-agent-{agent_uuid}.json
        baseName := strings.TrimSuffix(info.Name(), ".json")
        parts := strings.Split(baseName, "-agent-")
        sessionUUID := parts[0]
        agentUUID := ""
        if len(parts) > 1 {
            agentUUID = parts[1]
        }

        // Read and parse file
        content, err := os.ReadFile(path)
        if err != nil {
            errors = append(errors, fmt.Sprintf("%s: read error: %v", info.Name(), err))
            return nil
        }

        if len(content) < 3 { // Empty or just "[]"
            return nil
        }

        var todos []struct {
            Content    string `json:"content"`
            Status     string `json:"status"`
            ActiveForm string `json:"activeForm"`
        }

        if err := json.Unmarshal(content, &todos); err != nil {
            errors = append(errors, fmt.Sprintf("%s: JSON parse error: %v", info.Name(), err))
            return nil
        }

        // Insert/update in database
        // ... transaction with upsert logic ...

        todosIndexed += len(todos)
        return nil
    })

    if err != nil {
        errors = append(errors, fmt.Sprintf("walk error: %v", err))
    }

    return filesProcessed, todosIndexed, errors
}

// IndexPlans scans ~/.claude/plans/ and ingests into database
func (i *Indexer) IndexPlans() (int, []string) {
    plansDir := filepath.Join(i.claudeDir, "plans")
    plansIndexed := 0
    var errors []string

    err := filepath.Walk(plansDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
            return nil
        }

        content, err := os.ReadFile(path)
        if err != nil {
            errors = append(errors, fmt.Sprintf("%s: read error: %v", info.Name(), err))
            return nil
        }

        // Parse filename for display name: "peppy-yawning-teapot" -> "Peppy Yawning Teapot"
        baseName := strings.TrimSuffix(info.Name(), ".md")
        displayName := formatDisplayName(baseName)

        // Generate preview
        preview := string(content)
        if len(preview) > 200 {
            preview = preview[:200] + "..."
        }

        // Upsert into database
        // ...

        plansIndexed++
        return nil
    })

    if err != nil {
        errors = append(errors, fmt.Sprintf("walk error: %v", err))
    }

    return plansIndexed, errors
}

func formatDisplayName(s string) string {
    words := strings.Split(s, "-")
    for i, w := range words {
        if len(w) > 0 {
            words[i] = strings.ToUpper(w[:1]) + w[1:]
        }
    }
    return strings.Join(words, " ")
}
```

---

## API Handlers

Add to `data_handler.go`:

```go
// GetTodosV2 returns todo stats and sessions from database
func (h *DataHandler) GetTodosV2(w http.ResponseWriter, r *http.Request) {
    // Query database for:
    // - Total session count
    // - Sessions with todo_count > 0
    // - Sum of pending_count, in_progress_count, completed_count
    // - List of sessions sorted by modified_at desc

    response := map[string]interface{}{
        "total_files":      totalSessions,
        "non_empty_files":  nonEmptySessions,
        "status_breakdown": statusCounts,
        "sessions":         sessions,
        "last_indexed":     lastIndexedTime,
    }
    writeJSONResponse(w, response)
}

// GetTodoDetailV2 returns todos for a specific session
func (h *DataHandler) GetTodoDetailV2(w http.ResponseWriter, r *http.Request) {
    sessionUUID := mux.Vars(r)["session_uuid"]

    // Query claude_todos WHERE session_uuid = ?
    // ORDER BY item_index

    if len(todos) == 0 {
        writeErrorResponse(w, "Session not found", http.StatusNotFound)
        return
    }

    writeJSONResponse(w, response)
}

// GetPlansV2 returns all plans from database
func (h *DataHandler) GetPlansV2(w http.ResponseWriter, r *http.Request) {
    // Query claude_plans ORDER BY modified_at desc
    writeJSONResponse(w, response)
}

// GetPlanDetailV2 returns a specific plan
func (h *DataHandler) GetPlanDetailV2(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    // Query by id (integer) or file_name
    writeJSONResponse(w, response)
}

// ReindexTodosV2 triggers manual reindexing
func (h *DataHandler) ReindexTodosV2(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    filesProcessed, todosIndexed, errors := h.indexer.IndexTodos()
    plansIndexed, planErrors := h.indexer.IndexPlans()

    errors = append(errors, planErrors...)

    response := map[string]interface{}{
        "files_processed": filesProcessed,
        "todos_indexed":   todosIndexed,
        "plans_indexed":   plansIndexed,
        "errors":          errors,
        "duration":        time.Since(start).String(),
    }
    writeJSONResponse(w, response)
}
```

---

## Route Registration

Add to `proxy-data/main.go`:

```go
// Session Data endpoints
r.HandleFunc("/api/v2/claude/todos", h.GetTodosV2).Methods("GET")
r.HandleFunc("/api/v2/claude/todos/reindex", h.ReindexTodosV2).Methods("POST")
r.HandleFunc("/api/v2/claude/todos/{session_uuid}", h.GetTodoDetailV2).Methods("GET")
r.HandleFunc("/api/v2/claude/plans", h.GetPlansV2).Methods("GET")
r.HandleFunc("/api/v2/claude/plans/{id}", h.GetPlanDetailV2).Methods("GET")
```

Note: Order matters - `/todos/reindex` must come before `/todos/{session_uuid}`

---

## TypeScript Interfaces

Add to `types.ts`:

```typescript
// Todo types
export interface TodoStatusCounts {
  pending: number;
  in_progress: number;
  completed: number;
}

export interface TodoSession {
  id: number;
  session_uuid: string;
  agent_uuid: string;
  file_path: string;
  todo_count: number;
  pending_count: number;
  in_progress_count: number;
  completed_count: number;
  modified_at: string;
}

export interface TodosResponse {
  total_files: number;
  non_empty_files: number;
  status_breakdown: TodoStatusCounts;
  sessions: TodoSession[];
  last_indexed: string;
}

export interface TodoItem {
  content: string;
  status: 'pending' | 'in_progress' | 'completed';
  active_form: string;
}

export interface TodoDetailResponse {
  session_uuid: string;
  agent_uuid: string;
  file_path: string;
  modified_at: string;
  todos: TodoItem[];
}

// Plan types
export interface PlanSummary {
  id: number;
  file_name: string;
  display_name: string;
  preview: string;
  file_size: number;
  modified_at: string;
}

export interface PlansResponse {
  total_count: number;
  total_size: number;
  plans: PlanSummary[];
  last_indexed: string;
}

export interface PlanDetailResponse {
  id: number;
  file_name: string;
  display_name: string;
  content: string;
  file_size: number;
  modified_at: string;
}

export interface ReindexResponse {
  files_processed: number;
  todos_indexed: number;
  plans_indexed: number;
  errors: string[];
  duration: string;
}
```

---

## Query Hooks

Add to `api.ts`:

```typescript
export function useTodos() {
  return useQuery({
    queryKey: ['claude-todos'],
    queryFn: () => fetchAPI<TodosResponse>('/claude/todos'),
  });
}

export function useTodoDetail(sessionUuid: string | null) {
  return useQuery({
    queryKey: ['claude-todos', sessionUuid],
    queryFn: () => fetchAPI<TodoDetailResponse>(`/claude/todos/${sessionUuid}`),
    enabled: !!sessionUuid,
  });
}

export function usePlans() {
  return useQuery({
    queryKey: ['claude-plans'],
    queryFn: () => fetchAPI<PlansResponse>('/claude/plans'),
  });
}

export function usePlanDetail(id: number | null) {
  return useQuery({
    queryKey: ['claude-plans', id],
    queryFn: () => fetchAPI<PlanDetailResponse>(`/claude/plans/${id}`),
    enabled: id !== null,
  });
}

export function useReindexTodos() {
  return useMutation({
    mutationFn: () => fetch('/api/v2/claude/todos/reindex', { method: 'POST' }).then(r => r.json()),
  });
}
```

---

## Sample Data

### Todo File
Path: `~/.claude/todos/016c7c55-1a1e-4257-8a30-befc355f6541-agent-aee4a4e8-a8af-48df-a141-495997bb6019.json`
```json
[
  {
    "content": "Build Configuration endpoint",
    "status": "completed",
    "activeForm": "Building Configuration API endpoint"
  },
  {
    "content": "Update sidebar navigation",
    "status": "in_progress",
    "activeForm": "Updating sidebar navigation"
  }
]
```

### Plan File
Path: `~/.claude/plans/peppy-yawning-teapot.md`
- Display name: "Peppy Yawning Teapot"
- Content: Full markdown

---

## Startup Sequence

1. proxy-data starts
2. InitSchema() creates tables if not exist
3. Indexer runs on startup (or first request)
4. API serves from database

Consider: Background indexer goroutine that periodically re-indexes
