import { create } from "zustand";
import { isFunction } from "#lib/typeGuards";
import {
  STORAGE_KEY,
  DEFAULT_THEME,
  type ThemePreference,
  type ThemeId,
  resolveTheme,
  nextPreference,
  normalizeStored,
  applyThemeToDocument,
} from "#lib/themes";

function osDark(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) return true;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

function getInitialPreference(): ThemePreference {
  if (typeof window === "undefined") return DEFAULT_THEME;
  const stored = normalizeStored(window.localStorage.getItem(STORAGE_KEY));
  if (stored) return stored;
  return osDark() ? "atlas-dark" : "atlas-light";
}

function getInitialResolved(pref: ThemePreference): ThemeId {
  return resolveTheme(pref, osDark());
}

interface ThemeState {
  theme: ThemePreference;
  resolvedTheme: ThemeId;
  setTheme: (theme: ThemePreference) => void;
  cycleTheme: () => void;
  toggleTheme: () => void;
}

const initialPref = getInitialPreference();
const initialResolved = getInitialResolved(initialPref);

// Apply on load
if (typeof document !== "undefined") {
  applyThemeToDocument(initialResolved);
}

// Live OS tracking when preference is system
if (typeof window !== "undefined" && window.matchMedia) {
  const mql = window.matchMedia("(prefers-color-scheme: dark)");
  const handler = (e: MediaQueryListEvent | MediaQueryList) => {
    const { theme } = useThemeStore.getState();
    if (theme === "system") {
      const resolved = resolveTheme("system", e.matches);
      applyThemeToDocument(resolved);
      useThemeStore.setState({ resolvedTheme: resolved });
    }
  };
  // Support both modern and legacy
  if (isFunction(mql.addEventListener)) {
    mql.addEventListener("change", handler);
  } else if (isFunction(mql.addListener)) {
    mql.addListener(handler);
  }
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: initialPref,
  resolvedTheme: initialResolved,
  setTheme: (theme) => {
    const resolved = resolveTheme(theme, osDark());
    applyThemeToDocument(resolved);
    if (typeof window !== "undefined") window.localStorage.setItem(STORAGE_KEY, theme);
    set({ theme, resolvedTheme: resolved });
  },
  cycleTheme: () => get().setTheme(nextPreference(get().theme)),
  toggleTheme: () => get().cycleTheme(),
}));
