# 01: Reflection discovery + in-process gRPC test-server fixture

**What to build:** A developer can point at a gRPC endpoint and see which services and methods it exposes without any proto files. `internal/grpc` gains `Discover` over server reflection, the CLI gains `reqly grpc services <endpoint>`, and an in-process test server (reflection + echo unary + server-streaming services) exists as the shared fixture for all later M43 tickets.

**Blocked by:** None (can start immediately).

**Status:** shipped (PR #309, 2026-08-24)

- [x] `Discover` returns service names with their methods (input/output types) from a reflection-enabled server
- [x] Discovery against a reflection-disabled server fails with a clear error naming reflection as the problem
- [x] `reqly grpc services <endpoint>` prints services/methods; non-zero exit on connection or reflection failure
- [x] The in-process test fixture serves reflection plus one unary and one server-streaming echo method, reused by package tests
- [x] No CGO: only pure-Go dependencies added
