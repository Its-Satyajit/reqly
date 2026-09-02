# ADR 0045: Coverage — Variables 96.2% (M78)

## Status
Accepted — grill Q1 (tests)

## Context
`internal/variables` coverage 55.8% due to uncovered `Get`/`Range`/`Clone`/`UnknownDynamicTags`/`Generate`.

## Decision
`internal/variables/coverage_test.go` (M78): `TestCoverage_GetRangeClone` + `TestCoverage_UnknownDynamicTagsAndGenerate` (defaultTagGenerator 6 tags, `SetTagGeneratorForTest` fixed) → `variables` 55.8% → 96.2%, `go test -cover` avg ~80%+, `Milestones/06` Coverage within targets `[~]` → `[x]`.

## Consequences
Q1: Exporter 57.9% and docs 69% still low — deferred.
Q2: Thresholds tracked per PR, not enforced.
