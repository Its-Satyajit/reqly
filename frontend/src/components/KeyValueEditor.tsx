import type { KeyValueRow } from '../lib/request'
import { Button } from './ui/button'

interface KeyValueEditorProps {
  rows: KeyValueRow[]
  onChange: (rows: KeyValueRow[]) => void
  keyPlaceholder?: string
  valuePlaceholder?: string
}

const inputClass =
  'rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus-visible:border-ring'

export function KeyValueEditor({
  rows,
  onChange,
  keyPlaceholder = 'key',
  valuePlaceholder = 'value',
}: KeyValueEditorProps) {
  const update = (index: number, patch: Partial<KeyValueRow>) => {
    onChange(rows.map((row, i) => (i === index ? { ...row, ...patch } : row)))
  }

  return (
    <div className="flex flex-col gap-1">
      {rows.map((row, i) => (
        <div key={i} className="flex items-center gap-1">
          <input
            type="checkbox"
            checked={row.enabled}
            onChange={(e) => update(i, { enabled: e.target.checked })}
            title={row.enabled ? 'Enabled — click to disable' : 'Disabled — click to enable'}
            className="size-3.5 shrink-0 accent-(--primary)"
          />
          <input
            value={row.key}
            onChange={(e) => update(i, { key: e.target.value })}
            placeholder={keyPlaceholder}
            spellCheck={false}
            className={`${inputClass} min-w-0 flex-1 font-mono ${row.enabled ? '' : 'opacity-50'}`}
          />
          {row.file !== undefined ? (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.file}
                onChange={(e) => update(i, { file: e.target.value })}
                placeholder="file path"
                spellCheck={false}
                className={`${inputClass} min-w-0 flex-1 font-mono ${row.enabled ? '' : 'opacity-50'}`}
              />
              <input
                value={row.filename ?? ''}
                onChange={(e) => update(i, { filename: e.target.value || undefined })}
                placeholder="filename"
                spellCheck={false}
                className={`${inputClass} w-24 font-mono ${row.enabled ? '' : 'opacity-50'}`}
              />
              <Button type="button" variant="ghost" size="icon-xs" title="Use text value" onClick={() => update(i, { file: undefined, filename: undefined })}>T</Button>
            </div>
          ) : (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.value}
                onChange={(e) => update(i, { value: e.target.value })}
                placeholder={valuePlaceholder}
                spellCheck={false}
                className={`${inputClass} min-w-0 flex-1 font-mono ${row.enabled ? '' : 'opacity-50'}`}
              />
              <Button type="button" variant="ghost" size="icon-xs" title="Use file" onClick={() => update(i, { file: '', filename: undefined })}>F</Button>
            </div>
          )}
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            title="Remove row"
            onClick={() => onChange(rows.filter((_, j) => j !== i))}
          >
            ×
          </Button>
        </div>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="xs"
          onClick={() => onChange([...rows, { key: '', value: '', enabled: true }])}
        >
          Add row
        </Button>
      </div>
    </div>
  )
}
