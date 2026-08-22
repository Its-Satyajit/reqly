# ADR 0015: Dynamic Values & Template Tags (M23)

## Status
Accepted

## Context
`ROADMAP.md:83` (dynamic values & template tags) and `docs/features.md:15` are the last P0 variable gap: UUID, timestamp, random, runtime values that must be available in URLs, headers, bodies, auth config, and scripts without hand-editing. `internal/variables` already resolves `{{key}}` via scope precedence; `{{` without `$` must stay variable-only. The design questions are which tags ship in M23, what syntax avoids clashing with variables, where interpolation happens, whether a tag generates once per request or per occurrence, how unknown/parametric tags are handled, and how to keep history faithful while tests stay deterministic.

## Decision
1. **Core 5 tags, `{{$name}}` syntax (Postman-compatible).** `{{$uuid}}` (v4), `{{$timestamp}}` (unix sec), `{{$isoTimestamp}}` (ISO8601 UTC), `{{$randomInt}}` (0-1000), `{{$randomString}}` (8-char alphanum). Zero-arg for M23; regex `\{\{\$(.*?)\}\}` already captures space-separated args (`{{$randomInt 1 100}}`) but M23 impl ignores args and generates the default range — parametric `min max`/`length` becomes M23b. Custom tags deferred (no `RegisterTag` yet).
2. **Strict `{{$` vs `{{`.** `{{$tag}}` is *never* a variable lookup; `{{key}}` without `$` is *never* a dynamic tag. No fallback. Unknown `{{$unknown}}` is left literal (not error) with a non-blocking `saveWarnings` ("Unknown dynamic tag `{{$unknown}}` will be sent as-is"), mirroring M21 body warnings. Keeps `variables` scopes pure.
3. **Interpolated everywhere, per occurrence.** `variables.Interpolate` (used by `request.Client` URL/headers/query/body/auth and scripting) is the single pass that expands `{{$tag}}` alongside `{{var}}` — so tags work in every string variables do. Per occurrence generation: each `{{$tag}}` match in one request yields a fresh value (two `{{$uuid}}` → two UUIDs). History stores the resolved bytes (like variables), so `history.Show` is masked resolved and `HistoryReplay` is exact; the request file on disk retains the raw `{{$tag}}`.
4. **Desktop picker + autocomplete, history resolved.** `{{$` autocomplete (filter 5 tags) plus a `{{$}}` pill button beside URL/Body/Params/Headers inserts at cursor. History UI shows resolved values; no extra persistence.
5. **TagGenerator seam for determinism.** `variables` internal `TagGenerator` interface `Generate(tag string, args []string) (string, bool)` — `defaultGenerator` uses `uuid.New/time.Now/rand`; tests inject `fixedGenerator` (e.g. `FixedUUID = 00000000-...`). `Generate` returns `bool` = known tag, so future `RegisterTag` can plug in without changing `Interpolate`. `internal/variables` remains the highest seam; `request`/`scripting` stay unchanged beyond calling `Interpolate`.

## Considered Options
- **Full 9 tags (add `{{$randomEmail}}`, `{{$randomIP}}`)** — rejected: low M23 ROI, easy add later via TagGenerator without seam change.
- **Unknown → error (fail send)** — rejected: breaks save/send contract; non-blocking warning keeps send resilient like M21 body warnings.
- **Per-request cached (same `{{$uuid}}` twice → same value)** — rejected: Postman semantics are per occurrence; caching can be done by capturing one generation into a variable via post-script if needed.
- **`RegisterTag` now** — rejected: no plugin demand yet; adding registry now adds speculative generality — interface already supports it, registration is deferred.

## Consequences
- **Positive:** Closes last P0 variable gap with one seam (`variables.Interpolate` + `TagGenerator`), 5 tags everywhere, deterministic tests, faithful history, and a clear M23b path for parametric/custom tags without new seams.
- **Trade-off:** M23 parametric `{{$randomInt 1 100}}` generates default `0-1000` until M23b; custom tags remain deferred, so users cannot yet extend via plugins.

