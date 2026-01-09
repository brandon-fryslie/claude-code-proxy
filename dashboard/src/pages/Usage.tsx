import { useState } from 'react'
import { PageHeader, PageContent } from '@/components/layout'
import { BarChart3 } from 'lucide-react'
import { useModelStats, useWeeklyStats, formatTokens } from '@/lib/api'
import { useDateRange } from '@/lib/DateRangeContext'
import { WeeklyUsageChart, ModelBreakdownChart, ModelComparisonBar } from '@/components/charts'
import { toISODateString } from '@/lib/chartUtils'
import { RefreshButton } from '@/components/features/RefreshButton'
import { useQueryClient } from '@tanstack/react-query'

export function UsagePage() {
  const { dateRange, selectedDate, setSelectedDate, presetRange } = useDateRange()

  const { data: modelStats, isLoading: isLoadingModels } = useModelStats(dateRange)
  const { data: weeklyStats, isLoading: isLoadingWeekly } = useWeeklyStats()

  const queryClient = useQueryClient()
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  // Calculate tokens for selected range
  const rangeTokens = modelStats?.modelStats?.reduce((acc, m) => acc + m.tokens, 0) || 0
  const rangeRequests = modelStats?.modelStats?.reduce((acc, m) => acc + m.requests, 0) || 0

  // Calculate week tokens (last 7 days)
  const weekTokens =
    weeklyStats?.dailyStats?.slice(-7).reduce((acc, day) => acc + day.tokens, 0) || 0

  // Calculate month tokens (last 30 days)
  const monthTokens =
    weeklyStats?.dailyStats?.slice(-30).reduce((acc, day) => acc + day.tokens, 0) || 0

  // Format range label based on preset
  const rangeLabel = presetRange === 'today' ? 'Today' :
                     presetRange === 'week' ? 'Last 7 Days' :
                     presetRange === 'month' ? 'Last 30 Days' :
                     selectedDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['stats', 'models'] }),
        queryClient.invalidateQueries({ queryKey: ['stats', 'weekly'] })
      ])
      setLastRefresh(new Date())
    } finally {
      setIsRefreshing(false)
    }
  }

  return (
    <>
      <PageHeader
        title="Token Usage"
        description="Analyze token consumption across providers and models"
        actions={
          <RefreshButton
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastRefresh={lastRefresh}
          />
        }
      />
      <PageContent>
        {/* Stats Grid */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">{rangeLabel}</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingModels ? '--' : formatTokens(rangeTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Requests</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingModels ? '--' : rangeRequests.toLocaleString()}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">{rangeLabel.toLowerCase()}</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Last 7 Days</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingWeekly ? '--' : formatTokens(weekTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Last 30 Days</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoadingWeekly ? '--' : formatTokens(monthTokens)}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">tokens</p>
          </div>
        </div>

        {/* Weekly Usage Chart */}
        <div className="mb-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
            Weekly Usage Trend
          </h3>
          {isLoadingWeekly ? (
            <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>Loading usage data...</p>
            </div>
          ) : !weeklyStats?.dailyStats || weeklyStats.dailyStats.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
              <BarChart3 size={48} className="mb-4 opacity-50" />
              <p>No usage data available</p>
              <p className="text-sm mt-1">Make some requests to see token usage</p>
            </div>
          ) : (
            <WeeklyUsageChart
              data={weeklyStats.dailyStats.slice(-7)}
              selectedDate={toISODateString(selectedDate)}
              onDateSelect={(date) => setSelectedDate(new Date(date))}
              height={320}
            />
          )}
        </div>

        {/* Model Breakdown Charts */}
        <div className="grid grid-cols-2 gap-4 mb-6">
          {/* Pie Chart */}
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
              Token Distribution by Model
            </h3>
            {isLoadingModels ? (
              <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
                <BarChart3 size={48} className="mb-4 opacity-50" />
                <p>Loading model data...</p>
              </div>
            ) : !modelStats?.modelStats || modelStats.modelStats.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
                <BarChart3 size={48} className="mb-4 opacity-50" />
                <p>No model data available</p>
              </div>
            ) : (
              <ModelBreakdownChart
                data={modelStats.modelStats}
                metric="tokens"
                height={320}
              />
            )}
          </div>

          {/* Bar Chart */}
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
              Model Comparison
            </h3>
            {isLoadingModels ? (
              <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
                <BarChart3 size={48} className="mb-4 opacity-50" />
                <p>Loading model data...</p>
              </div>
            ) : !modelStats?.modelStats || modelStats.modelStats.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-80 text-[var(--color-text-muted)]">
                <BarChart3 size={48} className="mb-4 opacity-50" />
                <p>No model data available</p>
              </div>
            ) : (
              <ModelComparisonBar
                data={modelStats.modelStats}
                height={320}
              />
            )}
          </div>
        </div>

        {/* Model Stats Table */}
        {modelStats && modelStats.modelStats && modelStats.modelStats.length > 0 && (
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
              Detailed Breakdown
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
