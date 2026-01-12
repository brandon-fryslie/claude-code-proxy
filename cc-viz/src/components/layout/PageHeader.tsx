interface PageHeaderProps {
  title: string
  description?: string
  actions?: React.ReactNode
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
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
