# M43: gRPC Client (Unary + Server-Streaming)

Status: shipped spec · ADR 0028 · tracker: local markdown (`docs/internal` only)

## Problem Statement

API developers working with gRPC services have to leave Reqly for grpcurl/Postman: they cannot discover what services a server exposes, invoke procedures with proper auth/metadata, or keep those calls versioned next to their HTTP requests. Reqly's Git-native workspace has no representation for a gRPC call at all, so none of the existing machinery — environments, variable inheritance, secret masking, assertions, history — applies.

## Solution

A gRPC call becomes a first-class citizen of the same request file model as HTTP requests: an endpoint, a fully-qualified method, a JSON message, metadata, and timeout live in the request file's `grpc:` config block. The client learns schemas from **server reflection** by default, with explicit workspace-relative `.proto` files as fallback. Unary calls return one message; server-streaming calls stream messages into a live inspector — reusing the realtime tab's interaction model on desktop and line-delimited output on CLI. Assertions and scripts run against the response message exactly as they do for HTTP.

## User Stories

1. As an API developer, I want to discover a server's services and methods via reflection, so that I don't need proto files on hand to explore it.
2. As an API developer, I want to point a request at `.proto` files in my workspace, so that reflection-disabled servers still work.
3. As an API developer, I want my gRPC call saved as a plain-text request file in Git, so that it versions and reviews like every other request.
4. As an API developer, I want to write the request message as JSON, so that I don't hand-encode protobuf bytes.
5. As an API developer, I want response messages rendered as JSON, so that I can read them without tooling.
6. As an API developer, I want headers in my request file sent as gRPC metadata, so that auth tokens ride existing inheritance.
7. As an API developer, I want per-call deadlines (`timeout: 5s`), so that slow servers fail fast with a clear error.
8. As an API developer, I want plaintext h2c and TLS transports with optional skip-verify / custom CA path, so that local dev (h2c) and production (TLS) both work.
9. As an API developer, I want server-streaming responses streamed into a live message list, so that I watch events arrive like SSE.
10. As an API developer, I want non-OK gRPC statuses rendered as failed responses (code + message + details), so that failures look like HTTP failures.
11. As an API developer, I want `reqly.test()` assertions against the response message, so that my existing assertion vocabulary works unchanged.
12. As an API developer, I want pre/post scripts receiving the response message as JSON, so that I can chain values into variables.
13. As an API developer, I want gRPC sends recorded in history, so that I can search/replay them like HTTP traffic.
14. As an API developer, I want environment variables and `{{$tags}}` interpolated in URL/message/metadata, so that per-env endpoints just work.
15. As an API developer, I want `reqly grpc invoke <file>`, so that CI can exercise gRPC endpoints.
16. As an API developer, I want `reqly grpc services <endpoint>`, so that service discovery is scriptable.
17. As a desktop user, I want a dedicated gRPC tab kind with a schema browser for the resolved service, so that I pick methods instead of typing full paths blind.
18. As a desktop user, I want streaming messages appended to a live inspector with timestamps, so that long-lived streams are debuggable.
19. As a desktop user, I want Stop/cancel semantics identical to HTTP sends, so that cancellation behaves consistently.
20. As a CLI user, I want exit code 1 when any invocation fails or a status is non-OK, so that shell scripts branch correctly.

## Implementation Decisions

1. **New pure package `internal/grpc`** (replaces the doc stub): `Discover(endpoint, opts) ([]Service, error)` over server reflection; `Invoke(ctx, call, opts) (<-chan Event, func(), error)` returning a stream of message/error/status events plus cancel. Options carry TLS config, headers (metadata), and deadline. Dependencies: `google.golang.org/grpc`, `google.golang.org/protobuf` (+ `grpc` reflection client) — pure Go, no-CGO rule intact.
2. **Schema acquisition:** reflection first; fallback loads descriptors from explicit workspace-relative `protoFiles:` paths via protobuf descriptor creation (`protoparse`-style parsing of `.proto` source). No auto-discovery of directories (ADR 0028).
3. **Request file:** new optional `grpc:` block on the existing request file format — `{service, method, message (JSON object or string), timeout ("30s" default), protoFiles?[]}`; `url` carries `host:port`; `headers` are metadata. Format-preserving save extends naturally (same machinery as `auth:`/`retry:` blocks). No new file kind.
4. **Send pipeline integration:** a gRPC send flows through `core.RequestService` like HTTP — env selection, variable interpolation across url/message/metadata, secret masking of metadata and error text, history recording. Streaming sends record one history row summarizing the call (messages received count, final status).
5. **Error model:** non-OK status → returned as failure with code number/name, message, and details; assertions see no body. Transport errors map like network errors (retryable classification untouched — no automatic retry for v1).
6. **CLI:** `reqly grpc invoke <file> [--env name] [--deadline d] [--max-messages n]` prints each message (line-delimited JSON) and final status; exit 1 on transport error or non-OK status. `reqly grpc services <endpoint> [--proto f...]` lists services/methods.
7. **Desktop:** new tab kind `kind: "grpc"` (pattern: realtime tabs): endpoint+metadata editor, service/method picker fed by Discover (or protoFiles), JSON message editor, response pane switching between unary result and streaming message list (timestamps, Stop button). New AppService methods `GrpcServices` / `GrpcInvoke` / `GrpcCancel` bridging to the shared core — bridge parses, core owns pipeline steps (Send Fidelity).
8. **Out of the runner:** collection runs stay HTTP-only this milestone (ADR 0028).

## Testing Decisions

- **Good tests assert external behavior**: what a parsed request file round-trips to, what events `Invoke` emits against a real server, what the CLI prints/exits — never internal channel shapes.
- **In-process gRPC test server is the single fixture**: register a reflection service + echo/unary and a server-streaming service (grpc-go interop-style helpers), reuse across package, CLI, and bridge tests — mirrors how realtime tests use httptest loopback servers.
- **Modules tested:** `internal/grpc` (discovery, unary, streaming, status mapping, TLS/h2c, deadline); `requestfile` round-trip of the `grpc:` block; Cobra command output/exit codes (prior art: `apps/cli/cmd/realtime_test.go`, `run_test.go`); AppService bridge methods with captured frames (prior art: `realtime_test.go`, `docsview_test.go`).
- **Assertions parity** is tested through the shared `internal/testing` engine with a gRPC response message as input — no new assertion code paths.

## Out of Scope

Client-streaming and bidirectional RPC (fast-follow with interactive composer); mTLS client certificates (proxy & TLS controls milestone); collection-run steps of gRPC type; automatic retry of gRPC calls; proto compilation from `.proto` beyond descriptor loading (no codegen artifacts committed); GUI service-graph visualization.

## Further Notes

- ADR 0028 records the reflection-first and same-request-file decisions with rejected alternatives; CONTEXT.md gains *gRPC Call*, *Server Reflection*, *Message*.
- `{{$tag}}` generation and 6-scope variable resolution apply to the message JSON text before parse — malformed JSON after interpolation surfaces as a request validation error.
- Secret masking must cover metadata values and gRPC status details (they can echo header material).
