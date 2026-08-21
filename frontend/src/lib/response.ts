// Response-view helpers shared by the viewer features. Kept as plain
// functions so they can be exercised without a component tree.

export type HeaderMap = Record<string, string[]>

export function contentType(headers: HeaderMap | undefined): string {
  const newLocal = 'content-type';
  const value = headers?.[newLocal]?.[0]
  return value ?? ''
}

export function headerRows(headers: HeaderMap | undefined): { key: string; value: string }[] {
  const rows: { key: string; value: string }[] = []
  for (const [key, values] of Object.entries(headers ?? {})) {
    for (const value of values) rows.push({ key, value })
  }
  return rows
}

const EXTENSIONS = {
  json: 'json',
  xml: 'xml',
  html: 'html',
  text: 'txt',
  csv: 'csv',
  javascript: 'js',
  'x-www-form-urlencoded': 'txt',
} satisfies Record<string, string>

/** suggestedFilename picks a download name for a response body: the
 * Content-Disposition filename when present, else `response.<ext>` derived
 * from the Content-Type. */
export function suggestedFilename(
  headers: HeaderMap | undefined,
  contentType: string,
): string {
  const disposition = headers?.['content-disposition']?.[0]
  const match = disposition?.match(/filename="?([^";]+)"?/i)
  if (match?.[1]) return match[1]
  const type = contentType.split(';')[0].trim().toLowerCase()
  const media = type.split('/')[1] ?? ''
  // SAFETY: EXTENSIONS is an open map; a miss falls back to the default ext.
  const ext = EXTENSIONS[media as keyof typeof EXTENSIONS] ?? 'txt'
  return `response.${ext}`
}

export interface ResponseCookie {
  name: string
  value: string
  domain: string | null
  path: string | null
  expires: string | null
  maxAge: number | null
  secure: boolean
  httpOnly: boolean
  sameSite: string | null
}

function parseCookieDate(value: string): string | null {
  const time = Date.parse(value)
  return Number.isNaN(time) ? null : new Date(time).toISOString()
}

/** parseSetCookies parses every `set-cookie` response header (RFC 6265 §5.2)
 * into structured cookies for display. Attribute values come through raw so
 * the table can show exactly what the server sent; dates are normalized. */
export function parseSetCookies(headers: HeaderMap | undefined): ResponseCookie[] {
  const raw = (headers ?? {})['set-cookie'] ?? []
  const cookies: ResponseCookie[] = []
  for (const header of raw) {
    const parts = header.split(';')
    const first = parts.shift()?.split('=', 2)
    if (!first) continue
    const cookie: ResponseCookie = {
      name: first[0].trim(),
      value: (first[1] ?? '').trim(),
      domain: null,
      path: null,
      expires: null,
      maxAge: null,
      secure: false,
      httpOnly: false,
      sameSite: null,
    }
    for (const part of parts) {
      const [key, ...rest] = part.trim().split('=')
      const value = rest.join('=').trim()
      switch (key.toLowerCase()) {
        case 'domain':
          cookie.domain = value || null
          break
        case 'path':
          cookie.path = value || null
          break
        case 'expires':
          cookie.expires = parseCookieDate(value)
          break
        case 'max-age':
          cookie.maxAge = Number.parseInt(value, 10)
          if (!Number.isFinite(cookie.maxAge)) cookie.maxAge = null
          break
        case 'secure':
          cookie.secure = true
          break
        case 'httponly':
          cookie.httpOnly = true
          break
        case 'samesite':
          cookie.sameSite = value.toLowerCase() || null
          break
        default:
          break
      }
    }
    if (cookie.name) cookies.push(cookie)
  }
  return cookies
}

/** cookieExpiry renders a cookie's Max-Age or Expires attribute as a short
 * label; session cookies return null. */
export function cookieExpiry(cookie: ResponseCookie): string | null {
  if (cookie.maxAge !== null) {
    if (cookie.maxAge <= 0) return 'expired'
    if (cookie.maxAge < 60) return `${cookie.maxAge}s`
    if (cookie.maxAge < 3600) return `${Math.round(cookie.maxAge / 60)}m`
    if (cookie.maxAge < 86400) return `${Math.round(cookie.maxAge / 3600)}h`
    return `${Math.round(cookie.maxAge / 86400)}d`
  }
  if (cookie.expires) {
    const ms = Date.parse(cookie.expires) - Date.now()
    if (Number.isNaN(ms)) return cookie.expires
    if (ms <= 0) return 'expired'
    if (ms < 3600000) return `${Math.max(1, Math.round(ms / 60000))}m`
    if (ms < 86400000) return `${Math.round(ms / 3600000)}h`
    return `${Math.round(ms / 86400000)}d`
  }
  return null
}

/** copyText writes text to the clipboard, falling back to a silent no-op when
 * the Clipboard API is unavailable. */
export async function copyText(text: string): Promise<void> {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    // Clipboard API unavailable or denied — nothing else to try.
  }
}

function looksLikeJson(body: string): boolean {
  const trimmed = body.trim()
  return trimmed.startsWith('{') || trimmed.startsWith('[')
}

function looksLikeXml(body: string): boolean {
  const trimmed = body.trim()
  return trimmed.startsWith('<') && !trimmed.startsWith('{')
}

/** indentXml re-indents a well-formed XML document by tag depth. Returns ''
 * when the input has no element tags. */
export function indentXml(xml: string): string {
  const tokens = xml.replace(/>\s*</g, '><').match(/<[^>]+>|[^<]+/g)
  if (!tokens) return ''
  let depth = 0
  const out: string[] = []
  for (const token of tokens) {
    const t = token.trim()
    if (!t) continue
    if (t.startsWith('<?') || t.startsWith('<!')) {
      out.push(t)
      continue
    }
    const closing = t.startsWith('</')
    const selfClosing = t.endsWith('/>')
    if (closing) depth = Math.max(0, depth - 1)
    out.push('  '.repeat(depth) + t)
    if (!closing && !selfClosing) depth++
  }
  return out.join('\n')
}

/** prettyBody formats a response body: JSON gets pretty-printed, XML gets
 * indented when parseable; anything else returns unchanged. */
export function prettyBody(body: string, contentType: string): string {
  if (contentType.includes('json') || looksLikeJson(body)) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      // Not valid JSON — fall through to raw.
    }
  }
  if (contentType.includes('xml') || looksLikeXml(body)) {
    const indented = indentXml(body)
    if (indented) return indented
  }
  return body
}

/** jsonText renders a parsed JSON value as plain text (for tree search). */
export function jsonText(value: unknown): string {
  if (value === null) return 'null'
  if (typeof value === 'object') {
    if (Array.isArray(value)) {
      return value.map((v) => jsonText(v)).join(' ')
    }
    return Object.entries(value as Record<string, unknown>)
      .map(([k, v]) => `${k} ${jsonText(v)}`)
      .join(' ')
  }
  return String(value)
}

export interface SearchPart {
  text: string
  match: boolean
}

export interface SearchResult {
  count: number
  parts: SearchPart[]
}

export interface TableData {
  columns: string[]
  rows: string[][]
}

/** isTabular reports whether a body should offer the Table view (JSON array-of-objects or CSV). */
export function isTabular(body: string, ct: string): boolean {
  if (ct.includes('csv')) return true
  const t = body.trim()
  if (!t) return false
  if (t.startsWith('[')) {
    try {
      const v = JSON.parse(t)
      return Array.isArray(v) && v.length > 0 && typeof v[0] === 'object' && v[0] !== null && !Array.isArray(v[0])
    } catch {
      return false
    }
  }
  // heuristic CSV: contains comma and newline, first line has comma
  if (t.includes(',') && t.includes('\n')) {
    const first = t.split('\n')[0] ?? ''
    return first.includes(',')
  }
  return false
}

/** parseTable returns columns + rows for a tabular body, or null when not tabular. Caps at 1000 rows. */
export function parseTable(body: string, ct: string): TableData | null {
  const t = body.trim()
  if (ct.includes('csv') || (!t.startsWith('[') && t.includes(','))) {
    // CSV
    const lines = t.split('\n').filter((l) => l.trim() !== '')
    if (lines.length === 0) return null
    const rows = lines.map((l) => l.split(',').map((c) => c.trim()))
    const columns = rows.shift() ?? []
    return { columns, rows: rows.slice(0, 1000) }
  }
  try {
    const v = JSON.parse(t)
    if (!Array.isArray(v) || v.length === 0) return null
    const cols = new Set<string>()
    for (const row of v.slice(0, 1000)) {
      if (typeof row === 'object' && row !== null) {
        for (const k of Object.keys(row as Record<string, unknown>)) cols.add(k)
      }
    }
    const columns = [...cols]
    const rows = (v as Record<string, unknown>[]).slice(0, 1000).map((r) => columns.map((c) => String(r[c] ?? '')))
    return { columns, rows }
  } catch {
    return null
  }
}

export function binaryPreviewType(ct: string): 'image' | 'pdf' | 'hex' | 'none' {
  if (ct.startsWith('image/')) return 'image'
  if (ct.includes('pdf')) return 'pdf'
  if (ct && !ct.includes('json') && !ct.includes('xml') && !ct.includes('text') && !ct.includes('csv')) return 'hex'
  return 'none'
}

/** searchBody splits text into match/non-match segments for highlighting.
 * Returns null when there is nothing to search. */
export function searchBody(text: string, query: string): SearchResult | null {
  const q = query.trim()
  if (!q) return null
  const lower = text.toLowerCase()
  const needle = q.toLowerCase()
  const parts: SearchPart[] = []
  let count = 0
  let i = 0
  while (i < text.length) {
    const idx = lower.indexOf(needle, i)
    if (idx === -1) {
      parts.push({ text: text.slice(i), match: false })
      break
    }
    if (idx > i) parts.push({ text: text.slice(i, idx), match: false })
    parts.push({ text: text.slice(idx, idx + needle.length), match: true })
    count++
    i = idx + needle.length
  }
  if (count === 0) return { count: 0, parts: [{ text, match: false }] }
  return { count, parts }
}
