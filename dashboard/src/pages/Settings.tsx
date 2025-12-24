import { PageHeader, PageContent } from '@/components/layout'
import { Bell, Database, Key } from 'lucide-react'

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
      {children}
    </div>
  )
}

export function SettingsPage() {
  return (
    <>
      <PageHeader
        title="Settings"
        description="Configure proxy and dashboard settings"
      />
      <PageContent>
        <div className="max-w-2xl space-y-4">
          <SettingsSection
            title="Notifications"
            description="Configure push notifications for requests"
            icon={<Bell size={18} />}
          >
            <div className="space-y-3">
              <label className="flex items-center justify-between">
                <span className="text-sm text-[var(--color-text-secondary)]">
                  Notify on errors
                </span>
                <input
                  type="checkbox"
                  defaultChecked
                  className="rounded border-[var(--color-border)]"
                />
              </label>
              <label className="flex items-center justify-between">
                <span className="text-sm text-[var(--color-text-secondary)]">
                  Notify on high latency (&gt;5s)
                </span>
                <input
                  type="checkbox"
                  className="rounded border-[var(--color-border)]"
                />
              </label>
            </div>
          </SettingsSection>

          <SettingsSection
            title="Data Retention"
            description="Manage request log storage"
            icon={<Database size={18} />}
          >
            <div className="space-y-3">
              <div>
                <label className="text-sm text-[var(--color-text-secondary)] block mb-1">
                  Keep logs for
                </label>
                <select className="w-full p-2 text-sm bg-[var(--color-bg-tertiary)] border border-[var(--color-border)] rounded text-[var(--color-text-primary)]">
                  <option>7 days</option>
                  <option>30 days</option>
                  <option>90 days</option>
                  <option>Forever</option>
                </select>
              </div>
            </div>
          </SettingsSection>

          <SettingsSection
            title="API Keys"
            description="Manage provider API keys"
            icon={<Key size={18} />}
          >
            <p className="text-sm text-[var(--color-text-muted)]">
              API keys are configured in config.yaml.
              Future: Edit keys directly from this UI.
            </p>
          </SettingsSection>
        </div>
      </PageContent>
    </>
  )
}
