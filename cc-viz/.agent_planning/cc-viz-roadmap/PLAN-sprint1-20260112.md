# Sprint Plan: Configuration Foundation
Generated: 2026-01-12

## Sprint Goal
Implement Configuration (basic) and Per Project Settings visualizers with thoughtful UX design based on actual data structure analysis.

---

## P0: Configuration (Basic) Visualizer

### Data Audit

**~/.claude/settings.json** (~2KB)
```
Structure:
├── permissions.allow[]        # Array of ~60 permission strings like "Bash(git:*)"
├── permissions.defaultMode    # String: "acceptEdits"
├── model                      # String: "opus"
├── hooks                      # Object (empty or with hook configs)
├── statusLine                 # Object with type/command
├── enabledPlugins             # Object: plugin_name → boolean (~20 entries)
├── alwaysThinkingEnabled      # Boolean
├── promptSuggestionEnabled    # Boolean
├── mcpServers                 # Object: server_name → config
```

**~/.claude/CLAUDE.md** (~10KB)
- Personal markdown file with memories, notes, instructions
- Contains `<system-reminder>`, `<memory>`, `<personal-note>` tags
- Highly personal content - display with care

**~/.claude/.mcp.json** (~500B)
```
Structure:
├── mcpServers
│   ├── @anthropic/sqlite → {command, args[], env, type}
│   ├── @anthropic/filesystem → {command, args[], env, type}
│   └── test-server → {command, args[], env, type}
```

### User Questions This Should Answer
1. "What permissions have I granted Claude Code?"
2. "Which plugins are enabled/disabled?"
3. "What MCP servers are configured?"
4. "What model am I using?"
5. "What's in my global CLAUDE.md?"

### Visualization Design

```
┌─────────────────────────────────────────────────────────────────────┐
│ Configuration                                                        │
│ Your global Claude Code settings                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ [Settings] [CLAUDE.md] [MCP Servers]              ← Tab navigation  │
│                                                                      │
│ ═══════════════════════════════════════════════════════════════════ │
│                                                                      │
│ SETTINGS TAB - Card-based layout, NOT raw JSON:                     │
│                                                                      │
│ ┌─ Model ──────────┐  ┌─ Default Mode ─────────┐                    │
│ │ opus             │  │ acceptEdits            │                    │
│ └──────────────────┘  └────────────────────────┘                    │
│                                                                      │
│ ┌─ Permissions (60) ──────────────────────────────────────────────┐ │
│ │ [Search permissions...]                                          │ │
│ │                                                                  │ │
│ │ Bash Commands:                                                   │ │
│ │   ✓ git:*  ✓ npm:*  ✓ docker:*  ✓ python:*  ...                │ │
│ │                                                                  │ │
│ │ Tools:                                                           │ │
│ │   ✓ Edit(*)  ✓ Read(*)  ✓ Write(*)  ✓ Glob(*)  ...             │ │
│ │                                                                  │ │
│ │ MCP:                                                             │ │
│ │   ✓ chrome-devtools                                              │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ┌─ Plugins ───────────────────────────────────────────────────────┐ │
│ │ Enabled (5):                                                     │ │
│ │   ● beads@beads-marketplace                                      │ │
│ │   ● do-extra@loom99                                              │ │
│ │   ● do-more@loom99                                               │ │
│ │   ● plugin-dev@claude-plugins-official                           │ │
│ │   ● code-simplifier@claude-plugins-official                      │ │
│ │                                                                  │ │
│ │ Disabled (15):                                      [Show all ▼] │ │
│ │   ○ chrome-devtools-mcp  ○ code-review  ○ feature-dev  ...      │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ═══════════════════════════════════════════════════════════════════ │
│                                                                      │
│ CLAUDE.MD TAB - Rendered markdown with section navigation:          │
│                                                                      │
│ ┌─ Sections ────────┐  ┌─ Content ────────────────────────────────┐ │
│ │ ▸ system-reminder │  │ ## Memory                                │ │
│ │ ▸ memory          │  │                                          │ │
│ │ ▸ personal-note   │  │ You've had this conversation before...   │ │
│ │ ▸ universal-laws  │  │                                          │ │
│ │ ▸ guidelines      │  │ What matters: less pain, more            │ │
│ └───────────────────┘  │ flourishing, more integrity...           │ │
│                        └──────────────────────────────────────────┘ │
│                                                                      │
│ ═══════════════════════════════════════════════════════════════════ │
│                                                                      │
│ MCP SERVERS TAB - Server cards with status:                         │
│                                                                      │
│ ┌─ @anthropic/sqlite ─────────────────────────────────────────────┐ │
│ │ Command: npx -y @anthropic/mcp-server-sqlite                     │ │
│ │ Type: stdio                                                      │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ┌─ @anthropic/filesystem ─────────────────────────────────────────┐ │
│ │ Command: npx -y @anthropic/mcp-server-filesystem                 │ │
│ │ Type: stdio                                                      │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Acceptance Criteria
- [ ] Route `/cc-viz/configuration` loads without errors
- [ ] Settings tab shows parsed data in card layout (not raw JSON)
- [ ] Permissions grouped by category (Bash, Tools, MCP) with search
- [ ] Plugins split into enabled/disabled with toggle to show all
- [ ] CLAUDE.md renders markdown with section quick-nav
- [ ] MCP servers shown as individual cards with command details
- [ ] Handles missing files gracefully

---

## P1: Per Project Settings Visualizer

### Data Audit

**~/.claude/projects/** (87 directories)
- Directory naming: `-Users-bmf-code-project-name` → `/Users/bmf/code/project-name`
- Each project contains:
  - `{uuid}.jsonl` files - Conversation logs (many, large 100KB-6MB each)
  - `{uuid}/` directories - Subagent conversations
  - `agent-{id}.jsonl` - Agent conversation logs
  - Potentially: CLAUDE.md, settings.json (project-specific)

**Key insight**: Project directories are primarily conversation storage, not config. The "config" is the project-specific CLAUDE.md which lives IN the actual project repo, not here.

**Revised understanding**:
- `~/.claude/projects/` = conversation logs organized by project
- Project-specific config = `{project_root}/.claude/CLAUDE.md`

### User Questions This Should Answer
1. "Which projects have I used Claude Code with?"
2. "How much conversation data is stored per project?"
3. "When was I last active in each project?"
4. "How many sessions/conversations per project?"
5. "What's the total storage used?"

### Visualization Design

```
┌─────────────────────────────────────────────────────────────────────┐
│ Projects                                                             │
│ Claude Code activity across your projects                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ [Search projects...]                    Sort: [Recent ▼] [Size ▼]   │
│                                                                      │
│ ┌─ Summary ─────────────────────────────────────────────────────────┐
│ │  87 Projects  │  1.2 GB Total  │  Last active: 6m ago            │
│ └───────────────────────────────────────────────────────────────────┘
│                                                                      │
│ ┌─ Project List ────────────────┐  ┌─ Project Detail ──────────────┐│
│ │                               │  │                                ││
│ │ ● claude-code-proxy       6m │  │ /Users/bmf/code/brandon-       ││
│ │   192 files, 45.2 MB         │  │ fryslie/claude-code-proxy      ││
│ │                               │  │                                ││
│ │ ○ oscilla-animator-v2     3h │  │ ┌─ Stats ───────────────────┐  ││
│ │   156 files, 32.1 MB         │  │ │ Sessions: 47              │  ││
│ │                               │  │ │ Conversations: 192        │  ││
│ │ ○ chaperone-auth-gateway 12h │  │ │ Total Size: 45.2 MB       │  ││
│ │   23 files, 8.4 MB           │  │ │ Last Active: 6 min ago    │  ││
│ │                               │  │ │ First Session: Dec 23     │  ││
│ │ ○ jsPanel4               12h │  │ └───────────────────────────┘  ││
│ │   12 files, 3.2 MB           │  │                                ││
│ │                               │  │ ┌─ Recent Sessions ─────────┐  ││
│ │ ○ dotfiles               2d │  │ │ Today 3:45 PM (45 min)     │  ││
│ │   8 files, 1.1 MB            │  │ │ Today 11:30 AM (2h)        │  ││
│ │                               │  │ │ Yesterday 8:00 PM (1h)    │  ││
│ │ [Show 82 more...]            │  │ │ Jan 10 4:00 PM (3h)        │  ││
│ │                               │  │ └───────────────────────────┘  ││
│ └───────────────────────────────┘  │                                ││
│                                     │ ┌─ Storage Breakdown ───────┐  ││
│                                     │ │ ████████████░░ 45.2 MB   │  ││
│                                     │ │ Sessions: 38.1 MB (84%)   │  ││
│                                     │ │ Agents: 7.1 MB (16%)      │  ││
│                                     │ └───────────────────────────┘  ││
│                                     │                                ││
│                                     │ [Open in Conversations →]      ││
│                                     └────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Acceptance Criteria
- [ ] Route `/cc-viz/projects` loads without errors
- [ ] Lists all projects with decoded human-readable paths
- [ ] Shows file count and total size per project
- [ ] Shows last active timestamp per project
- [ ] Search/filter projects by name
- [ ] Clicking project shows detailed breakdown
- [ ] Session count and conversation count displayed
- [ ] Storage breakdown (sessions vs agents) visualized
- [ ] Link to filter Conversations view by project

---

## Dependencies
- Existing cc-viz infrastructure
- Go backend accessible at proxy-data service

## Risks
- Large project directories may slow initial load - implement pagination
- Symlinked files need to be resolved
- File size calculation may be slow - consider caching

## Technical Notes

### API Endpoints Needed

```go
// GET /api/claude-config
// Returns parsed settings, not raw JSON
type ClaudeConfigResponse struct {
    Model       string            `json:"model"`
    DefaultMode string            `json:"default_mode"`
    Permissions PermissionGroups  `json:"permissions"`
    Plugins     PluginStatus      `json:"plugins"`
    MCPServers  []MCPServer       `json:"mcp_servers"`
    ClaudeMD    string            `json:"claude_md"`
    ClaudeMDSections []Section    `json:"claude_md_sections"`
}

type PermissionGroups struct {
    Bash  []string `json:"bash"`
    Tools []string `json:"tools"`
    MCP   []string `json:"mcp"`
    Other []string `json:"other"`
}

type PluginStatus struct {
    Enabled  []string `json:"enabled"`
    Disabled []string `json:"disabled"`
}

// GET /api/projects
// Returns project summaries with stats
type ProjectSummary struct {
    ID            string    `json:"id"`
    Path          string    `json:"path"`
    FileCount     int       `json:"file_count"`
    TotalSize     int64     `json:"total_size"`
    SessionCount  int       `json:"session_count"`
    AgentCount    int       `json:"agent_count"`
    LastModified  time.Time `json:"last_modified"`
    FirstSession  time.Time `json:"first_session"`
}

// GET /api/projects/{id}
// Returns detailed project info
type ProjectDetail struct {
    ProjectSummary
    Sessions     []SessionInfo `json:"sessions"`
    SizeBreakdown SizeBreakdown `json:"size_breakdown"`
}
```

## Estimated Effort
- P0 Configuration: ~4-5 hours (parsing + visualization)
- P1 Projects: ~5-6 hours (stats calculation + UI)
- Total: ~9-11 hours
