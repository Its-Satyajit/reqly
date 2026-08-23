# 02: Unary invoke end-to-end (`grpc:` block + CLI)

**What to build:** A developer saves a gRPC call as a request file — endpoint, `grpc:` block with service/method/message/timeout, headers as metadata — and runs `reqly grpc invoke <file>`: the JSON response message prints, a non-OK status or transport error exits 1 with code + message.

**Blocked by:** 01 — Reflection discovery + in-process gRPC test-server fixture.

**Status:** shipped (PR #311, 2026-08-24)

- [x] Request-file schema gains the optional `grpc:` block (service, method, message, timeout defaulting to 30s, optional protoFiles) and round-trips format-preserving like auth/retry blocks
- [x] Unary invoke via reflection resolves descriptors, sends metadata, returns the message as canonical protobuf-JSON
- [x] protoFiles fallback loads explicit workspace-relative `.proto` paths when reflection is unavailable
- [x] Non-OK status surfaces as failure: gRPC code number/name + status message (+ details), exit code 1
- [x] Transport errors (unreachable, deadline) exit 1 with readable messages
- [x] CLI tests cover success, assertion-free failure output, and exit codes against the shared fixture
