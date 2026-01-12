import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  Activity,
  BarChart3,
  ChevronLeft,
  ChevronRight,
  GitBranch,
  Home,
  MessageSquare,
  Settings,
  Zap,
} from 'lucide-react'

interface NavItem {
  id: string
  label: string
  icon: React.ReactNode
  badge?: number
  href?: string
  external?: boolean
}

interface NavSection {
  title: string
  items: NavItem[]
}

const navSections: NavSection[] = [
  {
    title: 'Overview',
    items: [
      { id: 'dashboard', label: 'Dashboard', icon: <Home size={18} /> },
      { id: 'cc-viz', label: 'CC-VIZ', icon: <MessageSquare size={18} />, href: 'http://localhost:8174', external: true },
      { id: 'requests', label: 'Requests', icon: <Activity size={18} /> },
    ],
  },
  {
    title: 'Analytics',
    items: [
      { id: 'usage', label: 'Token Usage', icon: <BarChart3 size={18} /> },
      { id: 'performance', label: 'Performance', icon: <Zap size={18} /> },
      { id: 'routing', label: 'Provider Routing', icon: <GitBranch size={18} /> },
    ],
  },
  {
    title: 'Configuration',
    items: [
      { id: 'settings', label: 'Settings', icon: <Settings size={18} /> },
    ],
  },
]

interface SidebarProps {
  activeItem: string
  onItemSelect: (id: string) => void
}

export function Sidebar({ activeItem, onItemSelect }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <aside
      className={cn(
        'flex flex-col h-full border-r transition-all duration-200',
        'bg-[var(--color-bg-secondary)] border-[var(--color-border)]',
        collapsed ? 'w-12' : 'w-60'
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between h-12 px-3 border-b border-[var(--color-border)]">
        {!collapsed && (
          <span className="text-sm font-semibold text-[var(--color-text-primary)]">
            Proxy Monitor
          </span>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          className={cn(
            'p-1.5 rounded hover:bg-[var(--color-bg-hover)]',
            'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]',
            'transition-colors',
            collapsed && 'mx-auto'
          )}
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-2">
        {navSections.map((section) => (
          <div key={section.title} className="mb-4">
            {!collapsed && (
              <div className="px-3 mb-1">
                <span className="text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {section.title}
                </span>
              </div>
            )}
            <ul>
              {section.items.map((item) => {
                const isActive = activeItem === item.id
                const buttonClass = cn(
                  'w-full flex items-center gap-3 px-3 py-2 text-sm transition-colors',
                  'hover:bg-[var(--color-bg-hover)]',
                  isActive
                    ? 'bg-[var(--color-bg-active)] text-[var(--color-text-primary)]'
                    : 'text-[var(--color-text-secondary)]',
                  collapsed && 'justify-center px-0'
                )

                if (item.external && item.href) {
                  return (
                    <li key={item.id}>
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={buttonClass}
                        title={collapsed ? item.label : undefined}
                      >
                        <span className={cn(
                          isActive ? 'text-[var(--color-accent)]' : ''
                        )}>
                          {item.icon}
                        </span>
                        {!collapsed && <span>{item.label}</span>}
                        {!collapsed && item.badge !== undefined && (
                          <span className="ml-auto text-xs bg-[var(--color-accent)] text-white px-1.5 py-0.5 rounded">
                            {item.badge}
                          </span>
                        )}
                      </a>
                    </li>
                  )
                }

                return (
                  <li key={item.id}>
                    <button
                      onClick={() => onItemSelect(item.id)}
                      className={buttonClass}
                      title={collapsed ? item.label : undefined}
                    >
                      <span className={cn(
                        isActive ? 'text-[var(--color-accent)]' : ''
                      )}>
                        {item.icon}
                      </span>
                      {!collapsed && <span>{item.label}</span>}
                      {!collapsed && item.badge !== undefined && (
                        <span className="ml-auto text-xs bg-[var(--color-accent)] text-white px-1.5 py-0.5 rounded">
                          {item.badge}
                        </span>
                      )}
                    </button>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}
      </nav>

      {/* Footer */}
      {!collapsed && (
        <div className="px-3 py-2 border-t border-[var(--color-border)]">
          <div className="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
            <div className="w-2 h-2 rounded-full bg-[var(--color-success)]" />
            <span>Proxy running</span>
          </div>
        </div>
      )}
    </aside>
  )
}
