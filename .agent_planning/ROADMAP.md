# Project Roadmap

**Goal:** Achieve feature parity between old dashboard (web/) and new dashboard (dashboard/)

**Last Updated:** 2025-12-26

---

## Phase 1: Core Navigation & Data [ACTIVE]

**Goal:** Enable efficient browsing and filtering of large request datasets

### Topics

- **date-navigation-filtering** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-kbm`
  - Add date picker component
  - Add week navigation (prev/next)
  - Persist selected date across pages
  - Filter requests by date range

- **virtualized-request-list** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-pyf`
  - Implement @tanstack/react-virtual for request list
  - Handle 1000s of requests efficiently
  - Lazy loading as user scrolls

- **model-filter-dropdown** [PROPOSED]
  - Epic: `brandon-fryslie_claude-code-proxy-jx0`
  - Add model filter to request list
  - Add endpoint filter option

---

## Phase 2: Rich Content Display [ACTIVE]

**Goal:** Render message content with proper formatting and interactivity

### Topics

- **message-content-parser** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-kqw`
  - ✅ Parse and render Anthropic message content blocks
  - ✅ Display text with proper formatting
  - ✅ Handle multi-part content (text + tool_use + tool_result)
  - ✅ SystemReminder extraction and hiding
  - ✅ FunctionDefinitions display

- **code-viewer-highlighting** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-6j7`
  - ✅ Create CodeViewer component
  - ✅ Custom single-pass syntax highlighting (no external deps)
  - ✅ Support JS/TS, Python, Go, Rust, Bash
  - ✅ Line numbers, copy, download, fullscreen

- **copy-to-clipboard** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-nhw`
  - ✅ useCopyToClipboard hook with fallback
  - ✅ CopyButton component with visual feedback
  - ✅ Integrated in Requests.tsx for JSON copy

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
  - ✅ ToolUseContent component (expandable)
  - ✅ Special rendering for bash, read_file, write_file, edit_file
  - ✅ Generic tool input display
  - ✅ Copy tool ID

- **tool-result-display** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-k6y`
  - ✅ ToolResultContent component
  - ✅ Content type detection (text, code, json, blocks)
  - ✅ cat -n format extraction
  - ✅ Success/error styling

- **image-content-display** [COMPLETED]
  - Epic: `brandon-fryslie_claude-code-proxy-asz`
  - ✅ ImageContent component for base64 images
  - ✅ Display inline with content
  - Note: Lightbox/zoom can be added later if needed

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
