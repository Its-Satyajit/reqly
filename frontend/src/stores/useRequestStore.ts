import { create } from 'zustand'
import type { RequestInput, RequestSender, ResponseData } from '../lib/request'
import { fetchSender } from '../lib/request'

interface RequestState {
  response: ResponseData | null
  loading: boolean
  error: string | null
  sender: RequestSender

  setSender: (sender: RequestSender) => void
  send: (req: RequestInput) => Promise<void>
  clear: () => void
}

/**
 * Request execution state. `sender` is the pluggable transport: the Wails host
 * injects the Go-core bridge; browser dev mode defaults to fetchSender.
 */
export const useRequestStore = create<RequestState>((set, get) => ({
  response: null,
  loading: false,
  error: null,
  sender: fetchSender,

  setSender: (sender) => set({ sender }),

  send: async (req) => {
    set({ loading: true, error: null })
    try {
      const response = await get().sender(req)
      set({ response, loading: false })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : String(err),
      })
    }
  },

  clear: () => set({ response: null, error: null, loading: false }),
}))
