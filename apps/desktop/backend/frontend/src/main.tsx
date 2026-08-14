import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@reqly/frontend'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)