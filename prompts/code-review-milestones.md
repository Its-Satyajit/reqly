# Code Review per Milestone — LLM Prompt (two-axis `/code-review`)

> **Use:** Copy this prompt verbatim into any LLM (Muse, ChatGPT, Gemini, Cursor) to get a `/code-review`-style two-axis review for **one milestone at a time**. The skill runs two parallel sub-agents — **Standards** and **Spec** — against `git diff <fixed-point>...HEAD` and aggregates without reranking.

---

## ⚠️ Where the docs live — `docs/internal` is a quarantined branch

**On `main` you only see** `README.md` + `ROADMAP.md` + `LICENSE` + `Milestones/01-12` + source. **Everything else is quarantined on the `docs/internal` branch — never merged to `main` or PRs.** That includes:

- `AGENTS.md`, `GEMINI.md`, `.cursorrules`, `CONTEXT.md` (domain glossary), `NOTES.md`, `DESIGN.md`, `CHANGELOG.md`, `.github/copilot-instructions.md`
- `docs/adr/0001-0045` (45 ADRs), `docs/spec/*.md` (per-feature specs like `m33-cancel-in-flight-request.md`, `m34-postman-import.md` …), `docs/agents/issue-tracker.md` + `domain.md` + `gitbutler.md` + `triage-labels.md`, `docs/features.md`, `docs/reference/*.md`, `docs/internal/gui-roadmap.md`
- `.scratch/<feature>/` ticket markdown (historical ticket archive, 30+ features)

If you need any of those while on `main`, **do not checkout `docs/internal`** (it will detach workspace under GitButler). Use read-only access:

```bash
# fetch + list what's on that branch
git fetch origin docs/internal
git ls-tree -r --name-only docs/internal -- docs/ | sort
git ls-tree -r --name-only docs/internal | grep "^\.scratch" | head -n 50

# read a single file without checking out
git show docs/internal:CONTEXT.md | head -n 300
git show docs/internal:AGENTS.md | head -n 300
git show docs/internal:docs/adr/0033-workflow-engine.md
git show docs/internal:docs/spec/m65-workflow-engine.md
git show docs/internal:docs/agents/issue-tracker.md

# or mount it as a disposable worktree (cleanest for many reads)
git worktree add /tmp/docs-internal docs/internal
ls /tmp/docs-internal/docs/adr | head -n 20
cat /tmp/docs-internal/CONTEXT.md
git worktree remove /tmp/docs-internal
```

Rules from `AGENTS.md:1` (`docs/internal:AGENTS.md`): **Never merge `docs/internal`**. If a doc shows up in a diff bound for `main`, move it to `docs/internal` first. When you need to edit a quarantined file, checkout `docs/internal`, commit there, push that branch only.

---

## Inputs you must pin before starting

1. **Milestone under review** — one of:
   - `Milestones/01-phase-0-foundation.md` (Phase 0, 26 boxes) — P0 skeleton
   - `Milestones/02-phase-1-core-api-client.md` (100 boxes) — P0 §1.1-1.13 (request engine, auth, gRPC/SOAP P3 splits)
   - `Milestones/03-phase-2-differentiating-features.md` (28 boxes) — P1 §56.1-56.8 + M28-M37
   - `Milestones/04-phase-3-power-user-features.md` (15 boxes) — P2 §57 + M60-M66
   - `Milestones/05-phase-4-ecosystem-and-enterprise.md` (10 boxes) — P3 §58 + M67-M75
   - `Milestones/06-phase-5-mcp-ai-extensibility.md` (45 boxes) — P4/P5 §59 + docs/quality gates
   - Or `Milestones/07-historical-milestones-ledger.md` … `12-traceability-map.md` for ledger/UI-architecture
   - Corresponding `ROADMAP.md` phase section (canonical ledger) — `ROADMAP.md:26-280`

2. **Fixed point** — whatever the human said (`main`, a SHA, `HEAD~5`, tag). If they didn't say, **ask**. Default for Reqly is `main`.
   ```bash
   git rev-parse --verify <fixed-point>   # must succeed
   git diff <fixed-point>...HEAD --stat   # three-dot, merge-base — must be non-empty; fail here, not inside sub-agents
   git log <fixed-point>..HEAD --oneline  # capture commit list
   ```

3. **HEAD** — usually `git rev-parse HEAD` on `gitbutler/workspace` (GitButler integration branch, see `but status`). The diff you review is the applied virtual branches at `but diff`.

---

## Process (mirror `/code-review` skill — 5 steps)

### 1) Pin the fixed point
Capture `git diff <fixed-point>...HEAD --stat` + `git log <fixed-point>..HEAD --oneline` once. Verify `git rev-parse <fixed-point>` resolves and diff is non-empty.

### 2) Identify the spec source (in order)
1. Issue refs in commits (`#123`, `Closes #45`, `!67`) via `docs/internal:docs/agents/issue-tracker.md` workflow (`gh issue view <n> --comments`).
2. Path the human passed.
3. Spec file under `Milestones/`, `docs/spec/` (on `docs/internal`), or `.scratch/<feature>/` (on `docs/internal`) matching branch/feature — e.g. `M65` → `docs/spec/m65-workflow-engine.md` on `docs/internal`.
4. If nothing, ask human where spec is; if none, Spec sub-agent reports "no spec available".

For **all app code vs milestone** (the usual Reqly ask), the spec is the milestone file itself + its `ROADMAP.md` phase + `CONTEXT.md` + relevant `docs/adr/*` + `docs/spec/*` — all but the milestone are on `docs/internal`. Example map:
- M33 cancel-in-flight → `internal/request` + `docs/spec/m33-cancel-in-flight-request.md` (docs/internal) + ADR 0010
- M65 workflow → `internal/workflow` + `docs/spec/m65-workflow-engine.md` + ADR 0033 + CONTEXT `Workflow`/`Condition`
- M67 theme → `internal/theme` + `docs/spec/m67-theme-sharing.md` + ADR 0035

### 3) Identify the standards sources
- `oxlint.config.ts:22-38` — 13 `anti-slop/*: error` (see file), `require-safety-comment-for-type-assertion` (every `as` needs `// SAFETY:`, `as const` exempt), `allowInTypeGuards:true` for `no-runtime-typeof`, `ignorePatterns` `**/bindings/**` etc.
- `Makefile` + `.github/workflows/ci.yml` — `gofmt -l internal apps/cli` (empty), `go vet ./internal/... ./apps/cli/...` (pass), `go test -race ./...`, `nub run typecheck` (tsc --noEmit), `oxlint` pass.
- No `CODING_STANDARDS.md`/`CONTRIBUTING.md` — the above are authoritative.

**On top of that, always carry the smell baseline** (Fowler ch.3) — 12 heuristics, never hard violations, repo overrides, skip tooling-enforced:

- **Mysterious Name** — name doesn't reveal intent → rename; if no honest name, design's murky.
- **Duplicated Code** — same shape in >1 hunk/file → extract.
- **Feature Envy** — method uses another object's data more than own → move method.
- **Data Clumps** — same few fields/params travel together → bundle into type.
- **Primitive Obsession** — primitive/string for domain concept → small type.
- **Repeated Switches** — same switch/if-cascade on same type recurs → polymorphism or one map.
- **Shotgun Surgery** — one logical change scatters across many files → gather into one module.
- **Divergent Change** — one file edited for several unrelated reasons → split.
- **Speculative Generality** — abstraction/params for needs spec doesn't have → delete/inline.
- **Message Chains** — long `a.b().c().d()` → hide behind one method on first object.
- **Middle Man** — class/function mostly delegates → cut, call target direct.
- **Refused Bequest** — subclass ignores most of inheritance → composition.

### 4) Spawn both sub-agents in parallel (do not pollute each other's context)

**Standards sub-agent prompt** — include:
- Full diff command + commit list from step 1.
- List of standards-source files from step 3 **plus the 12-item smell baseline verbatim** (sub-agent has no other access).
- Brief: "Report, per file/hunk where relevant, (a) every place the diff violates a documented standard: cite the standard (file + rule); and (b) any baseline smell you spot: name it and quote the hunk. Distinguish hard violations from judgement calls: documented breaches can be hard, baseline always judgement, documented repo standard overrides baseline. Skip anything tooling enforces. Under 400 words. Include `file_path:line_number`."

**Spec sub-agent prompt** — include:
- Diff command + commit list.
- Path or fetched contents of the spec (milestone markdown + relevant ADR/spec excerpt — fetch from `docs/internal` via `git show docs/internal:<path>`).
- Brief: "Report: (a) requirements spec asked for that are missing/partial; (b) behaviour in diff not asked for (scope creep); (c) requirements that look implemented but where implementation looks wrong. Quote the spec line for each finding. Under 400 words. Include `file_path:line_number`."

If spec missing, skip Spec sub-agent and note "no spec available".

### 5) Aggregate
Present under `## Standards` and `## Spec` headings, verbatim or lightly cleaned — **do not merge or rerank** (two axes deliberately separate). End with one-line summary: total findings per axis + worst issue **within each axis** (if any). Don't pick a single winner across axes.

**Why two axes:** Standards-pass/Spec-fail (clean code, wrong feature) and Spec-pass/Standards-fail (correct feature, messy code) must not mask each other.

---

## Output contract for one milestone

```md
## Standards — <milestone> `main`→HEAD (`git diff main...HEAD`)
Hard violations: N / Judgement smells: M
- `file:line` — rule — hunk quote — hard/judgement
...

## Spec — <milestone> `main`→HEAD
(a) Missing/partial:
- `Milestones/0X:line` "spec line" → `file:line` status — quote spec line
(b) Scope creep:
- ...
(c) Looks implemented but wrong:
- ...

Summary: Standards 0 hard + M judgement (worst: `file:line` — smell), Spec A missing + B creep + C wrong (worst: `file:line` — spec line).
```

Then **update the Code Review Gate** so the human can tick `[x]`:

- In the milestone file (`Milestones/0X`) — `## Code Review Gate (/code-review — two-axis)` — tick both `[ ]` → `[x]` when both axes green (or document deferral as `[~]` with note).
- In `ROADMAP.md:308` `## Code Review Gates` — tick the phase row for that milestone.

Commit via GitButler (see `docs/internal:docs/agents/gitbutler.md` + `but` skill):
```bash
but diff                          # inspect hunks, note IDs (e.g. pu, pm)
but commit -b <branch> -m "chore(docs): tick code-review gate <milestone>" <id> <id>
but status                        # verify
```

---

## Milestone → code map (use to scope the diff)

| Milestone | ROADMAP § | Key code on `main` | Spec on `docs/internal` |
|-----------|-----------|--------------------|-------------------------|
| 01 Foundation | §0 | `go.mod`, `frontend/`, `apps/desktop/backend`, `internal/history/db` | `docs/adr/0001`, `0002` |
| 02 Phase 1 P0 | §1 | `internal/request`, `auth` (basic/bearer/jwt/digest/oauth1/oauth2/aws/edgegrid/custom), `variables`, `history`, `core`, `requestfile`, `collections` | `docs/spec/m20`–`m26`, `m28`, ADRs 0005-0021 |
| 03 Phase 2 P1 | §56 | `lib/specTree`, `lib/schemaGraph`, `internal/theme`, `mocking`, `openapi`, `importer` (postman/insomnia/bruno) | `docs/spec/m34`–`m40`, ADRs 0035-0036 |
| 04 Phase 3 P2 | §57 | `internal/monitor`, `perf`, `mqtt`, `socketio`, `automation`, `collab` | `docs/spec/m59`–`m64`, ADRs 0040-0043 |
| 05 Phase 4 P3 | §58 | `internal/audit`, `policy`, `rbac`, `secrets.VaultStore`, `sso`, `scim`, `collab.Server` | `docs/spec/m67`–`m75`, ADRs 0035-0042 |
| 06 Phase 5 | §59 | `internal/mcp`, `internal/ai` | `docs/spec/m60`–`m64` |
| 09-11 UI Architecture | §2-§4, §56-§63 | `frontend/src/components/shell`, `features/*` | `docs/Reqly Complete UI Architecture...` + `docs/internal/gui-roadmap.md` |

---

## Verification gates (must be green before ticking `[x]`)

| Gate | Command | Pass |
|------|---------|------|
| Go tests | `go test ./...` | 0, no skips |
| Go race | `go test -race ./...` | 0 |
| Go vet | `go vet ./...` | 0 |
| Gofmt | `gofmt -l internal apps/cli` | empty |
| CLI build | `go build -o reqly ./apps/cli` | binary |
| Frontend typecheck | `nub run typecheck` | 0 |
| Lint | `npx oxlint` or `nub run lint` | 0 (warnings from 4 relaxed JSON-boundary rules expected) |
| Vitest | `nub run --filter @reqly/frontend test` | 20 files / 185 tests |
| React doctor | `nubx react-doctor@latest` | 0/0/0 score 100 |

---

## Example invocation (copy-paste)

> **Human:** "Review `Milestones/02-phase-1-core-api-client.md` (P0) — fixed point `main`, HEAD is current workspace. Spec is that milestone file + ROADMAP Phase 1 + ADRs 0005-0021 + CONTEXT on `docs/internal`."
>
> **LLM does:** pin `main` (`git diff main...HEAD --stat` + `git log main..HEAD --oneline`), fetch `git show docs/internal:CONTEXT.md` + `docs/internal:docs/adr/0005-git-native-auth-schemes.md` + `docs/spec/m33-cancel-in-flight-request.md`, spawn Standards + Spec sub-agents with prompts above, aggregate under `## Standards` / `## Spec`, one-line summary, then propose `but diff` → `but commit -b fix/review-round3` to tick the gate in `Milestones/02:158` + `ROADMAP.md:313`.

---

## Quick checklist before you start an LLM review

- [ ] `git rev-parse --verify <fixed-point>` succeeds, `git diff <fixed-point>...HEAD --stat` non-empty (fail here, not inside sub-agents)
- [ ] You have fetched `docs/internal` and can `git show docs/internal:CONTEXT.md` and `docs/internal:docs/agents/issue-tracker.md`
- [ ] You have read the milestone markdown on `main` (`Milestones/0X`) and its `ROADMAP.md` phase
- [ ] You have the smell baseline verbatim in the Standards prompt
- [ ] You will keep the two axes separate in the final report
