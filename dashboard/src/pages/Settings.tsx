import { type FC, useState } from 'react'
import { PageHeader, PageContent } from '@/components/layout'
import { Bell, Clock, Database, RotateCcw, Trash2 } from 'lucide-react'
import { useSettings } from '@/lib/hooks/useSettings'
import { clearAllRequests, useRequestsSummary } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import { ConfirmDeleteModal } from '@/components/features/ConfirmDeleteModal'
import { cn } from '@/lib/utils'

interface SettingsSectionProps {
  title: string
  description: string
  icon: React.ReactNode
  children: React.ReactNode
}

function SettingsSection({ title, description, icon, children }: SettingsSectionProps) {
  return (
    <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <div className="flex items-start gap-3 mb-4">
        <div className="p-2 rounded bg-[var(--color-bg-tertiary)] text-[var(--color-text-muted)]">
          {icon}
        </div>
        <div>
          <h3 className="text-sm font-medium text-[var(--color-text-primary)]">{title}</h3>
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">{description}</p>
        </div>
      </div>
      <div className="space-y-3 ml-12">{children}</div>
    </div>
  )
}

const ToggleSetting: FC<{
  label: string
  description?: string
  checked: boolean
  onChange: (checked: boolean) => void
  disabled?: boolean
}> = ({ label, description, checked, onChange, disabled }) => (
  <label className={cn("flex items-start gap-3 cursor-pointer", disabled && "opacity-50 cursor-not-allowed")}>
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => onChange(e.target.checked)}
      disabled={disabled}
      className="mt-1 w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
    />
    <div>
      <div className="text-sm font-medium text-[var(--color-text-primary)]">{label}</div>
      {description && <div className="text-xs text-[var(--color-text-muted)]">{description}</div>}
    </div>
  </label>
)

const SelectSetting: FC<{
  label: string
  value: string
  options: { value: string; label: string }[]
  onChange: (value: string) => void
  disabled?: boolean
}> = ({ label, value, options, onChange, disabled }) => (
  <div className={cn("flex items-center justify-between", disabled && "opacity-50")}>
    <span className="text-sm text-[var(--color-text-primary)]">{label}</span>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      className="px-3 py-1.5 text-sm border rounded-lg bg-[var(--color-bg-tertiary)] border-[var(--color-border)] text-[var(--color-text-primary)] focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
    >
      {options.map(opt => (
        <option key={opt.value} value={opt.value}>{opt.label}</option>
      ))}
    </select>
  </div>
)

const NumberSetting: FC<{
  label: string
  value: number
  onChange: (value: number) => void
  suffix?: string
  min?: number
  max?: number
  step?: number
  disabled?: boolean
}> = ({ label, value, onChange, suffix, min, max, step, disabled }) => (
  <div className={cn("flex items-center justify-between", disabled && "opacity-50")}>
    <span className="text-sm text-[var(--color-text-primary)]">{label}</span>
    <div className="flex items-center gap-2">
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        min={min}
        max={max}
        step={step}
        disabled={disabled}
        className="w-24 px-3 py-1.5 text-sm border rounded-lg text-right focus:ring-2 focus:ring-blue-500 bg-[var(--color-bg-tertiary)] border-[var(--color-border)] text-[var(--color-text-primary)]"
      />
      {suffix && <span className="text-sm text-[var(--color-text-muted)]">{suffix}</span>}
    </div>
  </div>
)

export function SettingsPage() {
  const { settings, updateSettings, resetSettings } = useSettings()
  const queryClient = useQueryClient()
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  // Get request count for confirmation modal
  const { data: requests } = useRequestsSummary()
  const requestCount = requests?.length || 0

  const handleClearAll = async () => {
    await clearAllRequests()
    // Invalidate all queries to refresh the UI
    queryClient.invalidateQueries()
  }

  return (
    <>
      <PageHeader
        title="Settings"
        description="Configure dashboard preferences"
        actions={
          <button
            onClick={resetSettings}
            className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)] transition-colors"
          >
            <RotateCcw className="w-4 h-4" />
            Reset to Defaults
          </button>
        }
      />

      <PageContent>
        <div className="max-w-2xl space-y-4">
          {/* Auto-Refresh Section */}
          <SettingsSection
            icon={<Clock size={18} />}
            title="Auto-Refresh"
            description="Automatically refresh data at regular intervals"
          >
            <ToggleSetting
              label="Enable auto-refresh"
              checked={settings.autoRefreshEnabled}
              onChange={(checked) => updateSettings({ autoRefreshEnabled: checked })}
            />
            <SelectSetting
              label="Refresh interval"
              value={String(settings.autoRefreshInterval)}
              options={[
                { value: '15', label: '15 seconds' },
                { value: '30', label: '30 seconds' },
                { value: '60', label: '1 minute' },
                { value: '300', label: '5 minutes' },
              ]}
              onChange={(value) => updateSettings({ autoRefreshInterval: Number(value) })}
              disabled={!settings.autoRefreshEnabled}
            />
          </SettingsSection>

          {/* Notifications Section */}
          <SettingsSection
            icon={<Bell size={18} />}
            title="Notifications"
            description="Get alerted about important events"
          >
            <ToggleSetting
              label="Error notifications"
              description="Show notification when a request fails"
              checked={settings.notifyOnError}
              onChange={(checked) => updateSettings({ notifyOnError: checked })}
            />
            <ToggleSetting
              label="High latency warnings"
              description="Alert when response time exceeds threshold"
              checked={settings.notifyOnHighLatency}
              onChange={(checked) => updateSettings({ notifyOnHighLatency: checked })}
            />
            <NumberSetting
              label="Latency threshold"
              value={settings.highLatencyThreshold}
              onChange={(value) => updateSettings({ highLatencyThreshold: value })}
              suffix="ms"
              min={1000}
              max={30000}
              step={1000}
              disabled={!settings.notifyOnHighLatency}
            />
          </SettingsSection>

          {/* Data Management Section */}
          <SettingsSection
            icon={<Database size={18} />}
            title="Data Management"
            description="Control request data storage and retention"
          >
            <SelectSetting
              label="Keep request logs for"
              value={String(settings.dataRetentionDays)}
              options={[
                { value: '7', label: '7 days' },
                { value: '14', label: '14 days' },
                { value: '30', label: '30 days' },
                { value: '90', label: '90 days' },
                { value: '0', label: 'Forever' },
              ]}
              onChange={(value) => updateSettings({ dataRetentionDays: Number(value) })}
            />
            <div className="text-xs text-[var(--color-text-muted)] mt-2">
              Note: Data retention is applied server-side. Changes take effect on next cleanup cycle.
            </div>

            {/* Clear All Requests Button */}
            <div className="pt-4 border-t border-[var(--color-border)]">
              <button
                onClick={() => setShowDeleteModal(true)}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-500/10 text-red-500 hover:bg-red-500/20 transition-colors font-medium"
              >
                <Trash2 size={16} />
                Clear All Requests
              </button>
              <p className="text-xs text-[var(--color-text-muted)] mt-2">
                Permanently delete all {requestCount.toLocaleString()} request{requestCount === 1 ? '' : 's'} from the database.
                This cannot be undone.
              </p>
            </div>
          </SettingsSection>
        </div>
      </PageContent>

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleClearAll}
        requestCount={requestCount}
      />
    </>
  )
}
