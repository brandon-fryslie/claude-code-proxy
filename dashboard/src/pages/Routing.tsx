import { PageHeader, PageContent } from '@/components/layout'
import { GitBranch, ArrowRight, Server, Settings, Check, X, AlertCircle, Activity } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useSubagentStats,
  useRoutingConfig,
  useProviderHealth,
  formatTokens,
  formatDuration,
} from '@/lib/api'
import { useDateRange } from '@/lib/DateRangeContext'
import type { ProviderHealth } from '@/lib/types'

function ProviderHealthCard({ provider }: { provider: ProviderHealth }) {
  const getStatusIcon = () => {
    if (!provider.healthy) {
      return <X className="w-5 h-5 text-red-500" />
    }
    if (provider.circuit_breaker_state === 'half-open') {
      return <AlertCircle className="w-5 h-5 text-yellow-500" />
    }
    return <Check className="w-5 h-5 text-green-500" />
  }

  const getCircuitBreakerBadge = () => {
    if (!provider.circuit_breaker_state) return null

    const colors = {
      closed: 'bg-green-500/10 text-green-400',
      'half-open': 'bg-yellow-500/10 text-yellow-400',
      open: 'bg-red-500/10 text-red-400',
    }

    return (
      <span className={cn('px-2 py-1 rounded text-xs font-medium', colors[provider.circuit_breaker_state])}>
        Circuit: {provider.circuit_breaker_state}
      </span>
    )
  }

  return (
    <div className="flex items-start gap-4 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <div className="flex-shrink-0 mt-0.5">{getStatusIcon()}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-2">
          <h3 className="font-medium text-[var(--color-text-primary)]">{provider.name}</h3>
          {getCircuitBreakerBadge()}
        </div>

        <div className="space-y-1 text-sm text-[var(--color-text-secondary)]">
          <div className="flex items-center gap-2">
            <span className="text-[var(--color-text-muted)]">Status:</span>
            <span className={provider.healthy ? 'text-green-400' : 'text-red-400'}>
              {provider.healthy ? 'Healthy' : 'Unhealthy'}
            </span>
          </div>

          {provider.fallback_provider && (
            <div className="flex items-center gap-2">
              <span className="text-[var(--color-text-muted)]">Fallback:</span>
              <span className="font-mono text-xs">{provider.fallback_provider}</span>
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
  const { data: routingConfig, isLoading: configLoading } = useRoutingConfig()
  const { data: providerHealth, isLoading: healthLoading } = useProviderHealth()

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
        description="Monitor provider health, circuit breakers, and subagent routing"
      />
      <PageContent>
        <div className="max-w-5xl space-y-8">
          {/* Provider Health Section */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <Activity size={16} className="text-[var(--color-text-muted)]" />
              <h2 className="text-sm font-medium text-[var(--color-text-primary)]">
                Provider Health
              </h2>
            </div>

            {healthLoading ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                Loading provider health...
              </div>
            ) : !providerHealth || providerHealth.length === 0 ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                <div className="text-center">
                  <p>No provider health information available</p>
                  <p className="text-sm mt-1">Configure providers in config.yaml</p>
                </div>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {providerHealth.map((provider) => (
                  <ProviderHealthCard key={provider.name} provider={provider} />
                ))}
              </div>
            )}
          </div>

          {/* Providers Configuration Section */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <Server size={16} className="text-[var(--color-text-muted)]" />
              <h2 className="text-sm font-medium text-[var(--color-text-primary)]">
                Provider Configuration
              </h2>
            </div>

            {configLoading ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                Loading configuration...
              </div>
            ) : !routingConfig || Object.keys(routingConfig.providers).length === 0 ? (
              <div className="flex items-center justify-center h-32 text-[var(--color-text-muted)]">
                <div className="text-center">
                  <p>No providers configured</p>
                  <p className="text-sm mt-1">Configure providers in config.yaml</p>
                </div>
              </div>
            ) : (
              <div className="space-y-2">
                {Object.entries(routingConfig.providers).map(([name, config]) => (
                  <div
                    key={name}
                    className="flex items-start gap-4 p-3 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]"
                  >
                    <Server size={16} className="text-[var(--color-text-muted)] mt-0.5" />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-sm font-medium text-[var(--color-text-primary)]">
                          {name}
                        </span>
                        <span
                          className={cn(
                            'text-xs px-2 py-0.5 rounded',
                            config.format === 'anthropic'
                              ? 'bg-purple-500/10 text-purple-400'
                              : 'bg-green-500/10 text-green-400'
                          )}
                        >
                          {config.format}
                        </span>
                        {config.circuit_breaker?.enabled && (
                          <span className="text-xs px-2 py-0.5 rounded bg-blue-500/10 text-blue-400">
                            Circuit Breaker
                          </span>
                        )}
                      </div>
                      <div className="space-y-1 text-xs text-[var(--color-text-secondary)]">
                        <div>{config.base_url}</div>
                        {config.fallback_provider && (
                          <div className="flex items-center gap-1">
                            <span className="text-[var(--color-text-muted)]">Fallback:</span>
                            <span>{config.fallback_provider}</span>
                          </div>
                        )}
                        {config.circuit_breaker?.enabled && (
                          <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
                            <span>Max failures: {config.circuit_breaker.max_failures}</span>
                            <span>•</span>
                            <span>Timeout: {config.circuit_breaker.timeout}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
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

            {configLoading ? (
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
                      routingConfig?.subagents?.enable
                        ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                        : 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-muted)]'
                    )}
                  >
                    {routingConfig?.subagents?.enable ? 'Enabled' : 'Disabled'}
                  </span>
                </div>

                {/* Mappings */}
                {!routingConfig?.subagents?.enable ? (
                  <div className="p-4 rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-sm text-[var(--color-text-muted)]">
                    Subagent routing is disabled. Enable it in config.yaml to route subagents to
                    different providers.
                  </div>
                ) : !routingConfig?.subagents?.mappings ||
                  Object.keys(routingConfig.subagents.mappings).length === 0 ? (
                  <div className="p-4 rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-sm text-[var(--color-text-muted)]">
                    No subagent mappings configured. Add mappings in config.yaml.
                  </div>
                ) : (
                  <div className="space-y-2">
                    {Object.entries(routingConfig.subagents.mappings).map(([agentName, mapping]) => {
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
