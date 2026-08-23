# 01 — JUnit + JSON test reports (internal/runner + CLI)

**Blocked by:** None

**Status:** done

- [x] runner.Report JSON serialization round-trips
- [x] JUnit XML: testsuite per run, testcase per step, failure elements carry request/assertion messages, timing attributes
- [x] CLI flags `--report junit <file>` / `--report json <file>` on collection test (both combinable); report write failure warns, never changes exit code
- [x] Table-driven tests incl. malformed-input and empty-run cases
- [x] go vet/gofmt/go test green
