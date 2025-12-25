import { type FC } from 'react'
import { GitCompare, X } from 'lucide-react'

interface CompareModeBannerProps {
  selectedCount: number
  onCompare: () => void
  onCancel: () => void
}

export const CompareModeBanner: FC<CompareModeBannerProps> = ({
  selectedCount,
  onCompare,
  onCancel,
}) => {
  return (
    <div className="sticky top-0 z-40 bg-indigo-600 text-white px-4 py-2 flex items-center justify-between shadow-lg">
      <div className="flex items-center gap-3">
        <GitCompare className="w-5 h-5" />
        <span className="font-medium">Compare Mode</span>
        <span className="px-2 py-0.5 bg-white/20 rounded text-sm">
          {selectedCount}/2 selected
        </span>
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={onCompare}
          disabled={selectedCount !== 2}
          className="px-4 py-1.5 bg-white text-indigo-600 font-medium rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-indigo-50 transition-colors"
        >
          Compare Selected
        </button>
        <button
          onClick={onCancel}
          className="p-1.5 hover:bg-white/20 rounded-lg transition-colors"
          title="Exit compare mode"
        >
          <X className="w-5 h-5" />
        </button>
      </div>
    </div>
  )
}
