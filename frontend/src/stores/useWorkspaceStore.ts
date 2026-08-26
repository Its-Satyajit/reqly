import { create } from 'zustand'
import { fallbackEnvAdapter, type EnvAdapter } from '../lib/env'
import { isRecord } from '../lib/typeGuards'

type SaveError = Error | { code?: string; message?: string } | string
import {
  fallbackCollectionsAdapter,
  type CollectionsAdapter,
  type FileRequestInput,
  type ResolvedRequestInput,
  type WorkspaceTree,
} from '../lib/collections'
import type { RequestHeader } from '../lib/request'
import {
  useRequestStore,
  NEW_REQUEST_TAB_ID,
  tabIsDirty,
  fileInputFromDraft,
  emptyTabDraft,
  type TabDraft,
  type TabMeta,
} from './useRequestStore'
import { useCollectionRunStore } from './useCollectionRunStore'
import type { BodyType } from '../lib/body'

export interface Workspace {
  id: string
  name: string
  path: string
}

export interface Environment {
  id: string
  name: string
  description: string
  variables: Record<string, string>
  secrets: string[]
}

export interface RequestTab {
  id: string
  title: string
  /** Workspace-relative Request Path when opened from a collection. */
  requestPath?: string
  /** Tab kind: "request" (default) renders the request editor; "run" renders
   * the collection Run View; "test" renders the test runner. */
  kind?: 'request' | 'run' | 'test' | 'realtime'
  /** Workspace-relative file path for file-backed non-request tabs. */
  filePath?: string
}

export type WorkspaceView =
	| 'home'
	| 'requests'
	| 'environments'
	| 'history'
	| 'mocks'
	| 'diff'
	| 'jwt'
	| 'graphql'
	| 'runners'
	| 'explorer'
	| 'docs'
	| 'grpc'
	| 'websocket'
	| 'sse'
	| 'settings'

interface WorkspaceState {
  currentWorkspace: Workspace | null
  selectedCollectionId: string | null
  openTabs: RequestTab[]
  activeTabId: string | null
  activeView: WorkspaceView
  activeEnvironmentId: string | null
  environments: Environment[]
  environmentsError: string | null
  envAdapter: EnvAdapter
  workspaceTree: WorkspaceTree | null
  workspaceError: string | null
  workspaceAdapter: CollectionsAdapter
  expanded: Record<string, boolean>
  dirtyEditors: Record<string, boolean>
  hasUnsavedEnvChanges: boolean
  /** Transient error opening a specific request; never replaces the tree. */
  openError: string | null

  setCurrentWorkspace: (workspace: Workspace | null) => void
  selectCollection: (id: string | null) => void
  setActiveView: (view: WorkspaceView) => void
  /** Switch views, holding the switch behind a confirm when environment edits are unsaved. */
  requestView: (view: WorkspaceView) => void
  /** The view a switch is waiting on until unsaved env changes are resolved. */
  pendingView: WorkspaceView | null
  confirmPendingView: () => void
  cancelPendingView: () => void
  setEditorDirty: (key: string, dirty: boolean) => void
  openTab: (tab: RequestTab, seed?: Partial<import('./useRequestStore').TabDraft>) => void
  /** Close a tab. The caller decides the dirty-tab policy: pass force to
   * close without checking; otherwise a dirty file-backed tab is left open
   * and the UI is expected to confirm first. */
  closeTab: (id: string, opts?: { force?: boolean }) => void
  setActiveTab: (id: string | null) => void
  /** Open a collection request by Request Path into a tab, seeding the draft
   * from its raw file request and recording its version, base URL, inherited
   * headers, variable chain, and env pill. */
  openRequest: (path: string) => Promise<void>
  /** Save a file-backed tab's editable fields back to disk. On a
   * changed-on-disk conflict the tab flags the conflict instead of
   * clobbering the external edit. */
  saveRequest: (id: string) => Promise<void>
  /** Resolve a changed-on-disk conflict by overwriting: re-open the file and
   * save the current draft on top of it. */
  overwriteRequest: (id: string) => Promise<void>
  /** Resolve a changed-on-disk conflict by reloading: re-open the file and
   * discard the tab's edits, reseeding from the fresh file. */
  reloadRequest: (id: string) => Promise<void>
  setActiveEnvironment: (id: string | null) => void
  setEnvironments: (environments: Environment[]) => void
  setEnvAdapter: (adapter: EnvAdapter) => void
  setWorkspaceAdapter: (adapter: CollectionsAdapter) => void
  refreshWorkspace: () => Promise<void>
  toggleExpanded: (path: string) => void
  refreshEnvironments: () => Promise<void>
}

const toEnvironment = (name: string, src: { description?: string; variables?: Record<string, string>; secrets?: string[] }): Environment => ({
  id: name,
  name,
  description: src.description ?? '',
  variables: src.variables ?? {},
  secrets: src.secrets ?? [],
})

/** bodyTypeFor infers the editor's body type from an opened request's body and
 * Content-Type header, so what the tab shows matches what the core will send. */
export const bodyTypeFor = (req: {
  body?: string
  headers: { key: string; value: string }[]
}): BodyType => {
  if (!req.body) return 'none'
  const contentType = req.headers
    .find((h) => h.key.toLowerCase() === 'content-type')
    ?.value.toLowerCase()
  if (contentType?.includes('json')) return 'json'
  if (contentType?.includes('xml')) return 'xml'
  if (contentType?.includes('multipart/form-data')) return 'form-data'
  if (contentType?.includes('urlencoded')) return 'urlencoded'
  return 'raw'
}

/** draftFromFileRequest maps the raw file-owned request onto the editor draft
 * shape, preserving placeholders for send-time interpolation. Builder fields
 * plus the file's own auth are editable; inherited fields are recomputed at
 * send. */
export const draftFromFileRequest = (file: FileRequestInput): Partial<TabDraft> => {
  const toRows = (rows: { key: string; value: string }[]) =>
    rows.map(({ key, value }) => ({ key, value, enabled: true }))
  return {
    method: file.method,
    url: file.url,
    params: toRows(file.query),
    headers: toRows(file.headers),
    bodyType: bodyTypeFor(file),
    body: file.body,
    auth: file.auth,
    retry: file.retry,
  }
}

/** baseUrlFor derives the inherited base URL prefix from the opened request's
 * effective URL and the file's raw URL. An absolute file URL means no base;
 * otherwise the raw URL is stripped from the end of the effective URL. */
export const baseUrlFor = (fileUrl: string, effectiveUrl: string): string => {
  if (!fileUrl || fileUrl.includes('://')) return ''
  if (effectiveUrl.endsWith(fileUrl)) return effectiveUrl.slice(0, -fileUrl.length)
  if (effectiveUrl.endsWith('/' + fileUrl)) return effectiveUrl.slice(0, -fileUrl.length - 1)
  return ''
}

/** effectiveUrlFor renders the live Effective URL line: the draft URL joined
 * onto the inherited base, or the draft URL as-is when absolute/base-less. */
export const effectiveUrlFor = (draftUrl: string, baseUrl: string): string => {
  if (draftUrl.includes('://')) return draftUrl
  if (!baseUrl) return draftUrl
  return `${baseUrl}/${draftUrl.replace(/^\/+/, '')}`
}

/** inheritedHeadersFrom returns the resolved headers that are not declared by
 * the file itself (workspace → collection → folder), for the read-only group. */
export const inheritedHeadersFrom = (
  resolved: ResolvedRequestInput,
  file: FileRequestInput,
): RequestHeader[] => {
  const own = new Set((file.headers ?? []).map((h) => h.key.toLowerCase()))
  return (resolved.headers ?? []).filter((h) => !own.has(h.key.toLowerCase()))
}

/** isChangedOnDisk reports whether an error is a changed-on-disk save
 * conflict (the file changed under the editor). The bridge tags it with a
 * stable code; the message check is a fallback for non-bridge adapters. */
export function isChangedOnDisk(err: SaveError): boolean {
  if (isRecord(err) && 'code' in err) {
    // SAFETY: in operator narrows to object with code; string access validated via isRecord
    return (err as { code?: string }).code === 'ERR_CHANGED_ON_DISK'
  }
  return err instanceof Error && err.message.includes('changed on disk')
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  currentWorkspace: null,
  selectedCollectionId: null,
  openTabs: [],
  activeTabId: null,
  activeView: 'home',
  activeEnvironmentId: null,
  environments: [],
  environmentsError: null,
  envAdapter: fallbackEnvAdapter,
  workspaceTree: null,
  workspaceError: null,
  workspaceAdapter: fallbackCollectionsAdapter,
  expanded: {},
  dirtyEditors: {},
  hasUnsavedEnvChanges: false,
  openError: null,

  setCurrentWorkspace: (currentWorkspace) => set({ currentWorkspace }),
  selectCollection: (selectedCollectionId) => set({ selectedCollectionId }),
  setActiveView: (activeView) => set({ activeView }),
  requestView: (view) => {
    const { activeView, hasUnsavedEnvChanges } = get()
    if (activeView === view) return
    if (hasUnsavedEnvChanges) {
      set({ pendingView: view })
      return
    }
    set({ activeView: view })
  },
  pendingView: null,
  confirmPendingView: () => {
    const { pendingView } = get()
    if (pendingView) set({ activeView: pendingView })
    set({ pendingView: null })
  },
  cancelPendingView: () => set({ pendingView: null }),
  setEditorDirty: (key, dirty) =>
    set((state) => {
      const dirtyEditors = { ...state.dirtyEditors, [key]: dirty }
      return { dirtyEditors, hasUnsavedEnvChanges: Object.values(dirtyEditors).some(Boolean) }
    }),

  openTab: (tab, seed) => {
    const exists = get().openTabs.some((t) => t.id === tab.id)
    set((state) => ({
      openTabs: exists ? state.openTabs : [...state.openTabs, tab],
      activeTabId: tab.id,
    }))
    if (!exists && tab.kind !== 'run') {
      useRequestStore.getState().ensureDraft(tab.id, seed)
    }
  },

  closeTab: (id, opts) => {
    const { drafts, meta } = useRequestStore.getState()
    if (!opts?.force && tabIsDirty(drafts[id], meta[id])) return
    const closing = get().openTabs.find((t) => t.id === id)
    const index = get().openTabs.findIndex((t) => t.id === id)
    let openTabs = get().openTabs.filter((t) => t.id !== id)
    const restoringScratchpad =
      openTabs.length === 0
        ? [{ id: NEW_REQUEST_TAB_ID, title: 'New Request' }]
        : null
    if (restoringScratchpad) openTabs = restoringScratchpad
    set((state) => ({
      openTabs,
      activeTabId:
        state.activeTabId === id
          ? (openTabs[Math.max(0, index - 1)]?.id ?? null)
          : state.activeTabId,
    }))
    useRequestStore.getState().removeTab(id)
    if (restoringScratchpad) {
      useRequestStore.getState().ensureDraft(NEW_REQUEST_TAB_ID)
    }
    if (closing?.kind === 'run') {
      useCollectionRunStore.getState().reset()
    }
  },

  setActiveTab: (activeTabId) => {
    if (activeTabId) {
      const tab = get().openTabs.find((t) => t.id === activeTabId)
      if (!tab || tab.kind !== 'run') {
        useRequestStore.getState().ensureDraft(activeTabId)
      }
    }
    set({ activeTabId })
  },

  openRequest: async (path) => {
    const { workspaceAdapter } = get()
    try {
      const opened = await workspaceAdapter.open(path)
      set({ openError: null })
      const seed = draftFromFileRequest(opened.fileRequest)
      get().openTab(
        { id: opened.path, title: opened.name, requestPath: opened.path },
        seed,
      )
      const meta: TabMeta = {
        requestPath: opened.path,
        name: opened.name,
        variables: opened.variables,
        env: opened.fileEnv || undefined,
        auth: opened.request.auth,
        version: opened.version,
        baseUrl: baseUrlFor(opened.fileRequest.url, opened.request.url),
        inheritedHeaders: inheritedHeadersFrom(opened.request, opened.fileRequest),
        baseline: { ...emptyTabDraft(), ...seed },
      }
      useRequestStore.getState().setMeta(opened.path, meta)
    } catch (err) {
      set({ openError: err instanceof Error ? err.message : String(err) })
    }
  },

  saveRequest: async (id) => {
    const { workspaceAdapter } = get()
    const { drafts, meta } = useRequestStore.getState()
    const m = meta[id]
    const draft = drafts[id]
    if (!m?.requestPath || !draft || m.changedOnDisk) return
    try {
      const version = await workspaceAdapter.save(m.requestPath, fileInputFromDraft(draft), m.version ?? '')
      useRequestStore.getState().setMeta(id, {
        ...m,
        version,
        baseline: { ...draft },
        changedOnDisk: false,
      })
    } catch (err) {
      // SAFETY: caught error is SaveError shape (Error or {code} from adapter) at I/O boundary
      useRequestStore.getState().setMeta(id, {
        ...m,
        changedOnDisk: isChangedOnDisk(err as SaveError),
      })
      // SAFETY: same SaveError shape as above; validated via isRecord/code check
      if (!isChangedOnDisk(err as SaveError)) {
        set({ openError: err instanceof Error ? err.message : String(err) })
      }
    }
  },

  overwriteRequest: async (id) => {
    const { workspaceAdapter } = get()
    const { drafts, meta } = useRequestStore.getState()
    const m = meta[id]
    const draft = drafts[id]
    if (!m?.requestPath || !draft) return
    try {
      // Re-open the file for a fresh version, then save the current draft on
      // top of it — the editor's edits win over the external change.
      const opened = await workspaceAdapter.open(m.requestPath)
      const version = await workspaceAdapter.save(
        m.requestPath,
        fileInputFromDraft(draft),
        opened.version,
      )
      useRequestStore.getState().setMeta(id, {
        ...m,
        version,
        baseline: { ...draft },
        changedOnDisk: false,
      })
      set({ openError: null })
    } catch (err) {
      set({ openError: err instanceof Error ? err.message : String(err) })
    }
  },

  reloadRequest: async (id) => {
    const { workspaceAdapter } = get()
    const { meta } = useRequestStore.getState()
    const m = meta[id]
    if (!m?.requestPath) return
    try {
      // Discard the tab's edits and reseed from the fresh file on disk.
      const opened = await workspaceAdapter.open(m.requestPath)
      const seed = draftFromFileRequest(opened.fileRequest)
      useRequestStore.getState().updateDraft(id, { ...emptyTabDraft(), ...seed })
      useRequestStore.getState().setMeta(id, {
        ...m,
        version: opened.version,
        baseUrl: baseUrlFor(opened.fileRequest.url, opened.request.url),
        inheritedHeaders: inheritedHeadersFrom(opened.request, opened.fileRequest),
        baseline: { ...emptyTabDraft(), ...seed },
        changedOnDisk: false,
      })
      set({ openError: null })
    } catch (err) {
      set({ openError: err instanceof Error ? err.message : String(err) })
    }
  },

  setActiveEnvironment: (activeEnvironmentId) => set({ activeEnvironmentId }),

  setEnvironments: (environments) => set({ environments }),

  setEnvAdapter: (envAdapter) => set({ envAdapter }),

  setWorkspaceAdapter: (workspaceAdapter) => set({ workspaceAdapter }),

  refreshWorkspace: async () => {
    const { workspaceAdapter } = get()
    try {
      const tree = await workspaceAdapter.load()
      set({
        workspaceTree: tree,
        workspaceError: null,
        openError: null,
        currentWorkspace: tree.name
          ? { id: tree.path, name: tree.name, path: tree.path }
          : null,
      })
    } catch (err) {
      set({ workspaceError: err instanceof Error ? err.message : String(err) })
    }
  },

  toggleExpanded: (path) =>
    set((state) => ({ expanded: { ...state.expanded, [path]: !state.expanded[path] } })),

  refreshEnvironments: async () => {
    const { envAdapter } = get()
    try {
      const data = await envAdapter.list()
      set({
        environments: data.environments.map((e) => toEnvironment(e.name, e)),
        activeEnvironmentId: data.active || null,
        environmentsError: null,
      })
    } catch (err) {
      set({ environmentsError: err instanceof Error ? err.message : String(err) })
    }
  },
}))