import { create } from "zustand";
import {
  CATEGORIES,
  type RequestTemplate,
  type TemplateCategory,
} from "#lib/templates";

interface TemplateState {
  selectedCategory: TemplateCategory["id"] | null;
  selectedTemplateId: string | null;
  customTemplates: RequestTemplate[];
  selectCategory(id: TemplateCategory["id"] | null): void;
  selectTemplate(id: string | null): void;
  addCustomTemplate(template: RequestTemplate): void;
  removeCustomTemplate(id: string): void;
  getAllTemplates(): RequestTemplate[];
}

export const useTemplateStore = create<TemplateState>((set, get) => ({
  selectedCategory: null,
  selectedTemplateId: null,
  customTemplates: [],

  selectCategory(id) {
    set({ selectedCategory: id, selectedTemplateId: null });
  },

  selectTemplate(id) {
    set({ selectedTemplateId: id });
  },

  addCustomTemplate(template) {
    set((s) => ({ customTemplates: [...s.customTemplates, template] }));
  },

  removeCustomTemplate(id) {
    set((s) => ({
      customTemplates: s.customTemplates.filter((t) => t.id !== id),
    }));
  },

  getAllTemplates() {
    const builtIn = CATEGORIES.flatMap((c) => c.templates);
    return [...builtIn, ...get().customTemplates];
  },
}));
