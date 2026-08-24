import { create } from "zustand";
import {
  fallbackDocsAdapter,
  type DocsAdapter,
  type DocsResultView,
} from "../lib/docs";

interface DocsState {
  adapter: DocsAdapter
  /** Selected collection names; empty array = whole workspace. */
  selected: string[]
  outName: string
  busy: boolean
  error: string | null
  result: DocsResultView | null
  activeFile: string | null
  setAdapter(adapter: DocsAdapter): void
  toggleCollection(name: string): void
  selectAll(): void
  setOutName(outName: string): void
  setActiveFile(name: string): void
  generate(): Promise<void>
}

export const useDocsStore = create<DocsState>((set, get) => ({
  adapter: fallbackDocsAdapter,
  selected: [],
  outName: "",
  busy: false,
  error: null,
  result: null,
  activeFile: null,

  setAdapter(adapter) {
    set({ adapter });
  },

  toggleCollection(name) {
    set((s) => ({
      selected: s.selected.includes(name)
        ? s.selected.filter((c) => c !== name)
        : [...s.selected, name],
    }));
  },

  selectAll() {
    set({ selected: [] });
  },

  setOutName(outName) {
    set({ outName });
  },

  setActiveFile(name) {
    set({ activeFile: name });
  },

  async generate() {
    const s = get();
    set({ busy: true, error: null });
    try {
      const result = await s.adapter.generate({
        collections: s.selected.length > 0 ? s.selected : undefined,
        outName: s.outName.trim() === "" ? undefined : s.outName.trim(),
      });
      set({
        result,
        busy: false,
        activeFile: result.files.find((f) => f.name === "index.md")?.name ??
          result.files[0]?.name ?? null,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : String(err),
        busy: false,
      });
    }
  },
}));
