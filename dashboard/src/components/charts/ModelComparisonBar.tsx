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
