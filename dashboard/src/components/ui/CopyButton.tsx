import { type FC, useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface CopyButtonProps {
  content: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  title?: string;
}

export const CopyButton: FC<CopyButtonProps> = ({
  content,
  size = 'md',
  className,
  title = 'Copy to clipboard',
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  };

  const sizeClasses = {
    sm: 'p-1',
    md: 'p-1.5',
    lg: 'p-2',
  };

  const iconSizes = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  return (
    <button
      onClick={handleCopy}
      className={cn(
        'text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors',
        sizeClasses[size],
        className
      )}
      title={title}
    >
      {copied ? (
        <Check className={cn(iconSizes[size], 'text-green-500')} />
      ) : (
        <Copy className={iconSizes[size]} />
      )}
    </button>
  );
};
