# Project Roadmap

**Goal:** Achieve feature parity between old dashboard (web/) and new dashboard (dashboard/)

**Last Updated:** 2026-01-07

---

## Phase 1: Core Navigation & Data [COMPLETED]

**Goal:** Enable efficient browsing and filtering of large request datasets

### Topics

- **date-navigation-filtering** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-kbm`
  - ✅ Week boundary calculation (Sunday-Saturday)
  - ✅ Date picker with native input
  - ✅ Week navigation (prev/next/today buttons)
  - ✅ Date persistence (localStorage + URL params)
  - ✅ Smart refetching (same week = 2 calls, diff week = 3 calls)

- **virtualized-request-list** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-pyf`
  - ✅ @tanstack/react-virtual installed and configured
  - ✅ Handles 1000+ requests with ~30 visible DOM nodes
  - ✅ Smooth scrolling (60fps)
  - ✅ Lazy loading on item click

- **model-filter-dropdown** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-jx0`
  - ✅ Dropdown: All Models, Opus, Sonnet, Haiku
  - ✅ URL parameter persistence (?model=opus)
  - ✅ Backend filtering integrated

- **conversation-search-indexing** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-d0o`
  - ✅ SQLite FTS5 implementation with 1973 conversations indexed
  - ✅ File watcher for ~/.claude/projects/ indexing
  - ✅ `/api/conversations/search?q=...` endpoint with pagination
  - ✅ Comprehensive E2E tests (22 test cases)

---

## Phase 2: Rich Content Display [COMPLETED]

**Goal:** Render message content with proper formatting and interactivity

### Topics

- **message-content-parser** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-kqw`
  - ✅ MessageContent.tsx (373 lines) with type dispatch
  - ✅ Render all Anthropic content block types (text, tool_use, tool_result, image)
  - ✅ Multi-part content with recursive rendering
  - ✅ SystemReminder extraction (collapsed by default)
  - ✅ FunctionDefinitions with tool count display

- **code-viewer-highlighting** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-6j7`
  - ✅ CodeViewer component with syntax highlighting
  - ✅ Support JS/TS, Python, Go, Rust, Bash, SQL, JSON, HTML, CSS
  - ✅ Line numbers, copy button, download, fullscreen modal
  - ✅ Handles 2000+ line files smoothly

- **copy-to-clipboard** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-nhw`
  - ✅ useCopyToClipboard hook with navigator.clipboard fallback
  - ✅ CopyButton component with visual feedback
  - ✅ Integrated across Requests, CodeViewer, ToolUse

- **conversation-content-integration** [PROPOSED]
  - Deps: message-content-parser
  - Use MessageContent in ConversationsPage detail view
  - Render actual message content (not just metadata)
  - Show user/assistant messages with formatting

- **header-copy-buttons** [PROPOSED]
  - Deps: copy-to-clipboard
  - Add CopyButton to request headers in detail view
  - Copy individual header values

---

## Phase 3: Tool Support [COMPLETED]

**Goal:** Display tool usage with expandable, formatted content

### Topics

- **tool-use-display** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-62s`
  - ✅ ToolUseContainer with expandable parameters
  - ✅ Special renderers: BashTool, ReadTool, WriteTool, EditTool, etc.
  - ✅ Parameter count badge, copy tool ID
  - ✅ Complex object expansion, large string truncation

- **tool-result-display** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-k6y`
  - ✅ ToolResultContent with content type detection
  - ✅ Code detection + CodeViewer integration
  - ✅ JSON formatting, cat -n line number extraction
  - ✅ Success (green) / Error (red) styling with icons

- **image-content-display** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-asz`
  - ✅ ImageContent with data URI construction
  - ✅ Download button, fullscreen modal
  - ✅ Error handling for missing/corrupt data
  - ✅ Media type label display

---

## Phase 4: Charts & Analytics [QUEUED]

**Goal:** Comprehensive usage visualization and metrics

### Topics

- **weekly-usage-chart** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-loz`
  - Add 7-day bar chart (like old dashboard)
  - Show daily totals
  - Color by model

- **model-breakdown-stats** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-xwb`
  - Add model breakdown endpoint usage
  - Show per-model token usage on dashboard
  - Pie chart or bar breakdown

- **performance-metrics** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-eee`
  - Enhanced percentile displays
  - Response time trends
  - Latency distribution

---

## Phase 5: Advanced Features [QUEUED]

**Goal:** Power user features for deep analysis

### Topics

- **request-comparison** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-wa1`
  - Multi-select requests
  - Side-by-side diff view
  - Compare tokens, timing, content

- **conversation-threads** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-jaq`
  - Full message thread display
  - User/Assistant message bubbles
  - Tool calls inline

- **data-management** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-5jo`
  - Add refresh button
  - Add clear all requests button
  - Settings for auto-refresh

- **web-routing-configuration** [PROPOSED]
  - Configure subagent-to-provider routing from web UI
  - View/edit config.yaml subagent mappings
  - Add/remove/modify routing rules
  - Persist changes to config file
  - Reload proxy configuration

- **multi-provider-comparison** [PROPOSED]
  - Send same prompt to multiple providers simultaneously
  - Side-by-side response comparison view
  - Compare response quality, latency, token usage
  - Select providers/models to include in comparison
  - Save comparison results for later review

- **vibeproxy-feature-analysis** [PROPOSED]
  - Research: Analyze VibeProxy features for potential adoption
  - Compare provider support (Gemini, Qwen, Antigravity, GitHub Copilot)
  - Evaluate OAuth authentication approach
  - Assess multi-account round-robin/failover
  - Document feature gap analysis

- **oauth-authentication** [PROPOSED]
  - OAuth login to use existing AI subscriptions (ChatGPT Plus, Claude Pro, etc.)
  - Browser-based OAuth flow with callback handler
  - Token storage and refresh management
  - Dashboard OAuth account management UI
  - Support for: OpenAI, Anthropic, Google (Gemini)

---

## Phase 6: Testing & Polish [QUEUED]

**Goal:** Ensure reliability and production readiness

### Topics

- **component-unit-tests** [PROPOSED]
  - Add Vitest tests for Phase 2/3 components
  - Test MessageContent with various content types
  - Test CodeViewer highlighting
  - Test copy functionality

- **accessibility-improvements** [PROPOSED]
  - Keyboard navigation for all interactive elements
  - ARIA labels for copy buttons
  - Focus management in fullscreen mode

---

## State Legend

- **PROPOSED** - Identified but not started
- **PLANNING** - STATUS/PLAN files created
- **IN PROGRESS** - Implementation underway
- **COMPLETED** - All acceptance criteria met
- **ARCHIVED** - No longer maintained
