import { useState, useCallback } from 'react'
import { getSettings, saveSettings, type DashboardSettings } from '../storage'

export function useSettings() {
  // Initialize with getSettings directly, no need for useEffect
  const [settings, setSettingsState] = useState<DashboardSettings>(getSettings())

  const updateSettings = useCallback((updates: Partial<DashboardSettings>) => {
    setSettingsState(prev => {
      const updated = { ...prev, ...updates }
      saveSettings(updated)
      return updated
    })
  }, [])

  const resetSettings = useCallback(() => {
    localStorage.removeItem('claude-proxy-dashboard:settings')
    setSettingsState(getSettings())
  }, [])

  return {
    settings,
    updateSettings,
    resetSettings,
  }
}
