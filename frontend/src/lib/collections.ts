// Collections types + adapter shared by every Reqly front-end.
//
// Like environments and auth, collections never import the Wails generated
// bindings directly (they live in the host app and are regenerated on build).
// The host injects a CollectionsAdapter backed by the Go core's
// WorkspaceService; browser dev mode uses a read-only fallback.

import type { RequestAuth, RequestRetry, ResponseData } from './request'

/** WorkspaceRequest is a request file within a collection or folder, located
 * by its workspace-relative Request Path (e.g. "users/auth/login"). */
export interface WorkspaceRequest {
  name: string
  path: string
  /** HTTP method from the request file; drives the sidebar's method chip. */
  method?: string
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
 * workspace's collection tree, opens requests, runs collections, and saves
 * their editable builder fields back to disk. */
export interface CollectionsAdapter {
  load: () => Promise<WorkspaceTree>
  /** open resolves a request file by Request Path into its fully resolved
   * form (effective URL, merged headers, inherited auth, variable chain,
   * file environment) plus the raw file-owned request and version, ready to
   * seed an editor tab. */
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
  /** exportReport serializes a finished run as "json" or "junit" (JUnit XML),
   * writes it under .reqly/exports/, and returns the path + rendered text. */
  exportReport?: (
    format: "json" | "junit",
    report: {
      path?: string
      started?: string
      finished?: string
      durationMs: number
      steps: {
        name: string
        requestPath?: string
        passed: boolean
        requestError?: string
        durationMs?: number
        tests?: { name: string; passed: boolean }[]
      }[]
    },
  ) => Promise<{ format: string; path: string; content: string }>
  /** save persists a file-backed tab's editable builder fields to disk,
   * preserving format and every non-editable field. expectedVersion must
   * match the file's fingerprint from OpenedRequest.version; a mismatch
   * rejects the save (changed on disk) without touching the file. Resolves
   * to the new baseline version on success. */
  save: (path: string, draft: FileRequestInput, expectedVersion: string) => Promise<string>
  /** createCollection scaffolds collections/<name>/ with a descriptor. */
  createCollection: (name: string) => Promise<void>
  /** createFolder scaffolds <parent>/<name>/ with a descriptor. parent is a
   * container Request Path ("payments" or "payments/auth"). */
  createFolder: (parent: string, name: string) => Promise<void>
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

/** FileRequestInput is the raw, unmerged file-owned request: the editor seed.
 * It carries only what the file declares (no inherited base URL, headers, or
 * auth), and its builder fields (url/method/headers/query/body) plus its own
 * auth are editable — everything else is preserved verbatim on save. */
export interface FileRequestInput {
  method: string
  url: string
  headers: { key: string; value: string }[]
  query: { key: string; value: string }[]
  body: string
  /** The file's own auth (""/unset when the request declares none and inherits
   * from its containers). */
  auth?: RequestAuth
  /** The file's own retry policy (unset when the request retries nothing). */
  retry?: RequestRetry
}

/** OpenedRequest is a request file combined with its inherited configuration
 * and full variable chain, ready to be loaded into an editor. Placeholders
 * are left intact — they resolve at send time with the environment layer. */
export interface OpenedRequest {
  path: string
  name: string
  request: ResolvedRequestInput
  fileRequest: FileRequestInput
  variables: ResolvedVariable[]
  /** The request file's environment: field ("" when unset); the sending tab
   * uses it as its environment pill. */
  fileEnv: string
  /** Fingerprint of the raw file bytes at open time; a save is only accepted
   * while the on-disk bytes still match. */
  version: string
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
  save: async () => {
    throw new Error('Saving request files requires the desktop app.')
  },
  createCollection: async () => {
    throw new Error('Creating collections requires the desktop app.')
  },
  createFolder: async () => {
    throw new Error('Creating folders requires the desktop app.')
  },
}