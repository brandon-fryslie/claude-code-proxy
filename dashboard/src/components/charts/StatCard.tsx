import { type FC, type ReactNode } from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: ReactNode;
  trend?: {
    value: number;  // Percentage change
    label?: string;
  };
  className?: string;
}

export const StatCard: FC<StatCardProps> = ({
  title,
  value,
  subtitle,
  icon,
  trend,
  className,
}) => {
  const getTrendIcon = () => {
    if (!trend) return null;
    if (trend.value > 0) return <TrendingUp className="w-4 h-4" />;
    if (trend.value < 0) return <TrendingDown className="w-4 h-4" />;
    return <Minus className="w-4 h-4" />;
  };

  const getTrendColor = () => {
    if (!trend) return '';
    if (trend.value > 0) return 'text-green-600';
    if (trend.value < 0) return 'text-red-600';
    return 'text-gray-400';
  };

  return (
    <div className={cn(
      "bg-white rounded-xl border border-gray-200 p-4 shadow-sm",
      className
    )}>
      <div className="flex items-start justify-between">
        <div>
          <div className="text-sm text-gray-500 font-medium">{title}</div>
          <div className="text-2xl font-bold text-gray-900 mt-1">{value}</div>
          {subtitle && (
            <div className="text-xs text-gray-400 mt-1">{subtitle}</div>
          )}
        </div>
        {icon && (
          <div className="p-2 bg-gray-50 rounded-lg text-gray-400">
            {icon}
          </div>
        )}
      </div>

      {trend && (
        <div className={cn("flex items-center gap-1 mt-3 text-sm", getTrendColor())}>
          {getTrendIcon()}
          <span className="font-medium">
            {trend.value > 0 ? '+' : ''}{trend.value.toFixed(1)}%
          </span>
          {trend.label && <span className="text-gray-400">{trend.label}</span>}
        </div>
      )}
    </div>
  );
};
