import { useState, useCallback, useRef } from "react";
import { ZoomIn, ZoomOut, Maximize2, Move } from "lucide-react";
import { Button } from "#components/ui/button";
import { nodesForOpenApiContent, edgesForNodes, clampZoom, getTransform } from "../../lib/schemaGraph";
import { useSpecEditorStore } from "../../stores/useSpecEditorStore";

export function DepGraphView() {
  const content = useSpecEditorStore((s) => s.content);
  const nodes = nodesForOpenApiContent(content);
  const edges = edgesForNodes(nodes);
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const dragStart = useRef({ x: 0, y: 0 });

  const handleZoomIn = useCallback(() => setZoom((z) => clampZoom(z + 0.2)), []);
  const handleZoomOut = useCallback(() => setZoom((z) => clampZoom(z - 0.2)), []);
  const handleReset = useCallback(() => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
    setSelectedId(null);
  }, []);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      setIsDragging(true);
      dragStart.current = { x: e.clientX - pan.x, y: e.clientY - pan.y };
    },
    [pan],
  );
  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!isDragging) return;
      setPan({ x: e.clientX - dragStart.current.x, y: e.clientY - dragStart.current.y });
    },
    [isDragging],
  );
  const handleMouseUp = useCallback(() => setIsDragging(false), []);

  const selectedNode = selectedId ? nodes.find((n) => n.id === selectedId) : undefined;

  return (
    <div className="flex h-full flex-col gap-2 p-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold">Dependency Graph</h2>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={handleZoomOut} aria-label="Zoom out">
            <ZoomOut className="size-3" />
          </Button>
          <span className="w-10 text-center font-mono text-xs">{Math.round(zoom * 100)}%</span>
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={handleZoomIn} aria-label="Zoom in">
            <ZoomIn className="size-3" />
          </Button>
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={handleReset} aria-label="Reset view">
            <Maximize2 className="size-3" />
          </Button>
        </div>
      </div>
      <p className="flex items-center gap-1 text-xs text-muted-foreground">
        <Move className="size-3" />
        Drag to pan • Click node to select • {nodes.length} nodes, {edges.length} edges
      </p>
      {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions -- pan container handles drag */}
      <div
        className="select-none overflow-hidden rounded-lg border border-border bg-background"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        style={{ cursor: isDragging ? "grabbing" : "grab" }}
      >
        <svg viewBox="0 0 400 300" className="mx-auto h-[280px] w-full max-w-[480px]">
          <g transform={getTransform(zoom, pan.x, pan.y)}>
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
            {nodes.map((n) => {
              const isSelected = n.id === selectedId;
              return (
                <g
                  key={n.id}
                  transform={`translate(${n.x - 44}, ${n.y})`}
                  onClick={() => setSelectedId(n.id)}
                  role="button"
                  aria-label={`Select node ${n.label}`}
                  style={{ cursor: "pointer" }}
                >
                  <rect
                    width={88}
                    height={36}
                    rx={6}
                    className={isSelected ? "fill-primary/15 stroke-primary" : "fill-card stroke-border"}
                    strokeWidth={isSelected ? 1.5 : 1}
                  />
                  <text x={44} y={22} textAnchor="middle" className={isSelected ? "fill-primary font-mono text-[11px] font-medium" : "fill-foreground font-mono text-[11px]"}>
                    {n.label}
                  </text>
                </g>
              );
            })}
          </g>
        </svg>
      </div>
      {selectedNode && (
        <div className="rounded border border-border bg-card p-2 text-xs">
          <p className="font-medium">Selected: {selectedNode.label}</p>
          <p className="font-mono text-muted-foreground">id: {selectedNode.id}</p>
          <p className="text-muted-foreground">x: {selectedNode.x}, y: {selectedNode.y}</p>
          <div className="mt-1 flex gap-1">
            <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">{edges.filter((e) => e.from === selectedNode.id || e.to === selectedNode.id).length} connections</span>
          </div>
        </div>
      )}
    </div>
  );
}
