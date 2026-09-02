# ADR 0041: Shared Workspaces — Git-Native Collaboration (M74)

## Status
Accepted — grill Q1 (core + CLI + desktop, server deferred)

## Context
P3 §58.4 Team/Shared Workspaces needs Git-native sharing without cloud. Collections are already Git-native.

## Decision
`internal/collab` (M74): `SharedWorkspace{Path,Collaborators}` + `Collaborator{User,Role,AddedAt}` (viewer/editor/admin) + `Validate` + `Add/Remove/IsCollaborator` + `Load`/`Save`/`DefaultPath` (`.reqly/collab.yaml` 0600) + CLI `reqly collab list/add/remove` + desktop `CollabList/Add/Remove`.

## Consequences
Q1: File-based, not real-time — sharing via Git commit/push.
Q2: Server deferred — collab is file, not HTTP.
