# Claude Code Proxy

# Default: show available commands
default:
    @just --list

# Install dependencies (first-time setup)
install:
    cd proxy && go mod download
    cd web && pnpm install
    cd dashboard && pnpm install

# Build everything (Go binaries + web assets)
build:
    @echo "Building..."
    cd proxy && go build -tags "fts5" -o ../bin/proxy-core cmd/proxy-core/main.go
    cd proxy && go build -tags "fts5" -o ../bin/proxy-data cmd/proxy-data/main.go
    cd web && pnpm run build
    cd dashboard && pnpm run build
    @echo "Done"

# Run in development mode (Caddy + proxy-core + proxy-data + dashboards)
run:
    ./run-split.sh

# Run with Docker (backend in containers, frontends local for HMR)
docker:
    docker-compose -f docker-compose.backend.yml up -d --build
    @echo ""
    @echo "Backend running. Starting frontends..."
    @echo "  API:       http://localhost:3000"
    @echo "  Web:       http://localhost:5173"
    @echo "  Dashboard: http://localhost:5174"
    @echo ""
    cd web && pnpm dev & cd dashboard && pnpm dev

# Stop Docker services
stop:
    docker-compose -f docker-compose.backend.yml down
    docker-compose -f docker-compose.split.yml down 2>/dev/null || true

# Restart data service (zero-downtime update)
restart-data:
    docker-compose -f docker-compose.backend.yml up -d --no-deps --build proxy-data

# Start just Plano service
plano-up:
    docker-compose -f docker-compose.backend.yml up -d plano

# View Plano logs
plano-logs:
    docker-compose -f docker-compose.backend.yml logs -f plano

# Restart Plano service
plano-restart:
    docker-compose -f docker-compose.backend.yml restart plano

# Run all tests
test:
    cd proxy && go test -tags "fts5" ./...
    cd web && pnpm test

# Lint and type check
check:
    cd web && pnpm run typecheck && pnpm run lint
    cd dashboard && pnpm run lint

# Reset database
db:
    rm -f requests.db data/requests.db
    @echo "Database reset"

# Clean all build artifacts and Docker resources
clean:
    rm -rf bin/ web/build/ web/.cache/ dashboard/dist/
    docker-compose -f docker-compose.backend.yml down -v 2>/dev/null || true
    docker-compose -f docker-compose.split.yml down -v 2>/dev/null || true
    @echo "Cleaned"
