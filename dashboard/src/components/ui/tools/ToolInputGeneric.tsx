import { type FC, useState } from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';

export const ToolInputGeneric: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const entries = Object.entries(input);

  if (entries.length === 0) {
    return <span className="text-gray-400 text-sm italic">No parameters</span>;
  }

  // Show first 3 params, rest in expandable section
  const visibleParams = entries.slice(0, 3);
  const hiddenParams = entries.slice(3);
  const [showAll, setShowAll] = useState(false);

  return (
    <div className="space-y-2">
      {visibleParams.map(([key, value]) => (
        <ParamRow key={key} name={key} value={value} />
      ))}

      {hiddenParams.length > 0 && (
        <>
          {showAll &&
            hiddenParams.map(([key, value]) => <ParamRow key={key} name={key} value={value} />)}
          <button
            onClick={() => setShowAll(!showAll)}
            className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700"
          >
            <ChevronDown
              className={cn('w-3 h-3 transition-transform', showAll && 'rotate-180')}
            />
            {showAll ? 'Show less' : `Show ${hiddenParams.length} more parameters`}
          </button>
        </>
      )}
    </div>
  );
};

const ParamRow: FC<{ name: string; value: unknown }> = ({ name, value }) => {
  return (
    <div className="text-sm">
      <span className="text-gray-500 mr-2">{name}:</span>
      <ParamValue value={value} />
    </div>
  );
};

const ParamValue: FC<{ value: unknown }> = ({ value }) => {
  if (value === null || value === undefined) {
    return <span className="text-gray-400 italic">null</span>;
  }

  if (typeof value === 'boolean') {
    return (
      <span className={value ? 'text-green-600' : 'text-red-600'}>{String(value)}</span>
    );
  }

  if (typeof value === 'number') {
    return <span className="text-cyan-600 font-mono">{value}</span>;
  }

  if (typeof value === 'string') {
    // Truncate long strings
    if (value.length > 100 || value.includes('\n')) {
      return (
        <details className="inline">
          <summary className="cursor-pointer text-blue-600">
            String ({value.length} chars)
          </summary>
          <pre className="mt-1 p-2 bg-gray-50 rounded text-xs overflow-x-auto whitespace-pre-wrap max-h-48">
            {value}
          </pre>
        </details>
      );
    }
    return <span className="font-mono text-amber-700">"{value}"</span>;
  }

  if (typeof value === 'object') {
    const json = JSON.stringify(value, null, 2);
    return (
      <details className="inline">
        <summary className="cursor-pointer text-blue-600">
          {Array.isArray(value)
            ? `Array (${value.length})`
            : `Object (${Object.keys(value).length} keys)`}
        </summary>
        <pre className="mt-1 p-2 bg-gray-50 rounded text-xs overflow-x-auto max-h-48">
          {json}
        </pre>
      </details>
    );
  }

  return <span className="font-mono">{String(value)}</span>;
};
