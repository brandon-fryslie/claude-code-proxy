import { type FC, useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
  Cell,
} from 'recharts';

interface DailyTokens {
  date: string;
  tokens: number;
  requests: number;
  models?: Record<string, { tokens: number; requests: number }>;
}

interface WeeklyUsageChartProps {
  data: DailyTokens[];
  selectedDate?: string;
  onDateSelect?: (date: string) => void;
  height?: number;
}

interface ChartDataPoint {
  date: string;
  dayName: string;
  totalTokens: number;
  totalRequests: number;
  isToday: boolean;
  isSelected: boolean;
  [model: string]: unknown;
}

// Day names in order (Sunday first)
const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

// Model colors - consistent across all charts
export const MODEL_COLORS: Record<string, string> = {
  // Opus variants (purple)
  'claude-3-opus-20240229': '#9333ea',
  'claude-opus-4-20250514': '#9333ea',
  opus: '#9333ea',

  // Sonnet variants (blue)
  'claude-3-sonnet-20240229': '#3b82f6',
  'claude-3-5-sonnet-20240620': '#3b82f6',
  'claude-3-5-sonnet-20241022': '#3b82f6',
  'claude-sonnet-4-20250514': '#3b82f6',
  sonnet: '#3b82f6',

  // Haiku variants (green)
  'claude-3-haiku-20240307': '#10b981',
  'claude-3-5-haiku-20241022': '#10b981',
  haiku: '#10b981',

  // OpenAI (orange)
  'gpt-4': '#f97316',
  'gpt-4o': '#f97316',
  'gpt-4-turbo': '#f97316',

  // Default (gray)
  default: '#6b7280',
};

export const WeeklyUsageChart: FC<WeeklyUsageChartProps> = ({
  data,
  selectedDate,
  onDateSelect,
  height = 300,
}) => {
  // Get all unique models across all days
  const allModels = useMemo(() => {
    const modelSet = new Set<string>();
    data.forEach(day => {
      if (day.models) {
        Object.keys(day.models).forEach(model => modelSet.add(model));
      }
    });
    return Array.from(modelSet).sort();
  }, [data]);

  // Transform data for Recharts (needs flat structure with model keys)
  const chartData = useMemo(() => {
    return data.map(day => {
      const date = new Date(day.date + 'T00:00:00');
      const dayName = DAY_NAMES[date.getDay()];
      const isToday = day.date === getTodayDateString();
      const isSelected = day.date === selectedDate;

      const result: ChartDataPoint = {
        date: day.date,
        dayName,
        totalTokens: day.tokens,
        totalRequests: day.requests,
        isToday,
        isSelected,
      };

      // Add each model's tokens as a separate key
      allModels.forEach(model => {
        result[model] = day.models?.[model]?.tokens || 0;
      });

      return result;
    });
  }, [data, allModels, selectedDate]);

  // Calculate average for reference line
  const averageTokens = useMemo(() => {
    const total = data.reduce((sum, day) => sum + day.tokens, 0);
    return total / data.length;
  }, [data]);

  // Calculate max for Y-axis
  const maxTokens = useMemo(() => {
    return Math.max(...data.map(d => d.tokens), 1);
  }, [data]);

  // Custom tooltip
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;

    const dayData = chartData.find(d => d.dayName === label);
    if (!dayData) return null;

    return (
      <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm">
        <div className="font-semibold text-gray-900 mb-2">
          {label} - {dayData.date}
        </div>
        <div className="space-y-1">
          {payload
            .filter((p: any) => p.value > 0)
            .sort((a: any, b: any) => b.value - a.value)
            .map((p: any) => (
              <div key={p.dataKey} className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded"
                  style={{ backgroundColor: p.fill }}
                />
                <span className="text-gray-600">{getModelDisplayName(p.dataKey)}:</span>
                <span className="font-medium">{formatTokens(p.value)}</span>
              </div>
            ))}
        </div>
        <div className="mt-2 pt-2 border-t border-gray-100">
          <div className="flex justify-between">
            <span className="text-gray-500">Total:</span>
            <span className="font-semibold">{formatTokens(dayData.totalTokens)}</span>
          </div>
          <div className="flex justify-between text-gray-400">
            <span>Requests:</span>
            <span>{dayData.totalRequests}</span>
          </div>
        </div>
      </div>
    );
  };

  // Handle bar click
  const handleBarClick = (data: any) => {
    if (onDateSelect && data?.date) {
      onDateSelect(data.date);
    }
  };

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart
        data={chartData}
        margin={{ top: 20, right: 20, left: 20, bottom: 20 }}
        onClick={(e: any) => {
          if (e?.activePayload?.[0]) {
            handleBarClick(e.activePayload[0].payload);
          }
        }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />

        <XAxis
          dataKey="dayName"
          tick={{ fill: '#6b7280', fontSize: 12 }}
          tickLine={false}
          axisLine={{ stroke: '#e5e7eb' }}
        />

        <YAxis
          tickFormatter={formatTokens}
          tick={{ fill: '#6b7280', fontSize: 12 }}
          tickLine={false}
          axisLine={false}
          domain={[0, maxTokens * 1.1]}
        />

        <Tooltip content={<CustomTooltip />} />

        {/* Average line */}
        <ReferenceLine
          y={averageTokens}
          stroke="#9ca3af"
          strokeDasharray="5 5"
          label={{
            value: `Avg: ${formatTokens(averageTokens)}`,
            position: 'right',
            fill: '#9ca3af',
            fontSize: 11,
          }}
        />

        {/* Stacked bars for each model */}
        {allModels.map((model, index) => (
          <Bar
            key={model}
            dataKey={model}
            stackId="tokens"
            fill={getModelColor(model)}
            radius={index === allModels.length - 1 ? [4, 4, 0, 0] : undefined}
            cursor="pointer"
          >
            {/* Highlight selected/today bars */}
            {chartData.map((entry, i) => (
              <Cell
                key={i}
                opacity={entry.isSelected ? 1 : entry.isToday ? 0.9 : 0.8}
                strokeWidth={entry.isSelected ? 2 : 0}
                stroke={entry.isSelected ? '#1f2937' : undefined}
              />
            ))}
          </Bar>
        ))}

        <Legend
          formatter={(value) => getModelDisplayName(value)}
          wrapperStyle={{ paddingTop: 10 }}
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

// Utility functions
function getTodayDateString(): string {
  return new Date().toISOString().split('T')[0];
}

export function getModelColor(model: string): string {
  // Check exact match first
  if (MODEL_COLORS[model]) return MODEL_COLORS[model];

  // Check for partial matches
  const lowerModel = model.toLowerCase();
  if (lowerModel.includes('opus')) return MODEL_COLORS.opus;
  if (lowerModel.includes('sonnet')) return MODEL_COLORS.sonnet;
  if (lowerModel.includes('haiku')) return MODEL_COLORS.haiku;
  if (lowerModel.includes('gpt')) return MODEL_COLORS['gpt-4'];

  return MODEL_COLORS.default;
}

export function getModelDisplayName(model: string): string {
  // Simplify long model names
  if (model.includes('opus')) return 'Opus';
  if (model.includes('sonnet')) return 'Sonnet';
  if (model.includes('haiku')) return 'Haiku';
  if (model.includes('gpt-4o')) return 'GPT-4o';
  if (model.includes('gpt-4')) return 'GPT-4';

  // Fallback: extract the key part
  const parts = model.split('-');
  if (parts.length > 2) {
    return parts.slice(0, 2).join('-');
  }
  return model;
}

export function formatTokens(value: number): string {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(1)}B`;
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M`;
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`;
  }
  return value.toFixed(0);
}
