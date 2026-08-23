# 04 — Send fidelity (architecture review candidate 3)

**What to build:** Every entry point sends with identical fidelity: the Cookie Jar lives inside the execution pipeline (attach + ingest, independent of history recording), replay re-sends the stored request exactly as captured, and single-flight policy has one owner.

**Blocked by:** 01, 02 — Run pipeline + CLI migration

**Status:** ready-for-agent

- [x] `RunOptions.AttachCookies` (nil = on): jar matching/splicing moved from the Wails bridge into core; CLI sends now participate in the jar
- [x] Set-Cookie ingestion decoupled from history recording (`HistoryService.IngestSetCookies`); a send that opts out of recording still joins the jar
- [x] `HistoryService.ShowRaw` + faithful `HistoryReplay`: Authorization and captured Cookie header re-sent exactly; jar attach skipped for verbatim semantics; TODO comment resolved
- [x] Single-flight: bridge checks `CollectionRunService.Active()` for an immediate error; core `acquire()` remains the race-safe enforcer
- [x] Desktop token store opened through `secrets.OpenForWorkspace(root, "keychain")`; bridge-local `openAppTokenStore`, `resolveAppEnvironment`, `recordHistory`, inline cookie splice all deleted
- [x] Tests: jar attach + opt-out, ShowRaw vs Show masking, desktop suite green; `-race` clean
