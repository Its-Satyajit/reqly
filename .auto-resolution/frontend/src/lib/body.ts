// Request body type + encoders shared by the builder. Kept as plain
// functions so they can be exercised without a component tree.

import type { KeyValueRow } from './request'
import type { JsonValue } from './typeGuards'

export type BodyType = 'none' | 'json' | 'xml' | 'form-data' | 'urlencoded' | 'raw' | 'binary' | 'graphql'

export const bodyTypes: { id: BodyType; label: string }[] = [
  { id: 'none', label: 'None' },
  { id: 'json', label: 'JSON' },
  { id: 'xml', label: 'XML' },
  { id: 'form-data', label: 'Form data' },
  { id: 'urlencoded', label: 'URL encoded' },
  { id: 'raw', label: 'Raw text' },
  { id: 'binary', label: 'Binary' },
  { id: 'graphql', label: 'GraphQL' },
]

const CONTENT_TYPES = {
  json: 'application/json',
  xml: 'application/xml',
  urlencoded: 'application/x-www-form-urlencoded',
  graphql: 'application/json',
  binary: 'application/octet-stream',
} satisfies Record<Exclude<BodyType, 'none' | 'form-data' | 'raw'>, string>

/** contentTypeFor returns the Content-Type header a body type implies, or ''
 * when the type has no canonical type (raw text stays untyped, multipart needs
 * a boundary so it is handled by the encoder). */
export function contentTypeFor(type: BodyType): string {
  if (type === 'form-data') return 'multipart/form-data'
  // SAFETY: BodyType narrowed to keys of CONTENT_TYPES via Exclude union; fallback handles non-key
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
 * given boundary. Disabled rows and blank keys are dropped. File rows
 * (`row.file`) produce a `filename` + `Content-Type` part; the file bytes
 * are read at send time by the Go core (or bridge) from the Git-native path. */
export function encodeFormData(rows: KeyValueRow[], boundary: string): string {
  const parts: string[] = []
  for (const row of rows) {
    if (!row.enabled || row.key.trim() === '') continue
    if (row.file) {
      const filename = row.filename || row.file.split('/').pop() || 'file'
      const contentType = mimeForFile(row.file) || 'application/octet-stream'
      // File bytes are read at send; the placeholder is the file path.
      parts.push(
        `--${boundary}\r\nContent-Disposition: form-data; name="${row.key}"; filename="${filename}"\r\nContent-Type: ${contentType}\r\n\r\n[FILE:${row.file}]`,
      )
    } else {
      parts.push(
        `--${boundary}\r\nContent-Disposition: form-data; name="${row.key}"\r\n\r\n${row.value}`,
      )
    }
  }
  if (parts.length === 0) return ''
  return `${parts.join('\r\n')}\r\n--${boundary}--\r\n`
}

/** mimeForFile returns a mime type from a file extension, or '' if unknown. */
function mimeForFile(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'json':
      return 'application/json'
    case 'xml':
      return 'application/xml'
    case 'txt':
      return 'text/plain'
    case 'html':
      return 'text/html'
    case 'csv':
      return 'text/csv'
    case 'png':
      return 'image/png'
    case 'jpg':
    case 'jpeg':
      return 'image/jpeg'
    case 'pdf':
      return 'application/pdf'
    default:
      return ''
  }
}

export interface SerializedBody {
  /** The request body, or undefined when the type produces nothing. */
  body?: string
  /** Implied Content-Type for the body type, or '' when none applies. */
  contentType: string
}

/** serializeBody turns a RequestInput's body type + fields into the wire body
 * and implied Content-Type. 'none' yields no body; json/xml/raw/binary/graphql
 * send the editor text; form-data/urlencoded encode the key-value rows. */
export function serializeBody(input: {
  bodyType?: BodyType
  body?: string
  form?: KeyValueRow[]
  graphqlQuery?: string
  graphqlVariables?: string
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
    case 'binary': {
      const body = input.body ?? ''
      if (body === '') return { contentType: '' }
      const mime = mimeForFile(body) || contentTypeFor('binary')
      return { body, contentType: mime }
    }
    case 'graphql': {
      const query = input.graphqlQuery ?? input.body ?? ''
      if (query.trim() === '') return { contentType: '' }
      type GraphQLVariables = Record<string, JsonValue>
      let variables: GraphQLVariables = {}
      if (input.graphqlVariables?.trim()) {
        try {
          // SAFETY: GraphQL variables are parsed JSON object at I/O boundary; JSON.parse yields JsonValue map
          variables = JSON.parse(input.graphqlVariables) as GraphQLVariables
        } catch {
          // Invalid JSON will be surfaced as save warning; send empty variables.
          variables = {}
        }
      } else if (input.form?.length) {
        // Fallback: variables from form rows (key-value) if provided.
        const vars: Record<string, string> = {}
        for (const row of input.form) {
          if (row.enabled && row.key.trim() !== '') vars[row.key] = row.value
        }
        // SAFETY: string-valued map widens to JsonValue-valued variables; string is assignable to JsonValue
        variables = vars as GraphQLVariables
      }
      const body = JSON.stringify({ query, variables })
      return { body, contentType: contentTypeFor('graphql') }
    }
    default: {
      const body = input.body ?? ''
      if (body === '') return { contentType: '' }
      return { body, contentType: contentTypeFor(input.bodyType!) }
    }
  }
}
