# Claude Code Proxy

![Claude Code Proxy Demo](demo.gif)

A transparent proxy for capturing and visualizing in-flight Claude Code requests and conversations, with optional agent routing to different LLM providers.

## What It Does

Claude Code Proxy serves three main purposes:

1. **Claude Code Proxy**: Intercepts and monitors requests from Claude Code (claude.ai/code) to the Anthropic API, allowing you to see what Claude Code is doing in real-time
2. **Conversation Viewer**: Displays and analyzes your Claude API conversations with a beautiful web interface
3. **Agent Routing (Optional)**: Routes specific Claude Code agents to different LLM providers (e.g., route code-reviewer agent to GPT-4o)

## Features

- **Transparent Proxy**: Routes Claude Code requests through the monitor without disruption
- **Agent Routing (Optional)**: Map specific Claude Code agents to different LLM models
- **Request Monitoring**: SQLite-based logging of all API interactions
- **Live Dashboard**: Real-time visualization of requests and responses
- **Conversation Analysis**: View full conversation threads with tool usage
- **Easy Setup**: One-command startup for both services

## Quick Start

### Prerequisites
- **Option 1**: Go 1.20+ and Node.js 18+ (for local development)
- **Option 2**: Docker (for containerized deployment)
- Claude Code

### Installation

#### Option 1: Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/seifghazi/claude-code-proxy.git
   cd claude-code-proxy
   ```

2. **Configure the proxy**
   ```bash
   cp config.yaml.example config.yaml
   ```

3. **Install and run** (first time)
   ```bash
   make install  # Install all dependencies
   make dev      # Start both services
   ```

4. **Subsequent runs** (after initial setup)
   ```bash
   make dev
   # or
   ./run.sh
   ```

#### Option 2: Docker

1. **Clone the repository**
   ```bash
   git clone https://github.com/seifghazi/claude-code-proxy.git
   cd claude-code-proxy
   ```

2. **Configure the proxy**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml as needed
   ```

3. **Build and run with Docker**
   ```bash
   # Build the image
   docker build -t claude-code-proxy .
   
   # Run with default settings
   docker run -p 3001:3001 -p 5173:5173 claude-code-proxy
   ```

4. **Run with persistent data and custom configuration**
   ```bash
   # Create a data directory for persistent SQLite database
   mkdir -p ./data
   
   # Option 1: Run with config file (recommended)
   docker run -p 3001:3001 -p 5173:5173 \
     -v ./data:/app/data \
     -v ./config.yaml:/app/config.yaml:ro \
     claude-code-proxy
   
   # Option 2: Run with environment variables
   docker run -p 3001:3001 -p 5173:5173 \
     -v ./data:/app/data \
     -e ANTHROPIC_FORWARD_URL=https://api.anthropic.com \
     -e PORT=3001 \
     -e WEB_PORT=5173 \
     claude-code-proxy
   ```

5. **Docker Compose (alternative)**
   ```yaml
   # docker-compose.yml
   version: '3.8'
   services:
     claude-code-proxy:
       build: .
       ports:
         - "3001:3001"
         - "5173:5173"
       volumes:
         - ./data:/app/data
         - ./config.yaml:/app/config.yaml:ro  # Mount config file
       environment:
         - ANTHROPIC_FORWARD_URL=https://api.anthropic.com
         - PORT=3001
         - WEB_PORT=5173
         - DB_PATH=/app/data/requests.db
   ```

   Then run: `docker-compose up`

#### Option 3: Docker Split Architecture (Production)

For zero-downtime deployments, use the split architecture with separate containers:

1. **Clone and configure**
   ```bash
   git clone https://github.com/seifghazi/claude-code-proxy.git
   cd claude-code-proxy
   cp config.yaml.example config.yaml
   mkdir -p data
   ```

2. **Run split services with Docker Compose**
   ```bash
   # Build and start all services
   docker-compose -f docker-compose.split.yml up --build

   # Run in detached mode
   docker-compose -f docker-compose.split.yml up -d
   ```

3. **Zero-downtime updates**
   ```bash
   # Update proxy-data without affecting the main proxy
   docker-compose -f docker-compose.split.yml up -d --no-deps --build proxy-data

   # Update proxy-core (rare, but when needed)
   docker-compose -f docker-compose.split.yml up -d --no-deps --build proxy-core
   ```

4. **Access points (Docker split mode)**
   - **Unified Endpoint**: http://localhost:3000 (Caddy)
   - **Health Check**: http://localhost:3000/health

5. **Full development stack** (includes web dashboards)
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```
   - Web Dashboard: http://localhost:5173
   - New Dashboard: http://localhost:5174

### Using with Claude Code

To use this proxy with Claude Code, set:
```bash
export ANTHROPIC_BASE_URL=http://localhost:3001
```

Then launch Claude Code using the `claude` command.

This will route Claude Code's requests through the proxy for monitoring.

### Access Points
- **Web Dashboard**: http://localhost:5173
- **API Proxy**: http://localhost:3001
- **Health Check**: http://localhost:3001/health

## Advanced Usage

### Running Services Separately

If you need to run services independently:

```bash
# Run proxy only
make run-proxy

# Run web interface only (in another terminal)
make run-web
```

### Split Architecture (Zero-Downtime Deployments)

For production deployments requiring zero-downtime updates, the proxy can be split into two services:

- **proxy-core** (port 3001): Lightweight proxy that handles API forwarding and data recording. Rarely needs updates.
- **proxy-data** (port 3002): Dashboard APIs, statistics, conversation indexing. Can be updated independently.

**Running Split Architecture:**

```bash
# Build split services
make build-split

# Run with Caddy reverse proxy (unified endpoint on port 3000)
make dev-split
# or
./run-split.sh
```

**Access Points (Split Mode):**
- **Unified Endpoint**: http://localhost:3000 (Caddy routes to appropriate service)
- **proxy-core direct**: http://localhost:3001 (API proxy)
- **proxy-data direct**: http://localhost:3002 (Dashboard APIs)
- **Web Dashboard**: http://localhost:5173

**Requirements:**
- [Caddy](https://caddyserver.com/docs/install) for reverse proxy routing

### Available Make Commands

```bash
# Setup
make install        # Install all dependencies

# Build
make build          # Build monolith proxy + web
make build-proxy    # Build monolith proxy only
make build-split    # Build split services (proxy-core + proxy-data)

# Run (Development)
make dev            # Run monolith (proxy + web + dashboard)
make dev-split      # Run split architecture (Caddy + services + web)
make run-proxy      # Run monolith proxy only
make run-proxy-core # Run proxy-core only
make run-proxy-data # Run proxy-data only

# Cleanup
make clean          # Clean build artifacts
make db-reset       # Reset database
make help           # Show all commands
```

## Configuration

### Basic Setup

Create a `config.yaml` file (or copy from `config.yaml.example`):
```yaml
server:
  port: 3001

providers:
  anthropic:
    base_url: "https://api.anthropic.com"
    
  openai: # if enabling subagent routing
    api_key: "your-openai-key"  # Or set OPENAI_API_KEY env var

storage:
  db_path: "requests.db"
```

### Subagent Configuration (Optional)

The proxy supports routing specific Claude Code agents to different LLM providers. This is an **optional** feature that's disabled by default.

#### Enabling Subagent Routing

1. **Enable the feature** in `config.yaml`:
```yaml
subagents:
  enable: true  # Set to true to enable subagent routing
  mappings:
    code-reviewer: "gpt-4o"
    data-analyst: "o3"
    doc-writer: "gpt-3.5-turbo"
```

2. **Set up your Claude Code agents** following Anthropic's official documentation:
   - 📖 **[Claude Code Subagents Documentation](https://docs.anthropic.com/en/docs/claude-code/sub-agents)**

3. **How it works**: When Claude Code uses a subagent that matches one of your mappings, the proxy will automatically route the request to the specified model instead of Claude.

### Practical Examples

**Example 1: Code Review Agent → GPT-4o**
```yaml
# config.yaml
subagents:
  enable: true
  mappings:
    code-reviewer: "gpt-4o"
```
Use case: Route code review tasks to GPT-4o for faster responses while keeping complex coding tasks on Claude.

**Example 2: Reasoning Agent → O3**  
```yaml
# config.yaml
subagents:
  enable: true
  mappings:
    deep-reasoning: "o3"
```
Use case: Send complex reasoning tasks to O3 while using Claude for general coding.

**Example 3: Multiple Agents**
```yaml
# config.yaml
subagents:
  enable: true
  mappings:
    streaming-systems-engineer: "o3"
    frontend-developer: "gpt-4o-mini"
    security-auditor: "gpt-4o"
```
Use case: Different specialists for different tasks, optimizing for speed/cost/quality.

### Environment Variables

Override config via environment:
- `PORT` - Server port
- `OPENAI_API_KEY` - OpenAI API key
- `DB_PATH` - Database path
- `SUBAGENT_MAPPINGS` - Comma-separated mappings (e.g., `"code-reviewer:gpt-4o,data-analyst:o3"`)

### Docker Environment Variables

All environment variables can be configured when running the Docker container:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3001` | Proxy server port |
| `WEB_PORT` | `5173` | Web dashboard port |
| `READ_TIMEOUT` | `600` | Server read timeout (seconds) |
| `WRITE_TIMEOUT` | `600` | Server write timeout (seconds) |
| `IDLE_TIMEOUT` | `600` | Server idle timeout (seconds) |
| `ANTHROPIC_FORWARD_URL` | `https://api.anthropic.com` | Target Anthropic API URL |
| `ANTHROPIC_VERSION` | `2023-06-01` | Anthropic API version |
| `ANTHROPIC_MAX_RETRIES` | `3` | Maximum retry attempts |
| `DB_PATH` | `/app/data/requests.db` | SQLite database path |

Example with custom configuration:
```bash
docker run -p 3001:3001 -p 5173:5173 \
  -v ./data:/app/data \
  -e PORT=8080 \
  -e WEB_PORT=3000 \
  -e ANTHROPIC_FORWARD_URL=https://api.anthropic.com \
  -e DB_PATH=/app/data/custom.db \
  claude-code-proxy
```


## Project Structure

```
claude-code-proxy/
├── proxy/                  # Go proxy server
│   ├── cmd/
│   │   ├── proxy/         # Monolith entry point
│   │   ├── proxy-core/    # Split: lightweight proxy (API forwarding)
│   │   └── proxy-data/    # Split: data service (dashboard APIs)
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── handlers.go      # Monolith handler
│   │   │   ├── core_handler.go  # Split: proxy-core handlers
│   │   │   └── data_handler.go  # Split: proxy-data handlers
│   │   ├── service/
│   │   ├── config/
│   │   └── provider/
│   └── go.mod
├── web/                   # React Remix frontend
├── dashboard/             # New React dashboard
├── docker/                # Docker configurations
│   ├── Dockerfile.proxy-core   # proxy-core container
│   ├── Dockerfile.proxy-data   # proxy-data container
│   └── Caddyfile              # Caddy config for Docker
├── Dockerfile            # Monolith container
├── docker-compose.split.yml   # Split architecture (production)
├── docker-compose.dev.yml     # Full dev stack with dashboards
├── Caddyfile             # Reverse proxy config (local)
├── run.sh                # Start script (monolith)
├── run-split.sh          # Start script (split architecture)
├── config.yaml.example   # Configuration template
└── README.md
```

## Features in Detail

### Request Monitoring
- All API requests logged to SQLite database
- Searchable request history
- Request/response body inspection
- Conversation threading

### Web Dashboard
- Real-time request streaming
- Interactive request explorer
- Conversation visualization
- Performance metrics

## License

MIT License - see [LICENSE](LICENSE) for details.
