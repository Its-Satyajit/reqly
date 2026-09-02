// History adapter contract for desktop bridge.

import type { HeaderMap } from "./response"

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

export interface HistoryDetail extends HistoryEntry {
  reqHeaders: HeaderMap
  reqBody: string
  respHeaders: HeaderMap
  respBody: string
}

/** The response captured from replaying a stored request verbatim. */
export interface ReplayedResponse {
  statusCode: number
  durationMs: number
  size: number
  headers: HeaderMap
  body: string
}

export interface HistoryAdapter {
  list(limit: number, offset: number, status: string, env: string): Promise<HistoryEntry[]>
  show(id: string): Promise<HistoryDetail>
  search(query: string, limit: number): Promise<HistoryEntry[]>
  clear(env: string | null): Promise<void>
  replay(id: string): Promise<ReplayedResponse | null>
  replayWithVars?(id: string, vars: Record<string, string>): Promise<ReplayedResponse | null>
  listCookies(env: string): Promise<{ name: string; value: string; domain: string; path: string; env: string }[]>
  deleteCookie(name: string, domain: string, path: string, env: string): Promise<void>
  clearCookies(env: string | null): Promise<void>
}

export const fallbackHistoryAdapter: HistoryAdapter = {
  async list() { return [] },
  async show() { throw new Error("no history") },
  async search() { return [] },
  async clear() {},
  async replay() { return null },
  async listCookies() { return [] },
  async deleteCookie() {},
  async clearCookies() {},
}
