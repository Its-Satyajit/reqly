import { create } from "zustand";
import { generateCliCommand, type CicdPipeline } from "#lib/cicd";

interface CicdState {
  pipeline: CicdPipeline;
  setPipeline(patch: Partial<CicdPipeline>): void;
  addSecret(name: string): void;
  removeSecret(name: string): void;
  getCommand(): string;
}

export const useCicdStore = create<CicdState>((set, get) => ({
  pipeline: {
    name: "CI Tests",
    environment: "production",
    secrets: [],
  },

  setPipeline(patch) {
    set((s) => ({ pipeline: { ...s.pipeline, ...patch } }));
  },

  addSecret(name) {
    set((s) => ({
      pipeline: {
        ...s.pipeline,
        secrets: s.pipeline.secrets.includes(name)
          ? s.pipeline.secrets
          : [...s.pipeline.secrets, name],
      },
    }));
  },

  removeSecret(name) {
    set((s) => ({
      pipeline: {
        ...s.pipeline,
        secrets: s.pipeline.secrets.filter((n) => n !== name),
      },
    }));
  },

  getCommand() {
    return generateCliCommand(get().pipeline);
  },
}));
