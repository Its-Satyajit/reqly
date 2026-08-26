import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import '../index.css'

// Dev-only browser demo (`vite` + `?demo`): swap the Wails bridge seams for
// in-memory demo adapters so the full shell renders for design previews.
// Tree-shaken from production builds — the gate is a compile-time constant.
const bootstrap = import.meta.env.DEV &&
new URLSearchParams(window.location.search).has('demo')
  ? Promise.all([import('../lib/demo-workspace'), import('../stores')]).then(
      ([demo, stores]) => {
        stores.useWorkspaceBootstrapStore.getState().setAdapter(demo.demoBootstrapAdapter)
        stores.useWorkspaceStore.getState().setWorkspaceAdapter(demo.demoCollectionsAdapter)
        stores.useWorkspaceStore.getState().setEnvAdapter(demo.demoEnvAdapter)
        stores.useHistoryStore.getState().setAdapter(demo.demoHistoryAdapter)
      },
    )
  : undefined

void Promise.resolve(bootstrap).then(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
})
