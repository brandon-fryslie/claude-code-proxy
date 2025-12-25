import { type FC, useState } from 'react'
import { ChevronDown, Code } from 'lucide-react'
import { cn } from '@/lib/utils'
import { CodeViewer } from './CodeViewer'

interface FunctionDefinitionsProps {
  blocks: string[]
}

export const FunctionDefinitions: FC<FunctionDefinitionsProps> = ({ blocks }) => {
  const [expanded, setExpanded] = useState(false)

  if (blocks.length === 0) return null

  return (
    <details
      className="border border-blue-200 rounded-lg bg-blue-50 overflow-hidden"
      open={expanded}
      onToggle={(e) => setExpanded((e.target as HTMLDetailsElement).open)}
    >
      <summary className="flex items-center gap-2 px-4 py-2 cursor-pointer hover:bg-blue-100 transition-colors">
        <Code className="w-4 h-4 text-blue-600 flex-shrink-0" />
        <span className="text-sm font-medium text-blue-900">
          Function Definitions ({blocks.length})
        </span>
        <ChevronDown
          className={cn(
            "w-4 h-4 text-blue-600 transition-transform ml-auto",
            expanded && "rotate-180"
          )}
        />
      </summary>
      <div className="px-4 py-3 border-t border-blue-200 bg-white space-y-3">
        {blocks.map((block, i) => (
          <CodeViewer
            key={i}
            code={block}
            language="json"
            maxHeight={300}
            showControls={true}
          />
        ))}
      </div>
    </details>
  )
}
