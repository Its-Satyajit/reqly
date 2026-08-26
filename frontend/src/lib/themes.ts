export type ThemeAppearance = 'light' | 'dark'

export interface ThemeDef {
  id: string
  label: string
  appearance: ThemeAppearance
}

/** Open registry: a new theme is a token set in frontend/src/styles/tokens.css plus one entry here. */
export const THEMES: readonly ThemeDef[] = [
  { id: 'atlas-light', label: 'Atlas Light', appearance: 'light' },
  { id: 'atlas-dark', label: 'Atlas Dark', appearance: 'dark' },
] as const

export type ThemeId = (typeof THEMES)[number]['id']

export type ThemePreference = ThemeId | 'system'

export const DEFAULT_THEME: ThemeId = 'atlas-dark'

function isThemeId(value: string): value is ThemeId {
  return THEMES.some((t) => t.id === value)
}

export function resolveTheme(
  preference: ThemePreference,
  osPrefersDark: boolean,
): ThemeId {
  if (preference === 'system') {
    return osPrefersDark ? DEFAULT_THEME : firstWithAppearance('light')
  }
  return isThemeId(preference) ? preference : DEFAULT_THEME
}

function firstWithAppearance(appearance: ThemeAppearance): ThemeId {
  return THEMES.find((t) => t.appearance === appearance)?.id ?? DEFAULT_THEME
}

export { firstWithAppearance }

export function themeById(id: ThemeId): ThemeDef {
  return THEMES.find((t) => t.id === id) ?? THEMES[0]
}
