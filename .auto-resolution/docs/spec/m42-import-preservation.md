# Spec: Import Preservation & Unsupported-Feature Reporting (Milestone 42)

> **Status:** Shipped 2026-08-24 — grill settled 2026-08-23 (Q1–Q9 confirmed); tickets #257–#261 closed via PRs #265/#266/#267/#268
> **Scope:** Phase 1 §1.9 `ROADMAP.md` — "Import preservation (env/auth/scripts) + unsupported-feature reporting" (the last open Phase-1 checkbox)
> **Stack:** `internal/importer` (script translation + structured report + Postman environment writer) + `apps/cli/cmd/import.go`
> **Decision record:** [ADR 0026](../adr/0026-import-preservation-script-translation.md)

## Problem Statement

Importing a Postman, Insomnia, or Bruno collection into Reqly silently loses behavior. Pre-request and post-request scripts — often carrying auth flows, token extraction, and chaining logic — are dropped with a one-line warning. Postman collection-level variables never become an environment file, breaking parity with the Insomnia and Bruno importers. And every importer reports problems as free-text strings printed to stderr, so there is no way to group, count, filter, or render what was degraded — CLI output today is a flat wall of `warning: …` lines.

## Solution

* **Script preservation** — pre-request and post-request scripts from all three collection formats are imported into request files' existing `preRequest`/`postRequest` fields via best-effort translation of each format's scripting API onto Reqly's `reqly.*` sandbox surface. Lines that map cleanly are translated; lines that cannot be mapped are preserved verbatim as `// TODO(reqly-import): …` comments so nothing is silently lost and users can finish the port by hand.
* **Structured import report** — every degradation, skip, and translation across **all seven** importers is recorded as a typed report entry (`item path`, `category`, `severity`, `message`) instead of free text. The CLI prints a grouped summary; a future desktop import dialog renders the same structure.
* **Postman environments** — collection-level (and folder-level, merged) variables are written as `environments/<collection-name>.yaml`, matching Insomnia/Bruno parity.

## User Stories

1. As a developer migrating a Postman suite, I want my pre-request token-extraction scripts translated into `reqly.setVariable(...)` calls so my imported requests chain without hand-editing.
2. As a developer whose Postman tests use `pm.expect` chains, I want those lines preserved as commented TODOs in the imported file so I can see exactly what needs porting rather than discovering missing coverage at runtime.
3. As a developer importing a Bruno collection, I want `bru.getEnvVar("token")` calls rewritten to `reqly.getVariable("token")` so scripts run unmodified.
4. As a developer importing Postman collections with collection-level variables, I want them in `environments/<name>.yaml` so I can select and edit them with the standard environment tooling.
5. As a CI user running imports in a pipeline, I want the import summary to say how many scripts were translated, warned, or dropped per category so I can gate migration progress.
6. As a reviewer of an imported workspace, I want report entries to name the exact item path ("Orders folder > Create order") so I can navigate straight to the affected request file.
7. As a future desktop user, I want the same structured report available over the bridge so the import dialog can show degradations inline (rendering itself is out of scope here).
8. As a maintainer adding a new importer, I want a single report type to append entries to so I don't reinvent a warning convention per parser.

## Implementation Decisions

* **One new seam — script translation.** A pure function `TranslateScript(source, dialect) (translated string, entries []ReportEntry)` lives in `internal/importer`. It has no I/O and no dependency on parsing; dialects: `postman`, `bruno`, `insomnia`.
* **Mapping table (core only).** Variable access maps onto the sandbox surface: Postman `pm.environment/collectionVariables/variables/globals.get/set` → `reqly.getVariable`/`reqly.setVariable`; Bruno `bru.getEnvVar/getVar/setEnvVar/setVar` likewise; Insomnia `insomnia.environment.get` likewise; `hasVariable` used where source checks presence. Response reads that the sandbox supports map directly (`pm.response.code` → `reqly.response.status`; body/header reads → `reqly.response.body`/`.headers`). `pm.test(name, fn)` → `reqly.test(name, fn)` with the body recursively processed by the same table.
* **No assertion emulation.** `pm.expect`/chai `expect(...)` chains are not ported. Their enclosing lines become `// TODO(reqly-import): <original line>` comments plus one report entry each. Assertion semantics stay with the contract-testing milestone (§35/P1 #35).
* **Structured report.** `ImportReport{Importer string, Entries []ReportEntry}`; `ReportEntry{ItemPath, Category, Severity, Message}` with categories `auth | script | body | environment | schema | other` and severities `translated | warned | dropped`. Every parser signature changes from `(Result, []string, error)` to `(Result, *ImportReport, error)` — internal-only breakage; the CLI is the sole caller. Existing free-text warnings are re-expressed as entries with equivalent messages.
* **CLI rendering.** The import subcommands print entries grouped by category with item paths, then a one-line tally (`N translated, M warned, K dropped`). Exit code stays 0 unless parsing hard-fails — degradations never fail an import.
* **Postman environment writer.** Collection-level `variable` members merge with folder-level overrides into `environments/<collection-name>.yaml` (name sanitized like collection directories). Requests stop inlining values that now resolve through the environment layer.
* **All seven importers adopt the report type** mechanically; only Postman/Insomnia/Bruno gain new preservation behavior (cURL/HAR/WSDL have no scripts or collection variables to preserve).

## Testing Decisions

* Good tests assert external behavior: given fixture input, what lands in the workspace files and what does the report say — never internal call graphs.
* **Translation seam:** table-driven tests per dialect — each mapping row, unmappable-line preservation, nested calls inside `pm.test` bodies, empty/no-op scripts.
* **Parser tests:** extend the existing Postman/Insomnia/Bruno parser tests (prior art: vendored official example collections under `internal/importer/testdata/import-suite/`) to assert script fields in written request files, the new Postman environment file, and report contents.
* **CLI e2e:** grouped-summary rendering asserted in the style of `apps/cli/cmd/import_export_test.go`.

## Out of Scope

* Assertion/chai expectation emulation (deferred to the contract-testing milestone's assertion story)
* Desktop import dialog rendering the report (separate GUI milestone)
* Script preservation for cURL/HAR/WSDL/OpenAPI imports
* Re-translating scripts inside previously imported workspaces
* Fetching remote resources during import (file-only, unchanged)

## Further Notes

* Translation happens once at import time; imported files are plain Git-native truth afterwards — no runtime dialect detection, no compatibility shim in the sandbox.
* The `// TODO(reqly-import):` marker is a stable, greppable contract between this milestone and any future migration tooling.
* Roadmap §1.9 flips its last `[ ]` when this ships; §1.3 Digest-NTLM and OAuth 1.0 remain deferred-by-decision, not gaps.
