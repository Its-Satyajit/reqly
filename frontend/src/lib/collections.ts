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
 * workspace's collection tree and opens requests. */
export interface CollectionsAdapter {
  load: () => Promise<WorkspaceTree>
  /** open resolves a request file by Request Path into its fully resolved
   * form (effective URL, merged headers, inherited auth, variable chain,
   * file environment), ready to seed an editor tab. */
  open: (path: string) => Promise<OpenedRequest>
}

/** ResolvedVariable is one entry of an opened request's effective variable
 * chain, tagged with the scope that defined it. */
export interface ResolvedVariable {
  name: string
  value: string
  scope: string
}

/** ResolvedRequestInput is the resolved request fields an opened tab is
 * seeded with: the effective URL, merged headers (inherited + own), and
 * query/body. Inherited auth is applied silently by the core at send time. */
export interface ResolvedRequestInput {
  method: string
  url: string
  headers: { key: string; value: string }[]
  query: { key: string; value: string }[]
  body: string
}

/** OpenedRequest is a request file combined with its inherited configuration
 * and full variable chain, ready to be loaded into an editor. Placeholders
 * are left intact — they resolve at send time with the environment layer. */
export interface OpenedRequest {
  path: string
  name: string
  request: ResolvedRequestInput
  variables: ResolvedVariable[]
  /** The request file's environment: field ("" when unset); the sending tab
   * uses it as its environment pill. */
  fileEnv: string
}

/**
 * fallbackCollectionsAdapter is used in plain Vite dev mode (no Wails
 * bridge): it reports an empty tree so the sidebar renders its empty state,
 * and opening a request fails with a clear message.
 */
export const fallbackCollectionsAdapter: CollectionsAdapter = {
  load: async () => ({ name: '', path: '', collections: [] }),
  open: async () => {
    throw new Error('Opening collections requires the desktop app.')
  },
}