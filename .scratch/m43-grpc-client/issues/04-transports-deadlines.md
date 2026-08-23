# 04: Transports & deadlines (TLS, h2c, skip-verify, custom CA, timeout)

**What to build:** A developer can call plaintext h2c local dev servers and TLS production servers alike, trusting custom CAs or skipping verification when needed, with per-call deadlines enforced.

**Blocked by:** 02 — Unary invoke end-to-end (`grpc:` block + CLI).

**Status:** shipped (PR #314, 2026-08-24)

- [x] Plaintext h2c connections work against the fixture served without TLS
- [x] TLS works by default; skip-verify and custom CA path are configurable per request
- [x] Per-call `timeout` converts to a call deadline; expiry surfaces as a clear deadline-exceeded failure
- [x] Tests cover h2c, TLS with custom CA, skip-verify, and deadline expiry
