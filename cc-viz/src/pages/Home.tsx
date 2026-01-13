import { AppLayout } from '@/components/layout'
import { Link } from '@/components/ui/Link'
import {
  Settings,
  Bot,
  Puzzle,
  FolderKanban,
  FileText,
  History,
  Webhook,
  MonitorPlay,
  BarChart3,
  MessageSquare,
} from 'lucide-react'

interface CategoryCard {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  items: string[]
  status: 'available' | 'coming-soon'
  href?: string
}

const categories: CategoryCard[] = [
  // --- AVAILABLE NOW ---
  {
    id: 'conversations',
    title: 'Conversations',
    description: 'Browse and search through your Claude Code conversation logs',
    icon: <MessageSquare size={24} />,
    items: ['Session transcripts', 'Message threads', 'Tool usage', 'Subagent calls'],
    status: 'available',
    href: '/cc-viz/conversations',
  },
  {
    id: 'configuration',
    title: 'Configuration',
    description: 'View and manage your Claude Code settings and global configuration',
    icon: <Settings size={24} />,
    items: ['settings.json', 'CLAUDE.md', '.mcp.json', 'MCP servers'],
    status: 'available',
    href: '/cc-viz/configuration',
  },
  {
    id: 'projects',
    title: 'Projects',
    description: 'Claude Code activity and conversation history per project',
    icon: <FolderKanban size={24} />,
    items: ['Session history', 'Storage breakdown', 'Activity timeline', 'Agent usage'],
    status: 'available',
    href: '/cc-viz/projects',
  },
  // --- COMING SOON ---
  {
    id: 'agents',
    title: 'Agents',
    description: 'Custom subagent definitions from ~/.claude/agents/',
    icon: <Bot size={24} />,
    items: ['code-bloodhound.md', 'code-monkey-jr.md', 'codex-plan.md', 'work-decomposer.md'],
    status: 'coming-soon',
  },
  {
    id: 'commands',
    title: 'Commands',
    description: 'Slash commands from ~/.claude/commands/',
    icon: <Bot size={24} />,
    items: ['add-command.md', 'setup-mcp-docs.md', 'system-prompt-editor.md'],
    status: 'coming-soon',
  },
  {
    id: 'skills',
    title: 'Skills',
    description: 'Custom skills from ~/.claude/skills/',
    icon: <Bot size={24} />,
    items: ['Custom skill definitions', 'Skill configurations'],
    status: 'coming-soon',
  },
  {
    id: 'plugins',
    title: 'Plugins',
    description: 'Installed plugins, marketplaces, and plugin cache',
    icon: <Puzzle size={24} />,
    items: ['installed_plugins.json', 'Plugin cache', 'Marketplaces', 'plugins-config/'],
    status: 'coming-soon',
  },
  {
    id: 'session-data',
    title: 'Session Data',
    description: 'Debug logs, todos, plans, and session environment data',
    icon: <FileText size={24} />,
    items: ['debug/ logs', 'todos/ states', 'plans/', 'session-env/'],
    status: 'available',
    href: '/cc-viz/session-data',
  },
  {
    id: 'history',
    title: 'History',
    description: 'Conversation history and file change history',
    icon: <History size={24} />,
    items: ['history.jsonl', 'file-history/', 'paste-cache/', 'shell-snapshots/'],
    status: 'coming-soon',
  },
  {
    id: 'hooks',
    title: 'Hooks',
    description: 'Automation hooks for Claude Code events',
    icon: <Webhook size={24} />,
    items: ['Pre/Post tool hooks', 'Session hooks', 'Custom automations'],
    status: 'coming-soon',
  },
  {
    id: 'ide',
    title: 'IDE Integration',
    description: 'IDE-specific settings from ~/.claude/ide/',
    icon: <MonitorPlay size={24} />,
    items: ['VS Code settings', 'JetBrains settings', 'Editor configs'],
    status: 'coming-soon',
  },
  {
    id: 'telemetry',
    title: 'Telemetry & Stats',
    description: 'Usage statistics and telemetry data',
    icon: <BarChart3 size={24} />,
    items: ['stats-cache.json', 'telemetry/', 'statsig/', 'Feature flags'],
    status: 'coming-soon',
  },
]

function CategoryCardComponent({ category }: { category: CategoryCard }) {
  const isAvailable = category.status === 'available'

  const content = (
    <div
      className={`
        group relative p-5 rounded-lg border transition-all duration-200
        ${isAvailable
          ? 'border-[var(--color-border)] hover:border-[var(--color-accent)] hover:shadow-lg cursor-pointer bg-[var(--color-bg-secondary)]'
          : 'border-dashed border-[var(--color-border)] bg-[var(--color-bg-secondary)]/50 opacity-75'
        }
      `}
    >
      {/* Status Badge */}
      {!isAvailable && (
        <span className="absolute top-3 right-3 px-2 py-0.5 text-xs font-medium rounded-full bg-[var(--color-bg-hover)] text-[var(--color-text-muted)]">
          Coming Soon
        </span>
      )}

      {/* Icon and Title */}
      <div className="flex items-start gap-4 mb-3">
        <div className={`
          p-2.5 rounded-lg
          ${isAvailable
            ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] group-hover:bg-[var(--color-accent)]/20'
            : 'bg-[var(--color-bg-hover)] text-[var(--color-text-muted)]'
          }
        `}>
          {category.icon}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className={`
            font-semibold text-base mb-1
            ${isAvailable ? 'text-[var(--color-text-primary)]' : 'text-[var(--color-text-secondary)]'}
          `}>
            {category.title}
          </h3>
          <p className="text-sm text-[var(--color-text-muted)] line-clamp-2">
            {category.description}
          </p>
        </div>
      </div>

      {/* Items List */}
      <div className="flex flex-wrap gap-1.5 mt-3">
        {category.items.slice(0, 4).map((item) => (
          <span
            key={item}
            className={`
              px-2 py-0.5 text-xs rounded
              ${isAvailable
                ? 'bg-[var(--color-bg-hover)] text-[var(--color-text-secondary)]'
                : 'bg-[var(--color-bg-primary)] text-[var(--color-text-muted)]'
              }
            `}
          >
            {item}
          </span>
        ))}
      </div>
    </div>
  )

  if (isAvailable && category.href) {
    return (
      <Link href={category.href} className="block">
        {content}
      </Link>
    )
  }

  return content
}

export function HomePage() {
  return (
    <AppLayout
      title="CC-VIZ"
      description="Visualize your ~/.claude directory"
    >
      <div className="max-w-6xl mx-auto p-6">
        {/* Hero Section */}
        <div className="text-center mb-10">
          <h1 className="text-3xl font-bold text-[var(--color-text-primary)] mb-3">
            Claude Code Directory Visualizer
          </h1>
          <p className="text-lg text-[var(--color-text-secondary)] max-w-2xl mx-auto">
            Explore, search, and understand everything in your <code className="px-1.5 py-0.5 bg-[var(--color-bg-secondary)] rounded text-sm font-mono">~/.claude</code> directory
          </p>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-10">
          <QuickStat label="Visualizers" value={categories.length.toString()} />
          <QuickStat label="Available Now" value={categories.filter(c => c.status === 'available').length.toString()} />
          <QuickStat label="Coming Soon" value={categories.filter(c => c.status === 'coming-soon').length.toString()} />
          <QuickStat label="Data Sources" value="15+" />
        </div>

        {/* Categories Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {categories.map((category) => (
            <CategoryCardComponent key={category.id} category={category} />
          ))}
        </div>

        {/* Footer Note */}
        <div className="mt-10 text-center text-sm text-[var(--color-text-muted)]">
          <p>
            CC-VIZ is part of the Claude Code Proxy project.{' '}
            <a
              href="/dashboard/"
              className="text-[var(--color-accent)] hover:underline"
            >
              Back to Dashboard
            </a>
          </p>
        </div>
      </div>
    </AppLayout>
  )
}

function QuickStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-4 rounded-lg bg-[var(--color-bg-secondary)] border border-[var(--color-border)]">
      <div className="text-2xl font-bold text-[var(--color-text-primary)]">{value}</div>
      <div className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">{label}</div>
    </div>
  )
}
