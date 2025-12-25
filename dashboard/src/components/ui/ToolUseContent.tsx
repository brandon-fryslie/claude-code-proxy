import { type FC, useState } from 'react'
import { ChevronDown, Terminal, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { CodeViewer } from './CodeViewer'

interface ToolUseContentProps {
  id: string
  name: string
  input: Record<string, unknown>
}

// Special rendering for common tools
const SPECIAL_TOOLS = ['bash', 'read_file', 'write_file', 'edit_file', 'glob', 'grep']

export const ToolUseContent: FC<ToolUseContentProps> = ({ id, name, input }) => {
  const [expanded, setExpanded] = useState(false)
  const [copied, setCopied] = useState(false)

  const copyId = async () => {
    await navigator.clipboard.writeText(id)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const isSpecialTool = SPECIAL_TOOLS.includes(name)

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
  )
}

// Special rendering for common tools
const SpecialToolInput: FC<{ name: string; input: Record<string, unknown> }> = ({ name, input }) => {
  switch (name) {
    case 'bash': {
      return (
        <div className="font-mono text-sm bg-gray-900 text-gray-100 p-3 rounded">
          <span className="text-green-400">$ </span>
          {String(input.command || '')}
        </div>
      )
    }

    case 'read_file': {
      return (
        <div className="text-sm">
          <span className="text-gray-500">Reading: </span>
          <span className="font-mono text-blue-600">{String(input.path || input.file_path || '')}</span>
        </div>
      )
    }

    case 'write_file':
    case 'edit_file': {
      const content = String(input.content || input.new_content || '')
      const path = String(input.path || input.file_path || '')
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
      )
    }

    default:
      return <GenericToolInput input={input} />
  }
}

const GenericToolInput: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const entries = Object.entries(input)

  if (entries.length === 0) {
    return <span className="text-gray-400 text-sm italic">No parameters</span>
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
  )
}

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
      )
    }
    return <span className="font-mono">{value}</span>
  }

  if (typeof value === 'object' && value !== null) {
    return (
      <details className="inline">
        <summary className="cursor-pointer text-blue-600">
          Show object ({Object.keys(value as object).length} properties)
        </summary>
        <pre className="mt-1 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
          {JSON.stringify(value, null, 2)}
        </pre>
      </details>
    )
  }

  return <span className="font-mono">{String(value)}</span>
}

function getLanguageFromPath(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || ''
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
  }
  return languageMap[ext] || 'text'
}
