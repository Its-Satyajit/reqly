import { useEffect, useRef, useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";

type PerfResult = {
  rps: number;
  total: number;
  p50Ms: number;
  p95Ms: number;
  p99Ms: number;
  errorRate: number;
  statusCounts: Record<string, number>;
  latenciesMs?: number[];
};

export function PerfView() {
  const [file, setFile] = useState("collections/users/list.yaml");
  const [rps, setRps] = useState(10);
  const [duration, setDuration] = useState(10);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PerfResult | null>(null);
  const chartRef = useRef<HTMLDivElement>(null);
  const histRef = useRef<HTMLDivElement>(null);

  const run = async () => {
    setLoading(true);
    try {
      const wails = (window as unknown as { go?: { main?: { AppService?: { PerfRun?: (a: string, b: number, c: number, d: number) => Promise<{ result: PerfResult }> } } } }).go;
      const fn = wails?.main?.AppService?.PerfRun;
      if (!fn) {
        // Browser dev fallback: mock result
        setResult({ rps, total: rps * duration, p50Ms: 42, p95Ms: 89, p99Ms: 120, errorRate: 0.02, statusCounts: { "200": 98, "500": 2 }, latenciesMs: Array.from({ length: 20 }, (_, i) => 20 + i * 5) });
        return;
      }
      const res = await fn(file, rps, duration * 1000, rps);
      setResult(res.result);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!result || !chartRef.current || !histRef.current) return;
    let destroyed = false;
    // TanStack Charts DOM host — framework-agnostic per quick-start (mountChart)
    // Adapter @tanstack/charts/react requires React ^19 (tanstack-charts-installation) — DOM host works without adapter.
    // Fallback to simple div rendering when package not installed.
    (async () => {
      try {
        // @ts-ignore — optional dep, installed via pnpm add @tanstack/charts
        const charts = await import("@tanstack/charts");
        // @ts-ignore
        const scales = await import("@tanstack/charts/scales/linear");
        // @ts-ignore
        const point = await import("@tanstack/charts/scales/point");
        if (destroyed || !chartRef.current || !histRef.current) return;
        const latData = (result.latenciesMs ?? []).map((v, i) => ({ x: i, y: v }));
        const chart = (charts as unknown as { defineChart: (o: unknown) => unknown; mountChart: (el: HTMLElement, o: unknown) => { destroy: () => void } });
        // LineY latency snapshot
        const def = chart.defineChart({
          marks: [(charts as unknown as { lineY: (d: unknown, o: unknown) => unknown }).lineY(latData, { x: "x", y: "y", stroke: "#2563eb" })],
          scales: {
            x: { scale: () => (point as unknown as { scalePoint: () => unknown }).scalePoint(), axis: { label: "request" } },
            y: { scale: scales.scaleLinear, axis: { label: "ms" } },
          },
        } as unknown as never);
        const host = chart.mountChart(chartRef.current!, { definition: def, height: 200, ariaLabel: "latency" } as unknown as never);
        // Histogram barY
        const histData = Object.entries(result.statusCounts).map(([k, v]) => ({ code: k, count: v }));
        const histDef = chart.defineChart({
          marks: [(charts as unknown as { barY: (d: unknown, o: unknown) => unknown }).barY(histData, { x: "code", y: "count" })],
          scales: { x: { scale: () => point.scalePoint() }, y: { scale: scales.scaleLinear } },
        } as unknown as never);
        const histHost = chart.mountChart(histRef.current!, { definition: histDef, height: 160, ariaLabel: "status histogram" } as unknown as never);
        return () => {
          host.destroy();
          histHost.destroy();
        };
      } catch {
        // Library not installed — leave placeholder
      }
    })();
    return () => {
      destroyed = true;
    };
  }, [result]);

  return (
    <div className="flex h-full flex-col gap-3 p-3">
      <div className="flex items-center gap-2">
        <Input value={file} onChange={(e) => setFile(e.target.value)} placeholder="request file" className="flex-1" aria-label="request file" />
        <Input type="number" value={rps} onChange={(e) => setRps(Number(e.target.value))} className="w-20" aria-label="rps" />
        <Input type="number" value={duration} onChange={(e) => setDuration(Number(e.target.value))} className="w-20" aria-label="duration seconds" />
        <Button onClick={run} disabled={loading}>{loading ? "Running…" : "Run"}</Button>
      </div>
      {result && (
        <div className="flex flex-col gap-2">
          <div className="flex gap-4 text-xs">
            <span>p50 {result.p50Ms}ms</span><span>p95 {result.p95Ms}ms</span><span>p99 {result.p99Ms}ms</span><span>error {(result.errorRate * 100).toFixed(1)}%</span>
          </div>
          <div ref={chartRef} id="perf-latency-chart" className="rounded border border-border p-2" />
          <div ref={histRef} id="perf-hist-chart" className="rounded border border-border p-2" />
          <pre className="overflow-auto rounded bg-muted p-2 text-xs">{JSON.stringify(result, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}
