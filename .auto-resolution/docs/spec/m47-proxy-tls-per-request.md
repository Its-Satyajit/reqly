# Spec: Proxy & TLS per Request (Milestone 47 — §36 slice)

> **Status:** Draft — grill Q1–Q4 done 2026-08-27 (Recommend a/defer/hard/Request Settings), ADR 0031
> **Scope:** `ROADMAP.md:516` §36 — per-request `proxy` + `tls {insecureSkipVerify, caFile}` wired to transport; env proxy + mTLS + cert inspection deferred
> **Stack:** `internal/request` `Client.Execute` transport, `internal/requestfile` persistence, `apps/cli` flags, `frontend/src/features/request-editor/RequestSettings`

## Data Model
`request.Request { Proxy string `json:"proxy,omitempty"`; TLS *TLSConfig `json:"tls,omitempty"` }` `TLSConfig { InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`; CAFile string `json:"caFile,omitempty"` }` — `CAFile` workspace-relative (via `resolveTestPath` pattern `apps/desktop/backend/test.go:214`), `0644`. Interpolated via `variables` like headers.

## API Surface
- **Go core:** `Client.Execute` builds `http.Transport` per send: `ProxyURL` from `req.Proxy` (else `ProxyFromEnvironment`), `TLSClientConfig` with `InsecureSkipVerify` + `RootCAs` loaded from `CAFile` (PEM `x509.CertPool`); missing/invalid CA → `fmt.Errorf("tls: caFile …")` fail.
- **Request file:** `proxy: http://proxy:8080` + `tls: {insecureSkipVerify: true, caFile: ./certs/ca.pem}` preserved via `requestfile.Save` atomic.
- **CLI:** `reqly run/test --proxy <url> --insecure --ca-file <path>` overrides file (like `--env` precedence).
- **Desktop:** Request Settings panel (Bottom `RequestSettings`) — Proxy text + Insecure toggle + CA file picker; dirty-tracked, `useProxyTlsStore` removed in favour of request fields.

## Edge Cases
Empty `proxy` → `ProxyFromEnvironment`; invalid URL → send error; missing `caFile` → `tls: caFile not found`; bad PEM → `tls: failed to parse CA`; `insecureSkipVerify` + `caFile` both set → `InsecureSkipVerify` wins but CA still loaded (no conflict); no workspace → `resolveTestPath` outside-workspace → error.

## Testing Strategy
`internal/request` transport tests (proxy func returns URL, TLS config has `InsecureSkipVerify`/pool), `internal/requestfile` round-trip, CLI flag precedence, frontend Settings panel dirty/save tests; `go test -race` + `nub run typecheck` + `go vet`.
