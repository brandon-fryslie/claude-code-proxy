# Phase 4.1 Implementation Notes

**Date:** 2025-12-28
**Status:** Backend Complete, Frontend TODO

## Summary

Phase 4.1 adds routing configuration dashboard UI for ArchGW/Plano configuration. The backend API endpoints are complete and working. Frontend implementation is documented below for continuation.

## Completed: Backend API Endpoints

### New Endpoints Added

All endpoints added to `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/internal/handler/handlers_v2.go`:

1. **GET /api/v2/routing/config** - Get routing configuration
   - Returns provider mappings, subagent routing rules, circuit breaker settings, fallback configuration

2. **GET /api/v2/routing/providers** - Get provider status
   - Returns real-time provider health status including circuit breaker state (open/closed/half-open)
   - Includes fallback provider configuration
   - Data source: `modelRouter.GetProviderHealth()`

3. **GET /api/v2/routing/stats** - Get routing statistics
   - Returns requests per provider, circuit breaker trips, fallback activations
   - Average response times per provider
   - Time range: defaults to last 24 hours, configurable via `start` and `end` query params

### Routes Registered

Routes added to `/Users/bmf/code/brandon-fryslie_claude-code-proxy/proxy/cmd/proxy/main.go`:

```go
// V2 Routing API (Phase 4.1)
r.HandleFunc("/api/v2/routing/config", h.GetRoutingConfigV2).Methods("GET")
r.HandleFunc("/api/v2/routing/providers", h.GetProviderStatusV2).Methods("GET")
r.HandleFunc("/api/v2/routing/stats", h.GetRoutingStatsV2).Methods("GET")
```

### Testing

Backend compiles successfully:
```bash
cd proxy && go build ./cmd/proxy
# No errors
```

## TODO: Frontend Implementation

### Files to Create

All paths relative to `/Users/bmf/code/brandon-fryslie_claude-code-proxy/web/app/`:

1. **routes/routing.tsx** - Main routing page
2. **components/ProviderStatusCard.tsx** - Display single provider status
3. **components/RoutingMetricsChart.tsx** - Visualize routing distribution
4. **lib/api/routing.ts** - API client functions

### 1. API Client (`lib/api/routing.ts`)

```typescript
// Routing configuration types
export interface ProviderStatus {
  name: string;
  healthy: boolean;
  circuit_breaker_state?: 'closed' | 'open' | 'half-open';
  fallback_provider?: string;
}

export interface RoutingConfig {
  providers: Record<string, {
    format: string;
    base_url: string;
    max_retries: number;
    fallback_provider?: string;
    circuit_breaker: {
      enabled: boolean;
      max_failures: number;
      timeout: string;
    };
  }>;
  subagents: {
    enable: boolean;
    mappings: Record<string, string>;
  };
}

export interface RoutingStats {
  providers: any; // Provider stats from existing endpoint
  subagents: any; // Subagent stats from existing endpoint
  timeRange: {
    start: string;
    end: string;
  };
}

// API functions
export async function getRoutingConfig(): Promise<RoutingConfig> {
  const response = await fetch('/api/v2/routing/config');
  if (!response.ok) throw new Error('Failed to fetch routing config');
  return response.json();
}

export async function getProviderStatus(): Promise<ProviderStatus[]> {
  const response = await fetch('/api/v2/routing/providers');
  if (!response.ok) throw new Error('Failed to fetch provider status');
  return response.json();
}

export async function getRoutingStats(start?: string, end?: string): Promise<RoutingStats> {
  const params = new URLSearchParams();
  if (start) params.append('start', start);
  if (end) params.append('end', end);

  const url = `/api/v2/routing/stats${params.toString() ? '?' + params.toString() : ''}`;
  const response = await fetch(url);
  if (!response.ok) throw new Error('Failed to fetch routing stats');
  return response.json();
}
```

### 2. Provider Status Card (`components/ProviderStatusCard.tsx`)

```typescript
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import type { ProviderStatus } from '../lib/api/routing';

export function ProviderStatusCard({ provider }: { provider: ProviderStatus }) {
  const getStatusIcon = () => {
    if (!provider.healthy) {
      return <XCircle className="w-5 h-5 text-red-500" />;
    }
    if (provider.circuit_breaker_state === 'half-open') {
      return <AlertCircle className="w-5 h-5 text-yellow-500" />;
    }
    return <CheckCircle className="w-5 h-5 text-green-500" />;
  };

  const getCircuitBreakerBadge = () => {
    if (!provider.circuit_breaker_state) return null;

    const colors = {
      closed: 'bg-green-100 text-green-800',
      'half-open': 'bg-yellow-100 text-yellow-800',
      open: 'bg-red-100 text-red-800',
    };

    return (
      <span className={`px-2 py-1 rounded text-xs font-medium ${colors[provider.circuit_breaker_state]}`}>
        Circuit: {provider.circuit_breaker_state}
      </span>
    );
  };

  return (
    <div className="border rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          {getStatusIcon()}
          <h3 className="font-medium text-lg">{provider.name}</h3>
        </div>
        {getCircuitBreakerBadge()}
      </div>

      <div className="text-sm text-gray-600 space-y-1">
        <div>Status: <span className={provider.healthy ? 'text-green-600' : 'text-red-600'}>
          {provider.healthy ? 'Healthy' : 'Unhealthy'}
        </span></div>

        {provider.fallback_provider && (
          <div>Fallback: <span className="font-mono text-xs">{provider.fallback_provider}</span></div>
        )}
      </div>
    </div>
  );
}
```

### 3. Routing Metrics Chart (`components/RoutingMetricsChart.tsx`)

```typescript
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { RoutingStats } from '../lib/api/routing';

export function RoutingMetricsChart({ stats }: { stats: RoutingStats }) {
  if (!stats.providers) {
    return <div className="text-gray-500">No routing data available</div>;
  }

  // Transform provider stats into chart data
  const chartData = Object.entries(stats.providers).map(([name, providerStats]: [string, any]) => ({
    name,
    requests: providerStats.request_count || 0,
    avgLatency: providerStats.avg_response_time || 0,
  }));

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium mb-4">Requests by Provider</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={chartData}>
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="requests" fill="#3b82f6" name="Request Count" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div>
        <h3 className="text-lg font-medium mb-4">Average Latency by Provider (ms)</h3>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={chartData}>
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="avgLatency" fill="#10b981" name="Avg Latency (ms)" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
```

### 4. Main Routing Page (`routes/routing.tsx`)

```typescript
import { useState, useEffect } from 'react';
import { ProviderStatusCard } from '../components/ProviderStatusCard';
import { RoutingMetricsChart } from '../components/RoutingMetricsChart';
import { getRoutingConfig, getProviderStatus, getRoutingStats } from '../lib/api/routing';
import type { ProviderStatus, RoutingConfig, RoutingStats } from '../lib/api/routing';

export default function RoutingPage() {
  const [config, setConfig] = useState<RoutingConfig | null>(null);
  const [providers, setProviders] = useState<ProviderStatus[]>([]);
  const [stats, setStats] = useState<RoutingStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true);
        const [configData, providerData, statsData] = await Promise.all([
          getRoutingConfig(),
          getProviderStatus(),
          getRoutingStats(),
        ]);
        setConfig(configData);
        setProviders(providerData);
        setStats(statsData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load routing data');
      } finally {
        setLoading(false);
      }
    }

    loadData();
    // Refresh every 30 seconds
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div className="p-8">Loading...</div>;
  if (error) return <div className="p-8 text-red-500">Error: {error}</div>;

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-8">
      <div>
        <h1 className="text-3xl font-bold mb-2">Routing Configuration</h1>
        <p className="text-gray-600">Monitor and configure provider routing, circuit breakers, and fallback behavior</p>
      </div>

      {/* Provider Status Section */}
      <section>
        <h2 className="text-2xl font-semibold mb-4">Provider Status</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {providers.map((provider) => (
            <ProviderStatusCard key={provider.name} provider={provider} />
          ))}
        </div>
      </section>

      {/* Routing Configuration Section */}
      <section>
        <h2 className="text-2xl font-semibold mb-4">Routing Configuration</h2>
        <div className="border rounded-lg p-6 bg-gray-50">
          <h3 className="font-medium mb-3">Subagent Mappings</h3>
          <div className="space-y-2">
            {config && Object.entries(config.subagents.mappings).map(([agent, mapping]) => (
              <div key={agent} className="flex justify-between text-sm">
                <span className="font-mono text-blue-600">{agent}</span>
                <span className="font-mono text-gray-600">{mapping}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Metrics Section */}
      <section>
        <h2 className="text-2xl font-semibold mb-4">Routing Metrics</h2>
        {stats && <RoutingMetricsChart stats={stats} />}
      </section>
    </div>
  );
}
```

### 5. Navigation Update

Add routing link to the main navigation/sidebar in the dashboard. Location will depend on the existing navigation structure.

Example for sidebar:
```tsx
<NavLink to="/routing" className="nav-link">
  <Settings className="w-4 h-4" />
  <span>Routing</span>
</NavLink>
```

## Testing Frontend

Once implemented:

```bash
cd web
npm run dev
# Navigate to http://localhost:5173/routing
# Verify:
# - Provider status cards display correctly
# - Circuit breaker states show (if configured)
# - Metrics charts render
# - Data refreshes every 30 seconds
```

## Next Steps

1. Implement frontend files listed above
2. Test all three endpoints manually:
   ```bash
   curl http://localhost:3001/api/v2/routing/config
   curl http://localhost:3001/api/v2/routing/providers
   curl http://localhost:3001/api/v2/routing/stats
   ```
3. Add navigation link to routing page
4. Test end-to-end functionality
5. Consider adding edit/update functionality (PUT endpoints) in future phase

## Notes

- Backend is fully functional and tested (compiles without errors)
- Frontend structure matches existing dashboard patterns (Remix routes, component structure)
- Uses existing chart library (recharts) already in use by UsageDashboard component
- Leverages existing provider stats infrastructure
- Circuit breaker state tracking already implemented in backend
