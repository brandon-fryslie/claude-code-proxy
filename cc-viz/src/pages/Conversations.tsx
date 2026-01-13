import { useState, useMemo, useCallback } from 'react'
import { AppLayout } from '@/components/layout'
import { MessageSquare, ArrowUpDown } from 'lucide-react'
import { useConversations, useConversationDetail, useConversationMessages } from '@/lib/api'
import type { ConversationDetail, DBConversationMessage } from '@/lib/types'
import { ConversationThread } from '@/components/features/ConversationThread'
import { ConversationSearch } from '@/components/features/ConversationSearch'
import { ConversationList } from '@/components/features/ConversationList'
import { filterConversations } from '@/lib/search'

type SortOption = 'recent' | 'project' | 'messages'

function ConversationDetailPane({ conversationId }: { conversationId: string | null }) {
  const { data: conversation, isLoading } = useConversationDetail(conversationId)
  const [useDBMessages, setUseDBMessages] = useState(false)
  const { data: dbMessagesResponse, isLoading: isLoadingDBMessages } = useConversationMessages(
    useDBMessages ? conversationId : null,
    { includeSubagents: true }
  )

  if (!conversationId) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-[var(--color-text-muted)]">
        <MessageSquare size={48} className="mb-4 opacity-50" />
        <p>Select a conversation to view details</p>
      </div>
    )
  }

  if (isLoading || (useDBMessages && isLoadingDBMessages)) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        {useDBMessages ? 'Loading messages from database...' : 'Loading conversation...'}
      </div>
    )
  }

  if (!conversation && !useDBMessages) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Conversation not found
      </div>
    )
  }

  // Toggle to show database messages if available
  const showDBToggle = dbMessagesResponse && dbMessagesResponse.messages && dbMessagesResponse.messages.length > 0

  return (
    <div className="flex flex-col h-full">
      {showDBToggle && (
        <div className="px-4 py-2 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)] flex items-center justify-between">
          <span className="text-xs text-[var(--color-text-secondary)]">
            {useDBMessages ? `${dbMessagesResponse.total} messages from database` : 'Conversation file format'}
          </span>
          <button
            onClick={() => setUseDBMessages(!useDBMessages)}
            className="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded hover:bg-blue-200 dark:hover:bg-blue-900/50 transition-colors"
          >
            {useDBMessages ? 'Show File Format' : 'Show Database Format'}
          </button>
        </div>
      )}
      <div className="flex-1 overflow-hidden">
        {useDBMessages && dbMessagesResponse?.messages ? (
          <DBMessagesView messages={dbMessagesResponse.messages} />
        ) : conversation ? (
          <ConversationThread
            messages={conversation.messages}
            startTime={conversation.startTime}
            endTime={conversation.endTime}
          />
        ) : (
          <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
            No conversation data found
          </div>
        )}
      </div>
    </div>
  )
}

function DBMessagesView({ messages }: { messages: DBConversationMessage[] }) {
  const [searchQuery, setSearchQuery] = useState('')

  // Filter messages based on search
  const filteredMessages = useMemo(() => {
    if (!searchQuery) return messages
    const query = searchQuery.toLowerCase()
    return messages.filter(msg =>
      msg.content?.toString().toLowerCase().includes(query) ||
      msg.model?.toLowerCase().includes(query) ||
      msg.agentId?.toLowerCase().includes(query)
    )
  }, [messages, searchQuery])

  return (
    <div className="flex flex-col h-full">
      <div className="sticky top-0 z-10 px-4 py-2 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)]">
        <div className="flex items-center justify-between">
          <span className="text-sm text-[var(--color-text-secondary)]">
            {filteredMessages.length} / {messages.length} messages
          </span>
        </div>
        <div className="mt-2">
          <input
            type="text"
            placeholder="Search messages..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-3 py-1.5 text-sm bg-[var(--color-bg-primary)] border border-[var(--color-border)] rounded focus:outline-none focus:border-blue-500"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {filteredMessages.map((msg) => (
          <div
            key={msg.uuid}
            className="p-3 bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-lg"
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2 text-xs">
                <span className="font-medium text-[var(--color-text-primary)]">{msg.role || msg.type}</span>
                {msg.model && <span className="text-[var(--color-text-muted)]">{msg.model}</span>}
                {msg.agentId && (
                  <span className="px-1.5 py-0.5 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded text-xs">
                    Agent: {msg.agentId}
                  </span>
                )}
              </div>
              <span className="text-xs text-[var(--color-text-muted)]">
                {new Date(msg.timestamp).toLocaleTimeString()}
              </span>
            </div>

            {/* Token usage info */}
            {(msg.inputTokens || msg.outputTokens) && (
              <div className="mb-2 text-xs text-[var(--color-text-muted)]">
                Tokens: {msg.inputTokens || 0} in, {msg.outputTokens || 0} out
              </div>
            )}

            {/* Message content */}
            <div className="text-sm text-[var(--color-text-primary)] max-h-24 overflow-y-auto p-2 bg-[var(--color-bg-primary)] rounded">
              {typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content, null, 2)}
            </div>
          </div>
        ))}
        {filteredMessages.length === 0 && (
          <div className="text-center text-[var(--color-text-muted)] py-8">
            No messages match your search
          </div>
        )}
      </div>
    </div>
  )
}

export function ConversationsPage() {
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<SortOption>('recent')
  const [conversationDetailsCache] = useState<Map<string, ConversationDetail>>(
    new Map()
  )

  const { data: conversations, isLoading } = useConversations()

  // Update cache when a conversation detail is loaded
  const handleConversationSelect = useCallback((id: string) => {
    setSelectedConversationId(id)
  }, [])

  // Sort conversations
  const sortedConversations = useMemo(() => {
    if (!conversations) return []

    const sorted = [...conversations]
    switch (sortBy) {
      case 'recent':
        return sorted.sort((a, b) =>
          new Date(b.lastActivity).getTime() - new Date(a.lastActivity).getTime()
        )
      case 'project':
        return sorted.sort((a, b) =>
          (a.projectName || '').localeCompare(b.projectName || '')
        )
      case 'messages':
        return sorted.sort((a, b) => b.messageCount - a.messageCount)
      default:
        return sorted
    }
  }, [conversations, sortBy])

  // Filter by search query
  const filteredConversations = useMemo(() => {
    return filterConversations(sortedConversations, conversationDetailsCache, searchQuery)
  }, [sortedConversations, conversationDetailsCache, searchQuery])

  // Cycle through sort options
  const cycleSortOption = () => {
    setSortBy((current) => {
      if (current === 'recent') return 'project'
      if (current === 'project') return 'messages'
      return 'recent'
    })
  }

  const getSortLabel = () => {
    switch (sortBy) {
      case 'recent': return 'Recent'
      case 'project': return 'Project'
      case 'messages': return 'Messages'
    }
  }

  return (
    <AppLayout
      title="Conversations"
      description="Browse Claude Code conversation logs"
      activeItem="conversations"
    >
      <div className="flex-1 flex overflow-hidden h-full">
        {/* Left Sidebar - Conversation List */}
        <div className="w-80 border-r border-[var(--color-border)] bg-[var(--color-bg-secondary)] flex flex-col">
          {/* Search and Sort */}
          <div className="p-3 border-b border-[var(--color-border)] space-y-2">
            <ConversationSearch
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="Search conversations..."
            />
            <div className="flex items-center justify-between">
              <span className="text-xs text-[var(--color-text-muted)]">
                {filteredConversations.length} conversation{filteredConversations.length !== 1 ? 's' : ''}
              </span>
              <button
                onClick={cycleSortOption}
                className="flex items-center gap-1 px-2 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)] rounded transition-colors"
                title="Change sort order"
              >
                <ArrowUpDown size={12} />
                <span>Sort: {getSortLabel()}</span>
              </button>
            </div>
          </div>

          {/* Conversation List */}
          <div className="flex-1 overflow-hidden">
            <ConversationList
              conversations={filteredConversations}
              conversationDetails={conversationDetailsCache}
              selectedId={selectedConversationId}
              onSelect={handleConversationSelect}
              searchQuery={searchQuery}
              isLoading={isLoading}
            />
          </div>
        </div>

        {/* Right Pane - Conversation Detail */}
        <div className="flex-1 bg-[var(--color-bg-primary)] relative">
          <ConversationDetailPane conversationId={selectedConversationId} />
        </div>
      </div>
    </AppLayout>
  )
}
