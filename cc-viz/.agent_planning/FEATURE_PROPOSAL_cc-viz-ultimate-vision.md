# CC-VIZ Ultimate Vision: The Claude Code Data Cockpit

## The Problem

Claude Code generates a wealth of interconnected data across dozens of directories and file types, but this data sits in isolation - invisible, unexplorable, and disconnected. Users cannot:

- **See the full picture**: How does a conversation connect to the files it modified? What todos emerged from which sessions? Which projects are most active?
- **Navigate relationships**: Following a thread from prompt to tool call to file change to subagent spawn requires opening multiple files and correlating UUIDs manually
- **Search holistically**: Finding "that conversation where I fixed the authentication bug" means grepping through massive JSONL files
- **Understand their own patterns**: How do they use Claude Code? What tools do they use most? When are they most productive?
- **Manage their data**: Growing directories, failed telemetry, orphaned files - no visibility into what's consuming space or why

The data exists. The connections are there. But without visualization and navigation, it's like having a library with no catalog, no shelves, and no lights.

## The Vision

CC-VIZ becomes the **Mission Control** for Claude Code - a data cockpit where every piece of information is visible, connected, and actionable. When you open it, you see your entire Claude Code universe: active sessions pulsing, recent conversations flowing, file changes rippling across projects, subagents spinning off and reporting back.

**Imagine this experience:**

You click on a conversation from three days ago. Instantly you see not just the transcript, but a timeline showing every file it touched, every tool it called, every subagent it spawned. Click a file change and see the exact diff. Click a subagent and dive into its conversation. Click a plan it created and see how that plan evolved across sessions.

You search for "database migration" and results span conversations, plans, file changes, and even the prompts you typed. Each result shows its connections - this plan was created during that conversation, which also modified these files.

You hover over the activity heatmap and see January 3rd was your most productive day - 96 sessions, 12,000 messages. Click and see what you were working on. Compare it to today.

You notice a project with 500MB of data. Drill down: 400 sessions, 50 subagent chains, 10,000 file history snapshots. One-click cleanup options surface: archive old sessions, compress snapshots, export before deletion.

This is not a dashboard. This is a cockpit.

## Selected Ideas

### Idea 1: The Graph Navigator - Universal Data Interconnection

**User Story**: As a Claude Code user, I want to see and navigate all connections between my data so that I can understand context and follow any thread of work.

**The Experience**:

You're looking at a conversation about refactoring the authentication system. On the right side, a relationship panel shows:

- **12 files modified** - collapsed view showing filenames, expandable to show each change
- **3 subagents spawned** - each with their own conversation, linked back to the parent
- **2 plans referenced** - the plan that guided this work, and a new plan it created
- **1 todo list** - showing tasks that were in-progress during this conversation
- **8 tool calls** - Read, Write, Bash, etc. with quick preview on hover

Click any connection and you're there. The navigation is bidirectional - from a file's perspective, you see all conversations that modified it. From a plan's perspective, you see all sessions that referenced or updated it.

The underlying graph is queryable: "Show me all conversations that modified files in /src/auth/ in the last week" returns a filtered view.

**Why This Matters**: Currently, correlation is manual. You know a conversation happened but don't know what it touched. You know a file changed but not why. The graph navigator makes implicit connections explicit and traversable.

**Success Looks Like**:
- Any data entity shows its connections to other entities
- Clicking a connection navigates to that entity with context preserved
- Graph queries return meaningful filtered views
- Zero dead ends - every entity has paths to related entities

---

### Idea 2: The Unified Search Engine

**User Story**: As a power user, I want to search across ALL my Claude Code data with a single query so that I can find anything regardless of where it's stored.

**The Experience**:

A search bar sits at the top of every page. Type "postgres migration" and results stream in from everywhere:

```
Conversations (3 matches)
  [Jan 11] "Implementing database migration system" - 47 messages
    "...let's create a postgres migration that handles..."

Plans (1 match)
  peppy-yawning-teapot.md
    "## Database Migration Strategy\n..."

History (8 matches)
  "run the postgres migration"
  "check migration status"
  ...

File Changes (12 matches)
  migrations/001_initial.sql - created Jan 11
  src/db/migrate.go - modified 3 times

Todos (2 matches)
  [completed] "Implement postgres migration runner"
  [completed] "Test migration rollback"
```

Each result shows:
- Context: where it came from, when
- Snippet: the matching text with highlights
- Connections: what other entities link to this result

Filters narrow results: time range, data type, project, status. Saved searches persist for common queries.

Advanced syntax supports power users:
- `type:conversation project:cc-viz "token usage"` - conversations in cc-viz mentioning token usage
- `modified:>2026-01-01 tool:Bash` - recent sessions with Bash tool usage
- `status:in_progress` - active todos across all sessions

**Why This Matters**: Right now, finding something means knowing where it is. History is separate from conversations, which are separate from plans, which are separate from file changes. Unified search treats all data as one corpus.

**Success Looks Like**:
- Single search returns results from all data sources
- Results are ranked by relevance and recency
- Each result shows enough context to recognize it
- Clicking a result navigates to the full entity with search context highlighted

---

### Idea 3: The Activity Intelligence Dashboard

**User Story**: As a data-curious user, I want to understand my Claude Code usage patterns so that I can optimize my workflow and see my productivity trends.

**The Experience**:

The home page shows an intelligence dashboard that answers: "What's happening in my Claude Code universe?"

**Top Section - Live Status**:
- Currently active sessions (if any)
- Most recent conversation with one-click resume context
- Unfinished todos across all sessions
- In-progress plans

**Middle Section - Activity Patterns**:
```
This Week vs Last Week
+23% messages | +15% sessions | -8% tool calls

Most Active Hours: 2-4 AM, 9-11 AM (you're a night owl AND early bird)

Top Projects This Week:
1. claude-code-proxy - 45 sessions, 8.2MB data
2. oscilla-animator - 23 sessions, 4.1MB data
3. dotfiles - 12 sessions, 0.8MB data
```

**Bottom Section - Deep Insights**:
- Tool usage breakdown: Read 45%, Write 25%, Bash 20%, Search 10%
- Average conversation length: 34 messages
- Subagent spawn rate: 2.3 per session
- Plan creation rate: 0.4 per day
- Token usage by model over time (from stats-cache.json)

Click any insight to drill down. "Tool usage breakdown" expands to show trends over time, per-project breakdown, and links to conversations with unusual patterns.

**Why This Matters**: stats-cache.json already has this data but it's just numbers. The intelligence dashboard turns numbers into understanding: "You used Claude Code 40% more this week" or "You're most productive between 2-4 AM."

**Success Looks Like**:
- Dashboard loads instantly with cached stats
- Comparisons show meaningful trends (not just numbers)
- Every metric links to the underlying data
- Insights surface interesting patterns automatically

---

### Idea 4: The Extension Workshop - Agents, Commands, Skills, and Plugins

**User Story**: As a Claude Code customizer, I want to browse, understand, and manage all my extensions in one place so that I can see what's available and how it's all configured.

**The Experience**:

A unified extension browser shows everything:

**Agents Tab**:
```
5 Custom Agents                         [Create New]

code-bloodhound.md                      6.4 KB
  "Expert code investigator that traces..."
  Last used: 3 days ago | Used 12 times

code-monkey-jr.md                       3.4 KB
  "Junior developer persona..."
  Last used: never | Used 0 times

[Disabled]
  work-decomposer.md - moved to agents-disabled
```

Each agent shows:
- Name and description (extracted from content)
- Size and usage stats (correlated from session data)
- Quick edit link, enable/disable toggle
- Preview of the full markdown

**Commands Tab**:
Similar layout showing user commands, with execution history from history.jsonl

**Skills Tab**:
Installed skills with source marketplace, version, last update

**Plugins Tab**:
```
12 Plugins Installed                    [Check Updates]

beads@beads-marketplace                 v0.47.0
  Installed: Nov 5, 2025 | Updated: Jan 12, 2026
  Usage: 847 invocations

example-skills@anthropic-agent-skills
  Installed: Dec 21, 2025
  Status: Enabled | 0 invocations (unused?)
```

Plugin health indicators surface issues: outdated versions, unused plugins consuming space, disabled but not removed.

**Why This Matters**: Extensions are scattered across agents/, commands/, skills/, plugins/ - each with different formats and metadata. The workshop consolidates everything, adds usage analytics, and makes management visual.

**Success Looks Like**:
- All extension types visible in one interface
- Usage stats help identify valuable vs unused extensions
- Quick actions: edit, enable/disable, delete, update
- New extension creation is guided and easy

---

### Idea 5: The Time Machine - Session Replay and File History

**User Story**: As a developer, I want to see exactly what happened during a session including all file changes so that I can understand, recover, or learn from past work.

**The Experience**:

You open a conversation and click "Time Machine" mode. The interface transforms:

**Left panel**: The conversation transcript, now with timestamps and a timeline slider

**Right panel**: A file diff viewer showing the state of modified files

**Bottom panel**: A scrubber showing:
```
[10:32:14] Session start
[10:32:45] User: "Fix the auth bug"
[10:33:12] Read: src/auth/handler.go (v1)
[10:34:28] Write: src/auth/handler.go (v2)  [DIFF]
[10:35:01] Bash: go test ./...
[10:36:44] Write: src/auth/handler.go (v3)  [DIFF]
...
```

Drag the slider and watch:
- The conversation scroll to that point
- The file diff update to show the file's state at that moment
- The directory tree highlight which files existed/were modified

Click any file change to see the exact diff. Click "Restore" to get that version back. Click "Compare" to see multiple versions side-by-side.

The file-history directory makes this possible - it stores versions keyed by hash and version number. The time machine correlates these with conversation timestamps.

**Why This Matters**: File history data exists but there's no way to visualize it or correlate with conversations. The time machine makes "what did Claude do to my code?" answerable visually.

**Success Looks Like**:
- Timeline shows all events (messages, file changes, tool calls) chronologically
- Scrubbing updates all views in sync
- Any file version is recoverable
- Diffs are clear and syntax-highlighted

---

### Idea 6: The Subagent Observatory

**User Story**: As a user who works with subagents, I want to see the full hierarchy and flow of work across agent spawns so that I can understand how complex tasks are being decomposed and executed.

**The Experience**:

Any conversation that spawned subagents shows a "Subagent Map":

```
Main Session
  |
  +-- code-reviewer (agent-abc123)
  |     Task: "Review auth changes"
  |     Status: Completed
  |     Messages: 23 | Duration: 4m 32s
  |     |
  |     +-- analysis-helper (agent-def456)
  |           Task: "Check security patterns"
  |           Status: Completed
  |
  +-- test-writer (agent-ghi789)
        Task: "Write integration tests"
        Status: In Progress
        Messages: 45 | Duration: ongoing
```

Click any node to:
- View that agent's conversation in a split pane
- See its specific tool calls and file changes
- Jump to the parent message that spawned it
- See metrics: tokens used, time taken, files modified

The map shows flow: which agent spawned which, in what order, with what results. Color coding shows status (green=completed, blue=in-progress, red=failed).

For power users, aggregate views show:
- "Most commonly spawned agent types"
- "Average depth of agent chains"
- "Sessions with most subagents"

**Why This Matters**: Subagent transcripts exist in projects/{path}/{session}/subagents/ but there's no way to see hierarchy or flow. The observatory makes agent orchestration visible.

**Success Looks Like**:
- Tree visualization shows parent-child relationships
- Clicking any agent opens its conversation
- Status is clear at a glance
- Metrics help understand agent efficiency

---

### Idea 7: The Space Manager - Data Cleanup and Archival

**User Story**: As a user with limited disk space, I want to understand what's consuming space in my ~/.claude directory and clean it up safely.

**The Experience**:

A storage management view shows:

```
~/.claude Storage Analysis                 Total: 1.2 GB

[============================] projects/      680 MB (57%)
  |_ brandon-fryslie-claude-code-proxy       143 MB
  |_ oscilla-animator                         89 MB
  |_ (85 others)                             448 MB

[===============]              telemetry/    203 MB (17%)
  |_ 74 failed event files (stale?)          180 MB [Clean up]

[==========]                   debug/        145 MB (12%)
  |_ 944 debug logs                          145 MB [Compress]

[======]                       file-history/  98 MB  (8%)
  |_ 417 session histories                    98 MB

[===]                          history.jsonl   1.3 MB (0.1%)
  |_ All prompts since Dec 3                  1.3 MB

[=]                            (other)         73 MB (6%)
```

Each section expands to show:
- Individual items sorted by size
- Age (oldest files highlighted)
- Status (can this be safely deleted? archived? compressed?)

Actions:
- **Archive**: Move to compressed archive, keep index for search
- **Clean**: Delete with confirmation, show what will be removed
- **Compress**: Gzip large files, decompress on access
- **Export**: Create portable backup of selected data

Smart suggestions:
- "180 MB of failed telemetry events can be safely deleted"
- "23 sessions older than 30 days could be archived"
- "Debug logs older than 7 days serve no purpose"

**Why This Matters**: ~/.claude grows unbounded. Users don't know what's safe to delete, what's important, or how to reclaim space. The space manager makes cleanup safe and informed.

**Success Looks Like**:
- Clear visualization of space usage
- Every item shows whether it can be safely removed
- Bulk actions with preview before execution
- No accidental deletion of important data

---

### Idea 8: The Cross-Session Todo Board

**User Story**: As someone who works across many sessions, I want to see all my todos across all sessions in one kanban-style board so that I can track work regardless of where it originated.

**The Experience**:

A kanban board aggregates todos from all session files:

```
Pending (23)          In Progress (5)       Completed (147)
+------------------+  +------------------+  +------------------+
| Implement cache  |  | Add unit tests   |  | Fix auth bug     |
| Session: abc123  |  | Session: def456  |  | Session: ghi789  |
| Created: 2h ago  |  | Started: 1h ago  |  | Done: yesterday  |
+------------------+  +------------------+  +------------------+
| Review PR #45    |  | Update docs      |  | Refactor handler |
| Session: jkl012  |  | Session: abc123  |  | Session: mno345  |
+------------------+  +------------------+  +------------------+
```

Each card shows:
- The todo content
- Which session it belongs to (click to open)
- Status and timestamps
- Active form (the "doing" description)

Filters:
- By session
- By project (correlated from session data)
- By age
- By keywords

The board updates live if sessions are active. Completed todos can be archived to reduce clutter.

**Why This Matters**: Todos are stored per-session but work spans sessions. The board aggregates scattered todos into one actionable view.

**Success Looks Like**:
- All todos visible regardless of session
- Kanban layout shows workflow status
- Click-through to source session
- Filters help manage large todo counts

---

## Ideas Considered But Not Selected

1. **AI-Assisted Conversation Summarization**: Using an LLM to summarize conversations. Not selected because it adds external dependency, latency, and cost. The raw data with good navigation is more valuable than summaries.

2. **Real-Time Session Streaming**: Live view of ongoing Claude Code sessions. Not selected because it requires persistent connection, adds complexity, and the value vs. "just open Claude Code" is marginal.

3. **Collaborative Features**: Sharing conversations or configurations with team members. Not selected because ~/.claude is inherently single-user and adding collaboration fundamentally changes the product.

4. **Code Analysis on Modified Files**: Running static analysis on files Claude touched. Not selected because it's orthogonal to data visualization - better done in IDE or CI.

5. **Automated Backup to Cloud**: Syncing ~/.claude to cloud storage. Not selected because it raises privacy concerns and most users have their own backup solutions.

6. **Natural Language Query Interface**: "Show me conversations about the auth bug". Not selected for initial scope - unified search with structured syntax provides 80% of the value without the complexity.

7. **Cost Estimation from Token Usage**: Calculating API costs from stats-cache.json. Not selected because Claude Code uses subscription pricing, not API pricing, making cost calculation misleading.

8. **Configuration Editor**: In-app editing of settings.json and CLAUDE.md. Not selected because users already have editors they prefer, and we risk data corruption if our editor has bugs.

## Open Questions

1. **Data Freshness**: Should CC-VIZ monitor ~/.claude for changes and auto-refresh? Or require explicit refresh? Auto-refresh has UX benefits but adds complexity.

2. **Search Index Location**: Where should the search index live? In ~/.claude (alongside data) or in application-specific storage? Tradeoffs between portability and separation of concerns.

3. **Multi-Machine Support**: Do users want to aggregate data from multiple machines? If so, how do we handle session ID collisions and timestamp normalization?

4. **Privacy in Shared Environments**: If someone screenshots CC-VIZ, it exposes all their conversation history. Should there be a "privacy mode" that redacts content?

5. **Performance at Scale**: Some users have 1000+ sessions. What's the acceptable load time? Should we paginate aggressively or optimize for "show everything"?

6. **Archival Format**: If we add archival, what format? Keep JSONL but compressed? Convert to SQLite? Something portable like Markdown?

---

*This vision document represents the ultimate destination - a comprehensive data cockpit for Claude Code. Implementation phases will prioritize features by user impact and technical feasibility.*
