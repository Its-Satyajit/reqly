# ADR 0031: Proxy & TLS per Request (Q1)

## Status
Accepted — grill Q1 (per-request `proxy` + `insecureSkipVerify` + `caFile`, env proxy + mTLS deferred)

## Context
`ROADMAP.md:516` §36 lists per-env/per-request proxy, cert inspection, mTLS, custom CAs. No proxy/tls fields in `internal/request` or `requestfile`; data-layer `useProxyTlsStore` exists but no engine wiring.

## Decision
Q1 slice: **per-request only** — `request.proxy: string` (URL) + `request.tls: {insecureSkipVerify bool, caFile string}` stored in request file, applied in `Client.Execute` transport. Precedence request > OS env; env proxy + `clientCert`/`clientKey` + cert inspection are §36b.
Q2: **mTLS deferred** — §36-T1 ships `insecureSkipVerify`+`caFile` only; `clientCert`/`clientKey` are §36b (same `tls.Config` build).

## Consequences
Q3: **fail hard** — invalid CA/proxy → send error, no insecure fallback.
Q4: **Request Settings panel** — per-request `proxy` + `tls` fields (Settings panel, CLI `--proxy`/`--insecure`/`--ca-file`); env proxy deferred.

