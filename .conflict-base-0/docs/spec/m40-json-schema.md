# Spec: JSON Schema Validate / Inspect / Generate (Milestone 40)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q9 confirmed)
> **Scope:** Phase 1 §1.10 `ROADMAP.md` — JSON Schema validate, inspect, generate; schema *edit* deferred to M40b; test-assertion hook deferred to §35 contract testing
> **Stack:** new `internal/jsonschema` (santhosh-tekuri/jsonschema/v6 promoted from indirect) + new `apps/cli/cmd/schema.go` — no new external deps

## Problem Statement

Reqly validates OpenAPI documents but not the JSON Schemas inside or beside them. Developers must leave the tool to check a payload against a schema, understand a schema's shape, or produce a sample payload for a request. The mock server hand-rolls example synthesis with no constraint awareness.

## Solution

* **Validation** — `reqly schema validate <schema> [instance]`:
  * schema and instance accept JSON or YAML; instance omitted or `-` reads stdin
  * one line per violation: `$.items[2].price: got number, want string`; final count line
  * `--json`: machine-readable violations array `{path, message}`
  * `--draft <n>` overrides `$schema` detection (2020-12 default when absent)
  * exit 1 when violations exist; distinct errors for unreadable/uncompilable schema and unparseable instance
* **Inspection** — `reqly schema inspect <schema> [--json]`:
  * tree walk in the `graphql introspect` text style: each node shows type, required marker (`!`), inline constraints summary (`string enum:a|b maxLength:10`), nested under properties/items/additionalProperties
  * `$ref`s resolved with the target path shown once; cycles safe
  * `--json` dumps the keyword map
* **Instance generation** — `reqly schema generate <schema> [--seed n] [--optional]`:
  * deterministic by default — same schema, same document; `--seed <n>` varies choices
  * value precedence per node: `const` > first `enum` > `example` > `default` > synthesized
  * numbers honor minimum/maximum/multipleOf; strings satisfy minLength/maxLength; arrays emit max(minItems,1) generated items
  * known formats realistic: email, date-time, uri, uuid, ipv4, hostname; unknown formats fall back to plain synthesis
  * `pattern` with no other hint → generic `"string"` + warning; recursive `$ref`s capped at depth 8 + warning; `allOf` merged shallowly, `anyOf`/`oneOf` first branch chosen, `not` ignored — all warned, never fatal
  * optional properties omitted unless `--optional`
  * warnings on stderr like M39; never affect exit code when an instance was produced

## User Stories

1. As a developer, I pipe `curl -s api.test/users | reqly schema validate user.schema.json` and see exactly which fields violate.
2. As a spec reader, I run `reqly schema inspect pet.schema.yaml` instead of decoding YAML by eye.
3. As a tester, I generate a seed-varied sample payload to send with `reqly run`.
4. As a tool author, I script against `validate --json` violations.

## Implementation Decisions

- All logic in `internal/jsonschema` (`Compile`, `Validate`, `Inspect`, `Generate`) so the §35 contract-testing milestone can call it from the runner without touching the CLI.
- santhosh-tekuri/jsonschema/v6 is already in the dependency graph via kin-openapi; promoting it adds no new dependency.
- Deterministic generation keeps output testable and diff-friendly; randomness only behind `--seed`.
- No ADR: additive CLI surface following `graphql introspect` / `openapi explore` conventions.

## Out of Scope

- Schema editing UI (desktop, M40b)
- `AssertJSONSchema` assertion kind for test files (§35 contract testing)
- Schema inference from samples
- XML/XSD validation (separate roadmap line)
