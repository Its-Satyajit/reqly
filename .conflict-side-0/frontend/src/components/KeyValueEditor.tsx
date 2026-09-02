import { Paperclip, Type, X } from 'lucide-react'
import type { KeyValueRow } from '../lib/request'
import { cn } from '../lib/utils'
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
    <div className="flex flex-col gap-1.5">
      {rows.length > 0 && (
        <div className="flex items-center gap-1 px-1 font-mono text-[10px] uppercase tracking-wider text-muted-foreground select-none">
          <span className="w-4 shrink-0" />
          <span className="min-w-0 flex-1">{keyPlaceholder}</span>
          <span className="min-w-0 flex-1">{valuePlaceholder}</span>
          <span className="w-6 shrink-0" />
        </div>
      )}
      {rows.map((row, i) => (
        <div key={`${row.key || 'row'}-${i}`} className="flex items-center gap-1.5 group">
          <input
            type="checkbox"
            checked={row.enabled}
            onChange={(e) => update(i, { enabled: e.target.checked })}
            title={row.enabled ? 'Enabled — click to disable' : 'Disabled — click to enable'}
            aria-label={`${keyPlaceholder} enabled`}
            className="size-3.5 shrink-0 rounded border-border accent-primary cursor-pointer"
          />
          <input
            value={row.key}
            onChange={(e) => update(i, { key: e.target.value })}
            placeholder={keyPlaceholder}
            aria-label={`${keyPlaceholder} name`}
            spellCheck={false}
            className={cn(
              "h-7 min-w-0 flex-1 rounded border border-input/80 bg-background px-2 font-mono text-xs text-foreground placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none",
              !row.enabled && 'opacity-50 line-through text-muted-foreground'
            )}
          />
          {row.file !== undefined ? (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.file}
                onChange={(e) => update(i, { file: e.target.value })}
                placeholder="file path"
                aria-label={`${keyPlaceholder} file path`}
                spellCheck={false}
                className={cn(
                  "h-7 min-w-0 flex-1 rounded border border-input/80 bg-background px-2 font-mono text-xs text-foreground placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none",
                  !row.enabled && 'opacity-50'
                )}
              />
              <input
                value={row.filename ?? ''}
                onChange={(e) => update(i, { filename: e.target.value || undefined })}
                placeholder="filename"
                aria-label={`${keyPlaceholder} filename`}
                spellCheck={false}
                className={cn(
                  "h-7 w-24 rounded border border-input/80 bg-background px-2 font-mono text-xs text-foreground placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none",
                  !row.enabled && 'opacity-50'
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="h-7 w-7 text-muted-foreground hover:text-foreground"
                title="Use text value"
                aria-label="Use text value"
                onClick={() => update(i, { file: undefined, filename: undefined })}
              >
                <Type className="size-3" aria-hidden />
              </Button>
            </div>
          ) : (
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <input
                value={row.value}
                onChange={(e) => update(i, { value: e.target.value })}
                placeholder={valuePlaceholder}
                aria-label={`${keyPlaceholder} value`}
                spellCheck={false}
                className={cn(
                  "h-7 min-w-0 flex-1 rounded border border-input/80 bg-background px-2 font-mono text-xs text-foreground placeholder:text-muted-foreground/50 focus:border-ring focus:outline-none",
                  !row.enabled && 'opacity-50 text-muted-foreground'
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="h-7 w-7 text-muted-foreground hover:text-foreground opacity-60 group-hover:opacity-100"
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
            className="h-7 w-7 text-muted-foreground hover:text-destructive opacity-40 group-hover:opacity-100 transition-opacity"
            title="Remove row"
            aria-label={`Remove ${keyPlaceholder ?? 'key'} row`}
            onClick={() => onChange(rows.filter((_, j) => j !== i))}
          >
            <X className="size-3.5" aria-hidden />
          </Button>
        </div>
      ))}
      <div className="pt-0.5">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...rows, { key: '', value: '', enabled: true }])}
          className="h-6 gap-1 px-2 font-mono text-[11px] text-muted-foreground hover:text-foreground border-dashed"
        >
          + Add {keyPlaceholder}
        </Button>
      </div>
    </div>
  )
}
