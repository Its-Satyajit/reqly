# 01 — Introspection client + model (internal/graphql)

**Blocked by:** None

**Status:** done

- [x] Standard introspection query POSTed as {query} JSON; headers/timeout options
- [x] GraphQL errors[] → typed error with messages; non-2xx and malformed JSON distinct errors
- [x] Schema model: kinds, fields, args, wrapped TypeRefs ([X]!, X!)
- [x] Text summary renderer: roots first, types alphabetical, --type filter support
- [x] Table-driven tests with canned responses (wrappers, enums, errors, bad JSON)
- [x] go vet/gofmt/go test green
