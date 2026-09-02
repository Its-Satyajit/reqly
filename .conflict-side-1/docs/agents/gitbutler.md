# Version Control: GitButler (`but`)

All version-control **write** operations go through GitButler's `but` CLI, not raw git. Load the `but` skill before any VCS operation (commit/push/branch/PR/history edits).

## Rules

- Never run `git add`/`commit`/`push`/`checkout`/`merge`/`rebase`/`stash`. Translate to `but`.
  - One permitted exception: `git add -- <path>` to mark a conflicted uncommitted file resolved.
- Read-only git inspection (`git log`, `git show`, `git blame`) is fine.
- IDs are copied exactly from `but status` / `but diff` output — never invented or hardcoded.
- Branches marked `(merged upstream)` have landed; `but pull` cleans them up.
- Push a named branch only: `but push <branch>` (bare push pushes everything).
- Stacked PRs on GitHub must merge via the async REST API
  (`PUT /repos/:owner/:repo/pulls/<n>/merge-async`, then poll the returned URL);
  plain GraphQL merges 403 for stacks. The status endpoint may lag — poll the PR state itself.
- `but land` bypasses PRs/CI/branch protection; do not use it when an open PR exists.

## Quick map

| Task | Command |
|------|---------|
| Workspace overview | `but status` (add `-fv` for file/hunk IDs) |
| Uncommitted changes | `but diff` |
| Commit selected files/hunks | `but commit -b <branch> -m "<msg>" <id> <id>` |
| Amend | `but amend -t <commit-or-branch> <id>...` |
| Reorder / split history | `but move`, `but squash`, `but uncommit` |
| Update from target | `but pull` |
| Push | `but push <branch>` |
| Create PR | `but pr new <branch-id> -m "Title"` |
| Resolve conflicts | `but resolve <commit>` → edit files → `but resolve finish` |

Full syntax lives in the `but` skill (`~/.config/opencode/skills/gitbutler/`).

## Repo conventions

- Commit messages follow `type(scope): summary`.
- Docs/quarantined files are committed to the `docs/internal` branch, never to source branches.
- Dedicated GitButler branch per agent session, named `<name>/<short-description>`.
- Checkpoint-commit after each working turn; tidy unpublished local history on request.
