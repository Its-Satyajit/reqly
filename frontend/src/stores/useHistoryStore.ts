import { create } from "zustand"
import { fallbackHistoryAdapter, type HistoryAdapter, type HistoryEntry } from "../lib/history"

interface HistoryState {
  adapter: HistoryAdapter
  entries: HistoryEntry[]
  loading: boolean
  setAdapter(adapter: HistoryAdapter): void
  refresh(limit?: number): Promise<void>
}

export const useHistoryStore = create<HistoryState>((set, get) => ({
  adapter: fallbackHistoryAdapter,
  entries: [],
  loading: false,
  setAdapter(adapter) { set({ adapter }) },
  async refresh(limit = 50) {
    set({ loading: true })
    try {
      const entries = await get().adapter.list(limit, 0, "", "")
      set({ entries })
    } finally { set({ loading: false }) }
  },
}))
