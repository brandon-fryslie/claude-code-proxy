import { type FC } from 'react';
import { FileText, Code, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { formatLargeText } from '@/lib/formatters';
import { ToolUseContainer, ToolResultContent } from './tools';
import { ImageContent } from './ImageContent';

// Content block types from Anthropic API
interface TextContentBlock {
  type: 'text';
  text: string;
}

interface ToolUseContentBlock {
  type: 'tool_use';
  id: string;
  name: string;
  input: Record<string, unknown>;
}

interface ToolResultContentBlock {
  type: 'tool_result';
  tool_use_id?: string;
  content?: string | ContentBlock[];
  is_error?: boolean;
}

interface ImageContentBlock {
  type: 'image';
  source: {
    type: 'base64';
    media_type: 'image/jpeg' | 'image/png' | 'image/gif' | 'image/webp';
    data: string;
  };
}

type ContentBlock =
  | TextContentBlock
  | ToolUseContentBlock
  | ToolResultContentBlock
  | ImageContentBlock;

interface MessageContentProps {
  // Accept flexible content types - we'll handle the type checking internally
  content: string | ContentBlock | ContentBlock[] | unknown;
  className?: string;
}

export const MessageContent: FC<MessageContentProps> = ({ content, className }) => {
  // Handle string content
  if (typeof content === 'string') {
    return <TextContent text={content} className={className} />;
  }

  // Handle array of content blocks
  if (Array.isArray(content)) {
    return (
      <div className={cn('space-y-3', className)}>
        {content.map((block, index) => (
          <ContentBlockRenderer key={index} block={block} />
        ))}
      </div>
    );
  }

  // Handle single content block object
  if (content && typeof content === 'object') {
    return <ContentBlockRenderer block={content as ContentBlock | Record<string, unknown>} className={className} />;
  }

  // Fallback for unknown content
  return (
    <div className={cn('text-sm text-gray-500 italic', className)}>
      No content to display
    </div>
  );
};

// Render individual content blocks based on type
const ContentBlockRenderer: FC<{
  block: ContentBlock | Record<string, unknown>;
  className?: string;
}> = ({ block, className }) => {
  // Type guard for content blocks
  if (!block || typeof block !== 'object' || !('type' in block)) {
    return (
      <UnknownContentBlock content={block} className={className} />
    );
  }

  switch (block.type) {
    case 'text':
      return (
        <TextContent
          text={(block as TextContentBlock).text}
          className={className}
        />
      );

    case 'tool_use':
      const toolUse = block as ToolUseContentBlock;
      return (
        <ToolUseContainer
          id={toolUse.id || 'unknown'}
          name={toolUse.name || 'Unknown Tool'}
          input={toolUse.input || {}}
          isExecuting={false}
          defaultExpanded={false}
        />
      );

    case 'tool_result':
      const toolResult = block as ToolResultContentBlock;
      return (
        <ToolResultBlock
          toolUseId={toolResult.tool_use_id}
          content={toolResult.content}
          isError={toolResult.is_error}
          className={className}
        />
      );

    case 'image':
      const image = block as ImageContentBlock;
      if (image.source && image.source.data) {
        return <ImageContent source={image.source} />;
      }
      return (
        <div className="flex items-center gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-amber-700">
          <AlertCircle className="w-4 h-4" />
          <span className="text-sm">Image content missing source data</span>
        </div>
      );

    default:
      return <UnknownContentBlock content={block} className={className} />;
  }
};

// Text content with formatting
const TextContent: FC<{ text: string; className?: string }> = ({
  text,
  className,
}) => {
  // Check if content contains system reminders - show collapsed
  if (text.includes('<system-reminder>')) {
    return <SystemReminderContent text={text} className={className} />;
  }

  // Check if content contains tool definitions
  if (text.includes('<functions>')) {
    return <ToolDefinitionsContent text={text} className={className} />;
  }

  return (
    <div
      className={cn(
        'text-sm text-gray-700 bg-white rounded-lg p-4 border border-gray-200 leading-relaxed',
        className
      )}
      dangerouslySetInnerHTML={{ __html: formatLargeText(text) }}
    />
  );
};

// Tool result content
const ToolResultBlock: FC<{
  toolUseId?: string;
  content?: string | ContentBlock[];
  isError?: boolean;
  className?: string;
}> = ({ toolUseId, content, isError = false, className }) => {
  // Handle string content
  if (typeof content === 'string') {
    return (
      <div className={cn('rounded-lg border overflow-hidden', className)}>
        {toolUseId && (
          <div className="px-3 py-1.5 bg-gray-100 border-b text-xs text-gray-500 font-mono">
            Result for: {toolUseId.slice(-8)}
          </div>
        )}
        <ToolResultContent content={content} isError={isError} />
      </div>
    );
  }

  // Handle array of content blocks (nested)
  if (Array.isArray(content)) {
    return (
      <div className={cn('space-y-2', className)}>
        {toolUseId && (
          <div className="text-xs text-gray-500 font-mono">
            Result for: {toolUseId.slice(-8)}
          </div>
        )}
        <div className="space-y-2">
          {content.map((block, index) => (
            <ContentBlockRenderer key={index} block={block} />
          ))}
        </div>
      </div>
    );
  }

  // Handle empty content
  if (!content) {
    return (
      <div className={cn('text-sm text-gray-500 italic p-3 bg-gray-50 rounded', className)}>
        No result content
      </div>
    );
  }

  // Fallback - try to stringify
  return (
    <div className={cn('text-sm font-mono p-3 bg-gray-50 rounded', className)}>
      {JSON.stringify(content, null, 2)}
    </div>
  );
};

// System reminder content - collapsed by default
const SystemReminderContent: FC<{ text: string; className?: string }> = ({
  text,
  className,
}) => {
  // Split out system reminders
  const parts: Array<{ type: 'text' | 'reminder'; content: string }> = [];
  const reminderRegex = /<system-reminder>([\s\S]*?)<\/system-reminder>/g;
  let lastIndex = 0;
  let match;

  while ((match = reminderRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      const textPart = text.substring(lastIndex, match.index).trim();
      if (textPart) {
        parts.push({ type: 'text', content: textPart });
      }
    }
    parts.push({ type: 'reminder', content: match[1].trim() });
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    const textPart = text.substring(lastIndex).trim();
    if (textPart) {
      parts.push({ type: 'text', content: textPart });
    }
  }

  const reminderCount = parts.filter((p) => p.type === 'reminder').length;

  return (
    <div className={cn('space-y-3', className)}>
      {/* Regular text parts */}
      {parts
        .filter((p) => p.type === 'text')
        .map((part, index) => (
          <div
            key={`text-${index}`}
            className="text-sm text-gray-700 bg-white rounded-lg p-4 border border-gray-200 leading-relaxed"
            dangerouslySetInnerHTML={{ __html: formatLargeText(part.content) }}
          />
        ))}

      {/* System reminders - collapsed */}
      {reminderCount > 0 && (
        <details className="bg-gray-50 border border-gray-200 rounded-lg">
          <summary className="px-4 py-2 cursor-pointer text-sm text-gray-600 hover:text-gray-800 flex items-center gap-2">
            <AlertCircle className="w-4 h-4" />
            <span>
              {reminderCount} system reminder{reminderCount > 1 ? 's' : ''}
            </span>
          </summary>
          <div className="px-4 py-2 space-y-2 border-t border-gray-200">
            {parts
              .filter((p) => p.type === 'reminder')
              .map((part, index) => (
                <pre
                  key={`reminder-${index}`}
                  className="text-xs text-gray-600 font-mono whitespace-pre-wrap"
                >
                  {part.content}
                </pre>
              ))}
          </div>
        </details>
      )}
    </div>
  );
};

// Tool definitions content - collapsed by default
const ToolDefinitionsContent: FC<{ text: string; className?: string }> = ({
  text,
  className,
}) => {
  const functionsMatch = text.match(/<functions>([\s\S]*?)<\/functions>/);

  if (!functionsMatch) {
    return <TextContent text={text} className={className} />;
  }

  const beforeFunctions = text.substring(0, functionsMatch.index!);
  const afterFunctions = text.substring(
    functionsMatch.index! + functionsMatch[0].length
  );

  // Count function definitions
  const functionCount = (functionsMatch[1].match(/<function>/g) || []).length;

  return (
    <div className={cn('space-y-3', className)}>
      {beforeFunctions && (
        <div
          className="text-sm text-gray-700 bg-white rounded-lg p-4 border border-gray-200 leading-relaxed max-h-64 overflow-y-auto"
          dangerouslySetInnerHTML={{ __html: formatLargeText(beforeFunctions) }}
        />
      )}

      <details className="bg-emerald-50 border border-emerald-200 rounded-lg">
        <summary className="px-4 py-2 cursor-pointer text-sm text-emerald-700 hover:text-emerald-900 flex items-center gap-2">
          <Code className="w-4 h-4" />
          <span className="font-medium">
            {functionCount} available tools
          </span>
        </summary>
        <div className="px-4 py-2 border-t border-emerald-200 max-h-96 overflow-y-auto">
          <pre className="text-xs text-gray-700 font-mono whitespace-pre-wrap">
            {functionsMatch[1]}
          </pre>
        </div>
      </details>

      {afterFunctions && (
        <div
          className="text-sm text-gray-700 bg-white rounded-lg p-4 border border-gray-200 leading-relaxed max-h-64 overflow-y-auto"
          dangerouslySetInnerHTML={{ __html: formatLargeText(afterFunctions) }}
        />
      )}
    </div>
  );
};

// Unknown content fallback
const UnknownContentBlock: FC<{
  content: unknown;
  className?: string;
}> = ({ content, className }) => {
  const type = content && typeof content === 'object' && 'type' in content
    ? (content as Record<string, unknown>).type
    : 'unknown';

  return (
    <div className={cn('bg-amber-50 border border-amber-200 rounded-lg p-4', className)}>
      <div className="flex items-center gap-2 mb-2">
        <FileText className="w-4 h-4 text-amber-600" />
        <span className="text-sm font-medium text-amber-700">
          Unknown content type: {String(type)}
        </span>
      </div>
      <details className="cursor-pointer">
        <summary className="text-xs text-amber-600 hover:text-amber-800 underline">
          Show raw content
        </summary>
        <pre className="mt-2 text-xs overflow-x-auto bg-white rounded p-3 border border-amber-200 font-mono">
          {JSON.stringify(content, null, 2)}
        </pre>
      </details>
    </div>
  );
};
