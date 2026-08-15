import { CodeMirrorEditor } from '../../editors'
import { useRequestStore } from '../../stores'

function formatBody(body: string): string {
  try {
    return JSON.stringify(JSON.parse(body), null, 2)
  } catch {
    return body
  }
}

export function ResponseViewer() {
  const response = useRequestStore((s) => s.response)
  const loading = useRequestStore((s) => s.loading)
  const error = useRequestStore((s) => s.error)

  const content = loading
    ? '// Sending request…'
    : error
      ? `// Error: ${error}`
      : response
        ? `${response.proto ? `${response.proto} ` : ''}${response.statusCode} ${response.statusText} (${response.durationMs}ms)\n\n${Object.entries(response.headers)
            .map(([key, values]) => `${key}: ${values.join(', ')}`)
            .join('\n')}\n\n${formatBody(response.body)}`
        : '// Send a request to see the response'

  return (
    <div className="flex h-full flex-col">
      <p className="px-2 pb-1 pt-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
        Response
      </p>
      <div className="min-h-0 flex-1 p-2">
        <CodeMirrorEditor
          value={content}
          language="json"
          readOnly
          className="h-full overflow-hidden rounded-md border border-border"
        />
      </div>
    </div>
  )
}
