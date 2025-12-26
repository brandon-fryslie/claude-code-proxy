import { PageHeader, PageContent } from '@/components/layout'
import { Zap } from 'lucide-react'
import { usePerformanceStats, getTodayDateRange, formatDuration } from '@/lib/api'
import { PerformanceChart } from '@/components/charts'

export function PerformancePage() {
  const dateRange = getTodayDateRange()
  const { data: perfStats, isLoading } = usePerformanceStats(dateRange)

  // Calculate overall stats
  const overallStats = perfStats?.stats?.reduce(
    (acc, stat) => ({
      avgResponseMs: acc.avgResponseMs + stat.avgResponseMs * stat.requestCount,
      p50ResponseMs: acc.p50ResponseMs + stat.p50ResponseMs * stat.requestCount,
      p95ResponseMs: acc.p95ResponseMs + stat.p95ResponseMs * stat.requestCount,
      p99ResponseMs: acc.p99ResponseMs + stat.p99ResponseMs * stat.requestCount,
      requestCount: acc.requestCount + stat.requestCount,
    }),
    { avgResponseMs: 0, p50ResponseMs: 0, p95ResponseMs: 0, p99ResponseMs: 0, requestCount: 0 }
  )

  const avgResponse = overallStats && overallStats.requestCount > 0
    ? Math.round(overallStats.avgResponseMs / overallStats.requestCount)
    : 0

  const p50Response = overallStats && overallStats.requestCount > 0
    ? Math.round(overallStats.p50ResponseMs / overallStats.requestCount)
    : 0

  const p95Response = overallStats && overallStats.requestCount > 0
    ? Math.round(overallStats.p95ResponseMs / overallStats.requestCount)
    : 0

  const p99Response = overallStats && overallStats.requestCount > 0
    ? Math.round(overallStats.p99ResponseMs / overallStats.requestCount)
    : 0

  return (
    <>
      <PageHeader
        title="Performance"
        description="Response times and latency analysis"
      />
      <PageContent>
        {/* Stats Grid */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Avg Response</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoading ? '--' : avgResponse ? formatDuration(avgResponse) : '--'}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">Today's average</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">P50</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoading ? '--' : p50Response ? formatDuration(p50Response) : '--'}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">Median</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">P95</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoading ? '--' : p95Response ? formatDuration(p95Response) : '--'}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">95th percentile</p>
          </div>
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">P99</p>
            <p className="text-2xl font-semibold text-[var(--color-text-primary)] mt-1">
              {isLoading ? '--' : p99Response ? formatDuration(p99Response) : '--'}
            </p>
            <p className="text-xs text-[var(--color-text-secondary)] mt-1">99th percentile</p>
          </div>
        </div>

        {/* Performance Chart */}
        <div className="mb-6 p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
          <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">
            Response Times by Provider & Model
          </h3>
          {isLoading ? (
            <div className="flex flex-col items-center justify-center h-96 text-[var(--color-text-muted)]">
              <Zap size={48} className="mb-4 opacity-50" />
              <p>Loading performance data...</p>
            </div>
          ) : !perfStats?.stats || perfStats.stats.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-96 text-[var(--color-text-muted)]">
              <Zap size={48} className="mb-4 opacity-50" />
              <p>No performance data available</p>
              <p className="text-sm mt-1">Make some requests to see performance metrics</p>
            </div>
          ) : (
            <PerformanceChart data={perfStats.stats} height={400} />
          )}
        </div>

        {/* Detailed Stats Table */}
        {perfStats && perfStats.stats && perfStats.stats.length > 0 && (
          <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
            <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-4">Detailed Statistics</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-[var(--color-border)]">
                  <tr className="text-left text-[var(--color-text-muted)]">
                    <th className="pb-2 pr-4">Provider</th>
                    <th className="pb-2 pr-4">Model</th>
                    <th className="pb-2 pr-4 text-right">Requests</th>
                    <th className="pb-2 pr-4 text-right">Avg</th>
                    <th className="pb-2 pr-4 text-right">P50</th>
                    <th className="pb-2 pr-4 text-right">P95</th>
                    <th className="pb-2 pr-4 text-right">P99</th>
                    <th className="pb-2 text-right">Avg TTFB</th>
                  </tr>
                </thead>
                <tbody>
                  {perfStats.stats.map((stat, idx) => (
                    <tr key={idx} className="border-b border-[var(--color-border)] text-[var(--color-text-secondary)]">
                      <td className="py-2 pr-4">{stat.provider}</td>
                      <td className="py-2 pr-4">{stat.model}</td>
                      <td className="py-2 pr-4 text-right">{stat.requestCount}</td>
                      <td className="py-2 pr-4 text-right">{formatDuration(stat.avgResponseMs)}</td>
                      <td className="py-2 pr-4 text-right">{formatDuration(stat.p50ResponseMs)}</td>
                      <td className="py-2 pr-4 text-right">{formatDuration(stat.p95ResponseMs)}</td>
                      <td className="py-2 pr-4 text-right">{formatDuration(stat.p99ResponseMs)}</td>
                      <td className="py-2 text-right">{formatDuration(stat.avgFirstByteMs)}</td>
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
