import { useEffect, useRef, useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";

type MonitorPoint = { at: string; ok: number; latency: number };

export function MonitorView() {
  const [file] = useState("collections/users/list.yaml");
  const [points, setPoints] = useState<MonitorPoint[]>([{ at: "00:00", ok: 1, latency: 42 }]);
  const chartRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!chartRef.current || points.length === 0) return;
    let destroyed = false;
    (async () => {
      try {
        // @ts-ignore optional dep
        const charts = await import("@tanstack/charts");
        // @ts-ignore
        const linear = await import("@tanstack/charts/scales/linear");
        // @ts-ignore
        const point = await import("@tanstack/charts/scales/point");
        if (destroyed || !chartRef.current) return;
        const c = charts as unknown as { defineChart: (o: unknown) => unknown; mountChart: (el: HTMLElement, o: unknown) => { destroy: () => void } };
        const def = c.defineChart({
          marks: [(c as unknown as { lineY: (d: unknown, o: unknown) => unknown }).lineY(points, { x: "at", y: "ok", stroke: "#16a34a" })],
          scales: { x: { scale: () => (point as unknown as { scalePoint: () => unknown }).scalePoint() }, y: { scale: linear.scaleLinear } },
        } as unknown as never);
        const host = c.mountChart(chartRef.current!, { definition: def, height: 200, ariaLabel: "availability" } as unknown as never);
        return () => host.destroy();
      } catch {}
    })();
    return () => { destroyed = true; };
  }, [points]);

  return (
    <div className="flex h-full flex-col gap-3 p-3">
      <div className="flex items-center gap-2">
        <Input value={file} readOnly className="flex-1" aria-label="monitor file" />
        <Button onClick={() => setPoints((p) => [...p, { at: `${p.length}:00`, ok: Math.random() > 0.2 ? 1 : 0, latency: 20 + Math.random() * 80 }])}>Tick (mock)</Button>
      </div>
      <div ref={chartRef} className="rounded border border-border p-2" />
      <pre className="overflow-auto rounded bg-muted p-2 text-xs">{JSON.stringify(points, null, 2)}</pre>
    </div>
  );
}
