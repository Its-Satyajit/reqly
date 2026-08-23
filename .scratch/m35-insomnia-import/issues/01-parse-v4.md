# 01 — Insomnia v4 JSON parser (internal/importer)

**What to build:** Parse `__export_format: 4` exports: flat resources linked by parentId into folder tree; requests with method/url/headers(name→key)/parameters/body; environments collected for the writer; cookie jars dropped.

**Blocked by:** None

**Status:** done

- [x] Two-pass parentId resolution; orphans land at collection root
- [x] Headers: name→key, disabled respected; query parameters mapped; description dropped
- [x] Bodies: mimeType→implied Content-Type, urlencoded params encoded, form-data materialized (file rows warned), file mode warned+skipped
- [x] Auth: basic/bearer/apikey/digest mapped (location→in); others warned+dropped
- [x] Environments: name+data captured, nested values flattened to dotted keys w/ warning
- [x] Table-driven tests against vendored fixtures + inline cases
- [x] go vet/gofmt/go test green
