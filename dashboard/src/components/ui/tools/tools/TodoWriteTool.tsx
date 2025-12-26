import { type FC } from 'react';
import { CheckCircle, Circle, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Todo {
  content?: string;
  task?: string;
  description?: string;
  title?: string;
  text?: string;
  status: 'pending' | 'in_progress' | 'completed';
  activeForm?: string;
  priority?: 'high' | 'medium' | 'low';
}

interface TodoWriteInput {
  todos?: Todo[];
}

export const TodoWriteTool: FC<{ input: Record<string, unknown> }> = ({ input }) => {
  const { todos = [] } = input as TodoWriteInput;

  if (todos.length === 0) {
    return <div className="text-sm text-gray-500 italic">Empty todo list</div>;
  }

  // Group by status
  const grouped = {
    in_progress: todos.filter((t) => t.status === 'in_progress'),
    pending: todos.filter((t) => t.status === 'pending'),
    completed: todos.filter((t) => t.status === 'completed'),
  };

  // Count summary
  const total = todos.length;
  const done = grouped.completed.length;

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex items-center gap-3 text-sm">
        <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-green-500 transition-all"
            style={{ width: `${(done / total) * 100}%` }}
          />
        </div>
        <span className="text-gray-600">
          {done}/{total} complete
        </span>
      </div>

      {/* In Progress - show first */}
      {grouped.in_progress.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-blue-600 uppercase tracking-wide">
            In Progress
          </div>
          {grouped.in_progress.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}

      {/* Pending */}
      {grouped.pending.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            Pending
          </div>
          {grouped.pending.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}

      {/* Completed */}
      {grouped.completed.length > 0 && (
        <div className="space-y-2">
          <div className="text-xs font-medium text-green-600 uppercase tracking-wide">
            Completed
          </div>
          {grouped.completed.map((todo, i) => (
            <TodoItem key={i} todo={todo} />
          ))}
        </div>
      )}
    </div>
  );
};

const TodoItem: FC<{ todo: Todo }> = ({ todo }) => {
  // Get the todo text from various possible fields
  const text =
    todo.content || todo.task || todo.description || todo.title || todo.text || '';

  const statusConfig = {
    pending: {
      icon: Circle,
      iconClass: 'text-gray-400',
      textClass: 'text-gray-700',
    },
    in_progress: {
      icon: Clock,
      iconClass: 'text-blue-500 animate-pulse',
      textClass: 'text-blue-700 font-medium',
    },
    completed: {
      icon: CheckCircle,
      iconClass: 'text-green-500',
      textClass: 'text-gray-500 line-through',
    },
  };

  const config = statusConfig[todo.status] || statusConfig.pending;
  const Icon = config.icon;

  const priorityColors = {
    high: 'bg-red-100 text-red-700 border-red-200',
    medium: 'bg-amber-100 text-amber-700 border-amber-200',
    low: 'bg-blue-100 text-blue-700 border-blue-200',
  };

  return (
    <div className="flex items-start gap-2 py-1.5">
      <Icon className={cn('w-4 h-4 mt-0.5 flex-shrink-0', config.iconClass)} />
      <div className="flex-1 min-w-0">
        <span className={cn('text-sm', config.textClass)}>
          {todo.status === 'in_progress' && todo.activeForm ? todo.activeForm : text}
        </span>
        {todo.priority && (
          <span
            className={cn(
              'ml-2 px-1.5 py-0.5 text-xs rounded border',
              priorityColors[todo.priority]
            )}
          >
            {todo.priority}
          </span>
        )}
      </div>
    </div>
  );
};
