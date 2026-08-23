# 03: Pipeline integration (env/vars, masking, history)

**What to build:** A gRPC send behaves like every other send: environment selection and variable resolution apply across url/message/metadata, secrets in metadata or status details are masked in all output, and unary sends are recorded in history so they can be searched and replayed later.

**Blocked by:** 02 — Unary invoke end-to-end (`grpc:` block + CLI).

**Status:** shipped (PR #312, 2026-08-24)

- [x] Environment selection and 6-scope variables resolve across url, message JSON text, and metadata before parse; malformed JSON after interpolation is a validation error
- [x] Secret masking covers metadata values and gRPC status details in CLI output and history
- [x] Unary sends record a history row through the shared pipeline (Send Fidelity holds)
- [x] `{{$tags}}` generate per occurrence in message text
- [x] Pipeline tests assert masked output contains no secret values
