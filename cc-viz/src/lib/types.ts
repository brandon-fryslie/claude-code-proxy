// TypeScript types for cc-viz (conversation visualization)
// Extracted from proxy/internal/model/models.go

export interface AnthropicContentBlock {
  type: string
  text?: string
  id?: string
  name?: string
  input?: any
  tool_use_id?: string
  content?: string | AnthropicContentBlock[]
  is_error?: boolean
  source?: {
    type: string
    media_type: string
    data: string
  }
  [key: string]: any  // Index signature to allow any additional properties
}

// Conversation types
export interface Conversation {
  id: string
  projectName: string
  startTime: string
  lastActivity: string
  messageCount: number
  rootRequestId?: string
}

// Claude Code log message format
export interface ClaudeCodeMessage {
  type: string  // 'user' | 'assistant' | 'file-history-snapshot' | 'queue-operation' | 'system'
  message?: {
    role?: string
    content?: string | AnthropicContentBlock[]
  } | null
  uuid: string
  timestamp: string
  parentUuid?: string | null
  isSidechain?: boolean
  userType?: string
  cwd?: string
  sessionId?: string
  version?: string
}

// Database message format - includes full message data
export interface DBConversationMessage {
  uuid: string
  conversationId: string
  parentUuid?: string
  type: string
  role?: string
  timestamp: string
  cwd?: string
  gitBranch?: string
  sessionId?: string
  agentId?: string
  isSidechain?: boolean
  requestId?: string
  model?: string
  inputTokens?: number
  outputTokens?: number
  cacheReadTokens?: number
  cacheCreationTokens?: number
  content?: any
}

export interface ConversationMessagesResponse {
  conversationId: string
  messages: DBConversationMessage[] | null
  total: number
  offset: number
  limit: number
}

export interface ConversationDetail {
  sessionId: string
  projectName: string
  projectPath: string
  startTime: string
  endTime: string
  messageCount: number
  messages: ClaudeCodeMessage[]
}
