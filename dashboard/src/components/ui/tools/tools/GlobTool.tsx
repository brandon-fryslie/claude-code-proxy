import { type FC } from 'react';
import { Search, FolderOpen } from 'lucide-react';

export const GlobTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const pattern = String(input.pattern || '');
  const path = input.path as string | undefined;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Search className="w-4 h-4 text-gray-400" />
        <code className="px-2 py-1 bg-gray-100 rounded text-sm font-mono">{pattern}</code>
      </div>
      {path && (
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <FolderOpen className="w-3 h-3" />
          <span>in {path}</span>
        </div>
      )}
    </div>
  );
};
