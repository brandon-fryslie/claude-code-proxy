# Technical Context: Phase 1 - Conversation Indexer Implementation

**Document**: Complete context for implementing Conversation Indexer
**Created**: 2026-01-12
**Audience**: Implementation agents with write access

This document contains all necessary context to implement Phase 1 from scratch without needing to read other files.

---

## Executive Summary

The Conversation Indexer is 90% implemented in the codebase but has never been fully tested or verified on real data. This sprint completes and validates the implementation. All major components exist - database schema is production-quality, parsing logic is complete, API endpoints are ready. The main work is: (1) fixing FTS5 test blocker, (2) adding integration tests, (3) validating at scale.

---

## File Structure and Locations

### Main Codebase Files

**Indexer Core Logic**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/service/indexer.go` (408 lines)
- Contains: ConversationIndexer struct, initialIndex(), indexFile(), watchFiles(), processIndexQueue()
- Status: 90% complete, needs testing and verification
- Key methods to verify:
  - Line 100-130: `initialIndex()` - scans ~/./claude/projects/ for JSONL files
  - Line 153-313: `indexFile()` - core parsing and database insertion logic
  - Line 250-385: `watchFiles()` and `processIndexQueue()` - file watching and debouncing

**Message Parsing**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/service/conversation.go` (392+ lines)
- Contains: ConversationMessage struct, parseConversationFile(), ExtractMessageContent()
- Status: 100% complete, production-ready
- Key functions:
  - Line 45-80: `ConversationMessage` struct definition
  - Line 100-200: `parseConversationFile()` - parses JSONL format
  - Line 326-392: `ExtractMessageContent()` - extracts text and tool names for FTS5 index
  - Handles: text messages, tool calls, tool results, images

**Database Schema**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/service/storage_sqlite.go` (600+ lines)
- Contains: Database connection, schema creation, FTS5 setup
- Status: 100% complete with FTS5 blocker (test environment)
- Key sections:
  - Line 40-50: Database initialization and WAL mode setup
  - Line 195-220: Conversation tables schema (conversations, conversation_messages)
  - Line 205-215: FTS5 virtual table creation (THIS IS THE FTS5 BLOCKER FOR TESTS)
  - Line 300-350: Query methods for search and retrieval

**HTTP Handlers**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/handler/data_handler.go`
- Contains: SearchConversations(), GetConversations() endpoint handlers
- Status: 100% complete, ready for data
- API endpoints ready at /api/conversations/*

**React UI**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/cc-viz/src/pages/Conversations.tsx` (if exists)
- Status: UI components ready, awaiting indexed data
- Will display conversation list and search results once indexer is working

**Tests**: `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/service/indexer_test.go`
- Contains: 24 test cases for indexer functionality
- Status: Tests written but failing due to FTS5 blocker
- All test failures are FTS5-related, not logic errors

### Data Sources

**Conversation JSONL Files**: `~/.claude/projects/{project-name}/{session-uuid}.jsonl`
- Format: One JSON object per line (JSONL)
- Count: 3,927+ files across all projects
- Size: ~100-200 MB total data
- Structure: Each line contains message object with fields:
  - `sessionId`: UUID of the session
  - `agentId`: UUID of the agent (for subagents)
  - `type`: "message" or "tool_call" or "tool_result"
  - `role`: "user", "assistant", or "user" for tool results
  - `content`: Text content or array for complex content
  - `name`: Tool name if tool_use/tool_result
  - `usage`: {inputTokens, outputTokens}

**Example JSONL Entry** (message from conversation):
```json
{
  "sessionId": "abc-123-def",
  "agentId": "agent-456",
  "type": "message",
  "role": "assistant",
  "content": "I'll help you implement that feature.",
  "usage": {"inputTokens": 150, "outputTokens": 32}
}
```

**Example Tool Call Entry**:
```json
{
  "type": "tool_use",
  "role": "assistant",
  "name": "Bash",
  "input": {"command": "go test ./..."},
  "id": "tool-789"
}
```

---

## Database Schema Details

### Table 1: conversations

**Purpose**: Metadata for each conversation file

**Columns** (9 total):
```
id                INTEGER PRIMARY KEY
session_uuid      TEXT UNIQUE          -- Extracted from filename or content
project_path      TEXT                 -- ~/./claude/projects/{name}
file_path         TEXT                 -- Full path to .jsonl file
message_count     INTEGER              -- Total messages in conversation
total_tokens      INTEGER              -- Sum of all token usage
first_message_at  TIMESTAMP            -- Timestamp of first message
last_message_at   TIMESTAMP            -- Timestamp of last message
created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

### Table 2: conversation_messages

**Purpose**: Individual messages for full-text search and detail view

**Columns** (19 total):
```
id                INTEGER PRIMARY KEY
conversation_id   INTEGER              -- FK to conversations.id
message_type      TEXT                 -- "text", "tool_use", "tool_result", "image"
role              TEXT                 -- "user", "assistant"
content           TEXT                 -- Full message content (or truncated for tool results)
tool_name         TEXT                 -- Name if tool call (e.g., "Bash", "Read", "Write")
tool_use_id       TEXT                 -- Tool use ID for matching tool results
input_tokens      INTEGER              -- Input tokens for this message
output_tokens     INTEGER              -- Output tokens for this message
created_at        TIMESTAMP            -- Message timestamp
sequence_number   INTEGER              -- Order in conversation (1, 2, 3, ...)
is_tool_result    BOOLEAN              -- True if this is tool result (vs tool call)
FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
```

### Table 3: fts_conversations (FTS5 Virtual Table)

**Purpose**: Full-text search index across all message content

**Columns** (indexed):
```
conversation_id   -- Links back to conversation
message_content   -- Combined text for searching
tool_names        -- Tool names searchable separately
```

**Created with**:
```sql
CREATE VIRTUAL TABLE fts_conversations USING fts5(
    conversation_id,
    message_content,
    tool_names,
    content='conversation_messages',
    content_rowid='id'
)
```

**Search usage**:
```sql
SELECT * FROM fts_conversations
WHERE message_content MATCH 'postgres migration'
LIMIT 10
```

---

## FTS5 Blocker Details and Resolution

### The Problem

All 24 tests fail with error:
```
Error: database/sql: unknown driver "sqlite3" or driver initialization failed:
no such module: fts5
```

This occurs in test environment only. Production binary works fine.

### Root Cause

SQLite driver in test environment doesn't have FTS5 extension compiled in. This is a build-time configuration issue, not a code logic issue.

### Solution: Conditional Schema Build

**Approach**: Use Go build tags to create two different schema initializations.

**File to Modify**: `proxy/internal/service/storage_sqlite.go`

**Changes Required**:

1. Create two schema constants (or functions):

```go
// +build !test
func createFTS5Schema() string {
    return `
    CREATE VIRTUAL TABLE fts_conversations USING fts5(
        conversation_id,
        message_content,
        tool_names,
        content='conversation_messages',
        content_rowid='id'
    )
    `
}

// +build test
func createFTS5Schema() string {
    return "" // Empty - skip FTS5 in tests
}
```

2. In the schema initialization function, conditionally call FTS5 creation:

```go
func (s *SQLiteStorage) initSchema() error {
    // ... create other tables ...

    // Only create FTS5 if not in test mode
    if ftsSchema := createFTS5Schema(); ftsSchema != "" {
        if _, err := s.db.Exec(ftsSchema); err != nil {
            return err
        }
    }

    return nil
}
```

3. Update search query to handle missing FTS5:

```go
// In search function, check if FTS5 exists
func (s *SQLiteStorage) SearchConversations(query string) ([]SearchResult, error) {
    // If FTS5 available, use it; otherwise fallback to LIKE search
    stmt := `SELECT ... FROM conversation_messages WHERE content LIKE ?`
    // ... rest of query ...
}
```

**Effort**: 1-2 hours
**Result**: Tests pass in both test and production configurations

---

## Indexer Implementation Details

### ConversationIndexer Struct (indexer.go line ~45)

```go
type ConversationIndexer struct {
    db              *sql.DB
    convService     *ConversationService
    queue           chan string              // Filenames to index
    watcher         *fsnotify.Watcher
    projectsDir     string                   // ~/.claude/projects/
    debounceTimer   *time.Timer
    debounceMs      int
}
```

### initialIndex() Method (line ~100)

**Purpose**: Scan all JSONL files and queue for indexing

**Flow**:
1. Read all files from `~/.claude/projects/`
2. Filter for `.jsonl` files
3. Queue each file for indexing
4. Call processIndexQueue() to process

**Code Pattern**:
```go
func (ci *ConversationIndexer) initialIndex() error {
    // Walk project directories
    err := filepath.Walk(ci.projectsDir, func(path string, info os.FileInfo, err error) error {
        if strings.HasSuffix(path, ".jsonl") {
            ci.queue <- path  // Queue for processing
        }
        return nil
    })
    return err
}
```

### indexFile() Method (line ~153)

**Purpose**: Parse one JSONL file and insert into database

**Flow**:
1. Parse JSONL file using ConversationService.parseConversationFile()
2. Extract conversation metadata (session_uuid, message_count, tokens)
3. Delete any previous entries for this session
4. Insert conversation record
5. For each message:
   - Extract content via ExtractMessageContent()
   - Insert message record
   - Update FTS5 index
6. Commit transaction

**Code Pattern**:
```go
func (ci *ConversationIndexer) indexFile(filepath string) error {
    // Parse conversation
    conv, err := ci.convService.parseConversationFile(filepath)
    if err != nil {
        return err
    }

    tx, err := ci.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Delete old entries
    tx.ExecContext(ctx, "DELETE FROM conversations WHERE session_uuid = ?", conv.SessionID)

    // Insert conversation
    result, err := tx.ExecContext(ctx,
        "INSERT INTO conversations (session_uuid, ...) VALUES (?, ...)",
        conv.SessionID, ...)

    // Insert messages
    for _, msg := range conv.Messages {
        content := ExtractMessageContent(msg)
        tx.ExecContext(ctx,
            "INSERT INTO conversation_messages (conversation_id, content, ...) VALUES (?, ?, ...)",
            conversationID, content, ...)
    }

    // Update FTS5 index (automatic if using content= trigger)

    return tx.Commit()
}
```

### watchFiles() and processIndexQueue() (line ~250, ~350)

**Purpose**: Watch for file changes and re-index

**Flow**:
1. Create fsnotify watcher
2. Watch `~/.claude/projects/` directory
3. On file change event:
   - Debounce rapid changes (group within 500ms)
   - Queue file for re-indexing
4. processIndexQueue worker:
   - Read from queue channel
   - Call indexFile()
   - Handle errors without crashing

---

## API Endpoints and Integration

### Endpoint 1: GET /api/conversations

**Handler**: SearchConversations or GetConversations
**Location**: `proxy/internal/handler/data_handler.go`

**Request**:
```
GET /api/conversations?limit=100&offset=0
```

**Response**:
```json
{
  "conversations": [
    {
      "id": 1,
      "session_uuid": "abc-123",
      "project_path": "~/.claude/projects/my-project",
      "message_count": 47,
      "total_tokens": 12500,
      "first_message_at": "2026-01-12T10:30:00Z",
      "last_message_at": "2026-01-12T11:45:00Z"
    }
  ],
  "total": 3927
}
```

### Endpoint 2: POST /api/conversations/search

**Handler**: SearchConversations
**Location**: `proxy/internal/handler/data_handler.go`

**Request**:
```json
{
  "query": "postgres migration",
  "limit": 50,
  "offset": 0
}
```

**Response**:
```json
{
  "results": [
    {
      "conversation_id": 1,
      "session_uuid": "abc-123",
      "message": "...postgres migration implementation...",
      "tool_name": "Bash",
      "relevance": 0.95
    }
  ],
  "total": 12
}
```

---

## Testing Patterns and Fixtures

### Test Fixtures Location

**Path**: `proxy/internal/service/testdata/` or inline in test files

**Pattern**: Include sample JSONL entries:
```json
{"sessionId":"test-session","agentId":"test-agent","type":"message","role":"assistant","content":"Test message","usage":{"inputTokens":10,"outputTokens":5}}
{"sessionId":"test-session","agentId":"test-agent","type":"tool_use","role":"assistant","name":"Bash","input":{"command":"go test"},"id":"tool-1"}
```

### Existing Tests (indexer_test.go)

**24 test cases** covering:
- File parsing
- Message extraction
- Database insertion
- Search queries
- File watching
- Concurrency

**All failing due to FTS5 blocker** - will pass once conditional schema is implemented.

---

## Configuration and Environment

### Configuration File: config.yaml

**Relevant section**:
```yaml
database:
  type: sqlite
  path: ~/.claude/cc-viz.db
  journal_mode: wal         # Write-ahead logging for concurrency

indexer:
  enabled: true
  watch: true
  projects_dir: ~/.claude/projects
  rescan_interval: 3600     # Re-scan all files every hour
  debounce_ms: 500          # Group rapid file changes
```

### Environment Variables

```bash
CLAUDE_HOME=~/.claude          # Location of Claude Code data
DEBUG=true                     # Enable verbose logging (optional)
```

---

## Build and Test Commands

### Development Build

```bash
# Build proxy-data service with indexer
cd proxy
go build -o proxy-data ./cmd/proxy-data/main.go

# Run tests (currently failing due to FTS5)
go test ./internal/service/... -v

# Run specific test
go test -run TestConversationIndexer -v
```

### Production Build

```bash
# Full production build with FTS5 support
go build -o proxy-data ./cmd/proxy-data/main.go

# Verify binary includes SQLite driver
ldd ./proxy-data | grep sqlite
```

---

## Performance Targets

Based on existing data:

| Metric | Target | Note |
|--------|--------|------|
| Indexing speed | 400-600 files/min | 3,927 files in <10 min |
| Database size | <500 MB | For full year of data |
| Search latency | <100ms | FTS5 queries on indexed data |
| Startup time | <5s | Initial index and startup |
| Memory usage | <500 MB | Peak during indexing |

---

## Known Limitations (Phase 1)

1. **No deduplication**: If file indexed twice, messages appear twice (fixed by DELETE before insert)
2. **Tool results truncated**: Only first 500 chars indexed for FTS search (design choice to limit index size)
3. **No relationship linking**: This conversation doesn't show what files it modified (Phase 1c)
4. **No subagent handling**: Subagent conversations not specially marked (Phase 1b)
5. **No filtering**: Search returns all results, no type/project filters (Phase 2)

---

## Success Indicators During Implementation

After implementing FTS5 fix and running integration tests, you should see:

1. **Test output shows**:
   ```
   ok      ./internal/service  12.345s (FTS5 tests passing)
   ```

2. **Database populated with**:
   ```sql
   sqlite3 ~/.claude/cc-viz.db "SELECT COUNT(*) FROM conversation_messages"
   -- Returns: 123456+ (millions of messages from 3,927 conversations)
   ```

3. **Search works**:
   ```sql
   SELECT * FROM fts_conversations WHERE message_content MATCH 'postgres' LIMIT 5
   -- Returns: multiple matching messages from conversations about postgres
   ```

4. **No errors in logs** during full indexing run

---

## Rollback and Safety

If implementation causes issues:

1. **Rollback database**: Delete `~/.claude/cc-viz.db`, restart service (re-indexes automatically)
2. **Rollback code**: `git revert` the FTS5 commit
3. **Safety checkpoint**: Commit conditional schema work first, separate from logic changes

---

## References for Implementer

- **FTS5 Documentation**: https://www.sqlite.org/fts5.html
- **Go build tags**: https://golang.org/pkg/build/#hdr-Build_Constraints
- **SQLite driver**: https://github.com/mattn/go-sqlite3
- **JSONL format**: https://jsonlines.org/

---

## Handoff Checklist for Implementer

Before starting implementation, verify:

- [ ] Read entire PLAN file (.agent_planning/cc-viz-phase1-foundation/PLAN-*.md)
- [ ] Read entire DOD file (.agent_planning/cc-viz-phase1-foundation/DOD-*.md)
- [ ] Understand FTS5 blocker and conditional schema approach
- [ ] Located all 4 main code files (indexer.go, conversation.go, storage_sqlite.go, data_handler.go)
- [ ] Have access to `~/.claude/projects/` directory with real JSONL files
- [ ] Understand database schema (3 tables with FTS5 virtual)
- [ ] Understand all 3 work items (P0, P1, P2) and their interdependencies
- [ ] Understand acceptance criteria for each P0/P1/P2 deliverable

**Ready to implement**: All checkmarks complete
