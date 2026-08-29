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

// --- Custom theme sharing (M67) — Git-native, validated, shareable ---

export interface CustomTheme {
  id: string;
  label: string;
  appearance: "light" | "dark";
  tokens?: Record<string, string>;
}

const ID_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export function validateCustomTheme(t: CustomTheme): string | null {
  if (!t.id) return "id is required";
  if (!ID_RE.test(t.id)) return `id ${JSON.stringify(t.id)} must be kebab-case [a-z0-9-]`;
  if (!t.label || !t.label.trim()) return "label is required";
  if (t.appearance !== "light" && t.appearance !== "dark") return "appearance must be 'light' or 'dark'";
  return null;
}

function parseSimpleYaml(yaml: string): CustomTheme {
  const lines = yaml.split("\n");
  const result: Record<string, unknown> = {};
  let currentKey: string | null = null;
  let tokens: Record<string, string> = {};
  let inTokens = false;
  for (const raw of lines) {
    const line = raw.trimEnd();
    if (!line.trim() || line.trim().startsWith("#")) continue;
    if (inTokens) {
      const m = line.match(/^\s{2,}(\S+):\s*"?([^"]*)"?\s*$/);
      if (m) {
        tokens[m[1]] = m[2];
        continue;
      }
      inTokens = false;
    }
    const kv = line.match(/^(\S+):\s*"?([^"]*)"?\s*$/);
    if (!kv) continue;
    const k = kv[1];
    const v = kv[2];
    if (k === "tokens") {
      inTokens = true;
      tokens = {};
      result["tokens"] = tokens;
      currentKey = "tokens";
    } else {
      result[k] = v;
      currentKey = k;
    }
  }
  // Handle tokens already captured
  if (inTokens && currentKey === "tokens") {
    result["tokens"] = tokens;
  }
  return {
    id: String(result["id"] ?? ""),
    label: String(result["label"] ?? ""),
    appearance: String(result["appearance"] ?? "") as "light" | "dark",
    tokens: (result["tokens"] as Record<string, string>) ?? undefined,
  };
}

export function parseCustomTheme(input: string): CustomTheme {
  const trimmed = input.trim();
  // Try JSON first
  if (trimmed.startsWith("{")) {
    try {
      const parsed = JSON.parse(trimmed) as CustomTheme;
      const err = validateCustomTheme(parsed);
      if (err) throw new Error(err);
      return parsed;
    } catch (e) {
      throw new Error(`parse theme: ${(e as Error).message}`);
    }
  }
  // Fallback to simple YAML
  try {
    const parsed = parseSimpleYaml(trimmed);
    const err = validateCustomTheme(parsed);
    if (err) throw new Error(err);
    return parsed;
  } catch (e) {
    throw new Error(`parse theme: ${(e as Error).message}`);
  }
}

export function customThemeToCSS(t: CustomTheme): string {
  const err = validateCustomTheme(t);
  if (err) throw new Error(err);
  const tokens = t.tokens ?? {};
  const parts: string[] = [];
  for (const [k, v] of Object.entries(tokens)) {
    const key = k.startsWith("--") ? k.slice(2) : k;
    parts.push(` --${key}: ${v};`);
  }
  return `[data-theme="${t.id}"] {${parts.join("")} }`;
}
