# ADR 0044: Integration Tests — Core ↔ Persistence ↔ Engine (M77)

## Status
Accepted — grill Q1 (tests)

## Context
Milestones/06 Quality requires integration tests for core ↔ persistence ↔ engine, but only unit tests existed.

## Decision
`internal/integration/pipeline_test.go` (M77): `TestPipeline_WorkflowWithPolicyRBACAuditCollab` (httptest login→profile, `policy.Validate/Enforce/Save` 0600, `rbac` admin, `collab` add/save/load, `workflow` Execute Extract + Interpolation, `audit` JSONL 0600) + `Milestones/06` Progress Tracker Quality ~95% → ~98%.

## Consequences
Q1: Single pipeline — not full E2E (Playwright deferred).
Q2: No performance — not load testing.
