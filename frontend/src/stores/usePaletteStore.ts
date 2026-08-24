import { create } from 'zustand'

interface PaletteState {
  open: boolean
  /** Recently run command ids, most recent first (max 5). */
  recent: string[]
  openPalette: () => void
  close: () => void
  toggle: () => void
  recordRun: (commandId: string) => void
}

const RECENT_KEY = 'reqly-palette-recent.v1'
const MAX_RECENT = 5

function readRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.flatMap((entry) =>
      entry === null || entry === undefined || String(entry) === '' ? [] : [String(entry)],
    )
  } catch {
    return []
  }
}

function writeRecent(recent: string[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recent))
  } catch {
    // storage unavailable — in-memory only
  }
}

export const usePaletteStore = create<PaletteState>()((set, get) => ({
  open: false,
  recent: readRecent(),
  openPalette: () => set({ open: true }),
  close: () => set({ open: false }),
  toggle: () => set((s) => ({ open: !s.open })),
  recordRun: (commandId) => {
    const recent = [commandId, ...get().recent.filter((id) => id !== commandId)].slice(
      0,
      MAX_RECENT,
    )
    set({ recent })
    writeRecent(recent)
  },
}))
