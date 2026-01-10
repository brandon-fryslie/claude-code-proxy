import { type FC } from 'react';
import { Terminal, Clock, FolderOpen } from 'lucide-react';
import { cn } from '@/lib/utils';

interface BashToolInput {
  command?: string;
  description?: string;
  timeout?: number;
  run_in_background?: boolean;
  dangerouslyDisableSandbox?: boolean;
}

export const BashTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const {
    command,
    description,
    timeout,
    run_in_background,
    dangerouslyDisableSandbox,
  } = input as BashToolInput;

  return (
    <div className="space-y-2">
      {/* Description */}
      {description && <div className="text-sm text-gray-600 italic">{description}</div>}

      {/* Command display */}
      {command && (
        <div className="font-mono text-sm bg-gray-900 text-gray-100 p-2 rounded-lg overflow-x-auto">
          <div className="flex items-start gap-2">
            <span className="text-green-400 select-none">$</span>
            <span className="whitespace-pre-wrap break-all">{command}</span>
          </div>
        </div>
      )}

      {/* Metadata badges */}
      <div className="flex flex-wrap gap-2">
        {timeout && <Badge icon={Clock} label={`Timeout: ${timeout}ms`} />}
        {run_in_background && <Badge icon={FolderOpen} label="Background" variant="blue" />}
        {dangerouslyDisableSandbox && (
          <Badge icon={Terminal} label="No Sandbox" variant="red" />
        )}
      </div>
    </div>
  );
};

// Reusable badge component
const Badge: FC<{
  icon: FC<{ className?: string }>;
  label: string;
  variant?: 'gray' | 'blue' | 'red';
}> = ({ icon: Icon, label, variant = 'gray' }) => {
  const variants = {
    gray: 'bg-gray-100 text-gray-600 border-gray-200',
    blue: 'bg-blue-50 text-blue-600 border-blue-200',
    red: 'bg-red-50 text-red-600 border-red-200',
  };

  return (
    <div
      className={cn(
        'flex items-center gap-1.5 px-2 py-1 text-xs rounded-full border',
        variants[variant]
      )}
    >
      <Icon className="w-3 h-3" />
      <span>{label}</span>
    </div>
  );
};
