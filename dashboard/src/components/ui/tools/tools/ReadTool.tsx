import { type FC } from 'react';
import { FileText, Hash } from 'lucide-react';

export const ReadTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const path = input.file_path || input.path || '';
  const offset = input.offset as number | undefined;
  const limit = input.limit as number | undefined;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm">
        <FileText className="w-4 h-4 text-gray-400" />
        <span className="font-mono text-blue-600">{String(path)}</span>
      </div>
      {(offset || limit) && (
        <div className="flex gap-3 text-xs text-gray-500">
          {offset !== undefined && (
            <span className="flex items-center gap-1">
              <Hash className="w-3 h-3" />
              Offset: {offset}
            </span>
          )}
          {limit && (
            <span className="flex items-center gap-1">
              <Hash className="w-3 h-3" />
              Limit: {limit} lines
            </span>
          )}
        </div>
      )}
    </div>
  );
};
