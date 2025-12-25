import { useState } from 'react'
import { PageHeader } from '@/components/layout'
import { MessageSquare, Clock, Hash } from 'lucide-react'
import { useConversations, useConversationDetail } from '@/lib/api'
import { cn } from '@/lib/utils'
import { ConversationThread } from '@/components/features/ConversationThread'

function ConversationListItem({
  projectName,
  lastActivity,
  messageCount,
  isSelected,
  onClick,
}: {
  projectName: string
  lastActivity: string
  messageCount: number
  isSelected: boolean
  onClick: () => void
}) {
  const formattedDate = new Date(lastActivity).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full p-3 text-left border-b border-[var(--color-border)]',
        'hover:bg-[var(--color-bg-hover)] transition-colors',
        isSelected && 'bg-[var(--color-bg-active)]'
      )}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {projectName || 'Unknown Project'}
          </h3>
          <div className="flex items-center gap-3 mt-1 text-xs text-[var(--color-text-muted)]">
            <span className="flex items-center gap-1">
              <Clock size={10} />
              {formattedDate}
            </span>
            <span className="flex items-center gap-1">
              <Hash size={10} />
              {messageCount} messages
            </span>
          </div>
        </div>
      </div>
    </button>
  )
}

function ConversationDetail({ conversationId }: { conversationId: string | null }) {
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
      endTime={conversation.lastActivity}
    />
  )
}

export function ConversationsPage() {
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)
  const { data: conversations, isLoading } = useConversations()

  return (
    <>
      <PageHeader
        title="Conversations"
        description="View Claude Code conversation logs"
      />
      <div className="flex-1 flex overflow-hidden">
        {/* Conversation List */}
        <div className="w-80 border-r border-[var(--color-border)] bg-[var(--color-bg-secondary)] overflow-auto">
          {isLoading ? (
            <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
              Loading conversations...
            </div>
          ) : !conversations || conversations.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-[var(--color-text-muted)] p-4">
              <MessageSquare size={48} className="mb-4 opacity-50" />
              <p className="text-center">No conversations found</p>
              <p className="text-sm mt-1 text-center">
                Conversations are parsed from Claude Code logs
              </p>
            </div>
          ) : (
            conversations.map((conv) => (
              <ConversationListItem
                key={conv.id}
                projectName={conv.projectName}
                lastActivity={conv.lastActivity}
                messageCount={conv.messageCount}
                isSelected={selectedConversationId === conv.id}
                onClick={() => setSelectedConversationId(conv.id)}
              />
            ))
          )}
        </div>

        {/* Conversation Detail */}
        <div className="flex-1 bg-[var(--color-bg-primary)]">
          <ConversationDetail conversationId={selectedConversationId} />
        </div>
      </div>
    </>
  )
}
