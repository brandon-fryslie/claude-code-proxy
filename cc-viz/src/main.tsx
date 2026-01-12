import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { ConversationsPage } from './pages/Conversations'
import './index.css'

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <div className="h-screen flex flex-col bg-[var(--color-bg-primary)]">
        <ConversationsPage />
      </div>
    </QueryClientProvider>
  </React.StrictMode>,
)
