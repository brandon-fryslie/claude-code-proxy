# Agent Instructions

## Project Overview

Claude Code Proxy - A transparent proxy for monitoring Claude Code API requests with optional routing to different LLM providers.

**Architecture:** Split services for zero-downtime deployments
- `proxy-core` (port 3001) - Lightweight API proxy, rarely changes
- `proxy-data` (port 3002) - Dashboard APIs, stats, indexing
- Caddy (port 3000) - Reverse proxy routing

## Workflow Commands

All operations use `just`. These are the ONLY commands you need:

```bash
just install       # Install dependencies (first-time)
just build         # Build everything
just run           # Run locally (requires Caddy)
just docker        # Run with Docker (best for dev)
just stop          # Stop Docker
just restart-data  # Zero-downtime update
just test          # Run tests
just check         # Lint + typecheck
just db            # Reset database
just clean         # Clean everything
```

## Issue Tracking

Uses **bd** (beads) for issue tracking:

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Session Completion

**When ending a work session**, complete ALL steps:

1. **Quality gates** - `just test && just check`
2. **File issues** - `bd add` for remaining work
3. **Update issues** - Close finished, update in-progress
4. **Commit** - MANDATORY
5. **Hand off** - Context for next session

## Key Directories

```
proxy/cmd/proxy-core/     # Lightweight proxy entry point
proxy/cmd/proxy-data/     # Data service entry point
proxy/internal/handler/   # HTTP handlers (core_handler.go, data_handler.go)
proxy/internal/service/   # Business logic, storage, indexer
web/                      # Legacy Remix dashboard
dashboard/                # New React dashboard
docker/                   # Docker configs for split services
```

## Testing Changes

After making changes:

```bash
just test          # Must pass
just check         # Must pass
just run           # Verify it works
```

For Docker-based testing:
```bash
just docker        # Backend in Docker, frontends local (HMR works)
just restart-data  # Test zero-downtime updates
```
