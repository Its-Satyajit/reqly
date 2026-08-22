# Workflow: Triage

> **Trigger:** Event — new GitHub issue or PR opened in `Its-Satyajit/reqly` (webhook or `gh issue list --search "created:>1h"` poll fallback).
> **Checkpoint:** Human-in-the-loop after full triage — agent presents a **brief** (one decision-ready summary per issue, not raw output) and waits for approve/reject before writing `ready-for-agent`.
> **Push-right:** Agent does maximal work before checkpoint — categorise, verify repro, grill if needed, draft brief — so the human is asked once, late, with everything prepared.

## Goal

Turn every new issue/PR into an agent-ready brief within one checkpoint, without asking the human for anything the agent could look up itself.

## Steps

1. **Detect** new issue/PR via `gh issue view <n> --json title,body,labels,author` (or `gh pr view`). If `NOTES.md` is thin, first interview the user about tools/channels/terminology and record in `NOTES.md`.

2. **Categorise** per `triage` skill: bug / feature / docs / chore; check `CONTEXT.md` for terminology conflicts (call out if user’s term conflicts with glossary).

3. **Verify** repro:
   - For bugs: try to reproduce via `go test -run <name>` / `npm run typecheck` / `wails3 generate` as appropriate; dispatch sub-agents for facts, don’t ask user for anything you could look up.
   - For features: check `ROADMAP.md` and `docs/features.md` for duplicate/covered scope.

4. **Grill if needed** — if spec is thin or domain language is fuzzy, run `grill-with-docs` (one question per turn, update `CONTEXT.md` + `docs/adr/` inline) until frontier is empty. Grill questions are **not** part of the brief; they are resolved before drafting.

5. **Draft brief** — per `triage` skill, write an agent-ready brief: `Issue #123 — Category: bug | Missing: repro steps | Proposed: close as dupe of #92 | Labels: ready-for-agent?` Include `gh` link and `CONTEXT.md` terms used.

6. **Present brief** — at the checkpoint, show the brief (tight, decision-ready, with link to the drafted brief file or comment draft). Do **not** post the brief or add `ready-for-agent` label yet.

7. **Await approve/reject** — human reads the brief, not the raw issue. On **approve**, agent posts the brief as a GitHub comment, adds `ready-for-agent` label via `gh issue edit <n> --add-label ready-for-agent`, and creates/updates `workflows/triage.md` spec if the grilling changed the model. On **reject**, agent records the reason in `NOTES.md` and closes the issue as `invalid`/`duplicate` via `gh issue close`.

## Brief Format

```
Issue #123 — [Category] — [One-line summary]
- Missing: [what’s missing, or “none”]
- Proposed: [ready-for-agent / needs answers Q1-Q2 / close as dupe]
- Link: https://github.com/Its-Satyajit/reqly/issues/123
```

Speed of review is imperative — the brief is <5 lines.

## Tools & Channels

- **In-workspace** via `gh` + `but` + `go`/`npm` + `context-mode` for file search; reads `CONTEXT.md`, `ROADMAP.md`, `docs/adr/`, `NOTES.md`.
- **Output:** GitHub comment + label; spec source of truth is `workflows/triage.md` (this file).

## Definition of Done

An implementer could build this workflow without asking a single question: trigger is event, checkpoint is post-triage approve, brief format is fixed, tools are `gh`/`but`, and `NOTES.md` holds the shared language. Done when `workflows/triage.md` exists and `NOTES.md` has been sharpened at least once.
