import { PageHeader, PageContent } from '@/components/layout'
import { BarChart3 } from 'lucide-react'
import { useModelStats, useWeeklyStats, getTodayDateRange, formatTokens } from '@/lib/api'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'

export function UsagePage() {
  const dateRange = getTodayDateRange()
  const { data: modelStats, isLoading: isLoadingModels } = useModelStats(dateRange)
  const { data: weeklyStats, isLoading: isLoadingWeekly } = useWeeklyStats()

  // Calculate today's tokens
  const todayTokens = modelStats?.modelStats.reduce((acc, m) => acc + m.tokens, 0) || 0

  // Calculate week tokens (last 7 days)
  const weekTokens =
    weeklyStats?.dailyStats.slice(-7).reduce((acc, day) => acc + day.tokens, 0) || 0

  // Calculate month tokens (last 30 days)
  const monthTokens =
    weeklyStats?.dailyStats.slice(-30).reduce((acc, day) => acc + day.tokens, 0) || 0

  // Prepare chart data
  const modelChartData =
    modelStats?.modelStats.map((m) => ({
      model: m.model,
      tokens: m.tokens,
      requests: m.requests,
    })) || []

  const weeklyChartData =
    weeklyStats?.dailyStats.slice(-7).map((day) => ({
      date: new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      tokens: day.tokens,
      requests: day.requests,
    })) || []

  return (
    <>
      <PageHeader
        title="Token Usage"
        description="Analyze token consumption across providers and models"
      />
      <PageContent>
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Today</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingModels ? '--' : formatTokens(todayTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">This Week</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingWeekly ? '--' : formatTokens(weekTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">This Month</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingWeekly ? '--' : formatTokens(monthTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
        </div>

        {/* Token Usage by Model */}
        <div className="mb-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
            Token Usage by Model (Today)
          </h3>
          {isLoadingModels ? (
            <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>Loading usage data...</p>
            </div>
          ) : modelChartData.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>No usage data available</p>
              <p className="text-sm mt-1">Make some requests to see token usage</p>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={modelChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                <XAxis
                  dataKey="model"
                  stroke="var(--color-text-muted)"
                  style={{ fontSize: '12px' }}
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis stroke="var(--color-text-muted)" style={{ fontSize: '12px' }} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--color-bg-tertiary)',
                    border: '1px solid var(--color-border)',
                    borderRadius: '6px',
                    color: 'var(--color-text-primary)',
                  }}
                  formatter={(value) => value ? formatTokens(value as number) : "0"}
                />
                <Bar dataKey="tokens" fill="#8b5cf6" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Weekly Trend */}
        <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
            Weekly Trend (Last 7 Days)
          </h3>
          {isLoadingWeekly ? (
            <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>Loading trend data...</p>
            </div>
          ) : weeklyChartData.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>No trend data available</p>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={weeklyChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                <XAxis
                  dataKey="date"
                  stroke="var(--color-text-muted)"
                  style={{ fontSize: '12px' }}
                />
                <YAxis stroke="var(--color-text-muted)" style={{ fontSize: '12px' }} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--color-bg-tertiary)',
                    border: '1px solid var(--color-border)',
                    borderRadius: '6px',
                    color: 'var(--color-text-primary)',
                  }}
                  formatter={(value, name) =>
                    value && name === "tokens" ? formatTokens(value as number) : value || 0
                  }
                />
                <Legend />
                <Bar dataKey="tokens" fill="#8b5cf6" name="Tokens" />
                <Bar dataKey="requests" fill="#3b82f6" name="Requests" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Model Stats Table */}
        {modelStats && modelStats.modelStats.length > 0 && (
          <div className="mt-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
              Detailed Breakdown (Today)
            </h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-[var(--color-border)]">
                  <tr className="text-left text-[var(--color-text-muted)]">
                    <th className="pb-2 pr-4">Model</th>
                    <th className="pb-2 pr-4 text-right">Requests</th>
                    <th className="pb-2 pr-4 text-right">Total Tokens</th>
                    <th className="pb-2 text-right">Avg Tokens/Request</th>
                  </tr>
                </thead>
                <tbody>
                  {modelStats.modelStats.map((stat, idx) => (
                    <tr
                      key={idx}
                      className="border-b border-[var(--color-border)] text-[var(--color-text-secondary)]"
                    >
                      <td className="py-2 pr-4">{stat.model}</td>
                      <td className="py-2 pr-4 text-right">{stat.requests.toLocaleString()}</td>
                      <td className="py-2 pr-4 text-right">{formatTokens(stat.tokens)}</td>
                      <td className="py-2 text-right">
                        {stat.requests > 0
                          ? formatTokens(Math.round(stat.tokens / stat.requests))
                          : '--'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </PageContent>
    </>
  )
}
