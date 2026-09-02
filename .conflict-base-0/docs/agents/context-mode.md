# Session Memory & Context Protection: context-mode

The `context-mode` MCP tools are the highest-priority tool surface in this repo — they keep raw bytes out of the model's context window and persist session memory across `/clear` and `/compact`. Only `ctx purge` wipes it.

## Tool priority (strict order)

1. **context-mode** (`ctx_*`) — for anything a `ctx_*` tool can do: analysis, filtering, counting, searching, multi-command gathers (`ctx_batch_execute`), web fetches (`ctx_fetch_and_index`), file analysis (`ctx_execute_file`). Think-in-Code: program the derivation, print only the answer.
2. **Harness tools** (Read/Edit/Write/Grep/Glob) — when context-mode has no suitable operation. Read/Edit stay correct for file *mutation*; `ctx_execute_file` for read-only *analysis*.
3. **Bash** — last resort. Always allowed: git/but mutations, `mkdir`/`rm`/`mv`, `go build/test/vet`, `npm`.

## Everyday commands

| Need | Command |
|------|---------|
| Resume prior work | `ctx_search(queries: [...], sort: "timeline")` — search before asking the user |
| Store a decision | `ctx_index(content, source)` |
| Check what was decided | `ctx_search(queries: ["decision"], source: "decision", sort: "timeline")` |
| Multi-command research | `ctx_batch_execute(commands, queries, concurrency: 4-8)` |
| Web docs | `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` |

## Blocked / redirected

- `curl`/`wget` and inline HTTP are intercepted → use `ctx_fetch_and_index` or `fetch()` inside `ctx_execute`.
- Shell commands with large output → `ctx_batch_execute`.
- Reading a file to analyze it → `ctx_execute_file`; Read only when editing.

## Session hygiene

- Skills, roles, and decisions persist for the whole session; don't abandon them as it grows.
- Artifacts go to files — return the path plus a one-line description, never the full content inline.
- After `/clear`/`/compact`: knowledge base is preserved. `ctx stats` shows savings; `ctx doctor` diagnoses; `ctx purge` resets (destructive).

The `ctx-*` skills mirror these commands (`/context-mode:ctx-search`, `ctx-index`, `ctx-stats`, ...).
