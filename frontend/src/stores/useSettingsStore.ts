import { create } from "zustand";

export type SettingsTabId =
  | "appearance"
  | "workspace"
  | "storage"
  | "network"
  | "security"
  | "cicd"
  | "shortcuts"
  | "auth"
  | "about";

interface SettingsState {
  activeTab: SettingsTabId;
  setActiveTab: (tab: SettingsTabId) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  activeTab: "appearance",
  setActiveTab: (activeTab) => set({ activeTab }),
}));

