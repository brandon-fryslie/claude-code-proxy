# EVALUATION: Unified Conversation Search with Filters

**Date:** 2026-01-12
**Status:** CONTINUE - Ready for Planning

---

## 1. WHAT EXISTS

### Frontend Search
- **ConversationSearch.tsx** - Basic text input with Cmd+K shortcut
- **search.ts** - `filterConversations()`, `filterMessages()`, `matchesSearchQuery()`
- **Current limitation:** Only searches within selected conversation, not across all

### Backend Search
- **FTS5 Database** - `conversations_fts` virtual table with porter stemmer
- **Tables:** `conversations` (metadata), `conversation_messages` (full data with agent_id)
- **API:** `GET /api/v2/conversations/search?q=<query>` - WORKING
- **Returns:** ConversationMatch (conversation-level summary, not message-level)

### Data Available
- `conversation_messages.type` - "user", "assistant", "file-history-snapshot", etc.
- `conversation_messages.agent_id` - Identifies subagent messages
- `conversation_messages.content_json` - Full message content

---

## 2. WHAT'S MISSING

1. **Global Search Component** - Search across ALL conversations simultaneously
2. **Filter UI** - Checkboxes for User/Claude/Subagents/Other
3. **Message-Level Results** - API returns conversation summary, needs message fragments
4. **Frontend Hook** - No `useConversationSearch()` in api.ts
5. **Result Display** - No component for search results with highlighted fragments

---

## 3. FILTER LOGIC

Based on database schema:
- **User** → `type = 'user'`
- **Claude** → `type = 'assistant' AND agent_id IS NULL`
- **Subagents** → `type = 'assistant' AND agent_id IS NOT NULL`
- **Other** → `type NOT IN ('user', 'assistant')`

---

## 4. FILES TO MODIFY

### Backend
- `proxy/internal/handler/data_handler.go` - Add types filter param, return message-level results
- `proxy/internal/service/storage_sqlite.go` - Enhanced SQL with message snippets
- `proxy/internal/model/models.go` - Add SearchMessage, update SearchResults

### Frontend
- `dashboard/src/lib/api.ts` - Add useConversationSearch() hook
- `dashboard/src/lib/types.ts` - Add SearchMessage, SearchFilters types
- `dashboard/src/pages/Conversations.tsx` - Integrate global search
- NEW: `dashboard/src/components/features/GlobalSearchResults.tsx`

---

## 5. RISKS

- **Performance** - FTS5 across 90K+ messages (mitigate: pagination, tested <10ms)
- **Filter Edge Cases** - agent_id null handling (verify indexer populates correctly)

---

## 6. VERDICT: CONTINUE

Infrastructure ready. Changes are additive. No blockers.
