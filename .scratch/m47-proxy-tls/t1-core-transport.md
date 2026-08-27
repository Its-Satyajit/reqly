# T1 — Core transport (proxy + TLS)

> **Spec:** `docs/spec/m47-proxy-tls-per-request.md`
> **Blocks:** T2, T3

- `internal/request.Request {Proxy string; TLS *TLSConfig}` + `TLSConfig {InsecureSkipVerify bool; CAFile string}`
- `Client.Execute` builds `http.Transport` per send: proxy URL or `ProxyFromEnvironment`, `tls.Config` with `InsecureSkipVerify` + `RootCAs` from `CAFile` PEM; invalid CA/proxy → error fail-hard

**Done when:** `go test -race ./internal/request` transport tests green, `go vet` green
