# 01 — Postman parser (internal/importer)

**What to build:** `ParsePostman(data []byte) (*PostmanResult, []string, error)` for Postman v2.1 collection JSON: requests with method/URL (string or object form)/query/headers, nested folder tree, collection+request variables, bodies (raw/urlencoded/formdata/graphql; file mode → warning), auth mapping (basic/bearer/apikey only), script warnings.

**Blocked by:** None — can start immediately

**Status:** done

- [x] Mirror structs private; URL accepts both string and `{raw,host,path,query}` via custom unmarshal
- [x] v2.1 schema check — other versions parse but warn
- [x] Collection `variable[]` → result-level map; request `variable[]` → request-file variables; disabled skipped
- [x] Bodies: raw (+ language → implied Content-Type when header absent), urlencoded/formdata → enabled rows (`disabled` respected), formdata file rows keep local `src` + warning, graphql query+variables, `file` mode → warning + skip
- [x] Auth: basic{username,password}, bearer{token}, apikey{key,value,in}; others warned + dropped; request-level overrides collection-level
- [x] `event[]` scripts → one warning per scripted item ("script not imported")
- [x] Table-driven tests: fixture collections covering each body mode, both URL forms, folders, vars, auth variants, warnings
- [x] `go vet` + `gofmt -l` + `go test ./internal/importer` green
