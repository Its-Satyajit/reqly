import { create } from "zustand";

import type { SpecDiagnostic } from "#lib/specTree";
import { diagnosticsForSpec } from "#lib/specTree";

interface SpecEditorState {
  content: string;
  selectedId: string;
  dirty: boolean;
  filePath: string;
  diagnostics: SpecDiagnostic[];
  selectedOps: Set<string>;
  generateWarnings: string[];
  setContent: (c: string) => void;
  setSelected: (id: string) => void;
  markSaved: () => void;
  revalidate: () => void;
  setFilePath: (p: string) => void;
  loadContent: (content: string, filePath: string) => void;
  toggleOp: (id: string) => void;
  setGenerateWarnings: (w: string[]) => void;
}

export const useSpecEditorStore = create<SpecEditorState>((set, get) => ({
  content: "openapi: 3.1.0\ninfo:\n  title: Reqly API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: List users\n",
  selectedId: "info",
  dirty: false,
  filePath: "openapi.yaml",
  diagnostics: diagnosticsForSpec(
    "openapi: 3.1.0\ninfo:\n  title: Reqly API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: List users\n",
  ),
  selectedOps: new Set<string>(),
  generateWarnings: [],
  setContent: (content) =>
    set({
      content,
      dirty: true,
      diagnostics: diagnosticsForSpec(content),
    }),
  setSelected: (selectedId) => set({ selectedId }),
  markSaved: () => set({ dirty: false }),
  revalidate: () => set({ diagnostics: diagnosticsForSpec(get().content) }),
  setFilePath: (filePath) => set({ filePath }),
  loadContent: (content, filePath) =>
    set({ content, filePath, dirty: false, diagnostics: diagnosticsForSpec(content) }),
  toggleOp: (id) =>
    set((s) => {
      const next = new Set(s.selectedOps);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return { selectedOps: next };
    }),
  setGenerateWarnings: (generateWarnings) => set({ generateWarnings }),
}));
