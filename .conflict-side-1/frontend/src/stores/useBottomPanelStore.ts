import { create } from "zustand";
import type { BottomPanelId } from "#lib/bottomPanel";

interface BottomPanelState {
  active: BottomPanelId | null;
  collapsed: boolean;
  setActive: (id: BottomPanelId | null) => void;
  toggle: (id: BottomPanelId) => void;
  setCollapsed: (collapsed: boolean) => void;
}

export const useBottomPanelStore = create<BottomPanelState>((set, get) => ({
  active: null,
  collapsed: true,
  setActive: (active) => set({ active, collapsed: active === null }),
  toggle: (id) => {
    const { active, collapsed } = get();
    if (active === id && !collapsed) {
      set({ collapsed: true });
      return;
    }
    set({ active: id, collapsed: false });
  },
  setCollapsed: (collapsed) => set({ collapsed }),
}));
