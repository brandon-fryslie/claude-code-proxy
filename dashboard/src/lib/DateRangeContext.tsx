import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'

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
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      const start = new Date(end)
      start.setDate(start.getDate() - 6)
      start.setHours(0, 0, 0, 0)
      return { start: start.toISOString(), end: end.toISOString() }
    }
    case 'month': {
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      const start = new Date(end)
      start.setDate(start.getDate() - 29)
      start.setHours(0, 0, 0, 0)
      return { start: start.toISOString(), end: end.toISOString() }
    }
    case 'custom': {
      // For custom, use selectedDate as the end of range (showing that day)
      const start = new Date(selectedDate.getFullYear(), selectedDate.getMonth(), selectedDate.getDate())
      const end = new Date(selectedDate.getFullYear(), selectedDate.getMonth(), selectedDate.getDate(), 23, 59, 59)
      return { start: start.toISOString(), end: end.toISOString() }
    }
  }
}

interface DateRangeProviderProps {
  children: ReactNode
}

export function DateRangeProvider({ children }: DateRangeProviderProps) {
  const [selectedDate, setSelectedDateState] = useState(() => new Date())
  const [presetRange, setPresetRangeState] = useState<'today' | 'week' | 'month' | 'custom'>('today')

  const dateRange = getDateRangeForPreset(presetRange, selectedDate)

  const setSelectedDate = useCallback((date: Date) => {
    setSelectedDateState(date)
    // If selecting a specific date, switch to custom mode
    const today = new Date()
    const isToday = date.getFullYear() === today.getFullYear() &&
                    date.getMonth() === today.getMonth() &&
                    date.getDate() === today.getDate()
    if (!isToday && presetRange !== 'custom') {
      setPresetRangeState('custom')
    }
  }, [presetRange])

  const setPresetRange = useCallback((preset: 'today' | 'week' | 'month' | 'custom') => {
    setPresetRangeState(preset)
    // Reset selected date to today when switching presets (except custom)
    if (preset !== 'custom') {
      setSelectedDateState(new Date())
    }
  }, [])

  return (
    <DateRangeContext.Provider value={{
      dateRange,
      selectedDate,
      setSelectedDate,
      presetRange,
      setPresetRange,
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
