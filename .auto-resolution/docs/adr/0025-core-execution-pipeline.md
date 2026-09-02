# ADR 0025: Core Owns the Execution Pipeline

## Status
Accepted

## Context
`internal/core` is documented as the shared application-service boundary for Desktop, CLI, and MCP, but the CLI imports it in exactly one place. Root cause: `RequestService.Send` is shallow — hardwired `context.Background()`, no observer hooks, no acquired-token exposure, no masking — so every front-end dropped below it. Five CLI commands each rebuild the same preamble (env selection → variable layering → token-cache client → masking → Execute → token capture), and the desktop maintains its own drifted copies of token-store opening and environment resolution (its copy drops `FileEnv` from selection precedence). Dual parity holds by copy-paste. Separately, history recording exists only on the desktop path; `reqly run` writes nothing to `history.db`.

## Decision
1. **`core.RequestService.Run(ctx, req, RunOptions) (*RunResult, error)` becomes the single execution pipeline**: environment selection (`REQLY_ENV` → flag → file pill → workspace descriptor), variable layering, optional runtime-variable injection (Bulk/Pagination rows), secret masking, retry observer wiring, history recording, and acquired-token capture all live behind it. Callers parse flags and render results; they never receive secrets (masked-in-place output; raw bytes exist only long enough to be recorded).
2. **Backend selection moves into its packages**: `secrets.OpenForWorkspace(root, defaultBackend)` honors `REQLY_TOKEN_STORE` for every front-end; environment selection is constructed through `internal/environments` only. The two hand-maintained adapters are deleted.
3. **Migration is expand–contract**: add `Run`, migrate the five CLI commands, then delete the preambles and `Send`. The desktop migrates under separate work that also fixes cookie attach and replay fidelity (see architecture review candidates 3).

## Considered Options
- **New `internal/executor` package** — rejected: duplicates the documented boundary; fails the deletion test (two modules, one job).
- **Raw response + caller-applied masker** — rejected: masking-by-convention recreates the leak class this removes; masking must be an invariant.
- **Unify pagination/bulk/retry loops into one abstraction** — rejected: their stop conditions sit at different levels (attempt budget vs structural inspection vs data exhaustion vs fail-fast); a common loop would need callbacks for everything and be shallow by construction. Only their step records/printers may converge later.

## Consequences
- **Positive:** Parity becomes structural; masking bugs fix once; MCP gets the whole pipeline through one method; CLI gains history recording for free.
- **Trade-off:** `reqly run` now writes `.reqly/history.db` by default (`RecordHistory *bool` opts out); callers lose access to raw response bytes (acceptable — display surfaces want masked output anyway).
