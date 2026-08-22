# Workflow: Renovate

> **Trigger:** Event — new Renovate PR opened (`renovate/*` branch, e.g. #157 `golang-1.26-trixie`, #158 `alpine`, #159 `golang-1.x`, #177 `wails/v3`, #178 `goja`) with weekly schedule fallback (every Monday `gh pr list --label renovate` + `gh issue view 102` scan).
> **Checkpoint:** Human-in-the-loop after rebase + CI — agent presents a **brief** per PR and waits for approve before merging.
> **Push-right:** Agent does maximal work before checkpoint — rebases, waits for CI, briefs — so the human is asked once, late, with everything prepared.
> **Scope:** Renovate 5 only; `release-please` #142 stays manual.

## Goal

Keep dependency PRs green and merged within one checkpoint per PR, without asking the human to rebase or check CI.

## Steps

1. **Detect** new Renovate PR via `gh pr view <n> --json number,title,headRefOid,mergeable,mergeStateStatus` or weekly scan `gh pr list --search "label:renovate"` + `gh issue view 102 --json body`.

2. **Rebase** — `gh api -X PUT /repos/Its-Satyajit/reqly/pulls/<n>/update-branch -f expected_head_sha=<head>` (merge `main` `82a4116` into PR branch). This brings the `ci.yml:67` `-d ../frontend/bindings` + `bridge.ts` `noImplicitAny` fixes. If API returns `no new commits`, close as superseded: `gh pr close <n> --comment "Superseded by main 82a4116, Renovate will recreate"` and stop.

3. **Wait for CI** — poll `gh pr checks <n>` until `Frontend` + `Go core` are `pass` (or `UNSTABLE` → `CLEAN`). Timeout after 20 min → brief notes `CI still UNSTABLE`.

4. **Draft brief** — per-PR one-liner: `PR #177 wails/v3 beta.9→beta.11 — Go pass, Frontend pass after rebase — ready to merge` with link `https://github.com/Its-Satyajit/reqly/pull/177`.

5. **Present brief** — at the checkpoint, show the brief (tight, decision-ready, with link). Do **not** merge yet.

6. **Await approve** — human reads the brief, not the raw PR. On **approve**, `gh pr merge <n> --merge` (or `but land` if the PR is part of a GitButler stack and `gh pr merge --auto` says `must be merged using asynchronous merge REST API`, then use `but land <branch> --yes` as fallback). On **reject**, close with reason: `gh pr close <n> --comment "Rejected: ..."` and record in `NOTES.md`.

## Brief Format

```
PR #177 — [dep] — Go pass, Frontend pass after rebase — ready to merge
Link: https://github.com/Its-Satyajit/reqly/pull/177
```

Speed of review is imperative — <2 lines per PR.

## Tools & Channels

- **In-workspace** via `gh` + `but` + `gh pr checks`; reads `NOTES.md` for terminology (e.g. “Renovate 5” = #157, #158, #159, #177, #178).
- **Output:** GitHub merge (or close) + `NOTES.md` update.

## Definition of Done

Implementer could build without asking a question: trigger is event + weekly schedule, checkpoint is post-rebase/CI brief, brief format is fixed, tools are `gh` + `but`, and `NOTES.md` holds “Renovate 5” language. Done when `workflows/renovate.md` exists and `NOTES.md` has been sharpened.
