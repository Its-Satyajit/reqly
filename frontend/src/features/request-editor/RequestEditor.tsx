import { useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { Button } from '../../components/ui/button'
import { KeyValueEditor } from '../../components/KeyValueEditor'
import { useRequestStore, useWorkspaceStore } from '../../stores'
import { sentRows } from '../../lib/request'
import { bodyTypes, type BodyType } from '../../lib/body'
import type { ResolvedVariable } from '../../lib/collections'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const

type Tab = 'params' | 'headers' | 'body' | 'variables'

const tabs: { id: Tab; label: string }[] = [
  { id: 'params', label: 'Params' },
  { id: 'headers', label: 'Headers' },
  { id: 'body', label: 'Body' },
  { id: 'variables', label: 'Variables' },
]

const tabClass = (active: boolean) =>
  `rounded-md px-2 py-1 text-xs font-medium transition-colors ${
    active
      ? 'bg-muted text-foreground'
      : 'text-muted-foreground hover:text-foreground'
  }`

const bodyLanguage: Record<BodyType, 'json' | 'xml' | 'text'> = {
  none: 'text',
  json: 'json',
  xml: 'xml',
  'form-data': 'text',
  urlencoded: 'text',
  raw: 'text',
}

export function RequestEditor() {
  const activeTabId = useWorkspaceStore((s) => s.activeTabId)
  const draft = useRequestStore((s) => (activeTabId ? s.drafts[activeTabId] : undefined))
  const meta = useRequestStore((s) => (activeTabId ? s.meta[activeTabId] : undefined))
  const loading = useRequestStore((s) => (activeTabId ? s.responses[activeTabId]?.loading : false))
  const updateDraft = useRequestStore((s) => s.updateDraft)
  const send = useRequestStore((s) => s.send)
  const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId)
  const environments = useWorkspaceStore((s) => s.environments)
  const [tab, setTab] = useState<Tab>('params')

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

  const handleSend = () => {
    void send(activeTabId, {
      method: draft.method,
      url: draft.url,
      params: sentRows(draft.params),
      headers: sentRows(draft.headers).map(({ key, value }) => ({ key, value })),
      bodyType: draft.bodyType,
      body: draft.body,
      form: sentRows(draft.form),
      env: meta?.env,
      vars: meta?.variables,
      auth: meta?.auth,
    })
  }

  const patch = (p: Partial<Parameters<typeof updateDraft>[1]>) => updateDraft(activeTabId, p)

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 p-2">
        <select
          value={draft.method}
          onChange={(e) => patch({ method: e.target.value })}
          className="w-28 rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground"
        >
          {methods.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
        <input
          value={draft.url}
          onChange={(e) => patch({ url: e.target.value })}
          placeholder="https://reqly-test-api.vercel.app/api/users?page=1 — mock API for testing"
          className="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground"
        />
        <span
          title={meta?.env ? 'Environment pinned by the request file' : 'Environment from the app header'}
          className="shrink-0 rounded-full border border-border bg-muted/50 px-2 py-1 text-[10px] font-medium text-muted-foreground"
        >
          {envPill ?? 'No environment'}
          {meta?.env ? ' • file' : ''}
        </span>
        <Button size="sm" onClick={handleSend} disabled={loading}>
          {loading ? 'Sending…' : 'Send'}
        </Button>
      </div>

      <div className="flex items-center gap-1 px-2">
        {tabs
          .filter((t) => showVariables || t.id !== 'variables')
          .map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setTab(t.id)}
              className={tabClass(tab === t.id)}
            >
              {t.label}
            </button>
          ))}
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
          <KeyValueEditor
            rows={draft.headers}
            onChange={(rows) => patch({ headers: rows })}
            keyPlaceholder="header"
            valuePlaceholder="value"
          />
        ) : tab === 'variables' ? (
          <VariablesView
            variables={meta?.variables ?? []}
            env={meta?.env ?? null}
            headerEnv={environments.find((e) => e.id === activeEnvironmentId)?.name ?? null}
          />
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
