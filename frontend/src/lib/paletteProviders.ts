import { useWorkspaceStore, type WorkspaceView } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { useCommandPaletteStore } from "#stores/useCommandPaletteStore";
import { useThemeStore } from "#stores/useThemeStore";
import { useImportStore } from "#stores/useImportStore";
import { useExportStore } from "#stores/useExportStore";
import type { WorkspaceFolder } from "./collections";

export function registerDefaultPaletteProviders() {
  const store = useCommandPaletteStore.getState();
  // Commands - navigation, tab actions, import/export, theme
  const navViews: readonly WorkspaceView[] = [
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
    "settings",
    "spec-editor",
  ];
  navViews.forEach((v) => {
    store.registerCommand({
      id: `nav-${v}`,
      title: `Go to ${v}`,
      hint: "Navigation",
      keywords: v,
      run: () => useWorkspaceStore.getState().requestView(v),
    });
  });
  store.registerCommand({ id: "new-request", title: "New request", hint: "⌘N", keywords: "new request tab", run: () => useWorkspaceStore.getState().openTab({ id: `req-${Date.now()}`, title: "New Request" }) });
  store.registerCommand({ id: "close-tab", title: "Close tab", hint: "⌘W", keywords: "close", run: () => { const id = useWorkspaceStore.getState().activeTabId; if (id) useWorkspaceStore.getState().closeTab(id); } });
  store.registerCommand({ id: "toggle-theme", title: "Toggle theme", hint: "Theme", keywords: "theme dark light system", run: () => useThemeStore.getState().cycleTheme() });
  store.registerCommand({ id: "import", title: "Import collection", hint: "Import", keywords: "import openapi postman curl har", run: () => useImportStore.getState().setOpen(true) });
  store.registerCommand({ id: "export", title: "Export collection", hint: "Export", keywords: "export openapi postman curl har", run: () => useExportStore.getState().setOpen(true) });
  // Direct theme picks (story 18) — one palette hit per theme
  store.registerCommand({ id: "theme-light", title: "Theme: Light", hint: "Appearance", keywords: "theme light atlas-light", run: () => useThemeStore.getState().setTheme("atlas-light") });
  store.registerCommand({ id: "theme-dark", title: "Theme: Dark", hint: "Appearance", keywords: "theme dark atlas-dark", run: () => useThemeStore.getState().setTheme("atlas-dark") });
  store.registerCommand({ id: "theme-system", title: "Theme: System", hint: "Appearance", keywords: "theme system auto", run: () => useThemeStore.getState().setTheme("system") });

  // Data providers - capped at 50, graceful degrade
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
      id: `hist-${h.id}`, title: `${h.method} ${h.url}`, kind: "History", run: () => useWorkspaceStore.getState().requestView("history"),
    })),
  });
  store.registerProvider({
    id: "collections",
    kind: "Collection",
    getItems: () => {
      const tree = useWorkspaceStore.getState().workspaceTree;
      if (!tree) return [];
      const items: { id: string; title: string; kind: string; run: () => void }[] = [];
      const walkFolder = (folder: WorkspaceFolder, prefix: string) => {
        for (const req of folder.requests) {
          items.push({
            id: `col-${req.path}`,
            title: `${prefix} / ${req.name}`,
            kind: "Collection",
            run: () => void useWorkspaceStore.getState().openRequest(req.path),
          });
        }
        for (const sub of folder.folders) {
          walkFolder(sub, `${prefix} / ${sub.name}`);
        }
      };
      for (const col of tree.collections) {
        for (const req of col.requests) {
          items.push({
            id: `col-${req.path}`,
            title: `${col.name} / ${req.name}`,
            kind: "Collection",
            run: () => void useWorkspaceStore.getState().openRequest(req.path),
          });
        }
        for (const f of col.folders) {
          walkFolder(f, `${col.name} / ${f.name}`);
        }
      }
      return items.slice(0, 50);
    },
  });
}
