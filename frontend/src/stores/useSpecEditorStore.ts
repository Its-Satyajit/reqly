import { create } from "zustand";

interface SpecEditorState {
  content: string;
  selectedId: string;
  dirty: boolean;
  setContent: (c: string) => void;
  setSelected: (id: string) => void;
  markSaved: () => void;
}

export const useSpecEditorStore = create<SpecEditorState>((set) => ({
  content: "openapi: 3.1.0\ninfo:\n  title: Reqly API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: List users\n",
  selectedId: "info",
  dirty: false,
  setContent: (content) => set({ content, dirty: true }),
  setSelected: (selectedId) => set({ selectedId }),
  markSaved: () => set({ dirty: false }),
}));
