import { create } from "zustand";
import {
  fallbackExportAdapter,
  type ExportAdapter,
  type ExportFormat,
  type ExportOutcome,
} from "../lib/export";

interface ExportState {
  adapter: ExportAdapter
  open: boolean
  format: ExportFormat
  collection: string
  outName: string
  outcome: ExportOutcome | null
  busy: boolean
  error: string | null
  setAdapter(adapter: ExportAdapter): void
  setOpen(open: boolean): void
  setFormat(format: ExportFormat): void
  setCollection(collection: string): void
  setOutName(outName: string): void
  run(): Promise<ExportOutcome | null>
}

export const useExportStore = create<ExportState>((set, get) => ({
  adapter: fallbackExportAdapter,
  open: false,
  format: "postman",
  collection: "",
  outName: "",
  outcome: null,
  busy: false,
  error: null,

  setAdapter(adapter) {
    set({ adapter });
  },

  setOpen(open) {
    if (open) {
      set({ open: true, error: null, outcome: null, busy: false });
      return;
    }
    set({
      open: false,
      format: "postman",
      collection: "",
      outName: "",
      outcome: null,
      busy: false,
      error: null,
    });
  },

  setFormat(format) {
    set({ format, error: null });
  },

  setCollection(collection) {
    set({ collection, error: null });
  },

  setOutName(outName) {
    set({ outName, error: null });
  },

  async run() {
    const { adapter, format, collection, outName } = get();
    set({ busy: true, error: null });
    try {
      const outcome = await adapter.run({
        format,
        collection: collection.trim() === "" ? undefined : collection.trim(),
        outName: outName.trim() === "" ? undefined : outName.trim(),
      });
      set({ outcome, busy: false });
      return outcome;
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : String(err),
        busy: false,
      });
      return null;
    }
  },
}));
