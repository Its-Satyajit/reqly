import { create } from 'zustand'
import type { RequestInput, RequestSender, ResponseData, KeyValueRow, RequestAuth, RequestHeader, RequestRetry } from '../lib/request'
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
  /** GraphQL query and variables JSON for graphql body type. */
  graphqlQuery?: string
  graphqlVariables?: string
  /** The request's own auth (Inherit when unset — no own auth, the request
   * inherits from its containers). */
  auth?: RequestAuth
  /** The request's own retry policy (unset = no retries). */
  retry?: RequestRetry
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
  /** True when the last send was cancelled by the user (Stop); rendered
   * neutrally, distinct from an error. */
  cancelled?: boolean
}

export const emptyTabDraft = (): TabDraft => ({
  method: 'GET',
  url: '',
  bodyType: 'none',
  body: '{\n  \n}',
  form: [],
  params: [],
  headers: [],
  graphqlQuery: '',
  graphqlVariables: '{\n  \n}',
})

/** FileDraftInput is the file-owned request shape a save writes to disk. */
export interface FileDraftInput {
  method: string
  url: string
  headers: RequestHeader[]
  query: { key: string; value: string }[]
  body: string
  auth?: RequestAuth
  retry?: RequestRetry
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
  const { body, contentType } = serializeBody({
    bodyType: draft.bodyType,
    body: draft.body,
    form: draft.form,
    graphqlQuery: draft.graphqlQuery,
    graphqlVariables: draft.graphqlVariables,
  })
  if (contentType && !hasManualType) headers.push({ key: 'Content-Type', value: contentType })
  return {
    method: draft.method,
    url: draft.url,
    headers,
    query: sentRows(draft.params).map(({ key, value }) => ({ key, value })),
    body: body ?? '',
    auth: draft.auth,
    retry: draft.retry,
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
  /** Per-tab monotonically increasing send token; responses resolving under a
   * stale token are dropped so the newest Send always wins. */
  sendTokens: Record<string, number>
  /** Send id of the tab's in-flight send, when one is running. */
  activeSendIds: Record<string, string>
  /** Bridge-provided cancel seam (Go CancelSend); browser dev mode leaves it
   * unset — cancel still drops the token so late responses are discarded. */
  cancelSender: ((sendId: string) => Promise<void>) | null

  setSender: (sender: RequestSender) => void
  setCancelSender: (cancel: (sendId: string) => Promise<void>) => void
  /** Create a tab's draft if it does not exist yet, seeding from `seed`. */
  ensureDraft: (id: string, seed?: Partial<TabDraft>) => void
  updateDraft: (id: string, patch: Partial<TabDraft>) => void
  /** Set a tab's collection metadata (variables chain, environment pill). */
  setMeta: (id: string, meta: TabMeta) => void
  send: (id: string, req: RequestInput) => Promise<void>
  /** Abort the tab's in-flight send; unknown/finished sends are no-ops. */
  cancel: (id: string) => Promise<void>
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
  sendTokens: {},
  activeSendIds: {},
  cancelSender: null,

  setSender: (sender) => set({ sender }),

  setCancelSender: (cancelSender) => set({ cancelSender }),

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
    const token = (get().sendTokens[id] ?? 0) + 1
    const sendId = crypto.randomUUID()
    set((state) => ({
      sendTokens: { ...state.sendTokens, [id]: token },
      activeSendIds: { ...state.activeSendIds, [id]: sendId },
      responses: {
        ...state.responses,
        [id]: { ...state.responses[id], loading: true, error: null, cancelled: false },
      },
    }))
    const clearActive = () =>
      set((state) => {
        if (state.activeSendIds[id] !== sendId) return {}
        const activeSendIds = { ...state.activeSendIds }
        delete activeSendIds[id]
        return { activeSendIds }
      })
    try {
      const response = await get().sender({ ...req, sendId })
      clearActive()
      if (get().sendTokens[id] !== token) return
      set((state) => ({
        responses: {
          ...state.responses,
          [id]: { response, loading: false, error: null, cancelled: false },
        },
      }))
    } catch (err) {
      clearActive()
      if (get().sendTokens[id] !== token) return
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

  cancel: async (id) => {
    const sendId = get().activeSendIds[id]
    if (!sendId) return
    // Drop the pending token first so the aborted send's rejection cannot
    // overwrite the cancelled state, then tell the Go side to abort.
    set((state) => ({
      responses: {
        ...state.responses,
        [id]: { ...state.responses[id], loading: false, error: null, cancelled: true },
      },
    }))
    try {
      await get().cancelSender?.(sendId)
    } catch {
      // Cancel is best-effort; the send may have finished already.
    }
  },

  removeTab: (id) =>
    set((state) => {
      if (!state.drafts[id] && !state.responses[id] && !state.meta[id]) return {}
      const drafts = { ...state.drafts }
      const responses = { ...state.responses }
      const meta = { ...state.meta }
      const sendTokens = { ...state.sendTokens }
      const activeSendIds = { ...state.activeSendIds }
      delete drafts[id]
      delete responses[id]
      delete meta[id]
      delete sendTokens[id]
      delete activeSendIds[id]
      return { drafts, responses, meta, sendTokens, activeSendIds }
    }),
}))
