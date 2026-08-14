import { useState } from 'react'
import { CodeMirrorEditor } from '../../editors'
import { Button } from '../../components/ui/button'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const

export function RequestEditor() {
  const [method, setMethod] = useState<string>('GET')
  const [url, setUrl] = useState('')
  const [body, setBody] = useState('{\n  \n}')

  const send = () => {
    // TODO(core): dispatch to the Go request engine over the Wails bridge.
    console.log({ method, url, body })
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
          placeholder="https://api.example.com/users"
          className="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground"
        />
        <Button size="sm" onClick={send}>
          Send
        </Button>
      </div>
      <div className="min-h-0 flex-1 p-2">
        <CodeMirrorEditor
          value={body}
          language="json"
          onChange={setBody}
          className="h-full overflow-hidden rounded-md border border-border"
        />
      </div>
    </div>
  )
}