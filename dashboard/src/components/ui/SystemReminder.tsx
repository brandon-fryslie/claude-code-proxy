import { type FC, useState } from 'react'
import { ChevronDown, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SystemReminderProps {
  content: string
}

export const SystemReminder: FC<SystemReminderProps> = ({ content }) => {
  const [expanded, setExpanded] = useState(false)

  return (
    <details
      className="border border-amber-200 rounded-lg bg-amber-50 overflow-hidden"
      open={expanded}
      onToggle={(e) => setExpanded((e.target as HTMLDetailsElement).open)}
    >
      <summary className="flex items-center gap-2 px-4 py-2 cursor-pointer hover:bg-amber-100 transition-colors">
        <AlertCircle className="w-4 h-4 text-amber-600 flex-shrink-0" />
        <span className="text-sm font-medium text-amber-900">System Reminder</span>
        <ChevronDown
          className={cn(
            "w-4 h-4 text-amber-600 transition-transform ml-auto",
            expanded && "rotate-180"
          )}
        />
      </summary>
      <div className="px-4 py-3 border-t border-amber-200 bg-white">
        <div className="text-sm text-gray-700 whitespace-pre-wrap">
          {content}
        </div>
      </div>
    </details>
  )
}
