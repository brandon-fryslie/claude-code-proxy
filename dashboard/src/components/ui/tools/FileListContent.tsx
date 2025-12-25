import { type FC, useState } from 'react';
import { File, Folder, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileListContentProps {
  content: string;
}

export const FileListContent: FC<FileListContentProps> = ({ content }) => {
  const files = content.split('\n').filter((line) => line.trim());

  // Group by directory
  const grouped = groupByDirectory(files);

  return (
    <div className="text-sm font-mono space-y-1 max-h-64 overflow-y-auto">
      {files.length <= 20 ? (
        // Simple list for small results
        files.map((file, i) => <FileRow key={i} path={file} />)
      ) : (
        // Grouped view for large results
        <GroupedFileView groups={grouped} />
      )}
      <div className="text-xs text-gray-400 pt-2 border-t">{files.length} files</div>
    </div>
  );
};

const FileRow: FC<{ path: string }> = ({ path }) => {
  const isDirectory = path.endsWith('/');

  return (
    <div className="flex items-center gap-2 py-0.5 hover:bg-gray-50 rounded px-1">
      {isDirectory ? (
        <Folder className="w-4 h-4 text-blue-500" />
      ) : (
        <File className="w-4 h-4 text-gray-400" />
      )}
      <span className={cn(isDirectory ? 'text-blue-600' : 'text-gray-700')}>{path}</span>
    </div>
  );
};

const GroupedFileView: FC<{ groups: Record<string, string[]> }> = ({ groups }) => {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const toggleDir = (dir: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(dir)) {
        next.delete(dir);
      } else {
        next.add(dir);
      }
      return next;
    });
  };

  return (
    <div className="space-y-1">
      {Object.entries(groups).map(([dir, files]) => (
        <div key={dir}>
          <button
            onClick={() => toggleDir(dir)}
            className="flex items-center gap-2 w-full text-left py-0.5 hover:bg-gray-50 rounded px-1"
          >
            <ChevronRight
              className={cn(
                'w-3 h-3 text-gray-400 transition-transform',
                expanded.has(dir) && 'rotate-90'
              )}
            />
            <Folder className="w-4 h-4 text-blue-500" />
            <span className="text-blue-600">{dir}/</span>
            <span className="text-gray-400 text-xs">({files.length})</span>
          </button>
          {expanded.has(dir) && (
            <div className="ml-6 border-l border-gray-200 pl-2">
              {files.map((file, i) => (
                <FileRow key={i} path={file} />
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

function groupByDirectory(files: string[]): Record<string, string[]> {
  const groups: Record<string, string[]> = {};

  for (const file of files) {
    const parts = file.split('/');
    const dir = parts.length > 1 ? parts.slice(0, -1).join('/') : '.';
    const name = parts.pop() || file;

    if (!groups[dir]) {
      groups[dir] = [];
    }
    groups[dir].push(name);
  }

  return groups;
}
