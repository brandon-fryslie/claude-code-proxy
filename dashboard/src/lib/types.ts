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

// ============================================================================
// Configuration Types
// ============================================================================

/**
 * Provider configuration - matches proxy/internal/config/config.go ProviderConfig
 */
export interface ProviderConfig {
  /** Provider format: "anthropic" or "openai" */
  format: 'anthropic' | 'openai'
  /** API base URL */
  base_url: string
  /** API key (will be "***REDACTED***" in responses from server) */
  api_key?: string
  /** API version (for Anthropic-format providers) */
  version?: string
  /** Max retry attempts */
  max_retries?: number
}

/**
 * Subagent routing configuration - matches proxy/internal/config/config.go SubagentsConfig
 */
export interface SubagentsConfig {
  /** Whether subagent routing is enabled */
  enable: boolean
  /** Mapping of agent name to "provider:model" string */
  mappings: Record<string, string>
}

/**
 * Server configuration - matches proxy/internal/config/config.go ServerConfig
 */
export interface ServerConfig {
  port: string
  timeouts: {
    read: string
    write: string
    idle: string
  }
}

/**
 * Storage configuration - matches proxy/internal/config/config.go StorageConfig
 */
export interface StorageConfig {
  requests_dir: string
  db_path: string
}

/**
 * Full configuration - matches proxy/internal/config/config.go Config
 */
export interface Config {
  server: ServerConfig
  providers: Record<string, ProviderConfig>
  storage: StorageConfig
  subagents: SubagentsConfig
}

// ============================================================================
// Routing Configuration Types (Phase 4.1)
// ============================================================================

/**
 * Circuit breaker configuration - matches proxy/internal/config/config.go CircuitBreakerConfig
 */
export interface CircuitBreakerConfig {
  enabled: boolean
  max_failures: number
  timeout: string
}

/**
 * Extended provider configuration including circuit breaker and fallback settings
 * Matches the response from GET /api/v2/routing/config
 */
export interface RoutingProviderConfig {
  format: 'anthropic' | 'openai'
  base_url: string
  max_retries: number
  fallback_provider?: string
  circuit_breaker: CircuitBreakerConfig
}

/**
 * Full routing configuration response
 * Matches the response from GET /api/v2/routing/config
 */
export interface RoutingConfig {
  providers: Record<string, RoutingProviderConfig>
  subagents: {
    enable: boolean
    mappings: Record<string, string>
  }
}

/**
 * Provider health status - matches proxy/internal/service/model_router.go ProviderHealth
 */
export interface ProviderHealth {
  name: string
  healthy: boolean
  circuit_breaker_state?: 'closed' | 'open' | 'half-open'
  fallback_provider?: string
}

/**
 * Routing statistics response
 * Matches the response from GET /api/v2/routing/stats
 */
export interface RoutingStatsResponse {
  providers: ProviderStatsResponse
  subagents: SubagentStatsResponse
  timeRange: {
    start: string
    end: string
  }
}
