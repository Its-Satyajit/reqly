# ADR 0016: Code Generation (M24)

## Status
Accepted

## Context
`ROADMAP.md:203` (code generation) and `docs/features.md:43` list *request → cURL, JS, Python, Go* as a P1 differentiator, but the desktop has no “Copy as” and the CLI has no `export code`. The request can carry `BodyType: json/xml/form-data/urlencoded/raw/binary/graphql` (M21) and `Auth` (`basic/bearer/apikey/jwt/digest/oauth2/aws/edgegrid/none` + inheritance) — all via `request.Request` (resolved). The design questions are which languages ship in M24, how body/auth fidelity is handled, where the feature is exposed (CLI vs desktop vs MCP), and how secrets stay masked while `history` stores exact bytes.

## Decision
1. **Core 4 languages, `internal/exporter.Generate` single seam (beside `postman.go`).** `cURL` (faithful `--request/--header/--data-raw/--form/--data-binary`), `JavaScript` (`fetch` async/await), `Python` (`requests.request`), `Go` (`http.NewRequestWithContext` + `strings.NewReader` + `Header.Set`). Pure function `Generate(req Request, lang string, mask func(string) string) (string,error)` — `mask` is `environments.MaskValues` + `auth.MaskValues` (renders `[SECRET]` with `// [SECRET] masked` comment, never plaintext). Reuses `request.Request` directly; `variables` already resolved, so code shows concrete values (history’s resolved bytes, so “Copy as cURL” from History is exact).
2. **Body faithful per `BodyType`, form-data as `--form` for cURL only.** `json/xml/raw/graphql` → `--data-raw '...'` (cURL) / `body: '...'` (JS/Python/Go) with `graphql` JSON `{"query","variables"}`; `urlencoded` → `--data-urlencode`; `form-data` (file-aware rows) → `cURL --form 'key=value'` / `--form 'key=@./path'` (file rows, Git-native relative path) and `FormData`/`multipart` comment fallback for JS/Python/Go (full `multipart.Writer` is M24b); `binary` → `--data-binary @./relative/path` for cURL (file path, not inline bytes) and comment for others. No base64 inline.
3. **Auth `basic/bearer/apikey/jwt` as header/`--user`, rest TODO.** `basic` → `cURL --user user:pass` + `Authorization: Basic` header in JS/Python/Go; `bearer/apikey/jwt` → `Authorization` / `X-API-Key` header; `aws/edgegrid/oauth2/digest` → `// TODO: sign via SDK — not generated in M24` comment (low demand, high complexity); `none` → no header. Inherited auth already resolved into `req.Auth` by the workspace chain, so `Generate` sees the effective auth.
4. **Both CLI + desktop, shared `exporter`, no file download for M24.** CLI `reqly export code <request-file|collection-path> --lang cURL|js|python|go [--out <file>]` (like `reqly run`, `--env` respected, stdout when no `--out`). Desktop `RequestEditor` header bar “Copy as ▾” (cURL default) copies via `copyText` + toast, `HistoryView` row “Copy as cURL” from resolved entry, `ResponseViewer` “Copy as cURL” from last response’s request (when available). Shares `exporter` via `HistoryAdapter` pattern; MCP deferred (same `Generate` can be exposed later). No `collection` bulk export in M24 (single request/history entry only).
5. **Golden files, deterministic, no network.** `exporter/testdata/<lang>.golden` + table-driven `TestGenerate_<Lang>` (prior art `postman_test.go`) with `request.Request` fixtures (method/URL/headers/body/auth) → expected snippet literal; secrets in fixtures are masked via `mask` func, so tests assert `[SECRET]`.

## Considered Options
- **Extended 6 (add Node axios, Rust reqwest)** — rejected: low M24 ROI, easy add later via `exporter` without new seam.
- **Full `multipart.Writer` for Go in M24** — rejected: high complexity for `form-data` file rows; M24 `Go` fallback is comment + `strings.NewReader` raw, M24b will add `multipart.Writer`.
- **Base64 inline for `binary`** — rejected: Git-native relative path `@./path` is diff-friendly and matches `requestfile` `body.file` shape; inline bloats the snippet.
- **New `internal/codegen` package** — rejected: `internal/exporter` already is the “sharing request shapes” seam (`postman.go`); adding a new package fragments the seam — keep blast radius to one file `code.go`.

## Consequences
- **Positive:** Single pure `exporter.Generate` closes P1 code generation for 4 langs with one file, `Request` seam untouched, secrets never leak (masked), “Copy as” works from editor and history, CLI `export code` mirrors `run` resolution (`--env`).
- **Trade-off:** M24 `Go` `form-data` is comment fallback, `aws/edgegrid/oauth2/digest` are TODO comments, and bulk `collection` export is deferred — all are M24b follow-ups without new seams.

