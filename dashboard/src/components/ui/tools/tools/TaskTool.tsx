import { type FC } from 'react';
import { Users, FileText } from 'lucide-react';

export const TaskTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const subagentType = String(input.subagent_type || '');
  const description = String(input.description || '');
  const prompt = String(input.prompt || '');
  const model = input.model as string | undefined;
  const runInBackground = input.run_in_background as boolean | undefined;

  return (
    <div className="space-y-3">
      {/* Agent type */}
      <div className="flex items-center gap-2">
        <Users className="w-4 h-4 text-indigo-500" />
        <span className="font-medium text-indigo-700">{subagentType}</span>
        {model && (
          <span className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded">
            {model}
          </span>
        )}
        {runInBackground && (
          <span className="px-2 py-0.5 bg-blue-50 text-blue-600 text-xs rounded">
            Background
          </span>
        )}
      </div>

      {/* Description */}
      {description && <div className="text-sm text-gray-600">{description}</div>}

      {/* Prompt preview */}
      {prompt && (
        <details className="text-sm">
          <summary className="cursor-pointer text-blue-600 hover:text-blue-700">
            <FileText className="w-4 h-4 inline mr-1" />
            Show prompt ({prompt.length} chars)
          </summary>
          <pre className="mt-2 p-3 bg-gray-50 rounded text-xs overflow-x-auto whitespace-pre-wrap">
            {prompt}
          </pre>
        </details>
      )}
    </div>
  );
};
