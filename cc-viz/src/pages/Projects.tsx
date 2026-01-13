import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AppLayout } from '@/components/layout'
import {
  Search,
  ArrowUpDown,
  FolderOpen,
  Clock,
  HardDrive,
  FileText,
  ExternalLink,
} from 'lucide-react'

interface ProjectSession {
  id: string
  file: string
  size: number
  modified: string
  is_agent: boolean
}

interface Project {
  id: string
  path: string
  name: string
  file_count: number
  total_size: number
  session_count: number
  agent_count: number
  last_modified: string
  created: string
}

interface ProjectsResponse {
  projects: Project[]
  total_count: number
  total_size: number
}

interface ProjectDetail {
  id: string
  path: string
  name: string
  file_count: number
  total_size: number
  session_count: number
  agent_count: number
  last_modified: string
  sessions: ProjectSession[]
  size_breakdown: {
    sessions: number
    agents: number
  }
}

type SortOption = 'recent' | 'size' | 'name' | 'sessions'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

export function ProjectsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<SortOption>('recent')
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null)

  const { data, isLoading, error } = useQuery<ProjectsResponse>({
    queryKey: ['claude-projects'],
    queryFn: () => fetch('/api/v2/claude/projects').then(r => r.json()),
  })

  const { data: projectDetail, isLoading: isLoadingDetail } = useQuery<ProjectDetail>({
    queryKey: ['claude-project', selectedProjectId],
    queryFn: () => fetch(`/api/v2/claude/projects/${selectedProjectId}`).then(r => r.json()),
    enabled: !!selectedProjectId,
  })

  // Filter and sort projects
  const filteredProjects = useMemo(() => {
    if (!data?.projects) return []

    let filtered = data.projects
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.path.toLowerCase().includes(query)
      )
    }

    return [...filtered].sort((a, b) => {
      switch (sortBy) {
        case 'recent':
          return new Date(b.last_modified).getTime() - new Date(a.last_modified).getTime()
        case 'size':
          return b.total_size - a.total_size
        case 'name':
          return a.name.localeCompare(b.name)
        case 'sessions':
          return b.session_count - a.session_count
        default:
          return 0
      }
    })
  }, [data?.projects, searchQuery, sortBy])

  const cycleSortOption = () => {
    const options: SortOption[] = ['recent', 'size', 'name', 'sessions']
    const currentIndex = options.indexOf(sortBy)
    setSortBy(options[(currentIndex + 1) % options.length])
  }

  const getSortLabel = () => {
    switch (sortBy) {
      case 'recent': return 'Recent'
      case 'size': return 'Size'
      case 'name': return 'Name'
      case 'sessions': return 'Sessions'
    }
  }

  if (isLoading) {
    return (
      <AppLayout title="Projects" activeItem="projects">
        <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
          Loading projects...
        </div>
      </AppLayout>
    )
  }

  if (error) {
    return (
      <AppLayout title="Projects" activeItem="projects">
        <div className="flex items-center justify-center h-full text-red-500">
          Error loading projects: {(error as Error).message}
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout
      title="Projects"
      description="Claude Code activity across your projects"
      activeItem="projects"
    >
      <div className="flex flex-col h-full">
        {/* Summary Bar */}
        <div className="p-4 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)]">
          <div className="flex items-center gap-6 text-sm">
            <div className="flex items-center gap-2">
              <FolderOpen size={16} className="text-[var(--color-text-muted)]" />
              <span className="font-semibold text-[var(--color-text-primary)]">
                {data?.total_count || 0}
              </span>
              <span className="text-[var(--color-text-muted)]">Projects</span>
            </div>
            <div className="flex items-center gap-2">
              <HardDrive size={16} className="text-[var(--color-text-muted)]" />
              <span className="font-semibold text-[var(--color-text-primary)]">
                {formatBytes(data?.total_size || 0)}
              </span>
              <span className="text-[var(--color-text-muted)]">Total</span>
            </div>
            {data?.projects?.[0] && (
              <div className="flex items-center gap-2">
                <Clock size={16} className="text-[var(--color-text-muted)]" />
                <span className="text-[var(--color-text-muted)]">Last active:</span>
                <span className="text-[var(--color-text-secondary)]">
                  {formatRelativeTime(data.projects[0].last_modified)}
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="flex-1 flex overflow-hidden">
          {/* Project List */}
          <div className="w-80 border-r border-[var(--color-border)] bg-[var(--color-bg-secondary)] flex flex-col">
            {/* Search and Sort */}
            <div className="p-3 border-b border-[var(--color-border)] space-y-2">
              <div className="relative">
                <Search
                  size={16}
                  className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]"
                />
                <input
                  type="text"
                  placeholder="Search projects..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-9 pr-3 py-2 text-sm bg-[var(--color-bg-primary)] border border-[var(--color-border)] rounded focus:outline-none focus:border-[var(--color-accent)]"
                />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs text-[var(--color-text-muted)]">
                  {filteredProjects.length} project{filteredProjects.length !== 1 ? 's' : ''}
                </span>
                <button
                  onClick={cycleSortOption}
                  className="flex items-center gap-1 px-2 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)] rounded transition-colors"
                >
                  <ArrowUpDown size={12} />
                  Sort: {getSortLabel()}
                </button>
              </div>
            </div>

            {/* Project List */}
            <div className="flex-1 overflow-y-auto">
              {filteredProjects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => setSelectedProjectId(project.id)}
                  className={`
                    w-full text-left p-3 border-b border-[var(--color-border)] transition-colors
                    ${selectedProjectId === project.id
                      ? 'bg-[var(--color-bg-active)]'
                      : 'hover:bg-[var(--color-bg-hover)]'
                    }
                  `}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-medium text-[var(--color-text-primary)] truncate">
                      {project.name}
                    </span>
                    <span className="text-xs text-[var(--color-text-muted)]">
                      {formatRelativeTime(project.last_modified)}
                    </span>
                  </div>
                  <div className="text-xs text-[var(--color-text-muted)]">
                    {project.file_count} files, {formatBytes(project.total_size)}
                  </div>
                </button>
              ))}
              {filteredProjects.length === 0 && (
                <div className="p-4 text-center text-[var(--color-text-muted)]">
                  No projects match your search
                </div>
              )}
            </div>
          </div>

          {/* Project Detail */}
          <div className="flex-1 bg-[var(--color-bg-primary)] overflow-auto">
            {selectedProjectId ? (
              isLoadingDetail ? (
                <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
                  Loading project details...
                </div>
              ) : projectDetail ? (
                <ProjectDetailPane project={projectDetail} />
              ) : (
                <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
                  Project not found
                </div>
              )
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-[var(--color-text-muted)]">
                <FolderOpen size={48} className="mb-4 opacity-50" />
                <p>Select a project to view details</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  )
}

function ProjectDetailPane({ project }: { project: ProjectDetail }) {
  const sessionPercentage = project.total_size > 0
    ? Math.round((project.size_breakdown.sessions / project.total_size) * 100)
    : 0

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h2 className="text-xl font-bold text-[var(--color-text-primary)] mb-1">
          {project.name}
        </h2>
        <p className="text-sm text-[var(--color-text-muted)] font-mono">
          {project.path}
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <StatCard label="Sessions" value={project.session_count.toString()} />
        <StatCard label="Agents" value={project.agent_count.toString()} />
        <StatCard label="Total Size" value={formatBytes(project.total_size)} />
        <StatCard
          label="Last Active"
          value={formatRelativeTime(project.last_modified)}
        />
      </div>

      {/* Storage Breakdown */}
      <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4 mb-6">
        <h3 className="font-semibold text-[var(--color-text-primary)] mb-3">
          Storage Breakdown
        </h3>
        <div className="mb-2">
          <div className="h-3 bg-[var(--color-bg-hover)] rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all"
              style={{ width: `${sessionPercentage}%` }}
            />
          </div>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-[var(--color-text-secondary)]">
            Sessions: {formatBytes(project.size_breakdown.sessions)} ({sessionPercentage}%)
          </span>
          <span className="text-[var(--color-text-secondary)]">
            Agents: {formatBytes(project.size_breakdown.agents)} ({100 - sessionPercentage}%)
          </span>
        </div>
      </div>

      {/* Recent Sessions */}
      <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-4 mb-6">
        <h3 className="font-semibold text-[var(--color-text-primary)] mb-3">
          Recent Sessions
        </h3>
        <div className="space-y-2">
          {project.sessions.slice(0, 10).map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between py-2 px-3 bg-[var(--color-bg-primary)] rounded"
            >
              <div className="flex items-center gap-2">
                <FileText size={14} className="text-[var(--color-text-muted)]" />
                <span className="text-sm text-[var(--color-text-secondary)] font-mono truncate max-w-[200px]">
                  {session.id}
                </span>
                {session.is_agent && (
                  <span className="px-1.5 py-0.5 text-xs bg-purple-500/10 text-purple-400 rounded">
                    agent
                  </span>
                )}
              </div>
              <div className="flex items-center gap-4 text-xs text-[var(--color-text-muted)]">
                <span>{formatBytes(session.size)}</span>
                <span>{formatRelativeTime(session.modified)}</span>
              </div>
            </div>
          ))}
          {project.sessions.length > 10 && (
            <p className="text-center text-xs text-[var(--color-text-muted)] py-2">
              +{project.sessions.length - 10} more sessions
            </p>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <a
          href={`/cc-viz/conversations?project=${encodeURIComponent(project.path)}`}
          className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--color-accent)] text-white rounded hover:opacity-90 transition-opacity text-sm"
        >
          <ExternalLink size={14} />
          Open in Conversations
        </a>
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-[var(--color-bg-secondary)] rounded-lg border border-[var(--color-border)] p-3">
      <div className="text-xs text-[var(--color-text-muted)]">{label}</div>
      <div className="text-lg font-semibold text-[var(--color-text-primary)] mt-0.5">
        {value}
      </div>
    </div>
  )
}
