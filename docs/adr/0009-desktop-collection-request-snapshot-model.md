# ADR 0009: Desktop Collection Requests Use an Open-time Resolved Snapshot

## Status
Accepted

## Context
The desktop Collections Browser opens a collection request into the editor by resolving its inherited configuration (base URL, headers, auth) and variable chain server-side at open time. The question is what happens at *send* time: re-resolve from disk, or send the snapshot the user was shown.

## Decision
Sending an opened collection request uses the **open-time snapshot**: the resolved request fields the tab displays, plus the variable chain captured when the request was opened, plus the tab's environment pill (the request file's `environment:` if set, else the header-selected environment). The environment scope and process-env are layered *below* the snapshot at send, preserving precedence (request vars still beat env vars). The engine never re-reads the request file at send time. A workspace refresh reloads the sidebar tree only; open tabs keep their snapshot and are only replaced when the user re-opens the request.

## Considered Options
- **Re-derive from disk at send** — always current (picks up disk edits) but can disagree with what the tab displayed, breaks if the file was deleted mid-edit, and would clobber the WYSIWYG contract. Rejected: the editor is a draft; disk should not mutate it out from under the user.
- **Live file watching** — rejected for M16 in favor of on-demand refresh; the tree is a Git-native view and the user's editor is the source of truth until they refresh.

## Consequences
- **Positive:** WYSIWYG — what you saw is what's sent; safe against files changing or moving while a request is open; consistent with the per-tab draft model; matches the existing `request.Request` + `variables.Set` send seam.
- **Trade-off:** Disk edits to a request file are not reflected in an already-open tab until it is re-opened; the Variables tab shows the open-time values.