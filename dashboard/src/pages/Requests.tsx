import { useState, useRef } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { PageHeader } from '@/components/layout'
import { ResizablePanel, PanelGroup, Panel } from '@/components/layout'
import { cn } from '@/lib/utils'
import { ChevronRight, Clock, ArrowRight, Search, GitCompare, User, Bot } from 'lucide-react'
import { useRequestsSummary, useRequestDetail, formatDuration, clearAllRequests } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import type { RequestSummary as RequestSummaryType, RequestLog } from '@/lib/types'
import { CompareModeBanner } from '@/components/features/CompareModeBanner'
import { RequestCompareModal } from '@/components/features/RequestCompareModal'
import { DataManagementBar } from '@/components/features/DataManagementBar'
import { MessageContent } from '@/components/ui'

interface CompareState {
  enabled: boolean
  selectedIds: string[]
}

function RequestListItem({
  request,
  isSelected,
  onClick,
  compareMode,
  isCompareSelected,
  onCompareToggle,
}: {
  request: RequestSummaryType
  isSelected: boolean
  onClick: () => void
  compareMode: boolean
  isCompareSelected: boolean
  onCompareToggle: () => void
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
      {/* Compare mode checkbox */}
      {compareMode && (
        <input
          type="checkbox"
          checked={isCompareSelected}
          onChange={(e) => {
            e.stopPropagation()
            onCompareToggle()
          }}
          className="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
      )}
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

  return (
    <div className="p-4 h-full overflow-auto">
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

      <div className="space-y-4">
        {/* Request Messages */}
        <div>
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-2">Request Messages</h3>
          <div className="space-y-3 max-h-[50vh] overflow-auto">
            {request.body?.messages?.map((msg, idx) => (
              <div
                key={idx}
                className={cn(
                  'rounded-lg border p-3',
                  msg.role === 'user' ? 'bg-blue-500/10 border-blue-500/30 ml-4' : 'bg-[var(--color-bg-tertiary)] border-[var(--color-border)] mr-4'
                )}
              >
                <div className="flex items-center gap-2 mb-2">
                  {msg.role === 'user' ? (
                    <User size={14} className="text-blue-400" />
                  ) : (
                    <Bot size={14} className="text-green-400" />
                  )}
                  <span className="text-xs font-medium text-[var(--color-text-secondary)] uppercase">
                    {msg.role}
                  </span>
                </div>
                <MessageContent content={msg.content} />
              </div>
            )) || (
              <div className="p-3 rounded bg-[var(--color-bg-tertiary)] font-mono text-xs text-[var(--color-text-secondary)] overflow-auto">
                <pre>{JSON.stringify(request.body, null, 2)}</pre>
              </div>
            )}
          </div>
        </div>

        {/* Response Content */}
        {request.response && (
          <div>
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-2">Response</h3>
            <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] p-3 max-h-[50vh] overflow-auto">
              {request.response.body?.content ? (
                <MessageContent content={request.response.body.content} />
              ) : request.response.bodyText ? (
                <pre className="font-mono text-xs text-[var(--color-text-secondary)] whitespace-pre-wrap">
                  {request.response.bodyText}
                </pre>
              ) : (
                <div className="text-sm text-[var(--color-text-muted)] italic">No response content</div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export function RequestsPage() {
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [compareState, setCompareState] = useState<CompareState>({
    enabled: false,
    selectedIds: [],
  })
  const [showCompareModal, setShowCompareModal] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())
  const [compareRequests, setCompareRequests] = useState<{ request1: RequestLog; request2: RequestLog } | null>(null)
  const [isLoadingCompare, setIsLoadingCompare] = useState(false)

  // Ref for virtualization scroll container
  const parentRef = useRef<HTMLDivElement>(null)

  const queryClient = useQueryClient()
  // Remove limit to fetch all requests
  const { data: requests, isLoading, refetch } = useRequestsSummary()

  const filteredRequests = requests?.filter(r => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      r.model?.toLowerCase().includes(query) ||
      r.provider?.toLowerCase().includes(query) ||
      r.requestId.toLowerCase().includes(query)
    )
  }) || []

  // Set up virtualizer for requests list
  const virtualizer = useVirtualizer({
    count: filteredRequests.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60, // Estimated height of RequestListItem (smaller than old dashboard)
    overscan: 5, // Render 5 extra items above and below viewport for smooth scrolling
  })

  const toggleCompareMode = () => {
    setCompareState({
      enabled: !compareState.enabled,
      selectedIds: [],
    })
  }

  const toggleRequestSelection = (id: string) => {
    setCompareState(prev => {
      if (prev.selectedIds.includes(id)) {
        return {
          ...prev,
          selectedIds: prev.selectedIds.filter(x => x !== id),
        }
      }
      // Max 2 selected
      if (prev.selectedIds.length >= 2) {
        return {
          ...prev,
          selectedIds: [prev.selectedIds[1], id], // Remove oldest, add new
        }
      }
      return {
        ...prev,
        selectedIds: [...prev.selectedIds, id],
      }
    })
  }

  const handleCompare = async () => {
    if (compareState.selectedIds.length !== 2) return

    setIsLoadingCompare(true)
    try {
      // Fetch both requests - first try cache, then fetch
      const [id1, id2] = compareState.selectedIds

      let request1 = queryClient.getQueryData<RequestLog>(['requests', 'detail', id1])
      let request2 = queryClient.getQueryData<RequestLog>(['requests', 'detail', id2])

      // Fetch if not in cache
      if (!request1) {
        request1 = await queryClient.fetchQuery({
          queryKey: ['requests', 'detail', id1],
          queryFn: async () => {
            const response = await fetch(`/api/requests/${id1}`)
            if (!response.ok) throw new Error('Failed to fetch request')
            return response.json()
          }
        })
      }

      if (!request2) {
        request2 = await queryClient.fetchQuery({
          queryKey: ['requests', 'detail', id2],
          queryFn: async () => {
            const response = await fetch(`/api/requests/${id2}`)
            if (!response.ok) throw new Error('Failed to fetch request')
            return response.json()
          }
        })
      }

      if (request1 && request2) {
        setCompareRequests({ request1, request2 })
        setShowCompareModal(true)
      }
    } catch (error) {
      console.error('Failed to fetch requests for comparison:', error)
    } finally {
      setIsLoadingCompare(false)
    }
  }

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await refetch()
      setLastRefresh(new Date())
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleClearData = async () => {
    await clearAllRequests()
    queryClient.invalidateQueries({ queryKey: ['requests'] })
    setSelectedRequestId(null)
  }

  const handleCloseCompareModal = () => {
    setShowCompareModal(false)
    setCompareRequests(null)
  }

  return (
    <>
      {compareState.enabled && (
        <CompareModeBanner
          selectedCount={compareState.selectedIds.length}
          onCompare={handleCompare}
          onCancel={toggleCompareMode}
          isLoading={isLoadingCompare}
        />
      )}
      <PageHeader
        title="Requests"
        description="View and analyze API requests"
        actions={
          <div className="flex items-center gap-2">
            <DataManagementBar
              onRefresh={handleRefresh}
              onClearData={handleClearData}
              isRefreshing={isRefreshing}
              lastRefresh={lastRefresh}
            />
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
            <button
              onClick={toggleCompareMode}
              className={cn(
                "flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors",
                compareState.enabled
                  ? "bg-indigo-600 text-white"
                  : "bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:border-[var(--color-border-hover)]"
              )}
            >
              <GitCompare size={14} />
              {compareState.enabled ? 'Cancel' : 'Compare'}
            </button>
          </div>
        }
      />
      <div className="flex-1 overflow-hidden">
        <PanelGroup>
          <ResizablePanel defaultWidth={400} minWidth={300} maxWidth={600}>
            <div className="h-full flex flex-col bg-[var(--color-bg-secondary)] border-r border-[var(--color-border)]">
              {isLoading ? (
                <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                  Loading requests...
                </div>
              ) : filteredRequests.length === 0 ? (
                <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                  No requests found
                </div>
              ) : (
                <div
                  ref={parentRef}
                  className="flex-1 overflow-auto"
                >
                  <div
                    style={{
                      height: `${virtualizer.getTotalSize()}px`,
                      width: '100%',
                      position: 'relative',
                    }}
                  >
                    {virtualizer.getVirtualItems().map((virtualItem) => {
                      const request = filteredRequests[virtualItem.index]
                      return (
                        <div
                          key={request.requestId}
                          style={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            width: '100%',
                            transform: `translateY(${virtualItem.start}px)`,
                          }}
                        >
                          <RequestListItem
                            request={request}
                            isSelected={selectedRequestId === request.requestId}
                            onClick={() => setSelectedRequestId(request.requestId)}
                            compareMode={compareState.enabled}
                            isCompareSelected={compareState.selectedIds.includes(request.requestId)}
                            onCompareToggle={() => toggleRequestSelection(request.requestId)}
                          />
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          </ResizablePanel>
          <Panel className="bg-[var(--color-bg-primary)]">
            <RequestDetail requestId={selectedRequestId} />
          </Panel>
        </PanelGroup>
      </div>

      {/* Compare Modal */}
      {showCompareModal && compareRequests && (
        <RequestCompareModal
          request1={compareRequests.request1}
          request2={compareRequests.request2}
          onClose={handleCloseCompareModal}
        />
      )}
    </>
  )
}
