// Local storage utilities for dashboard settings

const STORAGE_PREFIX = 'claude-proxy-dashboard:'

export interface DashboardSettings {
  autoRefreshEnabled: boolean
  autoRefreshInterval: number // seconds
  defaultModelFilter: string
  darkMode: boolean
  notifyOnError: boolean
  notifyOnHighLatency: boolean
  highLatencyThreshold: number // ms
  dataRetentionDays: number
}

const DEFAULT_SETTINGS: DashboardSettings = {
  autoRefreshEnabled: false,
  autoRefreshInterval: 30,
  defaultModelFilter: 'all',
  darkMode: false,
  notifyOnError: true,
  notifyOnHighLatency: false,
  highLatencyThreshold: 5000,
  dataRetentionDays: 30,
}

export function getSettings(): DashboardSettings {
  try {
    const stored = localStorage.getItem(STORAGE_PREFIX + 'settings')
    if (stored) {
      return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) }
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
  return DEFAULT_SETTINGS
}

export function saveSettings(settings: Partial<DashboardSettings>): void {
  try {
    const current = getSettings()
    const updated = { ...current, ...settings }
    localStorage.setItem(STORAGE_PREFIX + 'settings', JSON.stringify(updated))
  } catch (e) {
    console.error('Failed to save settings:', e)
  }
}

export function getLastSelectedDate(): string | null {
  return localStorage.getItem(STORAGE_PREFIX + 'selectedDate')
}

export function saveSelectedDate(date: string): void {
  localStorage.setItem(STORAGE_PREFIX + 'selectedDate', date)
}

export function clearAllStorage(): void {
  Object.keys(localStorage)
    .filter(key => key.startsWith(STORAGE_PREFIX))
    .forEach(key => localStorage.removeItem(key))
}
