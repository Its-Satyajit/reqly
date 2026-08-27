import { nodesForOpenApiContent, edgesForNodes } from "../../lib/schemaGraph";
import { useSpecEditorStore } from "../../stores/useSpecEditorStore";

export function DepGraphView() {
  const content = useSpecEditorStore((s) => s.content);
  const nodes = nodesForOpenApiContent(content);
  const edges = edgesForNodes(nodes);
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));
  return (
    <div className="flex h-full flex-col gap-2 p-3">
      <h2 className="text-sm font-semibold">Dependency Graph</h2>
      <p className="text-xs text-muted-foreground">Static OpenAPI graph from openapi.yaml — schemas + paths via schemaGraph.</p>
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
            <rect width={88} height={36} rx={6} className="fill-card stroke-border" strokeWidth={1} />
            <text x={44} y={22} textAnchor="middle" className="fill-foreground font-mono text-[11px]">{n.label}</text>
          </g>
        ))}
      </svg>
    </div>
  );
}
