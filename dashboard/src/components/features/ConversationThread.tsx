import { type FC, useMemo } from 'react'
import { User, Bot, Clock } from 'lucide-react'
import type { ConversationMessage, AnthropicContentBlock } from '@/lib/types'

interface ConversationThreadProps {
  messages: ConversationMessage[]
  startTime: string
  endTime: string
}

export const ConversationThread: FC<ConversationThreadProps> = ({
  messages,
  startTime,
  endTime,
}) => {
  // Count by role
  const stats = useMemo(() => {
    const requests = messages.filter(m => m.request).length
    return { total: messages.length, requests }
  }, [messages])

  return (
    <div className="flex flex-col h-full">
      {/* Header with stats */}
      <div className="px-4 py-3 bg-gray-50 border-b flex items-center justify-between">
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <span className="flex items-center gap-1">
            <Clock className="w-4 h-4 text-gray-400" />
            {formatTimeRange(startTime, endTime)}
          </span>
          <span>{stats.total} messages</span>
        </div>
      </div>

      {/* Message list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((msg) => (
          <MessageBubble
            key={msg.requestId}
            message={msg}
          />
        ))}
      </div>
    </div>
  )
}

// Individual message bubble
interface MessageBubbleProps {
  message: ConversationMessage
}

const MessageBubble: FC<MessageBubbleProps> = ({
  message,
}) => {
  const request = message.request

  if (!request) {
    return (
      <div className="p-3 rounded-lg bg-gray-100 border border-gray-200">
        <div className="text-sm text-gray-500">
          Turn {message.turnNumber} - No request data
        </div>
      </div>
    )
  }

  // Determine if this is user or assistant based on the request/response structure
  const hasUserMessage = request.body?.messages?.some(m => m.role === 'user')
  const hasAssistantResponse = request.response?.body?.content

  return (
    <div className="space-y-2">
      {/* User message */}
      {hasUserMessage && request.body?.messages && (
        <div className="flex justify-end">
          <div className="max-w-[80%] bg-blue-50 border border-blue-200 rounded-xl p-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-7 h-7 rounded-full flex items-center justify-center bg-blue-100 text-blue-600">
                <User className="w-4 h-4" />
              </div>
              <span className="font-medium text-sm text-gray-700">User</span>
              <span className="text-xs text-gray-400">
                {formatTimestamp(message.timestamp)}
              </span>
            </div>
            <div className="prose prose-sm max-w-none">
              {request.body.messages
                .filter(m => m.role === 'user')
                .map((m, i) => (
                  <MessageContent key={i} content={m.content} />
                ))}
            </div>
          </div>
        </div>
      )}

      {/* Assistant response */}
      {hasAssistantResponse && (
        <div className="flex justify-start">
          <div className="max-w-[80%] bg-gray-50 border border-gray-200 rounded-xl p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <div className="w-7 h-7 rounded-full flex items-center justify-center bg-green-100 text-green-600">
                  <Bot className="w-4 h-4" />
                </div>
                <span className="font-medium text-sm text-gray-700">Assistant</span>
                <span className="text-xs text-gray-400">
                  {request.model}
                </span>
              </div>
              {request.response && (
                <span className="text-xs text-gray-400">
                  {formatDuration(request.response.responseTime)}
                </span>
              )}
            </div>
            <div className="prose prose-sm max-w-none">
              <MessageContent content={request.response?.body?.content} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const MessageContent: FC<{ content: string | AnthropicContentBlock[] | undefined }> = ({ content }) => {
  if (!content) return null

  if (typeof content === 'string') {
    return <div className="text-sm text-gray-700 whitespace-pre-wrap">{content}</div>
  }

  if (Array.isArray(content)) {
    return (
      <div className="space-y-2">
        {content.map((block, i) => (
          <div key={i}>
            {block.type === 'text' && block.text && (
              <div className="text-sm text-gray-700 whitespace-pre-wrap">{block.text}</div>
            )}
            {block.type === 'tool_use' && (
              <div className="text-sm bg-purple-50 border border-purple-200 rounded p-2">
                <div className="font-medium text-purple-700 mb-1">Tool: {block.name}</div>
                <pre className="text-xs text-gray-600 overflow-auto max-h-32">
                  {JSON.stringify(block.input, null, 2)}
                </pre>
              </div>
            )}
            {block.type === 'tool_result' && (
              <div className="text-sm bg-green-50 border border-green-200 rounded p-2">
                <div className="font-medium text-green-700 mb-1">Tool Result</div>
                <pre className="text-xs text-gray-600 overflow-auto max-h-32">
                  {typeof block.content === 'string' ? block.content : JSON.stringify(block.content, null, 2)}
                </pre>
              </div>
            )}
          </div>
        ))}
      </div>
    )
  }

  return <pre className="text-xs text-gray-500">{JSON.stringify(content, null, 2)}</pre>
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

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}
