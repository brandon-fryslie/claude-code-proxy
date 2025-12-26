import { PageHeader, PageContent } from '@/components/layout'
import { GitBranch, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSubagentStats, getTodayDateRange, formatTokens, formatDuration } from '@/lib/api'

function RouteRow({
  subagent,
  provider,
  model,
  requests,
  avgResponseMs,
  enabled = true,
}: {
  subagent: string
  provider: string
  model: string
  requests: number
  avgResponseMs: number
  enabled?: boolean
}) {
  return (
    <div className="flex items-center gap-4 p-3 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <div
        className={cn(
          'w-2 h-2 rounded-full',
          enabled ? 'bg-[var(--color-success)]' : 'bg-[var(--color-text-muted)]'
        )}
      />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-[var(--color-text-primary)]">{subagent}</span>
        </div>
        <div className="flex items-center gap-2 mt-1 text-xs text-[var(--color-text-secondary)]">
          <span>subagent</span>
          <ArrowRight size={10} className="text-[var(--color-text-muted)]" />
          <span>
            {provider}:{model}
          </span>
        </div>
      </div>
      <div className="text-right">
        <p className="text-sm font-medium text-[var(--color-text-primary)]">
          {requests.toLocaleString()}
        </p>
        <p className="text-xs text-[var(--color-text-muted)]">
          {formatDuration(avgResponseMs)} avg
        </p>
      </div>
    </div>
  )
}

export function RoutingPage() {
  const dateRange = getTodayDateRange()
  const { data: subagentStats, isLoading } = useSubagentStats(dateRange)

  const totalRouted = subagentStats?.subagents?.reduce((acc, s) => acc + s.requests, 0) || 0
  const totalTokens = subagentStats?.subagents?.reduce((acc, s) => acc + s.totalTokens, 0) || 0
  const avgLatency =
    subagentStats && subagentStats.subagents && subagentStats.subagents.length > 0
      ? Math.round(
          subagentStats.subagents.reduce((acc, s) => acc + s.avgResponseMs * s.requests, 0) /
            totalRouted
        )
      : 0

  return (
    <>
      <PageHeader
        title="Provider Routing"
        description="Configure and monitor subagent routing"
      />
      <PageContent>
        <div className="max-w-3xl">
          <div className="flex items-center gap-2 mb-4">
            <GitBranch size={16} className="text-[var(--color-text-muted)]" />
            <h2 className="text-sm font-medium text-[var(--color-text-primary)]">Active Routes</h2>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
              Loading routing data...
            </div>
          ) : !subagentStats || !subagentStats.subagents || subagentStats.subagents.length === 0 ? (
            <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
              <div className="text-center">
                <p>No subagent routing configured</p>
                <p className="text-sm mt-1">Configure subagent routing in config.yaml</p>
              </div>
            </div>
          ) : (
            <>
              <div className="space-y-2">
                {subagentStats.subagents.map((subagent) => (
                  <RouteRow
                    key={subagent.subagentName}
                    subagent={subagent.subagentName}
                    provider={subagent.provider}
                    model={subagent.targetModel}
                    requests={subagent.requests}
                    avgResponseMs={subagent.avgResponseMs}
                  />
                ))}
              </div>

              <div className="mt-8 p-4 rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)]">
                <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-2">
                  Routing Statistics (Today)
                </h3>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-xs text-[var(--color-text-muted)]">Total Routed</p>
                    <p className="text-lg font-semibold text-[var(--color-text-primary)]">
                      {totalRouted.toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-[var(--color-text-muted)]">Total Tokens</p>
                    <p className="text-lg font-semibold text-[var(--color-text-primary)]">
                      {formatTokens(totalTokens)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-[var(--color-text-muted)]">Avg Latency</p>
                    <p className="text-lg font-semibold text-[var(--color-text-primary)]">
                      {avgLatency ? formatDuration(avgLatency) : '--'}
                    </p>
                  </div>
                </div>
              </div>

              {/* Detailed Stats Table */}
              <div className="mt-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
                <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
                  Detailed Token Usage
                </h3>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="border-b border-[var(--color-border)]">
                      <tr className="text-left text-[var(--color-text-muted)]">
                        <th className="pb-2 pr-4">Subagent</th>
                        <th className="pb-2 pr-4">Provider:Model</th>
                        <th className="pb-2 pr-4 text-right">Requests</th>
                        <th className="pb-2 pr-4 text-right">Input Tokens</th>
                        <th className="pb-2 pr-4 text-right">Output Tokens</th>
                        <th className="pb-2 text-right">Total Tokens</th>
                      </tr>
                    </thead>
                    <tbody>
                      {subagentStats.subagents.map((stat, idx) => (
                        <tr
                          key={idx}
                          className="border-b border-[var(--color-border)] text-[var(--color-text-secondary)]"
                        >
                          <td className="py-2 pr-4">{stat.subagentName}</td>
                          <td className="py-2 pr-4">
                            {stat.provider}:{stat.targetModel}
                          </td>
                          <td className="py-2 pr-4 text-right">{stat.requests.toLocaleString()}</td>
                          <td className="py-2 pr-4 text-right">{formatTokens(stat.inputTokens)}</td>
                          <td className="py-2 pr-4 text-right">{formatTokens(stat.outputTokens)}</td>
                          <td className="py-2 text-right">{formatTokens(stat.totalTokens)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </>
          )}
        </div>
      </PageContent>
    </>
  )
}
