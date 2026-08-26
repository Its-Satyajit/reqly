import { DYNAMIC_TAGS } from "../lib/tags"

export function TagPicker({ onInsert }: { onInsert: (tag: string) => void }) {
  return (
    <div className="flex items-center gap-1">
      <span className="text-2xs text-muted-foreground">Insert tag:</span>
      {DYNAMIC_TAGS.map((t) => (
        <button key={t} type="button" onClick={() => onInsert("{{$" + t + "}}")} className="rounded border border-border px-1.5 py-0.5 text-xs hover:bg-muted">
          {t}
        </button>
      ))}
    </div>
  )
}
