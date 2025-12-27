import { type FC, useEffect, useRef, useMemo } from 'react'
import { MessageSquare, Clock, Hash, Wrench } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Conversation, ConversationDetail } from '@/lib/types'
import { formatTimeAgo, getLastMessagePreview, getToolsUsedInConversation } from '@/lib/search'
import { highlightMatches } from '@/lib/searchHighlight'

interface ConversationListItemProps {
  conversation: Conversation
  detail?: ConversationDetail
  isSelected: boolean
  onClick: () => void
  searchQuery: string
}

function ConversationListItem({
  conversation,
  detail,
  isSelected,
  onClick,
  searchQuery,
}: ConversationListItemProps) {
  const itemRef = useRef<HTMLButtonElement>(null)

  // Scroll into view when selected
  useEffect(() => {
    if (isSelected && itemRef.current) {
      itemRef.current.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }
  }, [isSelected])

  const preview = useMemo(() => {
    if (!detail) return ''
    return getLastMessagePreview(detail.messages)
  }, [detail])

  const hasTools = useMemo(() => {
    if (!detail) return false
    return getToolsUsedInConversation(detail.messages).length > 0
  }, [detail])

  return (
    <button
      ref={itemRef}
      onClick={onClick}
      className={cn(
        'w-full px-3 py-2 text-left border-b border-[var(--color-border)]',
        'hover:bg-[var(--color-bg-hover)] transition-colors',
        'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500',
        isSelected && 'bg-[var(--color-bg-active)]'
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          {/* Project name */}
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {highlightMatches(conversation.projectName || 'Unknown Project', searchQuery)}
          </h3>

          {/* Preview text */}
          {preview && (
            <p className="text-xs text-[var(--color-text-muted)] truncate mt-0.5">
              {preview}
            </p>
          )}

          {/* Metadata row */}
          <div className="flex items-center gap-2 mt-1 text-[10px] text-[var(--color-text-muted)]">
            <span className="flex items-center gap-1">
              <Clock size={10} />
              {formatTimeAgo(conversation.lastActivity)}
            </span>
            <span className="flex items-center gap-1">
              <Hash size={10} />
              {conversation.messageCount}
            </span>
            {hasTools && (
              <span className="flex items-center gap-1 text-purple-600 dark:text-purple-400">
                <Wrench size={10} />
              </span>
            )}
          </div>
        </div>
      </div>
    </button>
  )
}

interface ConversationListProps {
  conversations: Conversation[]
  conversationDetails: Map<string, ConversationDetail>
  selectedId: string | null
  onSelect: (id: string) => void
  searchQuery: string
  isLoading?: boolean
}

export const ConversationList: FC<ConversationListProps> = ({
  conversations,
  conversationDetails,
  selectedId,
  onSelect,
  searchQuery,
  isLoading,
}) => {
  const listRef = useRef<HTMLDivElement>(null)

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!conversations.length) return

      const currentIndex = selectedId
        ? conversations.findIndex(c => c.id === selectedId)
        : -1

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        const nextIndex = Math.min(currentIndex + 1, conversations.length - 1)
        onSelect(conversations[nextIndex].id)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        const prevIndex = Math.max(currentIndex - 1, 0)
        onSelect(conversations[prevIndex].id)
      } else if (e.key === 'Enter' && selectedId) {
        e.preventDefault()
        // Already selected, could trigger detail load or other action
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [conversations, selectedId, onSelect])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
        Loading conversations...
      </div>
    )
  }

  if (conversations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-[var(--color-text-muted)] p-4">
        <MessageSquare size={48} className="mb-4 opacity-50" />
        <p className="text-center">
          {searchQuery ? 'No conversations match your search' : 'No conversations found'}
        </p>
        {!searchQuery && (
          <p className="text-sm mt-1 text-center">
            Conversations are parsed from Claude Code logs
          </p>
        )}
      </div>
    )
  }

  return (
    <div ref={listRef} className="overflow-y-auto">
      {conversations.map((conv) => (
        <ConversationListItem
          key={conv.id}
          conversation={conv}
          detail={conversationDetails.get(conv.id)}
          isSelected={selectedId === conv.id}
          onClick={() => onSelect(conv.id)}
          searchQuery={searchQuery}
        />
      ))}
    </div>
  )
}
