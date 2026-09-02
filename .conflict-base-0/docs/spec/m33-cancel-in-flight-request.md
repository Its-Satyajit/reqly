# Spec: Cancel in-flight request (Milestone 33)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q5 confirmed)
> **Scope:** Phase 1 §1.5a `ROADMAP.md` — "cancel in-flight request" — desktop editor Send path only
> **Stack:** `internal/core` (ctx seam on `Send`) + `apps/desktop/backend` (`sendId` registry + `CancelSend` binding) + shared frontend (Stop button, cancelled state) — no new deps
> **Predecessor:** ADR 0010 (collection-run cancel via `runCancels` registry), ADR 0025 (core execution pipeline)

## Problem Statement

The transport is already cancellation-ready end to end: `http.NewRequestWithContext` (`internal/request/client.go:333`) and retry backoff selects on `ctx.Done()`. Collection runs can be cancelled through the `runCancels` registry (`app.go:311`). But a single desktop send has no cancellation path anywhere above the engine: `AppService.SendRequest` hard-codes `context.Background()` (`app.go:132`), and the Send button merely disables while loading. A hung endpoint pins the tab until timeout.

## Solution

* **Core seam** — `RequestService.Send(r request.Request, vars ...*variables.Set)` gains `ctx context.Context` as its first parameter and passes it to `client.Execute`, removing the last core send path that hard-codes `context.Background()`. Internal callers updated; no behavior change when callers pass `context.Background()`.

* **Desktop bridge** — `SendOptions` grows an optional `SendID string`. When non-empty, `SendRequest` derives `ctx, cancel := context.WithCancel(...)` around its pipeline call and registers `cancel` in a `sendMu sync.Mutex`-guarded `sendCancels map[string]context.CancelFunc`, mirroring `runCancels`; the map entry is removed in a defer. New binding `CancelSend(sendID string) error` looks up and calls cancel under lock — a missing/expired id is a no-op returning nil (send finished before Stop was pressed). Empty `SendID` → no registration, current behavior.

* **Cancelled-send artifacts** — none. Guaranteed by the existing early return in `core.Run` (`run.go:151-153`): engine error → no cookie ingest, no history record. A cancelled send leaves zero traces.

* **Frontend** — per-tab `sendId = crypto.randomUUID()` minted per send and passed through the sender adapter into `SendOptions`. While `loading`, the Send button toggles to **Stop**; clicking calls a new `cancel()` store action that invokes the adapter's `cancelSend(sendId)` and drops the token so a late response is discarded. On cancellation the response pane renders a distinct neutral "Request cancelled" state — not an error. Stale-response protection stays with the existing `sendTokens`.

* **Scope boundary** — editor Send only. History replay returns raw bytes with no network wait; CLI already honours SIGINT ctx; collection-run cancellation exists.

## User Stories

1. As a desktop user, I want to hit Stop while a request hangs so I am not stuck waiting for its timeout.
2. As a desktop user, I want a cancelled send to leave nothing in history so replays and search stay clean.
3. As a desktop user, I want Stop to be harmless after the response arrived (no-op) so races never surface as errors.
4. As a developer, I want `RequestService.Send` to accept ctx so no core seam bypasses cancellation.
5. As a developer, I want cancel-after-finish to be a silent no-op so UI and backend disagree safely.

## Implementation Decisions

- Registry pattern copied from `runCancels`/`WorkspaceRunCancel` (ADR 0010) — same mutex discipline, same delete-on-finish defer; no new abstraction ("sends" are not long-lived objects worth a service).
- Cancellation detection on the frontend: bridge maps Go error containing context canceled / "canceled" to a typed cancelled outcome; the store distinguishes `{status:'cancelled'}` from `{error}`.
- Wails `$CancellablePromise` is JS-side only — it cannot abort the Go call — hence the explicit `CancelSend` binding rather than promise abandonment.
- No ADR: hard-to-reverse? No. Surprising? No — it extends the established registry pattern. Real trade-off? Marginal.
- Glossary: add **Send Cancellation** to CONTEXT.md.
