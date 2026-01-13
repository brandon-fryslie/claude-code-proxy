import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AppLayout } from '@/components/layout'
import { Search, Check, X, Server, FileText, Settings as SettingsIcon } from 'lucide-react'

interface PermissionGroups {
  bash: string[]
  tools: string[]
  mcp: string[]
  other: string[]
}

interface PluginStatus {
  enabled: string[]
  disabled: string[]
}

interface ClaudeMdSection {
  name: string
  position: number
}

interface MCPServer {
  name: string
  command: string
  type: string
  args?: string[]
}

interface ClaudeConfig {
  settings: {
    model: string
    default_mode: string
    permissions: PermissionGroups
    plugins: PluginStatus
    raw: Record<string, unknown>
  } | null
  settings_error?: string
  claude_md: {
    content: string
    sections: ClaudeMdSection[]
  } | null
  claude_md_error?: string
  mcp_config: {
    servers: MCPServer[]
    raw: Record<string, unknown>
  } | null
  mcp_config_error?: string
}

type TabId = 'settings' | 'claude_md' | 'mcp'

const tabs: { id: TabId; label: string; icon: React.ReactNode }[] = [
  { id: 'settings', label: 'Settings', icon: <SettingsIcon size={16} /> },
  { id: 'claude_md', label: 'CLAUDE.md', icon: <FileText size={16} /> },
  { id: 'mcp', label: 'MCP Servers', icon: <Server size={16} /> },
]

export function ConfigurationPage() {
  const [activeTab, setActiveTab] = useState<TabId>('settings')
  const [permissionSearch, setPermissionSearch] = useState('')
  const [showAllDisabled, setShowAllDisabled] = useState(false)

  const { data, isLoading, error } = useQuery<ClaudeConfig>({
    queryKey: ['claude-config'],
    queryFn: () => fetch('/api/v2/claude/config').then(r => r.json()),
  })

  if (isLoading) {
    return (
      <AppLayout title="Configuration" activeItem="configuration">
        <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
          Loading configuration...
        </div>
      </AppLayout>
    )
  }

  if (error) {
    return (
      <AppLayout title="Configuration" activeItem="configuration">
        <div className="flex items-center justify-center h-full text-red-500">
          Error loading configuration: {(error as Error).message}
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout
      title="Configuration"
      description="Your global Claude Code settings"
      activeItem="configuration"
    >
      <div className="flex flex-col h-full">
        {/* Tab Navigation */}
        <div className="flex border-b border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors
                border-b-2 -mb-[2px]
                ${activeTab === tab.id
                  ? 'border-[var(--color-accent)] text-[var(--color-text-primary)]'
                  : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
                }
              `}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-auto p-6">
          {activeTab === 'settings' && (
            <SettingsTab
              config={data}
              permissionSearch={permissionSearch}
              setPermissionSearch={setPermissionSearch}
              showAllDisabled={showAllDisabled}
              setShowAllDisabled={setShowAllDisabled}
            />
          )}
          {activeTab === 'claude_md' && <ClaudeMdTab config={data} />}
          {activeTab === 'mcp' && <MCPServersTab config={data} />}
        </div>
      </div>
    </AppLayout>
  )
}

function SettingsTab({
  config,
  permissionSearch,
  setPermissionSearch,
  showAllDisabled,
  setShowAllDisabled,
}: {
  config: ClaudeConfig | undefined
  permissionSearch: string
  setPermissionSearch: (s: string) => void
  showAllDisabled: boolean
  setShowAllDisabled: (b: boolean) => void
}) {
  const settings = config?.settings

  if (!settings) {
    return (
      <div className="text-center py-8 text-[var(--color-text-muted)]">
        <p>{config?.settings_error || 'Settings not available'}</p>
      </div>
    )
  }

  // Filter permissions based on search
  const filteredPermissions = useMemo(() => {
    if (!permissionSearch) return settings.permissions
    const query = permissionSearch.toLowerCase()
    return {
      bash: settings.permissions.bash.filter(p => p.toLowerCase().includes(query)),
      tools: settings.permissions.tools.filter(p => p.toLowerCase().includes(query)),
      mcp: settings.permissions.mcp.filter(p => p.toLowerCase().includes(query)),
      other: settings.permissions.other.filter(p => p.toLowerCase().includes(query)),
    }
  }, [settings.permissions, permissionSearch])

  const totalPermissions =
    settings.permissions.bash.length +
    settings.permissions.tools.length +
    settings.permissions.mcp.length +
    settings.permissions.other.length

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Quick Info Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <InfoCard label="Model" value={settings.model || 'Not set'} />
        <InfoCard label="Default Mode" value={settings.default_mode || 'Not set'} />
        <InfoCard label="Permissions" value={`${totalPermissions} allowed`} />
        <InfoCard
          label="Plugins"
          value={`${settings.plugins.enabled.length} enabled`}
        />
      </div>

      {/* Permissions Section */}
      <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold text-[var(--color-text-primary)]">
            Permissions ({totalPermissions})
          </h3>
          <div className="relative">
            <Search
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]"
            />
            <input
              type="text"
              placeholder="Search permissions..."
              value={permissionSearch}
              onChange={(e) => setPermissionSearch(e.target.value)}
              className="pl-9 pr-3 py-1.5 text-sm bg-[var(--color-bg-primary)] border border-[var(--color-border)] rounded focus:outline-none focus:border-[var(--color-accent)]"
            />
          </div>
        </div>

        <div className="space-y-4">
          {filteredPermissions.bash.length > 0 && (
            <PermissionGroup title="Bash Commands" permissions={filteredPermissions.bash} />
          )}
          {filteredPermissions.tools.length > 0 && (
            <PermissionGroup title="Tools" permissions={filteredPermissions.tools} />
          )}
          {filteredPermissions.mcp.length > 0 && (
            <PermissionGroup title="MCP" permissions={filteredPermissions.mcp} />
          )}
          {filteredPermissions.other.length > 0 && (
            <PermissionGroup title="Other" permissions={filteredPermissions.other} />
          )}
          {Object.values(filteredPermissions).every(arr => arr.length === 0) && (
            <p className="text-center text-[var(--color-text-muted)] py-4">
              No permissions match your search
            </p>
          )}
        </div>
      </div>

      {/* Plugins Section */}
      <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4">
        <h3 className="font-semibold text-[var(--color-text-primary)] mb-4">
          Plugins
        </h3>

        {settings.plugins.enabled.length > 0 && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Enabled ({settings.plugins.enabled.length})
            </h4>
            <div className="flex flex-wrap gap-2">
              {settings.plugins.enabled.map((plugin) => (
                <span
                  key={plugin}
                  className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-green-500/10 text-green-400 rounded"
                >
                  <Check size={12} />
                  {plugin}
                </span>
              ))}
            </div>
          </div>
        )}

        {settings.plugins.disabled.length > 0 && (
          <div>
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-medium text-[var(--color-text-secondary)]">
                Disabled ({settings.plugins.disabled.length})
              </h4>
              <button
                onClick={() => setShowAllDisabled(!showAllDisabled)}
                className="text-xs text-[var(--color-accent)] hover:underline"
              >
                {showAllDisabled ? 'Show less' : 'Show all'}
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {(showAllDisabled
                ? settings.plugins.disabled
                : settings.plugins.disabled.slice(0, 5)
              ).map((plugin) => (
                <span
                  key={plugin}
                  className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-[var(--color-bg-hover)] text-[var(--color-text-muted)] rounded"
                >
                  <X size={12} />
                  {plugin}
                </span>
              ))}
              {!showAllDisabled && settings.plugins.disabled.length > 5 && (
                <span className="px-2 py-1 text-xs text-[var(--color-text-muted)]">
                  +{settings.plugins.disabled.length - 5} more
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function ClaudeMdTab({ config }: { config: ClaudeConfig | undefined }) {
  const claudeMd = config?.claude_md
  const [activeSection, setActiveSection] = useState<string | null>(null)

  if (!claudeMd) {
    return (
      <div className="text-center py-8 text-[var(--color-text-muted)]">
        <p>{config?.claude_md_error || 'CLAUDE.md not found'}</p>
        <p className="text-sm mt-2">
          Create a ~/.claude/CLAUDE.md file to add global instructions
        </p>
      </div>
    )
  }

  return (
    <div className="flex gap-6 h-full">
      {/* Section Navigation */}
      {claudeMd.sections.length > 0 && (
        <div className="w-48 flex-shrink-0">
          <h4 className="text-sm font-medium text-[var(--color-text-secondary)] mb-2">
            Sections
          </h4>
          <nav className="space-y-1">
            {claudeMd.sections.map((section) => (
              <button
                key={section.name}
                onClick={() => setActiveSection(section.name)}
                className={`
                  w-full text-left px-3 py-2 text-sm rounded transition-colors
                  ${activeSection === section.name
                    ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]'
                    : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)]'
                  }
                `}
              >
                {section.name}
              </button>
            ))}
          </nav>
        </div>
      )}

      {/* Content */}
      <div className="flex-1 bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4 overflow-auto">
        <pre className="text-sm text-[var(--color-text-primary)] whitespace-pre-wrap font-mono">
          {claudeMd.content}
        </pre>
      </div>
    </div>
  )
}

function MCPServersTab({ config }: { config: ClaudeConfig | undefined }) {
  const mcpConfig = config?.mcp_config

  if (!mcpConfig || !mcpConfig.servers || mcpConfig.servers.length === 0) {
    return (
      <div className="text-center py-8 text-[var(--color-text-muted)]">
        <Server size={48} className="mx-auto mb-4 opacity-50" />
        <p>{config?.mcp_config_error || 'No MCP servers configured'}</p>
        <p className="text-sm mt-2">
          Add servers to ~/.claude/.mcp.json to enable MCP integrations
        </p>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto space-y-4">
      <p className="text-sm text-[var(--color-text-muted)] mb-4">
        {mcpConfig.servers.length} MCP server{mcpConfig.servers.length !== 1 ? 's' : ''} configured
      </p>

      {mcpConfig.servers.map((server) => (
        <div
          key={server.name}
          className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4"
        >
          <div className="flex items-start justify-between mb-3">
            <h3 className="font-semibold text-[var(--color-text-primary)]">
              {server.name}
            </h3>
            <span className="px-2 py-0.5 text-xs bg-[var(--color-bg-hover)] text-[var(--color-text-muted)] rounded">
              {server.type}
            </span>
          </div>

          <div className="space-y-2 text-sm">
            <div>
              <span className="text-[var(--color-text-muted)]">Command: </span>
              <code className="text-[var(--color-text-secondary)] bg-[var(--color-bg-primary)] px-1 rounded">
                {server.command}
              </code>
            </div>
            {server.args && server.args.length > 0 && (
              <div>
                <span className="text-[var(--color-text-muted)]">Args: </span>
                <code className="text-[var(--color-text-secondary)] bg-[var(--color-bg-primary)] px-1 rounded">
                  {server.args.join(' ')}
                </code>
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4">
      <div className="text-sm text-[var(--color-text-muted)]">{label}</div>
      <div className="text-lg font-semibold text-[var(--color-text-primary)] mt-1">
        {value}
      </div>
    </div>
  )
}

function PermissionGroup({
  title,
  permissions,
}: {
  title: string
  permissions: string[]
}) {
  return (
    <div>
      <h4 className="text-sm font-medium text-[var(--color-text-secondary)] mb-2">
        {title} ({permissions.length})
      </h4>
      <div className="flex flex-wrap gap-1.5">
        {permissions.map((perm) => (
          <span
            key={perm}
            className="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-[var(--color-bg-hover)] text-[var(--color-text-secondary)] rounded"
          >
            <Check size={10} className="text-green-400" />
            {perm}
          </span>
        ))}
      </div>
    </div>
  )
}
