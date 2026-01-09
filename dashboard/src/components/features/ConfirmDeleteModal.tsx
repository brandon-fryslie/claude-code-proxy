import { type FC, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ConfirmDeleteModalProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => Promise<void>
  requestCount?: number
}

export const ConfirmDeleteModal: FC<ConfirmDeleteModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  requestCount,
}) => {
  const [confirmText, setConfirmText] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)

  const handleConfirm = async () => {
    if (confirmText !== 'DELETE') return
    setIsDeleting(true)
    try {
      await onConfirm()
      setConfirmText('') // Reset for next time
      onClose()
    } catch (error) {
      console.error('Failed to clear data:', error)
    } finally {
      setIsDeleting(false)
    }
  }

  const handleClose = () => {
    if (!isDeleting) {
      setConfirmText('') // Reset on close
      onClose()
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="relative bg-[var(--color-bg-primary)] rounded-lg shadow-xl border border-[var(--color-border)] p-6 max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 rounded-full bg-red-500/10">
            <AlertTriangle size={24} className="text-red-500" />
          </div>
          <h2 className="text-xl font-bold text-[var(--color-text-primary)]">
            Clear All Requests?
          </h2>
        </div>

        {/* Warning message */}
        <div className="mb-4 text-[var(--color-text-secondary)]">
          <p className="mb-2">
            This will permanently delete{' '}
            {requestCount !== undefined
              ? `${requestCount.toLocaleString()} request${requestCount === 1 ? '' : 's'}`
              : 'all requests'}{' '}
            from the database.
          </p>
          <p className="font-semibold text-red-500">
            This action cannot be undone.
          </p>
        </div>

        {/* Confirmation input */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-[var(--color-text-primary)] mb-2">
            Type <span className="font-mono font-bold">DELETE</span> to confirm:
          </label>
          <input
            type="text"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            placeholder="DELETE"
            disabled={isDeleting}
            className={cn(
              'w-full px-3 py-2 rounded-lg border',
              'bg-[var(--color-bg-secondary)] border-[var(--color-border)]',
              'text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)]',
              'focus:outline-none focus:border-[var(--color-accent)]',
              'disabled:opacity-50 disabled:cursor-not-allowed'
            )}
            autoFocus
          />
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={handleClose}
            disabled={isDeleting}
            className={cn(
              'flex-1 px-4 py-2 rounded-lg font-medium transition-colors',
              'bg-[var(--color-bg-tertiary)] border border-[var(--color-border)]',
              'text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]',
              'disabled:opacity-50 disabled:cursor-not-allowed'
            )}
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={confirmText !== 'DELETE' || isDeleting}
            className={cn(
              'flex-1 px-4 py-2 rounded-lg font-medium transition-colors',
              'bg-red-600 text-white hover:bg-red-700',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'flex items-center justify-center gap-2'
            )}
          >
            {isDeleting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Deleting...
              </>
            ) : (
              'Confirm Delete'
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
