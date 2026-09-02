# Spec: OpenAPI Endpoint Explorer + Request Generation (Milestone 39)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q7 confirmed)
> **Scope:** Phase 1 §1.10 `ROADMAP.md` — endpoint explorer + generate requests from spec; desktop explorer panel deferred to M39b
> **Stack:** `internal/openapi` (kin-openapi, already loads/validates 3.x) + new `apps/cli/cmd/openapi.go` — no new deps
> **Reuses:** `internal/requestfile` (on-disk request format), `internal/importer/openapi.go` naming/server conventions

## Problem Statement

`reqly import openapi` imports an entire spec into a workspace. Developers often want the opposite granularity: *browse* a spec's endpoints without importing anything, then generate runnable request files for just the operations they're working on. Today they must leave the tool (Swagger UI) or import everything and delete.

## Solution

* **Explorer** — `reqly openapi explore <spec> [--tag <t>] [--json]`:
  * default: aligned table of operations — method, path, operationId (or `—`), first tag, summary
  * `--tag payments`: filter to operations carrying that tag
  * `--json`: machine-readable array `{method, path, operationId, tags, summary}` for scripts/jq
* **Generator** — `reqly openapi generate <spec> [selectors] [--output dir]` writes native request files (requestfile YAML) for selected operations:
  * selectors: repeatable `--operation <opId>`, one `--method <m> --path <p>` pair, repeatable `--tag <t>`, or `--all`; no selector → error listing available operations
  * output: `<dir>/<filename>.yaml`, default dir `./requests`; filename follows the importer convention (`operationFilename`: operationId, else method+path-derived); collisions get `-2`/`-3` suffixes + warning
  * base URL: first `servers[]` entry templated as `{{baseUrl}}`, declared in each file's `variables`
  * path/query/header params: value resolved example → schema default; unresolved required params keep their `{name}` placeholder in the URL / empty variable and are listed in a per-file summary line on stderr
  * bodies: `application/json` rendered inline as literal JSON from operation example → examples → media-type example → schema default/example; other content types are warned and omitted
  * security: `bearer` → `Authorization: Bearer {{token}}`; `apiKey` in header → header with `{{apiKey}}`; both declared in file variables. OAuth2/openIdConnect/cookie/in-query keys are skipped with a warning
  * exit non-zero only if zero files were written; warnings go to stderr and never fail the run
* **Errors** — unparseable specs, unknown operationIds, unknown methods/paths, and selector conflicts produce distinct actionable messages.

## User Stories

1. As a developer, I run `reqly openapi explore api.yaml --tag users` and see exactly which user operations exist before touching my workspace.
2. As a developer, I run `reqly openapi generate api.yaml --operation createUser --operation getUser` and get two runnable request files.
3. As a tool author, I pipe `explore --json` into jq to script against a spec.
4. As a user whose spec uses oauth2, I still get generated files with the auth parts clearly warned about rather than silently missing.

## Implementation Decisions

- Explore/generate logic lives in `internal/openapi` (`Explore`, `Generate`) so it is testable without the CLI and reusable by the desktop explorer later (M39b).
- Generation reuses `importer.openapi`'s filename + server conventions via small shared helpers where extraction is clean; no importer behavior changes otherwise.
- Generated bodies are inline literals, not template tags — deterministic, nothing to configure.
- Unresolved params stay literal (`{id}`) instead of fake values: forces conscious filling over silently sending garbage.
- Security refinement during implementation: bearer/basic/api-key-header map to **native `request.Auth` blocks** with placeholder variables (e.g. `auth: {type: bearer, config: {token: "{{token}}"}}`) instead of raw headers — same UX as grilled, but integrates masking and `reqly run` auth handling like every importer does. Basic auth added alongside bearer/api-key for free.
- Desktop explorer panel deferred (M39b): needs UI tree/table decisions.
- No ADR: additive CLI surface following `graphql introspect` conventions.

## Out of Scope

- OpenAPI 3.1 exotic features beyond what kin-openapi resolves (webhooks render as operations when present, callbacks warned)
- Editing/authoring specs (§33 P1)
- Response validation against schemas (§35 contract testing)
