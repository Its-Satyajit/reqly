# Spec: Bruno Collection Import (Milestone 36)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q5 confirmed)
> **Scope:** Phase 1 §1.9 `ROADMAP.md` — Bruno collection import (new gap)
> **Stack:** `internal/importer/bruno.go` (stdlib `encoding/json`) + `apps/cli/cmd/import.go` — no new deps
> **Predecessor:** M34/M35 importer conventions (Parse → Result → Write; shared body/auth policy)
> **Fixtures:** `internal/importer/testdata/import-suite/bruno/fixtures/` (testbench, edge cases, malformed)

## Problem Statement

Reqly imports cURL, OpenAPI, HAR, Postman, and Insomnia. Bruno is a Git-native API client whose users are the closest natural adopters of Reqly's model; its single-file JSON export should convert losslessly into a Reqly workspace.

## Solution

* **Input** — one JSON export file (`reqly import bruno <file> [--output <dir>]`). A directory path errors with guidance ("export the collection as a single JSON file"). Malformed JSON hard-errors.
* **Tree** — `items[]` recursion: `type: folder` → folder descriptors; `type: http`/`graphql` → request files. Order preserved by array position (`seq` ignored).
* **Requests** — method/url/name; headers and query params respect `enabled`; body per mode:
  * `json`/`xml`/`text`/`sparql` member string → body text + implied Content-Type (json/xml/text) unless header sets one
  * `formUrlEncoded[]` → URL-encoded wire body + CT
  * `multipartForm[]` → materialized multipart (file entries warned, re-attach after import)
  * `graphql {query, variables}` → JSON body + application/json
  * `file[]` → warning + skipped
* **Auth** — request-level `auth {mode, <mode>: {...}}`: `basic`, `bearer`, `apikey` (placement headers→in:header, queryparams→in:query), `digest` map onto Reqly's registry; `oauth2`/`awsv4`/`hawk`/`ntlm`/`wssec` warn+drop. Collection-level `root.request.auth` and `root.request.headers` become the collection descriptor's `auth:`/`headers:` — Reqly inheritance applies them natively.
* **Environments** — each environment becomes `environments/<name>.yaml`; `secret:true` variables land under `secrets:` (masked everywhere), others under `variables:`; disabled skipped.
* **Warnings** — scripts (`script`), `assertions`, `tests`, `docs`, examples are warned + skipped, matching Postman M34 policy.

## User Stories

1. As a Bruno user, I run `reqly import bruno collection.json` and get a working workspace so switching costs nothing.
2. As a user, my folders, secrets-split environments, and collection-level auth survive the import.
3. As a user importing scripted collections, I get warnings listing what was not translated.
4. As a developer, I want table-driven tests against real fixtures (testbench covers every body/auth/script shape).

## Implementation Decisions

- Public surface mirrors Postman/Insomnia: `BrunoResult{Title, Collection, Root *PostmanFolder, Auth request.Auth, Headers []request.Header, Environments []InsomniaEnvironment}`, `Write(dir)` reusing the shared folder writer + env writer.
- `apikey.placement` maps queryparams→query, headers/default→header.
- No ADR: follows the established importer pattern.
