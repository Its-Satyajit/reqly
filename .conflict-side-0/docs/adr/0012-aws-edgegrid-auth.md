# ADR 0012: AWS SigV4 and Akamai EdgeGrid Auth

## Status
Accepted

## Context
`ROADMAP.md:100` (Phase 1 §1.3) lists `OAuth 1.0, AWS Signature, Akamai EdgeGrid, custom auth` as the last unchecked P0 auth box. Core `internal/auth/doc.go:19` already declares AWS and EdgeGrid but no scheme is registered (`internal/auth/schemes.go:106`). The desktop Auth tab (M19, ADR 0011) can edit every other core scheme but has no AWS/EdgeGrid forms. Both schemes sign every request with flat `auth.config` keys and have no token lifecycle (unlike OAuth). The design question is whether to bundle all four leftovers or close P0 with the two enterprise schemes that build directly on the M19 seams.

## Decision
1. **Scope is AWS SigV4 + EdgeGrid only.** M20 registers `aws` and `edgegrid` (`internal/auth/aws.go`, `edgegrid.go`). OAuth 1.0 and custom generic auth are deferred to a follow-up — they reuse the same `auth.config` + Auth tab seams, so no rework.
2. **Flat config, no inference.** AWS: `accessKey`/`secretKey`/`region`/`service` required, `sessionToken` optional (STS). EdgeGrid: `clientToken`/`clientSecret`/`accessToken`/`host` required. All camelCase to match existing `clientSecret`/`accessToken` naming (`internal/auth/schemes.go:83`). `region`/`service` are explicit; host is explicit for EdgeGrid; no `maxBody`/`headersToSign` v1.
3. **Plaintext, sensitive-flagged, non-blocking.** Like ADR 0011 §5–6: fields are `type="text"` with a `SecretBadge` derived from `SecretKeys()` (`secretKey`+`sessionToken` for AWS, `clientSecret`+`accessToken` for EdgeGrid), masked at send/output via `MaskValues` (`internal/auth/auth.go:53`). Missing required keys produce non-blocking save warnings in `RequestEditor` (reusing `authWarnings` `frontend/src/lib/authSchemes.ts:244`); `Apply` returns hard errors that surface on send. `Inherit`/`none` semantics unchanged (ADR 0011 §2–3) — `Inherit` removes the file's auth block, `none` writes `auth.type: none`.
4. **Per-request signing, no cache.** Signing is per-request (`X-Amz-Date`/`X-Amz-Content-Sha256`/`Authorization: AWS4-HMAC-SHA256` + `X-Amz-Security-Token`; `Authorization: EG1-HMAC-SHA256 ...`). No `CachedTokenSource` (`internal/auth/cache.go:1`) — unlike OAuth, there is no token to persist or refresh.
5. **Desktop is a thin extension of M19.** Two flat picker entries (`AUTH_SCHEME_LABELS` `frontend/src/lib/authSchemes.ts:168`) — "AWS SigV4" and "Akamai EdgeGrid" — each with a `FieldDef[]` form reusing `AuthFieldRow` (`frontend/src/features/auth-editor/AuthEditor.tsx:113`). No new components; `ORDERED_AUTH_SCHEMES` appends `aws`/`edgegrid`. The bridge (`apps/desktop/frontend/src/bridge.ts:181` `normalizeAuth`) already carries generic `auth` — no DTO change.

## Considered Options
- **Bundle OAuth 1.0 + custom in M20** — rejected: OAuth 1.0 (nonce/timestamp/HMAC-SHA1) is high complexity for low present demand; custom (generic header/query template) overlaps `apikey` and would complicate the first enterprise milestone. Keeping to 2 schemes keeps ticket count at 5, same sizing as M19.
- **Masked secret inputs (blank = unchanged)** — rejected per ADR 0011: request auth values are Git-visible, so plaintext editing is honest; masking would hide the file truth.
- **SDK-style inference (infer region/service/host)** — rejected: Reqly's `auth.config` is explicit and Git-native; inference would hide config and complicate `MaskValues`.
- **Grouped "Enterprise" submenu** — rejected: flat list is simpler and matches existing `apikey`/`jwt`/`digest` flatness.

## Consequences
- **Positive:** P0 auth closes to 100% (basic/bearer/apikey/jwt/digest/oauth2/aws/edgegrid + none), desktop has full parity with request files and CLI (`auth.Apply` `internal/auth/auth.go:97`), and the two highest-value enterprise auth types work end-to-end with minimal new UI.
- **Trade-off:** OAuth 1.0 and custom remain deferred; users needing them still hand-edit the file until the follow-up.
