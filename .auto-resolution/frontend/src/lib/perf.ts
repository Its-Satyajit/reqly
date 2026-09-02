export interface PerfResultView {
  rps: number;
  total: number;
  p50Ms: number;
  p95Ms: number;
  p99Ms: number;
  errorRate: number;
  statusCounts: Record<string, number>;
  latenciesMs?: number[];
}

export interface PerfAdapter {
  run(input: {
    filePath: string;
    rps: number;
    durationMs: number;
    concurrency: number;
  }): Promise<{ result: PerfResultView }>;
}

let perfBridge: PerfAdapter | null = null;

export function setPerfBridge(adapter: PerfAdapter): void {
  perfBridge = adapter;
}

export function getPerfBridge(): PerfAdapter {
  if (!perfBridge) {
    return {
      async run(input) {
        // Fallback for browser development mode
        const total = input.rps * Math.round(input.durationMs / 1000);
        return {
          result: {
            rps: input.rps,
            total,
            p50Ms: 42,
            p95Ms: 89,
            p99Ms: 120,
            errorRate: 0.02,
            statusCounts: { "200": Math.max(1, total - 2), "500": 2 },
            latenciesMs: Array.from({ length: 20 }, (_, i) => 20 + i * 5),
          },
        };
      },
    };
  }
  return perfBridge;
}
