import { type FC, useState } from 'react';
import { ChevronDown, Terminal, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { CopyButton } from '../CopyButton';
import { ToolInputGeneric } from './ToolInputGeneric';

// Import specialized tools
import { BashTool } from './tools/BashTool';
import { ReadTool } from './tools/ReadTool';
import { WriteTool } from './tools/WriteTool';
import { EditTool } from './tools/EditTool';
import { GlobTool } from './tools/GlobTool';
import { GrepTool } from './tools/GrepTool';
import { TaskTool } from './tools/TaskTool';
import { TodoWriteTool } from './tools/TodoWriteTool';

interface ToolUseContainerProps {
  id: string;
  name: string;
  input: Record<string, unknown>;
  isExecuting?: boolean;
  defaultExpanded?: boolean;
}

// Map tool names to specialized renderers
const TOOL_RENDERERS: Record<string, FC<{ input: Record<string, unknown> }>> = {
  Bash: BashTool,
  bash: BashTool,
  Read: ReadTool,
  read: ReadTool,
  read_file: ReadTool,
  Write: WriteTool,
  write: WriteTool,
  write_file: WriteTool,
  Edit: EditTool,
  edit: EditTool,
  edit_file: EditTool,
  Glob: GlobTool,
  glob: GlobTool,
  Grep: GrepTool,
  grep: GrepTool,
  Task: TaskTool,
  task: TaskTool,
  TodoWrite: TodoWriteTool,
  todowrite: TodoWriteTool,
};

export const ToolUseContainer: FC<ToolUseContainerProps> = ({
  id,
  name,
  input,
  isExecuting = true,
  defaultExpanded = false,
}) => {
  const [expanded, setExpanded] = useState(defaultExpanded);

  // Get specialized renderer or fall back to generic
  const ToolRenderer = TOOL_RENDERERS[name] || ToolInputGeneric;

  return (
    <div className="border border-indigo-200 rounded-lg bg-gradient-to-r from-indigo-50 to-blue-50 overflow-hidden shadow-sm">
      {/* Header - Always visible */}
      <ToolHeader
        name={name}
        id={id}
        expanded={expanded}
        isExecuting={isExecuting}
        onToggle={() => setExpanded(!expanded)}
      />

      {/* Expandable content */}
      {expanded && (
        <div className="px-4 py-3 border-t border-indigo-100 bg-white/60">
          <ToolRenderer input={input} />
        </div>
      )}

      {/* Execution indicator footer */}
      {isExecuting && (
        <div className="px-4 py-1.5 bg-indigo-100/50 text-xs text-indigo-600 flex items-center gap-2 border-t border-indigo-100">
          <Loader2 className="w-3 h-3 animate-spin" />
          <span>Executing {name}...</span>
        </div>
      )}
    </div>
  );
};

// Header Component
const ToolHeader: FC<{
  name: string;
  id: string;
  expanded: boolean;
  isExecuting: boolean;
  onToggle: () => void;
}> = ({ name, id, expanded, isExecuting, onToggle }) => {
  const shortId = id.slice(-8);

  return (
    <div
      className="flex items-center justify-between px-4 py-2.5 cursor-pointer hover:bg-white/50 transition-colors group"
      onClick={onToggle}
    >
      <div className="flex items-center gap-3">
        {/* Tool icon */}
        <div className="w-8 h-8 rounded-lg bg-indigo-100 flex items-center justify-center">
          <Terminal className="w-4 h-4 text-indigo-600" />
        </div>

        {/* Tool name and ID */}
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-indigo-900">{name}</span>
            {isExecuting && (
              <span className="w-2 h-2 bg-indigo-500 rounded-full animate-pulse" />
            )}
          </div>
          <div className="flex items-center gap-1 text-xs text-gray-400">
            <span className="font-mono">{shortId}</span>
            <CopyButton
              content={id}
              size="sm"
              className="opacity-0 group-hover:opacity-100"
            />
          </div>
        </div>
      </div>

      {/* Expand/collapse chevron */}
      <ChevronDown
        className={cn(
          'w-5 h-5 text-gray-400 transition-transform duration-200',
          expanded && 'rotate-180'
        )}
      />
    </div>
  );
};
