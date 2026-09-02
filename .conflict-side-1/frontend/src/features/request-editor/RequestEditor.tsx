import { useEffect, useState } from 'react'
import { CodeMirrorEditor } from '../../editors/CodeMirrorEditor'
import { Button } from '../../components/ui/button'
import { CompactSelect } from '../../components/CompactSelect'
import { KeyValueEditor } from '../../components/KeyValueEditor'
import { Loader2, SlidersHorizontal, Sparkles, MoreHorizontal } from 'lucide-react'
import { RequestSettingsDialog } from './RequestSettingsDialog'
import { AuthEditor } from '../auth-editor/AuthEditor'
import { authWarnings } from '../../lib/authSchemes'
import { useRequestStore, tabIsDirty } from '../../stores/useRequestStore'
import { useWorkspaceStore, effectiveUrlFor } from '../../stores/useWorkspaceStore'
import { sentRows } from '../../lib/request'
import { bodyTypes, type BodyType } from '../../lib/body'
import type { KeyValueRow, RequestAuth, RequestRetry } from '../../lib/request'
import type { ResolvedVariable } from '../../lib/collections'
import { TagPicker } from '../../components/TagPicker'
import { tagWarnings } from '../../lib/tags'
import { generateCode } from '../../lib/codegen'
import { copyText } from '../../lib/response'
import { notifyError } from '../../lib/notify'
import { handleTabArrowKeys } from '../../lib/ui'
import { cn } from '../../lib/utils'
import { TemplatePickerSheet } from './TemplatePickerSheet'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '../../components/ui/dropdown-menu'

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
      <div className="flex h-full flex-col items-center justify-center gap-2 p-6">
        <p className="text-sm font-medium text-foreground">No request open</p>
        <p className="max-w-sm text-center text-xs leading-relaxed text-muted-foreground">
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
    <div className="flex h-full flex-col bg-card">
      {meta?.changedOnDisk && (
        <div className="flex items-center justify-between gap-2 border-b border-status-warn/30 bg-status-warn/10 px-3 py-2">
          <p className="text-xs leading-snug text-status-warn">
            This request changed on disk since you opened it. Overwrite or reload.
          </p>
          <div className="flex shrink-0 gap-1.5">
            <Button size="xs" variant="outline" onClick={() => void reloadRequest(activeTabId)}>
              Reload
            </Button>
            <Button size="xs" onClick={() => void overwriteRequest(activeTabId)}>
              Overwrite
            </Button>
          </div>
        </div>
      )}
      {requestPath && dirty && warnings.length > 0 && (
        <div className="flex flex-col gap-0.5 border-b border-status-warn/30 bg-status-warn/10 px-3 py-2">
          {warnings.map((w) => (
            <p key={w} className="text-xs leading-snug text-status-warn">
              {w}
            </p>
          ))}
        </div>
      )}

      {/* Transmission rail — responsive: URL beam full-width on mobile, actions wrap */}
      <div className="border-b border-border bg-card">
        <div className="flex flex-col gap-2 px-3 py-3 sm:flex-row sm:items-center">
          <div className="flex min-w-0 w-full flex-1 items-center overflow-hidden rounded-md border border-input bg-background focus-within:border-ring focus-within:ring-1 focus-within:ring-ring/20">
            <CompactSelect
              value={draft.method}
              onChange={(method) => patch({ method })}
              ariaLabel="HTTP method"
              className="w-[88px] shrink-0 rounded-r-none border-0 border-r border-input bg-muted/50 font-mono text-xs font-bold tracking-wide"
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
              placeholder="https://api.example.com/v1/resource  or  {{baseUrl}}/users"
              spellCheck={false}
              aria-label="Request URL"
              className="min-w-0 flex-1 border-0 bg-transparent px-3 py-2 font-mono text-[13px] leading-none text-foreground placeholder:text-muted-foreground/50 focus:outline-none"
            />
          </div>

          <div className="flex w-full shrink-0 items-center justify-between gap-1.5 sm:w-auto sm:justify-end">
            <div className="flex items-center gap-1.5">
              {envPill && (
                <span
                  title={meta?.env ? 'Environment pinned by file' : 'Environment from header'}
                  className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/40 px-2 py-1 font-mono text-[11px] leading-none text-muted-foreground"
                >
                  <span className="size-1.5 rounded-full bg-status-ok" aria-hidden />
                  {envPill}
                </span>
              )}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setTemplatePickerOpen(true)}
                className="hidden h-8 gap-1.5 px-2.5 text-muted-foreground hover:text-foreground sm:inline-flex"
                title="Insert template"
              >
                <Sparkles className="size-3.5 text-primary" aria-hidden />
                <span className="hidden lg:inline text-xs">Templates</span>
              </Button>
              {/* mobile templates as icon only to avoid overlap */}
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => setTemplatePickerOpen(true)}
                className="h-8 w-8 text-muted-foreground hover:text-foreground sm:hidden"
                title="Insert template"
                aria-label="Insert template"
              >
                <Sparkles className="size-3.5 text-primary" aria-hidden />
              </Button>
            </div>

            <div className="flex shrink-0 items-center gap-1.5">
              {requestPath && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => void saveRequest(activeTabId)}
                  disabled={!canSave}
                  className="h-8 px-3 font-mono text-xs"
                >
                  {dirty ? 'Save' : 'Saved'}
                </Button>
              )}
              {loading ? (
                <Button size="sm" variant="destructive" onClick={() => void cancel(activeTabId)} className="h-8 gap-1.5 px-3.5 font-semibold">
                  <Loader2 className="size-3.5 animate-spin" aria-hidden />
                  Cancel
                </Button>
              ) : (
                <Button size="sm" onClick={handleSend} className="h-8 gap-1 px-4 font-semibold tracking-wide">
                  Send
                  <kbd className="hidden rounded bg-primary-foreground/15 px-1 py-0.5 font-mono text-[10px] font-normal tracking-normal lg:inline">⌘⏎</kbd>
                </Button>
              )}
            </div>
          </div>
        </div>

        {/* resolved line — quiet, mono, only when file-backed */}
        {requestPath && (
          <div className="flex items-center gap-2 border-t border-border/50 bg-muted/20 px-3 py-1.5">
            <span className="shrink-0 font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Resolved</span>
            <code className="min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground">
              {effectiveUrlFor(draft.url, meta?.baseUrl ?? '')}
            </code>
            <span className="hidden shrink-0 items-center gap-1.5 sm:flex">
              <CompactSelect
                value={codeLang}
                onChange={(v) => {
                  if (v === "curl" || v === "js" || v === "python" || v === "go") setCodeLang(v)
                }}
                ariaLabel="Snippet language"
                className="h-6 w-16 text-[11px] font-mono"
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
                className="h-6 px-2 font-mono text-[11px] text-muted-foreground hover:text-foreground"
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
            </span>
          </div>
        )}

        {/* transmission tracer — subtle progress when sending */}
        <div className="h-px w-full bg-border">
          <div
            className={cn(
              "h-px bg-primary transition-all duration-300",
              loading ? "w-full opacity-100" : "w-0 opacity-0"
            )}
            aria-hidden
          />
        </div>
      </div>

      {settingsOpen && (
        <RequestSettingsDialog
          settings={{ timeout: draft.timeout, followRedirects: draft.followRedirects }}
          onApply={(s) => patch(s)}
          onClose={() => setSettingsOpen(false)}
        />
      )}

      {/* Section rail — two distinct rows: Row 1 tabs, Row 2 insert tags + settings */}
      <div className="flex shrink-0 flex-col border-b border-border bg-card">
        {/* Row 1: Params · Headers · Body · Auth — always its own row, scrollable */}
        <div
          className="flex min-w-0 items-center gap-0.5 overflow-x-auto whitespace-nowrap border-b border-border/40 px-2 py-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
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
            const active = tab === t.id
            return (
              <button
                key={t.id}
                type="button"
                role="tab"
                aria-selected={active}
                tabIndex={active ? 0 : -1}
                onClick={() => setTab(t.id)}
                className={cn(
                  "relative shrink-0 rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors",
                  active
                    ? "bg-muted text-foreground"
                    : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                )}
              >
                <span className="inline-flex items-center gap-1.5">
                  {t.label}
                  {count > 0 && <span className="size-1.5 rounded-full bg-primary" aria-hidden />}
                </span>
              </button>
            )
          })}
          <div className="ml-1 flex shrink-0 items-center gap-1 border-l border-border pl-1">
            <DropdownMenu>
              <DropdownMenuTrigger
                aria-label="More sections"
                className={cn(
                  "inline-flex items-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium transition-colors",
                  overflowTabs.some((o) => o.id === tab)
                    ? "bg-muted text-foreground"
                    : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                )}
              >
                <MoreHorizontal className="size-3.5" aria-hidden />
                More
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="min-w-36">
                {overflowTabs.map((t) => {
                  if (t.id === 'variables' && !showVariables) return null
                  return (
                    <DropdownMenuItem
                      key={t.id}
                      onSelect={() => setTab(t.id)}
                      className={cn(tab === t.id && "bg-accent text-accent-foreground")}
                    >
                      {t.label}
                    </DropdownMenuItem>
                  )
                })}
              </DropdownMenuContent>
            </DropdownMenu>
            {overflowTabs.some((o) => o.id === tab) && (
              <span className="inline-flex items-center gap-1 text-xs font-medium text-foreground">
                · {overflowTabs.find((o) => o.id === tab)?.label}
              </span>
            )}
          </div>
        </div>

        {/* Row 2: Insert tags + settings — always separate, never overlaps tabs */}
        <div className="flex flex-wrap items-center justify-between gap-2 bg-muted/20 px-2 py-1.5">
          <div className="flex min-w-0 flex-1 flex-wrap items-center gap-1.5">
            <TagPicker onInsert={(tag) => patch({ url: draft.url + tag })} />
          </div>
          <button
            type="button"
            onClick={() => setSettingsOpen(true)}
            className="inline-flex shrink-0 items-center gap-1.5 rounded-md border border-border bg-background px-2 py-1 font-mono text-[11px] leading-none text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <SlidersHorizontal className="size-3" aria-hidden />
            {settingsSummary(draft)}
          </button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto bg-background p-3">
        {tab === 'params' ? (
          <KeyValueEditor
            rows={draft.params}
            onChange={(rows) => patch({ params: rows })}
            keyPlaceholder="key"
            valuePlaceholder="value"
          />
        ) : tab === 'headers' ? (
          <div className="flex flex-col gap-3">
            <KeyValueEditor
              rows={draft.headers}
              onChange={(rows) => patch({ headers: rows })}
              keyPlaceholder="header"
              valuePlaceholder="value"
            />
            {(meta?.inheritedHeaders.length ?? 0) > 0 && (
              <div className="rounded-md border border-border bg-muted/20">
                <div className="border-b border-border px-2.5 py-1.5">
                  <p className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                    Inherited — workspace / collection / folder
                  </p>
                </div>
                <div className="divide-y divide-border/50">
                  {meta!.inheritedHeaders.map((h) => (
                    <div key={`${h.key}:${h.value}`} className="flex items-center gap-3 px-2.5 py-1.5 text-xs">
                      <span className="shrink-0 font-medium text-foreground">{h.key}</span>
                      <code className="min-w-0 flex-1 truncate font-mono text-muted-foreground">{h.value}</code>
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
          <div className="flex h-full min-h-[160px] items-center justify-center rounded-md border border-dashed border-border bg-muted/10 p-6">
            <p className="text-center text-xs leading-relaxed text-muted-foreground">
              {tab === 'pre-request' && 'Pre-request scripts — coming soon.'}
              {tab === 'tests' && 'Tests — coming soon.'}
              {tab === 'docs' && 'Docs — coming soon.'}
            </p>
          </div>
        ) : tab === 'settings' ? (
          <div className="flex max-w-xl flex-col gap-6">
            <section className="flex flex-col gap-3">
              <div>
                <h3 className="text-sm font-semibold">Proxy & TLS</h3>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                  Override global proxy and certificate settings for this request only.
                </p>
              </div>
              <div className="flex flex-col gap-3">
                <label className="flex flex-col gap-1.5">
                  <span className="text-xs font-medium">Proxy URL</span>
                  <input
                    value={draft.proxy ?? ''}
                    onChange={(e) => patch({ proxy: e.target.value || undefined })}
                    placeholder="http://proxy:8080"
                    spellCheck={false}
                    className="rounded-md border border-input bg-background px-2.5 py-2 font-mono text-xs focus:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
                  />
                </label>
                <label className="flex flex-col gap-1.5">
                  <span className="text-xs font-medium">CA file <span className="font-normal text-muted-foreground">(workspace-relative PEM)</span></span>
                  <input
                    value={draft.tls?.caFile ?? ''}
                    onChange={(e) => patch({ tls: { ...draft.tls, caFile: e.target.value || undefined } })}
                    placeholder="./certs/ca.pem"
                    spellCheck={false}
                    className="rounded-md border border-input bg-background px-2.5 py-2 font-mono text-xs focus:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
                  />
                </label>
                <label className="flex cursor-pointer items-center gap-2 text-xs select-none">
                  <input
                    type="checkbox"
                    checked={Boolean(draft.tls?.insecureSkipVerify)}
                    onChange={(e) => patch({ tls: { ...draft.tls, insecureSkipVerify: e.target.checked } })}
                    className="size-3.5 rounded border-input accent-primary"
                  />
                  <span>Skip TLS verification</span>
                  <span className="text-muted-foreground">(insecure)</span>
                </label>
              </div>
            </section>
            <div className="h-px bg-border" />
            <section className="flex flex-col gap-3">
              <div>
                <h3 className="text-sm font-semibold">Retry policy</h3>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                  Retry on network errors and 429 / 502–504 unless a custom status set is declared in the file.
                </p>
              </div>
              <RetrySection retry={draft.retry} onChange={(retry) => patch({ retry })} />
            </section>
          </div>
        ) : (
          <div className="flex h-full min-h-0 flex-col gap-3">
            <div className="flex gap-1 overflow-x-auto whitespace-nowrap pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
              {bodyTypes.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => patch({ bodyType: t.id })}
                  className={cn(
                    "shrink-0 rounded-md border px-2.5 py-1 text-xs font-medium transition-colors",
                    draft.bodyType === t.id
                      ? "border-border bg-muted text-foreground"
                      : "border-transparent text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                  )}
                >
                  {t.label}
                </button>
              ))}
            </div>
            {draft.bodyType === 'form-data' || draft.bodyType === 'urlencoded' ? (
              <div className="min-h-0 flex-1 overflow-y-auto rounded-md border border-border bg-card p-2">
                <KeyValueEditor
                  rows={draft.form}
                  onChange={(rows) => patch({ form: rows })}
                  keyPlaceholder="field"
                  valuePlaceholder="value"
                />
              </div>
            ) : draft.bodyType === 'binary' ? (
              <div
                className="flex flex-col gap-2 rounded-md border border-border bg-card p-3"
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => {
                  e.preventDefault()
                  const file = e.dataTransfer.files[0]
                  if (file) patch({ body: `./fixtures/${file.name}` })
                }}
              >
                <span className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  File path — relative to request file
                </span>
                <input
                  value={draft.body}
                  onChange={(e) => patch({ body: e.target.value })}
                  placeholder="./fixtures/payload.bin — or drop a file"
                  aria-label="Binary file path"
                  spellCheck={false}
                  className="rounded-md border border-input bg-background px-2.5 py-2 font-mono text-xs focus:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
                />
                <p className="text-xs leading-relaxed text-muted-foreground">
                  Git-native path. Core reads bytes from disk at send time.
                </p>
              </div>
            ) : draft.bodyType === 'graphql' ? (
              <div className="flex min-h-0 flex-1 flex-col gap-3">
                <div className="flex min-h-[160px] flex-1 flex-col gap-1.5">
                  <span className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Query</span>
                  <CodeMirrorEditor
                    value={draft.graphqlQuery ?? draft.body}
                    language="graphql"
                    onChange={(graphqlQuery) => patch({ graphqlQuery })}
                    className="min-h-0 flex-1 overflow-hidden rounded-md border border-border"
                  />
                </div>
                <div className="flex min-h-[140px] flex-1 flex-col gap-1.5">
                  <span className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Variables (JSON)</span>
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

function RetrySection({
  retry,
  onChange,
}: {
  retry: RequestRetry | undefined
  onChange: (retry: RequestRetry | undefined) => void
}) {
  const enabled = Boolean(retry?.count)
  const patch = (p: Partial<RequestRetry>) => {
    const next = { ...retry, ...p }
    if (!next.count) {
      onChange(undefined)
      return
    }
    onChange(next)
  }
  const inputCls =
    'w-full rounded-md border border-input bg-background px-2.5 py-2 text-xs focus:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring'
  return (
    <div className="flex flex-col gap-3">
      <label className="flex cursor-pointer items-center gap-2 text-xs select-none">
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => {
            if (e.target.checked) {
              onChange({ count: 3, strategy: 'exponential', delayMs: 1000, maxDelayMs: 30000 })
            } else {
              onChange(undefined)
            }
          }}
          className="size-3.5 rounded border-input accent-primary"
        />
        <span>Enable retries</span>
        {enabled && <span className="text-muted-foreground">— {retrySummary(retry)}</span>}
      </label>
      {enabled && (
        <div className="grid grid-cols-2 gap-3">
          <label className="flex flex-col gap-1.5">
            <span className="text-xs font-medium">Retry count</span>
            <input
              type="number"
              min={1}
              max={10}
              value={retry?.count ?? ''}
              onChange={(e) => patch({ count: Math.max(0, Math.floor(Number(e.target.value))) || 0 })}
              placeholder="3"
              aria-label="Retry count"
              className={inputCls}
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-xs font-medium">Base delay (ms)</span>
            <input
              type="number"
              min={0}
              value={retry?.delayMs ?? ''}
              onChange={(e) => patch({ delayMs: Math.max(0, Number(e.target.value) || 0) })}
              placeholder="1000"
              aria-label="Base delay between retries in milliseconds"
              className={inputCls}
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-xs font-medium">Max delay (ms)</span>
            <input
              type="number"
              min={0}
              value={retry?.maxDelayMs ?? ''}
              onChange={(e) => patch({ maxDelayMs: Math.max(0, Number(e.target.value) || 0) })}
              placeholder="30000"
              aria-label="Maximum backoff delay in milliseconds"
              className={inputCls}
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-xs font-medium">Backoff strategy</span>
            <CompactSelect
              value={retry?.strategy === 'fixed' ? 'fixed' : 'exponential'}
              onChange={(strategy) => {
                if (strategy === 'fixed' || strategy === 'exponential') patch({ strategy })
              }}
              ariaLabel="Backoff strategy"
              options={retryStrategies.map((s) => ({ ...s }))}
            />
          </label>
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
}) {
  return (
    <div className="flex flex-col gap-2">
      <p className="text-xs leading-relaxed text-muted-foreground">
        Effective variables, highest → lowest precedence. Values are read-only, resolved at send time.
      </p>
      <p className="font-mono text-xs text-muted-foreground">
        Env: {env ?? headerEnv ?? 'none'} ({env ? 'pinned by file' : headerEnv ? 'from header' : 'no env'})
      </p>
      {variables.length === 0 ? (
        <p className="text-xs text-muted-foreground">No variables defined.</p>
      ) : (
        <div className="overflow-hidden rounded-md border border-border">
          {[...variables].reverse().map((v, i) => (
            <div
              key={`${v.scope}:${v.name}`}
              className={cn("flex items-start gap-2 px-2.5 py-1.5 text-xs", i % 2 === 0 ? 'bg-muted/30' : 'bg-card')}
            >
              <span className="w-20 shrink-0 rounded bg-muted px-1.5 py-0.5 text-center font-mono text-[10px] font-medium text-muted-foreground">
                {v.scope}
              </span>
              <span className="shrink-0 font-medium">{v.name}</span>
              <code className="min-w-0 flex-1 truncate font-mono text-muted-foreground">
                {v.value || '(empty)'}
              </code>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
