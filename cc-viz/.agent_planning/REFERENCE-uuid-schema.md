# UUID-Based Filename Schema Reference

**Purpose**: Complete reference for implementing graph-based data processing in CC-VIZ
**Created**: 2026-01-12
**Source**: Comprehensive exploration of ~/.claude directory structure

---

## Overview

The `~/.claude` directory uses UUIDs as the primary linking mechanism between data entities. Understanding these patterns is essential for building the Graph Navigator and other cross-referencing features.

### UUID Types

| Type | Format | Example | Semantics |
|------|--------|---------|-----------|
| Session UUID | UUID v4 (36 chars) | `016c7c55-1a1e-4257-8a30-befc355f6541` | Root identifier for a Claude Code session |
| Agent UUID | UUID v4 (36 chars) | `aee4a4e8-a8af-48df-a141-495997bb6019` | Subagent identifier (may equal session UUID) |
| Agent Shorthand | 7 hex chars | `a4a3a5e` | Abbreviated agent ID in project files |
| Message UUID | UUID v4 (36 chars) | `f8a91c32-...` | Individual message identifier |
| File Hash | 16 hex chars | `69219e99e98d8fec` | Content hash for file versions |

---

## 1. Session UUIDs

**Pattern**: `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` (UUID v4)

Session UUIDs are the **root identifier** - all work within a session shares this UUID.

### Where Session UUIDs Appear

| Location | Pattern | Example |
|----------|---------|---------|
| `~/.claude/session-env/` | Directory name | `session-env/016c7c55-1a1e-4257-8a30-befc355f6541/` |
| `~/.claude/debug/` | Filename | `016c7c55-1a1e-4257-8a30-befc355f6541.txt` |
| `~/.claude/file-history/` | Directory name | `file-history/016c7c55-1a1e-4257-8a30-befc355f6541/` |
| `~/.claude/todos/` | Filename prefix | `016c7c55-...-agent-....json` |
| `~/.claude/telemetry/` | Filename component | `1p_failed_events.016c7c55-....json` |
| `~/.claude/history.jsonl` | `sessionId` field | `{"sessionId": "016c7c55-..."}` |
| `~/.claude/projects/*/` | JSONL `sessionId` field | `{"sessionId": "016c7c55-..."}` |

---

## 2. Agent UUIDs

Agents can be the **main session** (agent UUID == session UUID) or **subagents** (different UUIDs).

### Two Format Types

**A. Full UUID Format** (in todos, telemetry)
```
{SESSION_UUID}-agent-{AGENT_UUID}.json
```

**B. Hex Shorthand Format** (in project JSONL files)
```
agent-{7_HEX_CHARS}.jsonl
```

### Relationship Rule

```
sessionId == agentId  →  Main session agent (primary user interaction)
sessionId != agentId  →  Subagent (spawned by parent session)
```

---

## 3. File Patterns by Directory

### `~/.claude/todos/`

**Pattern**: `{SESSION_UUID}-agent-{AGENT_UUID}.json`

**Examples**:
```
016c7c55-1a1e-4257-8a30-befc355f6541-agent-016c7c55-1a1e-4257-8a30-befc355f6541.json  # Main session
016c7c55-1a1e-4257-8a30-befc355f6541-agent-90023ae8-7ec8-4b00-810f-e53399c9b98a.json  # Subagent 1
016c7c55-1a1e-4257-8a30-befc355f6541-agent-aee4a4e8-a8af-48df-a141-495997bb6019.json  # Subagent 2
```

**Regex**:
```regex
^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})-agent-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.json$
```

**Content Structure**:
```json
[
  {
    "content": "Implement feature X",
    "status": "pending" | "in_progress" | "completed",
    "activeForm": "Implementing feature X"
  }
]
```

**Count**: ~1,050 files

---

### `~/.claude/projects/{PATH}/`

**Directory Name Pattern**: `-{PATH_WITH_SLASHES_AS_DASHES}`

**Examples**:
```
-Users-bmf-code-brandon-fryslie-claude-code-proxy
-Users-bmf-code-primitives
-Users-bmf--claude                # Note: double dash for single slash in path
```

**File Patterns Within**:

| Pattern | Example | Description |
|---------|---------|-------------|
| `{SESSION_UUID}.jsonl` | `8ba41bf5-ddb0-4b1d-af6b-1871922933a0.jsonl` | Full session transcript |
| `agent-{HEX}.jsonl` | `agent-a4a3a5e.jsonl` | Agent-specific transcript |
| `{SESSION}/subagents/agent-{HEX}.jsonl` | Nested subagent transcripts |

**JSONL Entry Structure**:
```json
{
  "sessionId": "UUID",
  "agentId": "UUID|SHORTHAND",
  "uuid": "UUID",
  "parentUuid": null | "UUID",
  "type": "user" | "assistant" | "system",
  "userType": "external",
  "timestamp": "2026-01-12T10:30:00.000Z",
  "cwd": "/Users/bmf/code/project",
  "gitBranch": "main",
  "version": "1.0.30",
  "isSidechain": false,
  "message": {
    "role": "user" | "assistant",
    "content": "string" | [{"type": "text", "text": "..."}]
  },
  "requestId": "string" | null,
  "usage": {
    "input_tokens": 1234,
    "output_tokens": 567,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 890
  }
}
```

---

### `~/.claude/file-history/`

**Structure**: `{SESSION_UUID}/{FILE_HASH}@v{N}`

**Example**:
```
~/.claude/file-history/
├── 016c7c55-1a1e-4257-8a30-befc355f6541/
│   ├── 69219e99e98d8fec@v1
│   ├── 69219e99e98d8fec@v2
│   └── c82431275e08e346@v1
```

**Components**:
- `FILE_HASH`: 16-char hex content hash
- `@vN`: Version number (increments on change)

**Content**: Raw file content (source code, config, etc.)

**Count**: ~417 sessions have file history

---

### `~/.claude/debug/`

**Pattern**: `{SESSION_UUID}.txt`

**Example**: `016c7c55-1a1e-4257-8a30-befc355f6541.txt`

**Content**: Plain text debug logs
```
2026-01-13T00:46:35.649Z [DEBUG] AutoUpdaterWrapper: Installation type: native
2026-01-13T00:46:35.690Z [DEBUG] Getting matching hook commands for SessionEnd...
```

**Count**: ~950 files

---

### `~/.claude/telemetry/`

**Pattern**: `1p_failed_events.{SESSION_UUID}.{AGENT_UUID}.json`

**Example**:
```
1p_failed_events.0baa8de5-c5d9-45ee-bcd4-ea1576bb596e.59c828cd-6f5a-492c-8239-c67018d00a8b.json
```

**Content**: JSON array of failed telemetry events

**Count**: ~77 files

---

### `~/.claude/session-env/`

**Pattern**: `{SESSION_UUID}/` (directory)

**Content**: Environment variable persistence (often empty)

**Count**: ~801 directories

---

### `~/.claude/history.jsonl`

**Single global file** with entries:

```json
{
  "display": "fix the auth bug",
  "pastedContents": {},
  "timestamp": 1736654400000,
  "project": "/Users/bmf/code/project",
  "sessionId": "016c7c55-1a1e-4257-8a30-befc355f6541"
}
```

**Links**: `sessionId` → session entities, `project` → project directory

---

### `~/.claude/plans/`

**Pattern**: `{DESCRIPTIVE_NAME}.md` (NOT UUID-based)

**Examples**:
```
ancient-seeking-rossum.md
peppy-yawning-teapot.md
idempotent-seeking-puzzle.md
```

**Note**: Plans use human-readable names (adjective-noun combinations), not UUIDs. Link to sessions via content references or creation timestamp correlation.

---

### `~/.claude/shell-snapshots/`

**Pattern**: `snapshot-bash-{TIMESTAMP}-{RANDOM_ID}.sh`

**Example**: `snapshot-bash-1766854480395-wp0aq9.sh`

**Components**:
- `TIMESTAMP`: Unix milliseconds
- `RANDOM_ID`: 6-char random identifier

**Count**: ~137 files

---

### `~/.claude/paste-cache/`

**Pattern**: `{HEX_16}.txt`

**Example**: `57bbac341b02760a.txt`

**Semantics**: Content-addressed cache (hash of pasted content)

**Count**: ~16 files

---

## 4. Entity Relationship Graph

```
┌─────────────────────────────────────────────────────────────────────┐
│                           SESSION (UUID)                            │
│  Root node - all entities trace back to session UUID                │
└─────────────────────────────────────────────────────────────────────┘
         │
         ├──► session-env/{SESSION}/
         ├──► debug/{SESSION}.txt
         ├──► file-history/{SESSION}/
         ├──► history.jsonl entries (sessionId field)
         │
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           AGENT (UUID)                              │
│  Primary (sessionId == agentId) or Subagent (different)             │
└─────────────────────────────────────────────────────────────────────┘
         │
         ├──► todos/{SESSION}-agent-{AGENT}.json
         ├──► telemetry/1p_failed_events.{SESSION}.{AGENT}.json
         ├──► projects/{PATH}/agent-{HEX}.jsonl
         │
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          MESSAGE (UUID)                             │
│  Linked via parentUuid field - creates conversation DAG             │
└─────────────────────────────────────────────────────────────────────┘
         │
         └──► Tool calls, file edits tracked in file-history
```

### Key Relationships

| From | To | Via |
|------|-----|-----|
| Session | Agents | `sessionId` field matches |
| Session | Todos | Filename prefix `{SESSION}-agent-*` |
| Session | Debug | Filename `{SESSION}.txt` |
| Session | File History | Directory `file-history/{SESSION}/` |
| Session | History | `sessionId` field in history.jsonl |
| Agent | Messages | `agentId` field in JSONL entries |
| Agent | Todos | Filename suffix `-agent-{AGENT}.json` |
| Message | Parent | `parentUuid` field (null for root) |
| Message | Files | Tool calls referencing file paths |

---

## 5. Graph Processing Implications

### For Indexing

1. **Parse session UUID from filename** using regex
2. **Extract agent UUID** from filename (todos) or `agentId` field (JSONL)
3. **Identify primary vs subagent** by comparing sessionId == agentId
4. **Build message chains** via parentUuid references
5. **Correlate file changes** via timestamps + file-history

### For Navigation

- **Session → All Agents**: Query by sessionId
- **Agent → All Messages**: Filter JSONL by agentId
- **Message → Context**: Follow parentUuid chain upward
- **File → Sessions**: Find sessions with matching file-history entries

### For Search

- **FTS5 indexes** should include sessionId/agentId for filtering
- **Cross-entity search** requires joining on UUIDs
- **Time-based queries** use timestamp fields

---

## 6. File Counts Summary

| Directory | Pattern | Count | Primary Key |
|-----------|---------|-------|-------------|
| `session-env/` | `{SESSION}/` | ~801 | Session UUID |
| `debug/` | `{SESSION}.txt` | ~950 | Session UUID |
| `file-history/` | `{SESSION}/{HASH}@v{N}` | ~417 sessions | Session UUID |
| `todos/` | `{SESSION}-agent-{AGENT}.json` | ~1,050 | Session + Agent UUID |
| `telemetry/` | `1p_failed_events.{S}.{A}.json` | ~77 | Session + Agent UUID |
| `projects/*/` | Various JSONL | varies | Project + Session/Agent |
| `shell-snapshots/` | `snapshot-bash-{TS}-{ID}.sh` | ~137 | Timestamp |
| `paste-cache/` | `{HEX_16}.txt` | ~16 | Content hash |
| `history.jsonl` | Single file | 1 | Session UUID field |
| `plans/` | `{name}.md` | varies | Name (not UUID) |

---

## 7. Regex Patterns for Parsing

### Session UUID
```regex
[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}
```

### Todo Filename
```regex
^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})-agent-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.json$
```
- Group 1: Session UUID
- Group 2: Agent UUID

### Agent Shorthand
```regex
agent-([0-9a-f]{7})
```
- Group 1: 7-char hex ID

### File History Version
```regex
([0-9a-f]{16})@v(\d+)
```
- Group 1: File hash
- Group 2: Version number

### Telemetry Filename
```regex
1p_failed_events\.([0-9a-f-]{36})\.([0-9a-f-]{36})\.json
```
- Group 1: Session UUID
- Group 2: Agent UUID

---

## 8. Implementation Checklist

When implementing graph-based processing:

- [ ] Parse todo filenames to extract session + agent UUIDs
- [ ] Parse JSONL entries for sessionId, agentId, uuid, parentUuid
- [ ] Build session → agents relationship table
- [ ] Build agent → messages relationship (via JSONL)
- [ ] Build message → parent chains (via parentUuid)
- [ ] Correlate file-history with sessions
- [ ] Index history.jsonl for session lookup
- [ ] Handle agent shorthand → full UUID mapping if needed
- [ ] Create cross-reference tables for navigation
