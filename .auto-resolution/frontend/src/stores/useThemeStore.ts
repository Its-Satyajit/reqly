import { create } from 'zustand'

import {
  firstWithAppearance,
  resolveTheme,
  THEMES,
  themeById,
  type ThemeAppearance,
  type ThemeId,
  type ThemePreference,
} from '#lib/themes'

const STORAGE_KEY = 'reqly-theme'

interface MediaQueryLike {
  matches: boolean
  addEventListener(type: string, cb: () => void): void
  removeEventListener(type: string, cb: () => void): void
}

export interface ThemeControllerOptions {
  /** OS dark preference at init (test seam). */
  osPrefersDark?: boolean
  /** matchMedia handle override (test seam); defaults to prefers-color-scheme. */
  mediaQuery?: MediaQueryLike
  onDispose?: () => void
}

export interface ThemeController {
  get(): { preference: ThemePreference; theme: ThemeId }
  set(preference: ThemePreference): void
  toggle(): void
  dispose(): void
}

function isThemePreference(value: string | null): value is ThemePreference {
  return value === 'system' || THEMES.some((t) => t.id === value)
}

function readStoredPreference(): ThemePreference {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (isThemePreference(stored)) return stored
  } catch {
    // storage unavailable — fall through to system
  }
  return 'system'
}

function safeMatchMedia(query: string): MediaQueryLike | undefined {
  try {
    return window.matchMedia(query)
  } catch {
    return undefined
  }
}

function osDark(mql?: MediaQueryLike): boolean {  if (mql) return mql.matches
  if (globalThis.window === undefined) return true
  try {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  } catch {
    return true
  }
}

export function applyThemeToDom(theme: ThemeId) {
  document.documentElement.dataset.theme = theme
  document.documentElement.classList.toggle('dark', themeById(theme).appearance === 'dark')
}

/**
 * Framework-agnostic theme state machine so the behavior is testable without
 * React. The zustand store below is a thin reactive wrapper.
 */
export function createThemeController(options: ThemeControllerOptions = {}): ThemeController {
  const mql =
    options.mediaQuery ??
    (globalThis.window === undefined
      ? undefined
      : safeMatchMedia('(prefers-color-scheme: dark)'))

  const dark = () => (options.osPrefersDark !== undefined ? options.osPrefersDark : osDark(mql))

  let preference = readStoredPreference()
  let theme = resolveTheme(preference, dark())
  applyThemeToDom(theme)

  const sync = (next: ThemePreference) => {
    preference = next
    theme = resolveTheme(next, dark())
    applyThemeToDom(theme)
    try {
      localStorage.setItem(STORAGE_KEY, next)
    } catch {
      // storage unavailable — in-memory only
    }
  }

  const onOsChange = () => {
    if (preference !== 'system') return
    theme = resolveTheme('system', dark())
    applyThemeToDom(theme)
  }
  mql?.addEventListener('change', onOsChange)

  return {
    get: () => ({ preference, theme }),
    set: sync,
    toggle: () =>
      sync(
        themeById(theme).appearance === 'dark'
          ? firstWithAppearance('light')
          : firstWithAppearance('dark'),
      ),
    dispose: () => {
      mql?.removeEventListener('change', onOsChange)
      options.onDispose?.()
    },
  }
}

interface ThemeState {
  preference: ThemePreference
  theme: ThemeId
  appearance: ThemeAppearance
  setTheme: (preference: ThemePreference) => void
  toggleTheme: () => void
}

export const useThemeStore = create<ThemeState>((set) => {
  const controller = createThemeController()
  const pull = () =>
    set({
      ...controller.get(),
      appearance: themeById(controller.get().theme).appearance,
    })
  const initial = { ...controller.get(), appearance: themeById(controller.get().theme).appearance }

  return {
    ...initial,
    setTheme: (next) => {
      controller.set(next)
      pull()
    },
    toggleTheme: () => {
      controller.toggle()
      pull()
    },
  }
})
