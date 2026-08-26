import { useEffect, useRef } from "react";
import { useCommandPaletteStore } from "#stores/useCommandPaletteStore";

export function CommandPalette() {
  const open = useCommandPaletteStore((s) => s.open);
  const query = useCommandPaletteStore((s) => s.query);
  const setQuery = useCommandPaletteStore((s) => s.setQuery);
  const setOpen = useCommandPaletteStore((s) => s.setOpen);
  const filtered = useCommandPaletteStore((s) => s.filtered());
  const inputRef = useRef<HTMLInputElement>(null);
  useEffect(() => { if (open) inputRef.current?.focus(); }, [open]);
  if (!open) return null;
  return (
    <div role="dialog" aria-label="Command palette" className="fixed inset-0 bg-black/20 p-8">
      <div className="mx-auto max-w-lg bg-popover rounded-lg border p-2">
        <input ref={inputRef} value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Type a command…" className="w-full px-2 py-1" onKeyDown={(e) => { if (e.key === "Escape") setOpen(false); if (e.key === "Enter" && filtered[0]) { filtered[0].run(); setOpen(false); } }} />
        <ul className="mt-2">
          {filtered.length === 0 ? <li className="text-xs text-muted-foreground px-2 py-1">No commands match “{query}”.</li> : filtered.map((c) => (
            <li key={c.id}><button onClick={() => { c.run(); setOpen(false); }} className="w-full text-left px-2 py-1 text-sm hover:bg-muted flex justify-between"><span>{c.title}</span>{c.hint && <span className="text-xs text-muted-foreground">{c.hint}</span>}</button></li>
          ))}
        </ul>
      </div>
    </div>
  );
}
