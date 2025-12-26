import { type FC } from 'react';
import { FileSearch, FolderOpen, FileType } from 'lucide-react';

export const GrepTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const pattern = String(input.pattern || '');
  const path = input.path as string | undefined;
  const glob = input.glob as string | undefined;
  const type = input.type as string | undefined;
  const outputMode = input.output_mode as string | undefined;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileSearch className="w-4 h-4 text-purple-500" />
        <code className="px-2 py-1 bg-purple-50 border border-purple-200 rounded text-sm font-mono text-purple-700">
          /{pattern}/
        </code>
      </div>
      <div className="flex flex-wrap gap-2 text-xs">
        {path && (
          <span className="flex items-center gap-1 text-gray-500">
            <FolderOpen className="w-3 h-3" />
            {path}
          </span>
        )}
        {glob && <span className="px-2 py-0.5 bg-gray-100 rounded">glob: {glob}</span>}
        {type && (
          <span className="flex items-center gap-1 px-2 py-0.5 bg-blue-50 text-blue-600 rounded">
            <FileType className="w-3 h-3" />
            {type}
          </span>
        )}
        {outputMode && <span className="px-2 py-0.5 bg-gray-100 rounded">{outputMode}</span>}
      </div>
    </div>
  );
};
