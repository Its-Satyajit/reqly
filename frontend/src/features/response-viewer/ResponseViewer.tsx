import { useState } from 'react'
import { CodeMirrorEditor } from '../../editors'

export interface ResponseViewerProps {
  value?: string
}

export function ResponseViewer({ value }: ResponseViewerProps) {
  const [draft, setDraft] = useState('// Send a request to see the response')

  const response = value ?? draft

  return (
    <div className="flex h-full flex-col">
      <p className="px-2 pb-1 pt-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
        Response
      </p>
      <div className="min-h-0 flex-1 p-2">
        <CodeMirrorEditor
          value={response}
          language="json"
          readOnly
          onChange={setDraft}
          className="h-full overflow-hidden rounded-md border border-border"
        />
      </div>
    </div>
  )
}