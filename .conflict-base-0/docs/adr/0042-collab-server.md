# ADR 0042: Collab Server — Self-Hosted HTTP Server (M75)

## Status
Accepted — grill Q1 (core + CLI + desktop)

## Context
P3 §58.4 self-hosted collaboration server needs local HTTP for shared workspaces without cloud.

## Decision
`internal/collab.Server` (M75): `Server{root,mux}` + `NewServer` + `Handler` + `/health` (JSON ok), `/collab` (GET `Load` → JSON `SharedWorkspace`), `/workspace` (GET `Load` + `Glob collections/*` → `{"path","collaborators","collections","health":"ok"}`), `net.Listen` ephemeral, `http.Serve`; CLI `reqly collab serve --port` + desktop `CollabServe(port)` (`net.Listen` 127.0.0.1:port, goroutine).

## Consequences
Q1: No auth — RBAC not enforced on server, open local.
Q2: No WebSocket — polling, not real-time.
