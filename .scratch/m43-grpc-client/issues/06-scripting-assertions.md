# 06: Scripting & assertions parity

**What to build:** A developer writes `reqly.test()` assertions and pre/post scripts for a gRPC request exactly as they would for HTTP: scripts receive the response message as JSON, assertions run unchanged, extracted values chain into variables.

**Blocked by:** 03 — Pipeline integration (env/vars, masking, history).

**Status:** shipped (PR #317, 2026-08-24)

- [x] Pre-request scripts can patch the outgoing message/metadata
- [x] Post-request scripts receive the response message as JSON and can set runtime variables
- [x] `reqly.test()` assertions (status-equivalent = gRPC OK, body JSON assertions) pass/fail identically to HTTP semantics
- [x] Non-OK status yields no response body for assertions, matching HTTP failure semantics
- [x] Tests exercise the shared testing/scripting engine with gRPC responses via the existing engine seams
