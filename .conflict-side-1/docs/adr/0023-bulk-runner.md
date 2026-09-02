# ADR 0023: Bulk Runner (M31)

## Status
Accepted

## Context
`ROADMAP.md:346` (`Bulk request execution — CSV/JSON inputs, sequential/parallel with concurrency control`) and `docs/features.md:26` (Bulk: CSV/JSON/Variable inputs, Sequential/Parallel/Concurrency) remain unchecked. `internal/pagination` loops one request via `sendFn` + `OnStep`, `internal/runner` is single-flight sequential, and `variables` already interpolates `{{var}}` per request via `ScopeRuntime`. Design questions: which inputs ship in M31 (CSV, JSON, variables, generated), execution mode (sequential vs parallel semaphore vs ordered output), where bulk config lives (request file vs CLI only), and CLI shape (`reqly bulk run` vs `reqly pagination` reuse).

## Decision
1. **M31 is CSV+JSON MVP without variable/generated datasets.** `reqly bulk run <request-file> --data data.csv|data.json [--parallel] [--concurrency 5] [--continue-on-error]` reads CSV header→`map[string]string` or JSON array of objects (values stringified), repeats the request per row interpolating row fields as `{{key}}`. Sequential default, parallel via `--parallel` + semaphore `concurrency` (default 5 parallel, 1 sequential); output ordered by input index but sent concurrent; `--continue-on-error` keeps going on non-2xx (default stop on first error). `internal/bulk.Run` pure `Run(ctx, req, rows, opts, sendFn, onStep)` reuses pagination/runner seams.

## Considered Options
- **CSV only for M31, JSON deferred** — rejected: JSON array is the natural counterpart and shares the same row-map seam.
- **Variable/generated datasets in M31** — rejected: they reuse `variables.TagGenerator` and are additive via `--var` flag in M31b.
- **`reqly pagination` reuse for bulk** — rejected: bulk is row-driven repetition, not next-link traversal; separate `bulk` root keeps CLI discoverable (`collection`/`pagination`/`bulk`).

## Consequences
- **Positive:** Closes `ROADMAP.md:346` inputs + modes with one pure package, one Cobra seam, no new storage, and clear M31b for variable/generated inputs + desktop.
- **Trade-off:** M31 has no `--var` rows or `{{$randomInt}}` generated rows; `--continue-on-error` is opt-in.
