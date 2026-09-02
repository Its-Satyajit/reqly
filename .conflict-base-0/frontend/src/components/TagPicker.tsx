import { DYNAMIC_TAGS } from "../lib/tags"

export function TagPicker({ onInsert }: { onInsert: (tag: string) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1">
      <span className="hidden shrink-0 whitespace-nowrap font-mono text-[10px] uppercase tracking-wide text-muted-foreground sm:inline">
        Insert:
      </span>
      {DYNAMIC_TAGS.map((t) => (
        <button
          key={t}
          type="button"
          onClick={() => onInsert("{{$" + t + "}}")}
          className="shrink-0 whitespace-nowrap rounded border border-border bg-background px-1.5 py-0.5 font-mono text-[11px] leading-none hover:bg-muted hover:text-foreground"
        >
          {t}
        </button>
      ))}
    </div>
  )
}
