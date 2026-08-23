import { useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { cn } from '../lib/utils'
import { jsonText } from '../lib/response'
import { isRecord, isString, type JsonObject, type JsonValue } from '../lib/typeGuards'

interface JsonTreeProps {
  data: JsonValue
  name?: string
  depth?: number
  filter?: string
}

function matches(value: JsonValue, needle: string): boolean {
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

  if (!isRecord(data) && !Array.isArray(data)) {
    const label = isString(data) ? JSON.stringify(data) : String(data)
    const className = isString(data) ? 'text-foreground' : 'text-primary'
    return <Row name={name} label={label} className={className} filter={filter} />
  }

  const isArray = Array.isArray(data)
  // SAFETY: Array.isArray narrows JsonValue to JsonValue[]; isRecord narrows to JsonObject with JsonValue values
  const entries: [string, JsonValue][] = isArray
    ? (data as JsonValue[]).map((value, i) => [String(i), value])
    : Object.entries(data as JsonObject)
  const summary = `${isArray ? 'Array' : 'Object'}(${entries.length})`

  // Subtree filter: hide when the query is set and nothing under this node
  // contains it.
  const visible = filter ? matches(data, needle) : true
  if (!visible) return null

  return (
    <div>
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 rounded px-1 py-0.5 text-xs hover:bg-muted"
      >
        <ChevronRight
          className={cn(
            'w-3 text-muted-foreground transition-transform',
            open && 'rotate-90',
          )}
          aria-hidden
        />
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

export function JsonTree({ data, filter }: { data: JsonValue; filter?: string }) {
  return <JsonNode data={data} depth={0} filter={filter} />
}
