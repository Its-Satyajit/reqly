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
    <div role="dialog" aria-label="Command palette" className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 p-4 pt-[15vh] backdrop-blur-[1px]">
      <div className="w-full max-w-lg rounded border border-border bg-popover p-1.5 shadow-2xl">
        <div className="flex items-center border-b border-border/70 px-2.5 pb-1.5 pt-1">
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Type a command or search… (e.g. 'Theme', 'Env', 'Go to')"
            className="w-full border-0 bg-transparent font-mono text-xs text-foreground placeholder:text-muted-foreground/60 focus:outline-none"
            onKeyDown={(e) => {
              if (e.key === "Escape") setOpen(false);
              if (e.key === "Enter" && filtered[0]) run(filtered[0]);
            }}
          />
          <kbd className="shrink-0 rounded border border-border/80 bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
            ESC
          </kbd>
        </div>
        <div className="mt-1 max-h-72 overflow-y-auto p-1">
          {filtered.length === 0 ? (
            <div className="px-3 py-6 text-center">
              <p className="font-mono text-xs text-muted-foreground">
                No matching commands found.
              </p>
            </div>
          ) : query.trim() === "" && filtered.length > 0 ? (
            <div className="space-y-2">
              {Array.from(grouped.entries()).map(([group, items]) => (
                <div key={group}>
                  <p className="px-2 py-0.5 font-mono text-[9px] font-bold uppercase tracking-wider text-muted-foreground/70">
                    {group}
                  </p>
                  <ul className="flex flex-col gap-0.5">
                    {items.map((c) => (
                      <li key={c.id}>
                        <button
                          type="button"
                          onClick={() => run(c)}
                          className="flex w-full items-center justify-between rounded px-2 py-1 text-left transition-colors hover:bg-muted/70 focus:bg-muted/70"
                        >
                          <span className="font-mono text-xs text-foreground/90">{c.title}</span>
                          {c.hint && (
                            <span className="ml-2 shrink-0 rounded border border-border/60 bg-muted/50 px-1.5 py-0.2 font-mono text-[9px] text-muted-foreground">
                              {c.hint}
                            </span>
                          )}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          ) : (
            <ul className="flex flex-col gap-0.5">
              {filtered.map((c) => (
                <li key={c.id}>
                  <button
                    type="button"
                    onClick={() => run(c)}
                    className="flex w-full items-center justify-between rounded px-2 py-1 text-left transition-colors hover:bg-muted/70 focus:bg-muted/70"
                  >
                    <span className="font-mono text-xs text-foreground/90">{c.title}</span>
                    {c.hint && (
                      <span className="ml-2 shrink-0 rounded border border-border/60 bg-muted/50 px-1.5 py-0.2 font-mono text-[9px] text-muted-foreground">
                        {c.hint}
                      </span>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="flex items-center justify-between border-t border-border/70 px-2.5 pt-1.5 text-[10px] font-mono text-muted-foreground">
          <span><kbd className="text-foreground">↑↓</kbd> navigate</span>
          <span><kbd className="text-foreground">↵</kbd> run command</span>
        </div>
      </div>
    </div>
  );
}
