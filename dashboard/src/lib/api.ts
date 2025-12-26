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
} from './types'

// Base API URL - proxy is configured in vite.config.ts
const API_BASE = '/api'

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

export function useRequestsSummary(params?: GetRequestsParams) {
  const queryString = buildQueryString(params || {})
  return useQuery({
    queryKey: ['requests', 'summary', params],
    queryFn: async () => {
      const response = await fetchAPI<{ requests: RequestSummary[] }>(`/requests/summary${queryString}`)
      return response.requests || []
    },
  })
}

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
      const response = await fetchAPI<{ date: string }>('/requests/latest-date')
      return response.date
    },
  })
}

// ============================================================================
// Data Management
// ============================================================================

export async function clearAllRequests(): Promise<void> {
  await fetch(`${API_BASE}/requests`, {
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

export function useConversations() {
  return useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const response = await fetchAPI<{ conversations: Conversation[] }>('/conversations')
      return response.conversations || []
    },
  })
}

export function useConversationDetail(id: string | null) {
  return useQuery({
    queryKey: ['conversations', 'detail', id],
    queryFn: () => fetchAPI<ConversationDetail>(`/conversations/${id}`),
    enabled: !!id,
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
