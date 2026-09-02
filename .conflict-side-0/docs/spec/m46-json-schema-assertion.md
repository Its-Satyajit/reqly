# Spec: JSON Schema Assertion (Milestone 46 — §35 slice)

> **Status:** Draft — backfilled 2026-08-27 from grill Q1–Q4 (recommended a/b/a/defer)
> **Scope:** `ROADMAP.md:515` §35 — contract testing, JSON Schema response validation pipeline (`internal/jsonschema/validate.go:42` `Compile`+`Validate`)
> **Grill:** Q1 test assertion `json_schema` first (vs standalone CLI) / Q2 file path (`schema: ./schemas/user.json`, workspace-relative) / Q3 full violation list `$.items[2].price: …` / Q4 OpenAPI response validation deferred to §35b
> **Stack:** `internal/testing/assertion.go:46` `AssertJSONSchema`, `internal/jsonschema`, `internal/response`

## Problem
`reqly schema validate/inspect/generate` are CLI-only; collection test runner (`internal/testing` `Suite.Run`) has no contract assertion. No way to fail a test when response violates a schema.

## Solution
- **Assertion kind:** `AssertJSONSchema = "json_schema"` (`internal/testing/assertion.go:46`). Fields: `Path` = schema file path (workspace-relative, `strings.TrimSpace` required), `Value` = optional draft override (`2020`/`2019`/`7`/`6`/`4`, empty → 2020-12 default via `jsonschema.Compile`).
- **Evaluate:** `evaluate` `json_schema` case: read `Path` via `os.ReadFile`, `jsonschema.Compile(data, a.Value)` (draft override), `jsonschema.Validate(sch, resp.Body)` (JSON or YAML instance via `decode`). 0 violations → `Passed=true` `"json_schema %q valid"`; n violations → `Passed=false` `"json_schema %q: $.a: msg; $.b: msg"` using `Violation.String()` (`validate.go:35`). Missing path / read / compile / validate errors → `Passed=false` with `json_schema: …` message (file path always quoted, never truncated).
- **Inversion:** `Not` wrapper at `assertion.go:192` inverts `Passed` and wraps `Message` as `NOT(…)`.
- **File path:** resolved relative to CWD/workspace root; caller (CLI `reqly test`/`collection test`) passes test file's dir as base in follow-up; v1 uses process CWD (same as `internal/jsonschema` CLI).

## Data Model
`Assertion {Kind: "json_schema", Path: "schemas/user.json", Value: "7" }` — no new struct fields (reuses `Path`/`Value`), no `SchemaDraft` separate field (defer — `Value` is draft).

## API Surface
No CLI flag change; test file YAML: `{kind: json_schema, path: ./schemas/user.json}` optionally `value: "7"` for draft. Programmatic `Suite.Run(resp)` returns `Result{Passed, Message}` with violations.

## Edge Cases
Missing `Path` → `missing schema path`; unreadable file → `read "…" : …`; invalid schema → `compile "…" : …`; invalid instance (non-JSON/YAML) → `validate: …`; empty body → violations per schema (e.g. `required`); `Not` inverts.

## Testing Strategy
Existing `internal/testing` table tests + new `json_schema` cases (valid pass, type mismatch fail, missing file, compile error, `Not` inversion) via temp files (`t.TempDir()` `0644`); `go test -race ./internal/testing:1`; `internal/jsonschema` existing `validate_test.go` covers draft/synthesis.
