# ADR 0026: Import Preserves Scripts by Translation, Not Emulation

## Status
Accepted

## Context
The Postman, Insomnia, and Bruno importers drop pre/post-request scripts with a free-text warning. The request-file format already runs scripts in a Goja sandbox exposing a small `reqly.*` global (`getVariable`/`hasVariable`/`setVariable`, `test`, request/response views). Imported scripts use each format's own API (`pm.*`, `bru.*`, `insomnia.*`) and chai-style assertion chains. The roadmap's §1.9 line "import preservation (env/auth/scripts)" requires deciding what "preserving" a foreign-API script means.

## Decision
1. **Translate the core scripting API at import time; preserve everything else verbatim as `// TODO(reqly-import): …` comments.** Variable get/set, test registration, and supported response reads map onto the existing sandbox surface through a pure translation function per dialect (postman/bruno/insomnia). Unmappable lines are never deleted: they stay in the file as greppable comments, and each becomes a structured report entry.
2. **Do not emulate assertion libraries.** `pm.expect`/chai chains are commented out, not ported. Assertion semantics are owned by the contract-testing milestone; building a chai shim now would freeze a compatibility surface ahead of that design.
3. **Degradations become a structured `ImportReport`** (item path, category, severity, message) returned by every importer parser in place of free-text warning strings. Rendering is a caller concern; CLI groups by category today, desktop renders the same structure later.
4. **Translation is one-shot at import time.** Imported files contain only Reqly-native script syntax plus TODO comments — no runtime dialect detection, no shim in the sandbox.

## Considered Options
- **Verbatim copy of source scripts** — rejected: ships dead code that fails on first send; violates the no-dummy-fallbacks rule while looking like success.
- **Full emulation (`pm.*` object graph in the sandbox)** — rejected: permanently widens the sandbox surface to match three foreign APIs and their assertion ecosystems; the maintenance tail outlives the migration value.
- **Keep dropping scripts with better warnings** — rejected: leaves §1.9 preservation unsolved and forces every migrating team to hand-port chaining logic.

## Consequences
- **Positive:** Imported collections chain out of the box for the high-frequency 80%; nothing is silently lost (file comments + report entries are dual records); one report type ends per-parser warning conventions.
- **Trade-off:** Translated scripts are best-effort — subtle semantics (Postman variable scope resolution order, dynamic `pm.` property access) can translate syntactically but behave differently; the TODO marker and report make the risk visible rather than eliminating it.
- **Follow-up:** Contract testing (P1 #35) defines the assertion story; a future pass may offer to upgrade `TODO(reqly-import)` assertion comments into native tests.
