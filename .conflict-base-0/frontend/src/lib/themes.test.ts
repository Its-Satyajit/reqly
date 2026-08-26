import { describe, expect, it } from 'vitest'

import { THEMES, resolveTheme, type ThemePreference } from './themes'

describe('theme registry', () => {
  it('ships atlas-dark and atlas-light', () => {
    expect(THEMES.map((t) => t.id).sort()).toEqual(['atlas-dark', 'atlas-light'])
  })

  it('exposes an appearance per theme', () => {
    for (const t of THEMES) {
      expect(['light', 'dark']).toContain(t.appearance)
      expect(t.label.length).toBeGreaterThan(0)
    }
  })
})

describe('resolveTheme', () => {
  it('resolves an explicit theme id to itself', () => {
    expect(resolveTheme('atlas-light', false)).toBe('atlas-light')
    expect(resolveTheme('atlas-dark', true)).toBe('atlas-dark')
  })

  it('resolves system to atlas-dark when the OS prefers dark', () => {
    expect(resolveTheme('system', true)).toBe('atlas-dark')
  })

  it('resolves system to atlas-light when the OS prefers light', () => {
    expect(resolveTheme('system', false)).toBe('atlas-light')
  })

  it('falls back to atlas-dark for unknown values (corrupt storage)', () => {
    // SAFETY: deliberately invalid value exercising the corrupt-storage fallback.
    expect(resolveTheme('bogus' as ThemePreference, false)).toBe('atlas-dark')
  })
})
