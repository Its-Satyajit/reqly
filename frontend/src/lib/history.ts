// History adapter contract for desktop bridge.

export interface HistoryEntry {
  id: string
  requestPath: string
  method: string
  url: string
  env: string
  status: number
  durationMs: number
  size: number
  createdAt: string
}

export interface HistoryAdapter {
  list(limit: number, offset: number, status: string, env: string): Promise<HistoryEntry[]>
  show(id: string): Promise<HistoryEntry & { reqHeaders: Record<string, string[]>; reqBody: string; respHeaders: Record<string, string[]>; respBody: string }>
  search(query: string, limit: number): Promise<HistoryEntry[]>
  clear(env: string | null): Promise<void>
  replay(id: string): Promise<void>
  listCookies(env: string): Promise<{ name: string; value: string; domain: string; path: string; env: string }[]>
  deleteCookie(name: string, domain: string, path: string, env: string): Promise<void>
  clearCookies(env: string | null): Promise<void>
}

export const fallbackHistoryAdapter: HistoryAdapter = {
  async list() { return [] },
  async show() { throw new Error("no history") },
  async search() { return [] },
  async clear() {},
  async replay() {},
  async listCookies() { return [] },
  async deleteCookie() {},
  async clearCookies() {},
}
