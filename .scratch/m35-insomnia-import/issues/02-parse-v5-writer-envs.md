# 02 — Insomnia v5 YAML parser + workspace writer with environments

**What to build:** Parse `type: collection.insomnia.rest/5.0` YAML: hierarchical collection[].children tree; same request/auth/body mapping as T1. Then `InsomniaResult.Write(dir)` writing the workspace incl. `environments/<name>.yaml` per imported environment.

**Blocked by:** 01

**Status:** done

- [x] v5 sniffing: type prefix collection.insomnia.rest/, version mismatch warns
- [x] children recursion → folders; settings/cookieJar blocks ignored
- [x] environments block accepts dict and list shapes; subEnvironments flattened as separate env files
- [x] Write: reqly.yaml + collections/<name>/ tree + environments/*.yaml (0600/0644)
- [x] Round-trip tests: write → collections.LoadWorkspace + env files loadable
- [x] go vet/gofmt/go test green
