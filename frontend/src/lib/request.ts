// Request types + senders shared by every Reqly front-end.
//
// The shared UI never imports the Wails-generated bindings directly (they live
// in the host app and are regenerated on build). Instead, request execution is
// injected through a RequestSender: the Wails host injects a sender backed by
// the Go core, while browser dev mode uses fetchSender.

export interface RequestHeader {
  key: string
  value: string
}

export interface RequestInput {
  method?: string
  url: string
  headers?: RequestHeader[]
  body?: string
  timeout?: number
}

export interface ResponseData {
  statusCode: number
  statusText: string
  proto: string
  headers: Record<string, string[]>
  body: string
  durationMs: number
  size: number
  ok: boolean
}

export type RequestSender = (req: RequestInput) => Promise<ResponseData>

const CONTENT_TYPE = 'Content-Type'

function detectContentType(body: string): string {
  const trimmed = body.trim()
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) return 'application/json'
  return 'text/plain'
}

/**
 * fetchSender runs a request in the browser (no Wails bridge). Used as the
 * default so the shared UI is usable in plain Vite dev mode.
 */
export const fetchSender: RequestSender = async (req) => {
  const started = performance.now()

  const headers: Record<string, string> = {}
  for (const h of req.headers ?? []) headers[h.key] = h.value
  if (req.body && !Object.keys(headers).some((k) => k.toLowerCase() === 'content-type')) {
    headers[CONTENT_TYPE] = detectContentType(req.body)
  }

  const res = await fetch(req.url, {
    method: req.method ?? 'GET',
    headers,
    body: req.body || undefined,
  })

  const body = await res.text()
  const responseHeaders: Record<string, string[]> = {}
  res.headers.forEach((value, key) => {
    ;(responseHeaders[key] ??= []).push(value)
  })

  return {
    statusCode: res.status,
    statusText: res.statusText,
    proto: '',
    headers: responseHeaders,
    body,
    durationMs: Math.round(performance.now() - started),
    size: new TextEncoder().encode(body).length,
    ok: res.ok,
  }
}
