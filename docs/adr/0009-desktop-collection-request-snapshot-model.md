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

## Amendment (Milestone 17): Draft-based Editing with Send-time Re-resolution

### Context
M16 shipped the open-time snapshot as a *send* decision, but the editor was still read-only. M17 makes collection tabs editable (Save, dirty tracking, conflict handling), which forces a re-examination: a saved file's inherited config can change under a tab, and the file's own non-builder fields (auth, timeout, scripts) must survive editing.

### Decision
File-backed tabs become **drafts over the raw file request**. Opening seeds the editor from the file's *unmerged* request (builder fields only: url/method/headers/query/body); saving writes only those fields back, preserving format (JSON/YAML by extension) and every non-editable field verbatim via an atomic temp-file+rename. Sends re-resolve the **live draft** through the full inheritance chain at send time (`ResolveSend`), so inherited base URL/headers/auth and the variable scopes are recomputed from the containers rather than taken from an open-time snapshot; the environment scope layers below them. A content fingerprint taken at open guards saves against concurrent on-disk edits, surfacing a changed-on-disk conflict (Overwrite/Reload) instead of clobbering. The scratchpad tab keeps its raw-send behavior.

### Supersedes
For file-backed tabs, the send-time snapshot from the original Decision is replaced by send-time re-resolution of the live draft. The open-time *resolved* view still drives display (Effective URL line, inherited-headers group) and the Variables tab; the environment pill and its layering rules are unchanged. The scratchpad was never snapshot-based.