# 01 — Validation core (internal/jsonschema)

**Blocked by:** None

**Status:** done

- [x] Promote santhosh-tekuri/jsonschema/v6 to direct dep
- [x] Compile(schemaData) with $schema draft detection + --draft override; YAML/JSON input
- [x] Validate(instance) → violations [{path, message, keywordLocation}]; stdin/- handling is CLI-side
- [x] Distinct errors: read failure, compile failure, parse failure
- [x] Table-driven tests across drafts, nested paths, arrays, bad inputs
