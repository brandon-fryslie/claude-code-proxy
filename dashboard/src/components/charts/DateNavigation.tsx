import { type FC } from 'react';
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DateNavigationProps {
  selectedDate: Date;
  onDateChange: (date: Date) => void;
  mode?: 'day' | 'week';
  disableForward?: boolean;  // Disable going past today
}

export const DateNavigation: FC<DateNavigationProps> = ({
  selectedDate,
  onDateChange,
  mode = 'day',
  disableForward = true,
}) => {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const isAtToday = selectedDate >= today;

  const goBack = () => {
    const newDate = new Date(selectedDate);
    if (mode === 'week') {
      newDate.setDate(newDate.getDate() - 7);
    } else {
      newDate.setDate(newDate.getDate() - 1);
    }
    onDateChange(newDate);
  };

  const goForward = () => {
    if (disableForward && isAtToday) return;

    const newDate = new Date(selectedDate);
    if (mode === 'week') {
      newDate.setDate(newDate.getDate() + 7);
    } else {
      newDate.setDate(newDate.getDate() + 1);
    }
    onDateChange(newDate);
  };

  const goToToday = () => {
    onDateChange(new Date());
  };

  const formatDate = (date: Date): string => {
    if (mode === 'week') {
      const start = getWeekStart(date);
      const end = new Date(start);
      end.setDate(end.getDate() + 6);
      return `${formatShortDate(start)} - ${formatShortDate(end)}`;
    }
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="flex items-center gap-2">
      <button
        onClick={goBack}
        className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
        title={mode === 'week' ? 'Previous week' : 'Previous day'}
      >
        <ChevronLeft className="w-5 h-5 text-gray-600" />
      </button>

      <div className="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg min-w-[180px] justify-center">
        <Calendar className="w-4 h-4 text-gray-400" />
        <span className="text-sm font-medium text-gray-700">
          {formatDate(selectedDate)}
        </span>
      </div>

      <button
        onClick={goForward}
        disabled={disableForward && isAtToday}
        className={cn(
          "p-2 rounded-lg transition-colors",
          disableForward && isAtToday
            ? "text-gray-300 cursor-not-allowed"
            : "hover:bg-gray-100 text-gray-600"
        )}
        title={mode === 'week' ? 'Next week' : 'Next day'}
      >
        <ChevronRight className="w-5 h-5" />
      </button>

      {!isAtToday && (
        <button
          onClick={goToToday}
          className="px-3 py-1.5 text-sm text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
        >
          Today
        </button>
      )}
    </div>
  );
};

function getWeekStart(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  d.setDate(d.getDate() - day);  // Go to Sunday
  d.setHours(0, 0, 0, 0);
  return d;
}

function formatShortDate(date: Date): string {
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}
