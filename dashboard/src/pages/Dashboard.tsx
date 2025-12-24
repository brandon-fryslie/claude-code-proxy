import { PageHeader, PageContent } from '@/components/layout'
import { Activity, BarChart3, Clock, Zap } from 'lucide-react'
import { useHourlyStats, useProviderStats, getTodayDateRange, formatDuration, formatTokens } from '@/lib/api'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

interface StatCardProps {
  label: string
  value: string
  subValue?: string
  icon: React.ReactNode
  isLoading?: boolean
}

function StatCard({ label, value, subValue, icon, isLoading }: StatCardProps) {
  return (
    <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">{label}</p>
          {isLoading ? (
            <p className="text-2xl font-semibold text-[var(--color-text-muted)] mt-1">--</p>
          ) : (
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">{value}</p>
          )}
          {subValue && (
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">{subValue}</p>
          )}
        </div>
        <div className="p-2 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-muted)]">
          {icon}
        </div>
      </div>
    </div>
  )
}

const CHART_COLORS = ['#8b5cf6', '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#ec4899']

export function DashboardPage() {
  const dateRange = getTodayDateRange()
  const { data: hourlyStats, isLoading: isLoadingHourly } = useHourlyStats(dateRange)
  const { data: providerStats, isLoading: isLoadingProviders } = useProviderStats(dateRange)

  // Prepare chart data
  const hourlyChartData = hourlyStats?.hourlyStats.map(h => ({
    hour: `${h.hour}:00`,
    requests: h.requests,
    tokens: h.tokens,
  })) || []

  const providerChartData = providerStats?.providers.map(p => ({
    name: p.provider,
    tokens: p.totalTokens,
    requests: p.requests,
  })) || []

  // Calculate stats
  const todayRequests = hourlyStats?.todayRequests || 0
  const todayTokens = hourlyStats?.todayTokens || 0
  const avgResponseTime = hourlyStats?.avgResponseTime || 0
  const activeProviders = providerStats?.providers.length || 0

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="Overview of proxy activity and usage"
      />
      <PageContent>
        {/* Stats Grid */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <StatCard
            label="Requests Today"
            value={todayRequests.toLocaleString()}
            subValue={`${formatTokens(todayTokens)} tokens`}
            icon={<Activity size={20} />}
            isLoading={isLoadingHourly}
          />
          <StatCard
            label="Tokens Used"
            value={formatTokens(todayTokens)}
            subValue="Today's total"
            icon={<BarChart3 size={20} />}
            isLoading={isLoadingHourly}
          />
          <StatCard
            label="Avg Response Time"
            value={avgResponseTime ? formatDuration(avgResponseTime) : '--'}
            subValue="Today's average"
            icon={<Clock size={20} />}
            isLoading={isLoadingHourly}
          />
          <StatCard
            label="Active Providers"
            value={activeProviders.toString()}
            subValue={providerStats?.providers.map(p => p.provider).join(', ') || 'None'}
            icon={<Zap size={20} />}
            isLoading={isLoadingProviders}
          />
        </div>

        {/* Charts */}
        <div className="grid grid-cols-2 gap-4">
          {/* Request Volume Chart */}
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)] h-80">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">Request Volume Today</h3>
            {isLoadingHourly ? (
              <div className="flex items-center justify-center h-64 text-[var(--color-text-muted)]">
                Loading...
              </div>
            ) : hourlyChartData.length === 0 ? (
              <div className="flex items-center justify-center h-64 text-[var(--color-text-muted)]">
                No data available
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="85%">
                <BarChart data={hourlyChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                  <XAxis
                    dataKey="hour"
                    stroke="var(--color-text-muted)"
                    style={{ fontSize: '12px' }}
                  />
                  <YAxis
                    stroke="var(--color-text-muted)"
                    style={{ fontSize: '12px' }}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--color-bg-tertiary)',
                      border: '1px solid var(--color-border)',
                      borderRadius: '6px',
                      color: 'var(--color-text-primary)',
                    }}
                  />
                  <Bar dataKey="requests" fill="#8b5cf6" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>

          {/* Token Usage by Provider Chart */}
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)] h-80">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">Token Usage by Provider</h3>
            {isLoadingProviders ? (
              <div className="flex items-center justify-center h-64 text-[var(--color-text-muted)]">
                Loading...
              </div>
            ) : providerChartData.length === 0 ? (
              <div className="flex items-center justify-center h-64 text-[var(--color-text-muted)]">
                No data available
              </div>
            ) : (
              <div className="flex items-center h-64">
                <ResponsiveContainer width="60%" height="100%">
                  <PieChart>
                    <Pie
                      data={providerChartData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="tokens"
                    >
                      {providerChartData.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={CHART_COLORS[index % CHART_COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--color-bg-tertiary)',
                        border: '1px solid var(--color-border)',
                        borderRadius: '6px',
                        color: 'var(--color-text-primary)',
                      }}
                      formatter={(value) => value ? formatTokens(value as number) : "0"}
                    />
                  </PieChart>
                </ResponsiveContainer>
                <div className="flex-1 pl-4">
                  <div className="space-y-2">
                    {providerChartData.map((entry, index) => (
                      <div key={entry.name} className="flex items-center gap-2">
                        <div
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: CHART_COLORS[index % CHART_COLORS.length] }}
                        />
                        <span className="text-xs text-[var(--color-text-secondary)] flex-1">
                          {entry.name}
                        </span>
                        <span className="text-xs text-[var(--color-text-primary)] font-medium">
                          {formatTokens(entry.tokens)}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </PageContent>
    </>
  )
}
