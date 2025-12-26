import { type FC, useMemo } from 'react'
import { User, Bot, Clock } from 'lucide-react'
import type { ClaudeCodeMessage, AnthropicContentBlock } from '@/lib/types'

interface ConversationThreadProps {
  messages: ClaudeCodeMessage[]
  startTime: string
  endTime: string
}

export const ConversationThread: FC<ConversationThreadProps> = ({
  messages,
  startTime,
  endTime,
}) => {
  // Filter to only user/assistant messages with valid content
  const chatMessages = useMemo(() => {
    return messages.filter(m =>
      (m.type === 'user' || m.type === 'assistant') &&
      m.message?.content
    )
  }, [messages])

  // Count by role
  const stats = useMemo(() => {
    const userCount = chatMessages.filter(m => m.type === 'user').length
    const assistantCount = chatMessages.filter(m => m.type === 'assistant').length
    return { total: chatMessages.length, userCount, assistantCount }
  }, [chatMessages])

  return (
    <div className="flex flex-col h-full">
      {/* Header with stats */}
      <div className="px-4 py-3 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)] flex items-center justify-between">
        <div className="flex items-center gap-4 text-sm text-[var(--color-text-secondary)]">
          <span className="flex items-center gap-1">
            <Clock className="w-4 h-4 text-[var(--color-text-muted)]" />
            {formatTimeRange(startTime, endTime)}
          </span>
          <span>{stats.total} messages</span>
          <span className="text-[var(--color-text-muted)]">
            ({stats.userCount} user, {stats.assistantCount} assistant)
          </span>
        </div>
      </div>

      {/* Message list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {chatMessages.map((msg, idx) => (
          <MessageBubble
            key={msg.uuid || idx}
            message={msg}
          />
        ))}
      </div>
    </div>
  )
}

// Individual message bubble
interface MessageBubbleProps {
  message: ClaudeCodeMessage
}

const MessageBubble: FC<MessageBubbleProps> = ({ message }) => {
  const isUser = message.type === 'user'
  const content = message.message?.content

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[80%] rounded-xl p-4 ${
          isUser
            ? 'bg-blue-50 border border-blue-200 dark:bg-blue-900/20 dark:border-blue-800'
            : 'bg-[var(--color-bg-secondary)] border border-[var(--color-border)]'
        }`}
      >
        <div className="flex items-center gap-2 mb-2">
          <div
            className={`w-7 h-7 rounded-full flex items-center justify-center ${
              isUser
                ? 'bg-blue-100 text-blue-600 dark:bg-blue-800 dark:text-blue-300'
                : 'bg-green-100 text-green-600 dark:bg-green-800 dark:text-green-300'
            }`}
          >
            {isUser ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
          </div>
          <span className="font-medium text-sm text-[var(--color-text-primary)]">
            {isUser ? 'User' : 'Assistant'}
          </span>
          <span className="text-xs text-[var(--color-text-muted)]">
            {formatTimestamp(message.timestamp)}
          </span>
        </div>
        <div className="prose prose-sm max-w-none dark:prose-invert">
          <MessageContent content={content} />
        </div>
      </div>
    </div>
  )
}

const MessageContent: FC<{ content: string | AnthropicContentBlock[] | undefined }> = ({
  content,
}) => {
  if (!content) return <span className="text-[var(--color-text-muted)]">No content</span>

  if (typeof content === 'string') {
    return (
      <div className="text-sm text-[var(--color-text-primary)] whitespace-pre-wrap">
        {content}
      </div>
    )
  }

  if (Array.isArray(content)) {
    return (
      <div className="space-y-2">
        {content.map((block, i) => (
          <div key={i}>
            {block.type === 'text' && block.text && (
              <div className="text-sm text-[var(--color-text-primary)] whitespace-pre-wrap">
                {block.text}
              </div>
            )}
            {block.type === 'tool_use' && (
              <div className="text-sm bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800 rounded p-2">
                <div className="font-medium text-purple-700 dark:text-purple-300 mb-1">
                  Tool: {block.name}
                </div>
                <pre className="text-xs text-[var(--color-text-secondary)] overflow-auto max-h-32">
                  {JSON.stringify(block.input, null, 2)}
                </pre>
              </div>
            )}
            {block.type === 'tool_result' && (
              <div className="text-sm bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded p-2">
                <div className="font-medium text-green-700 dark:text-green-300 mb-1">
                  Tool Result
                </div>
                <pre className="text-xs text-[var(--color-text-secondary)] overflow-auto max-h-32">
                  {typeof block.content === 'string'
                    ? block.content
                    : JSON.stringify(block.content, null, 2)}
                </pre>
              </div>
            )}
          </div>
        ))}
      </div>
    )
  }

  return (
    <pre className="text-xs text-[var(--color-text-muted)]">
      {JSON.stringify(content, null, 2)}
    </pre>
  )
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  })
}

function formatTimeRange(start: string, end: string): string {
  const startDate = new Date(start)
  const endDate = new Date(end)
  const durationMs = endDate.getTime() - startDate.getTime()
  const durationMins = Math.round(durationMs / 60000)

  if (durationMins < 60) {
    return `${durationMins}m`
  }
  return `${Math.round(durationMins / 60)}h ${durationMins % 60}m`
}
