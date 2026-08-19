import { useMemo, useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { JsonTree } from '../../components/JsonTree'
import { Button } from '../../components/ui/button'
import { useRequestStore, useWorkspaceStore } from '../../stores'
import {
  contentType,
  headerRows,
  parseSetCookies,
  prettyBody,
  searchBody,
  suggestedFilename,
  copyText,
  cookieExpiry,
} from '../../lib/response'
import { queryJSONPath, type JSONPathMatch } from '../../lib/jsonpath'

type View = 'raw' | 'pretty' | 'headers' | 'tree' | 'cookies'

const views: { id: View; label: string }[] = [
  { id: 'raw', label: 'Raw' },
  { id: 'pretty', label: 'Pretty' },
  { id: 'headers', label: 'Headers' },
  { id: 'tree', label: 'Tree' },
  { id: 'cookies', label: 'Cookies' },
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
  const activeTabId = useWorkspaceStore((s) => s.activeTabId)
  const tabState = useRequestStore((s) => (activeTabId ? s.responses[activeTabId] : undefined))
  const response = tabState?.response ?? null
  const loading = tabState?.loading ?? false
  const error = tabState?.error ?? null
  const [view, setView] = useState<View>('pretty')
  const [query, setQuery] = useState('')
  const [jsonPath, setJsonPath] = useState('')
  const [copied, setCopied] = useState(false)

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

  const filename = response ? suggestedFilename(response.headers, ct) : ''
  const headersText = headers.map((h) => `${h.key}: ${h.value}`).join('\n')
  const cookies = useMemo(
    () => (response ? parseSetCookies(response.headers) : []),
    [response],
  )
  const cookiesText = cookies
    .map(
      (c) =>
        `${c.name}=${c.value}; ${[
          c.domain ? `Domain=${c.domain}` : '',
          c.path ? `Path=${c.path}` : '',
          c.secure ? 'Secure' : '',
          c.httpOnly ? 'HttpOnly' : '',
        ]
          .filter(Boolean)
          .join('; ')}`,
    )
    .join('\n')
  const jsonPathResult = useMemo(() => {
    if (!parsed || !jsonPath.trim()) return null
    return queryJSONPath(parsed, jsonPath)
  }, [parsed, jsonPath])

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

  const filteredCookies =
    view === 'cookies' && query.trim()
      ? cookies.filter((c) =>
          `${c.name} ${c.value} ${c.domain ?? ''} ${c.path ?? ''}`
            .toLowerCase()
            .includes(query.toLowerCase()),
        )
      : cookies

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

      {response ? (
        <div className="flex items-center gap-1 border-t border-border/50 px-2 py-1">
          <Button
            size="xs"
            variant="ghost"
            onClick={() => {
              const text =
                view === 'headers' ? headersText : view === 'cookies' ? cookiesText : bodyView
              void copyText(text)
              setCopied(true)
              setTimeout(() => setCopied(false), 1500)
            }}
          >
            {copied ? 'Copied' : 'Copy'}
          </Button>
          <Button
            size="xs"
            variant="ghost"
            onClick={() => {
              void copyText(headersText)
              setCopied(true)
              setTimeout(() => setCopied(false), 1500)
            }}
          >
            Copy headers
          </Button>
          <Button
            size="xs"
            variant="ghost"
            onClick={() => {
              const blob = new Blob([raw], {
                type: ct || 'application/octet-stream',
              })
              const url = URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = url
              a.download = filename
              a.click()
              URL.revokeObjectURL(url)
            }}
          >
            Download
          </Button>
          <Button size="xs" variant="ghost" onClick={() => setView('pretty')}>
            Format
          </Button>
          <span className="ml-auto flex items-center gap-1 text-xs text-muted-foreground">
            <span>JSONPath</span>
            <input
              value={jsonPath}
              onChange={(e) => setJsonPath(e.target.value)}
              placeholder="$.users[*].name"
              disabled={parsed === null}
              spellCheck={false}
              className="w-48 rounded-md border border-input bg-background px-2 py-1 font-mono text-xs text-foreground placeholder:text-muted-foreground disabled:opacity-50"
            />
          </span>
        </div>
      ) : null}

      {parsed === null && response && jsonPath.trim() ? (
        <p className="border-t border-border/50 px-2 py-1 text-xs text-muted-foreground">
          This response is not JSON — JSONPath queries need a JSON body.
        </p>
      ) : null}

      <div className="min-h-0 flex-1 p-2">
        {jsonPathResult && !jsonPathResult.error && jsonPathResult.matches.length > 0 ? (
          <div className="flex h-full flex-col gap-1 overflow-y-auto rounded-md border border-border bg-background p-2">
            {jsonPathResult.matches.map((m) => (
              <JsonPathMatchRow key={m.path} match={m} />
            ))}
          </div>
        ) : jsonPathResult?.error ? (
          <div className="flex h-full items-start rounded-md border border-border bg-background p-2">
            <p className="text-xs text-destructive">{jsonPathResult.error}</p>
          </div>
        ) : jsonPathResult ? (
          <div className="flex h-full items-start rounded-md border border-border bg-background p-2">
            <p className="text-xs text-muted-foreground">No matches for this path.</p>
          </div>
        ) : loading ? (
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
        ) : view === 'cookies' ? (
          <div className="h-full overflow-y-auto rounded-md border border-border bg-background p-2">
            {filteredCookies.length === 0 ? (
              <div className="flex h-full flex-col items-start justify-center gap-2 px-4">
                <p className="text-sm font-medium text-foreground">
                  {cookies.length === 0
                    ? 'No cookies set by this response.'
                    : 'No cookies match your search.'}
                </p>
                {cookies.length === 0 ? (
                  <p className="max-w-sm text-xs text-muted-foreground">
                    Servers set cookies via <code className="font-mono">Set-Cookie</code> response
                    headers. Send a request to an endpoint that sets a cookie to see it here —
                    persistence is a separate roadmap item.
                  </p>
                ) : null}
              </div>
            ) : (
              <table className="w-full text-left text-xs">
                <tbody>
                  {filteredCookies.map((c) => (
                    <tr key={`${c.name}-${c.value}-${c.domain ?? ''}`} className="border-b border-border/50 last:border-0">
                      <td className="py-1 pr-3 align-top font-mono text-foreground">{c.name}</td>
                      <td className="py-1 pr-3 font-mono text-muted-foreground break-all">{c.value}</td>
                      <td className="py-1 pr-3 align-top text-muted-foreground">
                        {c.domain ?? '—'}
                      </td>
                      <td className="py-1 pr-3 align-top font-mono text-muted-foreground">
                        {c.path ?? '/'}
                      </td>
                      <td className="py-1 pr-3 align-top text-muted-foreground">
                        {cookieExpiry(c) ?? 'Session'}
                      </td>
                      <td className="py-1 align-top whitespace-nowrap text-muted-foreground">
                        {[
                          c.secure ? 'Secure' : '',
                          c.httpOnly ? 'HttpOnly' : '',
                          c.sameSite ? `SameSite=${c.sameSite}` : '',
                        ]
                          .filter(Boolean)
                          .join(' · ') || '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
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

      {view !== 'tree' && view !== 'headers' && view !== 'cookies' && response && query && searchResult && searchResult.count === 0 ? (
        <p className="px-2 pb-1 text-xs text-muted-foreground">No matches in the response body.</p>
      ) : null}
    </div>
  )
}

function JsonPathMatchRow({ match }: { match: JSONPathMatch }) {
  const text = useMemo(
    () =>
      match.value === null
        ? 'null'
        : typeof match.value === 'object'
          ? JSON.stringify(match.value, null, 2)
          : String(match.value),
    [match.value],
  )
  return (
    <div className="flex flex-col gap-0.5 rounded-md border border-border/50 bg-background px-2 py-1">
      <p className="font-mono text-xs text-muted-foreground">{match.path}</p>
      <pre className="overflow-x-auto whitespace-pre-wrap font-mono text-xs text-foreground">
        {text}
      </pre>
    </div>
  )
}
