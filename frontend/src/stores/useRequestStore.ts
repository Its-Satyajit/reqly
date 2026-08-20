import { create } from 'zustand'
import type { RequestInput, RequestSender, ResponseData, KeyValueRow, RequestAuth, RequestHeader } from '../lib/request'
import { fetchSender, sentRows } from '../lib/request'
import { serializeBody } from '../lib/body'
import type { BodyType } from '../lib/body'
import type { ResolvedVariable } from '../lib/collections'

/** The id of the persistent ad-hoc scratchpad tab (not tied to a request file). */
export const NEW_REQUEST_TAB_ID = 'new-request'

/** Per-tab editable request fields held in the editor. */
export interface TabDraft {
  method: string
  url: string
  bodyType: BodyType
  body: string
  form: KeyValueRow[]
  params: KeyValueRow[]
  headers: KeyValueRow[]
  /** The request's own auth (Inherit when unset — no own auth, the request
   * inherits from its containers). */
  auth?: RequestAuth
}

/** Per-tab metadata for requests opened from a collection: the source request
 * path, its effective variable chain (read-only), the environment pill, and
 * the inherited auth applied silently at send. The scratchpad tab has none
 * of these. */
export interface TabMeta {
  requestPath?: string
  name?: string
  variables: ResolvedVariable[]
  env?: string
  auth?: RequestAuth
  /** Fingerprint of the raw file bytes at open/last-save; a save is only
   * accepted while the on-disk bytes still match. */
  version?: string
  /** The inherited base URL prefix ("" when the file URL is absolute) used to
   * render the tab's live Effective URL line. */
  baseUrl?: string
  /** Headers inherited from the container chain (workspace → collection →
   * folder), displayed read-only alongside the editable own headers. */
  inheritedHeaders: RequestHeader[]
  /** The draft exactly as seeded from the file; dirty = draft ≠ baseline. */
  baseline?: TabDraft
  /** True when the last save attempt hit a concurrent on-disk edit; the
   * editor surfaces Overwrite / Reload. */
  changedOnDisk?: boolean
}

/** Per-tab execution state: the response (if any), in-flight flag, error. */
export interface TabResponse {
  response: ResponseData | null
  loading: boolean
  error: string | null
}

export const emptyTabDraft = (): TabDraft => ({
  method: 'GET',
  url: '',
  bodyType: 'none',
  body: '{\n  \n}',
  form: [],
  params: [],
  headers: [],
})

/** FileDraftInput is the file-owned request shape a save writes to disk. */
export interface FileDraftInput {
  method: string
  url: string
  headers: RequestHeader[]
  query: { key: string; value: string }[]
  body: string
  auth?: RequestAuth
}

/** fileInputFromDraft serializes a tab's editable fields into the
 * file-owned request shape a save writes to disk: the body type + body/form
 * collapse into request.body, with the implied Content-Type pushed onto the
 * headers (unless a manual Content-Type is already present) so the file
 * round-trips through the editor. The draft's own auth rides along (Inherit
 * writes none, so an existing file auth block is removed). */
export function fileInputFromDraft(draft: TabDraft): FileDraftInput {
  const headers = sentRows(draft.headers).map(({ key, value }) => ({ key, value }))
  const hasManualType = headers.some((h) => h.key.toLowerCase() === 'content-type')
  const { body, contentType } = serializeBody(draft)
  if (contentType && !hasManualType) headers.push({ key: 'Content-Type', value: contentType })
  return {
    method: draft.method,
    url: draft.url,
    headers,
    query: sentRows(draft.params).map(({ key, value }) => ({ key, value })),
    body: body ?? '',
    auth: draft.auth,
  }
}

/** tabIsDirty reports whether a tab's draft differs from its file baseline
 * (the seeded file request). Scratchpad tabs are never dirty. */
export function tabIsDirty(draft: TabDraft | undefined, meta: TabMeta | undefined): boolean {
  if (!draft || !meta?.requestPath || !meta.baseline) return false
  return JSON.stringify(draft) !== JSON.stringify(meta.baseline)
}

interface RequestState {
  drafts: Record<string, TabDraft>
  responses: Record<string, TabResponse>
  meta: Record<string, TabMeta>
  sender: RequestSender

  setSender: (sender: RequestSender) => void
  /** Create a tab's draft if it does not exist yet, seeding from `seed`. */
  ensureDraft: (id: string, seed?: Partial<TabDraft>) => void
  updateDraft: (id: string, patch: Partial<TabDraft>) => void
  /** Set a tab's collection metadata (variables chain, environment pill). */
  setMeta: (id: string, meta: TabMeta) => void
  send: (id: string, req: RequestInput) => Promise<void>
  removeTab: (id: string) => void
}

/**
 * Request execution + per-tab draft/response state. `sender` is the pluggable
 * transport: the Wails host injects the Go-core bridge; browser dev mode
 * defaults to fetchSender. Tabs are keyed by tab id; the open/active tab list
 * lives in the workspace store.
 */
export const useRequestStore = create<RequestState>((set, get) => ({
  drafts: {},
  responses: {},
  meta: {},
  sender: fetchSender,

  setSender: (sender) => set({ sender }),

  ensureDraft: (id, seed) =>
    set((state) => {
      if (state.drafts[id]) return {}
      return { drafts: { ...state.drafts, [id]: { ...emptyTabDraft(), ...seed } } }
    }),

  updateDraft: (id, patch) =>
    set((state) => {
      const draft = state.drafts[id]
      if (!draft) return {}
      return { drafts: { ...state.drafts, [id]: { ...draft, ...patch } } }
    }),

  setMeta: (id, meta) =>
    set((state) => ({ meta: { ...state.meta, [id]: meta } })),

  send: async (id, req) => {
    set((state) => ({
      responses: {
        ...state.responses,
        [id]: { ...state.responses[id], loading: true, error: null },
      },
    }))
    try {
      const response = await get().sender(req)
      set((state) => ({
        responses: {
          ...state.responses,
          [id]: { response, loading: false, error: null },
        },
      }))
    } catch (err) {
      set((state) => ({
        responses: {
          ...state.responses,
          [id]: {
            ...state.responses[id],
            loading: false,
            error: err instanceof Error ? err.message : String(err),
          },
        },
      }))
    }
  },

  removeTab: (id) =>
    set((state) => {
      if (!state.drafts[id] && !state.responses[id] && !state.meta[id]) return {}
      const drafts = { ...state.drafts }
      const responses = { ...state.responses }
      const meta = { ...state.meta }
      delete drafts[id]
      delete responses[id]
      delete meta[id]
      return { drafts, responses, meta }
    }),
}))
