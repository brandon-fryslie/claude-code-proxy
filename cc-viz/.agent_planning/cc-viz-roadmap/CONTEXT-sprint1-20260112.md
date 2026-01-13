# Implementation Context: Sprint 1 - Configuration Foundation
Generated: 2026-01-12

## Codebase Structure

### Frontend (cc-viz/)
```
cc-viz/
├── src/
│   ├── main.tsx                          # App entry, routing logic (lines 1-30)
│   ├── pages/
│   │   ├── Home.tsx                      # Landing page with category cards
│   │   └── Conversations.tsx             # Existing conversation browser
│   ├── components/
│   │   ├── layout/
│   │   │   └── AppLayout.tsx             # Sidebar + page header (lines 1-240)
│   │   ├── features/
│   │   │   └── ConversationThread.tsx    # Message display (reusable patterns)
│   │   └── ui/
│   │       └── Link.tsx                  # Simple link component
│   └── lib/
│       ├── api.ts                        # API hooks (useConversations, etc.)
│       └── utils.ts                      # cn() utility for classnames
```

### Backend (proxy/)
```
proxy/
├── cmd/
│   └── proxy-data/main.go                # Data service entry
├── internal/
│   └── handler/
│       └── data_handler.go               # API handlers (add new endpoints here)
```

## Key Files to Modify

### 1. Frontend Routing (src/main.tsx)
```typescript
// Current routing (lines 20-28):
function App() {
  const path = window.location.pathname
  if (path === '/cc-viz/conversations' || path === '/cc-viz/conversations/') {
    return <ConversationsPage />
  }
  return <HomePage />
}

// ADD: Routes for configuration and projects
if (path === '/cc-viz/configuration' || path === '/cc-viz/configuration/') {
  return <ConfigurationPage />
}
if (path === '/cc-viz/projects' || path === '/cc-viz/projects/') {
  return <ProjectsPage />
}
```

### 2. Sidebar Navigation (src/components/layout/AppLayout.tsx)
```typescript
// Lines 36-68: navSections array
// Currently "Configuration" and "Projects" are disabled
// Update to add href and remove disabled:

{ id: 'configuration', label: 'Configuration', icon: <Settings size={18} />, href: '/cc-viz/configuration' },
{ id: 'projects', label: 'Projects', icon: <FolderKanban size={18} />, href: '/cc-viz/projects' },
```

### 3. Home Page Cards (src/pages/Home.tsx)
```typescript
// Lines 22-30: Configuration category
// Update status to 'available' and add href:
{
  id: 'configuration',
  title: 'Configuration',
  ...
  status: 'available',
  href: '/cc-viz/configuration',
},

// Lines 50-58: Projects category
// Update status to 'available' and add href:
{
  id: 'projects',
  title: 'Projects',
  ...
  status: 'available',
  href: '/cc-viz/projects',
},
```

### 4. New API Endpoints (proxy/internal/handler/data_handler.go)

Add to DataHandler struct methods:

```go
// GET /api/claude-config
func (h *DataHandler) GetClaudeConfig(w http.ResponseWriter, r *http.Request) {
    homeDir, _ := os.UserHomeDir()
    claudeDir := filepath.Join(homeDir, ".claude")

    // Read settings.json
    settings, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))

    // Read CLAUDE.md (follow symlinks)
    claudeMd, _ := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))

    // Read .mcp.json
    mcpConfig, _ := os.ReadFile(filepath.Join(claudeDir, ".mcp.json"))

    response := map[string]interface{}{
        "settings":   json.RawMessage(settings),
        "claude_md":  string(claudeMd),
        "mcp_config": json.RawMessage(mcpConfig),
    }
    json.NewEncoder(w).Encode(response)
}

// GET /api/projects
func (h *DataHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
    homeDir, _ := os.UserHomeDir()
    projectsDir := filepath.Join(homeDir, ".claude", "projects")

    entries, _ := os.ReadDir(projectsDir)
    projects := []map[string]interface{}{}

    for _, entry := range entries {
        if entry.IsDir() {
            info, _ := entry.Info()
            // Decode path: "-Users-bmf-code-foo" -> "/Users/bmf/code/foo"
            decodedPath := strings.ReplaceAll(entry.Name(), "-", "/")
            projects = append(projects, map[string]interface{}{
                "id":           entry.Name(),
                "path":         decodedPath,
                "last_modified": info.ModTime(),
            })
        }
    }
    json.NewEncoder(w).Encode(projects)
}

// GET /api/projects/{id}
func (h *DataHandler) GetProject(w http.ResponseWriter, r *http.Request) {
    // Extract id from path
    // Read project's CLAUDE.md, settings.json if they exist
}
```

### 5. Register Routes (proxy/cmd/proxy-data/main.go or router setup)

```go
router.HandleFunc("/api/claude-config", handler.GetClaudeConfig).Methods("GET")
router.HandleFunc("/api/projects", handler.ListProjects).Methods("GET")
router.HandleFunc("/api/projects/{id}", handler.GetProject).Methods("GET")
```

## New Files to Create

### src/pages/Configuration.tsx
```typescript
import { AppLayout } from '@/components/layout'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'

interface ClaudeConfig {
  settings: object | null
  claude_md: string | null
  mcp_config: object | null
}

export function ConfigurationPage() {
  const [activeTab, setActiveTab] = useState<'settings' | 'claude_md' | 'mcp'>('settings')

  const { data, isLoading } = useQuery<ClaudeConfig>({
    queryKey: ['claude-config'],
    queryFn: () => fetch('/api/claude-config').then(r => r.json())
  })

  return (
    <AppLayout title="Configuration" activeItem="configuration">
      {/* Tab bar */}
      {/* Content display based on activeTab */}
    </AppLayout>
  )
}
```

### src/pages/Projects.tsx
```typescript
import { AppLayout } from '@/components/layout'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'

interface Project {
  id: string
  path: string
  last_modified: string
}

export function ProjectsPage() {
  const [selectedProject, setSelectedProject] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const { data: projects } = useQuery<Project[]>({
    queryKey: ['projects'],
    queryFn: () => fetch('/api/projects').then(r => r.json())
  })

  return (
    <AppLayout title="Projects" activeItem="projects">
      {/* Search bar */}
      {/* Project list (left) */}
      {/* Project detail (right) */}
    </AppLayout>
  )
}
```

## Data Format Examples

### ~/.claude/settings.json
```json
{
  "permissions": {
    "allow": ["Bash(git:*)", "Bash(npm:*)", ...],
    "deny": []
  },
  "theme": "dark",
  ...
}
```

### ~/.claude/.mcp.json
```json
{
  "mcpServers": {
    "@anthropic/sqlite": {
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-server-sqlite"],
      "type": "stdio"
    }
  }
}
```

### ~/.claude/projects/ directory naming
- Directory name: `-Users-bmf-code-brandon-fryslie-claude-code-proxy`
- Decoded path: `/Users/bmf/code/brandon-fryslie/claude-code-proxy`
- Contents: May contain CLAUDE.md, settings.json, agents/, commands/

## Styling Patterns

Use existing CSS variables from index.css:
- `var(--color-bg-primary)` - main background
- `var(--color-bg-secondary)` - card/panel background
- `var(--color-text-primary)` - main text
- `var(--color-text-muted)` - secondary text
- `var(--color-border)` - borders
- `var(--color-accent)` - highlight color

## Dependencies to Add

```bash
pnpm add react-markdown remark-gfm  # For markdown rendering
```

Or use existing syntax highlighting if available in the codebase.
