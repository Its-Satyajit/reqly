import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { Button } from "#components/ui/button";
import { SPEC_SECTIONS } from "#lib/specTree";
import { cn } from "#lib/utils";
import { useSpecEditorStore } from "#stores/useSpecEditorStore";

export function SpecEditorView() {
  const content = useSpecEditorStore((s) => s.content);
  const selectedId = useSpecEditorStore((s) => s.selectedId);
  const dirty = useSpecEditorStore((s) => s.dirty);
  const setContent = useSpecEditorStore((s) => s.setContent);
  const setSelected = useSpecEditorStore((s) => s.setSelected);
  const markSaved = useSpecEditorStore((s) => s.markSaved);

  return (
    <div className="flex h-full min-h-0">
      <aside className="w-[220px] shrink-0 border-r border-border bg-card/30 p-3">
        <p className="px-2 pb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Spec Tree</p>
        <ul className="space-y-0.5">
          {SPEC_SECTIONS.map((n) => (
            <li key={n.id}>
              <button
                type="button"
                onClick={() => setSelected(n.id)}
                className={cn("w-full rounded px-2 py-1 text-left text-xs", selectedId === n.id ? "bg-primary/12 font-medium text-primary" : "text-muted-foreground hover:bg-muted hover:text-foreground")}
              >
                {n.label}
              </button>
              {n.children?.map((c) => (
                <button
                  key={c.id}
                  type="button"
                  onClick={() => setSelected(c.id)}
                  className={cn("ml-4 mt-0.5 flex w-[calc(100%-1rem)] rounded px-2 py-1 text-left text-xs", selectedId === c.id ? "bg-primary/12 font-medium text-primary" : "text-muted-foreground hover:bg-muted hover:text-foreground")}
                >
                  {c.label}
                </button>
              ))}
            </li>
          ))}
        </ul>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex items-center justify-between border-b border-border px-3 py-2">
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-muted-foreground">openapi.yaml</span>
            {dirty && <span className="size-2 rounded-full bg-warning" aria-label="unsaved" />}
          </div>
          <Button size="sm" variant="secondary" onClick={markSaved} disabled={!dirty}>
            Save
          </Button>
        </div>
        <div className="min-h-0 flex-1">
          <CodeMirrorEditor value={content} language="yaml" onChange={setContent} className="h-full" />
        </div>
      </div>
    </div>
  );
}
