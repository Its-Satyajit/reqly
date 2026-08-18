import { useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { Button } from '../../components/ui/button'
import { KeyValueEditor } from '../../components/KeyValueEditor'
import { useRequestStore } from '../../stores'
import { sentRows, type KeyValueRow } from '../../lib/request'
import { bodyTypes, type BodyType } from '../../lib/body'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const

type Tab = 'params' | 'headers' | 'body'

const tabs: { id: Tab; label: string }[] = [
  { id: 'params', label: 'Params' },
  { id: 'headers', label: 'Headers' },
  { id: 'body', label: 'Body' },
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
  const [method, setMethod] = useState<string>('GET')
  const [url, setUrl] = useState('')
  const [bodyType, setBodyType] = useState<BodyType>('none')
  const [body, setBody] = useState('{\n  \n}')
  const [form, setForm] = useState<KeyValueRow[]>([])
  const [tab, setTab] = useState<Tab>('params')
  const [params, setParams] = useState<KeyValueRow[]>([])
  const [headers, setHeaders] = useState<KeyValueRow[]>([])
  const send = useRequestStore((s) => s.send)
  const loading = useRequestStore((s) => s.loading)

  const handleSend = () => {
    void send({
      method,
      url,
      params: sentRows(params),
      headers: sentRows(headers).map(({ key, value }) => ({ key, value })),
      bodyType,
      body,
      form: sentRows(form),
    })
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 p-2">
        <select
          value={method}
          onChange={(e) => setMethod(e.target.value)}
          className="w-28 rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground"
        >
          {methods.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://reqly-test-api.vercel.app/api/users?page=1 — mock API for testing"
          className="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground"
        />
        <Button size="sm" onClick={handleSend} disabled={loading}>
          {loading ? 'Sending…' : 'Send'}
        </Button>
      </div>

      <div className="flex items-center gap-1 px-2">
        {tabs.map((t) => (
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
            rows={params}
            onChange={setParams}
            keyPlaceholder="param"
            valuePlaceholder="value"
          />
        ) : tab === 'headers' ? (
          <KeyValueEditor
            rows={headers}
            onChange={setHeaders}
            keyPlaceholder="header"
            valuePlaceholder="value"
          />
        ) : (
          <div className="flex h-full flex-col gap-1">
            <div className="flex items-center gap-1">
              {bodyTypes.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => setBodyType(t.id)}
                  className={tabClass(bodyType === t.id)}
                >
                  {t.label}
                </button>
              ))}
            </div>
            {bodyType === 'form-data' || bodyType === 'urlencoded' ? (
              <div className="min-h-0 flex-1 overflow-y-auto rounded-md border border-border p-2">
                <KeyValueEditor
                  rows={form}
                  onChange={setForm}
                  keyPlaceholder="field"
                  valuePlaceholder="value"
                />
              </div>
            ) : (
              <CodeMirrorEditor
                value={body}
                language={bodyLanguage[bodyType]}
                onChange={setBody}
                className="min-h-0 flex-1 overflow-hidden rounded-md border border-border"
              />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
