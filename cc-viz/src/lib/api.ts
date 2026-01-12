import { useQuery } from '@tanstack/react-query'
import type {
  Conversation,
  ConversationDetail,
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
