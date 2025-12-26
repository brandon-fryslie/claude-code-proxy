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
