import { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface ResizablePanelProps {
  children: React.ReactNode
  defaultWidth?: number
  minWidth?: number
  maxWidth?: number
  direction?: 'left' | 'right'
  className?: string
}

export function ResizablePanel({
  children,
  defaultWidth = 400,
  minWidth = 200,
  maxWidth = 800,
  direction = 'right',
  className,
}: ResizablePanelProps) {
  const [width, setWidth] = useState(defaultWidth)
  const [isResizing, setIsResizing] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  const startResizing = useCallback(() => {
    setIsResizing(true)
  }, [])

  const stopResizing = useCallback(() => {
    setIsResizing(false)
  }, [])

  const resize = useCallback(
    (e: MouseEvent) => {
      if (!isResizing || !panelRef.current) return

      const rect = panelRef.current.getBoundingClientRect()
      let newWidth: number

      if (direction === 'right') {
        newWidth = e.clientX - rect.left
      } else {
        newWidth = rect.right - e.clientX
      }

      newWidth = Math.max(minWidth, Math.min(maxWidth, newWidth))
      setWidth(newWidth)
    },
    [isResizing, direction, minWidth, maxWidth]
  )

  useEffect(() => {
    if (isResizing) {
      window.addEventListener('mousemove', resize)
      window.addEventListener('mouseup', stopResizing)
    }

    return () => {
      window.removeEventListener('mousemove', resize)
      window.removeEventListener('mouseup', stopResizing)
    }
  }, [isResizing, resize, stopResizing])

  return (
    <div
      ref={panelRef}
      className={cn('relative flex-shrink-0', className)}
      style={{ width }}
    >
      {children}
      <div
        className={cn(
          'absolute top-0 bottom-0 w-1 cursor-col-resize',
          'hover:bg-[var(--color-accent)] transition-colors',
          isResizing && 'bg-[var(--color-accent)]',
          direction === 'right' ? 'right-0' : 'left-0'
        )}
        onMouseDown={startResizing}
      />
    </div>
  )
}

interface PanelGroupProps {
  children: React.ReactNode
  direction?: 'horizontal' | 'vertical'
  className?: string
}

export function PanelGroup({ children, direction = 'horizontal', className }: PanelGroupProps) {
  return (
    <div
      className={cn(
        'flex h-full',
        direction === 'vertical' && 'flex-col',
        className
      )}
    >
      {children}
    </div>
  )
}

interface PanelProps {
  children: React.ReactNode
  className?: string
}

export function Panel({ children, className }: PanelProps) {
  return (
    <div className={cn('flex-1 min-w-0 min-h-0 overflow-auto', className)}>
      {children}
    </div>
  )
}
