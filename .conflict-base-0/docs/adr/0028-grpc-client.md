# ADR 0028: gRPC Client — Reflection-First, Same Request File

## Status
Accepted

## Context
`ROADMAP.md` P1 lists **gRPC** (proto files, reflection, service/method discovery, unary + streaming) as unchecked; `internal/grpc/` holds only a package doc stub and `go.mod` has no grpc dependencies. The existing seams are strong: request files carry auth/headers/variables with workspace inheritance (`internal/collections`, `internal/variables`), the send pipeline is shared by CLI + Desktop (`core.RequestService`, Send Fidelity), realtime tabs established the dedicated-tab-kind pattern for non-HTTP protocols, and the assertion engine (`internal/testing`) consumes JSON. Design questions: protocol scope, schema acquisition, on-disk representation, GUI surface, CLI shape, scripting parity, history/collection-run integration, deadlines, metadata, proto location, and error semantics.

## Decision
1. **Scope is unary + server-streaming first** (client-stream/bidi deferred). These cover most real services and reuse the WS/SSE interaction model; interactive senders for client/bidi come later as a fast-follow.
2. **Reflection-first schema acquisition.** The server reflection service discovers services/methods/messages at runtime; explicit workspace-relative `protoFiles:` paths in the request's `grpc:` config block are the fallback for reflection-disabled servers. No auto-discovery of a `proto/` directory — explicit paths keep the contract predictable and Git-native.
3. **Same request file, new `grpc:` block.** No separate file kind: `url` = `host:port`, headers = gRPC metadata (inheritance for free), plus `grpc: {service, method, message (JSON), timeout ("30s" default), protoFiles?}`. This reuses environments, variables, secret masking, history recording, and editor tabs almost unchanged.
4. **Full scripting/assertions parity.** Response messages are JSON-mapped (canonical protobuf-JSON), so `reqly.test()` and pre/post scripts work identically to HTTP.
5. **History yes, collection-run no (v1).** gRPC sends record history through the normal pipeline; mixing steps into the sequential collection runner is deferred to its own pass.
6. **Transports:** plaintext h2c and TLS with optional skip-verify / custom CA path; mTLS client certs ride the later proxy & TLS controls milestone.
7. **Surfaces:** Desktop gets a dedicated tab kind (`kind: "grpc"`); CLI gets `reqly grpc invoke <file>` and `reqly grpc services <endpoint>`.
8. **Error model:** non-OK gRPC status renders as a failed response — code as status, status message, details shown — with no response body for assertions, matching HTTP failure semantics.

## Considered Options
- **All four RPC shapes in v1** — rejected: client-streaming/bidi need an interactive multi-message composer; shipping them half-designed blocks the 90% case.
- **Separate `.grpc` file kind** — rejected: duplicates auth/variable/environment machinery and fragments the workspace tree for zero gain.
- **Proto-file-only schemas** — rejected: reflection is zero-config against the majority of servers and enables service discovery in the GUI; file loading stays as the fallback.
- **Separate `metadata:` map** — rejected: headers already are key/value pairs with enable toggles and chain inheritance; a second map would need its own merge rules.
- **Interactive REPL (`reqly ws`-style) in v1** — rejected: file-driven invoke fits run/test pipelines and CI; REPL is additive.

## Consequences
- **Positive:** Closes the largest P1 protocol gap reusing nearly every existing seam (send pipeline, inheritance, assertions, history, tab kinds); reflection gives GUI service discovery without user-supplied protos.
- **Trade-off:** Adds `google.golang.org/grpc` + `protobuf` deps (pure Go, no-CGO rule intact). Client-streaming/bidi users wait for the follow-up; collection runs stay HTTP-only until the runner gains typed steps.
