# ADR 0036: Audit Logs — Local Append-Only Trail (M69)

## Status
Accepted — grill Q1 (core + CLI + desktop, org policies deferred)

## Context
Enterprise §58.5 requires audit trail without cloud. History is per-request SQLite, not per-action.

## Decision
`internal/audit` (M69): `Entry{ID,Timestamp,Actor,Action,Resource,Details}` + `Validate` (11 actions) + `Store` (`.reqly/audit.log` JSONL 0600, mutex, sort asc) + `NewStore` (0700, 0600) + `Add`/`List`/`Clear`; CLI `reqly audit list/clear` + desktop `AuditList/Add/Clear/Export`.

## Consequences
Q1: JSONL, not SQLite — simple append, no FTS, no retention.
Q2: Org policies deferred — audit is append-only, no policy enforcement.
