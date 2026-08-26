import { create } from "zustand";
import {
  fallbackImportAdapter,
  type ImportAdapter,
  type ImportOutcome,
} from "../lib/import";

export type ImportStage = "input" | "preview" | "results";

interface ImportState {
  adapter: ImportAdapter
  open: boolean
  stage: ImportStage
  content: string
  filename: string
  formatHint: string
  detected: { format: string; ok: boolean } | null
  outcome: ImportOutcome | null
  targetDir: string
  busy: boolean
  error: string | null
  setAdapter(adapter: ImportAdapter): void
  setOpen(open: boolean): void
  setContent(content: string, filename?: string): void
  setFormatHint(hint: string): void
  setTargetDir(dir: string): void
  runPreview(): Promise<void>
  commit(): Promise<ImportOutcome | null>
  back(): void
}

export const useImportStore = create<ImportState>((set, get) => {
  let detectTimer: ReturnType<typeof setTimeout> | null = null;

  const scheduleDetect = () => {
    if (detectTimer) clearTimeout(detectTimer);
    detectTimer = setTimeout(() => {
      const { adapter, content } = get();
      if (content.trim() === "") {
        set({ detected: null });
        return;
      }
      void adapter
        .detect(content)
        .then((detected) => {
          set({ detected });
        })
        .catch(() => {
          set({ detected: null });
        });
    }, 250);
  };

  return {
    adapter: fallbackImportAdapter,
    open: false,
    stage: "input",
    content: "",
    filename: "",
    formatHint: "",
    detected: null,
    outcome: null,
    targetDir: "",
    busy: false,
    error: null,

    setAdapter(adapter) {
      set({ adapter });
    },

    setOpen(open) {
      if (open) {
        set({ open: true, stage: "input", error: null });
        return;
      }
      set({
        open: false,
        stage: "input",
        content: "",
        filename: "",
        formatHint: "",
        detected: null,
        outcome: null,
        targetDir: "",
        busy: false,
        error: null,
      });
    },

    setContent(content, filename = "") {
      set({ content, filename, error: null });
      scheduleDetect();
    },

    setFormatHint(formatHint) {
      set({ formatHint, error: null });
      scheduleDetect();
    },

    setTargetDir(targetDir) {
      set({ targetDir, error: null });
    },

    async runPreview() {
      const { adapter, content, formatHint } = get();
      if (content.trim() === "") return;
      set({ busy: true, error: null });
      try {
        const outcome = await adapter.preview({
          content,
          formatHint: formatHint === "" ? undefined : formatHint,
        });
        set({
          outcome,
          stage: "preview",
          targetDir: outcome.targetDir ?? "",
          busy: false,
        });
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
      }
    },

    async commit() {
      const { adapter, content, formatHint, targetDir } = get();
      set({ busy: true, error: null });
      try {
        const outcome = await adapter.commit({
          content,
          formatHint: formatHint === "" ? undefined : formatHint,
          targetDir: targetDir.trim() === "" ? undefined : targetDir.trim(),
        });
        set({ outcome, stage: "results", busy: false });
        return outcome;
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
        return null;
      }
    },

    back() {
      set({ stage: "input", error: null });
    },
  };
});
