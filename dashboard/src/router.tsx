import { createRouter, createRootRoute, createRoute, Outlet, useNavigate } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { Sidebar, GlobalDatePicker } from '@/components/layout'
import { DateRangeProvider } from '@/lib/DateRangeContext'
import { DashboardPage } from '@/pages/Dashboard'
import { RequestsPage } from '@/pages/Requests'
import { UsagePage } from '@/pages/Usage'
import { PerformancePage } from '@/pages/Performance'
import { RoutingPage } from '@/pages/Routing'
import { SettingsPage } from '@/pages/Settings'

// Root route with layout
const rootRoute = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  const navigate = useNavigate()
  const [activeItem, setActiveItem] = useState('dashboard')

  // Sync active item with current route
  useEffect(() => {
    const path = window.location.pathname
    const item = path.slice(1) || 'dashboard'
    setActiveItem(item)
  }, [])

  const handleItemSelect = (id: string) => {
    setActiveItem(id)
    navigate({ to: `/${id === 'dashboard' ? '' : id}` })
  }

  return (
    <DateRangeProvider>
      <div className="flex h-screen overflow-hidden bg-[var(--color-bg-primary)]">
        <Sidebar activeItem={activeItem} onItemSelect={handleItemSelect} />
        <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
          {/* Global date picker header */}
          <div className="flex items-center justify-end h-10 px-4 border-b border-[var(--color-border)] bg-[var(--color-bg-tertiary)]">
            <GlobalDatePicker />
          </div>
          <Outlet />
        </main>
      </div>
    </DateRangeProvider>
  )
}

// Define routes
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: DashboardPage,
})

const requestsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/requests',
  component: RequestsPage,
})

const usageRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/usage',
  component: UsagePage,
})

const performanceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/performance',
  component: PerformancePage,
})

const routingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/routing',
  component: RoutingPage,
})

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
})

// Create route tree
const routeTree = rootRoute.addChildren([
  indexRoute,
  dashboardRoute,
  requestsRoute,
  usageRoute,
  performanceRoute,
  routingRoute,
  settingsRoute,
])

// Create router
export const router = createRouter({ routeTree })

// Type declaration for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
