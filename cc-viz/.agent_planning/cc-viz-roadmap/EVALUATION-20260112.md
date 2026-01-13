# CC-VIZ Roadmap Evaluation
Generated: 2026-01-12

## Current State

CC-VIZ has a landing page with 10 visualizer categories. Only **Conversations (basic)** is currently implemented.

### Existing Infrastructure
- React + Vite + TanStack Query frontend
- Go backend with SQLite for conversation indexing
- Sidebar navigation with all categories (disabled)
- Landing page with category cards

## User-Specified Priority Order

1. **Configuration (basic)** - settings.json, CLAUDE.md, .mcp.json display
2. **Per Project Settings** - projects/ directory browser
3. **Session Data** - todos/, plans/, debug/, session-env/
4. **Conversations (advanced)** - enhanced search, analytics, filtering
5. **Hooks** - hooks/*.json display
6. **Agents** - agents/*.md browser
7. **Commands** - commands/*.md browser
8. **Skills** - skills/*.md browser
9. **Plugins** - installed_plugins.json, cache/, marketplaces/
10. **Configuration (advanced)** - editing capabilities
11. **History** - history.jsonl, file-history/, paste-cache/
12. **IDE Integration** - ide/ directory
13. **Telemetry** - telemetry/, statsig/, stats-cache.json

## Data Sources Analysis

| Priority | Visualizer | Files | Effort | Notes |
|----------|------------|-------|--------|-------|
| 1 | Configuration (basic) | settings.json, CLAUDE.md, .mcp.json | Low | Read-only display |
| 2 | Per Project Settings | projects/*/ (~87 dirs) | Medium | Aggregate many dirs |
| 3 | Session Data | todos/, plans/, debug/, session-env/ | Medium | Multiple file types |
| 4 | Conversations (advanced) | Existing + enhancements | Medium | Extend existing |
| 5 | Hooks | hooks/*.json (1 file currently) | Very Low | Simple JSON |
| 6 | Agents | agents/*.md (~5 files) | Low | Markdown + frontmatter |
| 7 | Commands | commands/*.md (~3 files) | Low | Markdown + frontmatter |
| 8 | Skills | skills/*.md (~3 files) | Low | Markdown + frontmatter |
| 9 | Plugins | Complex tree structure | Medium-High | Nested JSON/cache |
| 10 | Configuration (advanced) | Same as basic + writes | High | Validation, editing |
| 11 | History | history.jsonl (1.2MB), file-history/ | Medium | Large files, pagination |
| 12 | IDE Integration | ide/ (~8 files) | Low | Simple JSON |
| 13 | Telemetry | telemetry/, statsig/ | Low | Analytics display |

## Sprint Breakdown

### Sprint 1: Configuration Foundation
- Configuration (basic)
- Per Project Settings

### Sprint 2: Session Intelligence
- Session Data
- Conversations (advanced)

### Sprint 3: Customization Browser
- Hooks
- Agents
- Commands
- Skills

### Sprint 4: Plugin Ecosystem
- Plugins

### Sprint 5: Advanced & Polish
- Configuration (advanced)
- History
- IDE Integration
- Telemetry

## Verdict: CONTINUE

Priority order established by user. Ready to plan Sprint 1.
