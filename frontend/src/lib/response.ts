// Response-view helpers shared by the viewer features. Kept as plain
// functions so they can be exercised without a component tree.

export type HeaderMap = Record<string, string[]>

export function contentType(headers: HeaderMap | undefined): string {
  const value = (headers ?? {})['content-type']?.[0]
  return value ?? ''
}

export function headerRows(headers: HeaderMap | undefined): { key: string; value: string }[] {
  const rows: { key: string; value: string }[] = []
  for (const [key, values] of Object.entries(headers ?? {})) {
    for (const value of values) rows.push({ key, value })
  }
  return rows
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
