# Handoff: Conversation Database Ingestion

**Created**: 2026-01-12T13:15:00Z
**For**: Any agent continuing conversation browsing feature development
**Status**: in-progress

---

## Objective

Store full conversation data from `~/.claude/projects/` JSONL files into SQLite database for fast, coherent browsing of conversations and their messages, including subagent conversations.

## Current State

### What's Been Done
- Added `conversation_messages` table to store full message data (uuid, content, tokens, model, etc.)
- Updated indexer to store complete message content (not just FTS text)
- Added `GET /api/v2/conversations/{id}/messages` endpoint for paginated message retrieval
- Added `POST /api/v2/conversations/reindex` endpoint to trigger re-indexing
- Added `GetConversationFilePath` for fast single-conversation lookup (fixed 16s → 34ms performance issue)
- Fixed JSON field name mismatch (PascalCase → snake_case) for TypeScript compatibility
- Added contract tests for JSON serialization

### What's In Progress
- Subagent messages are stored but in **separate conversation records**
- Need to merge parent + subagent messages into unified timeline view

### What Remains
- [ ] Add `?include_subagents=true` param to merge subagent messages with parent
- [ ] Link conversations to proxy requests (by timestamp correlation or requestId)
- [ ] Update frontend to display merged conversation+subagent timeline
- [ ] Add conversation search/filter by project, date range, model used

## Context & Background

### Why We're Doing This
The `~/.claude/projects/` directory contains valuable conversation history (JSONL files) that was previously only accessible by reading files directly. This was slow (16+ seconds) and didn't allow searching or filtering. By storing in SQLite:
1. Fast retrieval (<50ms)
2. Full-text search via FTS5
3. Queryable metadata (model, tokens, timestamps)
4. Can correlate with proxy request data

### Key Decisions Made
| Decision | Rationale | Date |
|----------|-----------|------|
| Store in SQLite, not separate DB | Reuse existing requests.db infrastructure | 2026-01-12 |
| Use INSERT OR REPLACE for messages | Handle duplicate UUIDs in JSONL files | 2026-01-12 |
| Keep subagents as separate conversations initially | Simpler indexing, merge at query time | 2026-01-12 |
| Use snake_case JSON field names | Match TypeScript frontend conventions | 2026-01-12 |

### Important Constraints
- Must mount `~/.claude/projects` into container (already configured in docker-compose.yml)
- JSONL files can be large (10MB+ buffer needed for scanner)
- Some files have duplicate UUIDs (file-history-snapshot entries)
- Subagent conversations reference parent via `sessionId` field

## Acceptance Criteria

How we'll know this is complete:

- [x] Conversations indexed from ~/.claude/projects into DB
- [x] Messages queryable via API with pagination
- [x] Token usage and model info stored per message
- [ ] Subagent messages mergeable with parent timeline
- [ ] Conversations linkable to proxy requests
- [ ] Frontend displays coherent conversation browser

## Scope

### Files Modified (uncommitted)
- `proxy/internal/service/storage_sqlite.go` - Added `conversation_messages` table, `GetConversationMessages`, `ReindexConversations`
- `proxy/internal/service/storage.go` - Added interface methods
- `proxy/internal/service/indexer.go` - Store full message data, INSERT OR REPLACE
- `proxy/internal/service/conversation.go` - Added `MessageContent`, `MessageUsage` types, more fields on `ConversationMessage`
- `proxy/internal/model/models.go` - Added `DBConversationMessage`, `ConversationMessagesResponse`
- `proxy/internal/handler/data_handler.go` - Added `GetConversationMessagesV2`, `ReindexConversationsV2`
- `proxy/cmd/proxy-data/main.go` - Added routes for new endpoints
- `proxy/internal/config/config.go` - Added JSON tags to all config structs
- `proxy/internal/config/config_test.go` - NEW: Contract tests for JSON serialization

### Related Components
- `docker-compose.yml:93` - Volume mount for ~/.claude/projects
- `dashboard/` - Frontend that will consume these APIs

### Out of Scope
- Storing actual file diffs/snapshots from file-history-snapshot messages
- Real-time streaming of conversation updates
- Cross-project conversation search (each project indexed separately)

## Implementation Approach

### Recommended Next Steps
1. **Add subagent merging** - Query conversations where `sessionId` matches parent, merge by timestamp
2. **Add request linking** - Messages have `requestId` field; match to `requests` table
3. **Update frontend** - Show conversations with expandable subagent threads

### Database Schema (conversation_messages)
```sql
CREATE TABLE conversation_messages (
    uuid TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    parent_uuid TEXT,
    type TEXT NOT NULL,           -- user, assistant, file-history-snapshot
    role TEXT,                    -- user, assistant
    timestamp DATETIME NOT NULL,
    cwd TEXT,
    git_branch TEXT,
    session_id TEXT,              -- Parent conversation ID (for subagents)
    agent_id TEXT,                -- Subagent ID (e.g., "ac0c8d2")
    is_sidechain BOOLEAN,         -- true for subagent messages
    request_id TEXT,              -- Links to requests table
    model TEXT,                   -- e.g., "claude-opus-4-5-20251101"
    input_tokens INTEGER,
    output_tokens INTEGER,
    cache_read_tokens INTEGER,
    cache_creation_tokens INTEGER,
    content_json TEXT,            -- Full message content as JSON
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
```

### API Endpoints Added
- `GET /api/v2/conversations/{id}/messages?limit=100&offset=0` - Paginated messages
- `POST /api/v2/conversations/reindex` - Trigger re-index

### Patterns to Follow
- Use `writeJSONResponse()` for all API responses
- Use `sql.NullString` for nullable DB fields
- Parse timestamps with `time.RFC3339`

### Known Gotchas
- Some JSONL files have messages with empty UUIDs (file-history-snapshot)
- Subagent conversations are in `<project>/<parent-session>/subagents/` directories
- Token counts are nested inside `message.usage` in the JSONL
- Container must be restarted to pick up new image (podman-compose doesn't auto-update)

## Reference Materials

### Planning Documents
- [HANDOFF-json-field-mismatch-20260112.md](.agent_planning/HANDOFF-json-field-mismatch-20260112.md) - JSON field fix (COMPLETED)
- [STATUS-20260112.md](.agent_planning/web-routing-configuration/STATUS-20260112.md) - Web routing investigation

### Codebase References
- `proxy/internal/service/indexer.go:152-310` - Main indexing logic
- `proxy/internal/service/storage_sqlite.go:207-247` - Schema creation
- `proxy/internal/service/storage_sqlite.go:1496-1607` - Query methods
- `proxy/internal/handler/data_handler.go:726-780` - API handlers

### Data Format (JSONL message structure)
```json
{
  "parentUuid": "uuid-of-parent",
  "isSidechain": true,
  "cwd": "/Users/bmf/code/project",
  "sessionId": "parent-conversation-uuid",
  "gitBranch": "main",
  "agentId": "ac0c8d2",
  "requestId": "req_abc123",
  "type": "assistant",
  "message": {
    "role": "assistant",
    "model": "claude-opus-4-5-20251101",
    "content": [...],
    "usage": {
      "input_tokens": 1234,
      "output_tokens": 567
    }
  },
  "uuid": "message-uuid",
  "timestamp": "2026-01-12T10:54:20.690Z"
}
```

## Questions & Blockers

### Open Questions
- [ ] How to handle very long conversations (1000+ messages) in UI?
- [ ] Should subagent messages be collapsed by default in timeline view?
- [ ] Should we store raw content_json or extract structured fields?

### Current Blockers
- None - ready to continue implementation

### Need User Input On
- Preferred UX for viewing subagent conversations (inline vs expandable vs separate view)

## Testing Strategy

### Existing Tests
- `proxy/internal/config/config_test.go` - Contract tests for JSON serialization
- All existing tests pass (`just test`)

### New Tests Needed
- [ ] Test `GetConversationMessages` with pagination
- [ ] Test subagent message merging (when implemented)
- [ ] Test re-index endpoint

### Manual Testing
```bash
# List conversations (should be ~100)
curl -s http://localhost:8000/api/v2/conversations | jq 'length'

# Get messages for a conversation
curl -s "http://localhost:8000/api/v2/conversations/<id>/messages?limit=10" | jq '.messages | length'

# Check subagent messages
curl -s "http://localhost:8000/api/v2/conversations/agent-<id>/messages?limit=5" | jq '.messages[] | {type, model, sessionId, agentId}'

# Trigger re-index
curl -X POST http://localhost:8000/api/v2/conversations/reindex
```

## Success Metrics

How to validate implementation:

- [x] All existing tests pass (`just test`)
- [x] Conversations load in <50ms (was 16s)
- [x] Messages include model, token counts
- [ ] Subagent messages accessible via parent conversation
- [ ] Frontend shows coherent timeline

---

## Next Steps for Agent

**Immediate actions**:
1. Commit current changes (significant uncommitted work)
2. Implement `?include_subagents=true` parameter for message merging
3. Add frontend component for conversation browsing

**Before starting implementation**:
- [ ] Review this handoff document
- [ ] Run `just docker` to start services
- [ ] Verify APIs work with curl commands above

**When complete**:
- [ ] Update this handoff with completed items
- [ ] Commit changes
- [ ] Update CLAUDE.md if new patterns established
