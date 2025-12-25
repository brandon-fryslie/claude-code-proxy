import { type FC, useMemo } from 'react'
import { SystemReminder } from './SystemReminder'
import { FunctionDefinitions } from './FunctionDefinitions'

interface TextContentProps {
  text: string
  showSystemReminders?: boolean
}

export const TextContent: FC<TextContentProps> = ({
  text,
  showSystemReminders = false
}) => {
  const { regularContent, systemReminders, functionBlocks } = useMemo(() => {
    return parseTextContent(text)
  }, [text])

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
  )
}

// Parse text to extract special sections
function parseTextContent(text: string): {
  regularContent: string
  systemReminders: string[]
  functionBlocks: string[]
} {
  const systemReminders: string[] = []
  const functionBlocks: string[] = []

  // Extract <system-reminder> tags
  let regularContent = text.replace(
    /<system-reminder>([\s\S]*?)<\/system-reminder>/g,
    (_, content) => {
      systemReminders.push(content.trim())
      return ''
    }
  )

  // Extract <functions> blocks
  regularContent = regularContent.replace(
    /<functions>([\s\S]*?)<\/functions>/g,
    (_, content) => {
      functionBlocks.push(content.trim())
      return ''
    }
  )

  return {
    regularContent: regularContent.trim(),
    systemReminders,
    functionBlocks,
  }
}

// Format text with markdown-like syntax
function formatText(text: string): string {
  let html = escapeHtml(text)

  // Convert markdown-like syntax
  // **bold** -> <strong>
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')

  // *italic* -> <em>
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')

  // `code` -> <code>
  html = html.replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1 rounded text-sm">$1</code>')

  // Line breaks
  html = html.replace(/\n\n/g, '</p><p class="mt-2">')
  html = html.replace(/\n/g, '<br>')

  return `<p>${html}</p>`
}

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}
