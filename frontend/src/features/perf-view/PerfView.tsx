import { useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { getPerfBridge, type PerfResultView } from "#lib/perf";

export function PerfView() {
  const [file, setFile] = useState("collections/users/list.yaml");
  const [rps, setRps] = useState(10);
  const [duration, setDuration] = useState(10);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PerfResultView | null>(null);

  const run = async () => {
    setLoading(true);
    try {
      const adapter = getPerfBridge();
      const res = await adapter.run({
        filePath: file,
        rps,
        durationMs: duration * 1000,
        concurrency: rps,
      });
      setResult(res.result);
    } finally {
      setLoading(false);
    }
  };

  const latencies = result?.latenciesMs ?? [];
  const maxLat = Math.max(1, ...latencies);
  const minLat = Math.min(...latencies);

  const statusEntries = Object.entries(result?.statusCounts ?? {});
  const maxCount = Math.max(1, ...statusEntries.map(([, count]) => count));

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
          <div className="flex gap-4 text-xs font-mono">
            <span>p50 {result.p50Ms}ms</span>
            <span>p95 {result.p95Ms}ms</span>
            <span>p99 {result.p99Ms}ms</span>
            <span>error {(result.errorRate * 100).toFixed(1)}%</span>
          </div>

          {/* Latency Trend */}
          <div id="perf-latency-chart" className="rounded border border-border p-2">
            <p className="text-[11px] font-semibold text-muted-foreground mb-1">Latency Trend (ms)</p>
            {latencies.length > 0 ? (
              <svg viewBox="0 0 500 120" className="h-28 w-full">
                <polyline
                  fill="none"
                  stroke="#2563eb"
                  strokeWidth="2"
                  points={latencies
                    .map((val, idx) => {
                      const x = latencies.length > 1 ? (idx / (latencies.length - 1)) * 480 + 10 : 250;
                      const range = Math.max(1, maxLat - minLat);
                      const y = 110 - ((val - minLat) / range) * 90;
                      return `${x},${y}`;
                    })
                    .join(" ")}
                />
              </svg>
            ) : (
              <p className="text-xs text-muted-foreground">No latency data</p>
            )}
          </div>

          {/* Status Code Histogram */}
          <div id="perf-hist-chart" className="rounded border border-border p-2">
            <p className="text-[11px] font-semibold text-muted-foreground mb-1">Status Code Distribution</p>
            <div className="flex items-end gap-3 h-24 pt-2">
              {statusEntries.map(([code, count]) => {
                const heightPercent = Math.max(10, Math.round((count / maxCount) * 100));
                return (
                  <div key={code} className="flex flex-col items-center gap-1 font-mono text-xs">
                    <div
                      style={{ height: `${heightPercent}%` }}
                      className="w-8 rounded-t bg-primary"
                      title={`${code}: ${count}`}
                    />
                    <span>{code} ({count})</span>
                  </div>
                );
              })}
            </div>
          </div>

          <pre className="overflow-auto rounded bg-muted p-2 text-xs font-mono">{JSON.stringify(result, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}

