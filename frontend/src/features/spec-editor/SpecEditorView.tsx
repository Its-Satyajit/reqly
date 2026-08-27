import { useState } from "react";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import { Button } from "#components/ui/button";
import { SPEC_SECTIONS } from "#lib/specTree";
import { nodesForSpec, edgesForNodes } from "#lib/schemaGraph";
import { cn } from "#lib/utils";
import { useSpecEditorStore } from "#stores/useSpecEditorStore";

function SchemaViz({ selectedId }: { selectedId: string }) {
  const nodes = nodesForSpec(selectedId);
  const edges = edgesForNodes(nodes);
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));
  return (
    <div className="h-full overflow-auto bg-card p-6">
      <p className="mb-4 font-mono text-xs text-muted-foreground">Schema Visualization — {selectedId}</p>
      <svg viewBox="0 0 400 300" className="mx-auto h-[280px] w-full max-w-[480px] rounded-lg border border-border bg-background">
        {edges.map((e) => {
          const a = nodeMap.get(e.from)!;
          const b = nodeMap.get(e.to)!;
          return <line key={`${e.from}->${e.to}`} x1={a.x} y1={a.y + 18} x2={b.x} y2={b.y} stroke="currentColor" className="text-border" strokeWidth={1.2} strokeDasharray="6 4" markerEnd="url(#arrow)" />;
        })}
        <defs>
          <marker id="arrow" viewBox="0 0 8 8" refX={6} refY={4} markerWidth={8} markerHeight={8} orient="auto-start-reverse">
            <path d="M 0 0 L 8 4 L 0 8 z" className="fill-muted-foreground" />
          </marker>
        </defs>
        {nodes.map((n) => (
          <g key={n.id} transform={`translate(${n.x - 44}, ${n.y})`}>
            <rect width={88} height={36} rx={6} className={n.id === "User" ? "fill-primary/12 stroke-primary" : "fill-card stroke-border"} strokeWidth={1} />
            <text x={44} y={22} textAnchor="middle" className={n.id === "User" ? "fill-primary font-mono text-[11px] font-medium" : "fill-foreground font-mono text-[11px]"}>
              {n.label}
            </text>
          </g>
        ))}
      </svg>
      {nodes.length === 1 && <p className="mt-3 text-center text-xs text-muted-foreground">Select Components → Schemas to see the full graph.</p>}
    </div>
  );
}

export function SpecEditorView() {
  const content = useSpecEditorStore((s) => s.content);
  const selectedId = useSpecEditorStore((s) => s.selectedId);
  const dirty = useSpecEditorStore((s) => s.dirty);
  const setContent = useSpecEditorStore((s) => s.setContent);
  const setSelected = useSpecEditorStore((s) => s.setSelected);
  const markSaved = useSpecEditorStore((s) => s.markSaved);
  const [tab, setTab] = useState<"editor" | "viz">("editor");

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
          <div className="flex items-center gap-1">
            <div className="mr-2 flex rounded-md border border-border p-0.5">
              <button type="button" onClick={() => setTab("editor")} className={cn("rounded px-2 py-0.5 text-xs", tab === "editor" ? "bg-muted font-medium" : "text-muted-foreground")}>
                Editor
              </button>
              <button type="button" onClick={() => setTab("viz")} className={cn("rounded px-2 py-0.5 text-xs", tab === "viz" ? "bg-muted font-medium" : "text-muted-foreground")}>
                Visualization
              </button>
            </div>
            <Button size="sm" variant="secondary" onClick={markSaved} disabled={!dirty}>
              Save
            </Button>
          </div>
        </div>
        <div className="min-h-0 flex-1">{tab === "editor" ? <CodeMirrorEditor value={content} language="yaml" onChange={setContent} className="h-full" /> : <SchemaViz selectedId={selectedId} />}</div>
      </div>
    </div>
  );
}
