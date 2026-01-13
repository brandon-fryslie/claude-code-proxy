# Relevant Files - Phase 1: Graph Foundation (Conversation Indexer)

## Critical Path Files

### Backend: Indexing Infrastructure
**Status: 60% Complete - Skeleton exists, implementation missing**

```
proxy/internal/service/indexer.go (208 lines)
├─ ConversationIndexer struct ✅
├─ Start() / Stop() lifecycle ✅
├─ initialIndex() walk implementation ✅
├─ needsIndexing() file change detection ✅
├─ indexFile() ❌ STUB - NEEDS IMPLEMENTATION
├─ processIndexQueue() ❌ STUB - needs worker logic
└─ watchFiles() ❌ STUB - needs file watcher activation

proxy/internal/service/conversation.go (310 lines)
├─ ConversationMessage struct ✅
├─ MessageContent, MessageUsage ✅
├─ Conversation struct ✅
├─ parseConversationFile() ✅ COMPLETE
├─ GetConversations() ✅ works but expensive
├─ GetConversation(projectPath, sessionID) ✅
└─ GetConversationsByProject() ✅

proxy/internal/service/storage_sqlite.go (600+ lines)
├─ CREATE TABLE conversations ✅
├─ CREATE TABLE conversation_messages ✅
├─ CREATE VIRTUAL TABLE conversations_fts ✅
├─ Indexes (11 total) ✅
├─ runConversationSearchMigrations() ✅
├─ SearchConversations() query ✅ (tested but untestable)
└─ GetConversationMessages() ✅
```

**Key Methods Missing**:
- `indexFile(path string) error` - Should: parse JSONL, extract messages, insert to DB
- `processIndexQueue()` - Should: dequeue paths, call indexFile, handle errors
- `watchFiles()` - Should: watch ~/.claude/projects/, queue changed files with debounce

### Backend: Database & Migrations
**Status: 100% Complete - Schema ready for use**

```
proxy/internal/service/storage_sqlite.go
├─ conversations (metadata table, 9 columns)
│  └─ Indexes: project_path, mtime, indexed_at
├─ conversation_messages (detail table, 19 columns)
│  └─ Indexes: conversation_id, timestamp, parent, session, agent, request
├─ conversations_fts (full-text search virtual table)
│  └─ Tokenizer: porter + unicode61
│  └─ Fields: message_id, content_text, tool_names
└─ Foreign key constraints
```

**Note**: FTS5 requires CGO with SQLite built with FTS5 enabled - see blocker in evaluation.

### Backend: API Handlers
**Status: 100% Complete - Handlers written and tested**

```
proxy/internal/handler/data_handler.go
├─ GetConversations(w, r) - List all conversations
├─ GetConversationByID(w, r) - Get specific conversation
├─ GetConversationsByProject(w, r) - Filter by project
├─ SearchConversations(w, r) - FTS5 search (POST /api/conversations/search?q=...)
└─ ReindexConversations(w, r) - Trigger full re-index (POST /api/conversations/reindex)
```

**Endpoints**:
- `GET /api/conversations` - Returns all conversations
- `GET /api/conversations/{id}` - Returns conversation with all messages
- `GET /api/conversations?project=<path>` - Filter by project
- `GET /api/conversations/search?q=<query>` - Full-text search with pagination
- `POST /api/conversations/reindex` - Trigger full database re-index

### Backend: Data Models
**Status: 100% Complete - Structs defined**

```
proxy/internal/model/models.go
├─ IndexedConversation - Metadata for listing
├─ DBConversationMessage - Message detail for display
├─ ConversationMessagesResponse - API response wrapper
└─ SearchResults - Search response with pagination
```

### Frontend: React Components
**Status: 90% Complete - UI ready, waiting on indexing**

```
cc-viz/src/components/features/
├─ ConversationList.tsx (160 lines) - Lists conversations
├─ ConversationThread.tsx (200 lines) - Displays messages
├─ ConversationSearch.tsx (180 lines) - Search interface
└─ layout/AppLayout.tsx - Page layout wrapper

cc-viz/src/pages/
├─ Conversations.tsx (220 lines) - Main conversation page
│  └─ Uses virtualization for performance
│  └─ Lazy loads conversation details
│  └─ Shows file format vs database format toggle
└─ SessionData.tsx (180 lines) - Shows todos/plans (Phase 1a complete)
```

**Key Features Ready**:
- Virtualized conversation list (handles 1000+ conversations)
- Message thread viewer with formatting
- Full-text search UI with result previews
- Toggle between file format and DB format

### Frontend: API Integration
**Status: 95% Complete - Hooks written**

```
cc-viz/src/lib/api.ts
├─ useConversations() - List all conversations
├─ useConversationDetail(id) - Get specific conversation
├─ useConversationMessages(id, opts) - Messages with subagent merge
├─ useConversationSearch(query) - FTS5 search
└─ reindexConversations() - Manual refresh
```

---

## Test Files

### Test Status: BROKEN (FTS5 Module Missing)

```
proxy/internal/service/
├─ indexer_test.go (80 lines)
│  └─ TestConversationIndexer - FAIL (no such module: fts5)
│  └─ TestNeedsIndexing - FAIL (no such module: fts5)
│
├─ storage_sqlite_test.go (300+ lines)
│  └─ TestSaveRequest_NewFields - FAIL
│  └─ TestUpdateRequestWithResponse_TokensAndTiming - FAIL
│  └─ TestMigration_ExistingDatabase - FAIL
│  └─ 14 more storage tests - all FAIL
│
├─ data_handler_test.go (400+ lines)
│  └─ TestSearchConversations_EmptyQuery - FAIL
│  └─ TestSearchConversations_NoResults - FAIL
│  └─ TestSearchConversations_Pagination - FAIL
│  └─ TestSearchConversations_ResponseFormat - FAIL
│  └─ TestSearchConversations_HTTPStatusCodes - FAIL
│  └─ TestSearchConversations_ProjectFilter - FAIL
│
└─ search_test.go (300+ lines)
   └─ TestSearchConversations - FAIL
   └─ TestSearchConversationsResponseFormat - FAIL
```

**Resolution Needed**: See EVALUATION-phase1-foundation-20260112.md Section 3 for FTS5 blocker options.

---

## Configuration & Setup

```
proxy/internal/config/config.go
├─ StorageConfig struct
├─ DBPath defaults to "proxy.db"
└─ SQLite WAL mode enabled (✅ good for concurrency)
```

---

## Source Data

### Real Data Locations
```
~/.claude/projects/                          (21+ project directories)
├─ -Users-bmf--claude-plugins/               (example)
│  ├─ session-uuid-1.jsonl                   (1-100+ messages)
│  ├─ session-uuid-2.jsonl
│  └─ subagents/                             (optional subagent conversations)
│     ├─ subagent-id-1.jsonl
│     └─ subagent-id-2.jsonl
├─ -Users-bmf-code-brandon-fryslie.../
│  └─ (many more project directories)
...

Total: 3,927 JSONL files (as of 2026-01-12)
```

### Database Storage
```
proxy/proxy.db                                (SQLite database)
├─ conversations table (metadata)
├─ conversation_messages table (details)
├─ conversations_fts table (searchable index)
└─ Indexes + foreign keys
```

---

## Dependencies & Relationships

### Phase 1 Dependencies
```
Session Data Indexer (✅ COMPLETE)
    ↓
Conversation Indexer (🔄 IN PROGRESS - 60% done)
    ↓
├─ Subagent Graph (💡 PROPOSED - blocked)
│   ↓
├─ Relationship Linker (💡 PROPOSED - blocked)
│   ↓
Unified Search (Phase 2)
```

### External Dependencies
- SQLite (go-sqlite3 binding)
- fsnotify (file watcher)
- React 19.2
- TanStack React Router & React Query
- Tailwind CSS

---

## Critical Implementation Gaps

### Gap 1: indexFile() - Core Indexing Logic
**Location**: `proxy/internal/service/indexer.go`
**Current**: Stub method signature only
**Needed**:
1. Open JSONL file from path
2. For each line:
   - Parse ConversationMessage JSON
   - Extract message.content[].text → searchable_text
   - Extract tool names from message.content[].name
   - Insert into conversation_messages table
   - Insert into conversations_fts table
3. Update indexed_at timestamp
4. Handle errors (collect, don't fail entire indexing)

**Estimated LOC**: 60-80

### Gap 2: processIndexQueue() - Queue Worker
**Location**: `proxy/internal/service/indexer.go`
**Current**: Stub goroutine that does nothing
**Needed**:
1. For each path in ci.indexQueue:
   - Call ci.indexFile(path)
   - Log results
   - Handle and collect errors
2. Handle shutdown signal (ci.done channel)

**Estimated LOC**: 20-30

### Gap 3: watchFiles() - File Monitoring
**Location**: `proxy/internal/service/indexer.go`
**Current**: Stub method
**Needed**:
1. Add ci.claudeProjects to watcher
2. Watch for WRITE and CREATE events
3. Debounce rapid changes (use ci.debounceTimers)
4. Queue files for re-indexing
5. Handle shutdown (ci.done channel)

**Estimated LOC**: 40-50

### Gap 4: Message Content Extraction - Parser Helper
**Location**: `proxy/internal/service/indexer.go` or new file
**Needed**: Function to:
1. Parse message.content JSON (nested structure)
2. Extract text from text blocks
3. Extract tool names from tool_use blocks
4. Combine for FTS index

**Estimated LOC**: 30-40

**Ambiguity**: Should tool names be indexed separately or combined with text?

---

## Quick Reference

### Endpoints to Test
```bash
# List conversations
curl http://localhost:8002/api/conversations

# Get specific conversation
curl http://localhost:8002/api/conversations/{id}

# Search conversations
curl http://localhost:8002/api/conversations/search?q=test

# Trigger re-index
curl -X POST http://localhost:8002/api/conversations/reindex
```

### Build & Test Commands
```bash
# Build Go backend
cd proxy && go build ./cmd/proxy-data

# Run tests (currently failing)
cd proxy && go test ./internal/service/...

# Build frontend
cd cc-viz && npm run build

# Run frontend dev
cd cc-viz && npm run dev
```

### Database Inspection
```bash
# Check if conversations table exists
sqlite3 proxy/proxy.db ".tables" | grep conversations

# Count indexed conversations
sqlite3 proxy/proxy.db "SELECT COUNT(*) FROM conversations"

# Check FTS5 status
sqlite3 proxy/proxy.db "SELECT name FROM sqlite_master WHERE type='table' AND name='conversations_fts'"
```

---

## Decision Points Before Implementation

These must be resolved before writing implementation:

1. **Tool Call Extraction** (Section 7, Ambiguity A in evaluation)
   - Include tool names as searchable items?

2. **Content Truncation** (Section 7, Ambiguity B in evaluation)
   - Full message content or summaries?

3. **Subagent Handling** (Section 7, Ambiguity C in evaluation)
   - Separate index for side-chains?

4. **Re-index Strategy** (Section 7, Ambiguity D in evaluation)
   - Append-only or full re-index?

---

## Status Summary

| Component | Status | LOC | Tests | Blocker |
|-----------|--------|-----|-------|---------|
| Schema | ✅ Done | - | FAIL | FTS5 |
| Parser | ✅ Done | 310 | N/A | - |
| Indexer | 60% | 208 | FAIL | FTS5 |
| API | ✅ Done | - | FAIL | FTS5 |
| UI | 90% | 700+ | N/A | Indexing |
| Tests | Broken | 1000+ | FAIL | FTS5 |

---

## Next Session Checklist

- [ ] Resolve FTS5 blocker (Option A recommended)
- [ ] Clarify 4 ambiguities
- [ ] Implement indexFile()
- [ ] Implement processIndexQueue()
- [ ] Implement watchFiles()
- [ ] Add message extraction helper
- [ ] Run on real JSONL data (3,927 files)
- [ ] Verify database populated
- [ ] Test search functionality
- [ ] Add integration tests for large files
- [ ] Measure indexing performance
