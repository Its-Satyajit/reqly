import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@reqly/frontend'
import { initRequestBridge } from './bridge'

initRequestBridge()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)