# Clarifications Needed Before Phase 1b Implementation

**Status**: PAUSE - These 4 questions need answers before implementing Conversation Indexer

**Time to Decide**: ~30 minutes
**Impact**: High - Wrong decisions here mean rework later

---

## Question 1: Tool Call Indexing

### What's Being Asked
Should tool calls (Bash, Read, Write, etc.) be indexed as searchable items in full-text search?

### Context
Current JSONL messages contain tool_use blocks like:
```json
{
  "message": {
    "content": [
      { "type": "text", "text": "Here's the output" },
      { "type": "tool_use", "id": "tooluse_123", "name": "Bash", "input": {"command": "ls -la"} }
    ]
  }
}
```

When indexing for FTS5, we could:
- Index only the text content (simpler)
- Index text + tool names (enables searching by tool type)
- Index text + tool names + arguments (most complete)

### Options

**Option A: Text Only** ✅ Simple
- Index only `.message.content[].text` fields
- Benefits: Simple implementation, smaller index
- Tradeoffs: Can't search "show me all files this agent read" - user must search by filename only
- Implementation time: 2 hours

**Option B: Text + Tool Names** ⭐ Recommended
- Index text + tool names from `.message.content[].name` (Bash, Read, Write, etc)
- Benefits: Users can search for "Bash" to find all Bash executions, "Write" to find file edits
- Tradeoffs: Slightly larger index, more parsing needed
- Implementation time: 3 hours

**Option C: Full Extraction** 🏋️ Comprehensive
- Index text + tool names + key arguments (command from Bash, path from Read/Write)
- Benefits: Maximum search capability
- Tradeoffs: Larger index, complex parsing, storage bloat
- Implementation time: 5 hours

### Decision Needed
**Choose: A | B | C**

### Why This Matters for Phase 1b
Phase 1b (Subagent Graph) will also index tool calls by agent to show "which agent ran which tools most". The indexing pattern here determines efficiency there.

---

## Question 2: Message Content Truncation

### What's Being Asked
How much of each message's content should be stored in the FTS5 index?

### Context
Messages can be very large:
- Typical: 500-5,000 characters
- Large: 50,000-500,000 characters
- Extreme: 1-10 MB (rare, but happens with code outputs)

FTS5 stores the indexed content for search performance. Full content = full storage.

### Options

**Option A: Full Content** 🎯 Complete
- Index entire message content
- Benefits: No loss of searchability, can find anything
- Tradeoffs: Database grows large (1000 conversations × avg 100KB per conversation = 100MB+), slower FTS search
- Estimated DB size: 500MB-1GB depending on conversation age
- Implementation time: 2 hours (no truncation logic)

**Option B: Truncated - First 5KB** ⭐ Recommended
- Index only first 5,000 characters of each message
- Benefits: Balances search capability vs storage (reduces by ~90% in practice)
- Tradeoffs: Can't find content deep in long messages (but those rare)
- Estimated DB size: 50-100MB
- Implementation time: 2 hours (simple truncation)

**Option C: Configurable** 🔧 Flexible
- Add configuration parameter for truncation length
- Default: 5KB, configurable from 1KB to unlimited
- Benefits: Can tune per deployment
- Tradeoffs: More complex, needs config management
- Implementation time: 3 hours

### Decision Needed
**Choose: A | B | C**
If B or C: **Truncation Length? [1KB | 5KB | 10KB | 50KB | unlimited]**

### Why This Matters for Performance
Search queries on a 1GB FTS5 table are noticeably slower than 100MB. Phase 2 (unified search) will compound this.

---

## Question 3: Subagent Conversation Handling

### What's Being Asked
How should subagent "side-chain" conversations (parallel agents running in background) be indexed?

### Context
Directory structure:
```
~/.claude/projects/{project}/
├─ session-uuid.jsonl                 (Main agent conversation)
└─ subagents/
   ├─ agent-2-session-uuid.jsonl     (Side-chain from main)
   └─ agent-3-session-uuid.jsonl     (Another side-chain)
```

Each subagent has its own JSONL file with complete message history. They're parallel to the main agent, started and managed by the main agent.

Schema supports this: `conversation_messages.is_sidechain BOOLEAN`

### Options

**Option A: Separate Conversations** 🔀 Isolated
- Index each JSONL as separate conversation (current schema)
- Benefits: Simple, matches file structure, clear separation
- Tradeoffs: Can't correlate subagent work with main agent in searches, relationship linking (Phase 1c) must add relationships explicitly
- Data shown: User sees 3 separate conversations (main + 2 subagents)
- Implementation time: 3 hours

**Option B: Linked Conversations** 🔗 Correlated
- Index each JSONL as separate conversation
- Add `parent_conversation_id` foreign key in schema (needs migration)
- Benefits: Searches can find "show main agent + all spawned subagents", relationship linking easier
- Tradeoffs: Schema change needed now, more complex indexing
- Data shown: User sees main conversation, can expand to see subagent transcripts
- Implementation time: 5 hours

**Option C: Merged Conversations** 🔀 Combined
- Merge subagent messages into main conversation by timestamp
- Single conversation record with all messages interleaved
- Benefits: Unified search across agent + subagents, realistic workflow view
- Tradeoffs: Loses information about which agent sent which message (unless we track agent_id), complex merge logic
- Data shown: Single timeline with main + subagent messages interleaved
- Implementation time: 6 hours

### Decision Needed
**Choose: A | B | C**

### Why This Matters for Phase 1c
Phase 1c (Relationship Linker) needs to connect sessions to "all work spawned from them". Choice here determines relationship table design.

---

## Question 4: Re-indexing Strategy

### What's Being Asked
When a conversation JSONL file changes, how should the database be updated?

### Context
File watcher detects changes to `~/.claude/projects/{project}/{session}.jsonl`. Two strategies:

1. **Full Reset**: Delete all messages for that conversation, re-parse JSONL, re-insert all
2. **Append-Only**: Track file position, append new messages since last index

### Options

**Option A: Full Reset** 🔄 Safe
- When file changes: DELETE FROM conversation_messages WHERE conversation_id = ?
- Re-parse entire JSONL file
- Insert all messages fresh
- Benefits: Simple, guaranteed correct, no edge cases
- Tradeoffs: Slow if file has 10,000+ messages (re-index 10,000 on every change)
- Risk: FTS5 queries might see stale results during re-index (~1 second delay)
- Implementation time: 2 hours

**Option B: Append-Only** ⚡ Fast
- Track `last_read_position` for each conversation
- On file change: Seek to last position, read new lines only
- Insert only new messages
- Benefits: Fast (only new messages processed), efficient
- Tradeoffs: Risk if file edited mid-stream (line insertions/deletions), needs position tracking
- Risk: JSONL format guarantees append-only in practice, but Python/JS client could re-write file
- Implementation time: 5 hours (complex seek logic, edge case handling)

**Option C: Hybrid** 🎯 Balanced
- Detect file size change
- If change > 10% OR last_read > 1 hour ago: Full reset
- Otherwise: Append-only
- Benefits: Balances speed (most updates are appends) vs correctness (periodic reset catches issues)
- Tradeoffs: More complex logic
- Implementation time: 4 hours

### Decision Needed
**Choose: A | B | C**

### Why This Matters for User Experience
- Option A: 1-5 second lag when user continues typing mid-conversation (re-indexing)
- Option B: 100ms lag (fast), but risky
- Option C: Balances both

---

## Decision Template

Copy and fill this out:

```
CLARIFICATION DECISIONS
=======================

Question 1: Tool Call Indexing
Choice: [A | B | C]
Reasoning: [Why this choice]

Question 2: Message Truncation
Choice: [A | B | C]
Truncation Length (if B/C): [1KB | 5KB | 10KB | 50KB | unlimited]
Reasoning: [Why this choice]

Question 3: Subagent Handling
Choice: [A | B | C]
Reasoning: [Why this choice]

Question 4: Re-indexing Strategy
Choice: [A | B | C]
Reasoning: [Why this choice]

Next: Proceed with implementation using these decisions
```

---

## Impact Summary Table

| Decision | Phase 1 Impact | Phase 2 Impact | Phase 1c Impact | Storage Impact | Complexity |
|----------|---|---|---|---|---|
| Q1-A | Simple | Limited search | - | Baseline | Low |
| Q1-B | Medium | Good search | ✅ Enables | +10% | Medium |
| Q1-C | Complex | Excellent search | ✅✅ Enables | +30% | High |
| Q2-A | Simple | Slower | - | 10x size | Low |
| Q2-B | Simple | Good | - | 1x baseline | Low |
| Q2-C | Complex | Configurable | - | Depends | Medium |
| Q3-A | Simple | No correlation | Rework needed | - | Low |
| Q3-B | Medium | Partial correlation | Good | - | Medium |
| Q3-C | Complex | Full correlation | ✅ Works | - | High |
| Q4-A | Safe | Simple | - | 100% overhead | Low |
| Q4-B | Fast | Efficient | - | Low overhead | High |
| Q4-C | Balanced | Configurable | - | Depends | Medium |

---

## Recommended Defaults (If Unsure)

If you want to move fast and optimize later:
```
Q1: B (Text + Tool Names)        - Balanced search capability
Q2: B (5KB truncation)            - Good storage/search tradeoff
Q3: A (Separate Conversations)    - Simple now, can add relationships later
Q4: A (Full Reset)                - Safe, simple, acceptable performance
```

**Estimated outcome**: 100MB database, 3-hour indexing 3,927 JSONL files, sub-100ms searches

---

## Next Action

Once you provide these 4 decisions:
1. FTS5 blocker will be resolved (Option A: conditional schema)
2. Implementation can begin with full context
3. No ambiguity-driven rework needed later

**Ready when you are.**
