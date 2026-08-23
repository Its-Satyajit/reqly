# 01 — Explore + Generate core (internal/openapi)

**Blocked by:** None

**Status:** done

- [x] `Explore(spec)` → ordered endpoint list {method, path, operationId, tags, summary}; tag filter helper
- [x] `Generate(spec, selectors, dir)` → written file paths + warnings; selector resolution (operation/method+path/tag/all), unknown-selector errors
- [x] Filename via operationId-else-method-path convention; `-2`/`-3` collision suffixes + warning
- [x] Base URL: first server → `{{baseUrl}}` variable per file
- [x] Params: example → default; unresolved required stay `{name}`/empty variable + summary warning; optional omitted
- [x] JSON bodies inline from example/examples/schema default-example; non-JSON content warned + omitted
- [x] Security: bearer/apiKey-header mapped to placeholder variables; oauth2/oidc/cookie/in-query warned + skipped
- [x] Table-driven tests (table-driven specs incl. $ref-heavy fixture); go vet/gofmt/go test green
