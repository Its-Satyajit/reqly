# 07: Desktop gRPC tab kind

**What to build:** A desktop user opens a gRPC tab, picks a service/method discovered from the endpoint (or loaded from protoFiles), edits the JSON message with environment interpolation visible, sends it — unary results render as JSON; streams append timestamped messages live — and Stop cancels like HTTP sends. Sends record history.

**Blocked by:** 03 — Pipeline integration (env/vars, masking, history); 05 — Server-streaming invoke.

**Status:** shipped (PR #318, 2026-08-24)

- [x] New dedicated tab kind following the realtime-tab pattern, listed alongside other kinds
- [x] Bridge methods GrpcServices / GrpcInvoke / GrpcCancel wrap the shared core; bridge parses input, core owns pipeline steps
- [x] Service/method picker fed by Discover, falling back to protoFiles declared in the request file
- [x] Unary result renders as pretty JSON with gRPC status pill; streaming appends timestamped rows into a capped inspector
- [x] Stop cancels the streaming session via GrpcCancel; user-cancelled sends leave no fabricated response (context cancellation surfaces as cancelled event)
- [x] React Doctor gate passes with no new warnings vs baseline
