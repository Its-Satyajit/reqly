import { beforeEach, describe, expect, it, vi } from 'vitest'

import { createThemeController } from './useThemeStore'

// jsdom has no matchMedia; tests that need controlled behavior pass their own.
vi.stubGlobal('matchMedia', (query: string) => ({
  matches: query.includes('dark'),
  addEventListener: () => {},
  removeEventListener: () => {},
}))

function freshDom() {
  const storage = new Map<string, string>()
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => storage.get(k) ?? null,
    setItem: (k: string, v: string) => void storage.set(k, v),
  })
  document.documentElement.removeAttribute('data-theme')
  document.documentElement.classList.remove('dark')
  return { storage }
}

function makeMql(matches: () => boolean, listeners = new Set<() => void>()) {
  return {
    get matches() {
      return matches()
    },
    addEventListener: (_: string, cb: () => void) => listeners.add(cb),
    removeEventListener: (_: string, cb: () => void) => listeners.delete(cb),
  }
}

describe('theme controller', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('applies data-theme and the dark class on set', () => {
    freshDom()
    const c = createThemeController({ osPrefersDark: true })
    c.set('atlas-dark')
    expect(document.documentElement.dataset.theme).toBe('atlas-dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('removes the dark class for light themes', () => {
    freshDom()
    const c = createThemeController({ osPrefersDark: false })
    c.set('atlas-light')
    expect(document.documentElement.dataset.theme).toBe('atlas-light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('persists the preference', () => {
    const { storage } = freshDom()
    const c = createThemeController({ osPrefersDark: false })
    c.set('atlas-dark')
    expect(storage.get('reqly-theme')).toBe('atlas-dark')
  })

  it('restores the persisted preference on init', () => {
    const { storage } = freshDom()
    storage.set('reqly-theme', 'atlas-light')
    const c = createThemeController({ osPrefersDark: true })
    expect(c.get().preference).toBe('atlas-light')
    expect(c.get().theme).toBe('atlas-light')
  })

  it('resolves system preference at init when nothing is stored', () => {
    freshDom()
    const c = createThemeController({ osPrefersDark: false })
    expect(c.get().preference).toBe('system')
    expect(c.get().theme).toBe('atlas-light')
  })

  it('tracks live OS changes while preference is system', () => {
    freshDom()
    let osDark = true
    const listeners = new Set<() => void>()
    const c = createThemeController({
      mediaQuery: makeMql(() => osDark, listeners),
    })
    expect(c.get().theme).toBe('atlas-dark')
    osDark = false
    for (const cb of listeners) cb()
    expect(c.get().theme).toBe('atlas-light')
    expect(document.documentElement.dataset.theme).toBe('atlas-light')
  })

  it('stops tracking OS changes once an explicit theme is set', () => {
    freshDom()
    const listeners = new Set<() => void>()
    let osDark = true
    const c = createThemeController({
      mediaQuery: makeMql(() => osDark, listeners),
    })
    c.set('atlas-light')
    osDark = true
    for (const cb of listeners) cb()
    expect(c.get().theme).toBe('atlas-light')
  })

  it('toggle flips between light and dark themes', () => {
    freshDom()
    const c = createThemeController({ osPrefersDark: true })
    c.set('atlas-dark')
    c.toggle()
    expect(c.get().theme).toBe('atlas-light')
    c.toggle()
    expect(c.get().theme).toBe('atlas-dark')
  })
})
