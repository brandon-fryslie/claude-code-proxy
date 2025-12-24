# Claude Code Proxy - Task runner

# Default recipe
default: help

# Install all dependencies
install:
    @echo "Installing Go dependencies..."
    cd proxy && go mod download
    @echo "Installing web dependencies..."
    cd web && pnpm install
    @echo "Installing dashboard dependencies..."
    cd dashboard && pnpm install

# Build all services
build: build-proxy build-web build-dashboard

# Build proxy server
build-proxy:
    @echo "Building proxy server..."
    cd proxy && go build -o ../bin/proxy cmd/proxy/main.go

# Build legacy web interface
build-web:
    @echo "Building web interface..."
    cd web && pnpm run build

# Build new dashboard
build-dashboard:
    @echo "Building dashboard..."
    cd dashboard && pnpm run build

# Run everything in development mode (proxy + web + dashboard)
dev:
    ./run.sh

# Run proxy only
run-proxy:
    cd proxy && go run cmd/proxy/main.go

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
    rm -rf requests/

# Type check all TypeScript
typecheck:
    cd web && pnpm run typecheck
    cd dashboard && pnpm run build --mode development

# Lint all code
lint:
    cd web && pnpm run lint
    cd dashboard && pnpm run lint

# Run proxy tests
test-proxy:
    cd proxy && go test ./...

# Run web tests
test-web:
    cd web && pnpm run test

# Show help
help:
    @echo "Claude Code Proxy - Available commands:"
    @echo ""
    @echo "  just install     - Install all dependencies"
    @echo "  just build       - Build all services"
    @echo "  just dev         - Run everything (proxy + web + dashboard)"
    @echo "  just run-proxy   - Run proxy server only"
    @echo "  just clean       - Clean build artifacts"
    @echo "  just db-reset    - Reset database"
    @echo "  just typecheck   - Type check TypeScript"
    @echo "  just lint        - Lint all code"
    @echo "  just test-proxy  - Run proxy tests"
    @echo "  just test-web    - Run web tests"
