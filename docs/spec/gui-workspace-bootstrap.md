# Spec: Workspace Bootstrap (GUI-2.x)

> **Status:** Shipped 2026-08-24 — grill settled 2026-08-24 (Q1–Q4 confirmed)
> **Scope:** Open / create / switch Reqly workspaces from the desktop GUI; persisted last-workspace
> **Stack:** `apps/desktop/backend` (bridge + rebuild + persistence) + root `frontend/` (empty state, switcher)

## Problem Statement

The desktop app resolves its workspace from the process working directory at startup. Launched normally (dock, launcher, arbitrary terminal), no `reqly.yaml` is found and every view reports "no workspace found" — with no way to open, create, or select a folder from the GUI. The app is unusable unless launched from inside a workspace directory.

## Solution

When no workspace resolves, an empty state offers **Open folder…** and **Create workspace here**. A native directory picker feeds new bridge methods that validate (or scaffold) the workspace, rebuild all services in place, and persist the choice. The last workspace reopens automatically on launch. Once inside, a switcher beside the workspace name opens the same picker at any time.

## User Stories

1. As a new user launching Reqly from my Applications folder, I see an empty state that explains I need to pick a folder and lets me do it immediately.
2. As a developer with an existing Reqly workspace, I point the picker at it and the collections tree appears without restarting.
3. As a developer starting fresh, I pick any empty folder and Reqly scaffolds `reqly.yaml` + `collections/` for me.
4. As a daily user, Reqly reopens my last workspace on launch.
5. As a multi-workspace maintainer, I switch workspaces from the sidebar without leaving the app, and my open tabs reset cleanly.

## Implementation Decisions

**Bridge (Q1):**
- `WorkspaceStatus()` returns `{found bool, path string}` so the UI can branch at startup.
- `WorkspaceOpen(dir)` validates `IsWorkspace`, else errors with guidance toward create; rebuilds services in place by re-running exactly what `NewAppService` does (extracted `rebuildServices(root)`); persists the choice; returns the loaded tree.
- `WorkspaceCreate(dir, name?)` scaffolds `reqly.yaml` (`name:` defaults to folder name) + empty `collections/`, then delegates to `WorkspaceOpen`.
- The native directory picker lives behind an injectable function field so tests never touch a real dialog; production wires `application.Get().Dialog.OpenFileWithOptions(CanChooseDirectories)` + `PromptForSingleSelection`.
- No restart required (Q1); frontend performs tab/store resets client-side after a successful open — no event plumbing.

**Persistence (Q3):**
- Last workspace path stored at `os.UserConfigDir()/reqly/desktop.json`. On launch: prefer the stored path when still a valid workspace, else fall back to the current CWD walk. Invalid stored paths are ignored, not deleted.

**Create semantics (Q2):**
- Scaffolding is minimal and Git-native: descriptor + empty `collections/` only. No hidden files, no templates. Existing workspaces are never overwritten.

**UI surface (Q4):**
- Full-window empty state when `WorkspaceStatus().found == false`: short explanation of the local-first model + two CTAs (Open folder… / Create workspace…).
- Switcher icon beside the workspace name in the sidebar header once a workspace is open.

## Testing Decisions

- Single behavioral seam: `AppService.WorkspaceStatus/Open/Create` against temp directories — launch preference order, validation failure messaging, scaffold shape, rebuild correctness (tree loads post-switch), persistence round-trip (config dir injected, not global).
- Frontend verified by typecheck + lint gates; empty state and switcher are thin over the tested bridge.

## Out of Scope

- Multi-window / multiple simultaneously open workspaces.
- Recent-workspaces list beyond last-one (follow-up candidate).
- CLI `reqly init` subcommand (core gap noted; separate milestone).
- Drag-and-drop folder onto the window to open.

## Further Notes

- Recorded as new gui-roadmap item G-2.x on docs/internal; flipped on ship.
