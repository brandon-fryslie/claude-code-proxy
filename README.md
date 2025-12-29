# Claude Code Proxy

![Claude Code Proxy Demo](demo.gif)

A transparent proxy for capturing and visualizing in-flight Claude Code requests and conversations, with optional agent routing to different LLM providers.

## What It Does

Claude Code Proxy serves three main purposes:

1. **Claude Code Proxy**: Intercepts and monitors requests from Claude Code (claude.ai/code) to the Anthropic API, allowing you to see what Claude Code is doing in real-time
2. **Conversation Viewer**: Displays and analyzes your Claude API conversations with a beautiful web interface
3. **Agent Routing (Optional)**: Routes specific Claude Code agents to different LLM providers (e.g., route code-reviewer agent to GPT-4o, or planner agent to Gemini via Plano)

## Features

- **Transparent Proxy**: Routes Claude Code requests through the monitor without disruption
- **Agent Routing (Optional)**: Map specific Claude Code agents to different LLM models
- **Multi-LLM Support via Plano**: Access Gemini, DeepSeek, Qwen, and more providers through integrated Plano router
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

1. **Clone and configure**
   ```bash
   git clone https://github.com/seifghazi/claude-code-proxy.git
   cd claude-code-proxy
   cp config.yaml.example config.yaml
   cp .env.example .env
   ```

2. **Configure API keys** (if using Plano for multi-LLM routing)

   Edit `.env` and add your provider API keys:
   ```bash
   OPENAI_API_KEY=sk-xxx
   GEMINI_API_KEY=your-gemini-key
   DEEPSEEK_API_KEY=sk-your-deepseek-key
   DASHSCOPE_API_KEY=sk-your-qwen-key
   ```

3. **Install dependencies**
   ```bash
   just install
   ```

4. **Run** (choose one)
   ```bash
   just run      # Local development (requires Caddy)
   just docker   # Docker backend + local frontends (best HMR)
   ```

### Prerequisites

- **[just](https://github.com/casey/just)** - Command runner (`brew install just`)
- **Go 1.20+** and **Node.js 18+** (for local dev)
- **[Caddy](https://caddyserver.com/docs/install)** (for `just run`)
- **Docker** (for `just docker`)

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

## Commands

All operations use `just`. Run `just` to see available commands.

```bash
just install       # Install dependencies (first-time setup)
just build         # Build everything (Go binaries + web assets)
just run           # Run in development mode (Caddy + services + dashboards)
just docker        # Run with Docker (backend containers + local frontends for HMR)
just stop          # Stop Docker services
just restart-data  # Restart data service (zero-downtime update)
just plano-up      # Start just Plano service
just plano-logs    # View Plano logs
just plano-restart # Restart Plano service
just test          # Run all tests
just check         # Lint and type check
just db            # Reset database
just clean         # Clean all build artifacts and Docker resources
```

## Architecture

The proxy runs as multiple services for zero-downtime deployments:

- **proxy-core** (port 3001): Lightweight API proxy. Rarely changes.
- **proxy-data** (port 3002): Dashboard APIs, statistics, indexing. Updated frequently.
- **plano** (port 8080): Multi-LLM router for intelligent provider routing (optional).

Caddy (port 3000) routes requests to the appropriate service.

**Access Points:**
- **API**: http://localhost:3000 (unified) or http://localhost:3001 (direct)
- **Web Dashboard**: http://localhost:5173
- **New Dashboard**: http://localhost:5174

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

  plano: # Optional: Multi-LLM router
    base_url: "http://localhost:8080"  # or http://plano:8080 in Docker
    format: "openai"

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
    code-reviewer: "openai:gpt-4o"
    planner: "plano:gemini/gemini-2.0-flash-exp"
    data-analyst: "plano:deepseek/deepseek-chat"
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
    code-reviewer: "openai:gpt-4o"
```
Use case: Route code review tasks to GPT-4o for faster responses while keeping complex coding tasks on Claude.

**Example 2: Planning Agent → Gemini via Plano**
```yaml
# config.yaml
providers:
  plano:
    base_url: "http://localhost:8080"
    format: "openai"

subagents:
  enable: true
  mappings:
    planner: "plano:gemini/gemini-2.0-flash-exp"
```
Use case: Route planning tasks to Gemini for cost-effective, fast responses.

**Example 3: Multiple Agents with Plano**
```yaml
# config.yaml
subagents:
  enable: true
  mappings:
    planner: "plano:gemini/gemini-2.0-flash-exp"
    code-generator: "plano:deepseek/deepseek-coder"
    security-auditor: "openai:gpt-4o"
    budget-helper: "plano:qwen/qwen-max"
```
Use case: Different specialists for different tasks, optimizing for speed/cost/quality.

### Plano Multi-LLM Router Setup

Plano enables routing to multiple LLM providers (Gemini, DeepSeek, Qwen, etc.) through a single endpoint.

#### Configuration

1. **Configure providers in `plano_config.yaml`** (already included):
   ```yaml
   version: v0.1.0

   listeners:
     - name: egress_traffic
       address: 0.0.0.0
       port: 8080
       message_format: openai

   llm_providers:
     - model: openai/gpt-4o
       access_key: $OPENAI_API_KEY

     - model: gemini/gemini-2.0-flash-exp
       access_key: $GEMINI_API_KEY

     - model: deepseek/deepseek-chat
       access_key: $DEEPSEEK_API_KEY
   ```

2. **Set API keys in `.env`**:
   ```bash
   GEMINI_API_KEY=your-gemini-api-key
   DEEPSEEK_API_KEY=sk-your-deepseek-key
   DASHSCOPE_API_KEY=sk-your-qwen-key
   ```

3. **Enable Plano in Docker Compose**:

   Plano is automatically included when using `just docker`. The service starts on port 8080 (internal to Docker network).

4. **Route subagents to Plano** in `config.yaml`:
   ```yaml
   providers:
     plano:
       base_url: "http://plano:8080"  # Docker service name
       format: "openai"

   subagents:
     enable: true
     mappings:
       planner: "plano:gemini/gemini-2.0-flash-exp"
       code-helper: "plano:deepseek/deepseek-coder"
   ```

#### Plano Commands

```bash
# Start Plano service
just plano-up

# View Plano logs
just plano-logs

# Restart Plano
just plano-restart

# Stop all services (including Plano)
just stop
```

### Environment Variables

Override config via environment:
- `PORT` - Server port
- `OPENAI_API_KEY` - OpenAI API key
- `DB_PATH` - Database path
- `SUBAGENT_MAPPINGS` - Comma-separated mappings (e.g., `"code-reviewer:gpt-4o,data-analyst:o3"`)

#### Plano Environment Variables

Required for Plano to route to upstream providers:
- `GEMINI_API_KEY` - Google Gemini API key
- `DEEPSEEK_API_KEY` - DeepSeek API key
- `DASHSCOPE_API_KEY` - Qwen (DashScope) API key

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
│   ├── Dockerfile.plano        # Plano multi-LLM router
│   └── Caddyfile              # Caddy config for Docker
├── Dockerfile            # Monolith container
├── docker-compose.split.yml   # Split architecture (production)
├── docker-compose.backend.yml # Backend only (development)
├── docker-compose.dev.yml     # Full dev stack with dashboards
├── Caddyfile             # Reverse proxy config (local)
├── run.sh                # Start script (monolith)
├── run-split.sh          # Start script (split architecture)
├── config.yaml.example   # Configuration template
├── plano_config.yaml     # Plano multi-LLM router config
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

### Multi-LLM Routing (via Plano)
- Access 11+ LLM providers: Gemini, DeepSeek, Qwen, Mistral, Groq, and more
- Cost-effective alternatives to Claude/GPT-4
- Seamless integration with subagent routing
- Automatic format conversion (Anthropic ↔ OpenAI)

## License

MIT License - see [LICENSE](LICENSE) for details.
