import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@api-client/frontend'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)