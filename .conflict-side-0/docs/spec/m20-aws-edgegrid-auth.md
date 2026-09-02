# Spec: Desktop AWS SigV4 + Akamai EdgeGrid Auth (Milestone 20)

> **Status:** Draft — grill settled 2026-08-21
> **Scope:** Finish P0 auth gap `ROADMAP.md:100` (AWS Signature, Akamai EdgeGrid) — 2 schemes, 5 tickets
> **Stack:** `internal/auth` + `apps/desktop/frontend/src/bridge.ts` + `frontend/src/lib/authSchemes.ts` + `frontend/src/features/auth-editor/AuthEditor.tsx`
> **Predecessor:** M19 `docs/adr/0011-desktop-request-auth-editing.md` (Auth tab, draft auth, non-blocking warnings)

## Problem Statement

`ROADMAP.md:100` lists `OAuth 1.0, AWS Signature, Akamai EdgeGrid, custom auth` as the last unchecked P0 auth box. Core `internal/auth/doc.go:19` advertises AWS and EdgeGrid but no scheme is registered (`internal/auth/schemes.go:106` registers only bearer/basic/apikey). A request needing AWS SigV4 (e.g. API Gateway, S3) or Akamai OPEN must today hand-edit `auth.config` and cannot be configured from the desktop Auth tab. OAuth 1.0/custom remain deferred to keep M20 small (same sizing as M19).

## Solution

Register two schemes in `internal/auth`:

* **`aws` — AWS Signature Version 4**: signs every request (method, canonical URI, query, headers, payload hash) with `accessKey`/`secretKey` + `region`/`service` (+ optional `sessionToken` for STS). Injects `Authorization: AWS4-HMAC-SHA256 ...` + `X-Amz-Date` (+ `X-Amz-Security-Token` when present).
* **`edgegrid` — Akamai EdgeGrid (OPEN)**: signs with `clientToken`/`clientSecret`/`accessToken`/`host`, producing `Authorization: EG1-HMAC-SHA256 ...` per Akamai spec.

Both implement `auth.Scheme` (`internal/auth/auth.go:38`) + `SecretKeyScheme` (`auth.go:46`) and interpolate config values via `vars.Interpolate` like `basicScheme.Apply` (`schemes.go:30`). Unknown `auth.type` stays an error (`auth.go:105`).

Desktop: add two flat entries to the Auth tab picker (`AUTH_SCHEME_LABELS` in `frontend/src/lib/authSchemes.ts`) — "AWS SigV4" and "Akamai EdgeGrid" — each with typed `FieldDef[]` forms reusing `AuthFieldRow`/`SecretBadge` and `isSensitiveKey` derived from `SecretKeys()`. Follow `docs/adr/0011...:6` plaintext-sensitive-flagged inputs, non-blocking save warnings in `RequestEditor`, hard errors on send.

## User Stories

1. As a desktop user, I want to pick "AWS SigV4" in the Auth tab and fill accessKey/secretKey/region/service (+ optional sessionToken), so my request signs for API Gateway/S3 without editing the file.
2. As a desktop user, I want to pick "Akamai EdgeGrid" and fill clientToken/clientSecret/accessToken/host, so my request signs for Akamai OPEN.
3. As a CLI user, I want `auth.type: aws` / `edgegrid` in a request file to sign via `auth.Apply`, with missing keys failing fast on send, so file-driven runs match desktop.
4. As a user, I want sensitive fields flagged visually and masked in output (via `MaskValues` `internal/auth/auth.go:53`), so secrets never leak.

## Auth Config

**AWS** `auth.type: "aws"` (`ROADMAP.md:100`):
```yaml
auth:
  type: aws
  config:
    accessKey: "{{aws_access_key}}"      # required
    secretKey: "{{aws_secret_key}}"      # required, sensitive
    region: "us-east-1"                  # required
    service: "execute-api"               # required
    sessionToken: "{{aws_token}}"        # optional, sensitive
```
`SecretKeys() = ["secretKey","sessionToken"]`

**EdgeGrid** `auth.type: "edgegrid"`:
```yaml
auth:
  type: edgegrid
  config:
    clientToken: "{{akamai_client_token}}"   # required
    clientSecret: "{{akamai_client_secret}}" # required, sensitive
    accessToken: "{{akamai_access_token}}"   # required, sensitive
    host: "akab-xxxx.luna.akamaiapis.net"    # required
```
`SecretKeys() = ["clientSecret","accessToken"]`

All values interpolated; empty required → `Apply` error. No `host` override for AWS v1, no `maxBody` for EdgeGrid v1.

## Inheritance & Lifecycle

Same as M19 `docs/adr/0011...:2-3`: `Inherit` = no `auth` block (re-resolved via `ResolveSend` seam, inherited auth applies), `none` = `auth.type: none` clears. No token cache/refresh (unlike `CachedTokenSource` `internal/auth/cache.go:1` for OAuth) — signing is per-request.

## Desktop UX

* Picker: `frontend/src/lib/authSchemes.ts` `AUTH_SCHEME_LABELS = { ..., aws: "AWS SigV4", edgegrid: "Akamai EdgeGrid" }` + `FieldDef` arrays.
* Forms: `frontend/src/features/auth-editor/AuthEditor.tsx` flat two entries (no grouping), `AuthFieldRow` per key, `type="text"` + `SecretBadge` for sensitive (ADR 0011 §5).
* Warnings: `frontend/src/features/request-editor/RequestEditor.tsx` save-warnings banner — missing required → warning (non-blocking, like JWT `expiresIn` `internal/auth/jwt.go:103`), `isSensitiveKey` from field metadata.
* Bridge: `apps/desktop/frontend/src/bridge.ts` `fileRequest.auth` already carries `auth` (M19 T2 `bridge.ts:94` `normalizeVariables` pattern), no new DTO — just new type strings.

## Validation

* **Save**: non-blocking warnings for missing required (reuse M19 T5 wiring).
* **Send**: `Apply` validates required, interpolates, computes signature; error surfaces in `ResponseViewer`/toast. No retry.

## Out of Scope

OAuth 1.0, custom generic auth, host override for AWS, EdgeGrid `maxBody`/`headersToSign`, STS AssumeRole flow, browser-specific signing.

## Verification

* `internal/auth/aws_test.go` + `edgegrid_test.go`: table-driven Apply success/failure, known-vector SigV4 (AWS test suite canonical), EdgeGrid known signature, secret masking via `MaskValues`, interpolation errors, unknown type still errors.
* `internal/core/workspace_test.go`: add `aws`/`edgegrid` to `TestWorkspaceServiceSaveRequestAuthSemantics` table (Inherit removes block, save writes block).
* `go test ./...` + `go test -race ./...`, `npm run lint` exit 0, `npm run typecheck` both frontends, `wails3 task build` (from `apps/desktop/backend` per `apps/desktop/backend/Taskfile.yml:38`).

## Docs

* New `docs/adr/0012-aws-edgegrid-auth.md` (why flat config, why no inference, why deferred OAuth1/custom).
* Update `CONTEXT.md` glossary: `AWS Signature V4`, `Akamai EdgeGrid`.
* Tick `ROADMAP.md:100` + `docs/features.md:16` + Progress Tracker Phase 1 %.

## Ticket Split (5)

* T1 Core schemes `internal/auth/aws.go` + `edgegrid.go` + tests
* T2 Bridge/types `apps/desktop/frontend/src/bridge.ts` + `frontend/src/lib/authSchemes.ts`
* T3 Auth tab forms `frontend/src/features/auth-editor/AuthEditor.tsx`
* T4 Warnings/validation `frontend/src/features/request-editor/RequestEditor.tsx` + `internal/auth` errors
* T5 Docs ADR + CONTEXT + ROADMAP
