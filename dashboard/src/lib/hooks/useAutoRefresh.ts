import { useEffect, useRef, useCallback } from 'react'
import { useSettings } from './useSettings'

interface UseAutoRefreshOptions {
  onRefresh: () => void | Promise<void>
  enabled?: boolean // Override settings
}

export function useAutoRefresh({ onRefresh, enabled }: UseAutoRefreshOptions) {
  const { settings } = useSettings()
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const isEnabled = enabled ?? settings.autoRefreshEnabled
  const intervalMs = settings.autoRefreshInterval * 1000

  const refresh = useCallback(async () => {
    await onRefresh()
  }, [onRefresh])

  // Set up interval
  useEffect(() => {
    if (!isEnabled) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
      return
    }

    intervalRef.current = setInterval(refresh, intervalMs)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [isEnabled, intervalMs, refresh])

  // Manual refresh trigger
  const triggerRefresh = useCallback(() => {
    refresh()
  }, [refresh])

  return {
    triggerRefresh,
    isAutoRefreshEnabled: isEnabled,
  }
}
