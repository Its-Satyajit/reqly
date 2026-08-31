import { useEffect, useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { Button } from '../../components/ui/button'
import { CompactSelect } from '../../components/CompactSelect'
import { KeyValueEditor } from '../../components/KeyValueEditor'
import { ChevronRight, Loader2, SlidersHorizontal, Sparkles } from 'lucide-react'
import { RequestSettingsDialog } from './RequestSettingsDialog'
import { AuthEditor } from '../auth-editor/AuthEditor'
import { authWarnings } from '../../lib/authSchemes'
import { useRequestStore, useWorkspaceStore } from '../../stores'
import { tabIsDirty } from '../../stores/useRequestStore'
import { effectiveUrlFor } from '../../stores/useWorkspaceStore'
import { sentRows } from '../../lib/request'
import { bodyTypes, type BodyType } from '../../lib/body'
import type { KeyValueRow, RequestAuth, RequestRetry } from '../../lib/request'
import type { ResolvedVariable } from '../../lib/collections'
import { TagPicker } from '../../components/TagPicker'
import { tagWarnings } from '../../lib/tags'
import { generateCode } from '../../lib/codegen'
import { copyText } from '../../lib/response'
import { notifyError } from '../../lib/notify'
import { handleTabArrowKeys, tabClass } from '../../lib/ui'
import { cn } from '../../lib/utils'
import { TemplatePickerSheet } from './TemplatePickerSheet'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'CONNECT', 'TRACE'] as const

type Tab = 'params' | 'headers' | 'auth' | 'body' | 'variables' | 'pre-request' | 'tests' | 'docs' | 'settings'

const tabs: { id: Tab; label: string }[] = [
  { id: 'params', label: 'Params' },
  { id: 'headers', label: 'Headers' },
  { id: 'body', label: 'Body' },
  { id: 'auth', label: 'Auth' },
]

const overflowTabs: { id: Tab; label: string }[] = [
  { id: 'pre-request', label: 'Pre-request' },
  { id: 'tests', label: 'Tests' },
  { id: 'docs', label: 'Docs' },
  { id: 'settings', label: 'Settings' },
  { id: 'variables', label: 'Variables' },
]

const bodyLanguage = {
  none: 'text',
  json: 'json',
  xml: 'xml',
  'form-data': 'text',
  urlencoded: 'text',
  raw: 'text',
  text: 'text',
  html: 'text',
  binary: 'text',
  graphql: 'graphql',
} satisfies Record<BodyType, 'json' | 'xml' | 'text' | 'graphql'>

/** saveWarnings validates a draft before it is persisted. Warnings do not
 * block a save — they flag values that would survive onto disk (unknown
 * method, malformed body, incomplete auth config) so the user can fix them
 * instead of persisting garbage. */
function saveWarnings(draft: {
  method: string
  url: string
  bodyType: BodyType
  body: string
  form?: KeyValueRow[]
  graphqlQuery?: string
  graphqlVariables?: string
  auth?: RequestAuth
  params?: KeyValueRow[]
  headers?: KeyValueRow[]
}): string[] {
  const warnings: string[] = []
  if (!methods.some((m) => m === draft.method)) {
    warnings.push(`Unknown method "${draft.method}" will be written to the file.`)
  }
  // dynamic tag unknowns across url/body/headers/params
  const tagSources = [draft.url, draft.body, draft.graphqlQuery ?? "", draft.graphqlVariables ?? "", ...(draft.headers ?? []).map((h) => `${h.key} ${h.value}`), ...(draft.params ?? []).map((p) => `${p.key} ${p.value}`)]
  for (const src of tagSources) {
    warnings.push(...tagWarnings(src))
  }
  if (draft.bodyType === 'json' && draft.body.trim() !== '') {
    try {
      JSON.parse(draft.body)
    } catch {
      warnings.push('The JSON body is malformed and will be saved as-is.')
    }
  }
  if (draft.bodyType === 'binary' && draft.body.trim() === '') {
    warnings.push('Binary body requires a file path.')
  }
  if (draft.bodyType === 'graphql') {
    const vars = draft.graphqlVariables ?? ''
    if (vars.trim() !== '') {
      try {
        JSON.parse(vars)
      } catch {
        warnings.push('GraphQL variables are not valid JSON and will be saved as-is.')
      }
    }
  }
  if (draft.bodyType === 'form-data' && draft.form) {
    for (const row of draft.form) {
      if (row.enabled && row.file !== undefined && !row.file.trim()) {
        warnings.push(`Form-data field "${row.key || 'unnamed'}" has an empty file path.`)
      }
    }
  }
  warnings.push(...authWarnings(draft.auth))
  return warnings
}

export function RequestEditor() {
  const activeTabId = useWorkspaceStore((s) => s.activeTabId)
  const draft = useRequestStore((s) => (activeTabId ? s.drafts[activeTabId] : undefined))
  const meta = useRequestStore((s) => (activeTabId ? s.meta[activeTabId] : undefined))
  const loading = useRequestStore((s) => (activeTabId ? s.responses[activeTabId]?.loading : false))
  const updateDraft = useRequestStore((s) => s.updateDraft)
  const send = useRequestStore((s) => s.send)
  const cancel = useRequestStore((s) => s.cancel)
  const saveRequest = useWorkspaceStore((s) => s.saveRequest)
  const overwriteRequest = useWorkspaceStore((s) => s.overwriteRequest)
  const reloadRequest = useWorkspaceStore((s) => s.reloadRequest)
  const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId)
  const environments = useWorkspaceStore((s) => s.environments)
  const [tab, setTab] = useState<Tab>('params')
  const [codeLang, setCodeLang] = useState<'curl' | 'js' | 'python' | 'go'>('curl')
  const [copiedCode, setCopiedCode] = useState(false)
  const [templatePickerOpen, setTemplatePickerOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
        e.preventDefault()
        void saveRequest(activeTabId ?? '')
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [activeTabId, saveRequest])

  if (!activeTabId || !draft) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 p-4">
        <p className="text-sm font-medium text-foreground">No request open</p>
        <p className="max-w-sm text-center text-xs text-muted-foreground">
          Open a request from the sidebar or create a new one to start sending.
        </p>
      </div>
    )
  }

  const envPill = meta?.env ?? environments.find((e) => e.id === activeEnvironmentId)?.name ?? null
  const showVariables = (meta?.variables.length ?? 0) > 0
  const dirty = tabIsDirty(draft, meta)
  const requestPath = meta?.requestPath
  const canSave =
    Boolean(requestPath) && dirty && draft.url.trim() !== '' && !meta?.changedOnDisk
  const warnings = saveWarnings(draft)

  const handleSend = () => {
    void send(activeTabId, {
      method: draft.method,
      url: draft.url,
      params: sentRows(draft.params),
      headers: sentRows(draft.headers).map(({ key, value }) => ({ key, value })),
      bodyType: draft.bodyType,
      body: draft.body,
      form: sentRows(draft.form),
      graphqlQuery: draft.graphqlQuery,
      graphqlVariables: draft.graphqlVariables,
      env: meta?.env,
      requestPath,
      auth: draft.auth,
      proxy: draft.proxy,
      tls: draft.tls,
      timeout: draft.timeout,
      followRedirects: draft.followRedirects,
    })
  }

  const patch = (p: Partial<Parameters<typeof updateDraft>[1]>) => updateDraft(activeTabId, p)

  return (
    <div className="flex h-full flex-col">
      {meta?.changedOnDisk && (
        <div className="flex items-center justify-between gap-2 border-b border-status-warn/40 bg-status-warn/10 px-3 py-1.5">
          <p className="text-xs text-status-warn">
            This request changed on disk since you opened it. Overwrite the
            file, or reload to keep the on-disk version.
          </p>
          <div className="flex shrink-0 gap-1">
            <Button
              size="sm"
              variant="outline"
              onClick={() => void reloadRequest(activeTabId)}
            >
              Reload
            </Button>
            <Button
              size="sm"
              variant="default"
              onClick={() => void overwriteRequest(activeTabId)}
            >
              Overwrite
            </Button>
          </div>
        </div>
      )}
      {requestPath && dirty && warnings.length > 0 && (
        <div className="flex flex-col gap-0.5 border-b border-status-warn/40 bg-status-warn/10 px-3 py-1.5">
          {warnings.map((w) => (
            <p key={w} className="text-xs text-status-warn">
              {w}
            </p>
          ))}
        </div>
      )}
      <div className="flex flex-col gap-2 border-b border-border/80 bg-background p-2.5">
        <div className="flex min-w-0 items-center gap-1.5">
          <div className="flex min-w-0 flex-1 items-center rounded border border-input bg-background focus-within:border-ring focus-within:ring-1 focus-within:ring-ring">
            <CompactSelect
              value={draft.method}
              onChange={(method) => patch({ method })}
              ariaLabel="HTTP method"
              className="w-24 shrink-0 rounded-r-none border-0 border-r border-input bg-muted/40 font-mono text-xs font-bold tracking-wide"
              options={methods.map((m) => ({ value: m, label: m }))}
            />
            <input
              value={draft.url}
              onChange={(e) => patch({ url: e.target.value })}
              onKeyDown={(e) => {
                if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                  e.preventDefault()
                  if (!loading) handleSend()
                }
              }}
              placeholder="https://api.example.com/v1/resource or {{baseUrl}}/users"
              spellCheck={false}
              aria-label="Request URL"
              className="min-w-0 flex-1 border-0 bg-transparent px-2.5 py-1.5 font-mono text-xs text-foreground placeholder:text-muted-foreground/60 focus:outline-none"
            />
            <div className="flex shrink-0 items-center pr-1.5">
              <span
                title={meta?.env ? 'Environment pinned by the request file' : 'Environment from the app header'}
                className="rounded border border-border/60 bg-muted/40 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground"
              >
                {envPill ?? 'No env'}
              </span>
            </div>
          </div>

          <Button
            size="sm"
            variant="ghost"
            onClick={() => setTemplatePickerOpen(true)}
            title="Insert template snippet"
            aria-label="Insert template"
            className="h-8 gap-1 px-2 text-xs text-muted-foreground hover:text-foreground"
          >
            <Sparkles className="size-3 text-primary" aria-hidden />
            <span className="hidden sm:inline text-[11px]">Templates</span>
          </Button>

          {requestPath && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => void saveRequest(activeTabId)}
              disabled={!canSave}
              className="h-8 px-2.5 text-xs font-mono"
            >
              {dirty ? 'Save' : 'Saved'}
            </Button>
          )}

          {loading ? (
            <Button size="sm" variant="destructive" onClick={() => void cancel(activeTabId)} className="h-8 gap-1.5 px-3 text-xs font-semibold">
              <Loader2 className="size-3.5 animate-spin" aria-hidden />
              <span>Cancel</span>
            </Button>
          ) : (
            <Button size="sm" onClick={handleSend} className="h-8 gap-1 px-3.5 text-xs font-semibold tracking-wide">
              <span>Send</span>
              <kbd className="hidden sm:inline font-mono text-[10px] opacity-70">⌘⏎</kbd>
            </Button>
          )}

          <div className="hidden lg:flex items-center gap-1 border-l border-border/70 pl-1.5">
            <CompactSelect
              value={codeLang}
              onChange={(v) => {
                if (v === "curl" || v === "js" || v === "python" || v === "go") setCodeLang(v)
              }}
              ariaLabel="Snippet language"
              className="h-8 w-16 text-[11px] font-mono"
              options={[
                { value: "curl", label: "cURL" },
                { value: "js", label: "JS" },
                { value: "python", label: "Python" },
                { value: "go", label: "Go" },
              ]}
            />
            <Button
              size="xs"
              variant="ghost"
              className="h-8 px-2 text-[11px] text-muted-foreground hover:text-foreground"
              onClick={() => {
                const code = generateCode(
                  {
                    method: draft.method,
                    url: draft.url,
                    headers: sentRows(draft.headers).map(({ key, value }) => ({ key, value })),
                    query: sentRows(draft.params).map(({ key, value }) => ({ key, value })),
                    body: draft.body,
                    auth: draft.auth,
                  },
                  codeLang,
                )
                void copyText(code).then((ok) => {
                  if (ok) {
                    setCopiedCode(true)
                    setTimeout(() => setCopiedCode(false), 1500)
                  } else {
                    notifyError('Copy failed', 'Clipboard access was denied — copy the snippet manually.')
                  }
                })
              }}
            >
              {copiedCode ? 'Copied' : 'Copy'}
            </Button>
          </div>
        </div>

        {requestPath && (
          <div className="flex items-center gap-1.5 px-0.5 text-[11px]">
            <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
              Resolved:
            </span>
            <code className="min-w-0 flex-1 truncate font-mono text-muted-foreground">
              {effectiveUrlFor(draft.url, meta?.baseUrl ?? '')}
            </code>
          </div>
        )}
      </div>

      {settingsOpen && (
        <RequestSettingsDialog
          settings={{ timeout: draft.timeout, followRedirects: draft.followRedirects }}
          onApply={(s) => patch(s)}
          onClose={() => setSettingsOpen(false)}
        />
      )}

      {/* Clean single tab bar with indicators and secondary action bar */}
      <div className="flex items-center justify-between border-b border-border/80 bg-muted/10 px-2 py-1">
        <div
          className="flex items-center gap-1"
          role="tablist"
          aria-label="Request sections"
          onKeyDown={(e) => handleTabArrowKeys(e)}
        >
          {tabs.map((t) => {
            const count =
              t.id === 'params'
                ? draft.params.filter((p) => p.enabled && p.key.trim()).length
                : t.id === 'headers'
                ? draft.headers.filter((h) => h.enabled && h.key.trim()).length
                : t.id === 'body' && draft.bodyType !== 'none'
                ? 1
                : t.id === 'auth' && draft.auth && draft.auth.type !== 'none' && draft.auth.type !== 'inherit'
                ? 1
                : 0
            return (
              <button
                key={t.id}
                type="button"
                role="tab"
                aria-selected={tab === t.id}
                tabIndex={tab === t.id ? 0 : -1}
                onClick={() => setTab(t.id)}
                className={cn(
                  "relative flex items-center gap-1.5 rounded px-2.5 py-1 text-xs font-medium transition-colors select-none",
                  tab === t.id
                    ? "bg-background text-foreground shadow-xs font-semibold"
                    : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                )}
              >
                {t.label}
                {count > 0 && (
                  <span className="flex size-1.5 rounded-full bg-primary" />
                )}
              </button>
            )
          })}
          <div className="relative ml-0.5">
            <button
              type="button"
              aria-label="More request sections"
              aria-haspopup="menu"
              onClick={() => setTab((prev) => (overflowTabs.some((o) => o.id === prev) ? prev : overflowTabs[0].id))}
              className={cn(
                "rounded px-2 py-1 text-xs transition-colors select-none",
                overflowTabs.some((o) => o.id === tab)
                  ? "bg-background text-foreground font-semibold shadow-xs"
                  : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
              )}
              title="More: Pre-request, Tests, Docs, Settings, Variables"
            >
              ⋮ More
            </button>
          </div>
          {overflowTabs.some((o) => o.id === tab) && (
            <div className="flex items-center gap-1 pl-1 border-l border-border/60">
              {overflowTabs
                .filter((t) => t.id !== 'variables' || showVariables)
                .map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    role="tab"
                    aria-selected={tab === t.id}
                    tabIndex={tab === t.id ? 0 : -1}
                    onClick={() => setTab(t.id)}
                    className={cn(
                      "rounded px-2 py-0.5 text-xs transition-colors",
                      tab === t.id
                        ? "bg-background text-foreground font-medium shadow-xs"
                        : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                    )}
                  >
                    {t.label}
                  </button>
                ))}
            </div>
          )}
        </div>

        {/* Right side contextual shortcuts: Tag picker & settings */}
        <div className="flex items-center gap-2">
          <TagPicker onInsert={(tag) => patch({ url: draft.url + tag })} />
          <div className="h-3.5 w-px bg-border/80" />
          <button
            type="button"
            onClick={() => setSettingsOpen(true)}
            title="Request settings…"
            aria-label="Request settings"
            className="flex items-center gap-1 rounded border border-border/80 bg-background/50 px-1.5 py-0.5 text-[10px] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground font-mono"
          >
            <SlidersHorizontal className="size-3" aria-hidden />
            {settingsSummary(draft)}
          </button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-2">
        {tab === 'params' ? (
          <KeyValueEditor
            rows={draft.params}
            onChange={(rows) => patch({ params: rows })}
            keyPlaceholder="param"
            valuePlaceholder="value"
          />
        ) : tab === 'headers' ? (
          <div className="flex h-full flex-col gap-2">
            <KeyValueEditor
              rows={draft.headers}
              onChange={(rows) => patch({ headers: rows })}
              keyPlaceholder="header"
              valuePlaceholder="value"
            />
            {(meta?.inheritedHeaders.length ?? 0) > 0 && (
              <div className="shrink-0">
                <p className="mb-1 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
                  Inherited from workspace / collection / folder — read-only
                </p>
                <div className="overflow-hidden rounded-md border border-border">
                  {meta!.inheritedHeaders.map((h, i) => (
                    <div
                      key={`${h.key}:${h.value}:${i}`}
                      className={`flex items-center gap-2 px-2 py-1 text-xs ${
                        i % 2 === 0 ? 'bg-muted/30' : ''
                      }`}
                    >
                      <span className="shrink-0 font-medium text-foreground">{h.key}</span>
                      <code className="min-w-0 flex-1 truncate text-muted-foreground">
                        {h.value}
                      </code>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : tab === 'auth' ? (
          <AuthEditor
            auth={draft.auth}
            onChange={(auth) => patch({ auth })}
            inherited={meta?.auth}
          />
        ) : tab === 'variables' ? (
          <VariablesView
            variables={meta?.variables ?? []}
            env={meta?.env ?? null}
            headerEnv={environments.find((e) => e.id === activeEnvironmentId)?.name ?? null}
          />
        ) : tab === 'pre-request' || tab === 'tests' || tab === 'docs' ? (
          <div className="flex h-full items-center justify-center p-4">
            <p className="text-xs text-muted-foreground">
              {tab === 'pre-request' && 'Pre-request scripts — coming soon.'}
              {tab === 'tests' && 'Tests — coming soon.'}
              {tab === 'docs' && 'Docs — coming soon.'}
            </p>
          </div>
        ) : tab === 'settings' ? (
          <div className="flex flex-col gap-4 p-3 max-w-lg">
            <div className="flex flex-col gap-2 rounded border border-border/80 bg-muted/10 p-3">
              <span className="text-[11px] font-semibold uppercase tracking-wide text-foreground">Proxy & TLS</span>
              <label className="flex flex-col gap-1 text-xs">
                <span className="font-mono text-[10px] uppercase tracking-wide text-muted-foreground">Proxy URL</span>
                <input value={draft.proxy ?? ''} onChange={(e) => patch({ proxy: e.target.value || undefined })} placeholder="http://proxy:8080" spellCheck={false} className="rounded border border-input bg-background px-2 py-1.5 font-mono text-xs" />
              </label>
              <label className="flex items-center gap-2 text-xs pt-1">
                <input type="checkbox" checked={Boolean(draft.tls?.insecureSkipVerify)} onChange={(e) => patch({ tls: { ...draft.tls, insecureSkipVerify: e.target.checked } })} className="size-3.5 rounded" />
                <span>Insecure skip TLS certificate verification</span>
              </label>
              <label className="flex flex-col gap-1 text-xs pt-1">
                <span className="font-mono text-[10px] uppercase tracking-wide text-muted-foreground">CA certificate file (workspace-relative PEM)</span>
                <input value={draft.tls?.caFile ?? ''} onChange={(e) => patch({ tls: { ...draft.tls, caFile: e.target.value || undefined } })} placeholder="./certs/ca.pem" spellCheck={false} className="rounded border border-input bg-background px-2 py-1.5 font-mono text-xs" />
              </label>
            </div>

            <div className="flex flex-col gap-2 rounded border border-border/80 bg-muted/10 p-3">
              <span className="text-[11px] font-semibold uppercase tracking-wide text-foreground">Automatic Retries</span>
              <RetrySection retry={draft.retry} onChange={(retry) => patch({ retry })} />
            </div>
          </div>
        ) : (
          <div className="flex h-full flex-col gap-1">
            <div className="flex items-center gap-1">
              {bodyTypes.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => patch({ bodyType: t.id })}
                  className={tabClass(draft.bodyType === t.id)}
                >
                  {t.label}
                </button>
              ))}
            </div>
            {draft.bodyType === 'form-data' || draft.bodyType === 'urlencoded' ? (
              <div className="min-h-0 flex-1 overflow-y-auto rounded-md border border-border p-2">
                <KeyValueEditor
                  rows={draft.form}
                  onChange={(rows) => patch({ form: rows })}
                  keyPlaceholder="field"
                  valuePlaceholder="value"
                />
              </div>
            ) : draft.bodyType === 'binary' ? (
              <div
                className="flex min-h-0 flex-1 flex-col gap-2 rounded-md border border-border p-2"
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => {
                  e.preventDefault()
                  const file = e.dataTransfer.files[0]
                  if (file) patch({ body: `./fixtures/${file.name}` })
                }}
              >
                <div className="flex flex-col gap-1">
                  <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                    File path — relative to the request file (drag & drop or type)
                  </span>
                  <input
                    value={draft.body}
                    onChange={(e) => patch({ body: e.target.value })}
                    placeholder="./fixtures/payload.bin — or drop a file"
                    aria-label="Binary file path"
                    spellCheck={false}
                    className="min-w-0 rounded-md border border-input bg-background px-2 py-1.5 text-xs font-mono text-foreground placeholder:text-muted-foreground"
                  />
                </div>
                <p className="text-[11px] leading-relaxed text-muted-foreground">
                  Enter a Git-native path relative to the request file or drop a file to set the name. The core reads the bytes from disk at send time.
                </p>
              </div>
            ) : draft.bodyType === 'graphql' ? (
              <div className="flex min-h-0 flex-1 flex-col gap-2">
                <div className="flex min-h-0 flex-1 flex-col gap-1">
                  <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Query</span>
                  <CodeMirrorEditor
                    value={draft.graphqlQuery ?? draft.body}
                    language="graphql"
                    onChange={(graphqlQuery) => patch({ graphqlQuery })}
                    className="min-h-0 flex-1 overflow-hidden rounded-md border border-border"
                  />
                </div>
                <div className="flex min-h-0 flex-1 flex-col gap-1">
                  <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Variables (JSON)</span>
                  <CodeMirrorEditor
                    value={draft.graphqlVariables ?? '{\n  \n}'}
                    language="json"
                    onChange={(graphqlVariables) => patch({ graphqlVariables })}
                    className="min-h-0 flex-1 overflow-hidden rounded-md border border-border"
                  />
                </div>
              </div>
            ) : (
              <CodeMirrorEditor
                value={draft.body}
                language={bodyLanguage[draft.bodyType]}
                onChange={(body) => patch({ body })}
                className="min-h-0 flex-1 overflow-hidden rounded-md border border-border"
              />
            )}
          </div>
        )}
      </div>

      <TemplatePickerSheet
        open={templatePickerOpen}
        onOpenChange={setTemplatePickerOpen}
        onSelect={(inst) => {
          patch({
            method: inst.method,
            url: inst.path,
            body: inst.body ?? draft.body,
            headers: [
              ...draft.headers,
              ...Object.entries(inst.headers).map(([key, value]) => ({
                key,
                value,
                enabled: true,
              })),
            ],
          })
        }}
      />
    </div>
  )
}

const retryStrategies = [
  { value: 'exponential', label: 'Exponential' },
  { value: 'fixed', label: 'Fixed' },
] as const

/** settingsSummary renders the compact one-line state of the request's own
 * send overrides for the collapsed settings chip ("Defaults" when unset). */
function settingsSummary(draft: { timeout?: number; followRedirects?: boolean }): string {
  const parts: string[] = []
  if (draft.timeout) parts.push(`${draft.timeout}ms`)
  if (draft.followRedirects === false) parts.push('no redirects')
  else if (draft.followRedirects === true) parts.push('redirects')
  return parts.length > 0 ? parts.join(' · ') : 'Settings'
}

function retrySummary(retry: RequestRetry | undefined): string {
  if (!retry || !retry.count) return 'Off'
  const strategy = retry.strategy === 'fixed' ? 'fixed' : 'exponential'
  return `${retry.count} retries · ${strategy} · ${retry.delayMs ?? 1000}ms base`
}

/** RetrySection is the progressive-disclosure control for the request's
 * retry policy. Collapsed it costs one quiet row ("Retry — Off"); expanded
 * it exposes exactly the four policy fields the file format supports.
 * An empty count means no retries, so clearing the field disables the
 * policy entirely and nothing is written to the file. */
function RetrySection({
  retry,
  onChange,
}: {
  retry: RequestRetry | undefined
  onChange: (retry: RequestRetry | undefined) => void
}) {
  const [open, setOpen] = useState(false)
  const enabled = Boolean(retry?.count)
  const field =
    'w-20 rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground'

  const patch = (p: Partial<RequestRetry>) => {
    const next = { ...retry, ...p }
    if (!next.count) {
      onChange(undefined)
      return
    }
    onChange(next)
  }

  return (
    <div>
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 rounded-md py-0.5 text-left hover:bg-muted/50"
      >
        <ChevronRight
          className={`size-3 shrink-0 transition-transform ${open ? 'rotate-90' : ''}`}
          aria-hidden
        />
        <span className="text-xs font-medium text-foreground">Retry</span>
        <span className={`text-[11px] ${enabled ? 'text-muted-foreground' : 'text-muted-foreground/70'}`}>
          {retrySummary(retry)}
        </span>
      </button>
      {open && (
        <div className="mt-1 flex flex-wrap items-center gap-3 rounded-md border border-border p-2">
          <label className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
            Retries
            <input
              type="number"
              min={0}
              max={10}
              value={retry?.count ?? ''}
              onChange={(e) => patch({ count: Math.max(0, Math.floor(Number(e.target.value))) || 0 })}
              placeholder="0"
              aria-label="Retry count (0 disables retries)"
              className={field}
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
            Base delay (ms)
            <input
              type="number"
              min={0}
              value={retry?.delayMs ?? ''}
              onChange={(e) => patch({ delayMs: Math.max(0, Number(e.target.value) || 0) })}
              placeholder="1000"
              aria-label="Base delay between retries in milliseconds"
              className={field}
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
            Strategy
            <CompactSelect
              value={retry?.strategy === 'fixed' ? 'fixed' : 'exponential'}
              onChange={(strategy) => {
                if (strategy === 'fixed' || strategy === 'exponential') patch({ strategy })
              }}
              ariaLabel="Backoff strategy"
              options={retryStrategies.map((s) => ({ ...s }))}
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
            Max delay (ms)
            <input
              type="number"
              min={0}
              value={retry?.maxDelayMs ?? ''}
              onChange={(e) => patch({ maxDelayMs: Math.max(0, Number(e.target.value) || 0) })}
              placeholder="30000"
              aria-label="Maximum backoff delay in milliseconds"
              className={field}
            />
          </label>
          <p className="w-full text-[11px] leading-relaxed text-muted-foreground">
            Retries fire on network errors and 429/502/503/504 unless a custom status set is declared
            in the request file. The server's Retry-After header wins when present.
          </p>
        </div>
      )}
    </div>
  )
}

function VariablesView({
  variables,
  env,
  headerEnv,
}: {
  variables: ResolvedVariable[]
  env: string | null
  headerEnv: string | null
}) {  return (
    <div className="flex flex-col gap-2">
      <p className="text-xs text-muted-foreground">
        Effective variables, from highest to lowest precedence. Values are
        read-only and resolve at send time with the environment layered in.
      </p>
      <p className="text-xs text-muted-foreground">
        Environment: {env ?? headerEnv ?? 'none'} ({env ? 'pinned by file' : headerEnv ? 'from app header' : 'no environment selected'})
      </p>
      {variables.length === 0 ? (
        <p className="text-xs text-muted-foreground">No variables defined.</p>
      ) : (
        <div className="overflow-hidden rounded-md border border-border">
          {[...variables].reverse().map((v, i) => (
            <div
              key={`${v.scope}:${v.name}`}
              className={`flex items-start gap-2 px-2 py-1 text-xs ${
                i % 2 === 0 ? 'bg-muted/30' : ''
              }`}
            >
              <span className="w-20 shrink-0 rounded bg-muted px-1.5 py-0.5 text-center text-[10px] font-medium text-muted-foreground">
                {v.scope}
              </span>
              <span className="shrink-0 font-medium text-foreground">{v.name}</span>
              <code className="min-w-0 flex-1 truncate text-muted-foreground">
                {v.value || '(empty)'}
              </code>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}