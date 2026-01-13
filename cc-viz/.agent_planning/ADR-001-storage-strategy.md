# ADR-001: Long-Term Storage Strategy for CC-VIZ

**Status**: Decided
**Date**: 2026-01-13
**Decision**: No optimization needed - let database grow naturally
**Revisit**: When database exceeds 10-20GB or performance degrades

---

## Context

During Phase 1 (Conversation Indexer) performance validation, we discovered:

- **Source data**: ~/.claude/projects/ contains 1.6GB of JSONL conversation files
- **Indexed database**: The SQLite database with FTS5 is 1.0GB
- **Combined storage**: ~2.6GB total
- **Content duplication**: Full message JSON is stored in both JSONL (source) and SQLite (for queries)

The question arose: **How do we prevent unbounded growth long-term?**

---

## User Requirements (Brandon's Input)

> "We're the only ones using this data"

- No need to optimize for distribution or other consumers
- Can make aggressive changes later without breaking compatibility

> "I want high fidelity long term data and we might think of new use cases once we have the data"

- Preservation > optimization
- Don't throw away data we might need for future features
- Raw data enables flexibility for unknown use cases

> "Storage is cheap - 1TB drive + 4TB external, 96GB RAM"

- Storage is not a real constraint on this system
- Performance headroom is abundant

> "Claude Code is auto-deleting the data (I've used it for ~6 months). So I'd like to keep data it's getting rid of"

- CC-VIZ serves as an **archive/backup** for data Claude Code discards
- This is a feature, not a bug - we're preserving history Claude Code throws away

---

## Options Analyzed

### Option A: Retention Window (Delete Old Data)
- **What**: Keep only last N days in database, delete older data entirely
- **Trade-off**: Lose ability to search/view old conversations
- **Bounded?**: Yes
- **Verdict**: ❌ Conflicts with "high fidelity long term data" requirement

### Option B: Index-Only (No Content Duplication)
- **What**: Database stores only metadata + FTS index, full content lives in JSONL
- **Trade-off**: Slightly slower content retrieval (read from disk instead of SQLite)
- **Bounded?**: Yes (database ~200-300MB), but JSONL still grows
- **Verdict**: ⚠️ Good option if storage becomes constrained

### Option C: Tiered Storage (Hot/Warm/Cold)
- **What**: Recent data in SQLite, older in compressed archive, oldest in cloud/deleted
- **Trade-off**: Complexity, requires multiple storage backends
- **Bounded?**: Yes at each tier
- **Verdict**: ❌ Over-engineered for single-user system

### Option D: Hybrid Hot/Cold Storage
- **What**: 7 days full content in SQLite (hot), older content references JSONL (cold)
- **Trade-off**: Slight latency for old content access
- **Implementation**:
  ```go
  // Hot data: content_json populated
  // Cold data: content_json = NULL, query falls back to JSONL
  func (s *Storage) GetMessageContent(uuid string) (string, error) {
      var content sql.NullString
      s.db.QueryRow("SELECT content_json FROM messages WHERE uuid = ?", uuid).Scan(&content)
      if content.Valid {
          return content.String, nil  // Hot - instant
      }
      return s.readFromJSONL(uuid)    // Cold - disk read
  }
  ```
- **Bounded?**: Yes (database ~150MB), JSONL still grows
- **Verdict**: ⚠️ Best balance if we need to optimize later

### Option E: Compression
- **What**: Compress content_json with zstd/gzip in SQLite
- **Trade-off**: CPU overhead on read/write
- **Bounded?**: No, just slower growth (~3-5x compression)
- **Verdict**: ❌ Doesn't solve the problem, just delays it

### Option F: Derived-Only Index (No Full Content)
- **What**: SQLite only stores derived/aggregated data:
  - Token usage stats per day
  - Tool usage counts
  - Session timelines
  - File relationship graph
- Full content always read from JSONL on-demand
- **Trade-off**: No full-text search without scanning JSONL
- **Bounded?**: Yes, very small (~10-50MB)
- **Verdict**: ⚠️ Good if search isn't critical

### Option G: External Search Engine
- **What**: Use Tantivy, Meilisearch, or Typesense for FTS
- SQLite stores relationships only, external engine handles search
- **Trade-off**: Another service to run, operational complexity
- **Bounded?**: Better compression, but still grows
- **Verdict**: ❌ Over-engineered for single-user

### Option H: On-Demand Indexing
- **What**: Don't index everything upfront. Index conversations when accessed.
- SQLite stores "stubs" (file paths + timestamps)
- Index on first access, LRU eviction of indexed content
- **Trade-off**: First access to old data is slow
- **Bounded?**: Yes, capped by LRU cache size
- **Verdict**: ⚠️ Good for sparse access patterns

### Option I: Semantic Summarization (Lossy Compression)
- **What**: For old data, replace full content with AI-generated summaries
- Keep: "Conversation about auth implementation, 5 files touched, 23 Bash calls"
- Discard: Full message text after 30 days
- **Trade-off**: Lossy - can't recover original text
- **Bounded?**: Yes, summaries are tiny
- **Verdict**: ❌ Conflicts with "high fidelity" requirement

### Option J: Follow Claude Code's Lead
- **What**: Match Claude Code's data lifecycle
- If Claude Code deletes after N days, we survive N days
- **Trade-off**: Dependent on upstream behavior
- **Verdict**: ❌ Claude Code is deleting data we want to keep!

---

## Decision

**No change for now. Let the database grow naturally.**

### Rationale

1. **Storage is not a constraint**: 1TB local + 4TB external makes even 50GB trivial
2. **Growth rate is manageable**: ~1GB/month → 12GB/year → decades before it matters
3. **Preservation > optimization**: Unknown future use cases need raw data
4. **CC-VIZ is the archive**: Claude Code deletes old data; we're the backup
5. **Performance is excellent**: 3.36ms search latency with 1GB database

### Current Metrics (Phase 1 Benchmark)

| Metric | Value |
|--------|-------|
| Source JSONL | 1.6 GB |
| SQLite Database | 1.0 GB |
| Files Indexed | 3,939 |
| Messages Indexed | 284,238 |
| FTS Entries | 250,742 |
| Search Latency | 3.36 ms |
| Index Time (full) | 4.89 min |

### Revisit Triggers

Re-evaluate this decision when:
- Database exceeds 10-20GB
- Search queries become noticeably slow (>500ms)
- Memory pressure becomes an issue
- CC-VIZ needs to be distributed to users with limited storage

---

## Future Options (If Needed)

If we need to optimize later, the recommended path is:

1. **First**: Option D (Hybrid Hot/Cold)
   - 7-day hot tier with full content in SQLite
   - Cold tier references JSONL files
   - Minimal code changes, transparent to query layer
   - Expected savings: 1GB → 150-200MB database

2. **Second**: Option E (Compression)
   - Add zstd compression to content_json column
   - Quick win, 3-5x size reduction
   - Can combine with Option D

3. **Last Resort**: Option B (Index-Only)
   - Remove content_json entirely from SQLite
   - FTS index + metadata only
   - All content reads go to JSONL
   - Maximum savings but query latency increases

---

## Implementation Notes

### If We Implement Option D Later

Schema change:
```sql
ALTER TABLE conversation_messages ADD COLUMN is_hot_tier BOOLEAN DEFAULT 1;
-- Cold data: SET content_json = NULL, is_hot_tier = 0
```

Query layer change:
```go
func (s *Storage) GetMessageContent(uuid string) (string, error) {
    var content sql.NullString
    var conversationID, filePath string

    err := s.db.QueryRow(`
        SELECT m.content_json, m.conversation_id, c.file_path
        FROM conversation_messages m
        JOIN conversations c ON m.conversation_id = c.id
        WHERE m.uuid = ?
    `, uuid).Scan(&content, &conversationID, &filePath)

    if err != nil {
        return "", err
    }

    if content.Valid {
        return content.String, nil  // Hot tier - instant
    }

    // Cold tier - read from JSONL
    return s.readMessageFromJSONL(filePath, uuid)
}
```

Archival job (runs daily):
```go
func (s *Storage) ArchiveColdData() error {
    cutoff := time.Now().AddDate(0, 0, -7)  // 7 days

    _, err := s.db.Exec(`
        UPDATE conversation_messages
        SET content_json = NULL, is_hot_tier = 0
        WHERE timestamp < ? AND is_hot_tier = 1
    `, cutoff.Format(time.RFC3339))

    return err
}
```

---

## References

- Phase 1 Benchmark Results: P2 validation showed 3.36ms search, 4.89min full index
- ROADMAP.md: Space Manager (Phase 3) includes cleanup features
- Claude Code behavior: Auto-deletes conversations older than ~30 days

---

## Changelog

- 2026-01-13: Initial decision - no optimization needed
