# Monorepo Restructure Evaluation

**Status**: Analysis Complete
**Created**: 2026-01-12
**Scope**: Restructure to monorepo with separate cc-viz project

## Executive Summary

The project is well-positioned for restructuring. Architecture already split by service (proxy-core, proxy-data) with independent deployment capability. Main activity: consolidate dashboards into unified cc-viz project.

**Minimal changes required**: Move frontend code, update build scripts, separate Go modules.

**Risk**: Low. Clear service boundaries exist.

---

## 1. Current Structure

**Root layout:**
- `proxy/` - Go monorepo (both proxy-core and proxy-data code)
- `dashboard/` - New React app (current)
- `web/` - Legacy Remix app (deprecated)
- `Justfile` - Build orchestration
- `docker-compose.yml` - Service orchestration

**Services:**
- proxy-core (port 8001) - Lightweight API proxy, writes requests
- proxy-data (port 8002) - Data APIs, conversation indexing, stats
- Caddy (port 8000) - Reverse proxy routing

**Build system**: Justfile with unified `install`, `build`, `docker`, `test` commands

---

## 2. Target Structure

```
project-root/
├── proxy-core/              # STABLE - Lightweight proxy (extract from proxy/)
│   ├── go.mod
│   ├── cmd/proxy-core/
│   └── internal/
│       ├── config/          # Shared
│       ├── handler/         # Core handlers only
│       ├── model/           # Shared types
│       └── provider/
│
├── proxy-data/              # EVOLVING - Data APIs (extract from proxy/)
│   ├── go.mod
│   ├── cmd/proxy-data/
│   └── internal/
│       ├── handler/         # Data handlers
│       └── service/         # Storage, indexing
│
├── cc-viz/                  # NEW - Visualization project
│   ├── apps/
│   │   ├── dashboard/       # Move from dashboard/
│   │   └── web/            # Move from web/ (deprecated)
│   └── packages/           # Shared UI (future)
│
├── Justfile                 # Updated paths
├── docker-compose.yml       # Unchanged
└── docker/                  # Unchanged
```

**Key change**: Split `proxy/` into separate `proxy-core/` and `proxy-data/` modules

---

## 3. What Needs Changes

### Backend: Split Go Modules

**proxy-core/go.mod** - Minimal, stable:
- gorilla/handlers, mux
- godotenv, yaml, sqlite3
- Internal: config, handler, model, provider

**proxy-data/go.mod** - Full dependencies:
- Same as proxy-core
- Plus: service (storage, indexing)
- Uses go.mod replacements to link to shared code in proxy-core

### Frontend: Move to cc-viz

**Simple move**:
- `dashboard/` → `cc-viz/apps/dashboard/`
- `web/` → `cc-viz/apps/web/`
- Update Justfile build commands
- Mark web/ as deprecated

### Build System Updates

**Justfile changes**:
```
build:
    cd proxy-core && go build ...
    cd proxy-data && go build ...
    cd cc-viz/apps/dashboard && pnpm run build
    cd cc-viz/apps/web && pnpm run build
```

**Docker updates**:
- `Dockerfile.proxy-core` - COPY from proxy-core/
- `Dockerfile.proxy-data` - COPY from proxy-data/, include proxy-core/ for replacements
- Frontend: No change (still copied from cc-viz/apps/)

---

## 4. Dependencies

### Service Graph

```
proxy-core (stable)
  ↓ (writes DB)
SQLite (requests.db)
  ↑ (reads DB)
proxy-data (evolving)
  ↑ (API calls)
cc-viz (dashboards)
```

### Module Dependencies

- **proxy-core**: Independent, no deps on proxy-data
- **proxy-data**: Depends on shared code from proxy-core via go.mod replacements
- **cc-viz**: No backend deps, pure frontend

---

## 5. Risks

| Risk | Likelihood | Mitigation |
|------|------------|-----------|
| go.mod circular deps | Low | Test `go mod tidy` in each dir |
| Docker build paths fail | Medium | Test each Dockerfile independently |
| Frontend build paths wrong | High | Update incrementally, test locally |
| Module replacement paths break | Medium | Document and test before deploy |

**Mitigation strategy**:
- Separate commits per phase
- Comprehensive testing after each phase
- Keep git history (don't squash)

---

## 6. Ambiguities Resolved

**Q: Should shared code (config, models) be in separate `shared/` package?**
A: No - keep in proxy-core, proxy-data imports via replacement. Simpler, can refactor later.

**Q: Web/ deprecation - move or delete?**
A: Move to cc-viz, mark deprecated, remove Q2 2026 if unused.

**Q: Should cc-viz use pnpm workspaces?**
A: Not initially - simple move is lower risk. Migrate to workspaces in Phase 2 if needed.

---

## 7. Implementation Phases

**Phase 1: proxy-core extraction** (1 week) - Low risk, stable service
**Phase 2: proxy-data extraction** (1 week) - Medium risk, has schema
**Phase 3: Frontend move to cc-viz** (1 week) - Low risk, self-contained
**Phase 4-5: Verification & cleanup** (1 week) - Testing, documentation

**Total: 4-5 weeks**

---

## Verdict: CONTINUE

Infrastructure ready, clear boundaries, low risk. Ready to proceed with planning.
