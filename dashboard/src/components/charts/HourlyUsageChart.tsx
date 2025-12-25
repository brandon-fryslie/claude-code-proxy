import { type FC, useMemo } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { formatTokens, getModelColor, getModelDisplayName } from './WeeklyUsageChart';

interface HourlyTokens {
  hour: number;
  tokens: number;
  requests: number;
  models?: Record<string, { tokens: number; requests: number }>;
}

interface HourlyUsageChartProps {
  data: HourlyTokens[];
  isToday?: boolean;
  height?: number;
}

interface HourChartDataPoint {
  hour: number;
  hourLabel: string;
  totalTokens: number;
  totalRequests: number;
  [model: string]: unknown;
}

// Hour labels
const HOUR_LABELS: Record<number, string> = {
  0: '12 AM',
  6: '6 AM',
  12: '12 PM',
  18: '6 PM',
};

export const HourlyUsageChart: FC<HourlyUsageChartProps> = ({
  data,
  isToday = false,
  height = 250,
}) => {
  // Get all unique models
  const allModels = useMemo(() => {
    const modelSet = new Set<string>();
    data.forEach(hour => {
      if (hour.models) {
        Object.keys(hour.models).forEach(model => modelSet.add(model));
      }
    });
    return Array.from(modelSet).sort();
  }, [data]);

  // Transform data - fill in missing hours
  const chartData = useMemo(() => {
    const fullDay: HourChartDataPoint[] = [];
    for (let h = 0; h < 24; h++) {
      const hourData = data.find(d => d.hour === h);
      const result: HourChartDataPoint = {
        hour: h,
        hourLabel: HOUR_LABELS[h] || '',
        totalTokens: hourData?.tokens || 0,
        totalRequests: hourData?.requests || 0,
      };

      allModels.forEach(model => {
        result[model] = hourData?.models?.[model]?.tokens || 0;
      });

      fullDay.push(result);
    }
    return fullDay;
  }, [data, allModels]);

  // Current hour for "now" indicator
  const currentHour = new Date().getHours();

  // Custom tooltip
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;

    const hourData = chartData.find(d => d.hour === label);
    if (!hourData) return null;

    const hourStr = formatHour(label);

    return (
      <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm">
        <div className="font-semibold text-gray-900 mb-2">{hourStr}</div>
        <div className="space-y-1">
          {payload
            .filter((p: any) => p.value > 0)
            .map((p: any) => (
              <div key={p.dataKey} className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded"
                  style={{ backgroundColor: p.fill || p.stroke }}
                />
                <span className="text-gray-600">{getModelDisplayName(p.dataKey)}:</span>
                <span className="font-medium">{formatTokens(p.value)}</span>
              </div>
            ))}
        </div>
        <div className="mt-2 pt-2 border-t border-gray-100 text-gray-500">
          {hourData.totalRequests} requests
        </div>
      </div>
    );
  };

  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 10 }}>
        <defs>
          {allModels.map(model => (
            <linearGradient key={model} id={`gradient-${model}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={getModelColor(model)} stopOpacity={0.8} />
              <stop offset="95%" stopColor={getModelColor(model)} stopOpacity={0.1} />
            </linearGradient>
          ))}
        </defs>

        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />

        <XAxis
          dataKey="hour"
          tick={{ fill: '#6b7280', fontSize: 11 }}
          tickFormatter={(h) => HOUR_LABELS[h] || ''}
          ticks={[0, 6, 12, 18]}
          tickLine={false}
          axisLine={{ stroke: '#e5e7eb' }}
        />

        <YAxis
          tickFormatter={formatTokens}
          tick={{ fill: '#6b7280', fontSize: 11 }}
          tickLine={false}
          axisLine={false}
          width={50}
        />

        <Tooltip content={<CustomTooltip />} />

        {/* Current time indicator (only if viewing today) */}
        {isToday && (
          <ReferenceLine
            x={currentHour}
            stroke="#ef4444"
            strokeWidth={2}
            label={{
              value: 'Now',
              position: 'top',
              fill: '#ef4444',
              fontSize: 11,
            }}
          />
        )}

        {/* Stacked areas for each model */}
        {allModels.map(model => (
          <Area
            key={model}
            type="monotone"
            dataKey={model}
            stackId="tokens"
            stroke={getModelColor(model)}
            fill={`url(#gradient-${model})`}
            strokeWidth={2}
          />
        ))}
      </AreaChart>
    </ResponsiveContainer>
  );
};

function formatHour(hour: number): string {
  if (hour === 0) return '12:00 AM';
  if (hour === 12) return '12:00 PM';
  if (hour < 12) return `${hour}:00 AM`;
  return `${hour - 12}:00 PM`;
}
