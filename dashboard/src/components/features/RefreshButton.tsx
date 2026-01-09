import { type FC } from 'react'
import { RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'

interface RefreshButtonProps {
  onRefresh: () => Promise<void>
  isRefreshing?: boolean
  lastRefresh?: Date
  showLabel?: boolean
}

export const RefreshButton: FC<RefreshButtonProps> = ({
  onRefresh,
  isRefreshing,
  lastRefresh,
  showLabel = true,
}) => {
  const formatTimeSince = (date: Date) => {
    const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
    if (seconds < 60) return `${seconds}s ago`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    return `${hours}h ago`
  }

  return (
    <button
      onClick={onRefresh}
      disabled={isRefreshing}
      className={cn(
        'flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors',
        'bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-[var(--color-text-primary)]',
        'hover:bg-[var(--color-bg-hover)]',
        isRefreshing && 'opacity-50 cursor-not-allowed'
      )}
      title={lastRefresh ? `Last refresh: ${lastRefresh.toLocaleTimeString()}` : 'Refresh data'}
    >
      <RefreshCw className={cn('w-4 h-4', isRefreshing && 'animate-spin')} />
      {showLabel && 'Refresh'}
      {lastRefresh && !showLabel && (
        <span className="text-xs text-[var(--color-text-muted)]">
          {formatTimeSince(lastRefresh)}
        </span>
      )}
    </button>
  )
}
