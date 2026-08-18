import { useMemo, useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { JsonTree } from '../../components/JsonTree'
import { useRequestStore } from '../../stores'
import {
  contentType,
  headerRows,
  prettyBody,
  searchBody,
} from '../../lib/response'

type View = 'raw' | 'pretty' | 'headers' | 'tree'

const views: { id: View; label: string }[] = [
  { id: 'raw', label: 'Raw' },
  { id: 'pretty', label: 'Pretty' },
  { id: 'headers', label: 'Headers' },
  { id: 'tree', label: 'Tree' },
]

const tabClass = (active: boolean) =>
  `rounded-md px-2 py-1 text-xs font-medium transition-colors ${
    active ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
  }`

function statusClass(code: number): string {
  return code >= 400 ? 'text-destructive' : 'text-foreground'
}

function formatBytes(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

export function ResponseViewer() {
  const response = useRequestStore((s) => s.response)
  const loading = useRequestStore((s) => s.loading)
  const error = useRequestStore((s) => s.error)
  const [view, setView] = useState<View>('pretty')
  const [query, setQuery] = useState('')

  const ct = response ? contentType(response.headers) : ''
  const pretty = useMemo(
    () => (response ? prettyBody(response.body, ct) : ''),
    [response, ct],
  )
  const raw = response?.body ?? ''
  const parsed = useMemo(() => {
    if (!response) return null
    try {
      return JSON.parse(response.body) as unknown
    } catch {
      return null
    }
  }, [response])
  const bodyView = view === 'pretty' ? pretty : view === 'raw' ? raw : ''
  const treeFallback = view === 'tree' && parsed === null && response != null
  const headers = response ? headerRows(response.headers) : []

  const statusLine = response
    ? `${response.proto ? `${response.proto} ` : ''}${response.statusCode} ${response.statusText}`
    : ''

  const body = loading
    ? '// Sending request…'
    : error
      ? `// Error: ${error}`
      : response
        ? bodyView
        : '// Send a request to see the response'

  const searchResult = searchBody(bodyView, query)
  const filteredHeaders =
    view === 'headers' && query.trim()
      ? headers.filter(
          (h) =>
            h.key.toLowerCase().includes(query.toLowerCase()) ||
            h.value.toLowerCase().includes(query.toLowerCase()),
        )
      : headers

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between px-2 pb-1 pt-2">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Response
        </p>
        {response ? (
          <p className="text-xs text-muted-foreground">
            <span className={statusClass(response.statusCode)}>{statusLine}</span>
            {' · '}
            {response.durationMs}ms · {formatBytes(response.size)}
            {ct ? ` · ${ct.split(';')[0]}` : ''}
          </p>
        ) : null}
      </div>

      {response ? (
        <div className="flex items-center gap-1 px-2 pb-1">
          {views.map((v) => (
            <button
              key={v.id}
              type="button"
              onClick={() => setView(v.id)}
              className={tabClass(view === v.id)}
            >
              {v.label}
            </button>
          ))}
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search response…"
            className="ml-auto w-44 rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground"
          />
          {searchResult && searchResult.count > 0 ? (
            <span className="text-xs text-muted-foreground">{searchResult.count} matches</span>
          ) : null}
        </div>
      ) : null}

      <div className="min-h-0 flex-1 p-2">
        {loading ? (
          <CodeMirrorEditor
            value={body}
            language="json"
            readOnly
            className="h-full overflow-hidden rounded-md border border-border"
          />
        ) : error ? (
          <CodeMirrorEditor
            value={body}
            language="text"
            readOnly
            className="h-full overflow-hidden rounded-md border border-border"
          />
        ) : view === 'tree' && parsed !== null ? (
          <div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
            <JsonTree data={parsed} filter={query} />
          </div>
        ) : view === 'tree' && treeFallback ? (
          <div className="flex h-full flex-col rounded-md border border-border">
            <p className="px-2 pt-2 text-xs text-muted-foreground">
              This response is not JSON — showing the raw body.
            </p>
            <CodeMirrorEditor
              value={raw}
              language="text"
              readOnly
              className="min-h-0 flex-1 overflow-hidden"
            />
          </div>
        ) : view === 'headers' ? (
          <div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
            {filteredHeaders.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                {query ? 'No headers match your search.' : 'No response headers.'}
              </p>
            ) : (
              <table className="w-full text-left text-xs">
                <tbody>
                  {filteredHeaders.map((h) => (
                    <tr key={`${h.key}-${h.value}`} className="border-b border-border/50 last:border-0">
                      <td className="py-1 pr-3 align-top font-mono text-muted-foreground">
                        {h.key}
                      </td>
                      <td className="py-1 font-mono text-foreground break-all">{h.value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        ) : response ? (
          <CodeMirrorEditor
            value={body}
            language="json"
            readOnly
            className="h-full overflow-hidden rounded-md border border-border"
          />
        ) : (
          <CodeMirrorEditor
            value={body}
            language="text"
            readOnly
            className="h-full overflow-hidden rounded-md border border-border"
          />
        )}
      </div>

      {view !== 'tree' && view !== 'headers' && response && query && searchResult && searchResult.count === 0 ? (
        <p className="px-2 pb-1 text-xs text-muted-foreground">No matches in the response body.</p>
      ) : null}
    </div>
  )
}
