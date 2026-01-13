# Definition of Done: Phase 1 - Conversation Indexer Foundation

**Sprint**: Phase 1 Foundation
**Created**: 2026-01-12
**Target**: Complete before Phase 2 (Unified Search) begins

---

## P0: Fix FTS5 Blocker and Verify Core Indexing

### Acceptance Criteria

- [ ] `go test ./proxy/internal/service/...` runs with 0 failures (all 24 tests pass)
- [ ] FTS5 is disabled in test builds (using build tags) but enabled in production binary
- [ ] Production binary still includes FTS5 virtual table for full-text search capability
- [ ] Test environment uses fallback schema without FTS5 (all required columns present)
- [ ] Database initialization completes without errors in both test and production configurations
- [ ] Indexing can proceed from P0 to P1 without FTS5-related blockers

---

## P1: Implement Integration Tests with Real Data

### Acceptance Criteria

- [ ] New integration test in `indexer_test.go` successfully indexes 10+ real JSONL files
- [ ] Indexed messages appear in database (verify `conversation_messages` table is populated)
- [ ] Full-text search on indexed content returns correct results (FTS5 virtual table working)
- [ ] File watcher detects file modifications and triggers re-indexing
- [ ] Test cleanup ensures temporary database is deleted (no artifacts left)
- [ ] All new tests pass with both conditional and full schemas

---

## P2: Verify Performance at Scale

### Acceptance Criteria

- [ ] All 3,927+ real JSONL files from ~/.claude/projects/ indexed successfully without errors
- [ ] No indexing failures or data corruption during full scan
- [ ] Indexing completes in under 10 minutes (target: <5 minutes on modern hardware)
- [ ] Database file size is reasonable (<500 MB for full data volume)
- [ ] Full-text search queries complete in <100ms on common queries
- [ ] No database locks, deadlocks, or concurrency issues during indexing
- [ ] Memory usage remains stable during and after indexing (no leaks)

---

## Completion Requirements

**ALL of the following must be true to mark sprint as DONE:**

1. ✅ **FTS5 blocker resolved**: Test suite passes, build tags working
2. ✅ **Integration tests added**: Real data indexing verified
3. ✅ **Scale validated**: 3,927+ files indexed, performance acceptable
4. ✅ **Code committed**: All changes merged to main with proper commit messages
5. ✅ **No regressions**: Existing Session Data page still works
6. ✅ **Ready for handoff**: Phase 1c (Relationship Linker) can proceed with confidence

---

## Verification Checklist

Run these commands to verify completion:

```bash
# Test 1: All tests pass
go test ./proxy/internal/service/... -v

# Test 2: Production build succeeds
go build -o proxy-core ./proxy/cmd/proxy-core

# Test 3: FTS5 in production (should not error)
sqlite3 test.db "CREATE VIRTUAL TABLE test USING fts5(content)"

# Test 4: Real data indexing (if safe to run)
./proxy-core --config=config.yaml  # Monitor logs for successful indexing

# Test 5: Search functionality
curl -X POST http://localhost:8002/api/conversations/search \
  -H "Content-Type: application/json" \
  -d '{"query": "postgres migration"}'

# Test 6: Message count
sqlite3 ~/.claude/cc-viz.db "SELECT COUNT(*) FROM conversation_messages"
```

---

## Sign-Off

**Sprint Complete When**: User confirms ALL acceptance criteria met and tests pass

**Record**: Save user approval in `.agent_planning/cc-viz-phase1-foundation/USER-RESPONSE-<timestamp>.md`
