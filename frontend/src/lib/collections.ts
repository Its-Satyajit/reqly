// Collections types + adapter shared by every Reqly front-end.
//
// Like environments and auth, collections never import the Wails generated
// bindings directly (they live in the host app and are regenerated on build).
// The host injects a CollectionsAdapter backed by the Go core's
// WorkspaceService; browser dev mode uses a read-only fallback.

/** WorkspaceRequest is a request file within a collection or folder, located
 * by its workspace-relative Request Path (e.g. "users/auth/login"). */
export interface WorkspaceRequest {
  name: string
  path: string
}

/** WorkspaceFolder is a nested container (recursively) within a collection. */
export interface WorkspaceFolder {
  name: string
  path: string
  folders: WorkspaceFolder[]
  requests: WorkspaceRequest[]
}

/** WorkspaceCollection is a top-level collection of folders and requests. */
export interface WorkspaceCollection {
  name: string
  path: string
  folders: WorkspaceFolder[]
  requests: WorkspaceRequest[]
}

/** WorkspaceTree is the bridge-friendly view of a workspace's collection
 * hierarchy: collections → folders → requests, all name-sorted. */
export interface WorkspaceTree {
  name: string
  path: string
  collections: WorkspaceCollection[]
}

/** CollectionsAdapter is the seam through which the desktop UI loads the
 * workspace's collection tree. */
export interface CollectionsAdapter {
  load: () => Promise<WorkspaceTree>
}

/**
 * fallbackCollectionsAdapter is used in plain Vite dev mode (no Wails
 * bridge): it reports an empty tree so the sidebar renders its empty state.
 */
export const fallbackCollectionsAdapter: CollectionsAdapter = {
  load: async () => ({ name: '', path: '', collections: [] }),
}