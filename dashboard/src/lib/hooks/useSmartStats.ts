import { useMemo } from 'react'
import { useDateRange } from '../DateRangeContext'
import { useWeeklyStats, useHourlyStats, useModelStats } from '../api'
import { getWeekBoundaries } from '../utils'

/**
 * Smart stats fetching hook that implements the week-aware caching pattern.
 *
 * When navigating within the same week:
 * - Weekly stats are cached (1 query, served from cache)
 * - Hourly stats refetch (1 API call)
 * - Model stats refetch (1 API call)
 * Total: 2 API calls
 *
 * When navigating to a different week:
 * - Weekly stats refetch (1 API call)
 * - Hourly stats refetch (1 API call)
 * - Model stats refetch (1 API call)
 * Total: 3 API calls
 *
 * TanStack Query handles the caching automatically based on query keys.
 */
export function useSmartStats() {
  const { dateRange, currentWeekStart } = useDateRange()

  // Calculate week boundaries for weekly stats query
  // Using currentWeekStart ensures the query key only changes when week changes
  const weekRange = useMemo(() => {
    if (!currentWeekStart) {
      // Fallback to calculating from dateRange.start
      const startDate = new Date(dateRange.start)
      const { weekStart, weekEnd } = getWeekBoundaries(startDate)
      return {
        start: weekStart.toISOString(),
        end: weekEnd.toISOString(),
      }
    }

    const { weekStart, weekEnd } = getWeekBoundaries(currentWeekStart)
    return {
      start: weekStart.toISOString(),
      end: weekEnd.toISOString(),
    }
  }, [currentWeekStart, dateRange.start])

  // Weekly stats: query key includes weekRange, so stays cached within same week
  const weeklyStatsQuery = useWeeklyStats(weekRange)

  // Hourly stats: query key includes dateRange, so refetches on day change
  const hourlyStatsQuery = useHourlyStats(dateRange)

  // Model stats: query key includes dateRange, so refetches on day change
  const modelStatsQuery = useModelStats(dateRange)

  return {
    // Weekly stats (cached within same week)
    weeklyStats: weeklyStatsQuery.data,
    isLoadingWeekly: weeklyStatsQuery.isLoading,
    weeklyError: weeklyStatsQuery.error,

    // Hourly stats (refetches on day change)
    hourlyStats: hourlyStatsQuery.data,
    isLoadingHourly: hourlyStatsQuery.isLoading,
    hourlyError: hourlyStatsQuery.error,

    // Model stats (refetches on day change)
    modelStats: modelStatsQuery.data,
    isLoadingModel: modelStatsQuery.isLoading,
    modelError: modelStatsQuery.error,

    // Overall loading state
    isLoading: weeklyStatsQuery.isLoading || hourlyStatsQuery.isLoading || modelStatsQuery.isLoading,
  }
}
