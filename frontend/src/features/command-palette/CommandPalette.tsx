import { useEffect, useMemo, useRef } from "react";
import { getFilteredResults, groupByHint, useCommandPaletteStore } from "#stores/useCommandPaletteStore";

export function CommandPalette() {
  const open = useCommandPaletteStore((s) => s.open);
  const query = useCommandPaletteStore((s) => s.query);
  const commands = useCommandPaletteStore((s) => s.commands);
  const providers = useCommandPaletteStore((s) => s.providers);
  const setQuery = useCommandPaletteStore((s) => s.setQuery);
  const setOpen = useCommandPaletteStore((s) => s.setOpen);
  const recordRun = useCommandPaletteStore((s) => s.recordRun);
  const filtered = useMemo(() => getFilteredResults(query, commands, providers), [query, commands, providers]);
  const grouped = useMemo(() => groupByHint(filtered), [filtered]);
  const inputRef = useRef<HTMLInputElement>(null);
  useEffect(() => { if (open) inputRef.current?.focus(); }, [open]);
  const run = (c: (typeof filtered)[number]) => {
    recordRun(c.id);
    c.run();
    setOpen(false);
  };
  if (!open) return null;
  return (
    <div role="dialog" aria-label="Command palette" className="fixed inset-0 z-50 flex items-start justify-center bg-black/20 p-8 pt-[20vh]">
      <div className="w-full max-w-lg rounded-lg border border-border bg-popover p-2 shadow-lg">
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search or jump… (try “Go to”, “Theme:”, “Environment:”)"
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          onKeyDown={(e) => {
            if (e.key === "Escape") setOpen(false);
            if (e.key === "Enter" && filtered[0]) run(filtered[0]);
          }}
        />
        <div className="mt-2 max-h-80 overflow-y-auto">
          {filtered.length === 0 ? (
            <p className="px-2 py-8 text-center text-sm text-muted-foreground">
              No matches for “{query}”.<br />
              <span className="text-xs">Try “Go to”, “Theme: Light”, or “Environment:” — or press Esc to close.</span>
            </p>
          ) : query.trim() === "" && filtered.length > 0 ? (
            // Grouped when no query — recent first already sorted, now visually grouped + hint
            <div className="space-y-3">
              {Array.from(grouped.entries()).map(([group, items]) => (
                <div key={group}>
                  <p className="px-2 py-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{group}</p>
                  <ul>
                    {items.map((c) => (
                      <li key={c.id}>
                        <button
                          onClick={() => run(c)}
                          className="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted"
                        >
                          <span>{c.title}</span>
                          {c.hint && <span className="ml-2 shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">{c.hint}</span>}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          ) : (
            <ul>
              {filtered.map((c) => (
                <li key={c.id}>
                  <button
                    onClick={() => run(c)}
                    className="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted"
                  >
                    <span>{c.title}</span>
                    {c.hint && <span className="ml-2 shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">{c.hint}</span>}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <p className="mt-2 border-t border-border px-2 pt-2 text-[11px] text-muted-foreground">
          <span className="font-mono">↑↓</span> navigate · <span className="font-mono">↵</span> run · <span className="font-mono">Esc</span> close
        </p>
      </div>
    </div>
  );
}
