import { create } from "zustand";
import Fuse from "fuse.js";

export interface Command {
  id: string;
  title: string;
  hint?: string;
  keywords?: string;
  run: () => void;
}

export interface DataProvider {
  id: string;
  kind: string;
  getItems: () => { id: string; title: string; kind: string; run: () => void }[];
}

interface PaletteState {
  open: boolean;
  query: string;
  commands: Command[];
  providers: DataProvider[];
  recent: string[];
  setOpen: (open: boolean) => void;
  setQuery: (q: string) => void;
  registerCommand: (c: Command) => void;
  unregisterCommand: (id: string) => void;
  registerProvider: (p: DataProvider) => void;
  recordRun: (id: string) => void;
}

const RECENT_KEY = "reqly-palette-recent.v1";
const MAX_RECENT = 5;
function readRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((v): v is string => typeof v === "string" && v !== "");
  } catch {
    return [];
  }
}
function writeRecent(recent: string[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recent));
  } catch {}
}

export const useCommandPaletteStore = create<PaletteState>((set) => ({
  open: false,
  query: "",
  commands: [],
  providers: [],
  recent: readRecent(),
  setOpen: (open) => set({ open }),
  setQuery: (query) => set({ query }),
  registerCommand: (c) => set((s) => ({ commands: [...s.commands.filter((x) => x.id !== c.id), c] })),
  unregisterCommand: (id) => set((s) => ({ commands: s.commands.filter((c) => c.id !== id) })),
  registerProvider: (p) =>
    set((s) => ({ providers: [...s.providers.filter((x) => x.id !== p.id), p] })),
  recordRun: (id) =>
    set((s) => {
      const recent = [id, ...s.recent.filter((x) => x !== id)].slice(0, MAX_RECENT);
      writeRecent(recent);
      return { recent };
    }),
}));

export function getFilteredResults(query: string, commands: Command[], providers: DataProvider[]): (Command & { kind?: string })[] {
  const dataItems = providers.flatMap((p) => {
    try {
      return p.getItems().map((i) => ({ ...i, hint: i.kind }));
    } catch {
      return [];
    }
  });
  const all = [
    ...commands.map((c) => ({ ...c, kind: "command" })),
    ...dataItems.map((d) => ({ id: d.id, title: d.title, hint: d.hint, keywords: d.kind, run: d.run, kind: d.kind })),
  ];
  const recent = (() => {
    try {
      // SAFETY: RECENT_KEY is our own JSON string array (writeRecent), validated via Array.isArray in readRecent
      return (JSON.parse(localStorage.getItem(RECENT_KEY) ?? "[]") as string[]) ?? [];
    } catch {
      return [];
    }
  })();
  const recentFirst = (a: (typeof all)[number], b: (typeof all)[number]) =>
    Number(recent.includes(b.id)) - Number(recent.includes(a.id));
  if (!query.trim()) return [...all].sort(recentFirst).slice(0, 20);
  const fuse = new Fuse(all, { keys: ["title", "keywords"], threshold: 0.4 });
  const results = fuse.search(query).map((r) => r.item);
  return [...results].sort(recentFirst).slice(0, 20);
}

export function groupByHint(items: (Command & { kind?: string })[]): Map<string, (Command & { kind?: string })[]> {
  const groups = new Map<string, (Command & { kind?: string })[]>();
  const order = ["Navigation", "Environment", "Collection", "History", "Theme", "command"];
  for (const item of items) {
    const key = item.hint ?? item.kind ?? "command";
    // Normalize hint to group: Navigation for nav-*, Theme for theme, etc.
    const group = key === "Navigation" || key === "Theme" || key === "Environment" || key === "Collection" || key === "History" ? key : item.id.startsWith("nav-") ? "Navigation" : item.id.startsWith("theme-") ? "Theme" : item.id.startsWith("env-") || item.id.startsWith("col-") || item.id.startsWith("hist-") ? item.kind ?? "command" : "command";
    const normalized = order.includes(group) ? group : "command";
    const list = groups.get(normalized) ?? [];
    list.push(item);
    groups.set(normalized, list);
  }
  // Return in consistent order
  const ordered = new Map<string, (Command & { kind?: string })[]>();
  for (const k of order) if (groups.has(k)) ordered.set(k, groups.get(k)!);
  for (const [k, v] of groups) if (!ordered.has(k)) ordered.set(k, v);
  return ordered;
}
