import { type FC, useState, useMemo, useRef } from 'react'
import { Copy, Check, Download, Maximize2, X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CodeViewerProps {
  code: string
  language?: string
  filename?: string
  maxHeight?: number
  showLineNumbers?: boolean
  showControls?: boolean
}

export const CodeViewer: FC<CodeViewerProps> = ({
  code,
  language: providedLanguage,
  filename,
  maxHeight = 400,
  showLineNumbers = true,
  showControls = true,
}) => {
  const [copied, setCopied] = useState(false)
  const [fullscreen, setFullscreen] = useState(false)
  const codeRef = useRef<HTMLPreElement>(null)

  // Determine language from filename or provided value
  const language = useMemo(() => {
    if (providedLanguage) return providedLanguage
    if (filename) return getLanguageFromFilename(filename)
    return 'text'
  }, [providedLanguage, filename])

  // Apply syntax highlighting
  const highlightedLines = useMemo(() => {
    const lines = code.split('\n')
    return lines.map(line => highlightLine(line, language))
  }, [code, language])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    const blob = new Blob([code], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename || `code.${getExtension(language)}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

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
  )

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
    )
  }

  return codeContent
}

// Syntax highlighting - single pass, no external dependencies
function highlightLine(line: string, language: string): string {
  let result = escapeHtml(line)

  // Don't highlight plain text
  if (language === 'text') return result

  // Order matters! Apply patterns from most to least specific

  // 1. Strings (double quotes, single quotes, backticks)
  result = result.replace(
    /("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`(?:[^`\\]|\\.)*`)/g,
    '<span class="text-amber-300">$1</span>'
  )

  // 2. Comments
  result = result.replace(
    /(\/\/.*$|#.*$)/gm,
    '<span class="text-gray-500 italic">$1</span>'
  )

  // 3. Keywords (language-specific)
  const keywords = getKeywords(language)
  if (keywords.length > 0) {
    const keywordPattern = new RegExp(
      `\\b(${keywords.join('|')})\\b(?![^<]*>)`,
      'g'
    )
    result = result.replace(
      keywordPattern,
      '<span class="text-purple-400">$1</span>'
    )
  }

  // 4. Literals (true, false, null, undefined, None, etc.)
  result = result.replace(
    /\b(true|false|null|undefined|None|True|False|nil)\b(?![^<]*>)/g,
    '<span class="text-orange-400">$1</span>'
  )

  // 5. Numbers
  result = result.replace(
    /\b(\d+\.?\d*)\b(?![^<]*>)/g,
    '<span class="text-cyan-300">$1</span>'
  )

  // 6. PascalCase identifiers (likely types/classes)
  result = result.replace(
    /\b([A-Z][a-zA-Z0-9]*)\b(?![^<]*>)/g,
    '<span class="text-yellow-200">$1</span>'
  )

  return result
}

function getKeywords(language: string): string[] {
  const keywordSets: Record<string, string[]> = {
    javascript: ['const', 'let', 'var', 'function', 'class', 'extends', 'return', 'if', 'else', 'for', 'while', 'switch', 'case', 'break', 'continue', 'import', 'export', 'default', 'from', 'async', 'await', 'try', 'catch', 'throw', 'new', 'this', 'typeof', 'instanceof'],
    typescript: ['const', 'let', 'var', 'function', 'class', 'extends', 'return', 'if', 'else', 'for', 'while', 'switch', 'case', 'break', 'continue', 'import', 'export', 'default', 'from', 'async', 'await', 'try', 'catch', 'throw', 'new', 'this', 'typeof', 'instanceof', 'interface', 'type', 'enum', 'implements', 'private', 'public', 'protected', 'readonly'],
    python: ['def', 'class', 'return', 'if', 'elif', 'else', 'for', 'while', 'try', 'except', 'finally', 'raise', 'import', 'from', 'as', 'with', 'lambda', 'yield', 'assert', 'pass', 'break', 'continue', 'global', 'nonlocal', 'async', 'await'],
    go: ['func', 'package', 'import', 'var', 'const', 'type', 'struct', 'interface', 'map', 'chan', 'go', 'defer', 'return', 'if', 'else', 'for', 'range', 'switch', 'case', 'default', 'break', 'continue', 'fallthrough', 'select'],
    rust: ['fn', 'let', 'mut', 'const', 'struct', 'enum', 'impl', 'trait', 'pub', 'mod', 'use', 'return', 'if', 'else', 'for', 'while', 'loop', 'match', 'break', 'continue', 'async', 'await', 'move', 'ref', 'self', 'Self', 'where'],
    bash: ['if', 'then', 'else', 'elif', 'fi', 'for', 'while', 'do', 'done', 'case', 'esac', 'function', 'return', 'exit', 'export', 'local', 'readonly', 'declare', 'unset', 'source', 'alias'],
  }

  return keywordSets[language] || keywordSets.javascript || []
}

function getLanguageFromFilename(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
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
  }
  return languageMap[ext] || 'text'
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
  }
  return extMap[language] || 'txt'
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}
