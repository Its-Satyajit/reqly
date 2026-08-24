import { create } from 'zustand'

export type InspectorTabId = string

interface ShellState {
  /** Right-hand inspector mount point; views populate content per tab. */
  inspectorOpen: boolean
  inspectorTab: InspectorTabId | null
  openInspector: (tab?: InspectorTabId) => void
  closeInspector: () => void
  toggleInspector: () => void
}

const OPEN_KEY = 'reqly-shell-inspector-open'
const TAB_KEY = 'reqly-shell-inspector-tab'

/** Seed for the store's initial state, so persistence is testable directly. */
export function initialShellState(): Pick<ShellState, 'inspectorOpen' | 'inspectorTab'> {
  return {
    inspectorOpen: readOpen(),
    inspectorTab: readTab(),
  }
}

function readOpen(): boolean {
  try {
    return localStorage.getItem(OPEN_KEY) === '1'
  } catch {
    return false
  }
}

function readTab(): InspectorTabId | null {
  try {
    return localStorage.getItem(TAB_KEY)
  } catch {
    return null
  }
}

function writeOpen(open: boolean) {
  try {
    if (open) localStorage.setItem(OPEN_KEY, '1')
    else localStorage.removeItem(OPEN_KEY)
  } catch {
    // storage unavailable — in-memory only
  }
}

function writeTab(tab: InspectorTabId | null) {
  try {
    if (tab === null) localStorage.removeItem(TAB_KEY)
    else localStorage.setItem(TAB_KEY, tab)
  } catch {
    // storage unavailable — in-memory only
  }
}

export const useShellStore = create<ShellState>()((set, get) => ({
  ...initialShellState(),
  openInspector: (tab) => {
    const next = tab ?? get().inspectorTab ?? 'default'
    set({ inspectorOpen: true, inspectorTab: next })
    writeOpen(true)
    writeTab(next)
  },
  closeInspector: () => {
    set({ inspectorOpen: false })
    writeOpen(false)
  },
  toggleInspector: () => {
    const open = !get().inspectorOpen
    set({ inspectorOpen: open })
    writeOpen(open)
    if (open) writeTab(get().inspectorTab)
  },
}))
