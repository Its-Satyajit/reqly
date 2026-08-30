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
- [x] Template picker UI in request builder — `features/request-editor/TemplatePickerSheet.tsx` (search, category filter, instantiate → draft) + `RequestEditor` Templates button + `lib/templates` 21 tests — 2026-08-26

### §56.4 Proxy / TLS Controls

- [x] Proxy & TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [x] Proxy/TLS configuration panel UI — `features/settings-view/ProxyTlsPanel.tsx` (proxy type/host/port/auth + TLS verify/minVersion) + `stores/useProxyTlsStore` + `lib/proxyTls` 22 tests — 2026-08-26

### §56.5 Data-driven Testing

- [x] Data-driven testing — CSV/JSON dataset lib + zustand store + 23 tests — 2026-08-26
- [ ] Dataset picker + runner integration UI (P1 GUI pending)

### §56.6 CI/CD Integration

- [x] CI/CD support — CLI command generation + GitHub Action YAML + zustand store + 13 tests — 2026-08-26
- [x] CI/CD configuration panel UI — `features/settings-view/CicdPanel.tsx` (pipeline/env/collection inputs → `lib/cicd` GitHub Action YAML + CLI command generator) — 2026-08-26

### §56.7 Full Mock Server GUI

- [x] Mock server — CLI `reqly mock` with path/method matching, schema/example generation, `--delay`, `--fail-every`; stateful mocks, scenarios, fault injection, and zustand store shipped — 2026-08-26
- [x] Full mock server GUI — `features/mock-view/MocksView.tsx` (route editor + scenario manager + fault injection + logs viewer + `stores/useMockStore`) — 2026-08-26

### §56.8 GraphQL / gRPC Documentation

- [x] GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26
- [x] Documentation browser UI — `features/docs-view/DocsView.tsx` + `features/graphql-browser/GraphqlBrowser.tsx` (SDL parse, search, Markdown) + `features/grpc-view/GrpcTab.tsx` — 2026-08-26

### Other P1 Items

- [x] Code generation — `internal/exporter.Generate` (request → cURL/JS/Python/Go, header/body/auth, `[SECRET]` masked, `reqly export code` + desktop `Copy as`, golden files) — [ADR 0016](docs/adr/0016-code-generation.md)
- [x] API diff + breaking-change detection (endpoints, params, schemas, auth, response types) — `internal/diffing` (`OpenAPIFiles` structural diff + `breaking.go` severity classification), `reqly diff <file1> <file2>`, desktop Diff view
- [x] Request/response diff (JSON structural) — `diffing.JSON` + `reqly diff`
- [x] Environment diff — `reqly env diff` + desktop env tools panel
- [x] HAR import/export + replay — import (`internal/importer/har.go`) + export (`internal/exporter/har.go`) + live HAR archive replay engine (`internal/importer/har_replay.go`, `reqly history replay --har archive.har`, Goja binding `reqly.replayHAR()`, [M55](docs/spec/m55-har-replay-engine.md), [ADR 0039](docs/adr/0039-har-replay-engine.md)) shipped
- [x] JWT tooling (decode, claims viewer, signing, verification) — decode/claims viewer (`reqly jwt decode`, ADR 0021), HMAC/RSA verification (`internal/jwt.VerifyToken`, `reqly jwt verify`, ADR 0037), and Goja sandbox assertion `reqly.verifyJWT()` shipped ([M53](docs/spec/m53-jwt-signature-verification.md))
- [x] GraphQL introspection / gRPC reflection tooling — GraphQL schema introspection (`internal/graphql/introspect.go`, `reqly graphql introspect`) and gRPC server reflection (`internal/grpc.Discover`, `reqly grpc reflect`, Goja binding `reqly.reflectGRPC()`, [M54](docs/spec/m54-grpc-server-reflection.md), [ADR 0038](docs/adr/0038-grpc-server-reflection.md)) shipped
- [x] Advanced HTTP: HTTP/1.1, HTTP/2 ALPN negotiation, chunked transfer, keep-alive controls — `internal/request.Request.HTTPVersion` & `DisableKeepAlives` (`internal/request/client.go`), CLI flags `--http2` & `--no-keepalive` on `reqly run` ([M56](docs/spec/m56-advanced-http-controls.md), [ADR 0040](docs/adr/0040-advanced-http-protocol-controls.md)) shipped
- [x] Pagination runner (page/offset/cursor/link-header, stop conditions, aggregation) — `internal/pagination` + `reqly pagination run` ([ADR 0022](docs/adr/0022-pagination-runner.md)) + desktop runners panel
- [x] Bulk request execution (CSV/JSON inputs, sequential/parallel, concurrency) — `internal/bulk` + `reqly bulk run --data` ([ADR 0023](docs/adr/0023-bulk-runner.md)) + desktop runners panel
- [x] Retry & resilience — engine-level `request.retry` block ([ADR 0024](docs/adr/0024-retry-resilience.md))
- [~] API documentation generation (REST + GraphQL + realtime) — REST shipped: `reqly docs generate` + desktop Docs panel (G-15); GraphQL SDL parser + zustand store shipped (2026-08-26); realtime doc output deferred

---

