import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ============================================================================
// Date Utilities
// ============================================================================

export interface WeekBoundaries {
  weekStart: Date
  weekEnd: Date
}

/**
 * Get Sunday-Saturday week boundaries for a given date.
 * Week starts on Sunday at 00:00:00 and ends on Saturday at 23:59:59.999
 *
 * @param date - The date to calculate week boundaries for
 * @returns Object with weekStart (Sunday 00:00:00) and weekEnd (Saturday 23:59:59.999)
 */
export function getWeekBoundaries(date: Date): WeekBoundaries {
  const weekStart = new Date(date)
  weekStart.setHours(0, 0, 0, 0)
  const dayOfWeek = weekStart.getDay() // 0 = Sunday
  weekStart.setDate(weekStart.getDate() - dayOfWeek) // Go back to Sunday

  const weekEnd = new Date(weekStart)
  weekEnd.setDate(weekEnd.getDate() + 6) // Saturday
  weekEnd.setHours(23, 59, 59, 999)

  return { weekStart, weekEnd }
}

/**
 * Get start and end of a specific day in local timezone
 *
 * @param date - The date to get boundaries for
 * @returns Object with start (00:00:00) and end (23:59:59.999) as ISO strings
 */
export function getLocalDayBoundaries(date: Date): { start: string; end: string } {
  const startOfDay = new Date(date)
  startOfDay.setHours(0, 0, 0, 0)

  const endOfDay = new Date(date)
  endOfDay.setHours(23, 59, 59, 999)

  return {
    start: startOfDay.toISOString(),
    end: endOfDay.toISOString()
  }
}

/**
 * Format a date for display (e.g., "Dec 24")
 */
export function formatShortDate(date: Date): string {
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

/**
 * Format a date range for display (e.g., "Dec 22 - Dec 28")
 */
export function formatWeekRange(weekStart: Date, weekEnd: Date): string {
  // If same month, show "Dec 22 - 28"
  if (weekStart.getMonth() === weekEnd.getMonth()) {
    const monthYear = weekStart.toLocaleDateString('en-US', { month: 'short' })
    return `${monthYear} ${weekStart.getDate()} - ${weekEnd.getDate()}`
  }
  // Different months, show full range
  return `${formatShortDate(weekStart)} - ${formatShortDate(weekEnd)}`
}

/**
 * Check if two dates are in the same week (Sunday-Saturday)
 */
export function isSameWeek(date1: Date, date2: Date): boolean {
  const week1 = getWeekBoundaries(date1)
  const week2 = getWeekBoundaries(date2)
  return week1.weekStart.getTime() === week2.weekStart.getTime()
}
