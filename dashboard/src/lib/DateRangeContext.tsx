import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { getWeekBoundaries, getLocalDayBoundaries, isSameWeek } from './utils'

export interface DateRange {
  start: string
  end: string
}

interface DateRangeContextValue {
  dateRange: DateRange
  selectedDate: Date
  setSelectedDate: (date: Date) => void
  presetRange: 'today' | 'week' | 'month' | 'custom'
  setPresetRange: (preset: 'today' | 'week' | 'month' | 'custom') => void
  currentWeekStart: Date | null
  navigateWeek: (direction: 'prev' | 'next') => void
  isAtToday: boolean
}

const DateRangeContext = createContext<DateRangeContextValue | null>(null)

function getDateRangeForPreset(preset: 'today' | 'week' | 'month' | 'custom', selectedDate: Date): DateRange {
  const now = new Date()

  switch (preset) {
    case 'today': {
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      return { start: start.toISOString(), end: end.toISOString() }
    }
    case 'week': {
      const { weekStart, weekEnd } = getWeekBoundaries(now)
      return { start: weekStart.toISOString(), end: weekEnd.toISOString() }
    }
    case 'month': {
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      const start = new Date(end)
      start.setDate(start.getDate() - 29)
      start.setHours(0, 0, 0, 0)
      return { start: start.toISOString(), end: end.toISOString() }
    }
    case 'custom': {
      // For custom, use selectedDate as the single day
      return getLocalDayBoundaries(selectedDate)
    }
  }
}

// LocalStorage key for persisting selected date
const STORAGE_KEY = 'dashboard.selectedDate'

// Load persisted date from localStorage, with validation
function loadPersistedDate(): Date | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (!stored) return null

    const date = new Date(stored)
    if (isNaN(date.getTime())) return null

    // If persisted date is >7 days old, ignore it
    const now = new Date()
    const daysDiff = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
    if (daysDiff > 7) return null

    // Don't allow future dates
    if (date > now) return null

    return date
  } catch {
    return null
  }
}

// Save date to localStorage
function persistDate(date: Date): void {
  try {
    localStorage.setItem(STORAGE_KEY, date.toISOString())
  } catch {
    // Silently fail if localStorage is not available
  }
}

interface DateRangeProviderProps {
  children: ReactNode
}

export function DateRangeProvider({ children }: DateRangeProviderProps) {
  // Check for URL param ?date=YYYY-MM-DD (overrides localStorage)
  const urlDate = (() => {
    const params = new URLSearchParams(window.location.search)
    const dateParam = params.get('date')
    if (!dateParam) return null

    const date = new Date(dateParam)
    if (isNaN(date.getTime())) return null

    // Don't allow future dates
    const now = new Date()
    if (date > now) return null

    return date
  })()

  const [selectedDate, setSelectedDateState] = useState(() => {
    return urlDate || loadPersistedDate() || new Date()
  })

  const [presetRange, setPresetRangeState] = useState<'today' | 'week' | 'month' | 'custom'>(() => {
    // If we loaded a custom date, start in custom mode
    if (urlDate || loadPersistedDate()) return 'custom'
    return 'today'
  })

  const [currentWeekStart, setCurrentWeekStart] = useState<Date | null>(() => {
    const { weekStart } = getWeekBoundaries(selectedDate)
    return weekStart
  })

  const dateRange = getDateRangeForPreset(presetRange, selectedDate)

  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const isAtToday = selectedDate >= today

  const setSelectedDate = useCallback((date: Date) => {
    const previousDate = selectedDate

    setSelectedDateState(date)
    persistDate(date)

    // If selecting a specific date, switch to custom mode
    const now = new Date()
    const isToday = date.getFullYear() === now.getFullYear() &&
                    date.getMonth() === now.getMonth() &&
                    date.getDate() === now.getDate()

    if (!isToday && presetRange !== 'custom') {
      setPresetRangeState('custom')
    } else if (isToday && presetRange === 'custom') {
      setPresetRangeState('today')
    }

    // Track current week start for smart refetching
    // Only update if we moved to a different week
    if (!isSameWeek(previousDate, date)) {
      const { weekStart } = getWeekBoundaries(date)
      setCurrentWeekStart(weekStart)
    }
  }, [presetRange, selectedDate])

  const setPresetRange = useCallback((preset: 'today' | 'week' | 'month' | 'custom') => {
    setPresetRangeState(preset)
    // Reset selected date to today when switching presets (except custom)
    if (preset !== 'custom') {
      const now = new Date()
      setSelectedDateState(now)
      persistDate(now)

      const { weekStart } = getWeekBoundaries(now)
      setCurrentWeekStart(weekStart)
    }
  }, [])

  const navigateWeek = useCallback((direction: 'prev' | 'next') => {
    const newDate = new Date(selectedDate)
    newDate.setDate(newDate.getDate() + (direction === 'next' ? 7 : -7))

    // Don't allow going into future
    const now = new Date()
    if (newDate > now) return

    setSelectedDate(newDate)
  }, [selectedDate, setSelectedDate])

  // Update URL param when date changes (for shareable links)
  useEffect(() => {
    if (presetRange === 'custom') {
      const dateStr = selectedDate.toISOString().split('T')[0] // YYYY-MM-DD
      const url = new URL(window.location.href)
      url.searchParams.set('date', dateStr)
      window.history.replaceState({}, '', url)
    } else {
      // Remove date param for preset ranges
      const url = new URL(window.location.href)
      url.searchParams.delete('date')
      window.history.replaceState({}, '', url)
    }
  }, [selectedDate, presetRange])

  return (
    <DateRangeContext.Provider value={{
      dateRange,
      selectedDate,
      setSelectedDate,
      presetRange,
      setPresetRange,
      currentWeekStart,
      navigateWeek,
      isAtToday,
    }}>
      {children}
    </DateRangeContext.Provider>
  )
}

export function useDateRange(): DateRangeContextValue {
  const context = useContext(DateRangeContext)
  if (!context) {
    throw new Error('useDateRange must be used within a DateRangeProvider')
  }
  return context
}
