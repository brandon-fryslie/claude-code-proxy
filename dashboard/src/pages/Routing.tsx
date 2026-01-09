import { useState } from 'react'
import { PageHeader, PageContent } from '@/components/layout'
import { GitBranch, ArrowRight, Server, Settings, Key } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useSubagentStats,
  useProviders,
  useSubagentConfig,
  formatTokens,
  formatDuration,
} from '@/lib/api'
import { useDateRange } from '@/lib/DateRangeContext'
import { RefreshButton } from '@/components/features/RefreshButton'
import { useQueryClient } from '@tanstack/react-query'
import type { ProviderConfig } from '@/lib/types'

function ProviderCard({ name, config }: { name: string; config: ProviderConfig }) {
  const hasApiKey = config.api_key && config.api_key !== ''
  const isRedacted = config.api_key === '***REDACTED***'

  return (
    <div className="flex items-start gap-4 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <Server size={20} className="text-[var(--color-text-muted)] mt-0.5" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-2">
          <h3 className="font-medium text-[var(--color-text-primary)]">{name}</h3>
          <span
            className={cn(
              'text-xs px-2 py-0.5 rounded font-medium',
              config.format === 'anthropic'
                ? 'bg-purple-500/10 text-purple-400'
                : 'bg-green-500/10 text-green-400'
            )}
          >
            {config.format}
          </span>
          {(hasApiKey || isRedacted) && (
            <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-blue-500/10 text-blue-400">
              <Key size={12} />
              API Key
            </span>
          )}
        </div>

        <div className="space-y-1 text-sm text-[var(--color-text-secondary)]">
          <div className="font-mono text-xs">{config.base_url}</div>
          {config.version && (
            <div className="text-xs">
              <span className="text-[var(--color-text-muted)]">Version:</span> {config.version}
            </div>
          )}
          {config.max_retries !== undefined && (
            <div className="text-xs">
              <span className="text-[var(--color-text-muted)]">Max retries:</span> {config.max_retries}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

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
  const { dateRange } = useDateRange()
  const { data: subagentStats, isLoading: statsLoading } = useSubagentStats(dateRange)
  const { data: providers, isLoading: providersLoading } = useProviders()
  const { data: subagentConfig, isLoading: subagentLoading } = useSubagentConfig()

  const queryClient = useQueryClient()
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const totalRouted = subagentStats?.subagents?.reduce((acc, s) => acc + s.requests, 0) || 0
  const totalTokens = subagentStats?.subagents?.reduce((acc, s) => acc + s.totalTokens, 0) || 0
  const avgLatency =
    subagentStats && subagentStats.subagents && subagentStats.subagents.length > 0
      ? Math.round(
          subagentStats.subagents.reduce((acc, s) => acc + s.avgResponseMs * s.requests, 0) /
            totalRouted
        )
      : 0

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['stats', 'subagents'] }),
        queryClient.invalidateQueries({ queryKey: ['config', 'providers'] }),
        queryClient.invalidateQueries({ queryKey: ['config', 'subagents'] })
      ])
      setLastRefresh(new Date())
    } finally {
      setIsRefreshing(false)
    }
  }

  return (
    <>
      <PageHeader
        title="Provider Routing"
        description="View provider configuration and subagent routing"
        actions={
          <RefreshButton
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastRefresh={lastRefresh}
          />
        }
      />
      <PageContent>
        <div className="max-w-5xl space-y-8">
          {/* Providers Section */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <Server size={16} className="text-[var(--color-text-muted)]" />
              <h2 className="text-sm font-medium text-[var(--color-text-primary)]">
                Provider Configuration
              </h2>
            </div>

            {providersLoading ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                Loading providers...
              </div>
            ) : !providers || Object.keys(providers).length === 0 ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                <div className="text-center">
                  <p>No providers configured</p>
                  <p className="text-sm mt-1">Add providers in config.yaml</p>
                </div>
              </div>
            ) : (
              <div className="grid gap-4 md:grid-cols-2">
                {Object.entries(providers).map(([name, config]) => (
                  <ProviderCard key={name} name={name} config={config} />
                ))}
              </div>
            )}
          </div>

          {/* Subagent Routing Configuration */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <Settings size={16} className="text-[var(--color-text-muted)]" />
              <h2 className="text-sm font-medium text-[var(--color-text-primary)]">
                Subagent Routing
              </h2>
            </div>

            {subagentLoading ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                Loading subagent configuration...
              </div>
            ) : (
              <div className="space-y-4">
                {/* Status Badge */}
                <div className="flex items-center gap-2">
                  <span className="text-sm text-[var(--color-text-secondary)]">Status:</span>
                  <span
                    className={cn(
                      'text-xs px-2 py-1 rounded font-medium',
                      subagentConfig?.enable
                        ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                        : 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-muted)]'
                    )}
                  >
                    {subagentConfig?.enable ? 'Enabled' : 'Disabled'}
                  </span>
                </div>

                {/* Mappings */}
                {!subagentConfig?.enable ? (
                  <div className="p-4 rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-sm text-[var(--color-text-muted)]">
                    Subagent routing is disabled. Enable it in config.yaml to route subagents to
                    different providers.
                  </div>
                ) : !subagentConfig?.mappings ||
                  Object.keys(subagentConfig.mappings).length === 0 ? (
                  <div className="p-4 rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-sm text-[var(--color-text-muted)]">
                    No subagent mappings configured. Add mappings in config.yaml.
                  </div>
                ) : (
                  <div className="space-y-2">
                    {Object.entries(subagentConfig.mappings).map(([agentName, mapping]) => {
                      const [provider, model] = mapping.split(':')
                      return (
                        <div
                          key={agentName}
                          className="flex items-center gap-4 p-3 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]"
                        >
                          <div className="w-2 h-2 rounded-full bg-[var(--color-success)]" />
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-[var(--color-text-primary)]">
                                {agentName}
                              </span>
                            </div>
                            <div className="flex items-center gap-2 mt-1 text-xs text-[var(--color-text-secondary)]">
                              <span>subagent</span>
                              <ArrowRight size={10} className="text-[var(--color-text-muted)]" />
                              <span>
                                {provider}:{model}
                              </span>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Active Routes (Usage Statistics) */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <GitBranch size={16} className="text-[var(--color-text-muted)]" />
              <h2 className="text-sm font-medium text-[var(--color-text-primary)]">
                Active Routes
              </h2>
            </div>

            {statsLoading ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                Loading routing data...
              </div>
            ) : !subagentStats || !subagentStats.subagents || subagentStats.subagents.length === 0 ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                <div className="text-center">
                  <p>No subagent routing activity in selected period</p>
                  <p className="text-sm mt-1">Routes will appear here once they are used</p>
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
                    Routing Statistics
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
        </div>
      </PageContent>
    </>
  )
}
