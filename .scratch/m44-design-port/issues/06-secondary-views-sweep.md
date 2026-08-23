# M44 T6 — Secondary views sweep

Blocked by: T5.
Blocks: nothing.

## Goal

Restyle every secondary view onto tokens + shell so the app is uniformly on the Atlas design language. One ticket per batch, merged here as checklist items; ship as a single PR or small stacked PRs.

## Checklist

- [ ] GraphQL browser + gRPC tab (views2 `gqlOps`, grpc response-only updates)
- [ ] WebSocket/SSE realtime view (`realtime-view`)
- [ ] Runners: test runner, bulk, pagination (`runners-panel`, `test-runner`)
- [ ] OpenAPI explorer + docs panel
- [ ] Import/export dialogs (curl/HAR/OpenAPI/Postman/Insomnia fixtures in views2)
- [ ] Mock servers view
- [ ] Breaking-change diff view
- [ ] History view
- [x] Settings view (incl. new Theme section from T1) — registry-driven theme picker incl. `system` preference + response layout; shipped as T6a. 2026-08-25

## Acceptance criteria

- [x] No view renders untokened styles — grep gate: zero hardcoded palette colors across features/components/editors/app. Both themes verified on Settings. 2026-08-25
- [ ] Loading/empty/error states present for every async view (demo parity).
- [x] typecheck + lint + React Doctor clean vs baseline (re-run at T6 closure); vitest 34/34. States audit: every async view exposes busy/error affordances; deferred polish tracked as Its-Satyajit/reqly#328. — 2026-08-25

## Reference

- `shared/views.js` + `shared/views2.js` per-view templates
