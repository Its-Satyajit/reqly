import { Paperclip, Type, X } from 'lucide-react'
import type { KeyValueRow } from '../lib/request'
import { cn } from '../lib/utils'
import { inputClass } from '../lib/ui'
import { Button } from './ui/button'

interface KeyValueEditorProps {
  rows: KeyValueRow[]
  onChange: (rows: KeyValueRow[]) => void
  keyPlaceholder?: string
  valuePlaceholder?: string
}

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
        // Rows are anonymous value objects (often several blank rows at once)
        // with no stable identity — positional keys are the only correct choice.
        // react-doctor-disable-next-line react-doctor/no-array-index-as-key
        <div key={i} className="flex items-center gap-1">
          <input
            type="checkbox"
            checked={row.enabled}
            onChange={(e) => update(i, { enabled: e.target.checked })}
            title={row.enabled ? 'Enabled — click to disable' : 'Disabled — click to enable'}
            aria-label={`${keyPlaceholder} enabled`}
            className="size-3.5 shrink-0 accent-(--primary)"
          />
          <input
            value={row.key}
            onChange={(e) => update(i, { key: e.target.value })}
            placeholder={keyPlaceholder}
            aria-label={`${keyPlaceholder} name`}
            spellCheck={false}
            className={cn(inputClass, 'min-w-0 flex-1 font-mono', !row.enabled && 'opacity-50')}
          />
          {row.file !== undefined ? (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.file}
                onChange={(e) => update(i, { file: e.target.value })}
                placeholder="file path"
                aria-label={`${keyPlaceholder} file path`}
                spellCheck={false}
                className={cn(inputClass, 'min-w-0 flex-1 font-mono', !row.enabled && 'opacity-50')}
              />
              <input
                value={row.filename ?? ''}
                onChange={(e) => update(i, { filename: e.target.value || undefined })}
                placeholder="filename"
                aria-label={`${keyPlaceholder} filename`}
                spellCheck={false}
                className={cn(inputClass, 'w-24 font-mono', !row.enabled && 'opacity-50')}
              />
              <Button type="button" variant="ghost" size="icon-xs" title="Use text value" aria-label="Use text value" onClick={() => update(i, { file: undefined, filename: undefined })}><Type className="size-3" aria-hidden /></Button>
            </div>
          ) : (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.value}
                onChange={(e) => update(i, { value: e.target.value })}
                placeholder={valuePlaceholder}
                aria-label={`${keyPlaceholder} value`}
                spellCheck={false}
                className={cn(inputClass, 'min-w-0 flex-1 font-mono', !row.enabled && 'opacity-50')}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Use file"
                aria-label="Use file"
                onClick={() => update(i, { file: '', filename: undefined })}
              >
                <Paperclip className="size-3" aria-hidden />
              </Button>
            </div>
          )}
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            title="Remove row"
            aria-label={`Remove ${keyPlaceholder ?? 'key'} row`}
            onClick={() => onChange(rows.filter((_, j) => j !== i))}
          >
            <X className="size-3" aria-hidden />
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
