.PHONY: all build run clean install dev

# Default target
all: install build

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	cd proxy && go mod download
	@echo "Installing Node dependencies..."
	cd web && npm install

# Build all services
build: build-proxy build-web

# Build the monolith proxy (legacy)
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

build-web:
	@echo "Building web interface..."
	cd web && npm run build

# Run in development mode (monolith)
dev:
	@echo "Starting development servers..."
	./run.sh

# Run split architecture in development mode
dev-split:
	@echo "Starting split architecture development servers..."
	./run-split.sh

# Run proxy only (monolith)
run-proxy:
	cd proxy && go run -tags "fts5" cmd/proxy/main.go

# Run split services individually
run-proxy-core:
	cd proxy && go run -tags "fts5" cmd/proxy-core/main.go

run-proxy-data:
	cd proxy && go run -tags "fts5" cmd/proxy-data/main.go

# Run web only
run-web:
	cd web && npm run dev

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/build/
	rm -rf web/.cache/
	rm -f requests.db
	rm -rf requests/

# Database operations
db-reset:
	@echo "Resetting database..."
	rm -f requests.db
	rm -rf requests/

# Help
help:
	@echo "Claude Code Monitor - Available targets:"
	@echo ""
	@echo "  Setup:"
	@echo "    make install       - Install all dependencies"
	@echo ""
	@echo "  Build:"
	@echo "    make build         - Build monolith proxy + web"
	@echo "    make build-proxy   - Build monolith proxy"
	@echo "    make build-split   - Build split services (proxy-core + proxy-data)"
	@echo "    make build-web     - Build web interface"
	@echo ""
	@echo "  Run (Development):"
	@echo "    make dev           - Run monolith (proxy + web + dashboard)"
	@echo "    make dev-split     - Run split architecture (Caddy + proxy-core + proxy-data + web)"
	@echo "    make run-proxy     - Run monolith proxy only"
	@echo "    make run-proxy-core- Run proxy-core only"
	@echo "    make run-proxy-data- Run proxy-data only"
	@echo "    make run-web       - Run web interface only"
	@echo ""
	@echo "  Cleanup:"
	@echo "    make clean         - Clean build artifacts"
	@echo "    make db-reset      - Reset database"
	@echo "    make help          - Show this help message"