import { type FC } from 'react';
import { FilePlus } from 'lucide-react';
import { CodeViewer } from '../../CodeViewer';

export const WriteTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const path = String(input.file_path || input.path || '');
  const content = String(input.content || '');

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm">
        <FilePlus className="w-4 h-4 text-green-500" />
        <span className="font-mono text-blue-600">{path}</span>
      </div>
      {content && (
        <CodeViewer code={content} filename={path} maxHeight={200} showControls={true} />
      )}
    </div>
  );
};
