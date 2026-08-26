import { create } from "zustand";
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
  // No stored value: default to resolved dark/light based on OS but store as explicit? Use system? Spec says default? Use system? We'll default to system? But keep stable: map to theme id? Instead default to system so OS tracking works. However existing behavior defaulted to dark. Use atlas-dark as fallback per ADR unknown->default.
  // If no stored value, use system preference? Ticket says theme persists across restarts, rail cycles light→dark→system. Default unspecified. Use system if no storage? Let's default to system-aware but persisted as system would cause follow. Use DEFAULT_THEME as fallback for now.
  // To preserve "follows OS" expectation, default to "system" when nothing stored? Choose DEFAULT_THEME? We'll choose "system" is more spec-compliant. But test expects atlas-dark when OS dark and no stored.
  // That test will pass either way because system+dark => atlas-dark. Check: getInitialPreference returning "system" gives resolved atlas-dark. So both work.
  // Return "system" only if we want live tracking by default. Let's return system? The legacy defaulted to explicit. We'll keep explicit default for backward compat: resolve via OS to explicit id.
  return osDark() ? "atlas-dark" : "atlas-light";
  // Alternative: return "system" — would also resolve same but enable live tracking. Choose explicit to avoid surprise live re-skin for new users? Keep explicit.
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
  const handler = () => {
    const { theme } = useThemeStore.getState();
    if (theme === "system") {
      const resolved = resolveTheme("system", mql.matches);
      applyThemeToDocument(resolved);
      useThemeStore.setState({ resolvedTheme: resolved });
    }
  };
  // Support both modern and legacy
  if (typeof mql.addEventListener === "function") mql.addEventListener("change", handler);
  else mql.addListener(handler as unknown as (e: MediaQueryListEvent) => void);
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
