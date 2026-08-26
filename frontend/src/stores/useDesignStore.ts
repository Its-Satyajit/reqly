import { create } from 'zustand'

import { designById, DESIGNS, resolveDesign, type DesignId } from '#lib/designs'

const STORAGE_KEY = 'reqly-design'

function readStoredPreference(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY)
  } catch {
    return null
  }
}

export function applyDesignToDom(design: DesignId) {
  document.documentElement.dataset.design = design
}

/**
 * The design axis is purely visual: switching only swaps the `data-design`
 * attribute, so every store (tabs, drafts, environment, history) is untouched.
 */
export interface DesignState {
  design: DesignId
  label: string
  setDesign: (design: DesignId) => void
}

export const useDesignStore = create<DesignState>((set) => {
  const initial = resolveDesign(readStoredPreference())
  applyDesignToDom(initial)
  return {
    design: initial,
    label: designById(initial).label,
    setDesign: (next: DesignId) => {
      applyDesignToDom(next)
      try {
        localStorage.setItem(STORAGE_KEY, next)
      } catch {
        // storage unavailable — in-memory only
      }
      set({ design: next, label: designById(next).label })
    },
  }
})

export { DESIGNS }
