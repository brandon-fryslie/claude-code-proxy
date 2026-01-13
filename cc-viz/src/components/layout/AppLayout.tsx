import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  ChevronLeft,
  ChevronRight,
  Home,
  MessageSquare,
  ExternalLink,
  Settings,
  Bot,
  Puzzle,
  FolderKanban,
  FileText,
  History,
  Webhook,
  MonitorPlay,
  BarChart3,
  LayoutDashboard,
} from 'lucide-react'

interface NavItem {
  id: string
  label: string
  icon: React.ReactNode
  badge?: number
  href?: string
  external?: boolean
  disabled?: boolean
}

interface NavSection {
  title: string
  items: NavItem[]
}

const navSections: NavSection[] = [
  {
    title: 'Overview',
    items: [
      { id: 'home', label: 'Home', icon: <LayoutDashboard size={18} />, href: '/cc-viz/' },
    ],
  },
  {
    title: 'Available',
    items: [
      { id: 'conversations', label: 'Conversations', icon: <MessageSquare size={18} />, href: '/cc-viz/conversations' },
      { id: 'configuration', label: 'Configuration', icon: <Settings size={18} />, href: '/cc-viz/configuration' },
      { id: 'projects', label: 'Projects', icon: <FolderKanban size={18} />, href: '/cc-viz/projects' },
    ],
  },
  {
    title: 'Extensibility',
    items: [
      { id: 'agents', label: 'Agents', icon: <Bot size={18} />, disabled: true },
      { id: 'commands', label: 'Commands', icon: <Bot size={18} />, disabled: true },
      { id: 'skills', label: 'Skills', icon: <Bot size={18} />, disabled: true },
      { id: 'plugins', label: 'Plugins', icon: <Puzzle size={18} />, disabled: true },
      { id: 'hooks', label: 'Hooks', icon: <Webhook size={18} />, disabled: true },
    ],
  },
  {
    title: 'Data',
    items: [
      { id: 'session-data', label: 'Session Data', icon: <FileText size={18} />, disabled: true },
      { id: 'history', label: 'History', icon: <History size={18} />, disabled: true },
      { id: 'telemetry', label: 'Telemetry & Stats', icon: <BarChart3 size={18} />, disabled: true },
    ],
  },
  {
    title: 'Integration',
    items: [
      { id: 'ide', label: 'IDE Integration', icon: <MonitorPlay size={18} />, disabled: true },
    ],
  },
  {
    title: 'Links',
    items: [
      { id: 'dashboard', label: 'Back to Dashboard', icon: <Home size={18} />, href: '/dashboard/', external: false },
    ],
  },
]

interface SidebarProps {
  activeItem: string
}

function Sidebar({ activeItem }: SidebarProps) {
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
            CC-VIZ
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

                if (item.href) {
                  return (
                    <li key={item.id}>
                      <a
                        href={item.href}
                        target={item.external ? '_blank' : undefined}
                        rel={item.external ? 'noopener noreferrer' : undefined}
                        className={buttonClass}
                        title={collapsed ? item.label : undefined}
                      >
                        <span className={cn(
                          isActive ? 'text-[var(--color-accent)]' : ''
                        )}>
                          {item.icon}
                        </span>
                        {!collapsed && (
                          <>
                            <span className="flex-1">{item.label}</span>
                            {item.external && <ExternalLink size={12} className="text-[var(--color-text-muted)]" />}
                          </>
                        )}
                      </a>
                    </li>
                  )
                }

                return (
                  <li key={item.id}>
                    <button
                      className={buttonClass}
                      title={collapsed ? item.label : undefined}
                      disabled
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
            <span>Connected</span>
          </div>
        </div>
      )}
    </aside>
  )
}

interface PageHeaderProps {
  title: string
  description?: string
  actions?: React.ReactNode
}

function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <header className="flex items-center justify-between h-12 px-4 border-b border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
      <div>
        <h1 className="text-sm font-semibold text-[var(--color-text-primary)]">{title}</h1>
        {description && (
          <p className="text-xs text-[var(--color-text-muted)]">{description}</p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  )
}

interface AppLayoutProps {
  children: React.ReactNode
  title: string
  description?: string
  actions?: React.ReactNode
  activeItem?: string
}

export function AppLayout({ children, title, description, actions, activeItem = 'home' }: AppLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-[var(--color-bg-primary)]">
      <Sidebar activeItem={activeItem} />
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <PageHeader title={title} description={description} actions={actions} />
        <div className="flex-1 overflow-auto">
          {children}
        </div>
      </main>
    </div>
  )
}
