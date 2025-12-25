import { type FC } from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCopyToClipboard } from '@/lib/hooks/useCopyToClipboard'

interface CopyButtonProps {
  content: string
  className?: string
  size?: 'sm' | 'md' | 'lg'
  variant?: 'default' | 'ghost' | 'outline'
  label?: string
}

export const CopyButton: FC<CopyButtonProps> = ({
  content,
  className,
  size = 'sm',
  variant = 'ghost',
  label,
}) => {
  const { copied, copy } = useCopyToClipboard()

  const sizeClasses = {
    sm: 'p-1.5',
    md: 'p-2',
    lg: 'p-2.5',
  }

  const iconSizes = {
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  }

  const variantClasses = {
    default: 'bg-gray-200 hover:bg-gray-300 text-gray-700',
    ghost: 'hover:bg-gray-100 text-gray-500 hover:text-gray-700',
    outline: 'border border-gray-300 hover:bg-gray-50 text-gray-600',
  }

  return (
    <button
      onClick={() => copy(content)}
      className={cn(
        'inline-flex items-center gap-1.5 rounded transition-colors',
        sizeClasses[size],
        variantClasses[variant],
        className
      )}
      title={copied ? 'Copied!' : 'Copy to clipboard'}
    >
      {copied ? (
        <Check className={cn(iconSizes[size], 'text-green-500')} />
      ) : (
        <Copy className={iconSizes[size]} />
      )}
      {label && <span className="text-xs">{label}</span>}
    </button>
  )
}
