/**
 * Search utilities for conversation filtering
 */

import type { Conversation, ConversationDetail, ClaudeCodeMessage, AnthropicContentBlock } from './types'

/**
 * Extract searchable text from message content
 */
export function extractTextFromContent(content: string | AnthropicContentBlock[] | undefined): string {
  if (!content) return ''

  if (typeof content === 'string') {
    return content
  }

  if (Array.isArray(content)) {
    return content
      .map(block => {
        if (block.type === 'text' && block.text) {
          return block.text
        }
        if (block.type === 'tool_use' && block.name) {
          return `[tool:${block.name}]`
        }
        return ''
      })
      .filter(Boolean)
      .join(' ')
  }

  return ''
}

/**
 * Extract all tool names used in a message
 */
export function extractToolNames(content: string | AnthropicContentBlock[] | undefined): string[] {
  if (!content || typeof content === 'string') return []

  if (Array.isArray(content)) {
    return content
      .filter(block => block.type === 'tool_use' && block.name)
      .map(block => block.name!)
  }

  return []
}

/**
 * Get preview text from the last message in a conversation
 */
export function getLastMessagePreview(messages: ClaudeCodeMessage[], maxLength = 80): string {
  // Find the last user or assistant message
  const lastMessage = [...messages]
    .reverse()
    .find(m => (m.type === 'user' || m.type === 'assistant') && m.message?.content)

  if (!lastMessage?.message?.content) return ''

  const text = extractTextFromContent(lastMessage.message.content)
  if (text.length <= maxLength) return text

  return text.substring(0, maxLength) + '...'
}

/**
 * Extract all tool names used in a conversation
 */
export function getToolsUsedInConversation(messages: ClaudeCodeMessage[]): string[] {
  const toolSet = new Set<string>()

  messages.forEach(msg => {
    if (msg.message?.content) {
      const tools = extractToolNames(msg.message.content)
      tools.forEach(tool => toolSet.add(tool))
    }
  })

  return Array.from(toolSet)
}

/**
 * Search query matching - checks if text contains all search terms (case-insensitive)
 */
export function matchesSearchQuery(text: string, query: string): boolean {
  if (!query.trim()) return true

  const normalizedText = text.toLowerCase()
  const searchTerms = query.toLowerCase().trim().split(/\s+/)

  return searchTerms.every(term => normalizedText.includes(term))
}

/**
 * Filter conversations by search query
 * Searches: project name, message content, tool names
 */
export function filterConversations(
  conversations: Conversation[],
  conversationDetails: Map<string, ConversationDetail>,
  query: string
): Conversation[] {
  if (!query.trim()) return conversations

  return conversations.filter(conv => {
    // Search project name
    if (matchesSearchQuery(conv.projectName, query)) return true

    // Search message content and tools (requires detail to be loaded)
    const detail = conversationDetails.get(conv.id)
    if (detail) {
      // Search message content
      const hasMatchingMessage = detail.messages.some(msg => {
        if (!msg.message?.content) return false
        const text = extractTextFromContent(msg.message.content)
        return matchesSearchQuery(text, query)
      })

      if (hasMatchingMessage) return true

      // Search tool names
      const tools = getToolsUsedInConversation(detail.messages)
      if (tools.some(tool => matchesSearchQuery(tool, query))) return true
    }

    return false
  })
}

/**
 * Filter messages within a conversation by search query
 */
export function filterMessages(
  messages: ClaudeCodeMessage[],
  query: string
): ClaudeCodeMessage[] {
  if (!query.trim()) return messages

  return messages.filter(msg => {
    if (!msg.message?.content) return false

    const text = extractTextFromContent(msg.message.content)
    const tools = extractToolNames(msg.message.content)

    return matchesSearchQuery(text, query) || tools.some(tool => matchesSearchQuery(tool, query))
  })
}

/**
 * Format time ago (e.g., "2h ago", "yesterday")
 */
export function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays === 1) return 'yesterday'
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
