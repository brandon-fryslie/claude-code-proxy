import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { ConversationsPage } from './pages/Conversations'
import { HomePage } from './pages/Home'
import { ConfigurationPage } from './pages/Configuration'
import { ProjectsPage } from './pages/Projects'
import SessionDataPage from './pages/SessionData'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
      refetchOnWindowFocus: false,
      gcTime: 1000 * 60 * 10, // 10 minutes
    },
  },
})

// Simple URL-based routing
function App() {
  const path = window.location.pathname

  // Route based on path
  if (path === '/cc-viz/conversations' || path === '/cc-viz/conversations/') {
    return <ConversationsPage />
  }
  if (path === '/cc-viz/configuration' || path === '/cc-viz/configuration/') {
    return <ConfigurationPage />
  }
  if (path === '/cc-viz/projects' || path === '/cc-viz/projects/') {
    return <ProjectsPage />
  }
  if (path === '/cc-viz/session-data' || path === '/cc-viz/session-data/') {
    return <SessionDataPage />
  }

  // Default to home page for /cc-viz/ or /cc-viz
  return <HomePage />
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>,
)
