# Phase 4: Charts & Analytics - Implementation Handoff

**Branch:** `phase-4-charts-analytics`
**Epic:** `brandon-fryslie_claude-code-proxy-loz` (weekly-usage-chart), `brandon-fryslie_claude-code-proxy-xwb` (model-breakdown-stats), `brandon-fryslie_claude-code-proxy-eee` (performance-metrics)

---

## Executive Summary

You are implementing the **Charts & Analytics** system for the new dashboard. Your work transforms raw token/request data into beautiful, interactive visualizations that help users understand their Claude API usage patterns.

**Prerequisites:** The new dashboard already has Recharts installed and basic charts in place. You're enhancing and expanding these visualizations.

Your deliverables:
1. **Weekly Usage Chart** - 7-day stacked bar chart with model breakdown
2. **Model Breakdown Stats** - Per-model token/request analysis with pie/bar charts
3. **Performance Metrics** - Response time percentiles, latency distribution

---

## Your Working Environment

### Directory Structure
```
/Users/bmf/code/brandon-fryslie_claude-code-proxy/
├── dashboard/              # NEW dashboard (your target)
│   ├── src/
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── ui/
│   │   │   └── charts/     # CREATE THIS - your chart components
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx   # Has basic charts (enhance)
│   │   │   ├── Usage.tsx       # Token usage page (enhance)
│   │   │   └── Performance.tsx # Performance page (enhance)
│   │   └── lib/
│   │       ├── types.ts    # TypeScript interfaces
│   │       ├── api.ts      # React Query hooks (already has stats hooks)
│   │       └── utils.ts
├── web/                    # OLD dashboard (reference)
│   └── app/components/
│       └── UsageDashboard.tsx  # CRITICAL REFERENCE
└── proxy/                  # Go backend
    └── internal/
        ├── handler/handlers.go  # Stats API handlers
        └── model/models.go      # Data structures
```

### Tech Stack
- React 19.2.0
- TypeScript 5.9.3
- **Recharts 3.6.0** - Already installed, use this for all charts
- Tailwind CSS 4.1.18
- TanStack React Query 5.90.12

### Existing Chart Infrastructure

The dashboard already has these working:
- `useWeeklyStats()` - Fetches 7-day aggregate data
- `useHourlyStats()` - Fetches 24-hour breakdown
- `useModelStats()` - Fetches per-model breakdown
- `useProviderStats()` - Fetches per-provider breakdown
- `usePerformanceStats()` - Fetches response time percentiles

---

## API Endpoints You'll Use

### GET /api/stats

Returns weekly aggregate data.

**Query Parameters:**
- `start`: ISO 8601 UTC timestamp (week start)
- `end`: ISO 8601 UTC timestamp (week end)

**Response:**
```typescript
interface DashboardStats {
  dailyStats: DailyTokens[];
}

interface DailyTokens {
  date: string;           // "2024-12-24" (YYYY-MM-DD)
  tokens: number;         // Total tokens (input + output)
  requests: number;       // Request count
  models: Record<string, ModelStats>;  // Per-model breakdown
}

interface ModelStats {
  tokens: number;
  requests: number;
}
```

### GET /api/stats/hourly

Returns hourly breakdown for a specific day.

**Query Parameters:**
- `start`: ISO 8601 UTC (day start)
- `end`: ISO 8601 UTC (day end)

**Response:**
```typescript
interface HourlyStatsResponse {
  hourlyStats: HourlyTokens[];
  todayTokens: number;      // Day total
  todayRequests: number;
  avgResponseTime: number;  // ms
}

interface HourlyTokens {
  hour: number;           // 0-23 (UTC)
  tokens: number;
  requests: number;
  models: Record<string, ModelStats>;
}
```

### GET /api/stats/models

Returns per-model breakdown for date range.

**Response:**
```typescript
interface ModelStatsResponse {
  modelStats: {
    model: string;
    tokens: number;
    requests: number;
  }[];
}
```

### GET /api/stats/performance

Returns response time percentiles.

**Response:**
```typescript
interface PerformanceStatsResponse {
  stats: PerformanceStats[];
  startTime: string;
  endTime: string;
}

interface PerformanceStats {
  provider: string;
  model: string;
  avgResponseMs: number;
  p50ResponseMs: number;  // Median
  p95ResponseMs: number;  // 95th percentile
  p99ResponseMs: number;  // 99th percentile
  avgFirstByteMs: number; // Time to first token (streaming)
  requestCount: number;
}
```

---

## Topic 1: Weekly Usage Chart

### What You're Building

A 7-day stacked bar chart showing token usage by model, similar to the old dashboard but with Recharts.

### Key Features
1. Sunday-Saturday week boundaries
2. Stacked bars colored by model (Opus=purple, Sonnet=blue, Haiku=green)
3. Day name labels (Sun, Mon, Tue, etc.)
4. Hover tooltips with model breakdown
5. Y-axis with smart token formatting (K, M, B)
6. Average line (dashed)
7. Highlight "today" if visible
8. Click-to-select day functionality

### WeeklyUsageChart.tsx

```tsx
// dashboard/src/components/charts/WeeklyUsageChart.tsx
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
import { formatTokens, getModelColor, getModelDisplayName } from '@/lib/chartUtils';

interface DailyTokens {
  date: string;
  tokens: number;
  requests: number;
  models: Record<string, { tokens: number; requests: number }>;
}

interface WeeklyUsageChartProps {
  data: DailyTokens[];
  selectedDate?: string;          // Currently selected date (YYYY-MM-DD)
  onDateSelect?: (date: string) => void;
  height?: number;
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
      Object.keys(day.models).forEach(model => modelSet.add(model));
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

      const result: Record<string, unknown> = {
        date: day.date,
        dayName,
        totalTokens: day.tokens,
        totalRequests: day.requests,
        isToday,
        isSelected,
      };

      // Add each model's tokens as a separate key
      allModels.forEach(model => {
        result[model] = day.models[model]?.tokens || 0;
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
            <span className="font-semibold">{formatTokens(dayData.totalTokens as number)}</span>
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
        onClick={(e) => e?.activePayload?.[0] && handleBarClick(e.activePayload[0].payload)}
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
```

### HourlyUsageChart.tsx

```tsx
// dashboard/src/components/charts/HourlyUsageChart.tsx
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
  models: Record<string, { tokens: number; requests: number }>;
}

interface HourlyUsageChartProps {
  data: HourlyTokens[];
  isToday?: boolean;
  height?: number;
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
      Object.keys(hour.models).forEach(model => modelSet.add(model));
    });
    return Array.from(modelSet).sort();
  }, [data]);

  // Transform data - fill in missing hours
  const chartData = useMemo(() => {
    const fullDay = [];
    for (let h = 0; h < 24; h++) {
      const hourData = data.find(d => d.hour === h);
      const result: Record<string, unknown> = {
        hour: h,
        hourLabel: HOUR_LABELS[h] || '',
        totalTokens: hourData?.tokens || 0,
        totalRequests: hourData?.requests || 0,
      };

      allModels.forEach(model => {
        result[model] = hourData?.models[model]?.tokens || 0;
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
                  style={{ backgroundColor: p.fill }}
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
```

---

## Topic 2: Model Breakdown Stats

### ModelBreakdownChart.tsx

```tsx
// dashboard/src/components/charts/ModelBreakdownChart.tsx
import { type FC, useMemo } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from 'recharts';
import { formatTokens, getModelColor, getModelDisplayName } from './WeeklyUsageChart';

interface ModelStat {
  model: string;
  tokens: number;
  requests: number;
}

interface ModelBreakdownChartProps {
  data: ModelStat[];
  metric?: 'tokens' | 'requests';
  height?: number;
}

export const ModelBreakdownChart: FC<ModelBreakdownChartProps> = ({
  data,
  metric = 'tokens',
  height = 300,
}) => {
  // Sort by the selected metric (descending)
  const sortedData = useMemo(() => {
    return [...data]
      .sort((a, b) => b[metric] - a[metric])
      .map(item => ({
        ...item,
        displayName: getModelDisplayName(item.model),
        color: getModelColor(item.model),
        value: item[metric],
      }));
  }, [data, metric]);

  // Calculate total
  const total = useMemo(() => {
    return sortedData.reduce((sum, item) => sum + item.value, 0);
  }, [sortedData]);

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (!active || !payload?.length) return null;
    const item = payload[0].payload;
    const percentage = ((item.value / total) * 100).toFixed(1);

    return (
      <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm">
        <div className="flex items-center gap-2 mb-1">
          <div
            className="w-3 h-3 rounded"
            style={{ backgroundColor: item.color }}
          />
          <span className="font-semibold">{item.displayName}</span>
        </div>
        <div className="text-gray-600">
          {metric === 'tokens' ? formatTokens(item.value) : item.value.toLocaleString()} {metric}
        </div>
        <div className="text-gray-400">{percentage}% of total</div>
      </div>
    );
  };

  // Custom legend with percentages
  const renderLegend = (props: any) => {
    const { payload } = props;
    return (
      <div className="flex flex-wrap justify-center gap-4 mt-4">
        {payload.map((entry: any, index: number) => {
          const item = sortedData[index];
          const percentage = ((item.value / total) * 100).toFixed(1);
          return (
            <div key={entry.value} className="flex items-center gap-2 text-sm">
              <div
                className="w-3 h-3 rounded"
                style={{ backgroundColor: entry.color }}
              />
              <span className="text-gray-600">{item.displayName}</span>
              <span className="text-gray-400">({percentage}%)</span>
            </div>
          );
        })}
      </div>
    );
  };

  if (sortedData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-400">
        No data available
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={height}>
      <PieChart>
        <Pie
          data={sortedData}
          cx="50%"
          cy="50%"
          innerRadius={60}
          outerRadius={100}
          paddingAngle={2}
          dataKey="value"
          nameKey="displayName"
        >
          {sortedData.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip content={<CustomTooltip />} />
        <Legend content={renderLegend} />

        {/* Center text showing total */}
        <text
          x="50%"
          y="50%"
          textAnchor="middle"
          dominantBaseline="middle"
          className="fill-gray-900 text-lg font-semibold"
        >
          {metric === 'tokens' ? formatTokens(total) : total.toLocaleString()}
        </text>
        <text
          x="50%"
          y="50%"
          dy={20}
          textAnchor="middle"
          dominantBaseline="middle"
          className="fill-gray-400 text-xs"
        >
          total {metric}
        </text>
      </PieChart>
    </ResponsiveContainer>
  );
};
```

### ModelComparisonBar.tsx

Horizontal bar chart for comparing models side-by-side.

```tsx
// dashboard/src/components/charts/ModelComparisonBar.tsx
import { type FC, useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { formatTokens, getModelColor, getModelDisplayName } from './WeeklyUsageChart';

interface ModelStat {
  model: string;
  tokens: number;
  requests: number;
}

interface ModelComparisonBarProps {
  data: ModelStat[];
  height?: number;
}

export const ModelComparisonBar: FC<ModelComparisonBarProps> = ({
  data,
  height = 250,
}) => {
  // Sort and enhance data
  const chartData = useMemo(() => {
    return [...data]
      .sort((a, b) => b.tokens - a.tokens)
      .map(item => ({
        ...item,
        displayName: getModelDisplayName(item.model),
        color: getModelColor(item.model),
        avgPerRequest: item.requests > 0 ? Math.round(item.tokens / item.requests) : 0,
      }));
  }, [data]);

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (!active || !payload?.length) return null;
    const item = payload[0].payload;

    return (
      <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm">
        <div className="font-semibold text-gray-900 mb-2">{item.displayName}</div>
        <div className="space-y-1">
          <div className="flex justify-between gap-4">
            <span className="text-gray-500">Tokens:</span>
            <span className="font-medium">{formatTokens(item.tokens)}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-gray-500">Requests:</span>
            <span className="font-medium">{item.requests.toLocaleString()}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-gray-500">Avg/Request:</span>
            <span className="font-medium">{formatTokens(item.avgPerRequest)}</span>
          </div>
        </div>
      </div>
    );
  };

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart
        data={chartData}
        layout="vertical"
        margin={{ top: 10, right: 30, left: 80, bottom: 10 }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" horizontal={false} />

        <XAxis
          type="number"
          tickFormatter={formatTokens}
          tick={{ fill: '#6b7280', fontSize: 11 }}
          tickLine={false}
          axisLine={{ stroke: '#e5e7eb' }}
        />

        <YAxis
          type="category"
          dataKey="displayName"
          tick={{ fill: '#374151', fontSize: 12 }}
          tickLine={false}
          axisLine={false}
          width={70}
        />

        <Tooltip content={<CustomTooltip />} />

        <Bar dataKey="tokens" radius={[0, 4, 4, 0]}>
          {chartData.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
};
```

---

## Topic 3: Performance Metrics

### PerformanceChart.tsx

Multi-bar chart showing response time percentiles by model.

```tsx
// dashboard/src/components/charts/PerformanceChart.tsx
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
} from 'recharts';
import { getModelDisplayName } from './WeeklyUsageChart';

interface PerformanceStats {
  provider: string;
  model: string;
  avgResponseMs: number;
  p50ResponseMs: number;
  p95ResponseMs: number;
  p99ResponseMs: number;
  avgFirstByteMs: number;
  requestCount: number;
}

interface PerformanceChartProps {
  data: PerformanceStats[];
  height?: number;
}

// Percentile colors
const PERCENTILE_COLORS = {
  p50: '#10b981',  // Green - median
  p95: '#f59e0b',  // Amber - 95th
  p99: '#ef4444',  // Red - 99th
};

export const PerformanceChart: FC<PerformanceChartProps> = ({
  data,
  height = 300,
}) => {
  // Transform data for display
  const chartData = useMemo(() => {
    return data
      .filter(d => d.requestCount > 0)
      .sort((a, b) => a.p50ResponseMs - b.p50ResponseMs)
      .map(item => ({
        ...item,
        displayName: getModelDisplayName(item.model),
        // Format for display
        p50: item.p50ResponseMs,
        p95: item.p95ResponseMs,
        p99: item.p99ResponseMs,
      }));
  }, [data]);

  // Custom tooltip
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload?.length) return null;

    const item = chartData.find(d => d.displayName === label);
    if (!item) return null;

    return (
      <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm">
        <div className="font-semibold text-gray-900 mb-2">{label}</div>
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: PERCENTILE_COLORS.p50 }} />
            <span className="text-gray-600">P50 (Median):</span>
            <span className="font-medium">{formatDuration(item.p50ResponseMs)}</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: PERCENTILE_COLORS.p95 }} />
            <span className="text-gray-600">P95:</span>
            <span className="font-medium">{formatDuration(item.p95ResponseMs)}</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: PERCENTILE_COLORS.p99 }} />
            <span className="text-gray-600">P99:</span>
            <span className="font-medium">{formatDuration(item.p99ResponseMs)}</span>
          </div>
        </div>
        <div className="mt-2 pt-2 border-t border-gray-100 text-gray-400">
          <div>Avg: {formatDuration(item.avgResponseMs)}</div>
          <div>TTFB: {formatDuration(item.avgFirstByteMs)}</div>
          <div>{item.requestCount.toLocaleString()} requests</div>
        </div>
      </div>
    );
  };

  if (chartData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-400">
        No performance data available
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 20 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />

        <XAxis
          dataKey="displayName"
          tick={{ fill: '#374151', fontSize: 12 }}
          tickLine={false}
          axisLine={{ stroke: '#e5e7eb' }}
        />

        <YAxis
          tickFormatter={formatDuration}
          tick={{ fill: '#6b7280', fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />

        <Tooltip content={<CustomTooltip />} />

        <Legend
          formatter={(value) => {
            const labels: Record<string, string> = {
              p50: 'P50 (Median)',
              p95: 'P95',
              p99: 'P99',
            };
            return labels[value] || value;
          }}
        />

        <Bar dataKey="p50" fill={PERCENTILE_COLORS.p50} name="p50" radius={[4, 4, 0, 0]} />
        <Bar dataKey="p95" fill={PERCENTILE_COLORS.p95} name="p95" radius={[4, 4, 0, 0]} />
        <Bar dataKey="p99" fill={PERCENTILE_COLORS.p99} name="p99" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
};

// Format milliseconds to human-readable
function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  return `${(ms / 60000).toFixed(1)}m`;
}
```

### LatencyDistributionChart.tsx

Histogram showing response time distribution.

```tsx
// dashboard/src/components/charts/LatencyDistributionChart.tsx
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

interface LatencyDistributionProps {
  responseTimes: number[];  // Array of response times in ms
  height?: number;
}

export const LatencyDistributionChart: FC<LatencyDistributionProps> = ({
  responseTimes,
  height = 200,
}) => {
  // Create histogram buckets
  const { buckets, stats } = useMemo(() => {
    if (responseTimes.length === 0) {
      return { buckets: [], stats: { p50: 0, p95: 0, p99: 0, avg: 0 } };
    }

    const sorted = [...responseTimes].sort((a, b) => a - b);
    const min = sorted[0];
    const max = sorted[sorted.length - 1];

    // Calculate percentiles
    const p50 = sorted[Math.floor(sorted.length * 0.5)];
    const p95 = sorted[Math.floor(sorted.length * 0.95)];
    const p99 = sorted[Math.floor(sorted.length * 0.99)];
    const avg = sorted.reduce((a, b) => a + b, 0) / sorted.length;

    // Create ~20 buckets
    const bucketCount = Math.min(20, sorted.length);
    const bucketSize = (max - min) / bucketCount || 1;

    const bucketMap: Record<number, number> = {};
    for (let i = 0; i < bucketCount; i++) {
      const bucketStart = min + i * bucketSize;
      bucketMap[bucketStart] = 0;
    }

    // Count values in each bucket
    sorted.forEach(value => {
      const bucket = Math.floor((value - min) / bucketSize) * bucketSize + min;
      bucketMap[bucket] = (bucketMap[bucket] || 0) + 1;
    });

    const buckets = Object.entries(bucketMap)
      .map(([start, count]) => ({
        start: Number(start),
        count,
        percentage: (count / sorted.length) * 100,
      }))
      .sort((a, b) => a.start - b.start);

    return { buckets, stats: { p50, p95, p99, avg } };
  }, [responseTimes]);

  if (buckets.length === 0) {
    return (
      <div className="flex items-center justify-center h-32 text-gray-400">
        No latency data
      </div>
    );
  }

  return (
    <div>
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={buckets} margin={{ top: 10, right: 10, left: 0, bottom: 10 }}>
          <defs>
            <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.1} />
            </linearGradient>
          </defs>

          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />

          <XAxis
            dataKey="start"
            tickFormatter={(v) => `${Math.round(v)}ms`}
            tick={{ fill: '#6b7280', fontSize: 10 }}
            tickLine={false}
            axisLine={{ stroke: '#e5e7eb' }}
          />

          <YAxis
            tickFormatter={(v) => `${v.toFixed(0)}%`}
            tick={{ fill: '#6b7280', fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            width={40}
          />

          <Tooltip
            formatter={(value: number) => [`${value.toFixed(1)}%`, 'Requests']}
            labelFormatter={(label) => `${Math.round(label)}ms`}
          />

          {/* Reference lines for percentiles */}
          <ReferenceLine
            x={stats.p50}
            stroke="#10b981"
            strokeDasharray="3 3"
            label={{ value: 'P50', position: 'top', fontSize: 10, fill: '#10b981' }}
          />
          <ReferenceLine
            x={stats.p95}
            stroke="#f59e0b"
            strokeDasharray="3 3"
            label={{ value: 'P95', position: 'top', fontSize: 10, fill: '#f59e0b' }}
          />

          <Area
            type="monotone"
            dataKey="percentage"
            stroke="#3b82f6"
            fill="url(#latencyGradient)"
            strokeWidth={2}
          />
        </AreaChart>
      </ResponsiveContainer>

      {/* Stats summary */}
      <div className="flex justify-around text-center mt-2">
        <StatBox label="Avg" value={`${Math.round(stats.avg)}ms`} />
        <StatBox label="P50" value={`${Math.round(stats.p50)}ms`} color="green" />
        <StatBox label="P95" value={`${Math.round(stats.p95)}ms`} color="amber" />
        <StatBox label="P99" value={`${Math.round(stats.p99)}ms`} color="red" />
      </div>
    </div>
  );
};

const StatBox: FC<{ label: string; value: string; color?: string }> = ({
  label,
  value,
  color = 'gray',
}) => {
  const colorClasses: Record<string, string> = {
    gray: 'text-gray-600',
    green: 'text-green-600',
    amber: 'text-amber-600',
    red: 'text-red-600',
  };

  return (
    <div>
      <div className="text-xs text-gray-400">{label}</div>
      <div className={`font-semibold ${colorClasses[color]}`}>{value}</div>
    </div>
  );
};
```

---

## Quick Stats Cards

Reusable stat card component for summary metrics.

```tsx
// dashboard/src/components/charts/StatCard.tsx
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
```

---

## Date Navigation Component

For navigating between dates/weeks.

```tsx
// dashboard/src/components/charts/DateNavigation.tsx
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
```

---

## Integration with Existing Pages

### Update Dashboard.tsx

```tsx
// Replace existing charts with your new components:

import { WeeklyUsageChart } from '@/components/charts/WeeklyUsageChart';
import { HourlyUsageChart } from '@/components/charts/HourlyUsageChart';
import { ModelBreakdownChart } from '@/components/charts/ModelBreakdownChart';
import { StatCard } from '@/components/charts/StatCard';
import { DateNavigation } from '@/components/charts/DateNavigation';

// In the component:
<DateNavigation
  selectedDate={selectedDate}
  onDateChange={setSelectedDate}
  mode="day"
/>

<WeeklyUsageChart
  data={weeklyStats.dailyStats}
  selectedDate={formatDateString(selectedDate)}
  onDateSelect={(date) => setSelectedDate(new Date(date))}
/>

<HourlyUsageChart
  data={hourlyStats.hourlyStats}
  isToday={isToday(selectedDate)}
/>
```

---

## Utility Functions

Create a central file for chart utilities:

```tsx
// dashboard/src/lib/chartUtils.ts
export { formatTokens, getModelColor, getModelDisplayName } from '@/components/charts/WeeklyUsageChart';

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

export function getWeekBoundaries(date: Date): { start: Date; end: Date } {
  const start = new Date(date);
  start.setDate(start.getDate() - start.getDay());  // Sunday
  start.setHours(0, 0, 0, 0);

  const end = new Date(start);
  end.setDate(end.getDate() + 6);  // Saturday
  end.setHours(23, 59, 59, 999);

  return { start, end };
}

export function getDayBoundaries(date: Date): { start: Date; end: Date } {
  const start = new Date(date);
  start.setHours(0, 0, 0, 0);

  const end = new Date(date);
  end.setHours(23, 59, 59, 999);

  return { start, end };
}

export function toISODateString(date: Date): string {
  return date.toISOString().split('T')[0];
}

export function isToday(date: Date): boolean {
  const today = new Date();
  return toISODateString(date) === toISODateString(today);
}
```

---

## Testing Checklist

### Weekly Usage Chart
- [ ] Shows 7 days (Sunday-Saturday)
- [ ] Stacked bars by model with correct colors
- [ ] Day name labels on X-axis
- [ ] Token values on Y-axis with K/M/B formatting
- [ ] Tooltip shows model breakdown
- [ ] Average reference line displayed
- [ ] "Today" highlighted if in view
- [ ] Click on bar selects that date
- [ ] Empty days show empty bar

### Hourly Usage Chart
- [ ] Shows 24 hours
- [ ] Stacked areas by model
- [ ] Hour labels at 12AM, 6AM, 12PM, 6PM
- [ ] "Now" indicator when viewing today
- [ ] Tooltip shows model breakdown
- [ ] Fills missing hours with zero

### Model Breakdown Charts
- [ ] Pie chart shows token distribution
- [ ] Center displays total
- [ ] Legend shows percentages
- [ ] Bar chart sorted by tokens (descending)
- [ ] Correct colors per model
- [ ] Tooltip shows tokens, requests, avg/request

### Performance Charts
- [ ] Multi-bar shows P50, P95, P99
- [ ] Colors distinguish percentiles (green/amber/red)
- [ ] Tooltip shows all stats including TTFB
- [ ] Distribution histogram if data available
- [ ] Stats summary below chart

### Date Navigation
- [ ] Prev/Next buttons work
- [ ] Can't go past today (by default)
- [ ] "Today" button appears when not at today
- [ ] Week mode shows date range

---

## Common Gotchas

1. **Timezone Handling**: All API dates are UTC. Convert to local for display, back to UTC for API calls.
2. **Sunday-Saturday Weeks**: Not Monday-Sunday. Use `date.getDay()` (Sunday = 0).
3. **Empty Data**: Always handle empty arrays gracefully - show "No data" message.
4. **Token Formatting**: Use consistent formatting (K, M, B) everywhere.
5. **Memoization**: Charts re-render expensively. Use `useMemo` for data transformations.
6. **Responsive Design**: All charts must work on different screen sizes. Use `ResponsiveContainer`.
7. **Color Consistency**: Model colors must be identical across all charts.

---

## Reference Files

Study these in the old dashboard:
- `web/app/components/UsageDashboard.tsx` - Complete reference (450+ lines)
- `web/app/routes/_index.tsx` - Date navigation, stats loading patterns

---

## Definition of Done

Phase 4 is complete when:

1. Weekly usage chart with model breakdown works
2. Hourly usage chart with "now" indicator works
3. Model breakdown pie and bar charts work
4. Performance percentile charts work
5. Date navigation works (day and week modes)
6. Stat cards display summary metrics
7. All charts use consistent colors and formatting
8. Charts are responsive and performant
9. No TypeScript errors in strict mode
10. Commit history shows logical, atomic commits

---

**Make the data beautiful. Users love charts that tell a story.**
