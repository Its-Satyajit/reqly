import { create } from 'zustand'
import type { RequestInput, RequestSender, ResponseData, KeyValueRow, RequestAuth } from '../lib/request'
import { fetchSender } from '../lib/request'
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
