import { create } from 'zustand'
import type { AuthAdapter, AuthStatusData } from '../lib/auth'
import { fallbackAuthAdapter } from '../lib/auth'

interface AuthState {
  status: AuthStatusData | null
  loading: boolean
  error: string | null
  adapter: AuthAdapter

  setAdapter: (adapter: AuthAdapter) => void
  refresh: () => Promise<void>
  login: (config: Record<string, string>, flow: string) => Promise<void>
  logout: () => Promise<void>
  clearError: () => void
}

/**
 * OAuth token state for the workspace. `adapter` is the pluggable transport:
 * the Wails host injects the Go-core AuthService bridge; browser dev mode
 * falls back to fallbackAuthAdapter.
 */
export const useAuthStore = create<AuthState>((set, get) => ({
  status: null,
  loading: false,
  error: null,
  adapter: fallbackAuthAdapter,

  setAdapter: (adapter) => set({ adapter }),

  refresh: async () => {
    set({ loading: true, error: null })
    try {
      const status = await get().adapter.status()
      set({ status, loading: false })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : String(err),
      })
    }
  },

  login: async (config, flow) => {
    set({ loading: true, error: null })
    try {
      await get().adapter.login(config, flow)
      const status = await get().adapter.status()
      set({ status, loading: false })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : String(err),
      })
    }
  },

  logout: async () => {
    set({ loading: true, error: null })
    try {
      await get().adapter.logout()
      const status = await get().adapter.status()
      set({ status, loading: false })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : String(err),
      })
    }
  },

  clearError: () => set({ error: null }),
}))
