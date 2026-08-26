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
  setOpen: (open: boolean) => void;
  setQuery: (q: string) => void;
  registerCommand: (c: Command) => void;
  unregisterCommand: (id: string) => void;
  registerProvider: (p: DataProvider) => void;
  filtered: () => (Command & { kind?: string })[];
}

export const useCommandPaletteStore = create<PaletteState>((set, get) => ({
  open: false,
  query: "",
  commands: [],
  providers: [],
  setOpen: (open) => set({ open }),
  setQuery: (query) => set({ query }),
  registerCommand: (c) => set((s) => ({ commands: [...s.commands.filter((x) => x.id !== c.id), c] })),
  unregisterCommand: (id) => set((s) => ({ commands: s.commands.filter((c) => c.id !== id) })),
  registerProvider: (p) =>
    set((s) => ({ providers: [...s.providers.filter((x) => x.id !== p.id), p] })),
  filtered: () => {
    const { query, commands, providers } = get();
    const dataItems = providers.flatMap((p) => p.getItems().map((i) => ({ ...i, hint: i.kind })));
    const all = [...commands.map((c) => ({ ...c, kind: "command" })), ...dataItems.map((d) => ({ id: d.id, title: d.title, hint: d.hint, keywords: d.kind, run: d.run, kind: d.kind }))];
    if (!query.trim()) return all.slice(0, 20);
    const fuse = new Fuse(all, { keys: ["title", "keywords"], threshold: 0.4 });
    return fuse.search(query).map((r) => r.item).slice(0, 20);
  },
}));
