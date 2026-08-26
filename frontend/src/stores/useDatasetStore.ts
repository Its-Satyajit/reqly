import { create } from "zustand";
import {
  parseCsv,
  parseJsonDataset,
  getRowVariables,
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
