# Phase 3: Tool Support - Implementation Handoff

**Branch:** `phase-3-tool-support`
**Epic:** `brandon-fryslie_claude-code-proxy-62s` (tool-use-display), `brandon-fryslie_claude-code-proxy-k6y` (tool-result-display), `brandon-fryslie_claude-code-proxy-asz` (image-content-display)

---

## Executive Summary

You are implementing the **Tool Support** system for the new dashboard. Claude Code uses dozens of tools (Bash, Read, Write, Edit, Glob, Grep, etc.), and your work makes these tool calls and their results beautifully displayable.

**Prerequisites:** Phase 2 (Rich Content Display) should be complete before you start. You'll build on top of `MessageContent`, `CodeViewer`, and `CopyButton` components.

Your deliverables:
1. **Tool Use Display** - Expandable, specialized views for each tool type
2. **Tool Result Display** - Formatted results with content type detection
3. **Image Content Display** - Base64 image rendering with zoom/lightbox

---

## Your Working Environment

### Directory Structure
```
/Users/bmf/code/brandon-fryslie_claude-code-proxy/
├── dashboard/              # NEW dashboard (your target)
│   ├── src/
│   │   ├── components/
│   │   │   ├── layout/     # Sidebar, AppLayout, ResizablePanel
│   │   │   └── ui/         # Should have Phase 2 components
│   │   │       ├── MessageContent.tsx    # From Phase 2
│   │   │       ├── CodeViewer.tsx        # From Phase 2
│   │   │       ├── CopyButton.tsx        # From Phase 2
│   │   │       └── ...                   # Your new components go here
│   │   ├── pages/
│   │   └── lib/
│   │       ├── types.ts    # TypeScript interfaces
│   │       ├── api.ts      # React Query hooks
│   │       └── utils.ts    # cn() utility
├── web/                    # OLD dashboard (reference)
│   └── app/components/
│       ├── ToolUse.tsx          # REFERENCE
│       ├── ToolResult.tsx       # REFERENCE
│       ├── ImageContent.tsx     # REFERENCE
│       ├── CodeDiff.tsx         # REFERENCE: For Edit tool
│       └── TodoList.tsx         # REFERENCE: For TodoWrite tool
└── proxy/                  # Go backend
```

### Tech Stack
- React 19.2.0
- TypeScript 5.9.3 (strict mode)
- Tailwind CSS 4.1.18
- Lucide React 0.562.0

---

## The Tools You'll Display

Claude Code has access to these tools. You need specialized rendering for each:

### High Priority (Most Common)
| Tool | Purpose | Special Rendering Needed |
|------|---------|-------------------------|
| `Bash` | Execute shell commands | Command preview, exit codes, output formatting |
| `Read` | Read file contents | File path display, line number support |
| `Write` | Write entire files | File path, content preview with syntax highlighting |
| `Edit` | Edit portions of files | **Side-by-side diff** (old vs new) |
| `Glob` | Find files by pattern | Pattern display, file list results |
| `Grep` | Search file contents | Pattern, matches with context |
| `Task` | Launch sub-agents | Agent type, prompt preview |

### Medium Priority
| Tool | Purpose | Special Rendering Needed |
|------|---------|-------------------------|
| `TodoWrite` | Manage todo lists | **Formatted todo list** with status icons |
| `WebFetch` | Fetch web content | URL display, content preview |
| `WebSearch` | Search the web | Query display, results list |
| `AskUserQuestion` | Ask user questions | Question display, options |
| `NotebookEdit` | Edit Jupyter notebooks | Cell number, content |

### Lower Priority (Generic OK)
| Tool | Purpose |
|------|---------|
| `MultiEdit` | Multiple edits in one call |
| `KillShell` | Terminate background shell |
| `Skill` | Invoke skills |
| `EnterPlanMode` | Enter planning mode |
| `ExitPlanMode` | Exit planning mode |

---

## Topic 1: Tool Use Display (Expandable)

### Component Architecture

```
components/ui/tools/
├── ToolUseContainer.tsx    # Main wrapper with expand/collapse
├── ToolHeader.tsx          # Tool name, ID, status indicator
├── ToolInputGeneric.tsx    # Fallback for unknown tools
├── tools/                  # Specialized tool renderers
│   ├── BashTool.tsx
│   ├── ReadTool.tsx
│   ├── WriteTool.tsx
│   ├── EditTool.tsx        # With CodeDiff
│   ├── GlobTool.tsx
│   ├── GrepTool.tsx
│   ├── TaskTool.tsx
│   └── TodoWriteTool.tsx
└── index.ts
```

### ToolUseContainer.tsx - Main Component

```tsx
// dashboard/src/components/ui/tools/ToolUseContainer.tsx
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
  isExecuting?: boolean;  // Show pulsing indicator
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
      className="flex items-center justify-between px-4 py-2.5 cursor-pointer hover:bg-white/50 transition-colors"
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
          "w-5 h-5 text-gray-400 transition-transform duration-200",
          expanded && "rotate-180"
        )}
      />
    </div>
  );
};
```

### BashTool.tsx - Shell Command Display

```tsx
// dashboard/src/components/ui/tools/tools/BashTool.tsx
import { type FC } from 'react';
import { Terminal, Clock, FolderOpen } from 'lucide-react';

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
    dangerouslyDisableSandbox
  } = input as BashToolInput;

  return (
    <div className="space-y-3">
      {/* Description */}
      {description && (
        <div className="text-sm text-gray-600 italic">
          {description}
        </div>
      )}

      {/* Command display */}
      {command && (
        <div className="font-mono text-sm bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto">
          <div className="flex items-start gap-2">
            <span className="text-green-400 select-none">$</span>
            <span className="whitespace-pre-wrap break-all">{command}</span>
          </div>
        </div>
      )}

      {/* Metadata badges */}
      <div className="flex flex-wrap gap-2">
        {timeout && (
          <Badge icon={Clock} label={`Timeout: ${timeout}ms`} />
        )}
        {run_in_background && (
          <Badge icon={FolderOpen} label="Background" variant="blue" />
        )}
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
    <div className={cn(
      "flex items-center gap-1.5 px-2 py-1 text-xs rounded-full border",
      variants[variant]
    )}>
      <Icon className="w-3 h-3" />
      <span>{label}</span>
    </div>
  );
};
```

### EditTool.tsx - With Side-by-Side Diff

This is the most complex tool renderer. The Edit tool shows what code is being changed.

```tsx
// dashboard/src/components/ui/tools/tools/EditTool.tsx
import { type FC, useMemo } from 'react';
import { FileEdit, ArrowRight } from 'lucide-react';
import { CodeViewer } from '../../CodeViewer';

interface EditToolInput {
  file_path?: string;
  old_string?: string;
  new_string?: string;
  replace_all?: boolean;
}

export const EditTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const { file_path, old_string, new_string, replace_all } = input as EditToolInput;

  return (
    <div className="space-y-3">
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

const CodeDiff: FC<CodeDiffProps> = ({ oldCode, newCode, language = 'text' }) => {
  const { oldLines, newLines, changes } = useMemo(() => {
    return computeDiff(oldCode, newCode);
  }, [oldCode, newCode]);

  return (
    <div className="grid grid-cols-2 gap-2">
      {/* Old code (deletions) */}
      <div className="rounded-lg overflow-hidden border border-red-200">
        <div className="bg-red-50 px-3 py-1.5 text-xs font-medium text-red-700 border-b border-red-200">
          Old
        </div>
        <div className="bg-gray-900 p-2 text-sm font-mono overflow-x-auto max-h-64">
          {oldLines.map((line, i) => (
            <div
              key={i}
              className={cn(
                "px-2 py-0.5",
                changes.removed.has(i) && "bg-red-900/30 text-red-300"
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
        <div className="bg-green-50 px-3 py-1.5 text-xs font-medium text-green-700 border-b border-green-200">
          New
        </div>
        <div className="bg-gray-900 p-2 text-sm font-mono overflow-x-auto max-h-64">
          {newLines.map((line, i) => (
            <div
              key={i}
              className={cn(
                "px-2 py-0.5",
                changes.added.has(i) && "bg-green-900/30 text-green-300"
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
    ts: 'typescript', tsx: 'typescript',
    js: 'javascript', jsx: 'javascript',
    py: 'python', go: 'go', rs: 'rust',
    md: 'markdown', json: 'json', yaml: 'yaml',
  };
  return langMap[ext] || 'text';
}
```

### TodoWriteTool.tsx - Formatted Todo List

```tsx
// dashboard/src/components/ui/tools/tools/TodoWriteTool.tsx
import { type FC } from 'react';
import { CheckCircle, Circle, Clock, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Todo {
  content?: string;
  task?: string;
  description?: string;
  title?: string;
  text?: string;
  status: 'pending' | 'in_progress' | 'completed';
  activeForm?: string;
  priority?: 'high' | 'medium' | 'low';
}

interface TodoWriteInput {
  todos?: Todo[];
}

export const TodoWriteTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const { todos = [] } = input as TodoWriteInput;

  if (todos.length === 0) {
    return (
      <div className="text-sm text-gray-500 italic">
        Empty todo list
      </div>
    );
  }

  // Group by status
  const grouped = {
    in_progress: todos.filter(t => t.status === 'in_progress'),
    pending: todos.filter(t => t.status === 'pending'),
    completed: todos.filter(t => t.status === 'completed'),
  };

  // Count summary
  const total = todos.length;
  const done = grouped.completed.length;

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex items-center gap-3 text-sm">
        <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-green-500 transition-all"
            style={{ width: `${(done / total) * 100}%` }}
          />
        </div>
        <span className="text-gray-600">
          {done}/{total} complete
        </span>
      </div>

      {/* In Progress - show first */}
      {grouped.in_progress.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-blue-600 uppercase tracking-wide">
            In Progress
          </div>
          {grouped.in_progress.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}

      {/* Pending */}
      {grouped.pending.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            Pending
          </div>
          {grouped.pending.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}

      {/* Completed */}
      {grouped.completed.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-green-600 uppercase tracking-wide">
            Completed
          </div>
          {grouped.completed.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}
    </div>
  );
};

const TodoItem: FC<{ todo: Todo }> = ({ todo }) => {
  // Get the todo text from various possible fields
  const text = todo.content || todo.task || todo.description || todo.title || todo.text || '';

  const statusConfig = {
    pending: {
      icon: Circle,
      iconClass: 'text-gray-400',
      textClass: 'text-gray-700',
    },
    in_progress: {
      icon: Clock,
      iconClass: 'text-blue-500 animate-pulse',
      textClass: 'text-blue-700 font-medium',
    },
    completed: {
      icon: CheckCircle,
      iconClass: 'text-green-500',
      textClass: 'text-gray-500 line-through',
    },
  };

  const config = statusConfig[todo.status] || statusConfig.pending;
  const Icon = config.icon;

  const priorityColors = {
    high: 'bg-red-100 text-red-700 border-red-200',
    medium: 'bg-amber-100 text-amber-700 border-amber-200',
    low: 'bg-blue-100 text-blue-700 border-blue-200',
  };

  return (
    <div className="flex items-start gap-2 py-1.5">
      <Icon className={cn("w-4 h-4 mt-0.5 flex-shrink-0", config.iconClass)} />
      <div className="flex-1 min-w-0">
        <span className={cn("text-sm", config.textClass)}>
          {todo.status === 'in_progress' && todo.activeForm
            ? todo.activeForm
            : text}
        </span>
        {todo.priority && (
          <span className={cn(
            "ml-2 px-1.5 py-0.5 text-xs rounded border",
            priorityColors[todo.priority]
          )}>
            {todo.priority}
          </span>
        )}
      </div>
    </div>
  );
};
```

### Other Tool Renderers

I'll provide compact versions of the other specialized tools:

```tsx
// dashboard/src/components/ui/tools/tools/ReadTool.tsx
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
          {offset && (
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

// dashboard/src/components/ui/tools/tools/WriteTool.tsx
import { type FC } from 'react';
import { FilePlus } from 'lucide-react';
import { CodeViewer } from '../../CodeViewer';

export const WriteTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const path = String(input.file_path || input.path || '');
  const content = String(input.content || '');

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 text-sm">
        <FilePlus className="w-4 h-4 text-green-500" />
        <span className="font-mono text-blue-600">{path}</span>
      </div>
      {content && (
        <CodeViewer
          code={content}
          filename={path}
          maxHeight={200}
          showControls={true}
        />
      )}
    </div>
  );
};

// dashboard/src/components/ui/tools/tools/GlobTool.tsx
import { type FC } from 'react';
import { Search, FolderOpen } from 'lucide-react';

export const GlobTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const pattern = String(input.pattern || '');
  const path = input.path as string | undefined;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Search className="w-4 h-4 text-gray-400" />
        <code className="px-2 py-1 bg-gray-100 rounded text-sm font-mono">
          {pattern}
        </code>
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

// dashboard/src/components/ui/tools/tools/GrepTool.tsx
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
        {glob && (
          <span className="px-2 py-0.5 bg-gray-100 rounded">
            glob: {glob}
          </span>
        )}
        {type && (
          <span className="flex items-center gap-1 px-2 py-0.5 bg-blue-50 text-blue-600 rounded">
            <FileType className="w-3 h-3" />
            {type}
          </span>
        )}
        {outputMode && (
          <span className="px-2 py-0.5 bg-gray-100 rounded">
            {outputMode}
          </span>
        )}
      </div>
    </div>
  );
};

// dashboard/src/components/ui/tools/tools/TaskTool.tsx
import { type FC } from 'react';
import { Users, FileText } from 'lucide-react';

export const TaskTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const subagentType = String(input.subagent_type || '');
  const description = String(input.description || '');
  const prompt = String(input.prompt || '');
  const model = input.model as string | undefined;
  const runInBackground = input.run_in_background as boolean | undefined;

  return (
    <div className="space-y-3">
      {/* Agent type */}
      <div className="flex items-center gap-2">
        <Users className="w-4 h-4 text-indigo-500" />
        <span className="font-medium text-indigo-700">{subagentType}</span>
        {model && (
          <span className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded">
            {model}
          </span>
        )}
        {runInBackground && (
          <span className="px-2 py-0.5 bg-blue-50 text-blue-600 text-xs rounded">
            Background
          </span>
        )}
      </div>

      {/* Description */}
      {description && (
        <div className="text-sm text-gray-600">
          {description}
        </div>
      )}

      {/* Prompt preview */}
      {prompt && (
        <details className="text-sm">
          <summary className="cursor-pointer text-blue-600 hover:text-blue-700">
            <FileText className="w-4 h-4 inline mr-1" />
            Show prompt ({prompt.length} chars)
          </summary>
          <pre className="mt-2 p-3 bg-gray-50 rounded text-xs overflow-x-auto whitespace-pre-wrap">
            {prompt}
          </pre>
        </details>
      )}
    </div>
  );
};
```

### Generic Fallback Renderer

```tsx
// dashboard/src/components/ui/tools/ToolInputGeneric.tsx
import { type FC, useState } from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';

export const ToolInputGeneric: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const entries = Object.entries(input);

  if (entries.length === 0) {
    return <span className="text-gray-400 text-sm italic">No parameters</span>;
  }

  // Show first 3 params, rest in expandable section
  const visibleParams = entries.slice(0, 3);
  const hiddenParams = entries.slice(3);
  const [showAll, setShowAll] = useState(false);

  return (
    <div className="space-y-2">
      {visibleParams.map(([key, value]) => (
        <ParamRow key={key} name={key} value={value} />
      ))}

      {hiddenParams.length > 0 && (
        <>
          {showAll && hiddenParams.map(([key, value]) => (
            <ParamRow key={key} name={key} value={value} />
          ))}
          <button
            onClick={() => setShowAll(!showAll)}
            className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700"
          >
            <ChevronDown className={cn("w-3 h-3 transition-transform", showAll && "rotate-180")} />
            {showAll ? 'Show less' : `Show ${hiddenParams.length} more parameters`}
          </button>
        </>
      )}
    </div>
  );
};

const ParamRow: FC<{ name: string; value: unknown }> = ({ name, value }) => {
  return (
    <div className="text-sm">
      <span className="text-gray-500 mr-2">{name}:</span>
      <ParamValue value={value} />
    </div>
  );
};

const ParamValue: FC<{ value: unknown }> = ({ value }) => {
  if (value === null || value === undefined) {
    return <span className="text-gray-400 italic">null</span>;
  }

  if (typeof value === 'boolean') {
    return (
      <span className={value ? 'text-green-600' : 'text-red-600'}>
        {String(value)}
      </span>
    );
  }

  if (typeof value === 'number') {
    return <span className="text-cyan-600 font-mono">{value}</span>;
  }

  if (typeof value === 'string') {
    // Truncate long strings
    if (value.length > 100 || value.includes('\n')) {
      return (
        <details className="inline">
          <summary className="cursor-pointer text-blue-600">
            String ({value.length} chars)
          </summary>
          <pre className="mt-1 p-2 bg-gray-50 rounded text-xs overflow-x-auto whitespace-pre-wrap max-h-48">
            {value}
          </pre>
        </details>
      );
    }
    return <span className="font-mono text-amber-700">"{value}"</span>;
  }

  if (typeof value === 'object') {
    const json = JSON.stringify(value, null, 2);
    return (
      <details className="inline">
        <summary className="cursor-pointer text-blue-600">
          {Array.isArray(value) ? `Array (${value.length})` : `Object (${Object.keys(value).length} keys)`}
        </summary>
        <pre className="mt-1 p-2 bg-gray-50 rounded text-xs overflow-x-auto max-h-48">
          {json}
        </pre>
      </details>
    );
  }

  return <span className="font-mono">{String(value)}</span>;
};
```

---

## Topic 2: Tool Result Display

Enhance the `ToolResultContent` component from Phase 2 with smarter content detection.

### Enhanced Content Type Detection

```tsx
// dashboard/src/components/ui/tools/ToolResultContent.tsx
// Enhance the existing component with better detection

// Add these detection functions:

interface ContentAnalysis {
  type: 'code' | 'json' | 'file_list' | 'error' | 'table' | 'text';
  language?: string;
  metadata?: Record<string, unknown>;
}

function analyzeContent(content: string): ContentAnalysis {
  const trimmed = content.trim();

  // Error detection
  if (
    trimmed.startsWith('Error:') ||
    trimmed.startsWith('error:') ||
    /^(ENOENT|EACCES|EPERM|EEXIST)/.test(trimmed) ||
    trimmed.includes('Permission denied') ||
    trimmed.includes('No such file or directory')
  ) {
    return { type: 'error' };
  }

  // JSON detection
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      JSON.parse(trimmed);
      return { type: 'json' };
    } catch {
      // Not valid JSON, continue checking
    }
  }

  // File list (from glob/find)
  if (isFileList(trimmed)) {
    return { type: 'file_list' };
  }

  // Code with line numbers (cat -n format)
  if (/^\s*\d+[→\t]/.test(trimmed)) {
    const language = detectLanguageFromContent(trimmed);
    return { type: 'code', language };
  }

  // Code by content patterns
  if (hasCodePatterns(trimmed)) {
    const language = detectLanguageFromContent(trimmed);
    return { type: 'code', language };
  }

  // Table format (from ls -la, ps, etc.)
  if (isTableFormat(trimmed)) {
    return { type: 'table' };
  }

  return { type: 'text' };
}

function isFileList(content: string): boolean {
  const lines = content.split('\n').filter(l => l.trim());
  if (lines.length < 2) return false;

  // Check if most lines look like file paths
  const pathLikeLines = lines.filter(line =>
    /^\.?\//.test(line) ||  // Starts with / or ./
    /\.(ts|tsx|js|jsx|py|go|rs|md|json|yaml|yml|toml|txt|css|html)$/.test(line) ||
    /^[a-zA-Z0-9_-]+\//.test(line)  // Looks like relative path
  );

  return pathLikeLines.length >= lines.length * 0.7;
}

function isTableFormat(content: string): boolean {
  const lines = content.split('\n').filter(l => l.trim());
  if (lines.length < 2) return false;

  // Check for consistent column structure (multiple whitespace-separated columns)
  const columnCounts = lines.map(line =>
    line.trim().split(/\s{2,}/).length
  );

  const avgColumns = columnCounts.reduce((a, b) => a + b, 0) / columnCounts.length;
  return avgColumns >= 3 && columnCounts.every(c => Math.abs(c - avgColumns) <= 2);
}

function hasCodePatterns(content: string): boolean {
  const patterns = [
    /^(import|from|export|const|let|var|function|class|def|func|package)\s/m,
    /[{}\[\]];?\s*$/m,
    /^\s*(if|for|while|return|throw|try|catch)\s*[\(\{]/m,
    /=>\s*[{\(]/,
    /\bexport\s+(default\s+)?/,
    /^#!/,  // Shebang
  ];
  return patterns.some(p => p.test(content));
}

function detectLanguageFromContent(content: string): string {
  // Python indicators
  if (/^(def|class|import|from)\s/.test(content) || /:\s*$/.test(content.split('\n')[0] || '')) {
    return 'python';
  }

  // Go indicators
  if (/^(func|package|import)\s/.test(content) || /\bfmt\./.test(content)) {
    return 'go';
  }

  // Rust indicators
  if (/^(fn|use|mod|impl|struct|enum)\s/.test(content) || /\bmut\s/.test(content)) {
    return 'rust';
  }

  // TypeScript/JavaScript (most common in this context)
  if (/^(const|let|var|function|class|interface|type)\s/.test(content)) {
    return 'typescript';
  }

  // Bash
  if (/^#!/.test(content) || /\$\{?\w+\}?/.test(content)) {
    return 'bash';
  }

  return 'text';
}
```

### File List Renderer

```tsx
// dashboard/src/components/ui/tools/FileListContent.tsx
import { type FC, useState } from 'react';
import { File, Folder, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileListContentProps {
  content: string;
}

export const FileListContent: FC<FileListContentProps> = ({ content }) => {
  const files = content.split('\n').filter(line => line.trim());

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
      <div className="text-xs text-gray-400 pt-2 border-t">
        {files.length} files
      </div>
    </div>
  );
};

const FileRow: FC<{ path: string }> = ({ path }) => {
  const isDirectory = path.endsWith('/');
  const name = path.split('/').pop() || path;

  return (
    <div className="flex items-center gap-2 py-0.5 hover:bg-gray-50 rounded px-1">
      {isDirectory ? (
        <Folder className="w-4 h-4 text-blue-500" />
      ) : (
        <File className="w-4 h-4 text-gray-400" />
      )}
      <span className={cn(
        isDirectory ? 'text-blue-600' : 'text-gray-700'
      )}>
        {path}
      </span>
    </div>
  );
};

const GroupedFileView: FC<{ groups: Record<string, string[]> }> = ({ groups }) => {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const toggleDir = (dir: string) => {
    setExpanded(prev => {
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
                "w-3 h-3 text-gray-400 transition-transform",
                expanded.has(dir) && "rotate-90"
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
```

---

## Topic 3: Image Content Display

### ImageContent.tsx

```tsx
// dashboard/src/components/ui/ImageContent.tsx
import { type FC, useState } from 'react';
import { Image as ImageIcon, Download, Maximize2, X, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ImageSource {
  type: 'base64';
  media_type: 'image/jpeg' | 'image/png' | 'image/gif' | 'image/webp';
  data: string;
}

interface ImageContentProps {
  source: ImageSource;
  alt?: string;
  maxHeight?: number;
}

export const ImageContent: FC<ImageContentProps> = ({
  source,
  alt = 'Image content',
  maxHeight = 400,
}) => {
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [loadError, setLoadError] = useState(false);

  // Construct data URI
  const dataUri = `data:${source.media_type};base64,${source.data}`;

  // Get file extension for download
  const extension = source.media_type.split('/')[1] || 'png';

  const handleDownload = () => {
    const link = document.createElement('a');
    link.href = dataUri;
    link.download = `image.${extension}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loadError) {
    return (
      <div className="flex items-center gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
        <AlertCircle className="w-5 h-5" />
        <div>
          <div className="font-medium">Failed to load image</div>
          <div className="text-sm text-red-600">
            Format: {source.media_type} | Size: {Math.round(source.data.length / 1024)}KB base64
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      {/* Main image display */}
      <div className="relative group rounded-lg overflow-hidden border border-gray-200 bg-gray-50">
        {/* Image info badge */}
        <div className="absolute top-2 left-2 z-10 flex items-center gap-1.5 px-2 py-1 bg-black/60 text-white text-xs rounded">
          <ImageIcon className="w-3 h-3" />
          <span>{source.media_type.split('/')[1].toUpperCase()}</span>
        </div>

        {/* Control buttons */}
        <div className="absolute top-2 right-2 z-10 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => setLightboxOpen(true)}
            className="p-1.5 bg-black/60 text-white rounded hover:bg-black/80 transition-colors"
            title="View fullscreen"
          >
            <Maximize2 className="w-4 h-4" />
          </button>
          <button
            onClick={handleDownload}
            className="p-1.5 bg-black/60 text-white rounded hover:bg-black/80 transition-colors"
            title="Download image"
          >
            <Download className="w-4 h-4" />
          </button>
        </div>

        {/* The image */}
        <img
          src={dataUri}
          alt={alt}
          onError={() => setLoadError(true)}
          className="w-full object-contain cursor-zoom-in"
          style={{ maxHeight }}
          onClick={() => setLightboxOpen(true)}
        />
      </div>

      {/* Lightbox */}
      {lightboxOpen && (
        <Lightbox
          src={dataUri}
          alt={alt}
          onClose={() => setLightboxOpen(false)}
          onDownload={handleDownload}
        />
      )}
    </>
  );
};

// Lightbox component for fullscreen view
interface LightboxProps {
  src: string;
  alt: string;
  onClose: () => void;
  onDownload: () => void;
}

const Lightbox: FC<LightboxProps> = ({ src, alt, onClose, onDownload }) => {
  // Close on escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
      onClick={onClose}
    >
      {/* Controls */}
      <div className="absolute top-4 right-4 flex gap-2 z-10">
        <button
          onClick={(e) => { e.stopPropagation(); onDownload(); }}
          className="p-2 bg-white/20 text-white rounded-lg hover:bg-white/30 transition-colors"
          title="Download"
        >
          <Download className="w-5 h-5" />
        </button>
        <button
          onClick={onClose}
          className="p-2 bg-white/20 text-white rounded-lg hover:bg-white/30 transition-colors"
          title="Close"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Image */}
      <img
        src={src}
        alt={alt}
        className="max-w-[90vw] max-h-[90vh] object-contain"
        onClick={(e) => e.stopPropagation()}
      />
    </div>
  );
};

// Don't forget the import at the top
import { useEffect } from 'react';
```

---

## Testing Checklist

### Tool Use Display
- [ ] BashTool shows command with proper formatting
- [ ] ReadTool shows file path and optional line range
- [ ] WriteTool shows file path and content preview
- [ ] EditTool shows side-by-side diff (old vs new)
- [ ] GlobTool shows pattern and optional path
- [ ] GrepTool shows regex pattern and options
- [ ] TaskTool shows agent type, description, and prompt preview
- [ ] TodoWriteTool shows formatted todo list with status grouping
- [ ] Generic fallback handles unknown tools gracefully
- [ ] Expand/collapse works on all tools
- [ ] Tool ID copy button works
- [ ] Execution indicator shows pulsing animation

### Tool Result Display
- [ ] Detects and renders code content correctly
- [ ] Detects and renders JSON content correctly
- [ ] Detects and renders file lists correctly
- [ ] Detects and renders error content correctly
- [ ] Truncates long content with "Show more" button
- [ ] Error results show red styling
- [ ] Success results show green styling

### Image Content
- [ ] Renders base64 JPEG images
- [ ] Renders base64 PNG images
- [ ] Renders base64 GIF images
- [ ] Download button works
- [ ] Lightbox opens on click
- [ ] Lightbox closes on Escape key
- [ ] Lightbox closes on backdrop click
- [ ] Error state shows helpful message

---

## Common Gotchas

1. **Tool Name Casing**: Tools can be `Bash`, `bash`, or `BASH`. Use case-insensitive matching.
2. **Missing Fields**: Tool inputs may not have all fields. Always use optional chaining.
3. **Large Content**: Tool results can be megabytes. Always truncate with expandable sections.
4. **Base64 Images**: Can be very large. Don't render multiple at once without lazy loading.
5. **Cat -n Format**: Strip line numbers `^\s*\d+[→\t]` before displaying code.
6. **Nested Tool Results**: Tool results can contain arrays of content blocks recursively.

---

## Reference Files

Study these in the old dashboard:
- `web/app/components/ToolUse.tsx` - Tool invocation patterns
- `web/app/components/ToolResult.tsx` - Result detection and rendering
- `web/app/components/ImageContent.tsx` - Image handling
- `web/app/components/CodeDiff.tsx` - Diff algorithm
- `web/app/components/TodoList.tsx` - Todo rendering

---

## Definition of Done

Phase 3 is complete when:

1. All specialized tool renderers work correctly
2. Generic fallback handles unknown tools gracefully
3. Tool results detect content types automatically
4. Image content displays with lightbox and download
5. All components integrate with Phase 2 components
6. Expand/collapse state works consistently
7. No TypeScript errors in strict mode
8. Components export from appropriate index files
9. Commit history shows logical, atomic commits

---

**Build on Phase 2's foundation. Make tools beautiful.**
