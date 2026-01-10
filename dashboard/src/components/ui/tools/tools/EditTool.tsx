import { type FC, useMemo } from 'react';
import { FileEdit } from 'lucide-react';
import { cn } from '@/lib/utils';

interface EditToolInput {
  file_path?: string;
  old_string?: string;
  new_string?: string;
  replace_all?: boolean;
}

export const EditTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const { file_path, old_string, new_string, replace_all } = input as EditToolInput;

  return (
    <div className="space-y-2">
      {/* File path */}
      {file_path && (
        <div className="flex items-center gap-2 text-sm">
          <FileEdit className="w-4 h-4 text-gray-400" />
          <span className="font-mono text-blue-600">{file_path}</span>
          {replace_all && (
            <span className="px-2 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">
              Replace All
            </span>
          )}
        </div>
      )}

      {/* Side-by-side diff */}
      {old_string !== undefined && new_string !== undefined && (
        <CodeDiff
          oldCode={old_string}
          newCode={new_string}
          language={getLanguageFromPath(file_path || '')}
        />
      )}
    </div>
  );
};

// Side-by-side diff component
interface CodeDiffProps {
  oldCode: string;
  newCode: string;
  language?: string;
}

const CodeDiff: FC<CodeDiffProps> = ({ oldCode, newCode }) => {
  const { oldLines, newLines, changes } = useMemo(() => {
    return computeDiff(oldCode, newCode);
  }, [oldCode, newCode]);

  return (
    <div className="grid grid-cols-2 gap-2">
      {/* Old code (deletions) */}
      <div className="rounded-lg overflow-hidden border border-red-200">
        <div className="bg-red-50 px-2 py-1 text-xs font-medium text-red-700 border-b border-red-200">
          Old
        </div>
        <div className="bg-gray-900 p-2 text-sm font-mono overflow-x-auto max-h-64">
          {oldLines.map((line, i) => (
            <div
              key={i}
              className={cn(
                'px-2 py-0.5',
                changes.removed.has(i) && 'bg-red-900/30 text-red-300'
              )}
            >
              <span className="text-gray-500 select-none mr-3">{i + 1}</span>
              <span className="text-gray-100">{line || ' '}</span>
            </div>
          ))}
        </div>
      </div>

      {/* New code (additions) */}
      <div className="rounded-lg overflow-hidden border border-green-200">
        <div className="bg-green-50 px-2 py-1 text-xs font-medium text-green-700 border-b border-green-200">
          New
        </div>
        <div className="bg-gray-900 p-2 text-sm font-mono overflow-x-auto max-h-64">
          {newLines.map((line, i) => (
            <div
              key={i}
              className={cn(
                'px-2 py-0.5',
                changes.added.has(i) && 'bg-green-900/30 text-green-300'
              )}
            >
              <span className="text-gray-500 select-none mr-3">{i + 1}</span>
              <span className="text-gray-100">{line || ' '}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// Simple line-based diff algorithm
function computeDiff(oldCode: string, newCode: string) {
  const oldLines = oldCode.split('\n');
  const newLines = newCode.split('\n');

  // Find common prefix
  let prefixLen = 0;
  while (
    prefixLen < oldLines.length &&
    prefixLen < newLines.length &&
    oldLines[prefixLen] === newLines[prefixLen]
  ) {
    prefixLen++;
  }

  // Find common suffix
  let suffixLen = 0;
  while (
    suffixLen < oldLines.length - prefixLen &&
    suffixLen < newLines.length - prefixLen &&
    oldLines[oldLines.length - 1 - suffixLen] === newLines[newLines.length - 1 - suffixLen]
  ) {
    suffixLen++;
  }

  // Mark changed lines
  const removed = new Set<number>();
  const added = new Set<number>();

  for (let i = prefixLen; i < oldLines.length - suffixLen; i++) {
    removed.add(i);
  }
  for (let i = prefixLen; i < newLines.length - suffixLen; i++) {
    added.add(i);
  }

  return {
    oldLines,
    newLines,
    changes: { removed, added },
  };
}

function getLanguageFromPath(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || '';
  const langMap: Record<string, string> = {
    ts: 'typescript',
    tsx: 'typescript',
    js: 'javascript',
    jsx: 'javascript',
    py: 'python',
    go: 'go',
    rs: 'rust',
    md: 'markdown',
    json: 'json',
    yaml: 'yaml',
  };
  return langMap[ext] || 'text';
}
