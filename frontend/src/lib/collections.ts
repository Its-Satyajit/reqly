// Collections types + adapter shared by every Reqly front-end.
//
// Like environments and auth, collections never import the Wails generated
// bindings directly (they live in the host app and are regenerated on build).
// The host injects a CollectionsAdapter backed by the Go core's
// WorkspaceService; browser dev mode uses a read-only fallback.

import type { RequestAuth, ResponseData } from './request'

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

/** RunTestResult is the outcome of one reqly.test() assertion. */
export interface RunTestResult {
  name: string
  passed: boolean
}

/** RunStep is the streamed outcome of one request in a collection run. */
export interface RunStep {
  name: string
  /** Workspace-relative Request Path of the step. */
  requestPath: string
  passed: boolean
  /** Transport/pre-script error text ("" when none). */
  requestError: string
  /** Received response (null on failure). */
  response: ResponseData | null
  tests: RunTestResult[]
  logs: string[]
}

/** RunReport is the aggregate result of a finished collection run. */
export interface RunReport {
  steps: RunStep[]
  started: string
  finished: string
  total: number
  passed: number
  failed: number
  durationMs: number
  ok: boolean
}

/** RunEvent is a streamed collection-run event delivered to the UI. */
export type RunEvent =
  | { type: 'step'; step: RunStep }
  | { type: 'done'; report: RunReport }
  | { type: 'error'; message: string }

/** CollectionsAdapter is the seam through which the desktop UI loads the
 * workspace's collection tree, opens requests, and runs collections. */
export interface CollectionsAdapter {
  load: () => Promise<WorkspaceTree>
  /** open resolves a request file by Request Path into its fully resolved
   * form (effective URL, merged headers, inherited auth, variable chain,
   * file environment), ready to seed an editor tab. */
  open: (path: string) => Promise<OpenedRequest>
  /** run starts a collection/folder run. env names the environment pill ("" or
   * null for the workspace default); onEvent receives streamed step/done/error
   * events for the run's lifecycle. Resolves with the run id. */
  run: (
    path: string,
    env: string | null,
    failFast: boolean,
    onEvent: (event: RunEvent) => void,
  ) => Promise<string>
  /** cancelRun aborts an in-flight collection run by id. */
  cancelRun: (id: string) => Promise<void>
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
 * query/body. Inherited auth is carried along so it is applied silently at
 * send time. */
export interface ResolvedRequestInput {
  method: string
  url: string
  headers: { key: string; value: string }[]
  query: { key: string; value: string }[]
  body: string
  auth?: RequestAuth
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
  run: async () => {
    throw new Error('Running collections requires the desktop app.')
  },
  cancelRun: async () => {
    throw new Error('Running collections requires the desktop app.')
  },
}
