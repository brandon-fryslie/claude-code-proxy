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

# Run with podman (complete dev stack with HMR)
dev:
    podman-compose up --build

# Stop podman services
stop:
    podman-compose down

# Restart data service (zero-downtime update)
restart-data:
    podman-compose up -d --no-deps --build proxy-data

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

# Clean all build artifacts and podman resources
clean:
    rm -rf bin/ web/build/ web/.cache/ dashboard/dist/
    podman-compose down -v 2>/dev/null || true
    @echo "Cleaned"
