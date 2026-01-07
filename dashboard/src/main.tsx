import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from './router'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
      refetchOnWindowFocus: false,
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime in v4)
    },
  },
})

// Configure cache limits for request details to prevent memory bloat
// Keep last 50 request details in cache, using LRU eviction
queryClient.setQueryDefaults(['requests', 'detail'], {
  staleTime: 1000 * 60 * 5, // 5 minutes - request details are stable
  gcTime: 1000 * 60 * 10, // 10 minutes - keep in cache longer for comparison feature
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
