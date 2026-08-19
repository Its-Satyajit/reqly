# ADR 0010: Desktop Collection Runs Stream Per-Step Results and Read from Disk

## Status
Accepted

## Context
The desktop needs a collection-run surface (M18). The runner engine (`internal/runner`) already executes collections sequentially with variable chaining, pre/post scripts, and `reqly.test()` assertions; the only question is how the desktop surfaces a run. Two axes need deciding: (1) how results reach the frontend, and (2) what the run executes — the open tab snapshots or the on-disk request files.

## Decision
- **Streaming:** results are delivered as Wails events, one per completed step (`reqly.run.<id>.step`) followed by a final `reqly.run.<id>.done` event carrying the complete report. The engine exposes a single new seam — `runner.Options.OnStep func(StepResult)` — invoked inside `RunCollection` after each step; a core `CollectionRunService` converts each raw `StepResult` into a JSON-safe DTO (errors as strings, response bodies base64, auth/credential values never serialized) and the desktop bridge forwards them as events. Cancel flows through the existing `ctx`; the desktop enforces single-flight (one run at a time, server- and client-side).
- **Fresh from disk:** a run reads the request files from disk at run time, exactly like the CLI. Unsaved **Request Tab** drafts are never part of a run.

## Considered Options
- **Snapshot from open tabs** — rejected: a run is a statement about the saved workspace (like the CLI), not about the editor session; unsaved drafts are ambiguous and the run must be reproducible from Git.
- **Long-polling / single final result** — rejected: defeats live progress; streaming events are the Wails-idiomatic path.
- **Per-run environment selector** — deferred: M18 uses the active environment, matching how Request Tabs resolve their env pill; a selector is a natural follow-up.

## Consequences
- **Positive:** one engine change serves CLI and desktop; live progress without polling; reproducible, Git-native runs; single-flight keeps run state unambiguous; cancel is safe (current step may finish, no further steps scheduled).
- **Trade-off:** a run ignores unsaved edits (visible hint shown in the Run View); one run at a time; per-run environment choice is not yet possible.