// Request body type + encoders shared by the builder. Kept as plain
// functions so they can be exercised without a component tree.

import type { KeyValueRow } from './request'

export type BodyType = 'none' | 'json' | 'xml' | 'form-data' | 'urlencoded' | 'raw'

export const bodyTypes: { id: BodyType; label: string }[] = [
  { id: 'none', label: 'None' },
  { id: 'json', label: 'JSON' },
  { id: 'xml', label: 'XML' },
  { id: 'form-data', label: 'Form data' },
  { id: 'urlencoded', label: 'URL encoded' },
  { id: 'raw', label: 'Raw text' },
]

const CONTENT_TYPES: Record<Exclude<BodyType, 'none' | 'form-data' | 'raw'>, string> = {
  json: 'application/json',
  xml: 'application/xml',
  urlencoded: 'application/x-www-form-urlencoded',
}

/** contentTypeFor returns the Content-Type header a body type implies, or ''
 * when the type has no canonical type (raw text stays untyped, multipart needs
 * a boundary so it is handled by the encoder). */
export function contentTypeFor(type: BodyType): string {
  if (type === 'form-data') return 'multipart/form-data'
  return CONTENT_TYPES[type as keyof typeof CONTENT_TYPES] ?? ''
}

/** boundaryFor generates a multipart boundary token for a form-data body. */
export function boundaryFor(): string {
  return `reqly-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

/** encodeUrlEncoded serializes rows into application/x-www-form-urlencoded
 * form, dropping disabled rows and blank keys. */
export function encodeUrlEncoded(rows: KeyValueRow[]): string {
  const parts: string[] = []
  for (const row of rows) {
    if (!row.enabled || row.key.trim() === '') continue
    parts.push(`${encodeURIComponent(row.key)}=${encodeURIComponent(row.value)}`)
  }
  return parts.join('&')
}

/** encodeFormData serializes rows into a multipart/form-data body with the
 * given boundary. Disabled rows and blank keys are dropped. */
export function encodeFormData(rows: KeyValueRow[], boundary: string): string {
  const parts: string[] = []
  for (const row of rows) {
    if (!row.enabled || row.key.trim() === '') continue
    parts.push(
      `--${boundary}\r\nContent-Disposition: form-data; name="${row.key}"\r\n\r\n${row.value}`,
    )
  }
  if (parts.length === 0) return ''
  return `${parts.join('\r\n')}\r\n--${boundary}--\r\n`
}

export interface SerializedBody {
  /** The request body, or undefined when the type produces nothing. */
  body?: string
  /** Implied Content-Type for the body type, or '' when none applies. */
  contentType: string
}

/** serializeBody turns a RequestInput's body type + fields into the wire body
 * and implied Content-Type. 'none' yields no body; json/xml/raw send the
 * editor text; form-data/urlencoded encode the key-value rows. */
export function serializeBody(input: {
  bodyType?: BodyType
  body?: string
  form?: KeyValueRow[]
}): SerializedBody {
  switch (input.bodyType ?? 'none') {
    case 'none':
      return { contentType: '' }
    case 'form-data': {
      const boundary = boundaryFor()
      const body = encodeFormData(input.form ?? [], boundary)
      if (body === '') return { contentType: '' }
      return { body, contentType: `multipart/form-data; boundary=${boundary}` }
    }
    case 'urlencoded': {
      const body = encodeUrlEncoded(input.form ?? [])
      if (body === '') return { contentType: '' }
      return { body, contentType: 'application/x-www-form-urlencoded' }
    }
    default: {
      const body = input.body ?? ''
      if (body === '') return { contentType: '' }
      return { body, contentType: contentTypeFor(input.bodyType!) }
    }
  }
}
