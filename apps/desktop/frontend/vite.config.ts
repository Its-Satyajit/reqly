import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import wails from '@wailsio/runtime/plugins/vite'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(root, '../../..')

export default defineConfig({
  plugins: [react(), tailwindcss(), wails('./bindings')],
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
    fs: {
      // Allow Vite to serve the shared `frontend` package (imported via the
      // npm workspace symlink) and the Wails dev server.
      allow: [repoRoot],
    },
  },
  optimizeDeps: {
    exclude: ['@reqly/frontend'],
  },
})