# Spec: Insomnia v4/v5 Import (Milestone 35)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q5 confirmed)
> **Scope:** Phase 1 §1.9 `ROADMAP.md` — "Import: Insomnia"
> **Stack:** `internal/importer/insomnia.go` (stdlib `encoding/json` + `gopkg.in/yaml.v3`) + `apps/cli/cmd/import.go` — no new deps
> **Predecessor:** M34 Postman import (same Parse → Result → Write pattern, same body/auth conventions)

## Problem Statement

`reqly import` covers cURL, OpenAPI, HAR, and Postman. Insomnia is the other major desktop API client; its exports are JSON (v4) or YAML (v5). Users migrating from Insomnia need a one-command path into a Git-native Reqly workspace.

## Solution

* **Format sniffing** — `ParseInsomnia(data []byte)` accepts:
  * **v4**: JSON with `__export_format: 4` and a flat `resources[]` list linked by `parentId` (workspace → request_group tree → request; environment; cookie_jar).
  * **v5**: YAML with `type: collection.insomnia.rest/5.0`, a hierarchical `collection[]` (`children`), and an `environments` block.
  Structural errors (malformed JSON/YAML, missing `collection` in v5, missing resources in v4) hard-error; version mismatches warn but parse.

* **Tree** — v4 `request_group` hierarchy and v5 `children` nesting both map to Reqly folder descriptors, preserving order (`metaSortKey` ignored for MVP).

* **Requests** — method/url/name/headers (`name`→key, `disabled` respected), query `parameters`, description preserved as request name comment? No: dropped silently (Reqly files have no description field). Bodies per M34 conventions: mimeType → implied Content-Type, `urlencoded` params encoded, `form-data` materialized to multipart with file rows warned, file bodies warned+skipped.

* **Auth** — `basic`, `bearer`, `apikey` (location header/query → `in`), `digest` (username/password/realm/nonce...) map onto Reqly's registry; `oauth2`, `hawk`, `ntlm`, `asap`, `oauth1`, others → warning + dropped. Workspace-level default auth applies where the request has none.

* **Environments** — each Insomnia environment becomes `environments/<name>.yaml` in the imported workspace (Reqly-native format: variables map). Nested data values are flattened to dotted keys ("user.name") with one warning per environment. Cookie jars are dropped silently. Private environments are imported like any other (flag not carried).

* **CLI** — `reqly import insomnia <file> [--output <dir>]`; default output = sanitized collection/workspace name. Warnings on stderr, success line on stdout, matching M34.

* **Fixtures** — the community suite fixtures already vendored under `testdata/import-suite/insomnia/fixtures/` drive table-driven tests (v4, v4-with-envs, v5, v5-with-envs, v5-dates, malformed, invalid-missing-collection).

## User Stories

1. As an Insomnia user, I run `reqly import insomnia export.json` (or `.yaml`) and get a working workspace so migration needs no manual rebuild.
2. As a user, my folder structure survives both formats.
3. As a user, my environments arrive as native Reqly environment files so `reqly env use` works immediately.
4. As a user importing collections with OAuth2/HAWK auth, I get clear warnings rather than silent loss.
5. As a developer, I want malformed exports to fail loudly with actionable errors.

## Implementation Decisions

- One new file `internal/importer/insomnia.go`; public surface mirrors Postman: `InsomniaResult{Title, Requests..., Environments}` with `Write(dir string) error`.
- v4 parentId resolution: two-pass (index by `_id`, then attach); orphan requests whose parent chain misses the workspace root go to collection root.
- `__export_format` ≠ 4 or `type` ≠ `/5.0` → warning, best-effort parse continues.
- No ADR: follows the established importer pattern.
