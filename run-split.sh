#!/bin/bash

# Claude Code Monitor - Split Architecture Run Script
# Runs: Caddy (port 8000) + proxy-core (port 8001) + proxy-data (port 8002) + web dashboards

set -e

echo "Claude Code Monitor - Split Architecture"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go 1.20 or higher.${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Node.js is not installed. Please install Node.js 18 or higher.${NC}"
    exit 1
fi

# Check if Caddy is installed
if ! command -v caddy &> /dev/null; then
    echo -e "${RED}Caddy is not installed. Please install Caddy:${NC}"
    echo "  brew install caddy  # macOS"
    echo "  apt install caddy   # Debian/Ubuntu"
    echo "  See: https://caddyserver.com/docs/install"
    exit 1
fi

# Check for .env file
if [ ! -f .env ]; then
    echo -e "${YELLOW}No .env file found. Creating from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}Created .env file.${NC}"
    else
        echo "No .env.example file found."
        exit 1
    fi
fi

# Create temp directory for logs
LOGDIR=$(mktemp -d)
trap "rm -rf $LOGDIR" EXIT

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    kill $CADDY_PID $PROXY_CORE_PID $PROXY_DATA_PID $WEB_PID $DASHBOARD_PID 2>/dev/null || true
    rm -rf "$LOGDIR"
    exit
}

trap cleanup EXIT INT TERM

# Function to extract URL from Vite output
wait_for_vite_url() {
    local logfile="$1"
    local service_name="$2"
    local max_wait=30
    local waited=0

    while [ $waited -lt $max_wait ]; do
        if [ -f "$logfile" ]; then
            local url=$(grep -o 'Local:   http://[^[:space:]]*' "$logfile" 2>/dev/null | sed 's/Local:   //' | head -1)
            if [ -n "$url" ]; then
                echo "$url"
                return 0
            fi
        fi
        sleep 0.5
        waited=$((waited + 1))
    done

    echo "http://localhost:????"
    return 1
}

# Build proxy-core and proxy-data
echo -e "\n${BLUE}Building proxy-core...${NC}"
cd proxy
go mod download
go build -tags "fts5" -o ../bin/proxy-core cmd/proxy-core/main.go
echo -e "${GREEN}proxy-core built${NC}"

echo -e "${BLUE}Building proxy-data...${NC}"
go build -tags "fts5" -o ../bin/proxy-data cmd/proxy-data/main.go
cd ..
echo -e "${GREEN}proxy-data built${NC}"

# Install web dependencies if needed
if [ ! -d "web/node_modules" ]; then
    echo -e "\n${BLUE}Installing web dependencies...${NC}"
    cd web
    pnpm install
    cd ..
    echo -e "${GREEN}Web dependencies installed${NC}"
fi

# Install dashboard dependencies if needed
if [ ! -d "dashboard/node_modules" ]; then
    echo -e "\n${BLUE}Installing dashboard dependencies...${NC}"
    cd dashboard
    pnpm install
    cd ..
    echo -e "${GREEN}Dashboard dependencies installed${NC}"
fi

# Start Caddy reverse proxy
echo -e "\n${BLUE}Starting Caddy reverse proxy...${NC}"
caddy run --config Caddyfile > "$LOGDIR/caddy.log" 2>&1 &
CADDY_PID=$!
sleep 1

# Start proxy-core
echo -e "${BLUE}Starting proxy-core (port 8001)...${NC}"
./bin/proxy-core > "$LOGDIR/proxy-core.log" 2>&1 &
PROXY_CORE_PID=$!

# Wait for proxy-core to start
sleep 2

# Start proxy-data
echo -e "${BLUE}Starting proxy-data (port 8002)...${NC}"
./bin/proxy-data > "$LOGDIR/proxy-data.log" 2>&1 &
PROXY_DATA_PID=$!

# Wait for proxy-data to start
sleep 2

# Start legacy web server
echo -e "${BLUE}Starting web interface...${NC}"
cd web
pnpm run dev > "$LOGDIR/web.log" 2>&1 &
WEB_PID=$!
cd ..

# Start new dashboard
echo -e "${BLUE}Starting new dashboard...${NC}"
cd dashboard
pnpm run dev > "$LOGDIR/dashboard.log" 2>&1 &
DASHBOARD_PID=$!
cd ..

# Wait for Vite servers to be ready and extract URLs
echo -e "${BLUE}Waiting for services to be ready...${NC}"
WEB_URL=$(wait_for_vite_url "$LOGDIR/web.log" "web")
DASHBOARD_URL=$(wait_for_vite_url "$LOGDIR/dashboard.log" "dashboard")

echo -e "\n${GREEN}All services started!${NC}"
echo "========================================="
echo -e "Architecture:     ${YELLOW}Split Services${NC}"
echo ""
echo -e "Unified Proxy:    ${BLUE}http://localhost:8000${NC} (Caddy)"
echo -e "  -> /v1/*        ${BLUE}http://localhost:8001${NC} (proxy-core)"
echo -e "  -> /api/*       ${BLUE}http://localhost:8002${NC} (proxy-data)"
echo ""
echo -e "Web Dashboard:    ${BLUE}${WEB_URL}${NC}"
echo -e "New Dashboard:    ${BLUE}${DASHBOARD_URL}${NC}"
echo ""
echo -e "Health Checks:"
echo -e "  Caddy:          ${BLUE}http://localhost:8000/health${NC}"
echo -e "  proxy-core:     ${BLUE}http://localhost:8001/health${NC}"
echo -e "  proxy-data:     ${BLUE}http://localhost:8002/health${NC}"
echo "========================================="
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}\n"

# Tail the logs so user can see output
tail -f "$LOGDIR/caddy.log" "$LOGDIR/proxy-core.log" "$LOGDIR/proxy-data.log" "$LOGDIR/web.log" "$LOGDIR/dashboard.log" &
TAIL_PID=$!

# Wait for processes
wait $CADDY_PID $PROXY_CORE_PID $PROXY_DATA_PID $WEB_PID $DASHBOARD_PID
