import { create } from "zustand";
import {
  CREATE_HINT,
  fallbackWorkspaceBootstrapAdapter,
  rememberWorkspace,
  type WorkspaceBootstrapAdapter,
  type WorkspaceStatus,
} from "../lib/workspace";
import { useWorkspaceStore } from "./useWorkspaceStore";
import { NEW_REQUEST_TAB_ID } from "./useRequestStore";

interface PendingCreate {
  dir: string;
  suggestedName: string;
}

interface WorkspaceBootstrapState {
  adapter: WorkspaceBootstrapAdapter
  status: WorkspaceStatus | null
  checked: boolean
  pendingCreate: PendingCreate | null
  busy: boolean
  error: string | null
  setAdapter(adapter: WorkspaceBootstrapAdapter): void
  init(): Promise<void>
  openFolder(): Promise<void>
  createPending(name?: string): Promise<void>
  cancelPendingCreate(): void
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
    busy: false,
    error: null,

    setAdapter(adapter) {
      set({ adapter });
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
  if (status.path) rememberWorkspace(status.path);
  resetTabsForWorkspaceSwitch();
  await Promise.all([
    useWorkspaceStore.getState().refreshWorkspace(),
    useWorkspaceStore.getState().refreshEnvironments(),
  ]);
  set({ status, checked: true, busy: false });
}
