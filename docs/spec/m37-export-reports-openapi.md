# Spec: Exports — Test Reports + OpenAPI (Milestone 37)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q4 confirmed)
> **Scope:** Phase 1 §1.9 `ROADMAP.md` — "Export: requests, OpenAPI, responses, test results"
> **Stack:** `internal/runner` (report serialization) + `internal/exporter/openapi.go` + CLI flags — no new deps beyond stdlib for XML

## Problem Statement

§1.9 lists four export gaps. Two are already satisfied by shipped features: *responses* via HAR export (`reqly export har`, M28) and desktop response download; *requests* via workspace export (`reqly export workspace`, M25) and code generation (`reqly export code`, M24). The genuine gaps are **machine-readable test results** and **OpenAPI generation**.

## Solution

### T1 — Test reports (`reqly collection test --report <fmt> <file>`)

* Flags: `--report junit <path>` and/or `--report json <path>`; both may be combined.
* **JSON report**: full `runner.Report` serialization (steps, per-test outcomes, logs, timing).
* **JUnit XML**: `<testsuite name="collection" time=... >` with `<testcase>` per step (classname = folder path, failures from failed assertions/request errors with message bodies). Skipped steps omitted.
* Report writing is best-effort after the run: exit code still reflects run success/failure.

### T2 — OpenAPI export (`reqly export openapi [src] [--out <file>]`)

* New `internal/exporter/openapi.go`: builds an OpenAPI **3.0** YAML document from a collection/workspace:
  * paths+operations from request files (grouped by path); operationId from request names
  * parameters from query params + headers (non-standard headers only)
  * requestBody Content-Type from the implied header; schema left as empty generic object (documented limitation: no invented schemas)
  * security schemes: basic/bearer/apikey from collection/request auth; security applied per-operation or at root when uniform
  * servers from the collection base URL when present; variables as server defaults
* CLI prints the spec to stdout by default, writes a file with `--out`.

## User Stories

1. As a CI user, I run `reqly collection test --report junit out.xml` so my CI dashboard shows per-request results.
2. As an API producer, I keep my requests as source of truth and generate an OpenAPI doc on demand.
3. As a user, combining both reports in one run never changes the run's exit code.
4. As a developer, I want golden-file tests for generated OpenAPI output (same convention as docs/code exporters).

## Implementation Decisions

- JUnit writer lives in `internal/runner/report.go`; JSON uses encoding/json directly on Report.
- OpenAPI builder takes `[]requestfile.File` + collection config; reuses existing inheritance resolution.
- No ADR: additive exporters following established patterns.
