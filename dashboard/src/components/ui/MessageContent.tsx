import { type FC } from 'react'
import { TextContent } from './TextContent'
import { ToolUseContent } from './ToolUseContent'
import { ToolResultContent } from './ToolResultContent'
import { ImageContent, type ImageSource } from './ImageContent'

interface ContentBlock {
  type: string
  [key: string]: unknown
}

interface MessageContentProps {
  content: string | ContentBlock[]
  showSystemReminders?: boolean  // Default: false (hide them)
}

export const MessageContent: FC<MessageContentProps> = ({
  content,
  showSystemReminders = false
}) => {
  // Handle string content (simple case)
  if (typeof content === 'string') {
    return <TextContent text={content} showSystemReminders={showSystemReminders} />
  }

  // Handle array of content blocks
  if (!Array.isArray(content)) {
    return <pre className="text-xs text-red-500">Unknown content format</pre>
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
  )
}

const ContentBlockRenderer: FC<{
  block: ContentBlock
  showSystemReminders: boolean
}> = ({ block, showSystemReminders }) => {
  switch (block.type) {
    case 'text':
      return (
        <TextContent
          text={block.text as string}
          showSystemReminders={showSystemReminders}
        />
      )

    case 'tool_use':
      return (
        <ToolUseContent
          id={block.id as string}
          name={block.name as string}
          input={block.input as Record<string, unknown>}
        />
      )

    case 'tool_result':
      return (
        <ToolResultContent
          toolUseId={block.tool_use_id as string}
          content={block.content as string | ContentBlock[]}
          isError={block.is_error as boolean | undefined}
        />
      )

    case 'image':
      return <ImageContent source={block.source as ImageSource} />

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
      )
  }
}
