# Phase 3: Power-User Features (P2)

## Phase 3 — Power-User Features (P2)

Advanced functionality for experienced developers and teams.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §57

### §57.1 API Monitoring Dashboard

- [x] Scheduled requests/collections, health checks, latency/availability trends, alerts, CLI `reqly monitor run`, and Desktop MonitorView (`internal/monitor`, [M38](docs/spec/m49-monitor-scheduler.md), [ADR 0033](docs/adr/0033-monitor-scheduler.md)) shipped

### §57.2 Performance Testing

- [x] RPS, latency, P95/P99, error rate, status distribution, CLI `reqly perf run`, and Desktop PerfView (`internal/perf`, [M37](docs/spec/m48-perf-runner.md), [ADR 0032](docs/adr/0032-perf-runner-tanstack-charts.md)) shipped

### §57.3 MQTT / Socket.IO

- [x] MQTT publish/subscribe, topics, QoS (0,1,2), retain, TLS, CLI `reqly mqtt pub/sub`, and Goja binding `reqly.mqtt` (`internal/mqtt`, [M57](docs/spec/m57-mqtt-protocol-engine.md), [ADR 0041](docs/adr/0041-mqtt-protocol-engine.md)) shipped
- [x] Socket.IO connections, events, rooms, namespaces, CLI `reqly socketio connect/emit`, and Goja binding `reqly.socketio` (`internal/socketio`, [M58](docs/spec/m58-socketio-protocol-engine.md), [ADR 0042](docs/adr/0042-socketio-protocol-engine.md)) shipped

### §57.4 Dependency Graph

- [x] API dependency graph visualization and variable propagation graph (`frontend/src/features/dep-graph/DepGraphView.tsx`) shipped

### §57.5 Request Replay

- [x] Exact / modified vars / other env / captured traffic replay via HAR live replay engine and CLI `reqly history replay --har` (`internal/importer`, [M55](docs/spec/m55-har-replay-engine.md), [ADR 0039](docs/adr/0039-har-replay-engine.md)) shipped

### §57.6 In-app Developer Tools / Debugger

- [x] Request/auth/variables/script/runtime/network inspection via Bottom Utility Panel (`Console`, `Network`, `Tests`, `Variables`, `Cookies`, `⌘J`) shipped

### §57.7 Git GUI

- [x] Git sidebar panel, stage/commit/diff/history/status indicators (`internal/git`, `frontend/src/features/git-view/GitView.tsx`) shipped

### §57.8 Network Interception / Timeline Debugging

- [x] Request timeline debugging (DNS/connect/TLS/request/server/response/transfer) — `response.Timings` (`internal/response`), `internal/request.Client` `httptrace` synthesis, CLI `reqly run --timeline`, and Goja binding `reqly.response.timings` ([M59](docs/spec/m59-request-timeline-debugging.md), [ADR 0043](docs/adr/0043-request-timeline-debugging.md)) shipped
- [x] Traffic capture, inspection, HAR export/replay (`internal/importer`, `internal/exporter`, M28, M55) shipped

### Other P2 Items

- [x] API changelog (from specs + Git changes) — `internal/diffing.GenerateChangelog`, Markdown/JSON formatting, SemVer bump classification, CLI `reqly changelog <old> <new> [--format markdown|json] [--fail-on-breaking]`, and Goja `reqly.generateChangelog` binding ([M60](docs/spec/m60-api-changelog-semver.md), [ADR 0044](docs/adr/0044-api-changelog-semver.md)) shipped
- [x] Browser integrations (Chrome/Firefox/Safari/Edge DevTools 'Copy as fetch' & cURL parser, CLI `reqly import fetch`, Goja `reqly.importFetch` binding) — `internal/importer.ParseFetch`, [M63](docs/spec/m63-devtools-fetch-importer.md), [ADR 0047](docs/adr/0047-devtools-fetch-importer.md) shipped
- [x] Advanced mock state (multi-scenario state machines, transition rules, status control endpoints `/__reqly/state`, CLI `reqly mock --scenario <file>`, and Goja `reqly.mock` bindings) — `internal/mocking.StateMachine`, [M64](docs/spec/m64-stateful-mock-engine.md), [ADR 0048](docs/adr/0048-stateful-mock-engine.md) shipped
- [x] Workflow engine — sequential workflow execution with variable extraction, conditional step evaluation via Goja, `{{var}}` interpolation for URL/headers/query/body, `internal/workflow` + CLI `reqly workflow <file>` + Goja `reqly.workflow.run` + desktop `AppService.WorkflowRun` binding; visual builder UI pending — `internal/workflow` (M65) shipped (core + CLI + desktop shipped / UI pending)
- [x] Self-hosted automation — local workflow scheduler (`Automation{Name, Workflow, Interval, Enabled, MaxRuns}` + `Scheduler.Run` with ticker, immediate first run, `MaxRuns` cap, `IsEnabled`/`Validate`), `internal/automation` + CLI `reqly automation run <file> [--once --interval --max-runs]` + desktop `AppService.AutomationRun`; cron/Git-ops UI pending — `internal/automation` (M66) shipped (core + CLI + desktop shipped / UI pending)

---

