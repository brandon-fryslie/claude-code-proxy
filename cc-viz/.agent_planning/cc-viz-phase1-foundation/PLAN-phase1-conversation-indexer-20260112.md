# Sprint Plan: Phase 1 - Conversation Indexer Foundation

**Generated**: 2026-01-12
**Status**: Ready for Implementation
**Estimated Duration**: 8-12 hours

---

## Sprint Goal

Activate the Conversation Indexer to enable unified search and graph navigation across all Claude Code session data, unblocking Phase 2 (Search & Discovery).

---

## Current State Assessment

**Infrastructure Status**: 70% Ready
- Database schema: ✅ Complete and production-quality (3 tables with FTS5)
- Conversation parser: ✅ Complete (tested on 3,927+ real JSONL files)
- Indexer core logic: ⚠️ 90% implemented but needs verification
- API endpoints: ✅ Complete and ready for data
- React UI components: ✅ Ready for indexed data
- Test suite: ⚠️ Exists but failing (FTS5 test environment blocker)

**Key Discovery**: The implementation is 90% done in code but never fully tested or verified on real data. This sprint completes the final 10% and validates everything works at scale.

---

## Scope

### In Scope (This Sprint)

**P0: Fix FTS5 Blocker and Verify Core Indexing**
- Resolve FTS5 test environment issue (conditional build)
- Verify indexFile() method works with real JSONL data
- Validate message extraction and database population
- Ensure existing code actually functions as written

**P1: Add Integration Tests**
- Test indexing 100+ real JSONL files from ~/.claude/projects/
- Verify search functionality returns correct results
- Test file watcher detects changes and re-indexes

**P2: Verify Performance and Scale**
- Confirm all 3,927+ JSONL files can be indexed
- Measure indexing time and database size
- Validate search performance (<100ms queries)

### Explicitly Out of Scope (Future Sprints)

- Advanced query syntax (`type:conversation project:cc-viz` filters)
- Relationship graph linking to other entities (Phase 1c: Relationship Linker)
- Search UI components (Phase 2)
- Tool call aggregations or advanced analytics
- Subagent conversation handling (Phase 1b: Subagent Graph)

---

## Work Items

### P0: Fix FTS5 Blocker and Verify Indexing Core

**Problem**: All 24 tests fail with "no such module: fts5" - blocks confidence in implementation.

**Solution**: Implement conditional schema build using Go build tags.

**Technical Details**:
- File: `proxy/internal/service/storage_sqlite.go`
- Current schema creates FTS5 virtual table on line ~205-215
- Create alternate schema without FTS5 for test builds
- Use `//go:build !test` tags to switch schemas
- Ensures unit tests pass while preserving FTS5 in production binary

**Acceptance Criteria**:
- [ ] `go test ./internal/service/...` passes all 24 tests (FTS5 blocker resolved)
- [ ] Production binary still includes FTS5 virtual table for full-text search
- [ ] Test schema includes all required columns (conversations, conversation_messages tables)
- [ ] Database initialization runs without errors in both test and production modes

**Implementation Notes**:
- See `storage_sqlite.go` lines 195-220 for current FTS5 schema
- Estimated effort: 1-2 hours
- No changes needed to existing indexer logic
- Pattern: Use build tags similar to lines 1-10 in any Go test file

---

### P1: Implement Integration Tests with Real Data

**Problem**: Tests exist but don't validate against real JSONL files. No confidence in scale.

**Solution**: Add integration tests that index real conversation files.

**Technical Details**:
- Test file: `proxy/internal/service/indexer_test.go`
- Source data: `~/.claude/projects/` (3,927+ real JSONL files available)
- Use test fixtures or sample of real conversations
- Verify indexFile() actually populates database correctly

**Acceptance Criteria**:
- [ ] New test indexes 10+ real JSONL files without errors
- [ ] Database contains indexed messages from test files (verify message count, content)
- [ ] Search query returns messages from indexed files (test FTS5 search works)
- [ ] File watcher correctly detects file modifications and re-indexes

**Implementation Notes**:
- Reference: `proxy/internal/service/conversation_test.go` for pattern (parseConversationFile tests)
- Existing test data: Sample JSONL files in test fixtures or can generate from real ~/.claude/projects/
- Verify ExtractMessageContent() output appears in FTS5 index
- Test cleanup: Ensure temporary test database is deleted after tests

---

### P2: Verify Performance at Scale

**Problem**: Unknown if implementation handles real-world scale (3,927 files, 10,000+ messages).

**Solution**: Run indexer against full dataset and measure.

**Technical Details**:
- Run ConversationIndexer.initialIndex() against actual ~/.claude/projects/ directory
- Measure: indexing time, database size, query performance
- Verify no errors, memory usage reasonable

**Acceptance Criteria**:
- [ ] All 3,927+ JSONL files indexed successfully without errors
- [ ] Indexing completes in reasonable time (<10 minutes for full scan)
- [ ] Database file size is <500 MB (reasonable for data volume)
- [ ] Search queries return results in <100ms (FTS5 performance verified)
- [ ] No database locks or concurrency issues observed

**Implementation Notes**:
- Method: Temporary binary or test that calls indexer against real data
- Safety: Run on copy of ~/.claude or isolated test environment
- Measurement: Use time.Since(), os.Stat() for size
- Reference: `proxy/internal/service/storage_sqlite.go` line 42-45 (database stats)

---

## Dependencies

### Required Before This Sprint
- Go 1.21+ with SQLite driver (already present)
- Access to ~/.claude/projects/ directory with real JSONL files
- All existing code in place (`indexer.go`, `conversation.go`, `storage_sqlite.go`)

### Blocking This Sprint
- FTS5 conditional build (P0) - **must complete before P1/P2**
- Message extraction logic verification (P0 prerequisite)

### Unblocking for Phase 2
- Successful completion of P0, P1, P2 enables Unified Search implementation
- Phase 2 depends on: search queries work, performance acceptable, scale validated

---

## Technical Context

### Conversation Indexer Architecture

**File**: `proxy/internal/service/indexer.go` (408 lines)

**Components**:
1. `ConversationIndexer` struct (line ~45)
   - fields: `db *sql.DB`, `convService *ConversationService`, `queue chan string`, `watcher *fsnotify.Watcher`

2. `initialIndex()` method (line ~100)
   - Scans `~/.claude/projects/` for JSONL files
   - Calls indexFile() for each found

3. `indexFile()` method (line ~153)
   - Parses conversation using ConversationService.parseConversationFile()
   - Extracts messages via ExtractMessageContent()
   - Inserts into database with transactions
   - Deletes old entries before re-indexing (line ~204)

4. `watchFiles()` method (line ~250)
   - Uses fsnotify to detect file changes
   - Debounces rapid changes (timer at line ~316)
   - Queues file for re-indexing

5. `processIndexQueue()` method (line ~350)
   - Worker loop processes queued files
   - Handles errors without crashing

### Database Schema (storage_sqlite.go)

**Table 1**: `conversations` (9 columns)
```sql
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY,
    session_uuid TEXT UNIQUE,
    project_path TEXT,
    file_path TEXT,
    message_count INTEGER,
    total_tokens INTEGER,
    first_message_at TIMESTAMP,
    last_message_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

**Table 2**: `conversation_messages` (19 columns)
```sql
CREATE TABLE conversation_messages (
    id INTEGER PRIMARY KEY,
    conversation_id INTEGER FOREIGN KEY,
    message_type TEXT,
    role TEXT,
    content TEXT,
    tool_name TEXT,
    tool_use_id TEXT,
    input_tokens INTEGER,
    output_tokens INTEGER,
    created_at TIMESTAMP,
    sequence_number INTEGER,
    is_tool_result BOOLEAN,
    FOREIGN KEY(conversation_id) REFERENCES conversations(id)
)
```

**Table 3**: `fts_conversations` (FTS5 virtual table)
```sql
CREATE VIRTUAL TABLE fts_conversations USING fts5(
    conversation_id,
    message_content,
    tool_names,
    content='conversation_messages',
    content_rowid='id'
)
```

### API Endpoints (data_handler.go)

**Endpoint 1**: `GET /api/conversations`
- Returns list of all conversations
- Parameters: limit, offset (for pagination)
- Response: array of conversation summaries

**Endpoint 2**: `GET /api/conversations/{id}/messages`
- Returns all messages for a conversation
- Parameters: limit, offset
- Response: array of messages with details

**Endpoint 3**: `POST /api/conversations/search`
- Full-text search across indexed content
- Body: `{"query": "search text"}`
- Response: array of messages + conversation context

**Endpoint 4**: `POST /api/conversations/reindex`
- Manually trigger full re-index
- Body: optional `{"path": "/specific/project"}`
- Response: status and file count processed

### Message Extraction (conversation.go)

**Function**: `ExtractMessageContent()` (line ~326)

**Inputs**: ConversationMessage struct with fields:
- `Type` (text, image, tool_use, tool_result)
- `Content` (string or array)
- `Name` (tool name if tool_use/tool_result)

**Output**: Combined string for FTS5 index
```
"tool_name: Bash | command: go test ./..."
"text: Here's the implementation..."
```

**Behavior**:
- Tool results truncated at 500 chars (line ~375)
- Tool names and text combined for search
- Empty content safely handled (line ~380)

### Configuration (config.yaml pattern)

**Indexer configuration** (loaded at startup):
```yaml
indexer:
  enabled: true
  watch: true              # Enable file watcher
  rescan_interval: 3600   # Re-scan all files daily
  debounce_ms: 500        # Group rapid file changes
```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| FTS5 blocker persists | Blocks all testing | Implement conditional schema (P0 priority) |
| Real data indexing fails | Scale unknown | Add error handling, test with subset first |
| Search performance poor | Phase 2 blocked | Verify index creation (line ~210 in storage) |
| Concurrency issues | Data corruption | Test rapid file changes with watcher |
| Database size explosion | Storage concerns | Set reasonable bounds, implement archival (Phase 3) |

---

## Success Metrics

After sprint completion, verify:

1. **All tests pass**
   - `go test ./internal/service/...` returns 24/24 passed
   - No FTS5-related failures

2. **Real data indexed successfully**
   - All 3,927+ JSONL files indexed
   - No errors in indexing log

3. **Search works**
   - FTS5 queries return relevant results
   - Queries complete in <100ms
   - Tool name search works (e.g., "Bash" returns bash tool calls)

4. **Scale validated**
   - Indexing completes in <10 minutes
   - Database <500 MB
   - Memory usage stable during indexing

5. **Ready for Phase 2**
   - Graph Navigator (phase-1c) can proceed with verified data
   - Unified Search (phase-2) has reliable backend

---

## Next Actions After Completion

1. **Commit to main**: All FTS5 fixes and new tests committed
2. **Create phase-1c plan**: Subagent Graph and Relationship Linker
3. **Unblock Phase 2**: Unified Search can now begin
4. **Monitor in production**: Track indexing performance over time

---

## References

- Vision: `.agent_planning/FEATURE_PROPOSAL_cc-viz-ultimate-vision.md`
- Handoff: `.agent_planning/HANDOFF-cc-viz-ultimate-vision-20260112.md`
- Roadmap: `.agent_planning/ROADMAP.md`
- Evaluation: `EVALUATION-phase1-foundation-20260112.md` (in this directory)
