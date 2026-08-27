# T3 — Desktop Perf view + TanStack Charts

> **Spec:** `docs/spec/m48-perf-runner.md`
> **Blocks:** T1
> **Chart:** `pnpm add @tanstack/charts` + React peers `react@19` (`tanstack-charts-installation: framework compatibility` — `@tanstack/charts/react` requires React `^19.0.0`, project `react@19.2.8` compatible; DOM host `mountChart(container,{definition,height,ariaLabel})` works framework-agnostic, React adapter `@tanstack/charts/react` optional — spec uses DOM host via `useEffect`/`ref` per quick-start, `defineChart` + `lineY`/`barY` + `scaleLinear`/`scalePoint` + `tooltip`)

- `frontend/src/features/perf-view/PerfView.tsx` — form RPS/duration/concurrency + Run → `AppService.PerfRun` → snapshot `lineY` (latency per second) + `barY` (status histogram), `host.update`/`host.destroy`
- `apps/desktop/backend/perf.go` — `PerfRun` bridge

**Done when:** `nub run typecheck` + `mountChart` renders in `jsdom` mock
