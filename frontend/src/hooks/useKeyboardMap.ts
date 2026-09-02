import { useEffect } from "react";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useCommandPaletteStore } from "#stores/useCommandPaletteStore";
import { tabIsDirty, useRequestStore } from "#stores/useRequestStore";
import { useBottomPanelStore } from "#stores/useBottomPanelStore";
import { notifyWarning } from "#lib/notify";

const RAIL_VIEWS = [
  "home",
  "requests",
  "environments",
  "history",
  "mocks",
  "diff",
  "jwt",
  "graphql",
  "runners",
  "explorer",
  "docs",
  "grpc",
  "websocket",
  "sse",
  "mqtt",
  "socketio",
  "policy",
  "audit",
  "sso",
  "collab",
] as const;

function isTypingTarget(target: HTMLElement | null): boolean {
  if (!target) return false;
  if (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable) return true;
  // CodeMirror editor surface uses contenteditable internally
  return Boolean(target.closest("[contenteditable='true']"));
}

export function useKeyboardMap(onToggleSidebar?: () => void) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target instanceof HTMLElement ? e.target : null;
      const mod = e.metaKey || e.ctrlKey;
      if (!mod) return;
      const key = e.key.toLowerCase();
      if (key === "k") {
        e.preventDefault();
        const cur = useCommandPaletteStore.getState().open;
        useCommandPaletteStore.getState().setOpen(!cur);
        return;
      }
      if (key === "b") {
        if (isTypingTarget(target)) return;
        e.preventDefault();
        onToggleSidebar?.();
        return;
      }
      if (key === "w") {
        if (isTypingTarget(target)) return;
        e.preventDefault();
        const { activeTabId, closeTab } = useWorkspaceStore.getState();
        if (!activeTabId) return;
        const req = useRequestStore.getState();
        if (tabIsDirty(req.drafts[activeTabId], req.meta[activeTabId])) {
          notifyWarning("Tab has unsaved changes", "Close it with the tab's ✕ to discard them.");
          return;
        }
        closeTab(activeTabId, { force: true });
        return;
      }
      if (e.key === "Enter") {
        if (isTypingTarget(target)) return;
        e.preventDefault();
        const ws = useWorkspaceStore.getState();
        const rs = useRequestStore.getState();
        const active = ws.activeTabId;
        // Only send when a request tab is active
        if (!active) return;
        const tab = ws.openTabs.find((t) => t.id === active);
        if (tab && tab.kind && tab.kind !== "request" && tab.kind !== undefined) {
          // Non-request tabs: no-op per spec
          return;
        }
        // SAFETY: RequestEditor send seam expects tab id and draft input
        const draft = rs.drafts[active];
        if (!draft) return;
        void rs.send(active, {
          method: draft.method,
          url: draft.url,
          params: draft.params,
          headers: draft.headers,
          bodyType: draft.bodyType,
          body: draft.body,
          form: draft.form,
          graphqlQuery: draft.graphqlQuery,
          graphqlVariables: draft.graphqlVariables,
          auth: draft.auth,
          proxy: draft.proxy,
          tls: draft.tls,
        });
        return;
      }
      if (key === "j") {
        if (isTypingTarget(target)) return;
        e.preventDefault();
        const { active, collapsed } = useBottomPanelStore.getState();
        if (collapsed || !active) useBottomPanelStore.getState().setActive("console");
        else useBottomPanelStore.getState().setCollapsed(true);
        return;
      }
      if (/^[1-9]$/.test(e.key)) {
        if (isTypingTarget(target)) return;
        e.preventDefault();
        const idx = Number(e.key) - 1;
        const view = RAIL_VIEWS[idx];
        if (view) useWorkspaceStore.getState().requestView(view);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onToggleSidebar]);
}
