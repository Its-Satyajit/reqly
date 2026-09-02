import { create } from "zustand";
import {
  parseCsv,
  parseJsonDataset,
  getRowVariables,
  validateDataset,
  type Dataset,
} from "#lib/datasets";

export interface DatasetRunConfig {
  iterations: number;
  concurrency: number;
  bindTo: "request" | "collection" | "runtime";
}

interface DatasetState {
  dataset: Dataset | null;
  config: DatasetRunConfig;
  currentIteration: number;
  isRunning: boolean;
  results: { iteration: number; passed: boolean; error?: string }[];
  loadCsv(content: string, name?: string): void;
  loadJson(content: string, name?: string): void;
  loadDataset(content: string, name?: string): void;
  getValidationErrors(): string[];
  setConfig(patch: Partial<DatasetRunConfig>): void;
  getRowVariables(rowIndex: number): { [key: string]: string };
  clearDataset(): void;
}

export const useDatasetStore = create<DatasetState>((set, get) => ({
  dataset: null,
  config: { iterations: 1, concurrency: 1, bindTo: "request" },
  currentIteration: 0,
  isRunning: false,
  results: [],

  loadCsv(content, name) {
    set({ dataset: parseCsv(content, name), currentIteration: 0, results: [] });
  },

  loadJson(content, name) {
    set({ dataset: parseJsonDataset(content, name), currentIteration: 0, results: [] });
  },

  loadDataset(content, name) {
    const trimmed = content.trim();
    // Content sniffing wins over filename — JSON arrays are unambiguous.
    if (trimmed.startsWith("[")) {
      try {
        const ds = parseJsonDataset(content, name);
        set({ dataset: ds, currentIteration: 0, results: [] });
        return;
      } catch {
        // fall through to CSV
      }
    }
    // Fallback: treat as CSV (covers CSV content with .json name, empty, etc.)
    set({ dataset: parseCsv(content, name), currentIteration: 0, results: [] });
  },

  getValidationErrors() {
    const ds = get().dataset;
    if (!ds) return [];
    return validateDataset(ds);
  },

  setConfig(patch) {
    set((s) => ({ config: { ...s.config, ...patch } }));
  },

  getRowVariables(rowIndex) {
    const ds = get().dataset;
    if (!ds) return {};
    return getRowVariables(ds, rowIndex);
  },

  clearDataset() {
    set({ dataset: null, currentIteration: 0, results: [] });
  },
}));

// selectBulkData is the single selector for RunnersPanel bulk mode —
// it prefers the dataset store's rawContent (with validation/preview) over
// the legacy textarea fallback, hiding the store walk behind one call.
export function selectBulkData(mode: string, fallbackData: string): string | undefined {
  if (mode !== "bulk") return undefined;
  const ds = useDatasetStore.getState().dataset;
  if (ds?.rawContent && ds.rawContent.trim() !== "") return ds.rawContent;
  return fallbackData.trim() !== "" ? fallbackData : undefined;
}
