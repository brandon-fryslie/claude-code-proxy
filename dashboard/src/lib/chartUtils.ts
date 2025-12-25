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
