# Phase 2: Rich Content Display - Implementation Handoff

**Branch:** `phase-2-rich-content-display`
**Epic:** `brandon-fryslie_claude-code-proxy-kqw` (message-content-parser), `brandon-fryslie_claude-code-proxy-6j7` (code-viewer), `brandon-fryslie_claude-code-proxy-nhw` (copy-to-clipboard)

---

## Executive Summary

You are implementing the **Rich Content Display** system for the new dashboard. This is the foundation for all message rendering - without this, users see raw JSON instead of formatted content. Your work enables:

1. **Message Content Parser** - Transform Anthropic message blocks into React components
2. **Code Viewer with Syntax Highlighting** - Display code with line numbers and language detection
3. **Copy-to-Clipboard** - Universal copy buttons with visual feedback

When complete, users will see beautifully formatted messages instead of raw JSON blobs.

---

## Your Working Environment

### Directory Structure
```
/Users/bmf/code/brandon-fryslie_claude-code-proxy/
├── dashboard/              # NEW dashboard (your target)
│   ├── src/
│   │   ├── components/
│   │   │   ├── layout/     # Sidebar, AppLayout, ResizablePanel
│   │   │   └── ui/         # EMPTY - you'll add components here
│   │   ├── pages/
│   │   │   └── Requests.tsx  # Currently shows raw JSON
│   │   ├── lib/
│   │   │   ├── types.ts    # TypeScript interfaces
│   │   │   ├── api.ts      # React Query hooks
│   │   │   └── utils.ts    # cn() utility
│   │   └── index.css       # Theme variables
├── web/                    # OLD dashboard (reference implementation)
│   └── app/
│       ├── components/
│       │   ├── MessageContent.tsx   # REFERENCE: Message rendering
│       │   ├── CodeViewer.tsx       # REFERENCE: Code display
│       │   ├── ToolUse.tsx          # REFERENCE: Tool rendering
│       │   └── ToolResult.tsx       # REFERENCE: Result rendering
│       └── utils/
│           └── formatters.ts        # REFERENCE: Text utilities
└── proxy/                  # Go backend (API reference)
```

### Tech Stack (Dashboard)
- React 19.2.0
- TypeScript 5.9.3 (strict mode)
- Vite 7.2.4
- Tailwind CSS 4.1.18
- TanStack React Query 5.90.12
- Lucide React 0.562.0 (icons)

### Running the Dashboard
```bash
cd dashboard
pnpm install
pnpm dev  # Starts on http://localhost:5174
```

The backend proxy must be running on port 3001:
```bash
cd proxy && go run cmd/proxy/main.go
```

---

## Topic 1: Message Content Parser

### What You're Building

A component system that transforms Anthropic API message content into React components. Messages can contain:

1. **Text content** - Plain text or markdown-like formatting
2. **Tool use blocks** - Claude invoking tools with parameters
3. **Tool result blocks** - Results returned from tool execution
4. **Image content** - Base64 encoded images
5. **Special XML tags** - `<system-reminder>`, `<functions>` blocks

### Anthropic Message Format

Messages come from the API in this structure:

```typescript
// From dashboard/src/lib/types.ts
interface AnthropicMessage {
  role: 'user' | 'assistant' | 'system';
  content: string | ContentBlock[];
}

// Content can be a string OR an array of blocks:
type ContentBlock =
  | TextBlock
  | ToolUseBlock
  | ToolResultBlock
  | ImageBlock;

interface TextBlock {
  type: 'text';
  text: string;
}

interface ToolUseBlock {
  type: 'tool_use';
  id: string;          // e.g., "toolu_01XYZ..."
  name: string;        // e.g., "bash", "file_editor", "read_file"
  input: Record<string, unknown>;  // Tool-specific parameters
}

interface ToolResultBlock {
  type: 'tool_result';
  tool_use_id: string;  // References the tool_use block
  content: string | ContentBlock[];
  is_error?: boolean;
}

interface ImageBlock {
  type: 'image';
  source: {
    type: 'base64';
    media_type: 'image/jpeg' | 'image/png' | 'image/gif' | 'image/webp';
    data: string;  // Base64 encoded
  };
}
```

### Component Architecture

Create these files in `dashboard/src/components/ui/`:

```
components/ui/
├── MessageContent.tsx      # Main entry point - routes to sub-components
├── TextContent.tsx         # Renders text with formatting
├── ToolUseContent.tsx      # Renders tool invocations
├── ToolResultContent.tsx   # Renders tool results
├── ImageContent.tsx        # Renders base64 images
├── SystemReminder.tsx      # Handles <system-reminder> tags
├── FunctionDefinitions.tsx # Handles <functions> blocks
└── index.ts                # Barrel exports
```

### MessageContent.tsx - Main Component

```tsx
// dashboard/src/components/ui/MessageContent.tsx
import { type FC } from 'react';
import { TextContent } from './TextContent';
import { ToolUseContent } from './ToolUseContent';
import { ToolResultContent } from './ToolResultContent';
import { ImageContent } from './ImageContent';

interface ContentBlock {
  type: string;
  [key: string]: unknown;
}

interface MessageContentProps {
  content: string | ContentBlock[];
  showSystemReminders?: boolean;  // Default: false (hide them)
}

export const MessageContent: FC<MessageContentProps> = ({
  content,
  showSystemReminders = false
}) => {
  // Handle string content (simple case)
  if (typeof content === 'string') {
    return <TextContent text={content} showSystemReminders={showSystemReminders} />;
  }

  // Handle array of content blocks
  if (!Array.isArray(content)) {
    return <pre className="text-xs text-red-500">Unknown content format</pre>;
  }

  return (
    <div className="space-y-3">
      {content.map((block, index) => (
        <ContentBlockRenderer
          key={`${block.type}-${index}`}
          block={block}
          showSystemReminders={showSystemReminders}
        />
      ))}
    </div>
  );
};

const ContentBlockRenderer: FC<{
  block: ContentBlock;
  showSystemReminders: boolean;
}> = ({ block, showSystemReminders }) => {
  switch (block.type) {
    case 'text':
      return (
        <TextContent
          text={block.text as string}
          showSystemReminders={showSystemReminders}
        />
      );

    case 'tool_use':
      return (
        <ToolUseContent
          id={block.id as string}
          name={block.name as string}
          input={block.input as Record<string, unknown>}
        />
      );

    case 'tool_result':
      return (
        <ToolResultContent
          toolUseId={block.tool_use_id as string}
          content={block.content as string | ContentBlock[]}
          isError={block.is_error as boolean | undefined}
        />
      );

    case 'image':
      return <ImageContent source={block.source as ImageSource} />;

    default:
      // Fallback for unknown types - show raw JSON
      return (
        <details className="text-xs">
          <summary className="cursor-pointer text-gray-500">
            Unknown block type: {block.type}
          </summary>
          <pre className="mt-2 p-2 bg-gray-100 rounded overflow-x-auto">
            {JSON.stringify(block, null, 2)}
          </pre>
        </details>
      );
  }
};
```

### TextContent.tsx - Text Rendering

The text content component needs to handle:
1. Basic text display
2. `<system-reminder>` tags (usually hidden)
3. `<functions>` blocks (tool definitions from system prompts)
4. Markdown-like formatting (bold, italic, code)

```tsx
// dashboard/src/components/ui/TextContent.tsx
import { type FC, useMemo } from 'react';
import { SystemReminder } from './SystemReminder';
import { FunctionDefinitions } from './FunctionDefinitions';

interface TextContentProps {
  text: string;
  showSystemReminders?: boolean;
}

export const TextContent: FC<TextContentProps> = ({
  text,
  showSystemReminders = false
}) => {
  const { regularContent, systemReminders, functionBlocks } = useMemo(() => {
    return parseTextContent(text);
  }, [text]);

  return (
    <div className="space-y-2">
      {/* Regular text content */}
      {regularContent && (
        <div
          className="prose prose-sm max-w-none"
          dangerouslySetInnerHTML={{ __html: formatText(regularContent) }}
        />
      )}

      {/* Function definitions (from system prompts) */}
      {functionBlocks.length > 0 && (
        <FunctionDefinitions blocks={functionBlocks} />
      )}

      {/* System reminders (collapsible, usually hidden) */}
      {showSystemReminders && systemReminders.length > 0 && (
        <div className="space-y-2">
          {systemReminders.map((reminder, i) => (
            <SystemReminder key={i} content={reminder} />
          ))}
        </div>
      )}
    </div>
  );
};

// Parse text to extract special sections
function parseTextContent(text: string): {
  regularContent: string;
  systemReminders: string[];
  functionBlocks: string[];
} {
  const systemReminders: string[] = [];
  const functionBlocks: string[] = [];

  // Extract <system-reminder> tags
  let regularContent = text.replace(
    /<system-reminder>([\s\S]*?)<\/system-reminder>/g,
    (_, content) => {
      systemReminders.push(content.trim());
      return '';
    }
  );

  // Extract <functions> blocks
  regularContent = regularContent.replace(
    /<functions>([\s\S]*?)<\/functions>/g,
    (_, content) => {
      functionBlocks.push(content.trim());
      return '';
    }
  );

  return {
    regularContent: regularContent.trim(),
    systemReminders,
    functionBlocks,
  };
}

// Format text with markdown-like syntax
function formatText(text: string): string {
  let html = escapeHtml(text);

  // Convert markdown-like syntax
  // **bold** -> <strong>
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');

  // *italic* -> <em>
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // `code` -> <code>
  html = html.replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1 rounded text-sm">$1</code>');

  // Line breaks
  html = html.replace(/\n\n/g, '</p><p class="mt-2">');
  html = html.replace(/\n/g, '<br>');

  return `<p>${html}</p>`;
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
```

### ToolUseContent.tsx - Tool Invocation Display

```tsx
// dashboard/src/components/ui/ToolUseContent.tsx
import { type FC, useState } from 'react';
import { ChevronDown, Terminal, Copy, Check } from 'lucide-react';
import { cn } from '@/lib/utils';
import { CodeViewer } from './CodeViewer';

interface ToolUseContentProps {
  id: string;
  name: string;
  input: Record<string, unknown>;
}

// Special rendering for common tools
const SPECIAL_TOOLS = ['bash', 'read_file', 'write_file', 'edit_file', 'glob', 'grep'];

export const ToolUseContent: FC<ToolUseContentProps> = ({ id, name, input }) => {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);

  const copyId = async () => {
    await navigator.clipboard.writeText(id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isSpecialTool = SPECIAL_TOOLS.includes(name);

  return (
    <div className="border rounded-lg bg-gradient-to-r from-indigo-50 to-blue-50 overflow-hidden">
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-2 cursor-pointer hover:bg-white/50 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-indigo-600" />
          <span className="font-medium text-indigo-900">{name}</span>
          <button
            onClick={(e) => { e.stopPropagation(); copyId(); }}
            className="text-xs text-gray-400 hover:text-gray-600 flex items-center gap-1"
          >
            {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
            <span className="font-mono">{id.slice(-8)}</span>
          </button>
        </div>
        <ChevronDown
          className={cn(
            "w-4 h-4 text-gray-400 transition-transform",
            expanded && "rotate-180"
          )}
        />
      </div>

      {/* Tool Input */}
      {expanded && (
        <div className="px-4 py-3 border-t bg-white/50">
          {isSpecialTool ? (
            <SpecialToolInput name={name} input={input} />
          ) : (
            <GenericToolInput input={input} />
          )}
        </div>
      )}

      {/* Execution indicator (pulsing dot) */}
      <div className="px-4 py-1 bg-indigo-100/50 text-xs text-indigo-600 flex items-center gap-2">
        <span className="w-2 h-2 bg-indigo-500 rounded-full animate-pulse" />
        Executing tool...
      </div>
    </div>
  );
};

// Special rendering for common tools
const SpecialToolInput: FC<{ name: string; input: Record<string, unknown> }> = ({ name, input }) => {
  switch (name) {
    case 'bash':
      return (
        <div className="font-mono text-sm bg-gray-900 text-gray-100 p-3 rounded">
          <span className="text-green-400">$ </span>
          {String(input.command || '')}
        </div>
      );

    case 'read_file':
      return (
        <div className="text-sm">
          <span className="text-gray-500">Reading: </span>
          <span className="font-mono text-blue-600">{String(input.path || input.file_path || '')}</span>
        </div>
      );

    case 'write_file':
    case 'edit_file':
      const content = String(input.content || input.new_content || '');
      const path = String(input.path || input.file_path || '');
      return (
        <div className="space-y-2">
          <div className="text-sm">
            <span className="text-gray-500">{name === 'write_file' ? 'Writing to' : 'Editing'}: </span>
            <span className="font-mono text-blue-600">{path}</span>
          </div>
          {content && (
            <CodeViewer
              code={content}
              language={getLanguageFromPath(path)}
              maxHeight={200}
            />
          )}
        </div>
      );

    default:
      return <GenericToolInput input={input} />;
  }
};

const GenericToolInput: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const entries = Object.entries(input);

  if (entries.length === 0) {
    return <span className="text-gray-400 text-sm italic">No parameters</span>;
  }

  return (
    <div className="space-y-2">
      {entries.map(([key, value]) => (
        <div key={key} className="text-sm">
          <span className="text-gray-500">{key}: </span>
          <ToolInputValue value={value} />
        </div>
      ))}
    </div>
  );
};

const ToolInputValue: FC<{ value: unknown }> = ({ value }) => {
  if (typeof value === 'string') {
    // Truncate long strings
    if (value.length > 200 || value.includes('\n')) {
      return (
        <details className="inline">
          <summary className="cursor-pointer text-blue-600">
            Show content ({value.length} chars)
          </summary>
          <pre className="mt-1 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
            {value}
          </pre>
        </details>
      );
    }
    return <span className="font-mono">{value}</span>;
  }

  if (typeof value === 'object') {
    return (
      <details className="inline">
        <summary className="cursor-pointer text-blue-600">
          Show object ({Object.keys(value as object).length} properties)
        </summary>
        <pre className="mt-1 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
          {JSON.stringify(value, null, 2)}
        </pre>
      </details>
    );
  }

  return <span className="font-mono">{String(value)}</span>;
};

function getLanguageFromPath(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || '';
  const languageMap: Record<string, string> = {
    ts: 'typescript', tsx: 'typescript',
    js: 'javascript', jsx: 'javascript',
    py: 'python',
    go: 'go',
    rs: 'rust',
    md: 'markdown',
    json: 'json',
    yaml: 'yaml', yml: 'yaml',
    sh: 'bash', bash: 'bash',
    css: 'css', scss: 'css',
    html: 'html',
    sql: 'sql',
  };
  return languageMap[ext] || 'text';
}
```

### ToolResultContent.tsx - Tool Result Display

```tsx
// dashboard/src/components/ui/ToolResultContent.tsx
import { type FC, useState, useMemo } from 'react';
import { CheckCircle, XCircle, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { CodeViewer } from './CodeViewer';
import { MessageContent } from './MessageContent';

interface ToolResultContentProps {
  toolUseId: string;
  content: string | ContentBlock[];
  isError?: boolean;
}

interface ContentBlock {
  type: string;
  [key: string]: unknown;
}

export const ToolResultContent: FC<ToolResultContentProps> = ({
  toolUseId,
  content,
  isError = false
}) => {
  const [expanded, setExpanded] = useState(!isError); // Collapse errors by default

  // Detect content type
  const { contentType, processedContent } = useMemo(() => {
    if (typeof content !== 'string') {
      return { contentType: 'blocks' as const, processedContent: content };
    }

    // Detect code (cat -n format with line numbers)
    if (/^\s*\d+[→\t]/.test(content)) {
      return {
        contentType: 'code' as const,
        processedContent: extractCodeFromCatN(content)
      };
    }

    // Detect JSON
    if (content.trim().startsWith('{') || content.trim().startsWith('[')) {
      try {
        JSON.parse(content);
        return { contentType: 'json' as const, processedContent: content };
      } catch {
        // Not valid JSON
      }
    }

    // Detect code by keywords
    if (hasCodeIndicators(content)) {
      return { contentType: 'code' as const, processedContent: content };
    }

    return { contentType: 'text' as const, processedContent: content };
  }, [content]);

  const bgColor = isError
    ? 'bg-gradient-to-r from-red-50 to-rose-50'
    : 'bg-gradient-to-r from-emerald-50 to-green-50';

  const borderColor = isError ? 'border-red-200' : 'border-emerald-200';

  return (
    <div className={cn("border rounded-lg overflow-hidden", borderColor, bgColor)}>
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-2 cursor-pointer hover:bg-white/50 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          {isError ? (
            <XCircle className="w-4 h-4 text-red-500" />
          ) : (
            <CheckCircle className="w-4 h-4 text-emerald-500" />
          )}
          <span className={cn(
            "font-medium",
            isError ? "text-red-700" : "text-emerald-700"
          )}>
            {isError ? 'Error' : 'Result'}
          </span>
          <span className="text-xs text-gray-400 font-mono">
            {toolUseId.slice(-8)}
          </span>
          <span className="text-xs text-gray-400 px-2 py-0.5 bg-white/50 rounded">
            {contentType}
          </span>
        </div>
        <ChevronDown
          className={cn(
            "w-4 h-4 text-gray-400 transition-transform",
            expanded && "rotate-180"
          )}
        />
      </div>

      {/* Content */}
      {expanded && (
        <div className="px-4 py-3 border-t bg-white/50">
          <ToolResultBody
            contentType={contentType}
            content={processedContent}
            isError={isError}
          />
        </div>
      )}
    </div>
  );
};

const ToolResultBody: FC<{
  contentType: 'text' | 'code' | 'json' | 'blocks';
  content: string | ContentBlock[];
  isError: boolean;
}> = ({ contentType, content, isError }) => {
  if (contentType === 'blocks') {
    return <MessageContent content={content as ContentBlock[]} />;
  }

  const text = content as string;

  // Truncate very long content
  const MAX_LENGTH = 500;
  const [showFull, setShowFull] = useState(text.length <= MAX_LENGTH);
  const displayText = showFull ? text : text.slice(0, MAX_LENGTH);

  switch (contentType) {
    case 'code':
      return (
        <div>
          <CodeViewer code={displayText} language="text" maxHeight={300} />
          {!showFull && (
            <button
              onClick={() => setShowFull(true)}
              className="mt-2 text-sm text-blue-600 hover:underline"
            >
              Show full content ({text.length} chars)
            </button>
          )}
        </div>
      );

    case 'json':
      return (
        <pre className="p-3 bg-gray-900 text-gray-100 rounded text-xs overflow-x-auto">
          {JSON.stringify(JSON.parse(text), null, 2)}
        </pre>
      );

    default:
      return (
        <div className={cn(
          "text-sm whitespace-pre-wrap",
          isError && "text-red-600"
        )}>
          {displayText}
          {!showFull && (
            <>
              <span className="text-gray-400">...</span>
              <button
                onClick={() => setShowFull(true)}
                className="ml-2 text-blue-600 hover:underline"
              >
                Show more
              </button>
            </>
          )}
        </div>
      );
  }
};

// Extract code from cat -n format (line numbers with arrow or tab)
function extractCodeFromCatN(text: string): string {
  return text
    .split('\n')
    .map(line => {
      // Match: "   123→content" or "   123\tcontent"
      const match = line.match(/^\s*\d+[→\t](.*)$/);
      return match ? match[1] : line;
    })
    .join('\n');
}

// Detect if content looks like code
function hasCodeIndicators(text: string): boolean {
  const codePatterns = [
    /^(import|from|const|let|var|function|class|def|func|package)\s/m,
    /[{}\[\]];?\s*$/m,
    /^\s*(if|for|while|return|throw)\s*\(/m,
    /=>\s*{/,
    /\bexport\s+(default\s+)?/,
  ];
  return codePatterns.some(pattern => pattern.test(text));
}
```

---

## Topic 2: Code Viewer with Syntax Highlighting

### What You're Building

A syntax-highlighted code viewer with:
- Custom single-pass syntax highlighting (no external library)
- Line numbers
- Language detection from filename
- Copy button
- Optional fullscreen mode
- Download button

### CodeViewer.tsx

```tsx
// dashboard/src/components/ui/CodeViewer.tsx
import { type FC, useState, useMemo, useRef } from 'react';
import { Copy, Check, Download, Maximize2, X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface CodeViewerProps {
  code: string;
  language?: string;
  filename?: string;
  maxHeight?: number;
  showLineNumbers?: boolean;
  showControls?: boolean;
}

export const CodeViewer: FC<CodeViewerProps> = ({
  code,
  language: providedLanguage,
  filename,
  maxHeight = 400,
  showLineNumbers = true,
  showControls = true,
}) => {
  const [copied, setCopied] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const codeRef = useRef<HTMLPreElement>(null);

  // Determine language from filename or provided value
  const language = useMemo(() => {
    if (providedLanguage) return providedLanguage;
    if (filename) return getLanguageFromFilename(filename);
    return 'text';
  }, [providedLanguage, filename]);

  // Apply syntax highlighting
  const highlightedLines = useMemo(() => {
    const lines = code.split('\n');
    return lines.map(line => highlightLine(line, language));
  }, [code, language]);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDownload = () => {
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename || `code.${getExtension(language)}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const codeContent = (
    <div className={cn(
      "relative bg-gray-900 rounded-lg overflow-hidden",
      fullscreen && "fixed inset-4 z-50 flex flex-col"
    )}>
      {/* Header with controls */}
      {showControls && (
        <div className="flex items-center justify-between px-4 py-2 bg-gray-800 text-gray-400 text-xs">
          <div className="flex items-center gap-2">
            {filename && <span className="font-mono">{filename}</span>}
            <span className="px-2 py-0.5 bg-gray-700 rounded">{language}</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleCopy}
              className="p-1.5 hover:bg-gray-700 rounded transition-colors"
              title="Copy to clipboard"
            >
              {copied ? (
                <Check className="w-4 h-4 text-green-400" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
            </button>
            <button
              onClick={handleDownload}
              className="p-1.5 hover:bg-gray-700 rounded transition-colors"
              title="Download file"
            >
              <Download className="w-4 h-4" />
            </button>
            <button
              onClick={() => setFullscreen(!fullscreen)}
              className="p-1.5 hover:bg-gray-700 rounded transition-colors"
              title={fullscreen ? "Exit fullscreen" : "Fullscreen"}
            >
              {fullscreen ? (
                <X className="w-4 h-4" />
              ) : (
                <Maximize2 className="w-4 h-4" />
              )}
            </button>
          </div>
        </div>
      )}

      {/* Code content */}
      <pre
        ref={codeRef}
        className={cn(
          "overflow-auto text-sm leading-relaxed",
          fullscreen ? "flex-1" : ""
        )}
        style={{ maxHeight: fullscreen ? undefined : maxHeight }}
      >
        <table className="w-full border-collapse">
          <tbody>
            {highlightedLines.map((html, i) => (
              <tr key={i} className="hover:bg-gray-800/50">
                {showLineNumbers && (
                  <td className="select-none text-right pr-4 pl-4 text-gray-500 border-r border-gray-700 align-top">
                    {i + 1}
                  </td>
                )}
                <td
                  className="pl-4 pr-4 text-gray-100"
                  dangerouslySetInnerHTML={{ __html: html || '&nbsp;' }}
                />
              </tr>
            ))}
          </tbody>
        </table>
      </pre>
    </div>
  );

  // Fullscreen backdrop
  if (fullscreen) {
    return (
      <>
        <div
          className="fixed inset-0 bg-black/80 z-40"
          onClick={() => setFullscreen(false)}
        />
        {codeContent}
      </>
    );
  }

  return codeContent;
};

// Syntax highlighting - single pass, no external dependencies
function highlightLine(line: string, language: string): string {
  let result = escapeHtml(line);

  // Don't highlight plain text
  if (language === 'text') return result;

  // Order matters! Apply patterns from most to least specific

  // 1. Strings (double quotes, single quotes, backticks)
  result = result.replace(
    /("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`(?:[^`\\]|\\.)*`)/g,
    '<span class="text-amber-300">$1</span>'
  );

  // 2. Comments
  result = result.replace(
    /(\/\/.*$|#.*$)/gm,
    '<span class="text-gray-500 italic">$1</span>'
  );

  // 3. Keywords (language-specific)
  const keywords = getKeywords(language);
  if (keywords.length > 0) {
    const keywordPattern = new RegExp(
      `\\b(${keywords.join('|')})\\b(?![^<]*>)`,
      'g'
    );
    result = result.replace(
      keywordPattern,
      '<span class="text-purple-400">$1</span>'
    );
  }

  // 4. Literals (true, false, null, undefined, None, etc.)
  result = result.replace(
    /\b(true|false|null|undefined|None|True|False|nil)\b(?![^<]*>)/g,
    '<span class="text-orange-400">$1</span>'
  );

  // 5. Numbers
  result = result.replace(
    /\b(\d+\.?\d*)\b(?![^<]*>)/g,
    '<span class="text-cyan-300">$1</span>'
  );

  // 6. PascalCase identifiers (likely types/classes)
  result = result.replace(
    /\b([A-Z][a-zA-Z0-9]*)\b(?![^<]*>)/g,
    '<span class="text-yellow-200">$1</span>'
  );

  return result;
}

function getKeywords(language: string): string[] {
  const keywordSets: Record<string, string[]> = {
    javascript: ['const', 'let', 'var', 'function', 'class', 'extends', 'return', 'if', 'else', 'for', 'while', 'switch', 'case', 'break', 'continue', 'import', 'export', 'default', 'from', 'async', 'await', 'try', 'catch', 'throw', 'new', 'this', 'typeof', 'instanceof'],
    typescript: ['const', 'let', 'var', 'function', 'class', 'extends', 'return', 'if', 'else', 'for', 'while', 'switch', 'case', 'break', 'continue', 'import', 'export', 'default', 'from', 'async', 'await', 'try', 'catch', 'throw', 'new', 'this', 'typeof', 'instanceof', 'interface', 'type', 'enum', 'implements', 'private', 'public', 'protected', 'readonly'],
    python: ['def', 'class', 'return', 'if', 'elif', 'else', 'for', 'while', 'try', 'except', 'finally', 'raise', 'import', 'from', 'as', 'with', 'lambda', 'yield', 'assert', 'pass', 'break', 'continue', 'global', 'nonlocal', 'async', 'await'],
    go: ['func', 'package', 'import', 'var', 'const', 'type', 'struct', 'interface', 'map', 'chan', 'go', 'defer', 'return', 'if', 'else', 'for', 'range', 'switch', 'case', 'default', 'break', 'continue', 'fallthrough', 'select'],
    rust: ['fn', 'let', 'mut', 'const', 'struct', 'enum', 'impl', 'trait', 'pub', 'mod', 'use', 'return', 'if', 'else', 'for', 'while', 'loop', 'match', 'break', 'continue', 'async', 'await', 'move', 'ref', 'self', 'Self', 'where'],
    bash: ['if', 'then', 'else', 'elif', 'fi', 'for', 'while', 'do', 'done', 'case', 'esac', 'function', 'return', 'exit', 'export', 'local', 'readonly', 'declare', 'unset', 'source', 'alias'],
  };

  return keywordSets[language] || keywordSets.javascript || [];
}

function getLanguageFromFilename(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  const languageMap: Record<string, string> = {
    ts: 'typescript', tsx: 'typescript', mts: 'typescript', cts: 'typescript',
    js: 'javascript', jsx: 'javascript', mjs: 'javascript', cjs: 'javascript',
    py: 'python', pyw: 'python',
    go: 'go',
    rs: 'rust',
    md: 'markdown', mdx: 'markdown',
    json: 'json', jsonc: 'json',
    yaml: 'yaml', yml: 'yaml',
    sh: 'bash', bash: 'bash', zsh: 'bash',
    css: 'css', scss: 'css', sass: 'css', less: 'css',
    html: 'html', htm: 'html',
    sql: 'sql',
    dockerfile: 'docker',
    makefile: 'make',
    toml: 'toml',
    xml: 'xml',
  };
  return languageMap[ext] || 'text';
}

function getExtension(language: string): string {
  const extMap: Record<string, string> = {
    typescript: 'ts',
    javascript: 'js',
    python: 'py',
    go: 'go',
    rust: 'rs',
    bash: 'sh',
    json: 'json',
    yaml: 'yaml',
    markdown: 'md',
    html: 'html',
    css: 'css',
    sql: 'sql',
  };
  return extMap[language] || 'txt';
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}
```

---

## Topic 3: Copy-to-Clipboard

### What You're Building

A reusable hook and button component for copying content with visual feedback.

### useCopyToClipboard.ts

```tsx
// dashboard/src/lib/hooks/useCopyToClipboard.ts
import { useState, useCallback } from 'react';

interface UseCopyToClipboardReturn {
  copied: boolean;
  copy: (text: string) => Promise<void>;
  reset: () => void;
}

export function useCopyToClipboard(resetDelay = 2000): UseCopyToClipboardReturn {
  const [copied, setCopied] = useState(false);

  const copy = useCallback(async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), resetDelay);
    } catch (error) {
      console.error('Failed to copy:', error);
      // Fallback for older browsers
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand('copy');
        setCopied(true);
        setTimeout(() => setCopied(false), resetDelay);
      } finally {
        document.body.removeChild(textarea);
      }
    }
  }, [resetDelay]);

  const reset = useCallback(() => setCopied(false), []);

  return { copied, copy, reset };
}
```

### CopyButton.tsx

```tsx
// dashboard/src/components/ui/CopyButton.tsx
import { type FC } from 'react';
import { Copy, Check } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useCopyToClipboard } from '@/lib/hooks/useCopyToClipboard';

interface CopyButtonProps {
  content: string;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'ghost' | 'outline';
  label?: string;
}

export const CopyButton: FC<CopyButtonProps> = ({
  content,
  className,
  size = 'sm',
  variant = 'ghost',
  label,
}) => {
  const { copied, copy } = useCopyToClipboard();

  const sizeClasses = {
    sm: 'p-1.5',
    md: 'p-2',
    lg: 'p-2.5',
  };

  const iconSizes = {
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  const variantClasses = {
    default: 'bg-gray-200 hover:bg-gray-300 text-gray-700',
    ghost: 'hover:bg-gray-100 text-gray-500 hover:text-gray-700',
    outline: 'border border-gray-300 hover:bg-gray-50 text-gray-600',
  };

  return (
    <button
      onClick={() => copy(content)}
      className={cn(
        'inline-flex items-center gap-1.5 rounded transition-colors',
        sizeClasses[size],
        variantClasses[variant],
        className
      )}
      title={copied ? 'Copied!' : 'Copy to clipboard'}
    >
      {copied ? (
        <Check className={cn(iconSizes[size], 'text-green-500')} />
      ) : (
        <Copy className={iconSizes[size]} />
      )}
      {label && <span className="text-xs">{label}</span>}
    </button>
  );
};
```

---

## Integration Points

### Where to Use These Components

1. **Requests.tsx** - Replace raw JSON display with MessageContent
2. **ConversationDetail** - Use MessageContent for conversation messages
3. **Request Headers** - Use CopyButton for copying header values
4. **Request/Response JSON** - Use CopyButton for full JSON copy

### Example Integration in Requests.tsx

```tsx
// In the request detail panel, replace:
<pre>{JSON.stringify(request.body, null, 2)}</pre>

// With:
<MessageContent
  content={request.body.messages || []}
  showSystemReminders={false}
/>
```

---

## Testing Checklist

### Message Content Parser
- [ ] Renders plain text content
- [ ] Renders array of text blocks
- [ ] Renders tool_use blocks with all parameter types
- [ ] Renders tool_result blocks (success and error)
- [ ] Renders image blocks (base64)
- [ ] Extracts and hides system-reminder tags
- [ ] Extracts and displays function blocks
- [ ] Falls back gracefully for unknown block types
- [ ] Handles nested content in tool results

### Code Viewer
- [ ] Shows line numbers correctly
- [ ] Highlights syntax for JavaScript/TypeScript
- [ ] Highlights syntax for Python
- [ ] Highlights syntax for Go
- [ ] Copy button works and shows feedback
- [ ] Download button triggers file download
- [ ] Fullscreen mode works
- [ ] Language detection from filename works
- [ ] Handles very long files (performance)

### Copy to Clipboard
- [ ] Copy works in modern browsers
- [ ] Fallback works for older browsers
- [ ] Visual feedback shows for 2 seconds
- [ ] Multiple copy buttons track state independently

---

## Common Gotchas

1. **XSS Prevention**: Always escape HTML before inserting into DOM with dangerouslySetInnerHTML
2. **Content Type Detection**: The order of checks matters - check most specific patterns first
3. **Performance**: Memoize heavy computations (syntax highlighting, content parsing)
4. **Keyboard Accessibility**: Copy buttons should be focusable and work with Enter key
5. **Cat -n Format**: Tool results often contain line numbers like "   1→code" - strip them before display

---

## Reference Files (in /web directory)

Study these files in the old dashboard for patterns:
- `web/app/components/MessageContent.tsx` - Content rendering logic
- `web/app/components/CodeViewer.tsx` - Syntax highlighting approach
- `web/app/components/ToolUse.tsx` - Tool invocation display
- `web/app/components/ToolResult.tsx` - Result rendering with type detection
- `web/app/utils/formatters.ts` - Text formatting utilities

---

## Definition of Done

Phase 2 is complete when:

1. All three topics are implemented and tested
2. `MessageContent` correctly renders all Anthropic message types
3. `CodeViewer` displays code with syntax highlighting and controls
4. `CopyButton` provides consistent copy-to-clipboard across the app
5. `Requests.tsx` uses these components instead of raw JSON
6. No TypeScript errors in strict mode
7. All components export from `components/ui/index.ts`
8. Commit history shows logical, atomic commits

---

**Good luck, island agent. You've got this.**
