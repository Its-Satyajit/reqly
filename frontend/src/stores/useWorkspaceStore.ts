import { create } from 'zustand'

export interface Workspace {
  id: string
  name: string
  path: string
}

export interface Environment {
  id: string
  name: string
  variables: Record<string, string>
}

export interface RequestTab {
  id: string
  title: string
  requestId?: string
}

interface WorkspaceState {
  currentWorkspace: Workspace | null
  selectedCollectionId: string | null
  openTabs: RequestTab[]
  activeTabId: string | null
  activeEnvironmentId: string | null
  environments: Environment[]

  setCurrentWorkspace: (workspace: Workspace | null) => void
  selectCollection: (id: string | null) => void
  openTab: (tab: RequestTab) => void
  closeTab: (id: string) => void
  setActiveTab: (id: string | null) => void
  setActiveEnvironment: (id: string | null) => void
  setEnvironments: (environments: Environment[]) => void
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  currentWorkspace: null,
  selectedCollectionId: null,
  openTabs: [],
  activeTabId: null,
  activeEnvironmentId: null,
  environments: [],

  setCurrentWorkspace: (currentWorkspace) => set({ currentWorkspace }),
  selectCollection: (selectedCollectionId) => set({ selectedCollectionId }),

  openTab: (tab) =>
    set((state) => {
      const exists = state.openTabs.some((t) => t.id === tab.id)
      return {
        openTabs: exists ? state.openTabs : [...state.openTabs, tab],
        activeTabId: tab.id,
      }
    }),

  closeTab: (id) =>
    set((state) => {
      const index = state.openTabs.findIndex((t) => t.id === id)
      const openTabs = state.openTabs.filter((t) => t.id !== id)
      const activeTabId =
        state.activeTabId === id
          ? (openTabs[Math.max(0, index - 1)]?.id ?? null)
          : state.activeTabId
      return { openTabs, activeTabId }
    }),

  setActiveTab: (activeTabId) => set({ activeTabId }),
  setActiveEnvironment: (activeEnvironmentId) => set({ activeEnvironmentId }),
  setEnvironments: (environments) => set({ environments }),
}))