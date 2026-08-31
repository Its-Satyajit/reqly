import { create } from "zustand";
import {
  CREATE_HINT,
  fallbackWorkspaceBootstrapAdapter,
  type WorkspaceBootstrapAdapter,
  type WorkspaceStatus,
} from "../lib/workspace";
import { useWorkspaceStore } from "./useWorkspaceStore";
import { NEW_REQUEST_TAB_ID } from "./useRequestStore";

interface PendingCreate {
  dir: string;
  suggestedName: string;
}

interface RecentWorkspace {
  name: string;
  path: string;
  lastOpened: number;
}

interface WorkspaceBootstrapState {
  adapter: WorkspaceBootstrapAdapter;
  status: WorkspaceStatus | null;
  checked: boolean;
  pendingCreate: PendingCreate | null;
  createModalOpen: boolean;
  busy: boolean;
  error: string | null;
  recentWorkspaces: RecentWorkspace[];
  setAdapter(adapter: WorkspaceBootstrapAdapter): void;
  setCreateModalOpen(open: boolean): void;
  init(): Promise<void>;
  openFolder(): Promise<void>;
  openDirect(dir: string): Promise<void>;
  createInFolder(dir: string, name?: string): Promise<void>;
  createPending(name?: string): Promise<void>;
  cancelPendingCreate(): void;
  clearRecentWorkspaces(): void;
}

const RECENT_WORKSPACES_KEY = "reqly:recentWorkspaces";

function getStoredRecentWorkspaces(): RecentWorkspace[] {
  try {
    const raw = localStorage.getItem(RECENT_WORKSPACES_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) return parsed;
  } catch {
    // ignore
  }
  return [];
}

function saveRecentWorkspace(path: string, name?: string) {
  try {
    const list = getStoredRecentWorkspaces().filter((item) => item.path !== path);
    const inferredName = name || path.split(/[\\/]/).filter(Boolean).pop() || "Workspace";
    const updated: RecentWorkspace[] = [
      { name: inferredName, path, lastOpened: Date.now() },
      ...list,
    ].slice(0, 8);
    localStorage.setItem(RECENT_WORKSPACES_KEY, JSON.stringify(updated));
    return updated;
  } catch {
    return [];
  }
}

function resetTabsForWorkspaceSwitch() {
  const workspace = () => useWorkspaceStore.getState();
  for (const tab of workspace().openTabs) {
    if (tab.id !== NEW_REQUEST_TAB_ID) workspace().closeTab(tab.id);
  }
  if (!workspace().openTabs.some((t) => t.id === NEW_REQUEST_TAB_ID)) {
    workspace().openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
  }
  workspace().setActiveView("requests");
}

export const useWorkspaceBootstrapStore = create<WorkspaceBootstrapState>(
  (set, get) => ({
    adapter: fallbackWorkspaceBootstrapAdapter,
    status: null,
    checked: false,
    pendingCreate: null,
    createModalOpen: false,
    busy: false,
    error: null,
    recentWorkspaces: typeof window !== "undefined" ? getStoredRecentWorkspaces() : [],

    setAdapter(adapter) {
      set({ adapter });
    },

    setCreateModalOpen(createModalOpen) {
      set({ createModalOpen });
    },

    clearRecentWorkspaces() {
      try {
        localStorage.removeItem(RECENT_WORKSPACES_KEY);
      } catch {
        // ignore
      }
      set({ recentWorkspaces: [] });
    },

    async init() {
      const { adapter } = get();
      set({ busy: true, error: null });
      try {
        let status = await adapter.status();
        if (!status.found) {
          status = await adapter.restoreLast();
        }
        set({ status, checked: true, busy: false });
        if (status.found) {
          await Promise.all([
            useWorkspaceStore.getState().refreshWorkspace(),
            useWorkspaceStore.getState().refreshEnvironments(),
          ]);
        }
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
          checked: true,
        });
      }
    },

    async openFolder() {
      const { adapter } = get();
      set({ busy: true, error: null });
      try {
        const dir = await adapter.pickFolder();
        if (dir === "") {
          set({ busy: false });
          return;
        }
        try {
          await adapter.open(dir);
        } catch (err) {
          const message = err instanceof Error ? err.message : String(err);
          if (message.includes(CREATE_HINT)) {
            set({
              busy: false,
              pendingCreate: {
                dir,
                suggestedName: dir.split(/[\\/]/).filter(Boolean).pop() ?? "",
              },
            });
            return;
          }
          throw err;
        }
        await finishSwitch(set);
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
      }
    },

    async openDirect(dir: string) {
      const { adapter } = get();
      if (!dir) return;
      set({ busy: true, error: null });
      try {
        try {
          await adapter.open(dir);
        } catch (err) {
          const message = err instanceof Error ? err.message : String(err);
          if (message.includes(CREATE_HINT)) {
            set({
              busy: false,
              pendingCreate: {
                dir,
                suggestedName: dir.split(/[\\/]/).filter(Boolean).pop() ?? "",
              },
            });
            return;
          }
          throw err;
        }
        await finishSwitch(set);
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
      }
    },

    async createInFolder(dir: string, name?: string) {
      const { adapter } = get();
      if (!dir) return;
      set({ busy: true, error: null });
      try {
        await adapter.create(dir, name);
        set({ createModalOpen: false, pendingCreate: null });
        await finishSwitch(set);
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
      }
    },

    async createPending(name) {
      const { adapter, pendingCreate } = get();
      if (!pendingCreate) return;
      set({ busy: true, error: null });
      try {
        await adapter.create(pendingCreate.dir, name);
        set({ pendingCreate: null }, false);
        await finishSwitch(set);
      } catch (err) {
        set({
          error: err instanceof Error ? err.message : String(err),
          busy: false,
        });
      }
    },

    cancelPendingCreate() {
      set({ pendingCreate: null, error: null });
    },
  }),
);

async function finishSwitch(
  set: (partial: Partial<WorkspaceBootstrapState>) => void,
) {
  const status = await useWorkspaceBootstrapStore
    .getState()
    .adapter.status();
  if (status.found && status.path) {
    const treeName = useWorkspaceStore.getState().workspaceTree?.name;
    const recents = saveRecentWorkspace(status.path, treeName);
    set({ recentWorkspaces: recents });
  }
  resetTabsForWorkspaceSwitch();
  await Promise.all([
    useWorkspaceStore.getState().refreshWorkspace(),
    useWorkspaceStore.getState().refreshEnvironments(),
  ]);
  const treeName = useWorkspaceStore.getState().workspaceTree?.name;
  if (status.found && status.path) {
    const recents = saveRecentWorkspace(status.path, treeName);
    set({ recentWorkspaces: recents });
  }
  set({ status, checked: true, busy: false });
}

