import { type FC, useState } from 'react'
import { RefreshCw, Trash2, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface DataManagementBarProps {
  onRefresh: () => Promise<void>
  onClearData: () => Promise<void>
  isRefreshing?: boolean
  lastRefresh?: Date
}

export const DataManagementBar: FC<DataManagementBarProps> = ({
  onRefresh,
  onClearData,
  isRefreshing,
  lastRefresh,
}) => {
  const [showClearConfirm, setShowClearConfirm] = useState(false)
  const [isClearing, setIsClearing] = useState(false)

  const handleClear = async () => {
    setIsClearing(true)
    try {
      await onClearData()
      setShowClearConfirm(false)
    } finally {
      setIsClearing(false)
    }
  }

  return (
    <div className="flex items-center gap-2">
      {/* Refresh button */}
      <button
        onClick={onRefresh}
        disabled={isRefreshing}
        className={cn(
          'flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors',
          'bg-gray-100 hover:bg-gray-200 text-gray-700',
          isRefreshing && 'opacity-50 cursor-not-allowed'
        )}
        title={lastRefresh ? `Last refresh: ${lastRefresh.toLocaleTimeString()}` : 'Refresh data'}
      >
        <RefreshCw className={cn('w-4 h-4', isRefreshing && 'animate-spin')} />
        Refresh
      </button>

      {/* Clear data button */}
      <div className="relative">
        <button
          onClick={() => setShowClearConfirm(true)}
          className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
        >
          <Trash2 className="w-4 h-4" />
          Clear Data
        </button>

        {/* Confirmation popover */}
        {showClearConfirm && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setShowClearConfirm(false)}
            />
            <div className="absolute top-full mt-2 right-0 z-50 bg-white rounded-lg shadow-lg border p-4 w-64">
              <div className="text-sm font-medium text-gray-900 mb-2">
                Clear all request data?
              </div>
              <div className="text-xs text-gray-500 mb-4">
                This will permanently delete all logged requests. This action cannot be undone.
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setShowClearConfirm(false)}
                  className="flex-1 px-3 py-1.5 text-sm rounded bg-gray-100 hover:bg-gray-200 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleClear}
                  disabled={isClearing}
                  className="flex-1 px-3 py-1.5 text-sm rounded bg-red-600 text-white hover:bg-red-700 transition-colors disabled:opacity-50"
                >
                  {isClearing ? (
                    <Loader2 className="w-4 h-4 mx-auto animate-spin" />
                  ) : (
                    'Delete All'
                  )}
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
