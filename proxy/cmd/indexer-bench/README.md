# Conversation Indexer Benchmark (P2)

This benchmark validates Phase 1 P2 acceptance criteria: performance at scale.

## Purpose

Validates that the conversation indexer can handle real-world data volumes:
- Indexes all JSONL files from `~/.claude/projects/` (typically 3,927+ files)
- Measures indexing duration, database size, search performance, and memory usage
- Validates against P2 acceptance criteria from the Definition of Done

## Building

The benchmark must be built with FTS5 support enabled:

```bash
CGO_ENABLED=1 go build -tags="fts5" -o bin/indexer-bench ./cmd/indexer-bench
```

## Running

### With Temporary Database (Default)

```bash
./bin/indexer-bench
```

Creates a temporary database, runs the benchmark, displays results, and cleans up.

### With Persistent Database

```bash
./bin/indexer-bench -db=/path/to/benchmark.db
```

Preserves the database after benchmarking for manual inspection.

### Verbose Logging

```bash
./bin/indexer-bench -v
```

Shows detailed indexing progress (file-by-file).

## P2 Acceptance Criteria

The benchmark validates:

1. **All files indexed successfully** - No errors during indexing
2. **Indexing completes in <10 minutes** - Performance target
3. **Database size <500 MB** - Storage efficiency
4. **Search queries <100ms** - FTS5 performance (if enabled)
5. **Memory usage stable** - No leaks during indexing
6. **No database locks** - Concurrent safety

## Example Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Conversation Indexer Performance Benchmark (P2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📁 Using database: /tmp/p2-benchmark.db
📂 Indexing directory: /Users/user/.claude/projects

📊 Baseline memory usage: 0.37 MB

🔍 Starting full indexing benchmark...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Benchmark Results
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏱️  Indexing Duration:     4m32.5s (4.54 min)
📁 Files Found:           3927
✅ Files Indexed:         3927
❌ Indexing Errors:       0

📊 Database Statistics:
   Conversations:         3500
   Messages:              125000
   FTS Entries:           125000
   Database Size:         385.42 MB

💾 Memory Usage:
   Baseline:              0.37 MB
   Peak:                  45.12 MB
   Delta:                 44.75 MB

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  P2 Acceptance Criteria Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ All files indexed successfully (no errors)
✅ Indexing completed in 4.54 minutes (target: <10 minutes)
✅ Database size 385.42 MB (target: <500 MB)
✅ Search query completed in 23ms (target: <100ms)
✅ Memory usage stable (44.75 MB increase)
✅ No database locks or concurrency issues detected during indexing

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✅ ALL P2 ACCEPTANCE CRITERIA PASSED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1 Foundation is validated and ready for production!
```

## Technical Details

### What Gets Measured

- **Indexing Duration**: Wall-clock time to walk and index all JSONL files
- **Files Found/Indexed**: Total .jsonl files discovered and successfully indexed
- **Error Count**: Number of files that failed to index
- **Database Statistics**: Row counts for conversations, messages, and FTS entries
- **Database Size**: On-disk size including WAL and SHM files
- **Memory Usage**: Go runtime allocated memory before and after indexing
- **Search Performance**: Time to execute a sample FTS5 query

### Implementation

The benchmark uses the public `RunFullIndexBenchmark()` method from `ConversationIndexer`:

```go
stats, err := indexer.RunFullIndexBenchmark()
```

This is a synchronous version of `initialIndex()` that:
1. Walks ~/.claude/projects/
2. Indexes every .jsonl file (ignoring staleness checks for full benchmark)
3. Collects timing and count statistics
4. Returns structured results

### Exit Codes

- `0` - All P2 criteria passed
- `1` - One or more criteria failed (details in output)

## Related Documentation

- **Definition of Done**: `.agent_planning/cc-viz-phase1-foundation/DOD-phase1-conversation-indexer-20260112.md`
- **Sprint Plan**: `.agent_planning/cc-viz-phase1-foundation/PLAN-phase1-conversation-indexer-20260112.md`
- **Indexer Implementation**: `internal/service/indexer.go`
