# Claude Code Proxy - Task runner

# Default recipe
default: help

# ============================================================================
# Setup
# ============================================================================

# Install all dependencies
install:
    @echo "Installing Go dependencies..."
    cd proxy && go mod download
    @echo "Installing web dependencies..."
    cd web && pnpm install
    @echo "Installing dashboard dependencies..."
    cd dashboard && pnpm install

# ============================================================================
# Build (Local)
# ============================================================================

# Build all services
build: build-proxy build-web build-dashboard

# Build proxy server (monolith)
build-proxy:
    @echo "Building proxy server (monolith)..."
    cd proxy && go build -tags "fts5" -o ../bin/proxy cmd/proxy/main.go

# Build split services
build-proxy-core:
    @echo "Building proxy-core..."
    cd proxy && go build -tags "fts5" -o ../bin/proxy-core cmd/proxy-core/main.go

build-proxy-data:
    @echo "Building proxy-data..."
    cd proxy && go build -tags "fts5" -o ../bin/proxy-data cmd/proxy-data/main.go

build-split: build-proxy-core build-proxy-data
    @echo "Split services built successfully"

# Build legacy web interface
build-web:
    @echo "Building web interface..."
    cd web && pnpm run build

# Build new dashboard
build-dashboard:
    @echo "Building dashboard..."
    cd dashboard && pnpm run build

# ============================================================================
# Run (Local Development)
# ============================================================================

# Run everything in development mode (monolith: proxy + web + dashboard)
dev:
    ./run.sh

# Run split architecture in development mode
dev-split:
    ./run-split.sh

# Run proxy only (monolith)
run-proxy:
    cd proxy && go run -tags "fts5" cmd/proxy/main.go

# Run split services individually
run-proxy-core:
    cd proxy && go run -tags "fts5" cmd/proxy-core/main.go

run-proxy-data:
    cd proxy && go run -tags "fts5" cmd/proxy-data/main.go

# Run web dashboard only
run-web:
    cd web && pnpm dev

# Run new dashboard only
run-dashboard:
    cd dashboard && pnpm dev

# ============================================================================
# Docker Commands
# ============================================================================

# Build Docker images for split services
docker-build:
    @echo "Building Docker images..."
    docker build -f docker/Dockerfile.proxy-core -t claude-proxy-core .
    docker build -f docker/Dockerfile.proxy-data -t claude-proxy-data .

# Build monolith Docker image
docker-build-monolith:
    @echo "Building monolith Docker image..."
    docker build -t claude-code-proxy .

# Run split architecture in Docker (production-like)
docker-up:
    @echo "Starting Docker split services..."
    docker-compose -f docker-compose.split.yml up --build

# Run split architecture in Docker (detached)
docker-up-detached:
    docker-compose -f docker-compose.split.yml up -d --build

# Run full dev stack in Docker (includes frontends with HMR)
docker-dev:
    @echo "Starting Docker dev stack with HMR..."
    docker-compose -f docker-compose.dev.yml up --build

# Run backend in Docker, frontends locally (best HMR experience)
docker-backend:
    @echo "Starting Docker backend only..."
    @echo "Run frontends locally with: just run-web & just run-dashboard"
    docker-compose -f docker-compose.backend.yml up --build

# Run backend in Docker (detached), then start local frontends
docker-hybrid:
    @echo "Starting hybrid dev mode (Docker backend + local frontends)..."
    docker-compose -f docker-compose.backend.yml up -d --build
    @echo ""
    @echo "Backend running in Docker. Starting frontends..."
    @echo "Web:       http://localhost:5173"
    @echo "Dashboard: http://localhost:5174"
    @echo "API:       http://localhost:3000"
    @echo ""
    cd web && pnpm dev &
    cd dashboard && pnpm dev

# Stop all Docker services
docker-down:
    docker-compose -f docker-compose.split.yml down 2>/dev/null || true
    docker-compose -f docker-compose.dev.yml down 2>/dev/null || true
    docker-compose -f docker-compose.backend.yml down 2>/dev/null || true

# View Docker logs
docker-logs:
    docker-compose -f docker-compose.split.yml logs -f

# Update proxy-data without downtime
docker-update-data:
    @echo "Updating proxy-data (zero-downtime)..."
    docker-compose -f docker-compose.split.yml up -d --no-deps --build proxy-data

# Update proxy-core (rare)
docker-update-core:
    @echo "Updating proxy-core..."
    docker-compose -f docker-compose.split.yml up -d --no-deps --build proxy-core

# Clean Docker resources
docker-clean:
    @echo "Cleaning Docker resources..."
    docker-compose -f docker-compose.split.yml down -v --rmi local 2>/dev/null || true
    docker-compose -f docker-compose.dev.yml down -v --rmi local 2>/dev/null || true
    docker-compose -f docker-compose.backend.yml down -v --rmi local 2>/dev/null || true

# ============================================================================
# Testing
# ============================================================================

# Run proxy tests
test-proxy:
    cd proxy && go test -tags "fts5" ./...

# Run web tests
test-web:
    cd web && pnpm run test

# Type check all TypeScript
typecheck:
    cd web && pnpm run typecheck
    cd dashboard && pnpm run build --mode development

# Lint all code
lint:
    cd web && pnpm run lint
    cd dashboard && pnpm run lint

# ============================================================================
# Cleanup
# ============================================================================

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf web/build/
    rm -rf web/.cache/
    rm -rf dashboard/dist/

# Reset database
db-reset:
    @echo "Resetting database..."
    rm -f requests.db
    rm -f data/requests.db
    rm -rf requests/

# Full clean (including Docker)
clean-all: clean docker-clean
    @echo "Full cleanup complete"

# ============================================================================
# Help
# ============================================================================

# Show help
help:
    @echo "Claude Code Proxy - Available commands:"
    @echo ""
    @echo "  Setup:"
    @echo "    just install          - Install all dependencies"
    @echo ""
    @echo "  Build (Local):"
    @echo "    just build            - Build all services (monolith + web)"
    @echo "    just build-proxy      - Build monolith proxy"
    @echo "    just build-split      - Build split services (proxy-core + proxy-data)"
    @echo "    just build-web        - Build web interface"
    @echo "    just build-dashboard  - Build dashboard"
    @echo ""
    @echo "  Run (Local):"
    @echo "    just dev              - Run monolith (proxy + web + dashboard)"
    @echo "    just dev-split        - Run split (Caddy + services + dashboards)"
    @echo "    just run-proxy        - Run monolith proxy only"
    @echo "    just run-proxy-core   - Run proxy-core only"
    @echo "    just run-proxy-data   - Run proxy-data only"
    @echo "    just run-web          - Run web dashboard only"
    @echo "    just run-dashboard    - Run new dashboard only"
    @echo ""
    @echo "  Docker:"
    @echo "    just docker-build     - Build Docker images"
    @echo "    just docker-up        - Run split services in Docker"
    @echo "    just docker-dev       - Run full dev stack (with HMR)"
    @echo "    just docker-backend   - Run backend only in Docker"
    @echo "    just docker-hybrid    - Docker backend + local frontends (best DX)"
    @echo "    just docker-down      - Stop all Docker services"
    @echo "    just docker-logs      - View Docker logs"
    @echo "    just docker-update-data - Update proxy-data (zero-downtime)"
    @echo "    just docker-clean     - Clean Docker resources"
    @echo ""
    @echo "  Testing:"
    @echo "    just test-proxy       - Run proxy tests"
    @echo "    just test-web         - Run web tests"
    @echo "    just typecheck        - Type check TypeScript"
    @echo "    just lint             - Lint all code"
    @echo ""
    @echo "  Cleanup:"
    @echo "    just clean            - Clean build artifacts"
    @echo "    just clean-all        - Clean everything (including Docker)"
    @echo "    just db-reset         - Reset database"
