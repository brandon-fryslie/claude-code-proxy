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
