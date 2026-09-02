import { Plus, Trash2 } from "lucide-react"
import { Button } from "#components/ui/button"
import { cn } from "#lib/utils"

interface FieldListProps<T> {
  rows: T[]
  onChange: (rows: T[]) => void
  getId: (row: T, index: number) => string
  renderRow: (row: T, index: number) => React.ReactNode
  addLabel: string
  onAdd: () => T
  emptyMessage?: string
  className?: string
  rowClassName?: string
  addButtonVariant?: "default" | "outline" | "ghost" | "secondary"
  addButtonSize?: "xs" | "sm" | "default"
}

/**
 * FieldList — generic add/remove row list.
 * Use for env vars, mock route rows, broker credentials, etc.
 * For request-header/param rows prefer components/KeyValueEditor (K + V + enabled + file).
 */
export function FieldList<T>({
  rows,
  onChange,
  getId,
  renderRow,
  addLabel,
  onAdd,
  emptyMessage,
  className,
  rowClassName,
  addButtonVariant = "outline",
  addButtonSize = "sm",
}: FieldListProps<T>) {
  const removeAt = (index: number) => {
    onChange(rows.filter((_, j) => j !== index))
  }

  const add = () => {
    onChange([...rows, onAdd()])
  }

  return (
    <div className={cn("flex flex-col gap-2", className)}>
      {rows.length === 0 ? (
        emptyMessage ? (
          <p className="rounded-md border border-dashed border-border/60 bg-muted/10 px-3 py-4 text-center font-mono text-xs text-muted-foreground">
            {emptyMessage}
          </p>
        ) : null
      ) : (
        <div className="flex flex-col gap-1.5">
          {rows.map((row, i) => (
            <div
              key={getId(row, i)}
              className={cn("flex items-start gap-1.5 rounded-md border border-border/40 bg-card/40 p-2", rowClassName)}
            >
              <div className="min-w-0 flex-1">{renderRow(row, i)}</div>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="mt-0.5 text-muted-foreground hover:text-destructive"
                aria-label={`remove row ${i + 1}`}
                title="Remove row"
                onClick={() => removeAt(i)}
              >
                <Trash2 className="size-3.5" aria-hidden />
              </Button>
            </div>
          ))}
        </div>
      )}
      <Button type="button" variant={addButtonVariant} size={addButtonSize} className="self-start gap-1" onClick={add}>
        <Plus className="size-3.5" aria-hidden />
        {addLabel}
      </Button>
    </div>
  )
}
