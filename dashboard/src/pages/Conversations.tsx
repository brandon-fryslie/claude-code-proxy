import { useState, useMemo, useCallback } from 'react'
import { PageHeader } from '@/components/layout'
import { MessageSquare, ArrowUpDown } from 'lucide-react'
import { useConversations, useConversationDetail } from '@/lib/api'
import type { ConversationDetail } from '@/lib/types'
import { ConversationThread } from '@/components/features/ConversationThread'
import { ConversationSearch } from '@/components/features/ConversationSearch'
import { ConversationList } from '@/components/features/ConversationList'
import { filterConversations } from '@/lib/search'

type SortOption = 'recent' | 'project' | 'messages'

function ConversationDetailPane({ conversationId }: { conversationId: string | null }) {
  const { data: conversation, isLoading } = useConversationDetail(conversationId)

  if (!conversationId) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-[var(--color-text-muted)]">
        <MessageSquare size={48} className="mb-4 opacity-50" />
        <p>Select a conversation to view details</p>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Loading conversation...
      </div>
    )
  }

  if (!conversation) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Conversation not found
      </div>
    )
  }

  return (
    <ConversationThread
      messages={conversation.messages}
      startTime={conversation.startTime}
      endTime={conversation.endTime}
    />
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
    <>
      <PageHeader
        title="Conversations"
        description="Browse Claude Code conversation logs"
      />
      <div className="flex-1 flex overflow-hidden">
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
    </>
  )
}
