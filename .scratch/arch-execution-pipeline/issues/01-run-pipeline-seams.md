# 01 — Run pipeline + workspace-wiring seams

**What to build:** One deep method owns everything a send needs: `core.RequestService.Run(ctx, req, RunOptions)` resolves the environment (`REQLY_ENV` → flag → file pill → workspace descriptor), layers variables plus optional runtime vars, opens the token store, executes with retry observers, records history with raw bytes, and returns masked-in-place output with the acquired token never exposed. Token-store backend selection moves into `secrets.OpenForWorkspace` so every front-end shares one adapter.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [x] `secrets.Opened{Store, Backend, Warning}` + `OpenForWorkspace(root, defaultBackend string) Opened` honoring `REQLY_TOKEN_STORE`, keychain→file fallback preserved, no stderr printing inside the package; table-driven tests
- [x] `core.RunOptions{EnvFlag, FileEnv string; RuntimeVars *variables.Set; RecordHistory *bool; OnRetry func(request.RetryEvent)}` and `RunResult{Response *response.Response (masked in place)}`
- [x] `Run` composes: environment Selection → variable layering (file/env under runtime) → token-cache client → Execute → history record (raw) → mask headers/body/error; acquired token never serialized or returned
- [x] No workspace found → executes without caching/recording (matches today's CLI behavior), warning surfaced via `RunResult`
- [x] Unit tests with httptest: precedence order, runtime-var overlay wins over env scope, masking of secret env values in returned body/headers/error, history entry written raw then shown masked, OnRetry observed, RecordHistory=false writes nothing
- [x] `go vet` + `gofmt -l` + `go test -race ./internal/core ./internal/secrets` green
