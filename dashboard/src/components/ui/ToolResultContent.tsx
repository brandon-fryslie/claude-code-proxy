import { type FC, useState, useMemo } from 'react'
import { CheckCircle, XCircle, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { CodeViewer } from './CodeViewer'

interface ToolResultContentProps {
  toolUseId: string
  content: string | ContentBlock[]
  isError?: boolean
}

interface ContentBlock {
  type: string
  [key: string]: unknown
}

export const ToolResultContent: FC<ToolResultContentProps> = ({
  toolUseId,
  content,
  isError = false
}) => {
  const [expanded, setExpanded] = useState(!isError) // Collapse errors by default

  // Detect content type
  const { contentType, processedContent } = useMemo(() => {
    if (typeof content !== 'string') {
      return { contentType: 'blocks' as const, processedContent: content }
    }

    // Detect code (cat -n format with line numbers)
    if (/^\s*\d+[→\t]/.test(content)) {
      return {
        contentType: 'code' as const,
        processedContent: extractCodeFromCatN(content)
      }
    }

    // Detect JSON
    if (content.trim().startsWith('{') || content.trim().startsWith('[')) {
      try {
        JSON.parse(content)
        return { contentType: 'json' as const, processedContent: content }
      } catch {
        // Not valid JSON
      }
    }

    // Detect code by keywords
    if (hasCodeIndicators(content)) {
      return { contentType: 'code' as const, processedContent: content }
    }

    return { contentType: 'text' as const, processedContent: content }
  }, [content])

  const bgColor = isError
    ? 'bg-gradient-to-r from-red-50 to-rose-50'
    : 'bg-gradient-to-r from-emerald-50 to-green-50'

  const borderColor = isError ? 'border-red-200' : 'border-emerald-200'

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
  )
}

const ToolResultBody: FC<{
  contentType: 'text' | 'code' | 'json' | 'blocks'
  content: string | ContentBlock[]
  isError: boolean
}> = ({ contentType, content, isError }) => {
  const MAX_LENGTH = 500
  const [showFull, setShowFull] = useState(false)

  if (contentType === 'blocks') {
    // Import MessageContent lazily to avoid circular dependency
    // For now, render as JSON
    return (
      <pre className="p-3 bg-gray-900 text-gray-100 rounded text-xs overflow-x-auto">
        {JSON.stringify(content, null, 2)}
      </pre>
    )
  }

  const text = content as string
  const displayText = (showFull || text.length <= MAX_LENGTH) ? text : text.slice(0, MAX_LENGTH)

  switch (contentType) {
    case 'code':
      return (
        <div>
          <CodeViewer code={displayText} language="text" maxHeight={300} />
          {!showFull && text.length > MAX_LENGTH && (
            <button
              onClick={() => setShowFull(true)}
              className="mt-2 text-sm text-blue-600 hover:underline"
            >
              Show full content ({text.length} chars)
            </button>
          )}
        </div>
      )

    case 'json':
      return (
        <pre className="p-3 bg-gray-900 text-gray-100 rounded text-xs overflow-x-auto">
          {JSON.stringify(JSON.parse(text), null, 2)}
        </pre>
      )

    default:
      return (
        <div className={cn(
          "text-sm whitespace-pre-wrap",
          isError && "text-red-600"
        )}>
          {displayText}
          {!showFull && text.length > MAX_LENGTH && (
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
      )
  }
}

// Extract code from cat -n format (line numbers with arrow or tab)
function extractCodeFromCatN(text: string): string {
  return text
    .split('\n')
    .map(line => {
      // Match: "   123→content" or "   123\tcontent"
      const match = line.match(/^\s*\d+[→\t](.*)$/)
      return match ? match[1] : line
    })
    .join('\n')
}

// Detect if content looks like code
function hasCodeIndicators(text: string): boolean {
  const codePatterns = [
    /^(import|from|const|let|var|function|class|def|func|package)\s/m,
    /[{}[\]];?\s*$/m,
    /^\s*(if|for|while|return|throw)\s*\(/m,
    /=>\s*{/,
    /\bexport\s+(default\s+)?/,
  ]
  return codePatterns.some(pattern => pattern.test(text))
}
