import { useState } from 'react'
import { PageHeader, PageContent } from '@/components/layout'
import { BarChart3 } from 'lucide-react'
import { useModelStats, useWeeklyStats, getTodayDateRange, formatTokens } from '@/lib/api'
import { WeeklyUsageChart, ModelBreakdownChart, ModelComparisonBar, DateNavigation } from '@/components/charts'
import { toISODateString } from '@/lib/chartUtils'

export function UsagePage() {
  const [selectedDate, setSelectedDate] = useState(new Date())
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

  return (
    <>
      <PageHeader
        title="Token Usage"
        description="Analyze token consumption across providers and models"
      />
      <PageContent>
        {/* Stats Grid */}
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

        {/* Weekly Usage Chart with Date Navigation */}
        <div className="mb-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)]">
              Weekly Usage Trend
            </h3>
            <DateNavigation
              selectedDate={selectedDate}
              onDateChange={setSelectedDate}
              mode="day"
            />
          </div>
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
              Token Distribution by Model (Today)
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
              Model Comparison (Today)
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
        {modelStats && modelStats.modelStats.length > 0 && (
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
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
