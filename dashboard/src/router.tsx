import { createRouter, createRootRoute, createRoute, Outlet, useNavigate } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { Sidebar } from '@/components/layout/Sidebar'
import { DashboardPage } from '@/pages/Dashboard'
import { RequestsPage } from '@/pages/Requests'
import { ConversationsPage } from '@/pages/Conversations'
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
    <div className="flex h-screen overflow-hidden bg-[var(--color-bg-primary)]">
      <Sidebar activeItem={activeItem} onItemSelect={handleItemSelect} />
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Outlet />
      </main>
    </div>
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

const conversationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/conversations',
  component: ConversationsPage,
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
  conversationsRoute,
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
