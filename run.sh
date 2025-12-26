#!/bin/bash

# Claude Code Monitor - Build and Run Script

set -e

echo "Claude Code Monitor - Starting Services"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.20 or higher."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Node.js is not installed. Please install Node.js 18 or higher."
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
    kill $PROXY_PID $WEB_PID $DASHBOARD_PID 2>/dev/null || true
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
            # Look for the Local URL line
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

# Build and start proxy server
echo -e "\n${BLUE}Building proxy server...${NC}"
cd proxy
go mod download
go build -o ../bin/proxy cmd/proxy/main.go
cd ..

echo -e "${GREEN}Proxy server built${NC}"

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

# Start proxy server
echo -e "\n${BLUE}Starting proxy server...${NC}"
./bin/proxy > "$LOGDIR/proxy.log" 2>&1 &
PROXY_PID=$!

# Wait for proxy to start
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
echo -e "Web Dashboard:  ${BLUE}${WEB_URL}${NC}"
echo -e "New Dashboard:  ${BLUE}${DASHBOARD_URL}${NC}"
echo -e "API Proxy:      ${BLUE}http://localhost:3001${NC}"
echo -e "Health Check:   ${BLUE}http://localhost:3001/health${NC}"
echo "========================================="
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}\n"

# Tail the logs so user can see output
tail -f "$LOGDIR/proxy.log" "$LOGDIR/web.log" "$LOGDIR/dashboard.log" &
TAIL_PID=$!

# Wait for processes
wait $PROXY_PID $WEB_PID $DASHBOARD_PID
