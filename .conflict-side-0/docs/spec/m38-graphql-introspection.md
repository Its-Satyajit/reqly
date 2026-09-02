# Spec: GraphQL Introspection CLI (Milestone 38)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q3 confirmed)
> **Scope:** Phase 1 §1.8 `ROADMAP.md` — GraphQL introspection; desktop schema browser deferred to M38b
> **Stack:** `internal/graphql` (stdlib net/http + encoding/json) + new `apps/cli/cmd/graphql.go` — no new deps
> **Predecessor:** ADR 0013 GraphQL body support

## Problem Statement

Reqly can send GraphQL requests but cannot inspect a schema. Developers must leave the tool to explore types or craft queries blind. The package doc already promises "introspection, schema browsing, query validation" — this milestone delivers introspection via CLI.

## Solution

* **Client** — `graphql.Introspect(ctx, endpoint string, opts IntrospectOptions)` POSTs the standard full introspection query as `{query}` JSON (application/json). Options: repeatable headers (auth), timeout. GraphQL `errors[]` in a 200 response are returned as a typed error listing messages.
* **Model** — parsed `Schema`: types with kind/name/description/fields; fields carry argument lists and wrapped type references rendered GraphQL-style (`[String!]!`, `ID`, `[Episode!]`).
* **CLI** — `reqly graphql introspect <url> [--header "k: v"]... [--type <Name>] [--json]`:
  * default: text summary — Query/Mutation/Subscription root fields first (`field(arg: T): R`), remaining object types alphabetically; enums show values inline; scalars/input/unions listed by kind
  * `--type User`: renders only that type's fields
  * `--json`: pretty-printed raw introspection result for downstream tools
* **Errors** — non-2xx transport errors, malformed JSON, and GraphQL-level errors each produce distinct actionable messages.

## User Stories

1. As a developer, I run `reqly graphql introspect https://api.test/graphql --header "Authorization: Bearer t"` and read the schema without leaving the terminal.
2. As a developer, I use `--type User` to focus on one type while writing a query.
3. As a tool author, I pass `--json` output into jq/scripts.
4. As a user hitting an unauthorized endpoint, I see the GraphQL error messages, not a silent empty schema.

## Implementation Decisions

- Standard introspection meta-query (not the deprecated partial one) so all servers respond.
- Text renderer lives in internal/graphql (testable without CLI); CLI only wires flags/stdout.
- Desktop schema browser deferred (M38b): needs UI tree component decisions.
- No ADR: additive, follows ws/sse command conventions.
