# ADR 0032: Perf Runner — CLI + GUI with TanStack Charts (Q1)

## Status
Accepted — grill Q1 (CLI + GUI together, chart = TanStack Charts)

## Context
`ROADMAP.md:517` §37 needs RPS/P95/P99 + histogram + GUI. User requested TanStack Charts (`https://tanstack.com/charts/latest/docs/quick-start.md`): `pnpm add @tanstack/charts`, `defineChart/lineY/mountChart` DOM host, `scaleLinear`/`scalePoint`, `ResizeObserver` width, `initialWidth` fallback.

## Decision
§37 ships CLI `reqly perf run` + GUI Perf dashboard in one slice using TanStack Charts (DOM host `mountChart(container,{definition})` adapted to React via `useEffect`/`ref`).

## Consequences
Q2: constant RPS ticker + concurrency cap + sorted slice P50/P95/P99 + status histogram (no hdrhistogram dep).
Q3: snapshot after run (lineY latency + barY histogram, tooltip), live as follow-up.
Q4: separate `internal/perf` result (no history pollution), wiring deferred.
