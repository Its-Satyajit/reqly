import { useEffect, useState } from "react";
import { Activity, Play, Square, RefreshCw } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Badge } from "#components/ui/badge";

type MonitorPoint = {
  at: string;
  ok: boolean;
  status: number;
  latencyMs: number;
};

export function MonitorView() {
  const [file, setFile] = useState("collections/users/list.yaml");
  const [points, setPoints] = useState<MonitorPoint[]>([
    { at: "00:00", ok: true, status: 200, latencyMs: 42 },
    { at: "00:05", ok: true, status: 200, latencyMs: 38 },
    { at: "00:10", ok: true, status: 200, latencyMs: 45 },
    { at: "00:15", ok: false, status: 503, latencyMs: 210 },
    { at: "00:20", ok: true, status: 200, latencyMs: 40 },
  ]);
  const [isMonitoring, setIsMonitoring] = useState(false);

  // Summary Metrics calculation
  const total = points.length;
  const okCount = points.filter((p) => p.ok).length;
  const availability = total > 0 ? ((okCount / total) * 100).toFixed(2) : "100.00";
  const avgLatency =
    total > 0
      ? Math.round(points.reduce((acc, p) => acc + p.latencyMs, 0) / total)
      : 0;

  const tick = () => {
    const d = new Date();
    const at = `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}:${String(d.getSeconds()).padStart(2, "0")}`;
    const ok = Math.random() > 0.15;
    const status = ok ? 200 : 500;
    const latencyMs = Math.round(30 + Math.random() * 80);
    setPoints((prev) => [...prev.slice(-19), { at, ok, status, latencyMs }]);
  };

  useEffect(() => {
    let timer: number | undefined;
    if (isMonitoring) {
      timer = window.setInterval(tick, 3000);
    }
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [isMonitoring]);

  const maxLatency = Math.max(1, ...points.map((p) => p.latencyMs));
  const minLatency = Math.min(...points.map((p) => p.latencyMs));

  return (
    <section className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="API monitoring dashboard">
      <PageHeader
        icon={Activity}
        title="API Monitoring Dashboard"
        description="Scheduled endpoint health checks, availability calculation, and latency trends"
      />
      <div className="flex flex-col gap-4 p-4">
        {/* Controls */}
        <div className="flex flex-wrap items-end gap-2">
          <div className="flex min-w-64 flex-1 flex-col gap-1">
            <label htmlFor="monitor-file" className="text-xs font-medium">
              Target Request File
            </label>
            <Input
              id="monitor-file"
              value={file}
              onChange={(e) => setFile(e.target.value)}
              placeholder="collections/users/list.yaml"
              className="font-mono text-xs"
            />
          </div>
          <Button
            size="sm"
            variant={isMonitoring ? "destructive" : "default"}
            onClick={() => setIsMonitoring(!isMonitoring)}
          >
            {isMonitoring ? (
              <>
                <Square className="size-3.5" data-icon="inline-start" />
                Stop Polling
              </>
            ) : (
              <>
                <Play className="size-3.5" data-icon="inline-start" />
                Start Polling (3s)
              </>
            )}
          </Button>
          <Button size="sm" variant="outline" onClick={tick}>
            <RefreshCw className="size-3.5" data-icon="inline-start" />
            Check Now
          </Button>
        </div>

        {/* Stats Strip §57.1 */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div className="rounded-lg border border-border bg-card p-3">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">Total Checks</p>
            <p className="mt-1 text-xl font-bold font-mono">{total}</p>
          </div>
          <div className="rounded-lg border border-border bg-card p-3">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">Availability</p>
            <p className="mt-1 text-xl font-bold font-mono text-status-ok">{availability}%</p>
          </div>
          <div className="rounded-lg border border-border bg-card p-3">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">Avg Latency</p>
            <p className="mt-1 text-xl font-bold font-mono">{avgLatency}ms</p>
          </div>
          <div className="rounded-lg border border-border bg-card p-3">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">Health Status</p>
            <div className="mt-1 flex items-center gap-1.5">
              <Badge variant={Number(availability) > 95 ? "default" : "destructive"}>
                {Number(availability) > 95 ? "HEALTHY" : "DEGRADED"}
              </Badge>
            </div>
          </div>
        </div>

        {/* Latency Chart / Points */}
        <div className="rounded-lg border border-border bg-card p-3 space-y-2">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-semibold">Latency History (ms)</h3>
            <span className="text-[11px] font-mono text-muted-foreground">last {points.length} samples (min: {minLatency}ms, max: {maxLatency}ms)</span>
          </div>
          <div className="rounded border border-border/40 bg-background/50 p-2">
            <svg viewBox="0 0 500 120" className="h-28 w-full">
              <polyline
                fill="none"
                stroke="var(--color-primary, #c93517)"
                strokeWidth="2"
                points={points
                  .map((p, idx) => {
                    const x = points.length > 1 ? (idx / (points.length - 1)) * 480 + 10 : 250;
                    const range = Math.max(1, maxLatency - minLatency);
                    const y = 110 - ((p.latencyMs - minLatency) / range) * 90;
                    return `${x},${y}`;
                  })
                  .join(" ")}
              />
              {points.map((p, idx) => {
                const x = points.length > 1 ? (idx / (points.length - 1)) * 480 + 10 : 250;
                const range = Math.max(1, maxLatency - minLatency);
                const y = 110 - ((p.latencyMs - minLatency) / range) * 90;
                return (
                  <circle
                    key={idx}
                    cx={x}
                    cy={y}
                    r="3.5"
                    className={p.ok ? "fill-primary" : "fill-destructive"}
                  />
                );
              })}
            </svg>
          </div>
        </div>

        {/* History Table */}
        <div className="rounded-lg border border-border bg-card p-3 space-y-2">
          <h3 className="text-xs font-semibold">Recent Health Checks</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-left font-mono text-xs">
              <thead>
                <tr className="border-b border-border text-[11px] text-muted-foreground uppercase">
                  <th className="py-1.5 px-2">Time</th>
                  <th className="py-1.5 px-2">Status</th>
                  <th className="py-1.5 px-2">Latency</th>
                  <th className="py-1.5 px-2">Result</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {[...points].reverse().map((p, idx) => (
                  <tr key={idx} className="hover:bg-muted/40">
                    <td className="py-1 px-2 text-muted-foreground">{p.at}</td>
                    <td className="py-1 px-2">{p.status}</td>
                    <td className="py-1 px-2">{p.latencyMs}ms</td>
                    <td className="py-1 px-2">
                      <span className={p.ok ? "text-status-ok font-semibold" : "text-status-error font-semibold"}>
                        {p.ok ? "PASS" : "FAIL"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </section>
  );
}
