import { type FC, useMemo } from 'react'
import { X, Clock, Hash, Zap } from 'lucide-react'
import { formatDuration, formatTokens } from '@/lib/api'
import { cn } from '@/lib/utils'
import type { RequestLog, AnthropicContentBlock } from '@/lib/types'

interface RequestCompareModalProps {
  request1: RequestLog
  request2: RequestLog
  onClose: () => void
}

export const RequestCompareModal: FC<RequestCompareModalProps> = ({
  request1,
  request2,
  onClose,
}) => {
  // Handle escape key
  useMemo(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-2xl shadow-2xl max-w-[95vw] max-h-[90vh] w-full mx-4 overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b bg-gray-50">
          <h2 className="text-lg font-semibold text-gray-900">Compare Requests</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-200 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto">
          {/* Stats Comparison */}
          <div className="grid grid-cols-2 gap-4 p-4 bg-gray-50 border-b">
            <StatsCard request={request1} label="Request 1" />
            <StatsCard request={request2} label="Request 2" />
          </div>

          {/* Side-by-side Content */}
          <div className="grid grid-cols-2 divide-x">
            {/* Request 1 */}
            <div className="p-4 space-y-4">
              <SectionHeader title="Messages" requestId={request1.requestId} />
              <RequestContent request={request1} />
            </div>

            {/* Request 2 */}
            <div className="p-4 space-y-4">
              <SectionHeader title="Messages" requestId={request2.requestId} />
              <RequestContent request={request2} />
            </div>
          </div>

          {/* Response Comparison */}
          <div className="grid grid-cols-2 divide-x border-t">
            <div className="p-4 space-y-4">
              <SectionHeader title="Response" />
              <ResponseContent response={request1.response} />
            </div>
            <div className="p-4 space-y-4">
              <SectionHeader title="Response" />
              <ResponseContent response={request2.response} />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Stats card for quick comparison
const StatsCard: FC<{ request: RequestLog; label: string }> = ({ request, label }) => {
  const usage = request.response?.body?.usage

  return (
    <div className="bg-white rounded-lg p-4 border">
      <div className="text-xs text-gray-500 mb-2">{label}</div>
      <div className="grid grid-cols-4 gap-4 text-sm">
        <div>
          <div className="text-gray-500 text-xs">Model</div>
          <div className="font-medium truncate" title={request.model}>
            {getModelShortName(request.model || 'unknown')}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Response Time</div>
          <div className="font-medium flex items-center gap-1">
            <Clock className="w-3 h-3 text-gray-400" />
            {formatDuration(request.response?.responseTime || 0)}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Input Tokens</div>
          <div className="font-medium flex items-center gap-1">
            <Hash className="w-3 h-3 text-gray-400" />
            {formatTokens(usage?.input_tokens || 0)}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Output Tokens</div>
          <div className="font-medium flex items-center gap-1">
            <Zap className="w-3 h-3 text-gray-400" />
            {formatTokens(usage?.output_tokens || 0)}
          </div>
        </div>
      </div>
    </div>
  )
}

const SectionHeader: FC<{ title: string; requestId?: string }> = ({ title, requestId }) => (
  <div className="flex items-center justify-between">
    <h3 className="font-medium text-gray-900">{title}</h3>
    {requestId && (
      <span className="text-xs font-mono text-gray-400">{requestId.slice(-8)}</span>
    )}
  </div>
)

const RequestContent: FC<{ request: RequestLog }> = ({ request }) => {
  const messages = request.body?.messages || []

  return (
    <div className="space-y-3 max-h-64 overflow-y-auto">
      {messages.map((msg, i) => (
        <div
          key={i}
          className={cn(
            'p-3 rounded-lg border',
            msg.role === 'user' ? 'bg-blue-50 border-blue-200' : 'bg-gray-50 border-gray-200'
          )}
        >
          <div className="text-xs font-medium text-gray-500 mb-1">
            {msg.role}
          </div>
          <MessageContent content={msg.content} />
        </div>
      ))}
    </div>
  )
}

const ResponseContent: FC<{ response?: RequestLog['response'] }> = ({ response }) => {
  if (!response?.body?.content) {
    return <div className="text-gray-400 italic">No response</div>
  }

  return (
    <div className="max-h-64 overflow-y-auto">
      <MessageContent content={response.body.content} />
    </div>
  )
}

const MessageContent: FC<{ content: string | AnthropicContentBlock[] }> = ({ content }) => {
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
                <pre className="text-xs text-gray-600 overflow-auto">
                  {JSON.stringify(block.input, null, 2)}
                </pre>
              </div>
            )}
          </div>
        ))}
      </div>
    )
  }

  return <div className="text-gray-400 italic">Unknown content format</div>
}

function getModelShortName(model: string): string {
  if (model.includes('opus')) return 'Opus'
  if (model.includes('sonnet')) return 'Sonnet'
  if (model.includes('haiku')) return 'Haiku'
  if (model.includes('gpt-4o')) return 'GPT-4o'
  if (model.includes('gpt-4')) return 'GPT-4'
  return model.split('-').slice(0, 2).join('-')
}
