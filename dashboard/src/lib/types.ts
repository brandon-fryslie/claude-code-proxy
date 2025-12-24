// TypeScript types matching Go models from proxy/internal/model/models.go

export interface AnthropicUsage {
  input_tokens: number
  output_tokens: number
  cache_creation_input_tokens?: number
  cache_read_input_tokens?: number
  service_tier?: string
}

export interface RequestSummary {
  requestId: string
  timestamp: string
  method: string
  endpoint: string
  model?: string
  originalModel?: string
  routedModel?: string
  provider?: string
  subagentName?: string
  toolsUsed?: string[]
  toolCallCount?: number
  statusCode?: number
  responseTime?: number
  firstByteTime?: number
  usage?: AnthropicUsage
}

export interface AnthropicContentBlock {
  type: string
  text?: string
  id?: string
  name?: string
  input?: any
}

export interface AnthropicMessage {
  role: string
  content: string | AnthropicContentBlock[]
}

export interface AnthropicSystemMessage {
  text: string
  type: string
  cache_control?: {
    type: string
  }
}

export interface Tool {
  name: string
  description: string
  input_schema: {
    type: string
    properties: Record<string, any>
    required?: string[]
  }
}

export interface AnthropicRequest {
  model: string
  messages: AnthropicMessage[]
  max_tokens: number
  temperature?: number
  system?: AnthropicSystemMessage[]
  stream?: boolean
  tools?: Tool[]
  tool_choice?: any
}

export interface ResponseLog {
  statusCode: number
  headers: Record<string, string[]>
  body?: any
  bodyText?: string
  responseTime: number
  firstByteTime?: number
  streamingChunks?: string[]
  isStreaming: boolean
  completedAt: string
  toolCallCount?: number
}

export interface PromptGrade {
  score: number
  maxScore: number
  feedback: string
  improvedPrompt: string
  criteria: Record<string, {
    score: number
    feedback: string
  }>
  gradingTimestamp: string
  isProcessing: boolean
}

export interface RequestLog {
  requestId: string
  timestamp: string
  method: string
  endpoint: string
  headers: Record<string, string[]>
  body?: AnthropicRequest
  model?: string
  originalModel?: string
  routedModel?: string
  provider?: string
  subagentName?: string
  toolsUsed?: string[]
  toolCallCount?: number
  userAgent: string
  contentType: string
  promptGrade?: PromptGrade
  response?: ResponseLog
}

// Dashboard stats structures
export interface DailyTokens {
  date: string
  tokens: number
  requests: number
  models?: Record<string, ModelStats>
}

export interface HourlyTokens {
  hour: number
  tokens: number
  requests: number
  models?: Record<string, ModelStats>
}

export interface ModelStats {
  tokens: number
  requests: number
}

export interface ModelTokens {
  model: string
  tokens: number
  requests: number
}

export interface DashboardStats {
  dailyStats: DailyTokens[]
}

export interface HourlyStatsResponse {
  hourlyStats: HourlyTokens[]
  todayTokens: number
  todayRequests: number
  avgResponseTime: number
}

export interface ModelStatsResponse {
  modelStats: ModelTokens[]
}

// Provider analytics
export interface ProviderStats {
  provider: string
  requests: number
  inputTokens: number
  outputTokens: number
  totalTokens: number
  avgResponseMs: number
  errorCount: number
}

export interface ProviderStatsResponse {
  providers: ProviderStats[]
  startTime: string
  endTime: string
}

// Subagent analytics
export interface SubagentStats {
  subagentName: string
  provider: string
  targetModel: string
  requests: number
  inputTokens: number
  outputTokens: number
  totalTokens: number
  avgResponseMs: number
}

export interface SubagentStatsResponse {
  subagents: SubagentStats[]
  startTime: string
  endTime: string
}

// Tool analytics
export interface ToolStats {
  toolName: string
  usageCount: number
  callCount: number
  avgCallsPerRequest: number
}

export interface ToolStatsResponse {
  tools: ToolStats[]
  startTime: string
  endTime: string
}

// Performance analytics
export interface PerformanceStats {
  provider: string
  model: string
  avgResponseMs: number
  p50ResponseMs: number
  p95ResponseMs: number
  p99ResponseMs: number
  avgFirstByteMs: number
  requestCount: number
}

export interface PerformanceStatsResponse {
  stats: PerformanceStats[]
  startTime: string
  endTime: string
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

export interface ConversationMessage {
  requestId: string
  timestamp: string
  turnNumber: number
  isRoot: boolean
  request?: RequestLog
}

export interface ConversationDetail extends Conversation {
  messages: ConversationMessage[]
}
