import { create } from "zustand"
import {
  fallbackHistoryAdapter,
  type HistoryAdapter,
  type HistoryEntry,
  type ReplayedResponse,
} from "../lib/history"

export const HISTORY_PAGE_SIZE = 50

/** How many recent entries the fuzzy search pools over. History is bounded
 * by retention, so this covers the practical search space without IPC per
 * keystroke. */
export const HISTORY_SEARCH_POOL = 500

interface HistoryState {
  adapter: HistoryAdapter
  entries: HistoryEntry[]
  /** Raw recent-entry pool the fuzzy search filters client-side. */
  pool: HistoryEntry[]
  poolLoaded: boolean
  loading: boolean
  error: string | null
  /** The status-class filter of the currently loaded page ("" = all). */
  statusFilter: string
  /** Search query backing the currently loaded page (FTS when non-empty). */
  query: string
  replayed: ReplayedResponse | null
  setAdapter(adapter: HistoryAdapter): void
  /** Load one page. offset is derived from the caller's page number. */
  load(opts?: { limit?: number; offset?: number; status?: string; query?: string }): Promise<void>
  loadPool(): Promise<void>
  refresh(): Promise<void>
  clear(env: string | null): Promise<void>
  replay(id: string): Promise<void>
  dismissReplay(): void
}

export const useHistoryStore = create<HistoryState>((set, get) => ({
  adapter: fallbackHistoryAdapter,
  entries: [],
  pool: [],
  poolLoaded: false,
  loading: false,
  error: null,
  statusFilter: "",
  query: "",
  replayed: null,

  setAdapter(adapter) {
    set({ adapter })
  },

  async load(opts = {}) {
    const limit = opts.limit ?? HISTORY_PAGE_SIZE
    const offset = opts.offset ?? 0
    const status = opts.status ?? get().statusFilter
    const query = opts.query ?? get().query
    set({ loading: true, statusFilter: status, query })
    try {
      if (!get().poolLoaded) await get().loadPool()
      const entries = await get().adapter.list(limit, offset, status, "")
      set({ entries, error: null })
    } catch (err) {
      set({ error: err instanceof Error ? err.message : String(err) })
    } finally {
      set({ loading: false })
    }
  },

  async refresh() {
    await get().load({ offset: 0 })
  },

  async loadPool() {
    try {
      const pool = await get().adapter.list(HISTORY_SEARCH_POOL, 0, "", "")
      set({ pool, poolLoaded: true })
    } catch (err) {
      set({ error: err instanceof Error ? err.message : String(err) })
    }
  },

  async clear(env) {
    try {
      await get().adapter.clear(env)
      await get().refresh()
    } catch (err) {
      set({ error: err instanceof Error ? err.message : String(err) })
    }
  },

  async replay(id) {
    try {
      const response = await get().adapter.replay(id)
      if (!response) {
        set({ error: "Replay returned no response." })
        return
      }
      set({ replayed: response, error: null })
    } catch (err) {
      set({ error: err instanceof Error ? err.message : String(err) })
    }
  },

  dismissReplay() {
    set({ replayed: null })
  },
}))
