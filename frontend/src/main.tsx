import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { LoginGate } from './components/LoginGate'
import './index.css'

const rootElement = document.getElementById('root')
if (!rootElement) throw new Error('Root element not found')

createRoot(rootElement).render(
  <React.StrictMode>
    <LoginGate>
      <App />
    </LoginGate>
  </React.StrictMode>,
)
