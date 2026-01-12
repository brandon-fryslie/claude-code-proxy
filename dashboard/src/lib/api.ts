import { useQuery } from '@tanstack/react-query'
import type {
  RequestSummary,
  RequestLog,
  DashboardStats,
  HourlyStatsResponse,
  ModelStatsResponse,
  ProviderStatsResponse,
  SubagentStatsResponse,
  ToolStatsResponse,
  PerformanceStatsResponse,
  Conversation,
  ConversationDetail,
  Config,
  ProviderConfig,
  SubagentsConfig,
  RoutingConfig,
  ProviderHealth,
  RoutingStatsResponse,
  ConversationMessagesResponse,
} from './types'

// Use V2 API for cleaner responses
const API_BASE = '/api/v2'

// Generic fetch wrapper with error handling
async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(`API error: ${response.status} ${errorText}`)
  }

  return response.json()
}

// Helper to build query string
function buildQueryString(params: Record<string, any>): string {
  const searchParams = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      searchParams.append(key, String(value))
    }
  })
  const qs = searchParams.toString()
  return qs ? `?${qs}` : ''
}

// ============================================================================
// Request Queries
// ============================================================================

interface GetRequestsParams {
  model?: string
  start?: string
  end?: string
  offset?: number
  limit?: number
}

// V2 API returns array directly
export function useRequestsSummary(params?: GetRequestsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['requests', 'summary', params],
    queryFn: () => fetchAPI<RequestSummary[]>(`/requests/summary${queryString}`),
  })
}

// V2 API returns request directly (not wrapped)
export function useRequestDetail(id: string | null) {
  return useQuery({
    queryKey: ['requests', 'detail', id],
    queryFn: () => fetchAPI<RequestLog>(`/requests/${id}`),
    enabled: !!id,
  })
}

export function useLatestRequestDate() {
  return useQuery({
    queryKey: ['requests', 'latest-date'],
    queryFn: async () => {
      // This endpoint doesn't have a v2 version yet, use v1
      const response = await fetch('/api/requests/latest-date', {
        headers: { 'Content-Type': 'application/json' },
      })
      if (!response.ok) throw new Error('Failed to get latest date')
      const data = await response.json()
      return data.date
    },
  })
}

// ============================================================================
// Data Management
// ============================================================================

export async function clearAllRequests(): Promise<void> {
  await fetch('/api/requests', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
  })
}

// ============================================================================
// Stats Queries
// ============================================================================

interface StatsParams {
  start?: string
  end?: string
}

// V2 API ensures arrays are never null
export function useWeeklyStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'weekly', params],
    queryFn: () => fetchAPI<DashboardStats>(`/stats${queryString}`),
  })
}

export function useHourlyStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'hourly', params],
    queryFn: () => fetchAPI<HourlyStatsResponse>(`/stats/hourly${queryString}`),
  })
}

export function useModelStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'models', params],
    queryFn: () => fetchAPI<ModelStatsResponse>(`/stats/models${queryString}`),
  })
}

export function useProviderStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'providers', params],
    queryFn: () => fetchAPI<ProviderStatsResponse>(`/stats/providers${queryString}`),
  })
}

export function useSubagentStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'subagents', params],
    queryFn: () => fetchAPI<SubagentStatsResponse>(`/stats/subagents${queryString}`),
  })
}

export function useToolStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'tools', params],
    queryFn: () => fetchAPI<ToolStatsResponse>(`/stats/tools${queryString}`),
  })
}

export function usePerformanceStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['stats', 'performance', params],
    queryFn: () => fetchAPI<PerformanceStatsResponse>(`/stats/performance${queryString}`),
  })
}

// ============================================================================
// Conversation Queries
// ============================================================================

// V2 API returns array directly
export function useConversations() {
  return useQuery({
    queryKey: ['conversations'],
    queryFn: () => fetchAPI<Conversation[]>('/conversations'),
  })
}

// V2 API returns conversation directly (no project param needed)
export function useConversationDetail(id: string | null) {
  return useQuery({
    queryKey: ['conversations', 'detail', id],
    queryFn: () => fetchAPI<ConversationDetail>(`/conversations/${id}`),
    enabled: !!id,
  })
}

// V2 API returns paginated messages from database
export function useConversationMessages(
  id: string | null,
  options?: { limit?: number; offset?: number; includeSubagents?: boolean }
) {
  const queryString = buildQueryString({
    limit: options?.limit || 100,
    offset: options?.offset || 0,
    include_subagents: options?.includeSubagents ? 'true' : undefined,
  })
  return useQuery({
    queryKey: ['conversations', 'messages', id, options],
    queryFn: () => fetchAPI<ConversationMessagesResponse>(`/conversations/${id}/messages${queryString}`),
    enabled: !!id,
  })
}

// ============================================================================
// Configuration Queries
// ============================================================================

/**
 * Fetches the full configuration (sanitized - API keys redacted)
 */
export function useConfig() {
  return useQuery({
    queryKey: ['config'],
    queryFn: () => fetchAPI<Config>('/config'),
    staleTime: 60000, // Config doesn't change frequently - cache for 60 seconds
  })
}

/**
 * Fetches all provider configurations (sanitized - API keys redacted)
 */
export function useProviders() {
  return useQuery({
    queryKey: ['config', 'providers'],
    queryFn: () => fetchAPI<Record<string, ProviderConfig>>('/config/providers'),
    staleTime: 60000,
  })
}

/**
 * Fetches subagent routing configuration
 */
export function useSubagentConfig() {
  return useQuery({
    queryKey: ['config', 'subagents'],
    queryFn: () => fetchAPI<SubagentsConfig>('/config/subagents'),
    staleTime: 60000,
  })
}

// ============================================================================
// Routing Configuration Queries (Phase 4.1)
// ============================================================================

/**
 * Fetches the full routing configuration including circuit breaker and fallback settings
 * GET /api/v2/routing/config
 */
export function useRoutingConfig() {
  return useQuery({
    queryKey: ['routing', 'config'],
    queryFn: () => fetchAPI<RoutingConfig>('/routing/config'),
    staleTime: 30000, // Routing config changes less often - cache for 30 seconds
  })
}

/**
 * Fetches real-time provider health status including circuit breaker state
 * GET /api/v2/routing/providers
 */
export function useProviderHealth() {
  return useQuery({
    queryKey: ['routing', 'providers'],
    queryFn: () => fetchAPI<ProviderHealth[]>('/routing/providers'),
    staleTime: 5000, // Refresh health status more frequently - cache for 5 seconds
    refetchInterval: 10000, // Auto-refresh every 10 seconds
  })
}

/**
 * Fetches routing statistics including provider usage and circuit breaker trips
 * GET /api/v2/routing/stats
 */
export function useRoutingStats(params?: StatsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['routing', 'stats', params],
    queryFn: () => fetchAPI<RoutingStatsResponse>(`/routing/stats${queryString}`),
    staleTime: 10000, // Cache for 10 seconds
  })
}

// ============================================================================
// Utility Functions
// ============================================================================

export function getTodayDateRange(): { start: string; end: string } {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)

  return {
    start: start.toISOString(),
    end: end.toISOString(),
  }
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

export function formatTokens(tokens: number): string {
  if (tokens < 1000) return tokens.toString()
  if (tokens < 1000000) return `${(tokens / 1000).toFixed(1)}K`
  return `${(tokens / 1000000).toFixed(1)}M`
}
