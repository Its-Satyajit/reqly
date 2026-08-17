import { useState } from 'react'
import { jsonText } from '../lib/response'

interface JsonTreeProps {
  data: unknown
  name?: string
  depth?: number
  filter?: string
}

function matches(value: unknown, needle: string): boolean {
  if (!needle) return true
  return jsonText(value).toLowerCase().includes(needle)
}

function JsonNode({ data, name, depth = 0, filter }: JsonTreeProps) {
  const [open, setOpen] = useState(depth === 0)
  const needle = (filter ?? '').trim().toLowerCase()

  if (data === null) {
    return (
      <Row name={name} label="null" className="text-muted-foreground" filter={filter} />
    )
  }

  const type = typeof data
  if (type !== 'object') {
    const label = type === 'string' ? JSON.stringify(data) : String(data)
    const className = type === 'string' ? 'text-foreground' : 'text-primary'
    return <Row name={name} label={label} className={className} filter={filter} />
  }

  const isArray = Array.isArray(data)
  const entries: [string, unknown][] = isArray
    ? (data as unknown[]).map((value, i) => [String(i), value])
    : Object.entries(data as Record<string, unknown>)
  const summary = `${isArray ? 'Array' : 'Object'}(${entries.length})`

  // Subtree filter: hide when the query is set and nothing under this node
  // contains it.
  const visible = filter ? matches(data, needle) : true
  if (!visible) return null

  return (
    <div>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 rounded px-1 py-0.5 text-xs hover:bg-muted"
      >
        <span className="w-3 text-muted-foreground">{open ? '▾' : '▸'}</span>
        {name ? (
          <span className="font-medium text-foreground">{name}:</span>
        ) : null}
        <span className="text-muted-foreground">{summary}</span>
      </button>
      {open ? (
        <div className="border-l border-border pl-3">
          {entries.map(([key, value]) => (
            <JsonNode
              key={key}
              name={key}
              data={value}
              depth={depth + 1}
              filter={filter}
            />
          ))}
        </div>
      ) : null}
    </div>
  )
}

function Row({
  name,
  label,
  className,
  filter,
}: {
  name?: string
  label: string
  className: string
  filter?: string
}) {
  const needle = (filter ?? '').trim().toLowerCase()
  if (needle && !label.toLowerCase().includes(needle) && !(name ?? '').toLowerCase().includes(needle)) {
    return null
  }
  return (
    <div className="flex items-center gap-1 rounded px-1 py-0.5 pl-4 text-xs">
      {name ? <span className="font-medium text-foreground">{name}:</span> : null}
      <span className={className}>{label}</span>
    </div>
  )
}

export function JsonTree({ data, filter }: { data: unknown; filter?: string }) {
  return <JsonNode data={data} depth={0} filter={filter} />
}
