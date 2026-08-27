import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useCommandPaletteStore } from "#stores/useCommandPaletteStore";
import { useThemeStore } from "#stores/useThemeStore";

export function registerDefaultPaletteProviders() {
  const store = useCommandPaletteStore.getState();
  // Commands - navigation, tab actions, import/export, theme
  const navViews = ["home","requests","environments","history","websocket","sse","settings","mocks","docs"] as const;
  navViews.forEach((v) => {
    store.registerCommand({
      id: `nav-${v}`,
      title: `Go to ${v}`,
      hint: "Navigation",
      keywords: v,
      run: () => useWorkspaceStore.getState().requestView(v as never),
    });
  });
  store.registerCommand({ id: "new-request", title: "New request", hint: "⌘N", keywords: "new request tab", run: () => useWorkspaceStore.getState().openTab({ id: `req-${Date.now()}`, title: "New Request" }) });
  store.registerCommand({ id: "close-tab", title: "Close tab", hint: "⌘W", keywords: "close", run: () => { const id = useWorkspaceStore.getState().activeTabId; if (id) useWorkspaceStore.getState().closeTab(id); } });
  store.registerCommand({ id: "toggle-theme", title: "Toggle theme", hint: "Theme", keywords: "theme dark light system", run: () => useThemeStore.getState().cycleTheme() });

  // Data providers - capped at 20, graceful degrade
  store.registerProvider({
    id: "environments",
    kind: "Environment",
    getItems: () => useWorkspaceStore.getState().environments.slice(0, 50).map((e) => ({
      id: `env-${e.id}`, title: e.name, kind: "Environment", run: () => useWorkspaceStore.getState().setActiveEnvironment(e.id),
    })),
  });
  store.registerProvider({
    id: "history",
    kind: "History",
    getItems: () => useHistoryStore.getState().pool.slice(0, 50).map((h) => ({
      id: `hist-${h.id}`, title: `${h.method} ${h.url}`, kind: "History", run: () => useWorkspaceStore.getState().requestView("history" as never),
    })),
  });
  store.registerProvider({
    id: "collections",
    kind: "Collection",
    getItems: () => {
      const tree = useWorkspaceStore.getState().workspaceTree;
      if (!tree) return [];
      const items: { id: string; title: string; kind: string; run: () => void }[] = [];
      const walk = (nodes: unknown[], prefix = "") => {
        for (const n of nodes as { name: string; path?: string; children?: unknown[] }[]) {
          const title = prefix ? `${prefix} / ${n.name}` : n.name;
          if (n.path) items.push({ id: `col-${n.path}`, title, kind: "Collection", run: () => void useWorkspaceStore.getState().openRequest(n.path!) });
          if (n.children) walk(n.children, title);
        }
      };
      walk((tree as unknown as { children?: unknown[] }).children ?? []);
      return items.slice(0, 50);
    },
  });
}
