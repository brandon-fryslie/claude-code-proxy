import { useState } from 'react'
import { AppLayout } from '@/components/layout'
import { useTodos, useTodoDetail, usePlans, usePlanDetail, reindexSessionData } from '@/lib/api'
import type { TodoSession, PlanSummary } from '@/lib/types'
import { Search, RefreshCw, CheckCircle2, Circle, Clock, File } from 'lucide-react'

type Tab = 'todos' | 'plans'

export default function SessionData() {
  const [activeTab, setActiveTab] = useState<Tab>('todos')
  const [selectedSession, setSelectedSession] = useState<string | null>(null)
  const [selectedPlan, setSelectedPlan] = useState<number | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [isReindexing, setIsReindexing] = useState(false)

  const { data: todosData, isLoading: todosLoading, refetch: refetchTodos } = useTodos()
  const { data: plansData, isLoading: plansLoading, refetch: refetchPlans } = usePlans()
  const { data: todoDetail } = useTodoDetail(selectedSession)
  const { data: planDetail } = usePlanDetail(selectedPlan)

  const handleRefresh = async () => {
    setIsReindexing(true)
    try {
      await reindexSessionData()
      await refetchTodos()
      await refetchPlans()
    } catch (error) {
      console.error('Reindex failed:', error)
    } finally {
      setIsReindexing(false)
    }
  }

  const formatRelativeTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins}m ago`
    const diffHours = Math.floor(diffMins / 60)
    if (diffHours < 24) return `${diffHours}h ago`
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays}d ago`
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
  }

  const truncateUuid = (uuid: string) => {
    return `${uuid.slice(0, 8)}...${uuid.slice(-4)}`
  }

  const filteredSessions = todosData?.sessions?.filter((session: TodoSession) =>
    session.session_uuid.toLowerCase().includes(searchTerm.toLowerCase())
  ) || []

  const filteredPlans = plansData?.plans?.filter((plan: PlanSummary) =>
    plan.display_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    plan.preview.toLowerCase().includes(searchTerm.toLowerCase())
  ) || []

  return (
    <AppLayout title="Session Data" description="Debug logs, todos, plans, and session environment">
      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <div className="flex space-x-8">
          <button
            onClick={() => setActiveTab('todos')}
            className={`pb-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'todos'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Todos
          </button>
          <button
            onClick={() => setActiveTab('plans')}
            className={`pb-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'plans'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Plans
          </button>
        </div>
      </div>

      {activeTab === 'todos' && (
        <div className="space-y-6">
          {/* Summary Cards */}
          {todosData && (
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white p-4 rounded-lg border border-gray-200">
                <div className="text-2xl font-bold text-gray-900">{todosData.total_files}</div>
                <div className="text-sm text-gray-500">Total Files</div>
              </div>
              <div className="bg-white p-4 rounded-lg border border-gray-200">
                <div className="text-2xl font-bold text-gray-900">{todosData.non_empty_files}</div>
                <div className="text-sm text-gray-500">Non-empty</div>
              </div>
              <div className="bg-white p-4 rounded-lg border border-gray-200">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Circle className="h-3 w-3 text-gray-400" />
                    <span className="text-sm">{todosData.status_breakdown.pending} Pending</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Clock className="h-3 w-3 text-blue-500" />
                    <span className="text-sm">{todosData.status_breakdown.in_progress} In Progress</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    <span className="text-sm">{todosData.status_breakdown.completed} Completed</span>
                  </div>
                </div>
              </div>
              <div className="bg-white p-4 rounded-lg border border-gray-200 flex items-center justify-between">
                <div>
                  <div className="text-xs text-gray-500">Last indexed</div>
                  <div className="text-sm text-gray-900">
                    {todosData.last_indexed ? formatRelativeTime(todosData.last_indexed) : 'Never'}
                  </div>
                </div>
                <button
                  onClick={handleRefresh}
                  disabled={isReindexing}
                  className="p-2 rounded-lg hover:bg-gray-100 transition-colors disabled:opacity-50"
                  title="Refresh index"
                >
                  <RefreshCw className={`h-4 w-4 ${isReindexing ? 'animate-spin' : ''}`} />
                </button>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Sessions List */}
            <div className="bg-white rounded-lg border border-gray-200">
              <div className="p-4 border-b border-gray-200">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    placeholder="Search sessions..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
              <div className="divide-y divide-gray-200 max-h-96 overflow-y-auto">
                {todosLoading ? (
                  <div className="p-8 text-center text-gray-500">Loading...</div>
                ) : filteredSessions.length === 0 ? (
                  <div className="p-8 text-center text-gray-500">No sessions found</div>
                ) : (
                  filteredSessions.map((session: TodoSession) => (
                    <button
                      key={session.file_path}
                      onClick={() => setSelectedSession(session.agent_uuid)}
                      className={`w-full text-left p-4 hover:bg-gray-50 transition-colors ${
                        selectedSession === session.agent_uuid ? 'bg-blue-50 border-l-4 border-blue-500' : ''
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="font-mono text-sm text-gray-900">
                            {truncateUuid(session.session_uuid)}
                            {session.agent_uuid !== session.session_uuid && (
                              <span className="text-gray-400 ml-1">/{truncateUuid(session.agent_uuid)}</span>
                            )}
                          </div>
                          <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                            <span>{session.todo_count} todos</span>
                            <span className="flex items-center gap-1">
                              <CheckCircle2 className="h-3 w-3 text-green-500" />
                              {session.completed_count}
                            </span>
                            <span className="flex items-center gap-1">
                              <Clock className="h-3 w-3 text-blue-500" />
                              {session.in_progress_count}
                            </span>
                            <span className="flex items-center gap-1">
                              <Circle className="h-3 w-3 text-gray-400" />
                              {session.pending_count}
                            </span>
                          </div>
                        </div>
                        <div className="text-xs text-gray-400 ml-4">{formatRelativeTime(session.modified_at)}</div>
                      </div>
                    </button>
                  ))
                )}
              </div>
            </div>

            {/* Session Detail */}
            <div className="bg-white rounded-lg border border-gray-200">
              <div className="p-4 border-b border-gray-200">
                <h3 className="font-medium text-gray-900">Session Detail</h3>
              </div>
              <div className="p-4 max-h-96 overflow-y-auto">
                {!selectedSession ? (
                  <div className="text-center text-gray-500 py-8">Select a session to view details</div>
                ) : !todoDetail ? (
                  <div className="text-center text-gray-500 py-8">Loading...</div>
                ) : (
                  <div className="space-y-4">
                    <div>
                      <div className="text-xs text-gray-500">Session UUID</div>
                      <div className="font-mono text-sm text-gray-900">{todoDetail.session_uuid}</div>
                    </div>
                    <div>
                      <div className="text-xs text-gray-500">Agent UUID</div>
                      <div className="font-mono text-sm text-gray-900">{todoDetail.agent_uuid || 'N/A'}</div>
                    </div>
                    <div className="border-t border-gray-200 pt-4">
                      <div className="text-sm font-medium text-gray-900 mb-2">Todos</div>
                      <div className="space-y-2">
                        {todoDetail.todos.map((todo, idx) => (
                          <div key={idx} className="flex items-start gap-2">
                            {todo.status === 'completed' && (
                              <CheckCircle2 className="h-4 w-4 text-green-500 flex-shrink-0 mt-0.5" />
                            )}
                            {todo.status === 'in_progress' && (
                              <Clock className="h-4 w-4 text-blue-500 flex-shrink-0 mt-0.5" />
                            )}
                            {todo.status === 'pending' && (
                              <Circle className="h-4 w-4 text-gray-400 flex-shrink-0 mt-0.5" />
                            )}
                            <div className="flex-1 min-w-0">
                              <div className="text-sm text-gray-900">{todo.content}</div>
                              {todo.active_form && (
                                <div className="text-xs text-gray-500 italic mt-1">{todo.active_form}</div>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'plans' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Plans List */}
          <div className="bg-white rounded-lg border border-gray-200">
            <div className="p-4 border-b border-gray-200">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search plans..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              {plansData && (
                <div className="mt-3 text-xs text-gray-500">
                  {plansData.total_count} plans ({formatBytes(plansData.total_size)})
                </div>
              )}
            </div>
            <div className="divide-y divide-gray-200 max-h-96 overflow-y-auto">
              {plansLoading ? (
                <div className="p-8 text-center text-gray-500">Loading...</div>
              ) : filteredPlans.length === 0 ? (
                <div className="p-8 text-center text-gray-500">No plans found</div>
              ) : (
                filteredPlans.map((plan: PlanSummary) => (
                  <button
                    key={plan.id}
                    onClick={() => setSelectedPlan(plan.id)}
                    className={`w-full text-left p-4 hover:bg-gray-50 transition-colors ${
                      selectedPlan === plan.id ? 'bg-blue-50 border-l-4 border-blue-500' : ''
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-sm text-gray-900">{plan.display_name}</div>
                        <div className="text-xs text-gray-500 mt-1 line-clamp-2">{plan.preview}</div>
                        <div className="flex items-center gap-2 mt-2 text-xs text-gray-400">
                          <File className="h-3 w-3" />
                          {formatBytes(plan.file_size)}
                        </div>
                      </div>
                      <div className="text-xs text-gray-400 ml-4">{formatRelativeTime(plan.modified_at)}</div>
                    </div>
                  </button>
                ))
              )}
            </div>
          </div>

          {/* Plan Detail */}
          <div className="bg-white rounded-lg border border-gray-200">
            <div className="p-4 border-b border-gray-200">
              <h3 className="font-medium text-gray-900">Plan Content</h3>
            </div>
            <div className="p-4 max-h-96 overflow-y-auto">
              {!selectedPlan ? (
                <div className="text-center text-gray-500 py-8">Select a plan to view content</div>
              ) : !planDetail ? (
                <div className="text-center text-gray-500 py-8">Loading...</div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <div className="text-xs text-gray-500">File</div>
                    <div className="text-sm text-gray-900">{planDetail.file_name}</div>
                  </div>
                  <div>
                    <div className="text-xs text-gray-500">Size</div>
                    <div className="text-sm text-gray-900">{formatBytes(planDetail.file_size)}</div>
                  </div>
                  <div className="border-t border-gray-200 pt-4">
                    <div className="text-sm font-medium text-gray-900 mb-2">Content</div>
                    <div className="prose prose-sm max-w-none">
                      <pre className="text-xs bg-gray-50 p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">
                        {planDetail.content}
                      </pre>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </AppLayout>
  )
}
