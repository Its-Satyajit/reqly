# Phase 2: Differentiating Features (P1)

## Phase 2 — Differentiating Features (P1)

Features that make Reqly more capable than a basic API client.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §56

> **GUI milestones:** Most Phase 2 features have core/CLI shipped but no desktop GUI.
> See [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) for GUI-specific tracking (GUI-5 through GUI-16).

### §56.1 OpenAPI Spec Editor

- [x] Spec editor tree + YAML CodeMirror editor — `features/spec-editor/SpecEditorView.tsx` + `stores/useSpecEditorStore.ts` — 2026-08-26
- [ ] Interactive endpoint editing with validation (P1 GUI pending)

### §56.2 Schema Visualization

- [x] Schema graph — `lib/schemaGraph.ts` + `lib/schemaGraph.test.ts` — 2026-08-26
- [ ] Interactive graph UI with zoom/pan/node selection (P1 GUI pending)

### §56.3 Request Templates

- [x] Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [ ] Template picker UI in request builder (P1 GUI pending)

### §56.4 Proxy / TLS Controls

- [x] Proxy & TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [ ] Proxy/TLS configuration panel UI (P1 GUI pending)

### §56.5 Data-driven Testing

- [x] Data-driven testing — CSV/JSON dataset lib + zustand store + 23 tests — 2026-08-26
- [ ] Dataset picker + runner integration UI (P1 GUI pending)

### §56.6 CI/CD Integration

- [x] CI/CD support — CLI command generation + GitHub Action YAML + zustand store + 13 tests — 2026-08-26
- [ ] CI/CD configuration panel UI (P1 GUI pending)

### §56.7 Full Mock Server GUI

- [~] Mock server — CLI `reqly mock` with path/method matching, schema/example generation, `--delay`, `--fail-every`; stateful mocks, scenarios, fault injection, and zustand store shipped — 2026-08-26
- [ ] Full mock server GUI with route editor, scenario manager, logs viewer (P1 GUI pending)

### §56.8 GraphQL / gRPC Documentation

- [x] GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26
- [ ] Documentation browser UI (P1 GUI pending)

### Other P1 Items

- [x] Code generation — `internal/exporter.Generate` (request → cURL/JS/Python/Go, header/body/auth, `[SECRET]` masked, `reqly export code` + desktop `Copy as`, golden files) — [ADR 0016](docs/adr/0016-code-generation.md)
- [x] API diff + breaking-change detection (endpoints, params, schemas, auth, response types) — `internal/diffing` (`OpenAPIFiles` structural diff + `breaking.go` severity classification), `reqly diff <file1> <file2>`, desktop Diff view
- [x] Request/response diff (JSON structural) — `diffing.JSON` + `reqly diff`
- [x] Environment diff — `reqly env diff` + desktop env tools panel
- [~] HAR import/export + replay — import (`internal/importer/har.go`) + export (`internal/exporter/har.go`) shipped; HAR-specific replay pending (history replay via `HistoryReplay` shipped)
- [x] JWT tooling (decode, claims viewer, signing, verification) — decode/claims viewer (`reqly jwt decode`, ADR 0021), HMAC/RSA verification (`internal/jwt.VerifyToken`, `reqly jwt verify`, ADR 0037), and Goja sandbox assertion `reqly.verifyJWT()` shipped ([M53](docs/spec/m53-jwt-signature-verification.md))
- [~] GraphQL introspection / gRPC reflection tooling — GraphQL schema introspection + summary shipped (`internal/graphql/introspect.go`, desktop GraphQL browser); gRPC reflection not started
- [ ] Advanced HTTP: HTTP/2, HTTP/3, streaming, chunked transfer, keep-alive
- [x] Pagination runner (page/offset/cursor/link-header, stop conditions, aggregation) — `internal/pagination` + `reqly pagination run` ([ADR 0022](docs/adr/0022-pagination-runner.md)) + desktop runners panel
- [x] Bulk request execution (CSV/JSON inputs, sequential/parallel, concurrency) — `internal/bulk` + `reqly bulk run --data` ([ADR 0023](docs/adr/0023-bulk-runner.md)) + desktop runners panel
- [x] Retry & resilience — engine-level `request.retry` block ([ADR 0024](docs/adr/0024-retry-resilience.md))
- [~] API documentation generation (REST + GraphQL + realtime) — REST shipped: `reqly docs generate` + desktop Docs panel (G-15); GraphQL SDL parser + zustand store shipped (2026-08-26); realtime doc output deferred

---

