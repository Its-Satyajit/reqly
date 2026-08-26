import { useEffect } from "react";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useCommandPaletteStore } from "#stores/useCommandPaletteStore";
import { useRequestStore } from "#stores/useRequestStore";

const RAIL_VIEWS = ["home","requests","environments","history","mocks","diff","jwt","graphql"] as const;

export function useKeyboardMap() {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      const isTyping = target && (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable);
      const mod = e.metaKey || e.ctrlKey;
      if (mod && e.key.toLowerCase() === "k") {
        e.preventDefault();
        useCommandPaletteStore.getState().setOpen(!useCommandPaletteStore.getState().open);
        return;
      }
      if (mod && e.key === "Enter") {
        if (isTyping) return;
        e.preventDefault();
        const ws = useWorkspaceStore.getState();
        const rs = useRequestStore.getState();
        const active = ws.activeTabId;
        if (active) void rs.send(active, {} as unknown as never);
        return;
      }
      if (mod && /^[1-8]$/.test(e.key)) {
        if (isTyping) return;
        e.preventDefault();
        const idx = Number(e.key) - 1;
        const view = RAIL_VIEWS[idx];
        if (view) useWorkspaceStore.getState().requestView(view as never);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);
}
