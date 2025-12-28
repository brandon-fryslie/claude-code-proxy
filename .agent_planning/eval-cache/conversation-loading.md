# Conversation Loading Infrastructure

**Status:** COMPLETE
**Confidence:** FRESH (verified 2025-12-27 23:30)
**Source:** proxy/internal/service/conversation.go

## Overview
The proxy has a complete, working system for loading Claude Code conversation logs from `~/.claude/projects/` directory.

## Implementation Details

### Service Interface
```go
type ConversationService interface {
    GetConversations() (map[string][]*Conversation, error)
    GetConversation(projectPath, sessionID string) (*Conversation, error)
    GetConversationsByProject(projectPath string) ([]*Conversation, error)
}
```

**Location:** `proxy/internal/service/conversation.go:14-18`

### Data Structures

#### Conversation
```go
type Conversation struct {
    SessionID    string                 // Filename without .jsonl
    ProjectPath  string                 // Relative path from .claude/projects/
    ProjectName  string                 // Same as ProjectPath
    Messages     []*ConversationMessage
    StartTime    time.Time              // First message timestamp
    EndTime      time.Time              // Last message timestamp
    MessageCount int
    FileModTime  time.Time              // For sorting
}
```

**Location:** `proxy/internal/service/conversation.go:46-56`

#### ConversationMessage
```go
type ConversationMessage struct {
    ParentUUID  *string         // Parent message UUID (threading)
    IsSidechain bool            // Is this a sidechain message
    UserType    string          // "user" or "assistant"
    CWD         string          // Working directory
    SessionID   string          // Session identifier
    Version     string          // Claude Code version
    Type        string          // Message type (see below)
    Message     json.RawMessage // ⚠️ Raw JSON - format varies by Type
    UUID        string          // Unique message ID
    Timestamp   string          // RFC3339 format
    ParsedTime  time.Time       // Parsed timestamp
}
```

**Location:** `proxy/internal/service/conversation.go:32-44`

## Key Features

### JSONL Parsing
- ✅ Handles large messages (10MB buffer)
- ✅ Skips malformed lines gracefully
- ✅ Multiple timestamp format support (RFC3339, RFC3339Nano)
- ✅ Sorts messages by timestamp
- ✅ Extracts session metadata (start/end time, count)

**Location:** `proxy/internal/service/conversation.go:160-295`

### File Discovery
- Walks `~/.claude/projects/` recursively
- Filters for `.jsonl` files only
- Groups by project path
- Sorts by file modification time (newest first)

**Location:** `proxy/internal/service/conversation.go:59-112`

### Error Handling
- Gracefully handles missing files
- Logs parsing errors but continues processing
- Returns empty conversation on complete parse failure
- Uses file modification time as fallback for timestamps

## Message Types (From Code)

The `Type` field indicates message type. Common values:
- `"userMessage"` - User input
- `"assistantMessage"` - Assistant response
- `"toolResult"` - Tool execution result
- Others - undocumented, needs investigation of real files

**⚠️ CRITICAL:** The `Message` field structure varies by `Type`. This is a `json.RawMessage` that needs type-specific parsing.

## Usage Example

```go
cs := service.NewConversationService()

// Get all conversations, grouped by project
conversations, err := cs.GetConversations()
// Returns: map[projectPath][]*Conversation

// Get specific conversation
conv, err := cs.GetConversation("project-name", "session-abc123")

// Get all conversations for a project
convs, err := cs.GetConversationsByProject("project-name")
```

## API Integration

**Handlers exist in:** `proxy/internal/handler/handlers.go`
- `GetConversations(w, r)` - line 871
- `GetConversation(projectPath, sessionID)` - line 961
- `GetConversationsByProject(projectPath)` - line 971

## Limitations

1. **No indexing** - files parsed on every request
2. **No caching** - repeated calls re-read files
3. **Message parsing incomplete** - `Message` field is raw JSON
4. **No search** - client must filter in memory

## For Search Implementation

To add full-text search, you need:
1. Parse `Message json.RawMessage` based on `Type`
2. Extract text content from parsed messages
3. Index content in SQLite FTS5 table
4. Detect file changes for incremental indexing

**Next step:** Research JSONL message format to understand `Message` field structure.

## Reuse For
- Implementer working on conversation search indexing
- Anyone needing to understand how conversations are loaded
- Future conversation-related features

## Performance Notes
- **Current implementation:** File I/O on every request
- **Scalability:** Works for <100 projects, slow for 1000+
- **Improvement needed:** Add caching or indexing for production use

