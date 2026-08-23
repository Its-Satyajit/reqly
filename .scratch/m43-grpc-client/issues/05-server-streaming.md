# 05: Server-streaming invoke

**What to build:** A developer invokes a server-streaming method and receives each message as it arrives — line-delimited JSON on the CLI with `--max-messages` capping consumption — with cancel tearing down the stream cleanly.

**Blocked by:** 02 — Unary invoke end-to-end (`grpc:` block + CLI).

**Status:** shipped (PR #315, 2026-08-24)

- [x] Server-streaming calls emit each message as an event as it arrives (not buffered to completion)
- [x] `reqly grpc invoke --max-messages n` stops after n messages; final status still reported
- [x] Cancellation mid-stream closes the connection and exits cleanly
- [x] Stream errors surface with the same failure semantics as unary (code + message, exit 1)
- [x] Fixture streaming service exercises multi-message delivery ordering
