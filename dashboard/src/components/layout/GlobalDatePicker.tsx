import { type FC } from 'react'
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react'
import { cn, getWeekBoundaries, formatWeekRange } from '@/lib/utils'
import { useDateRange } from '@/lib/DateRangeContext'

export const GlobalDatePicker: FC = () => {
  const { selectedDate, setSelectedDate, presetRange, setPresetRange, navigateWeek, isAtToday } = useDateRange()

  const goBack = () => {
    const newDate = new Date(selectedDate)
    newDate.setDate(newDate.getDate() - 1)
    setSelectedDate(newDate)
  }

  const goForward = () => {
    if (isAtToday) return
    const newDate = new Date(selectedDate)
    newDate.setDate(newDate.getDate() + 1)
    setSelectedDate(newDate)
  }

  const goToToday = () => {
    setSelectedDate(new Date())
    setPresetRange('today')
  }

  const formatDate = (date: Date): string => {
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    })
  }

  const { weekStart, weekEnd } = getWeekBoundaries(selectedDate)
  const weekRangeDisplay = formatWeekRange(weekStart, weekEnd)

  const presets: { value: 'today' | 'week' | 'month'; label: string }[] = [
    { value: 'today', label: 'Today' },
    { value: 'week', label: '7d' },
    { value: 'month', label: '30d' },
  ]

  return (
    <div className="flex items-center gap-3">
      {/* Preset buttons */}
      <div className="flex items-center bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-0.5">
        {presets.map(({ value, label }) => (
          <button
            key={value}
            onClick={() => setPresetRange(value)}
            className={cn(
              'px-3 py-1 text-xs font-medium rounded-md transition-colors',
              presetRange === value
                ? 'bg-[var(--color-bg-active)] text-[var(--color-text-primary)]'
                : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
            )}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Date navigation - Day mode */}
      <div className="flex items-center gap-1">
        <button
          onClick={goBack}
          className="p-1.5 rounded-md hover:bg-[var(--color-bg-hover)] transition-colors text-[var(--color-text-secondary)]"
          title="Previous day"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <div className="flex items-center gap-1.5 px-2 py-1 bg-[var(--color-bg-secondary)] rounded-md border border-[var(--color-border)] min-w-[140px] justify-center">
          <Calendar className="w-3.5 h-3.5 text-[var(--color-text-muted)]" />
          <span className="text-xs font-medium text-[var(--color-text-primary)]">
            {formatDate(selectedDate)}
          </span>
        </div>

        <button
          onClick={goForward}
          disabled={isAtToday}
          className={cn(
            'p-1.5 rounded-md transition-colors',
            isAtToday
              ? 'text-[var(--color-text-muted)] cursor-not-allowed'
              : 'hover:bg-[var(--color-bg-hover)] text-[var(--color-text-secondary)]'
          )}
          title="Next day"
        >
          <ChevronRight className="w-4 h-4" />
        </button>

        {!isAtToday && (
          <button
            onClick={goToToday}
            className="ml-1 px-2 py-1 text-xs text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-md transition-colors"
          >
            Today
          </button>
        )}
      </div>

      {/* Week navigation */}
      <div className="flex items-center gap-1 border-l border-[var(--color-border)] pl-3">
        <button
          onClick={() => navigateWeek('prev')}
          className="p-1.5 rounded-md hover:bg-[var(--color-bg-hover)] transition-colors text-[var(--color-text-secondary)]"
          title="Previous week"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <div className="px-2 py-1 bg-[var(--color-bg-secondary)] rounded-md border border-[var(--color-border)] min-w-[140px] text-center">
          <span className="text-xs font-medium text-[var(--color-text-primary)]">
            {weekRangeDisplay}
          </span>
        </div>

        <button
          onClick={() => navigateWeek('next')}
          disabled={isAtToday}
          className={cn(
            'p-1.5 rounded-md transition-colors',
            isAtToday
              ? 'text-[var(--color-text-muted)] cursor-not-allowed'
              : 'hover:bg-[var(--color-bg-hover)] text-[var(--color-text-secondary)]'
          )}
          title="Next week"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  )
}
