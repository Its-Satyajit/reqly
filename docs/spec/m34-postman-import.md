# Spec: Postman Collection Import v2.1 (Milestone 34)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q5 confirmed)
> **Scope:** Phase 1 §1.9 `ROADMAP.md` — "Import: Postman"
> **Stack:** `internal/importer/postman.go` (stdlib `encoding/json`) + `apps/cli/cmd/import.go` subcommand — no new deps
> **Predecessor:** M28 HAR import (`internal/importer/har.go`, same Parse → Result → Write pattern), ADR 0005 auth scheme registry

## Problem Statement

`reqly import` supports cURL, OpenAPI, and HAR. Postman is the most common collection format in the wild; users migrating to Reqly must hand-rebuild collections or detour through Postman's cURL export. The importer pattern (parse → result → Git-native workspace writer) and the descriptor/folder model already exist.

## Solution

* **Parse** — `ParsePostman(data []byte) (*PostmanResult, []string, error)` accepts a Postman **v2.1** collection JSON (`info.schema` containing `schema.getpostman.com/json/collection/v2.1`). Tolerant of both URL forms (`url` as string or `{raw, host[], path[], query[]}`).
* **Tree** — nested `item[].item` folders are preserved 1:1 as Reqly folder descriptors (`reqly.yaml` per folder); requests become request files under their folder.
* **Variables** — collection `variable[]` (key/value/enabled) → collection descriptor `variables:` map; request-level `variable[]` → request-file `variables:`. `{{var}}` text is left untouched (identical syntax in both tools — no interpolation at import time). Disabled variables are skipped with no warning.
* **Bodies** — `raw` (+ `options.raw.language` json/xml/html/javascript → implied Content-Type when no header sets one), `urlencoded[]` → URL-encoded wire body + `application/x-www-form-urlencoded`, `formdata[]` → materialized multipart body (text fields inlined; `type: file` rows cannot be inlined and emit a re-attach warning), `graphql {query, variables}` → JSON body + `application/json`, `file` mode → warning + skipped.
* **Auth** — collection- and request-level `auth`: `basic` (username/password), `bearer` (token), `apikey` (key/value/in) map onto Reqly's `auth.config`; `noauth` maps silently to none; any other type (`oauth2`, `digest`, `hawk`, `edgegrid`, …) emits a warning and is dropped. Request-level auth overrides collection-level. Auth parameter rows accept both object and `{key, value, type}` array forms with non-string scalars.
* **Legacy v2 shorthands** — also accepted: `request` as a plain URL string (method GET), `header` as a raw newline-separated `Key: Value` string, items carrying both a request and sub-items (request imported into its own folder), requests with no URL (warned + skipped), missing/empty/whitespace methods normalized to GET.
* **Scripts** — `event[]` (prerequest/test scripts using the Postman `pm.*` API) emit warnings ("script not imported"); not translated.
* **CLI** — `reqly import postman <file.json> [--output <dir>]` writes a Git-native workspace via `PostmanResult.Write(dir)` (HAR-style): `<dir>/reqly.yaml` + `collections/<name>/…`. Default output directory = sanitized collection name. Unsupported features are reported on stderr as warnings, matching the other importers.
* **Naming** — request/folder names sanitized and deduplicated (same rules as HAR import: spaces→dashes, collapse dashes, 100-char cap).

## User Stories

1. As a Postman user, I can run `reqly import postman my-api.postman_collection.json` and get a working Reqly workspace so migration needs no manual rebuild.
2. As a user, I want folders preserved so my organization survives the import.
3. As a user, I want `{{baseUrl}}` style variables to keep working via the collection variables map.
4. As a user importing a collection with tests, I want a clear warning listing what was not imported rather than silent loss.
5. As a developer, I want the importer table-driven tested against fixture collections (string URLs, object URLs, all body modes, auth variants).

## Implementation Decisions

- One new file `internal/importer/postman.go` mirroring `har.go`; mirror structs private, `PostmanResult` public with `Write(dir string) error`.
- Folder recursion in `Write`: each folder gets `<dir>/<sanitized-name>/reqly.yaml`; requests written alongside; filename dedupe within a folder scope.
- `--output` default: sanitized info.name in the CWD (consistent with `import openapi` behavior of writing a workspace dir).
- Warnings returned by Parse AND accumulated during tree walk (auth/scripts/file-body/unsupported schema version).
- No ADR: follows the established importer pattern; nothing hard-to-reverse.
