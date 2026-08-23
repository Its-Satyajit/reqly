# 02 — Workspace writer with collection defaults + secret-split env files

**What to build:** `BrunoResult.Write(dir)` writing reqly.yaml, collections/<name>/reqly.yaml carrying auth+headers defaults, folder tree via the shared writer, and environments/<name>.yaml with variables+secrets maps.

**Blocked by:** 01

**Status:** done

- [x] Collection descriptor carries root auth + headers
- [x] Env files: secrets map populated from secret:true vars
- [x] Round-trip test: write → collections.LoadWorkspace → tree/auth/headers/env files verified
- [x] go vet/gofmt/go test green
