import { useState } from 'react'
import { PageHeader } from '@/components/layout'
import { ResizablePanel, PanelGroup, Panel } from '@/components/layout'
import { cn } from '@/lib/utils'
import { ChevronRight, Clock, ArrowRight, Filter, Search, User, Bot, ChevronDown } from 'lucide-react'
import { useRequestsSummary, useRequestDetail, formatDuration } from '@/lib/api'
import { MessageContent } from '@/components/ui'
import type { RequestSummary as RequestSummaryType, AnthropicMessage } from '@/lib/types'

function RequestListItem({
  request,
  isSelected,
  onClick,
}: {
  request: RequestSummaryType
  isSelected: boolean
  onClick: () => void
}) {
  const status = request.statusCode && request.statusCode >= 200 && request.statusCode < 300 ? 'success' :
                request.statusCode && request.statusCode >= 400 ? 'error' : 'pending'


  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full flex items-center gap-3 px-3 py-2 text-left border-b border-[var(--color-border)]',
        'hover:bg-[var(--color-bg-hover)] transition-colors',
        isSelected && 'bg-[var(--color-bg-active)]'
      )}
    >
      <div
        className={cn(
          'w-2 h-2 rounded-full flex-shrink-0',
          status === 'success' && 'bg-[var(--color-success)]',
          status === 'error' && 'bg-[var(--color-error)]',
          status === 'pending' && 'bg-[var(--color-warning)]'
        )}
      />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {request.model || request.originalModel || 'unknown'}
          </span>
          {request.provider && (
            <>
              <ArrowRight size={12} className="text-[var(--color-text-muted)] flex-shrink-0" />
              <span className="text-xs text-[var(--color-text-secondary)]">
                {request.provider}
              </span>
            </>
          )}
        </div>
        <div className="flex items-center gap-3 mt-1 text-xs text-[var(--color-text-muted)]">
          <span>
            {request.usage?.input_tokens?.toLocaleString() || 0} → {request.usage?.output_tokens?.toLocaleString() || 0}
          </span>
          {request.responseTime && (
            <span className="flex items-center gap-1">
              <Clock size={10} />
              {formatDuration(request.responseTime)}
            </span>
          )}
        </div>
      </div>
      <ChevronRight size={14} className="text-[var(--color-text-muted)] flex-shrink-0" />
    </button>
  )
}

// Message bubble component for chat-style display
function MessageBubble({ message }: { message: AnthropicMessage }) {
  const [expanded, setExpanded] = useState(false)
  const isUser = message.role === 'user'
  const isAssistant = message.role === 'assistant'

  // Get content preview
  const content = message.content
  const hasMultipleBlocks = Array.isArray(content) && content.length > 1
  const firstBlock = Array.isArray(content) ? content[0] : content

  // Check if this is a tool result message
  const isToolResult = Array.isArray(content) && content.some(
    (block): block is { type: string } =>
      block && typeof block === 'object' && 'type' in block && block.type === 'tool_result'
  )

  return (
    <div className={cn(
      'rounded-lg border overflow-hidden',
      isUser && 'bg-blue-50/50 border-blue-200',
      isAssistant && 'bg-gray-50 border-gray-200',
      !isUser && !isAssistant && 'bg-amber-50/50 border-amber-200'
    )}>
      {/* Header */}
      <div
        className={cn(
          'flex items-center gap-2 px-3 py-2 cursor-pointer',
          isUser && 'bg-blue-100/50',
          isAssistant && 'bg-gray-100',
          !isUser && !isAssistant && 'bg-amber-100/50'
        )}
        onClick={() => setExpanded(!expanded)}
      >
        <div className={cn(
          'w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0',
          isUser && 'bg-blue-500 text-white',
          isAssistant && 'bg-gray-700 text-white',
          !isUser && !isAssistant && 'bg-amber-500 text-white'
        )}>
          {isUser ? <User size={14} /> : <Bot size={14} />}
        </div>
        <div className="flex-1 min-w-0">
          <span className={cn(
            'text-sm font-medium capitalize',
            isUser && 'text-blue-800',
            isAssistant && 'text-gray-800',
            !isUser && !isAssistant && 'text-amber-800'
          )}>
            {message.role}
          </span>
          {hasMultipleBlocks && (
            <span className="ml-2 text-xs text-gray-500">
              ({(content as unknown[]).length} blocks)
            </span>
          )}
          {isToolResult && (
            <span className="ml-2 text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded">
              tool result
            </span>
          )}
        </div>
        <ChevronDown
          size={16}
          className={cn(
            'text-gray-400 transition-transform',
            expanded && 'rotate-180'
          )}
        />
      </div>

      {/* Content */}
      {expanded && (
        <div className="p-3">
          <MessageContent content={content} />
        </div>
      )}

      {/* Preview when collapsed */}
      {!expanded && (
        <div className="px-3 py-2 text-sm text-gray-600 truncate">
          {typeof firstBlock === 'string'
            ? firstBlock.substring(0, 100) + (firstBlock.length > 100 ? '...' : '')
            : firstBlock && typeof firstBlock === 'object' && 'type' in firstBlock
              ? `[${firstBlock.type}]`
              : '[content]'
          }
        </div>
      )}
    </div>
  )
}

// Response content display
function ResponseContent({ response }: { response: { body?: unknown; bodyText?: string } }) {
  const [expanded, setExpanded] = useState(true)

  const body = response.body
  const content = body && typeof body === 'object' && 'content' in body
    ? (body as { content: unknown }).content
    : null

  return (
    <div className="rounded-lg border border-green-200 bg-green-50/50 overflow-hidden">
      <div
        className="flex items-center gap-2 px-3 py-2 bg-green-100/50 cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="w-6 h-6 rounded-full bg-green-600 text-white flex items-center justify-center flex-shrink-0">
          <Bot size={14} />
        </div>
        <span className="text-sm font-medium text-green-800">Response</span>
        <ChevronDown
          size={16}
          className={cn(
            'ml-auto text-gray-400 transition-transform',
            expanded && 'rotate-180'
          )}
        />
      </div>

      {expanded && (
        <div className="p-3">
          {content ? (
            <MessageContent content={content} />
          ) : (
            <pre className="text-xs font-mono bg-white rounded p-3 border border-green-200 overflow-auto max-h-96">
              {JSON.stringify(body || response.bodyText, null, 2)}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}

function RequestDetail({ requestId }: { requestId: string | null }) {
  const { data: request, isLoading } = useRequestDetail(requestId)

  if (!requestId) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Select a request to view details
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Loading...
      </div>
    )
  }

  if (!request) {
    return (
      <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
        Request not found
      </div>
    )
  }

  const status = request.response?.statusCode && request.response.statusCode >= 200 && request.response.statusCode < 300 ? 'success' : 'error'
  const messages = request.body?.messages || []

  return (
    <div className="p-4 h-full overflow-auto">
      {/* Header */}
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">
          Request Details
        </h2>
        <span
          className={cn(
            'px-2 py-0.5 text-xs rounded',
            status === 'success' && 'bg-green-500/20 text-green-400',
            status === 'error' && 'bg-red-500/20 text-red-400'
          )}
        >
          {status}
        </span>
      </div>

      {/* Metadata grid */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
          <p className="text-xs text-[var(--color-text-muted)] mb-1">Model</p>
          <p className="text-sm text-[var(--color-text-primary)]">{request.model || 'N/A'}</p>
        </div>
        <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
          <p className="text-xs text-[var(--color-text-muted)] mb-1">Provider</p>
          <p className="text-sm text-[var(--color-text-primary)]">{request.provider || 'N/A'}</p>
        </div>
        <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
          <p className="text-xs text-[var(--color-text-muted)] mb-1">Input Tokens</p>
          <p className="text-sm text-[var(--color-text-primary)]">
            {request.response?.body?.usage?.input_tokens?.toLocaleString() || 'N/A'}
          </p>
        </div>
        <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
          <p className="text-xs text-[var(--color-text-muted)] mb-1">Output Tokens</p>
          <p className="text-sm text-[var(--color-text-primary)]">
            {request.response?.body?.usage?.output_tokens?.toLocaleString() || 'N/A'}
          </p>
        </div>
        {request.response?.responseTime && (
          <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
            <p className="text-xs text-[var(--color-text-muted)] mb-1">Response Time</p>
            <p className="text-sm text-[var(--color-text-primary)]">
              {formatDuration(request.response.responseTime)}
            </p>
          </div>
        )}
        {request.subagentName && (
          <div className="p-3 rounded bg-[var(--color-bg-tertiary)]">
            <p className="text-xs text-[var(--color-text-muted)] mb-1">Subagent</p>
            <p className="text-sm text-[var(--color-text-primary)]">{request.subagentName}</p>
          </div>
        )}
      </div>

      {/* Messages */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-[var(--color-text-primary)]">
          Messages ({messages.length})
        </h3>

        <div className="space-y-3">
          {messages.map((message, index) => (
            <MessageBubble key={index} message={message} />
          ))}
        </div>

        {/* Response */}
        {request.response && (
          <div className="mt-6">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-3">
              Response
            </h3>
            <ResponseContent response={request.response} />
          </div>
        )}
      </div>
    </div>
  )
}

export function RequestsPage() {
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const { data: requests, isLoading } = useRequestsSummary({ limit: 100 })

  const filteredRequests = requests?.filter(r => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      r.model?.toLowerCase().includes(query) ||
      r.provider?.toLowerCase().includes(query) ||
      r.requestId.toLowerCase().includes(query)
    )
  }) || []

  return (
    <>
      <PageHeader
        title="Requests"
        description="View and analyze API requests"
        actions={
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="Search requests..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-7 pr-3 py-1.5 text-sm bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] rounded text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-[var(--color-accent)]"
              />
            </div>
            <button className="p-1.5 rounded bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:border-[var(--color-border-hover)]">
              <Filter size={14} />
            </button>
          </div>
        }
      />
      <div className="flex-1 overflow-hidden">
        <PanelGroup>
          <ResizablePanel defaultWidth={400} minWidth={300} maxWidth={600}>
            <div className="h-full overflow-auto bg-[var(--color-bg-secondary)] border-r border-[var(--color-border)]">
              {isLoading ? (
                <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                  Loading requests...
                </div>
              ) : filteredRequests.length === 0 ? (
                <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                  No requests found
                </div>
              ) : (
                filteredRequests.map((request) => (
                  <RequestListItem
                    key={request.requestId}
                    request={request}
                    isSelected={selectedRequestId === request.requestId}
                    onClick={() => setSelectedRequestId(request.requestId)}
                  />
                ))
              )}
            </div>
          </ResizablePanel>
          <Panel className="bg-[var(--color-bg-primary)]">
            <RequestDetail requestId={selectedRequestId} />
          </Panel>
        </PanelGroup>
      </div>
    </>
  )
}
