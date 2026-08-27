export const THEMES = [
  { id: "atlas-light" as const, label: "Atlas Light", appearance: "light" as const },
  { id: "atlas-dark" as const, label: "Atlas Dark", appearance: "dark" as const },
] as const;

export type ThemeId = (typeof THEMES)[number]["id"];
export type ThemePreference = ThemeId | "system";

export const DEFAULT_THEME: ThemeId = "atlas-dark";
export const STORAGE_KEY = "reqly-theme";

const THEME_IDS = new Set<string>(THEMES.map((t) => t.id));

export function isThemeId(v: string): v is ThemeId {
  return THEME_IDS.has(v);
}

export function isThemePreference(v: string): v is ThemePreference {
  return v === "system" || isThemeId(v);
}

export function resolveTheme(preference: ThemePreference, osDark: boolean): ThemeId {
  if (preference === "system") return osDark ? "atlas-dark" : "atlas-light";
  return preference;
}

export function appearanceFor(themeId: ThemeId): "light" | "dark" {
  const found = THEMES.find((t) => t.id === themeId);
  return found?.appearance ?? "dark";
}

export function nextPreference(current: ThemePreference): ThemePreference {
  if (current === "atlas-light") return "atlas-dark";
  if (current === "atlas-dark") return "system";
  return "atlas-light";
}

export function normalizeStored(value: string | null): ThemePreference | null {
  if (value === null) return null;
  if (value === "system" || isThemeId(value)) return value;
  if (value === "light") return "atlas-light";
  if (value === "dark") return "atlas-dark";
  return null;
}

export function applyThemeToDocument(themeId: ThemeId) {
  const root = document.documentElement;
  // data-theme is the token selector (one block per theme in tokens.css / index.css)
  root.setAttribute("data-theme", themeId);
  // Keep .dark mirror for Tailwind @custom-variant dark
  root.classList.toggle("dark", appearanceFor(themeId) === "dark");
}
