# Extreme Quality & Lint Audit Report

- **Generated At:** 2026-09-01 15:54:07 UTC
- **Completed At:** 2026-09-01 15:54:30 UTC
- **Overall Status:** 🔴 **FAILED / ISSUES FOUND**

---

## Executive Summary

| # | Check / Tool | Status | Duration | Command |
|---|--------------|:------:|:--------:|---------|
| 1 | **Frontend TypeScript Typecheck** | 🟢 PASS | 6s | `nub run typecheck` |
| 2 | **Frontend AST Lint (oxlint)** | 🟢 PASS | 6s | `nub run lint:frontend` |
| 3 | **Go Compiler Build** | 🟢 PASS | 6s | `go build ./...` |
| 4 | **Go Vet (Suspicious Constructs)** | 🟢 PASS | 1s | `go vet ./...` |
| 5 | **Go Unit & Integration Tests** | 🟢 PASS | 10s | `go test ./...` |
| 6 | **Go Race Detector Tests** | 🟢 PASS | 5s | `go test -race ./...` |
| 7 | **Go Deep Static Analysis (staticcheck)** | 🔴 FAIL | 2s | `staticcheck ./...` |
| 8 | **Go Dead Code / Reachability (deadcode)** | 🟢 PASS | 11s | `deadcode ./...` |
| 9 | **JS/TS Dead Code (fallow dead-code)** | 🔴 FAIL | 6s | `nubx -y fallow dead-code` |
| 10 | **JS/TS Duplication & Clones (fallow dupes)** | 🟢 PASS | 6s | `nubx -y fallow dupes` |
| 11 | **JS/TS Health & Complexity (fallow health)** | 🔴 FAIL | 6s | `nubx -y fallow health` |
| 12 | **JS/TS Advisory Review (fallow review)** | 🟢 PASS | 7s | `nubx -y fallow review` |
| 13 | **React Architecture & Performance (react-doctor)** | 🟢 PASS | 23s | `CI=1 printf '\n' | nubx -y react-doctor@latest` |

---

## Detailed Tool Output Logs

### 1. Frontend TypeScript Typecheck

- **Command:** `nub run typecheck`
- **Status:** 🟢 PASS
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text
$ nub run --recursive --if-present typecheck
Scope: 2 of 3 workspace projects
frontend typecheck$ tsc --noEmit
frontend typecheck: Done
apps/desktop/frontend typecheck$ tsc --noEmit
apps/desktop/frontend typecheck: Done
```

</details>

---

### 2. Frontend AST Lint (oxlint)

- **Command:** `nub run lint:frontend`
- **Status:** 🟢 PASS
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text
$ oxlint
Found 0 warnings and 0 errors.
Finished in 5.6s on 189 files with 111 rules using 12 threads.
```

</details>

---

### 3. Go Compiler Build

- **Command:** `go build ./...`
- **Status:** 🟢 PASS
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text

```

</details>

---

### 4. Go Vet (Suspicious Constructs)

- **Command:** `go vet ./...`
- **Status:** 🟢 PASS
- **Duration:** 1s

<details>
<summary>Click to expand full output</summary>

```text

```

</details>

---

### 5. Go Unit & Integration Tests

- **Command:** `go test ./...`
- **Status:** 🟢 PASS
- **Duration:** 10s

<details>
<summary>Click to expand full output</summary>

```text
?   	github.com/Its-Satyajit/reqly/apps/cli	[no test files]
ok  	github.com/Its-Satyajit/reqly/apps/cli/cmd	6.630s
?   	github.com/Its-Satyajit/reqly/apps/desktop	[no test files]
ok  	github.com/Its-Satyajit/reqly/apps/desktop/backend	(cached)
ok  	github.com/Its-Satyajit/reqly/e2e	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/ai	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/audit	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/auth	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/automation	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/bulk	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/collab	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/collections	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/core	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/diffing	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/docs	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/environments	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/exporter	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/git	1.569s
ok  	github.com/Its-Satyajit/reqly/internal/git/provider	0.231s
ok  	github.com/Its-Satyajit/reqly/internal/graphql	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/grpc	(cached)
?   	github.com/Its-Satyajit/reqly/internal/grpc/testsrv	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/history	(cached)
?   	github.com/Its-Satyajit/reqly/internal/history/db	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/importer	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/integration	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/jsonschema	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/jwt	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mcp	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mocking	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/monitor	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mqtt	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/openapi	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/pagination	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/perf	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/plugin	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/policy	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/rbac	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/request	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/requestfile	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/response	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/runner	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/scim	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/scripting	0.107s
ok  	github.com/Its-Satyajit/reqly/internal/secrets	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/socketio	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/sse	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/sso	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/testing	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/testsupport	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/theme	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/update	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/validation	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/variables	(cached)
?   	github.com/Its-Satyajit/reqly/internal/version	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/websocket	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/workflow	(cached)
```

</details>

---

### 6. Go Race Detector Tests

- **Command:** `go test -race ./...`
- **Status:** 🟢 PASS
- **Duration:** 5s

<details>
<summary>Click to expand full output</summary>

```text
?   	github.com/Its-Satyajit/reqly/apps/cli	[no test files]
ok  	github.com/Its-Satyajit/reqly/apps/cli/cmd	(cached)
?   	github.com/Its-Satyajit/reqly/apps/desktop	[no test files]
ok  	github.com/Its-Satyajit/reqly/apps/desktop/backend	(cached)
ok  	github.com/Its-Satyajit/reqly/e2e	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/ai	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/audit	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/auth	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/automation	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/bulk	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/collab	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/collections	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/core	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/diffing	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/docs	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/environments	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/exporter	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/git	2.856s
ok  	github.com/Its-Satyajit/reqly/internal/git/provider	1.273s
ok  	github.com/Its-Satyajit/reqly/internal/graphql	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/grpc	(cached)
?   	github.com/Its-Satyajit/reqly/internal/grpc/testsrv	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/history	(cached)
?   	github.com/Its-Satyajit/reqly/internal/history/db	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/importer	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/integration	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/jsonschema	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/jwt	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mcp	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mocking	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/monitor	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/mqtt	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/openapi	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/pagination	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/perf	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/plugin	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/policy	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/rbac	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/request	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/requestfile	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/response	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/runner	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/scim	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/scripting	1.266s
ok  	github.com/Its-Satyajit/reqly/internal/secrets	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/socketio	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/sse	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/sso	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/testing	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/testsupport	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/theme	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/update	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/validation	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/variables	(cached)
?   	github.com/Its-Satyajit/reqly/internal/version	[no test files]
ok  	github.com/Its-Satyajit/reqly/internal/websocket	(cached)
ok  	github.com/Its-Satyajit/reqly/internal/workflow	(cached)
```

</details>

---

### 7. Go Deep Static Analysis (staticcheck)

- **Command:** `staticcheck ./...`
- **Status:** 🔴 FAIL
- **Duration:** 2s

<details>
<summary>Click to expand full output</summary>

```text
apps/cli/cmd/client.go:34:6: func maskAcquiredToken is unused (U1000)
apps/cli/cmd/environment.go:64:6: func mergeEnvScope is unused (U1000)
apps/cli/cmd/test_test.go:228:13: unnecessary use of fmt.Sprintf (S1039)
apps/desktop/backend/export.go:122:3: this value of path is never used (SA4006)
apps/desktop/backend/import_test.go:111:2: this value of res is never used (SA4006)
internal/collections/workspace_test.go:155:2: this value of coll is never used (SA4006)
internal/history/cookie_match_test.go:23:2: this value of got is never used (SA4006)
internal/history/history.go:564:6: func scanEntries is unused (U1000)
internal/importer/wsdl.go:379:20: func (*wsdlNode).text is unused (U1000)
internal/importer/wsdl.go:414:6: func collectSchemas is unused (U1000)
internal/jsonschema/generate_test.go:203:5: should omit nil check; len() for nil slices is defined as zero (S1009)
internal/mocking/server.go:253:6: type failureCountKey is unused (U1000)
internal/openapi/generate_test.go:97:6: func findGenerated is unused (U1000)
internal/pagination/pagination.go:301:3: should replace this if statement with an unconditional strings.TrimPrefix (S1017)
```

</details>

---

### 8. Go Dead Code / Reachability (deadcode)

- **Command:** `deadcode ./...`
- **Status:** 🟢 PASS
- **Duration:** 11s

<details>
<summary>Click to expand full output</summary>

```text
apps/cli/cmd/client.go:34:6: unreachable func: maskAcquiredToken
apps/cli/cmd/environment.go:64:6: unreachable func: mergeEnvScope
apps/desktop/backend/app.go:376:6: unreachable func: setEmitRunEvent
apps/desktop/backend/app.go:382:6: unreachable func: getEmitRunEvent
internal/collab/collab.go:83:6: unreachable func: IsCollaborator
internal/core/environment.go:60:6: unreachable func: NewEnvironmentService
internal/core/request.go:48:6: unreachable func: NewRequestService
internal/core/request.go:56:6: unreachable func: NewCachedRequestService
internal/environments/dotenv.go:85:6: unreachable func: LoadDotEnv
internal/exporter/postman.go:157:6: unreachable func: ParsePostman
internal/git/provider/provider.go:193:6: unreachable func: ListSupportedProviders
internal/grpc/testsrv/server.go:60:6: unreachable func: Start
internal/grpc/testsrv/server.go:84:19: unreachable func: fixture.Echo
internal/grpc/testsrv/server.go:88:19: unreachable func: fixture.Slow
internal/grpc/testsrv/server.go:93:19: unreachable func: fixture.Boom
internal/grpc/testsrv/server.go:97:19: unreachable func: fixture.StreamEcho
internal/grpc/testsrv/server.go:108:6: unreachable func: StartPlain
internal/grpc/testsrv/server.go:124:6: unreachable func: StartTLS
internal/grpc/testsrv/server.go:150:6: unreachable func: write
internal/grpc/testsrv/server.go:158:6: unreachable func: selfSigned
internal/history/history.go:564:6: unreachable func: scanEntries
internal/importer/swagger2.go:30:6: unreachable func: ParseSwagger2
internal/importer/swagger2.go:44:23: unreachable func: swagger2Doc.ToOpenAPIResult
internal/importer/wsdl.go:222:6: unreachable func: ParseWSDL
internal/importer/wsdl.go:379:20: unreachable func: wsdlNode.text
internal/importer/wsdl.go:414:6: unreachable func: collectSchemas
internal/jwt/jwt.go:113:6: unreachable func: IsExpired
internal/policy/policy.go:72:6: unreachable func: EnforceWorkflow
internal/rbac/rbac.go:136:6: unreachable func: Save
internal/request/client.go:72:6: unreachable func: WithTimeout
internal/request/client.go:80:6: unreachable func: WithHTTPClient
internal/request/client.go:101:6: unreachable func: WithOnRetry
internal/scim/scim.go:42:6: unreachable func: ValidateGroup
internal/scim/scim.go:91:17: unreachable func: Store.GetUser
internal/scim/scim.go:113:17: unreachable func: Store.DeactivateUser
internal/scim/scim.go:126:17: unreachable func: Store.CreateGroup
internal/scim/scim.go:143:17: unreachable func: Store.AddUserToGroup
internal/scripting/runtime.go:31:6: unreachable func: NewRuntime
internal/scripting/runtime.go:36:19: unreachable func: Runtime.RunScript
internal/scripting/runtime.go:41:19: unreachable func: Runtime.ensureVM
internal/scripting/sandbox.go:226:19: unreachable func: Sandbox.RunString
internal/secrets/backend.go:55:6: unreachable func: SetKeychainOpenerForTest
internal/sse/sse.go:84:6: unreachable func: WithHTTPClient
internal/sse/sse.go:89:6: unreachable func: WithStatusHandler
internal/sso/sso.go:65:6: unreachable func: IsGroupAllowed
internal/testsupport/testsupport.go:38:6: unreachable func: Workspace
internal/testsupport/testsupport.go:47:6: unreachable func: Files
internal/theme/theme.go:86:6: unreachable func: MarshalJSON
internal/theme/theme.go:110:6: unreachable func: IsBuiltIn
internal/variables/variables.go:56:6: unreachable func: SetTagGeneratorForTest
internal/variables/variables.go:122:6: unreachable func: UnknownDynamicTags
internal/websocket/websocket.go:89:6: unreachable func: WithSubprotocols
internal/websocket/websocket.go:94:6: unreachable func: WithStatusHandler
```

</details>

---

### 9. JS/TS Dead Code (fallow dead-code)

- **Command:** `nubx -y fallow dead-code`
- **Status:** 🔴 FAIL
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text
nub: pnpm-workspace.yaml is not read under nub identity — migrate it (`nub pm use nub`), delete it, or return to pnpm (`nub pm use pnpm`).
nub 0.7.5
░░░░░░░░░░░░░░░   15/71 pkgs · resolving
███████████████    7/7 pkgs
✓ resolved 7 · reused 7 in 3.4s
dependencies:
+ fallow@3.21.0

loaded config: /home/satyajit/Documents/GitHub/OSS/api-client/reqly-main-brancn/.fallowrc.json
   0.433672053s  WARN Skipped 9 package.json entry points outside project root or containing parent directory traversal: /usr/local/bin/reqly (3x), /usr/local/bin/reqly-desktop (3x), /tmp/reqly (2x), ../frontend/bindings
  30 entry points detected (28 plugin, 2 manual entry)

── Unused Code ─────────────────────────────────────

● Unused files (23)
  frontend/src/components/ui/tabs.tsx
  frontend/src/components/ui/tooltip.tsx
  frontend/src/features/dep-graph/DepGraphView.tsx
  frontend/src/features/git-view/GitView.tsx
  frontend/src/features/monitor-view/MonitorView.tsx
  frontend/src/features/perf-view/PerfView.tsx
  frontend/src/lib/git.ts
  frontend/src/lib/perf.ts
  tools/oxlint/anti-slop/effect/index.ts
  tools/oxlint/anti-slop/effect/rules/no-service-constructor-imports.test.ts
  ... and 13 more (--format json for full list)
  Files not reachable from any entry point — https://docs.fallow.tools/explanations/dead-code#unused-files
  To suppress: // fallow-ignore-file unused-file

● Unused exports (78)
  apps/desktop/frontend/src/bridge.ts (18)
    :71 wailsSender
    :129 wailsAuthAdapter
    :171 wailsEnvAdapter
    :370 wailsCollectionsAdapter
    :494 wailsHistoryAdapter
    ... and 13 more (--format json for full list)
  frontend/src/components/ui/toast.tsx (11)
    :222 Toast
    :223 ToastAction
    :224 ToastClose
    :225 ToastContent
    :226 ToastDescription
    ... and 6 more (--format json for full list)
  frontend/src/components/ui/dropdown-menu.tsx (9)
    :252 DropdownMenuPortal
    :255 DropdownMenuGroup
    :258 DropdownMenuCheckboxItem
    :259 DropdownMenuRadioGroup
    :260 DropdownMenuRadioItem
    ... and 4 more (--format json for full list)
  frontend/src/lib/schemaGraph.ts (6)
    :12 SCHEMA_EDGES
    :35 ZOOM_MIN
    :36 ZOOM_MAX
    :37 NODE_WIDTH
    :38 NODE_HEIGHT
    ... and 1 more (--format json for full list)
  frontend/src/components/ui/select.tsx (5)
    :193 SelectGroup
    :195 SelectLabel
    :196 SelectScrollDownButton
    :197 SelectScrollUpButton
    :198 SelectSeparator
  frontend/src/components/ui/alert-dialog.tsx (4)
    :180 AlertDialogMedia
    :181 AlertDialogOverlay
    :182 AlertDialogPortal
    :184 AlertDialogTrigger
  frontend/src/components/ui/dialog.tsx (4)
    :151 DialogClose
    :156 DialogOverlay
    :157 DialogPortal
    :159 DialogTrigger
  frontend/src/components/ui/menubar.tsx (4)
    :184 MenubarGroup
    :188 MenubarSub
    :189 MenubarSubTrigger
    :190 MenubarSubContent
  frontend/src/lib/typeGuards.ts (3)
    :13 isNumber
    :21 isObject
    :25 isDefinedString
  frontend/src/components/ui/alert.tsx (2)
    :76 AlertAction
    :76 AlertTitle
  ... and 12 more in 10 files (--format json for full list)
  Exported symbols with no known consumers — https://docs.fallow.tools/explanations/dead-code#unused-exports
  To auto-fix: fallow fix --dry-run
  To suppress: // fallow-ignore-next-line unused-export
  (5 more in files already reported as unused)

── Dependencies ─────────────────────────────────────

● Unused dependencies (2)
  frontend
  next-themes
  Listed in dependencies but never imported — https://docs.fallow.tools/explanations/dead-code#unused-dependencies

── Structure ─────────────────────────────────────

● Circular dependencies (2)
  frontend/src/components/CrashReportDialog.tsx
    → ErrorBoundary.tsx → CrashReportDialog.tsx

  frontend/src/stores/useCollectionRunStore.ts
    → useWorkspaceStore.ts → useCollectionRunStore.ts

  Import cycles that can cause initialization failures and prevent tree-shaking — https://docs.fallow.tools/explanations/dead-code#circular-dependencies

✗ 23 files · 78 exports · 2 unused dependencies · 2 circular dependencies (0.57s)
```

</details>

---

### 10. JS/TS Duplication & Clones (fallow dupes)

- **Command:** `nubx -y fallow dupes`
- **Status:** 🟢 PASS
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text
nub: pnpm-workspace.yaml is not read under nub identity — migrate it (`nub pm use nub`), delete it, or return to pnpm (`nub pm use pnpm`).
nub 0.7.5
░░░░░░░░░░░░░░░   13/71 pkgs · resolving
███████████████    7/7 pkgs
✓ resolved 7 · reused 7 in 3.0s
dependencies:
+ fallow@3.21.0

loaded config: /home/satyajit/Documents/GitHub/OSS/api-client/reqly-main-brancn/.fallowrc.json
note: skipped 34 files matching default duplicates ignores (use --explain-skipped for the list)
note: module wiring excluded from clone detection (--no-ignore-imports to include it)

● Duplicates (22 clone groups)

     12 lines  2 instances  spread 2  dup:c77b3abb6f87acd9-10
    frontend/src/features/dep-graph/DepGraphView.tsx:75-86
    frontend/src/features/spec-editor/SpecEditorView.tsx:18-29

     23 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-18
    frontend/src/stores/useWorkspaceBootstrap.ts:143-165
    frontend/src/stores/useWorkspaceBootstrap.ts:173-195

     20 lines  3 instances  spread 2  dup:7eb52920
    tools/oxlint/anti-slop/rules/no-known-value-widening.ts:29-48
    tools/oxlint/anti-slop/rules/no-module-mocking.ts:7-23
    tools/oxlint/anti-slop/shared/reflect-method.ts:3-21

     24 lines  2 instances  dup:c77b3abb6f87acd9-7
    tools/oxlint/anti-slop/rules/no-object-parameters.ts:7-30
    tools/oxlint/anti-slop/rules/no-unknown-parameters.ts:4-27

      8 lines  3 instances  spread 1  dup:c77b3abb6f87acd9-4
    frontend/src/stores/useAuthStore.ts:34-41
    frontend/src/stores/useAuthStore.ts:48-55
    frontend/src/stores/useAuthStore.ts:62-69

    100 lines  2 instances  dup:bc6ab6e5
    tools/oxlint/anti-slop/rules/no-unknown-returns.ts:16-115
    tools/oxlint/anti-slop/rules/no-unknown-type-aliases.ts:5-70

     19 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-14
    frontend/src/features/command-palette/CommandPalette.tsx:57-74
    frontend/src/features/command-palette/CommandPalette.tsx:79-97

     10 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-3
    tools/oxlint/anti-slop/shared/dictionary-types.ts:343-352
    tools/oxlint/anti-slop/shared/dictionary-types.ts:416-423

      5 lines  2 instances  spread 1  dup:e82d7ce7
    frontend/src/components/WorkspaceSidebar.tsx:175-179
    frontend/src/components/shell/ContextSidebar.tsx:34-38

     15 lines  2 instances  spread 1  dup:31961b11
    tools/oxlint/anti-slop/shared/dictionary-types.ts:234-243
    tools/oxlint/anti-slop/shared/dictionary-types.ts:450-464

  ... and 12 more clone groups
  Duplicate code blocks - https://docs.fallow.tools/explanations/duplication#clone-groups

● Clone families (3 with multiple groups)

  2 groups, 27 lines across frontend/src/features/environments-view/EnvironmentEditor.tsx, frontend/src/features/environments-view/SecretsEditor.tsx
    → Extract shared function (15 lines) from EnvironmentEditor.tsx, SecretsEditor.tsx
    → Extract shared function (12 lines) from EnvironmentEditor.tsx, SecretsEditor.tsx

  2 groups, 17 lines across frontend/src/stores/useAuthStore.ts
    → Extract shared function (9 lines) from useAuthStore.ts, useAuthStore.ts
    → Extract shared function (8 lines) from useAuthStore.ts, useAuthStore.ts, useAuthStore.ts

  3 groups, 42 lines across tools/oxlint/anti-slop/shared/dictionary-types.ts
    → Extract shared function (10 lines) from dictionary-types.ts, dictionary-types.ts
    → Extract shared function (15 lines) from dictionary-types.ts, dictionary-types.ts
    → Extract shared function (17 lines) from dictionary-types.ts, dictionary-types.ts

  Groups of related clones across the same files — https://docs.fallow.tools/explanations/duplication#clone-families

✗ 689 lines (2.7%) duplicated across 26 files (0.50s)
```

</details>

---

### 11. JS/TS Health & Complexity (fallow health)

- **Command:** `nubx -y fallow health`
- **Status:** 🔴 FAIL
- **Duration:** 6s

<details>
<summary>Click to expand full output</summary>

```text
nub: pnpm-workspace.yaml is not read under nub identity — migrate it (`nub pm use nub`), delete it, or return to pnpm (`nub pm use pnpm`).
nub 0.7.5
░░░░░░░░░░░░░░░   13/71 pkgs · resolving
█████░░░░░░░░░░    7/32 pkgs · ~187.1 MB
███████████████    7/7 pkgs
✓ resolved 7 · reused 7 in 4.3s
dependencies:
+ fallow@3.21.0

loaded config: /home/satyajit/Documents/GitHub/OSS/api-client/reqly-main-brancn/.fallowrc.json
   0.155578447s  WARN Skipped 9 package.json entry points outside project root or containing parent directory traversal: /usr/local/bin/reqly (3x), /usr/local/bin/reqly-desktop (3x), /tmp/reqly (2x), ../frontend/bindings

● Health score: 60 C
  Deductions: hotspots -10.0 · unit size -10.0 · unused deps -4.4 · circular deps -4.4 · maintainability -3.9 · dead exports -2.6 · coupling -2.4 · dead files -2.0 · complexity -0.4

■ Metrics: 28,508 LOC · dead files 10.1% · dead exports 13.0% · avg cyclomatic 2.3 · p90 cyclomatic 5 · maintainability 89.1 (good) · 2 churn hotspots (since 6 months) · 2 circular deps · 2 unused deps · duplication 2.7%

  Function size: 84% low · 8% medium · 4% high · 4% very high  (1-15 / 16-30 / 31-60 / >60 LOC)

  Render fan-in: <Button> 44 parents (124 incl. repeats) · <Input> 20 parents (47 incl. repeats) · <Alert> 17 parents (19 incl. repeats) · <AlertDescription> 17 parents (19 incl. repeats) · <Spinner> 15 parents (24 incl. repeats)

● Large functions (10 shown, 94 total)
  frontend/src/features/request-editor/RequestEditor.tsx
    :113 RequestEditor  532 lines
  frontend/src/features/mock-view/MocksView.tsx
    :30 MocksView  514 lines
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :42 ResponseViewer  510 lines
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :40 RunnersPanel  326 lines
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :32 EnvironmentsView  302 lines
  frontend/src/features/history-view/HistoryView.tsx
    :37 HistoryView  295 lines
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :27 OpenapiExplorer  278 lines
  frontend/src/app/App.tsx
    :57 App  268 lines
  frontend/src/stores/useWorkspaceStore.ts
    :220 useWorkspaceStore  266 lines
  frontend/src/features/import-dialog/ImportDialog.tsx
    :112 ImportDialog  244 lines
  Functions exceeding 60 lines of code (very high risk): https://docs.fallow.tools/explanations/health#unit-size
  use --top 94 to see all

● High complexity functions (255)
  CRAP scores are estimated from export references; run `fallow health --coverage <coverage-final.json>` for exact scores.
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :42 ResponseViewer CRITICAL
          78 ! cyclomatic  271 ! cognitive  510 lines
         react: 16 hooks (4 state, 10 memo, 2 custom), JSX depth 7
         6162.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :113 RequestEditor CRITICAL
          61 ! cyclomatic   95 ! cognitive  532 lines
         react: 18 hooks (5 state, 1 effect, 12 custom), max effect deps 2, JSX depth 6
         3782.0 ! CRAP
  frontend/src/lib/specTree.ts
    :49 tryParseYaml CRITICAL
          47 ! cyclomatic   89 ! cognitive  103 lines
         524.1 ! CRAP
  frontend/src/lib/codegen.ts
    :24 generateCode CRITICAL
          34 ! cyclomatic   46 ! cognitive   54 lines
         1190.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :112 ImportDialog CRITICAL
          32 ! cyclomatic   45 ! cognitive  244 lines
         react: 1 props, 19 hooks (1 state, 18 custom), JSX depth 6
         blast radius: <ImportDialog> rendered in 3 places
         1056.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :278 normalizeOpenedRequest CRITICAL
          31 ! cyclomatic   15   cognitive   38 lines
         992.0 ! CRAP
  frontend/src/features/mock-view/MocksView.tsx
    :30 MocksView CRITICAL
          29 ! cyclomatic   64 ! cognitive  514 lines
         react: 29 hooks (2 state, 1 effect, 26 custom), max effect deps 0, JSX depth 8
         870.0 ! CRAP
  frontend/src/hooks/useKeyboardMap.ts
    :29 handler CRITICAL
          27 ! cyclomatic   36 ! cognitive   78 lines
         756.0 ! CRAP
  frontend/src/lib/body.ts
    :118 serializeBody CRITICAL
          26 ! cyclomatic   32 ! cognitive   59 lines
         702.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :200 knownValueEvidence CRITICAL
          26 ! cyclomatic   18 ! cognitive   62 lines
         702.0 ! CRAP
  frontend/src/app/App.tsx
    :57 App CRITICAL
          25 ! cyclomatic  211 ! cognitive  268 lines
         react: 19 hooks (1 state, 4 effect, 14 custom), max effect deps 2, JSX depth 10
         650.0 ! CRAP
  frontend/src/components/shell/BottomPanel.tsx
    :20 PanelContent CRITICAL
          25 ! cyclomatic   38 ! cognitive  222 lines
         react: 1 props, 9 hooks (9 custom), JSX depth 6
         650.0 ! CRAP
  frontend/src/features/test-runner/TestTab.tsx
    :202 TestTab CRITICAL
          25 ! cyclomatic   25 ! cognitive  123 lines
         react: 1 props, 7 hooks (1 effect, 6 custom), max effect deps 1, JSX depth 5
         650.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :105 DiffView CRITICAL
          24 ! cyclomatic   25 ! cognitive  181 lines
         react: 1 props, 11 hooks (9 state, 1 effect, 1 custom), max effect deps 1, JSX depth 6
         blast radius: <ChangesList> rendered in 2 places
         600.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :190 unsafeDirectValue CRITICAL
          24 ! cyclomatic   29 ! cognitive   55 lines
         600.0 ! CRAP
  frontend/src/features/env-tools/EnvToolsPanel.tsx
    :99 EnvToolsPanel CRITICAL
          23 ! cyclomatic   27 ! cognitive  183 lines
         react: 10 hooks (9 state, 1 custom), JSX depth 5
         blast radius: <EnvPicker> rendered in 3 places
         552.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :61 saveWarnings CRITICAL
          23 ! cyclomatic   28 ! cognitive   51 lines
         552.0 ! CRAP
  frontend/src/lib/datasets.ts
    :56 parseCsv CRITICAL
          23 ! cyclomatic   45 ! cognitive   65 lines
  frontend/src/lib/specTree.ts
    :268 patchEndpointInContent HIGH
          23 ! cyclomatic   39 ! cognitive   90 lines
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :51 createTypeEnvironment CRITICAL
          23 ! cyclomatic   37 ! cognitive   46 lines
         552.0 ! CRAP
  frontend/src/features/grpc-view/GrpcTab.tsx
    :39 GrpcTab CRITICAL
          22 ! cyclomatic   25 ! cognitive  155 lines
         react: 1 props, 9 hooks (2 effect, 7 custom), max effect deps 1, JSX depth 5
         506.0 ! CRAP
  frontend/src/lib/specTree.ts
    :233 yamlEscape CRITICAL
          22 ! cyclomatic    3   cognitive   28 lines
         126.5 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :52 isBroadRecordType CRITICAL
          22 ! cyclomatic   13   cognitive   31 lines
         506.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :246 dictionaryValueTypes CRITICAL
          22 ! cyclomatic   26 ! cognitive   60 lines
         506.0 ! CRAP
  frontend/src/components/RunView.tsx
    :109 RunView CRITICAL
          21 ! cyclomatic   39 ! cognitive  182 lines
         react: 15 hooks (4 state, 11 custom), JSX depth 7
         462.0 ! CRAP
  frontend/src/stores/useCommandPaletteStore.ts
    :99 groupByHint CRITICAL
          21 ! cyclomatic   28 ! cognitive   18 lines
         116.3 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :44 RealtimeTab CRITICAL
          20   cyclomatic   30 ! cognitive  209 lines
         react: 1 props, 11 hooks (3 state, 1 effect, 7 custom), max effect deps 1, JSX depth 4
         blast radius: <RealtimeTab> rendered in 3 places
         420.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :340 classifyWideningTarget CRITICAL
          20   cyclomatic   24 ! cognitive   42 lines
         420.0 ! CRAP
  frontend/src/features/test-runner/TestTab.tsx
    :25 AssertionBuilder CRITICAL
          19   cyclomatic   20 ! cognitive  148 lines
         react: 1 props, 4 hooks (2 state, 2 custom), JSX depth 6
         380.0 ! CRAP
  frontend/src/lib/response.ts
    :79 parseSetCookies CRITICAL
          19   cyclomatic   19 ! cognitive   52 lines
         380.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :411 classifyAliasBroadTarget CRITICAL
          19   cyclomatic   22 ! cognitive   55 lines
         380.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :27 OpenapiExplorer CRITICAL
          18   cyclomatic   18 ! cognitive  278 lines
         react: 6 hooks (2 state, 4 custom), JSX depth 6
         342.0 ! CRAP
  frontend/src/components/CollectionTree.tsx
    :17 treeKeyDown CRITICAL
          17   cyclomatic   18 ! cognitive   38 lines
         blast radius: <CollectionBranch> rendered in 2 places
         306.0 ! CRAP
  frontend/src/lib/themes.ts
    :80 parseSimpleYaml HIGH
          17   cyclomatic   23 ! cognitive   51 lines
          79.4 ! CRAP
  frontend/src/features/graphql-browser/GraphqlBrowser.tsx
    :92 GraphqlBrowser CRITICAL
          16   cyclomatic   18 ! cognitive  180 lines
         react: 7 hooks (7 state), JSX depth 6
         blast radius: <FieldRow> rendered in 2 places
         272.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :71 wailsSender CRITICAL
          15   cyclomatic   16 ! cognitive   52 lines
         240.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :67 displayEntries CRITICAL
          15   cyclomatic   18 ! cognitive   11 lines
         240.0 ! CRAP
  frontend/src/features/settings-view/SettingsView.tsx
    :47 SettingsView CRITICAL
          15   cyclomatic   19 ! cognitive  191 lines
         react: 7 hooks (1 state, 6 custom), JSX depth 8
         240.0 ! CRAP
  frontend/src/lib/request.ts
    :128 fetchSender CRITICAL
          15   cyclomatic   14   cognitive   51 lines
         240.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :263 widenedBinding CRITICAL
          15   cyclomatic    9   cognitive   37 lines
         240.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-returns.ts
    :42 resolvesToUnknown CRITICAL
          15   cyclomatic   12   cognitive   35 lines
         240.0 ! CRAP
  frontend/src/components/JsonTree.tsx
    :19 JsonNode CRITICAL
          14   cyclomatic   16 ! cognitive   64 lines
         react: 4 props, 1 hooks (1 state), JSX depth 3
         blast radius: <JsonNode> rendered in 2 places
         210.0 ! CRAP
  frontend/src/lib/jsonpath.ts
    :93 parsePath CRITICAL
          14   cyclomatic   26 ! cognitive   40 lines
         210.0 ! CRAP
  frontend/src/stores/useGrpcStore.ts
    :122 send CRITICAL
          14   cyclomatic   18 ! cognitive   67 lines
         210.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :480 isKnownEvidenceExpression CRITICAL
          14   cyclomatic    4   cognitive   23 lines
         210.0 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :164 <arrow> CRITICAL
          14   cyclomatic   12   cognitive   37 lines
         react: JSX depth 2
         blast radius: <RealtimeTab> rendered in 3 places
         210.0 ! CRAP
  frontend/src/features/docs-view/DocsView.tsx
    :17 DocsView CRITICAL
          13   cyclomatic   27 ! cognitive  168 lines
         react: 14 hooks (1 state, 1 effect, 12 custom), max effect deps 1, JSX depth 8
         182.0 ! CRAP
  frontend/src/features/spec-editor/SpecEditorView.tsx
    :43 SpecEditorView CRITICAL
          13   cyclomatic   26 ! cognitive  159 lines
         react: 15 hooks (1 state, 1 memo, 13 custom), JSX depth 6
         182.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :779 decode CRITICAL
          13   cyclomatic    7   cognitive   20 lines
         182.0 ! CRAP
  frontend/src/components/shell/ContextSidebar.tsx
    :450 ContextSidebar CRITICAL
          13   cyclomatic    3   cognitive   62 lines
         react: 1 props, 1 hooks (1 custom), JSX depth 2
         blast radius: <SectionLabel> rendered in 13 places
         182.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :672 RetrySection CRITICAL
          13   cyclomatic    8   cognitive   97 lines
         react: 2 props, 1 hooks (1 state), JSX depth 4
         182.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-runtime-typeof.ts
    :51 UnaryExpression CRITICAL
          13   cyclomatic    6   cognitive   22 lines
         182.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :40 RunnersPanel CRITICAL
          12   cyclomatic   18 ! cognitive  326 lines
         react: 3 hooks (1 state, 1 effect, 1 custom), max effect deps 1, JSX depth 7
         156.0 ! CRAP
  frontend/src/lib/jsonpath.ts
    :38 walk CRITICAL
          12   cyclomatic   16 ! cognitive   54 lines
         156.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :143 isDefinitelyNarrowerRecordType CRITICAL
          12   cyclomatic    8   cognitive   18 lines
         156.0 ! CRAP
  frontend/src/components/RunView.tsx
    :18 StepRow CRITICAL
          12   cyclomatic   11   cognitive   87 lines
         react: 2 props, 1 hooks (1 state), JSX depth 5
         156.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-object-parameters.ts
    :52 resolvesToObject CRITICAL
          12   cyclomatic    7   cognitive   30 lines
         156.0 ! CRAP
  frontend/src/lib/authSchemes.ts
    :306 authWarnings CRITICAL
          12   cyclomatic   11   cognitive   18 lines
         156.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-module-mocking.ts
    :51 moduleMockCall CRITICAL
          12   cyclomatic   11   cognitive   16 lines
         156.0 ! CRAP
  frontend/src/lib/response.ts
    :270 parseTable CRITICAL
          12   cyclomatic   15   cognitive   29 lines
         156.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :364 <arrow> CRITICAL
          12   cyclomatic   15   cognitive   33 lines
         react: JSX depth 2
         156.0 ! CRAP
  tools/oxlint/anti-slop/shared/lexical-type-parameters.ts
    :35 lexicalTypeParameterNames CRITICAL
          12   cyclomatic   15   cognitive   27 lines
         156.0 ! CRAP
  frontend/src/features/export-dialog/ExportDialog.tsx
    :22 ExportDialog CRITICAL
          11   cyclomatic   18 ! cognitive  111 lines
         react: 11 hooks (11 custom), JSX depth 6
         blast radius: <ExportDialog> rendered in 3 places
         132.0 ! CRAP
  frontend/src/lib/response.ts
    :134 cookieExpiry CRITICAL
          11   cyclomatic   18 ! cognitive   18 lines
         132.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :122 isDefinitelyObjectType CRITICAL
          11   cyclomatic    2   cognitive   20 lines
         132.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :216 InheritedAuth CRITICAL
          11   cyclomatic   11   cognitive   40 lines
         react: 1 props, JSX depth 2
         blast radius: <AuthFieldRow> rendered in 2 places
         132.0 ! CRAP
  frontend/src/features/graphql-browser/GraphqlBrowser.tsx
    :18 FieldRow CRITICAL
          11   cyclomatic    8   cognitive   47 lines
         react: 2 props, 1 hooks (1 state), JSX depth 3
         blast radius: <FieldRow> rendered in 2 places
         132.0 ! CRAP
  frontend/src/features/git-view/GitView.tsx
    :6 GitView CRITICAL
          10   cyclomatic   17 ! cognitive   65 lines
         react: 8 hooks (6 state, 1 effect, 1 callback), max effect deps 1, JSX depth 4
         110.0 ! CRAP
  frontend/src/features/jwt-inspector/JwtInspector.tsx
    :54 <arrow> CRITICAL
          10   cyclomatic    4   cognitive   16 lines
         react: JSX depth 3
         blast radius: <ClaimsTable> rendered in 2 places
         110.0 ! CRAP
  frontend/src/components/KeyValueEditor.tsx
    :33 <arrow> CRITICAL
          10   cyclomatic    9   cognitive   96 lines
         react: JSX depth 4
         blast radius: <KeyValueEditor> rendered in 5 places
         110.0 ! CRAP
  frontend/src/features/request-editor/TemplatePickerSheet.tsx
    :26 TemplatePickerSheet CRITICAL
          10   cyclomatic   14   cognitive  130 lines
         react: 3 props, 3 hooks (2 state, 1 custom), JSX depth 5
         110.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :290 <arrow> CRITICAL
          10   cyclomatic    9   cognitive   31 lines
         blast radius: <ImportDialog> rendered in 3 places
         110.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :383 isBroadMappedKey CRITICAL
          10   cyclomatic    8   cognitive   27 lines
         110.0 ! CRAP
  frontend/src/features/auth-panel/AuthPanel.tsx
    :37 AuthPanel CRITICAL
          10   cyclomatic   13   cognitive  139 lines
         react: 5 hooks (3 state, 1 effect, 1 custom), max effect deps 1, JSX depth 4
         blast radius: <AuthPanel> rendered in 2 places
         110.0 ! CRAP
  frontend/src/lib/body.ts
    :83 mimeForFile CRITICAL
          10   cyclomatic    1   cognitive   24 lines
         110.0 ! CRAP
  frontend/src/lib/response.ts
    :247 isTabular CRITICAL
          10   cyclomatic    9   cognitive   21 lines
         110.0 ! CRAP
  frontend/src/features/spec-editor/EndpointEditor.tsx
    :17 EndpointEditor CRITICAL
          10   cyclomatic   13   cognitive  122 lines
         react: 3 props, 7 hooks (5 state, 2 memo), JSX depth 4
         110.0 ! CRAP
  frontend/src/features/workspace-home/HomeView.tsx
    :32 HomeView HIGH
           9   cyclomatic   18 ! cognitive  162 lines
         react: 12 hooks (1 effect, 11 custom), max effect deps 2, JSX depth 6
         blast radius: <Stat> rendered in 4 places
          90.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :20 ProxyPanel HIGH
           9   cyclomatic   10   cognitive  134 lines
         react: 4 hooks (4 custom), JSX depth 6
         blast radius: <ProxyPanel> rendered in 2 places
          90.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :274 closeTab HIGH
           9   cyclomatic    6   cognitive   26 lines
          90.0 ! CRAP
    :366 saveRequest HIGH
           9   cyclomatic    9   cognitive   26 lines
          90.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :174 aliasSubstitution HIGH
           9   cyclomatic    7   cognitive   15 lines
          90.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :178 <arrow> HIGH
           9   cyclomatic    8   cognitive   82 lines
         react: JSX depth 4
          90.0 ! CRAP
  frontend/src/lib/body.ts
    :61 encodeFormData HIGH
           9   cyclomatic   10   cognitive   20 lines
          90.0 ! CRAP
  frontend/src/lib/response.ts
    :177 indentXml HIGH
           9   cyclomatic   12   cognitive   20 lines
          90.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :65 onSave HIGH
           9   cyclomatic   12   cognitive   31 lines
          90.0 ! CRAP
  frontend/src/components/shell/ToolRail.tsx
    :63 ToolRail HIGH
           9   cyclomatic   10   cognitive  103 lines
         react: 2 props, 2 hooks (2 custom), JSX depth 5
          90.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :421 <arrow> HIGH
           9   cyclomatic    8   cognitive   31 lines
         react: JSX depth 2
          90.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :94 start HIGH
           9   cyclomatic    9   cognitive   38 lines
          90.0 ! CRAP
  frontend/src/stores/useGrpcStore.ts
    :154 <arrow> HIGH
           9   cyclomatic   10   cognitive   26 lines
          90.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :232 <arrow> HIGH
           9   cyclomatic    9   cognitive   67 lines
         react: JSX depth 4
          90.0 ! CRAP
  frontend/src/stores/useDocsStore.ts
    :58 generate HIGH
           9   cyclomatic    6   cognitive   21 lines
          90.0 ! CRAP
  frontend/src/components/shell/TopBar.tsx
    :36 TopBar HIGH
           8   cyclomatic   18 ! cognitive  165 lines
         react: 12 hooks (12 custom), JSX depth 8
          72.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :18 EnvironmentEditor HIGH
           8   cyclomatic   17 ! cognitive  190 lines
         react: 2 props, 11 hooks (4 state, 1 effect, 3 memo, 3 custom), max effect deps 3, JSX depth 4
          72.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :37 HistoryView HIGH
           8   cyclomatic   27 ! cognitive  295 lines
         react: 19 hooks (7 state, 1 effect, 11 custom), max effect deps 1, JSX depth 6
          72.0 ! CRAP
  frontend/src/lib/datasets.ts
    :24 parseCsvLine
           8   cyclomatic   19 ! cognitive   31 lines
  apps/desktop/frontend/src/bridge.ts
    :349 normalizeRunReport HIGH
           8   cyclomatic    7   cognitive   20 lines
          72.0 ! CRAP
    :717 endpoints HIGH
           8   cyclomatic   10   cognitive   16 lines
          72.0 ! CRAP
  frontend/src/features/runners-panel/DatasetPicker.tsx
    :10 DatasetPicker HIGH
           8   cyclomatic   15   cognitive  134 lines
         react: 9 hooks (2 state, 3 callback, 4 custom), JSX depth 5
          72.0 ! CRAP
  frontend/src/lib/crash.ts
    :98 formatReport HIGH
           8   cyclomatic   11   cognitive   41 lines
          72.0 ! CRAP
  frontend/src/features/test-runner/TestTab.tsx
    :174 buildAssertionLines HIGH
           8   cyclomatic    7   cognitive   11 lines
          72.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :134 openFolder HIGH
           8   cyclomatic   10   cognitive   33 lines
          72.0 ! CRAP
    :168 openDirect HIGH
           8   cyclomatic   10   cognitive   29 lines
          72.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-module-mocking.ts
    :25 isTestFrameworkObject HIGH
           8   cyclomatic    7   cognitive   25 lines
          72.0 ! CRAP
  frontend/src/lib/response.ts
    :300 binaryPreviewType HIGH
           8   cyclomatic    4   cognitive    6 lines
          72.0 ! CRAP
  frontend/src/features/perf-view/PerfView.tsx
    :6 PerfView HIGH
           8   cyclomatic   10   cognitive   98 lines
         react: 5 hooks (5 state), JSX depth 5
          72.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :115 <arrow> HIGH
           8   cyclomatic    7   cognitive   55 lines
         react: JSX depth 2
          72.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :203 <arrow> HIGH
           8   cyclomatic    7   cognitive   21 lines
         react: JSX depth 1
          72.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :128 run HIGH
           8   cyclomatic    7   cognitive   19 lines
         blast radius: <ChangesList> rendered in 2 places
          72.0 ! CRAP
  frontend/src/stores/useHistoryStore.ts
    :55 load HIGH
           8   cyclomatic    8   cognitive   16 lines
          72.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-known-value-widening.ts
    :58 hasKnownEvidence HIGH
           8   cyclomatic    6   cognitive   21 lines
          72.0 ! CRAP
  tools/oxlint/anti-slop/shared/lexical-type-parameters.ts
    :14 collectInferTypeParameterNames HIGH
           8   cyclomatic   12   cognitive   19 lines
          72.0 ! CRAP
  tools/oxlint/anti-slop/shared/reflect-method.ts
    :24 isGlobalReflectMethodCall HIGH
           8   cyclomatic    6   cognitive   12 lines
          72.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-type-aliases.ts
    :31 resolvesToUnknown HIGH
           8   cyclomatic    7   cognitive   17 lines
          72.0 ! CRAP
  frontend/src/lib/graphql.ts
    :37 gqlTypeRef HIGH
           8   cyclomatic    8   cognitive    9 lines
          72.0 ! CRAP
  frontend/src/components/RequestTabs.tsx
    :23 TabItem HIGH
           8   cyclomatic   15   cognitive  133 lines
         react: 3 props, 8 hooks (3 state, 5 custom), JSX depth 5
          72.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :32 EnvironmentsView HIGH
           7   cyclomatic   18 ! cognitive  302 lines
         react: 11 hooks (4 state, 1 effect, 6 custom), max effect deps 1, JSX depth 5
          56.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :331 normalizeRunStep HIGH
           7   cyclomatic    6   cognitive   17 lines
          56.0 ! CRAP
    :394 save HIGH
           7   cyclomatic    6   cognitive   25 lines
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :330 checkAssertion HIGH
           7   cyclomatic    4   cognitive   27 lines
          56.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :46 AuthEditor HIGH
           7   cyclomatic    8   cognitive   61 lines
         react: 3 props, JSX depth 3
         blast radius: <AuthFieldRow> rendered in 2 places
          56.0 ! CRAP
  frontend/src/stores/useCollectionRunStore.ts
    :61 startRun HIGH
           7   cyclomatic    8   cognitive   45 lines
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-object-parameters.ts
    :101 Program HIGH
           7   cyclomatic    7   cognitive   13 lines
          56.0 ! CRAP
  frontend/src/components/status.tsx
    :22 statusTier HIGH
           7   cyclomatic    6   cognitive    8 lines
         blast radius: <StatusPill> rendered in 3 places
          56.0 ! CRAP
  frontend/src/stores/useRealtimeStore.ts
    :74 <arrow> HIGH
           7   cyclomatic    6   cognitive   14 lines
          56.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :59 visibleGroups HIGH
           7   cyclomatic   10   cognitive   14 lines
         blast radius: <ImportDialog> rendered in 3 places
          56.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :236 finishSwitch HIGH
           7   cyclomatic    4   cognitive   23 lines
          56.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :63 filteredEndpoints HIGH
           7   cyclomatic    4   cognitive    7 lines
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-module-mocking.ts
    :41 <arrow> HIGH
           7   cyclomatic    5   cognitive    8 lines
          56.0 ! CRAP
  frontend/src/lib/response.ts
    :200 prettyBody HIGH
           7   cyclomatic    8   cognitive   14 lines
          56.0 ! CRAP
  frontend/src/features/settings-view/CicdPanel.tsx
    :10 CicdPanel HIGH
           7   cyclomatic   14   cognitive  170 lines
         react: 8 hooks (3 state, 5 custom), JSX depth 5
          56.0 ! CRAP
  frontend/src/components/shell/ToolRail.tsx
    :67 railButton HIGH
           7   cyclomatic    6   cognitive   35 lines
         react: JSX depth 2
          56.0 ! CRAP
  frontend/src/components/ui/toast.tsx
    :135 ToastIcon HIGH
           7   cyclomatic    6   cognitive   46 lines
         react: 1 props, JSX depth 1
          56.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :343 <arrow> HIGH
           7   cyclomatic    6   cognitive   17 lines
         react: JSX depth 3
          56.0 ! CRAP
  frontend/src/components/CollectionTree.tsx
    :251 CollectionTree HIGH
           7   cyclomatic   13   cognitive  107 lines
         react: 7 hooks (1 state, 1 memo, 5 custom), JSX depth 4
         blast radius: <CollectionBranch> rendered in 2 places
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-returns.ts
    :16 referencedAliasName HIGH
           7   cyclomatic    5   cognitive    9 lines
          56.0 ! CRAP
  frontend/src/features/workspace-bootstrap/WorkspaceEmptyState.tsx
    :13 WorkspaceEmptyState HIGH
           7   cyclomatic   13   cognitive   83 lines
         react: 8 hooks (1 state, 7 custom), JSX depth 5
          56.0 ! CRAP
  frontend/src/features/monitor-view/MonitorView.tsx
    :15 MonitorView HIGH
           7   cyclomatic   10   cognitive  184 lines
         react: 4 hooks (3 state, 1 effect), max effect deps 1, JSX depth 8
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-type-aliases.ts
    :5 referencedAliasName HIGH
           7   cyclomatic    5   cognitive    9 lines
          56.0 ! CRAP
    :50 Program HIGH
           7   cyclomatic    8   cognitive   18 lines
          56.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unsafe-dictionary-type.ts
    :75 shouldReportType HIGH
           7   cyclomatic    7   cognitive   11 lines
          56.0 ! CRAP
  frontend/src/components/JsonTree.tsx
    :84 Row HIGH
           7   cyclomatic    5   cognitive   22 lines
         react: 4 props, JSX depth 2
         blast radius: <JsonNode> rendered in 2 places
          56.0 ! CRAP
  frontend/src/features/jwt-inspector/JwtInspector.tsx
    :75 JwtInspector
           6   cyclomatic    9   cognitive  119 lines
         react: 4 hooks (4 state), JSX depth 6
         blast radius: <ClaimsTable> rendered in 2 places
          42.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :452 exportReport
           6   cyclomatic    5   cognitive   18 lines
          42.0 ! CRAP
    :499 show
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
    :895 invoke
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :115 <arrow>
           6   cyclomatic    3   cognitive    6 lines
         blast radius: <ProxyPanel> rendered in 2 places
          42.0 ! CRAP
    :128 <arrow>
           6   cyclomatic    3   cognitive    6 lines
         blast radius: <ProxyPanel> rendered in 2 places
          42.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :83 onSave
           6   cyclomatic    7   cognitive   22 lines
          42.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :39 isBroadRecordKeyType
           6   cyclomatic    4   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/features/test-runner/TestTab.tsx
    :187 insertIntoFirstAssertionsBlock
           6   cyclomatic    4   cognitive   14 lines
          42.0 ! CRAP
  tools/oxlint/anti-slop/effect/rules/no-service-constructor-imports.ts
    :34 ImportDeclaration
           6   cyclomatic    7   cognitive   16 lines
          42.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-object-parameters.ts
    :17 parameterAnnotation
           6   cyclomatic    5   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/stores/useRealtimeStore.ts
    :124 sendBinary
           6   cyclomatic    5   cognitive   17 lines
          42.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :145 bodyTypeFor
           6   cyclomatic    5   cognitive   14 lines
          42.0 ! CRAP
    :338 duplicateTab
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
    :393 overwriteRequest
           6   cyclomatic    5   cognitive   26 lines
          42.0 ! CRAP
  frontend/src/lib/ui.ts
    :26 handleTabArrowKeys
           6   cyclomatic    6   cognitive   21 lines
          42.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :467 isPopulatedObjectExpression
           6   cyclomatic    3   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/components/CreateWorkspaceModal.tsx
    :26 handlePickFolder
           6   cyclomatic    7   cognitive   14 lines
          42.0 ! CRAP
  frontend/src/lib/response.ts
    :45 suggestedFilename
           6   cyclomatic    3   cognitive   13 lines
          42.0 ! CRAP
    :312 normalizeHeaderKeys
           6   cyclomatic    7   cognitive   11 lines
          42.0 ! CRAP
    :360 searchBody
           6   cyclomatic    7   cognitive   22 lines
          42.0 ! CRAP
  frontend/src/components/ErrorBoundary.tsx
    :63 render
           6   cyclomatic    5   cognitive   35 lines
         react: JSX depth 3
         blast radius: <CopyReportButton> rendered in 32 places
          42.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :770 VariablesView
           6   cyclomatic    5   cognitive   42 lines
         react: 3 props, JSX depth 2
          42.0 ! CRAP
  frontend/src/components/shell/StatusBar.tsx
    :5 StatusBar
           6   cyclomatic    7   cognitive   53 lines
         react: 2 hooks (2 custom), JSX depth 4
          42.0 ! CRAP
  frontend/src/components/WorkspaceSidebar.tsx
    :31 WorkspaceSidebar
           6   cyclomatic   15   cognitive  142 lines
         react: 10 hooks (1 state, 9 custom), JSX depth 6
          42.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :21 SecretsEditor
           6   cyclomatic   15   cognitive  199 lines
         react: 3 props, 10 hooks (4 state, 1 effect, 2 memo, 3 custom), max effect deps 3, JSX depth 4
          42.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-parameters.ts
    :14 parameterAnnotation
           6   cyclomatic    5   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :71 ChangeRow
           6   cyclomatic    6   cognitive   33 lines
         react: 2 props, 1 hooks (1 state), JSX depth 4
         blast radius: <ChangesList> rendered in 2 places
          42.0 ! CRAP
  frontend/src/stores/useHistoryStore.ts
    :107 replayWithVars
           6   cyclomatic    7   cognitive   17 lines
          42.0 ! CRAP
  frontend/src/lib/jsonpath.ts
    :134 stripQuotes
           6   cyclomatic    5   cognitive   10 lines
          42.0 ! CRAP
  frontend/src/stores/useImportStore.ts
    :104 runPreview
           6   cyclomatic    6   cognitive   22 lines
          42.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-known-value-widening.ts
    :15 unwrapExpression
           6   cyclomatic    2   cognitive   13 lines
          42.0 ! CRAP
    :89 enclosingFunction
           6   cyclomatic    5   cognitive   14 lines
          42.0 ! CRAP
    :110 functionName
           6   cyclomatic    5   cognitive    9 lines
          42.0 ! CRAP
    :200 AssignmentExpression
           6   cyclomatic    5   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/features/spec-editor/SpecEditorView.tsx
    :69 handleGenerate
           6   cyclomatic    6   cognitive   22 lines
          42.0 ! CRAP
  frontend/src/features/grpc-view/GrpcTab.tsx
    :12 statusBadge
           6   cyclomatic    1   cognitive   23 lines
         react: JSX depth 2
          42.0 ! CRAP
  frontend/src/components/RequestTabs.tsx
    :173 <arrow>
           6   cyclomatic    6   cognitive   12 lines
          42.0 ! CRAP
  frontend/src/stores/useExportStore.ts
    :68 run
           5   cyclomatic    5   cognitive   19 lines
          30.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :371 load
           5   cyclomatic    4   cognitive   16 lines
          30.0 ! CRAP
    :458 steps
           5   cyclomatic    4   cognitive    8 lines
          30.0 ! CRAP
    :600 toDiffResultView
           5   cyclomatic    2   cognitive   17 lines
          30.0 ! CRAP
    :629 toGqlType
           5   cyclomatic    4   cognitive   21 lines
          30.0 ! CRAP
    :619 introspect
           5   cyclomatic    4   cognitive   40 lines
         react: 3 props
          30.0 ! CRAP
    :807 responses
           5   cyclomatic    4   cognitive    9 lines
          30.0 ! CRAP
    :872 generate
           5   cyclomatic    4   cognitive   15 lines
          30.0 ! CRAP
  frontend/src/components/ThemeToggle.tsx
    :7 ThemeToggle
           5   cyclomatic    8   cognitive   23 lines
         react: 2 hooks (2 custom), JSX depth 2
          30.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :155 TlsSecurityPanel
           5   cyclomatic    8   cognitive  103 lines
         react: 4 hooks (4 custom), JSX depth 5
         blast radius: <ProxyPanel> rendered in 2 places
          30.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :36 dirty
           5   cyclomatic    5   cognitive   10 lines
          30.0 ! CRAP
    :52 duplicateKey
           5   cyclomatic    6   cognitive   11 lines
          30.0 ! CRAP
    :67 secretLikeWarnings
           5   cyclomatic    6   cognitive   12 lines
          30.0 ! CRAP
  frontend/src/lib/paletteProviders.ts
    :67 getItems
           5   cyclomatic    6   cognitive   32 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-widen-then-assert.ts
    :84 broadTypeKind
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
    :301 assertionIsNarrower
           5   cyclomatic    4   cognitive   12 lines
          30.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :111 AuthFieldRow
           5   cyclomatic    4   cognitive   41 lines
         react: 4 props, JSX depth 3
         blast radius: <AuthFieldRow> rendered in 2 places
          30.0 ! CRAP
  frontend/src/features/test-runner/TestTab.tsx
    :215 <arrow>
           5   cyclomatic    2   cognitive    5 lines
          30.0 ! CRAP
  frontend/src/stores/useCollectionRunStore.ts
    :80 onEvent
           5   cyclomatic    2   cognitive   14 lines
          30.0 ! CRAP
    :107 cancelRun
           5   cyclomatic    6   cognitive   11 lines
          30.0 ! CRAP
    :119 exportReport
           5   cyclomatic    4   cognitive   20 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-object-parameters.ts
    :83 checkParameters
           5   cyclomatic    6   cognitive   16 lines
          30.0 ! CRAP
  frontend/src/hooks/useKeyboardMap.ts
    :20 isTypingTarget
           5   cyclomatic    3   cognitive    6 lines
          30.0 ! CRAP
  frontend/src/components/status.tsx
    :35 StatusPill
           5   cyclomatic    4   cognitive   28 lines
         react: 2 props, JSX depth 2
         blast radius: <StatusPill> rendered in 3 places
          30.0 ! CRAP
  frontend/src/features/git-view/GitView.tsx
    :13 load
           5   cyclomatic    4   cognitive   13 lines
          30.0 ! CRAP
  frontend/src/stores/useRealtimeStore.ts
    :63 connect
           5   cyclomatic    5   cognitive   46 lines
          30.0 ! CRAP
    :110 send
           5   cyclomatic    4   cognitive   13 lines
          30.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :184 baseUrlFor
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
    :420 reloadRequest
           5   cyclomatic    4   cognitive   23 lines
          30.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :45 groups
           5   cyclomatic    3   cognitive   10 lines
         blast radius: <ImportDialog> rendered in 3 places
          30.0 ! CRAP
  frontend/src/components/shell/ContextSidebar.tsx
    :40 openTestTab
           5   cyclomatic    5   cognitive    7 lines
         blast radius: <SectionLabel> rendered in 13 places
          30.0 ! CRAP
    :160 MocksContext
           5   cyclomatic   11   cognitive   81 lines
         react: 7 hooks (7 custom), JSX depth 4
         blast radius: <SectionLabel> rendered in 13 places
          30.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :44 getStoredRecentWorkspaces
           5   cyclomatic    4   cognitive   11 lines
          30.0 ! CRAP
    :110 init
           5   cyclomatic    5   cognitive   23 lines
          30.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :92 onReplayWithVars
           5   cyclomatic    5   cognitive    9 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/shared/dictionary-types.ts
    :106 isUnappliedReferenceTo
           5   cyclomatic    2   cognitive   10 lines
          30.0 ! CRAP
    :132 isEffectivelyEmptyMember
           5   cyclomatic    1   cognitive    9 lines
          30.0 ! CRAP
    :146 isEffectivelyEmptyInterface
           5   cyclomatic    3   cognitive   11 lines
          30.0 ! CRAP
    :158 resolvedSubstitutionArgument
           5   cyclomatic    4   cognitive   15 lines
          30.0 ! CRAP
  frontend/src/components/CreateWorkspaceModal.tsx
    :41 handleCreate
           5   cyclomatic    5   cognitive   16 lines
          30.0 ! CRAP
  frontend/src/lib/authSchemes.ts
    :274 authForScheme
           5   cyclomatic    3   cognitive   11 lines
          30.0 ! CRAP
  frontend/src/components/ErrorBoundary.tsx
    :114 <arrow>
           5   cyclomatic    4   cognitive   15 lines
         blast radius: <CopyReportButton> rendered in 32 places
          30.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :133 onKeyDown
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
    :233 <arrow>
           5   cyclomatic    5   cognitive    6 lines
          30.0 ! CRAP
    :293 <arrow>
           5   cyclomatic    2   cognitive    3 lines
          30.0 ! CRAP
    :416 <arrow>
           5   cyclomatic    4   cognitive   21 lines
         react: JSX depth 1
          30.0 ! CRAP
    :653 settingsSummary
           5   cyclomatic    4   cognitive    7 lines
          30.0 ! CRAP
    :661 retrySummary
           5   cyclomatic    4   cognitive    5 lines
          30.0 ! CRAP
  frontend/src/components/shell/BottomPanel.tsx
    :190 <arrow>
           5   cyclomatic    3   cognitive   11 lines
         react: JSX depth 2
          30.0 ! CRAP
  frontend/src/components/WorkspaceSidebar.tsx
    :181 openTestTab
           5   cyclomatic    5   cognitive    7 lines
          30.0 ! CRAP
  frontend/src/lib/crashReporter.ts
    :42 onKey
           5   cyclomatic    2   cognitive    9 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-parameters.ts
    :27 parameterName
           5   cyclomatic    4   cognitive   14 lines
          30.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :83 cookiesText
           5   cyclomatic    4   cognitive    9 lines
          30.0 ! CRAP
    :106 imageDataUrl
           5   cyclomatic    3   cognitive    4 lines
          30.0 ! CRAP
  frontend/src/features/command-palette/CommandPalette.tsx
    :4 CommandPalette
           5   cyclomatic   14   cognitive  102 lines
         react: 9 hooks (1 effect, 8 custom), max effect deps 1, JSX depth 5
          30.0 ! CRAP
  frontend/src/stores/useRequestStore.ts
    :129 tabIsDirty
           5   cyclomatic    2   cognitive    4 lines
          30.0 ! CRAP
    :196 send
           5   cyclomatic    5   cognitive   43 lines
          30.0 ! CRAP
  frontend/src/stores/useTestStore.ts
    :152 run
           5   cyclomatic    4   cognitive   28 lines
          30.0 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :25 statusBadge
           5   cyclomatic    1   cognitive   18 lines
         react: JSX depth 2
         blast radius: <RealtimeTab> rendered in 3 places
          30.0 ! CRAP
    :64 getPlaceholder
           5   cyclomatic    1   cognitive   14 lines
         blast radius: <RealtimeTab> rendered in 3 places
          30.0 ! CRAP
  frontend/src/features/request-editor/RequestSettingsDialog.tsx
    :43 RequestSettingsDialog
           5   cyclomatic    6   cognitive   82 lines
         react: 3 props, 2 hooks (2 state), JSX depth 5
          30.0 ! CRAP
  frontend/src/stores/useImportStore.ts
    :127 commit
           5   cyclomatic    5   cognitive   19 lines
          30.0 ! CRAP
  frontend/src/stores/useGrpcStore.ts
    :103 discover
           5   cyclomatic    5   cognitive   18 lines
          30.0 ! CRAP
    :156 <arrow>
           5   cyclomatic    4   cognitive    7 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-known-value-widening.ts
    :42 variableDeclarator
           5   cyclomatic    3   cognitive    7 lines
          30.0 ! CRAP
    :149 reportFlow
           5   cyclomatic    4   cognitive   19 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/require-safety-comment-for-type-assertion.ts
    :23 hasSafetyComment
           5   cyclomatic    6   cognitive   14 lines
          30.0 ! CRAP
  frontend/src/features/mock-view/MocksView.tsx
    :227 <arrow>
           5   cyclomatic    4   cognitive   95 lines
         react: JSX depth 4
          30.0 ! CRAP
    :351 <arrow>
           5   cyclomatic    4   cognitive   44 lines
         react: JSX depth 5
          30.0 ! CRAP
  tools/oxlint/anti-slop/shared/reflect-method.ts
    :16 isGlobalReflect
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unknown-returns.ts
    :93 Program
           5   cyclomatic    5   cognitive   10 lines
          30.0 ! CRAP
  frontend/src/features/import-dialog/ImportReportView.tsx
    :46 groups
           5   cyclomatic    4   cognitive   14 lines
         blast radius: <ImportReportView> rendered in 2 places
          30.0 ! CRAP
    :45 ImportReportView
           5   cyclomatic    4   cognitive   74 lines
         react: 1 props, 1 hooks (1 memo), JSX depth 1
         blast radius: <ImportReportView> rendered in 2 places
          30.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-unsafe-dictionary-type.ts
    :69 isPlainAliasConsumerUse
           5   cyclomatic    3   cognitive    5 lines
          30.0 ! CRAP
    :119 TSIndexSignature
           5   cyclomatic    3   cognitive   13 lines
          30.0 ! CRAP
  Functions exceeding cyclomatic, cognitive, or CRAP thresholds; ! marks the dimension that breached (https://docs.fallow.tools/explanations/health#complexity-metrics)
  To suppress: // fallow-ignore-next-line complexity

● File health scores (193 files) · sorted by triage concern

   81.3    frontend/src/features/response-viewer/ResponseViewer.tsx  risk
            572 LOC    2 fan-in   12 fan-out    0% dead  0.28 density  >999 risk

   78.8    frontend/src/features/request-editor/RequestEditor.tsx  risk
            811 LOC    2 fan-in   20 fan-out    0% dead  0.30 density  >999 risk

   80.9    frontend/src/lib/codegen.ts                     risk
             78 LOC    1 fan-in    2 fan-out    0% dead  0.49 density  >999 risk

   80.8    frontend/src/features/import-dialog/ImportDialog.tsx  risk
            356 LOC    4 fan-in   14 fan-out    0% dead  0.28 density  >999 risk

   68.6    apps/desktop/frontend/src/bridge.ts             risk
            984 LOC    1 fan-in    1 fan-out   95% dead  0.32 density  992.0 risk

   84.1    frontend/src/features/mock-view/MocksView.tsx   risk
            544 LOC    2 fan-in   11 fan-out    0% dead  0.20 density  870.0 risk

   83.2    frontend/src/hooks/useKeyboardMap.ts            risk
            111 LOC    1 fan-in    5 fan-out    0% dead  0.32 density  756.0 risk

   86.6    frontend/src/lib/body.ts                        risk
            177 LOC    5 fan-in    2 fan-out    0% dead  0.30 density  702.0 risk

   88.0    tools/oxlint/anti-slop/rules/no-widen-then-assert.ts  risk
            367 LOC    2 fan-in    0 fan-out    0% dead  0.40 density  702.0 risk

   80.6    frontend/src/app/App.tsx                        risk
            325 LOC    2 fan-in   40 fan-out    0% dead  0.15 density  650.0 risk

  ... and 183 more files (--format json for full list)

  Sorted by triage concern: the larger of low-MI concern and CRAP risk. The risk / structure tag marks which one placed each file. MI reflects complexity, coupling, and dead code; risk reflects untested complexity (CRAP) and can diverge from MI. Risk: low <15, moderate 15-30, high >=30. CRAP estimated from export references (85% direct, 40% indirect, 0% untested). Run `fallow health --coverage <coverage-final.json>` for exact scores. https://docs.fallow.tools/explanations/health#file-health-scores

● Hotspots (70 files, since 6 months)

   64.0 ▼  apps/desktop/frontend/src/bridge.ts
          35 commits   1912 churn  0.32 density   1 fan-in  ▼ cooling

   52.9 ▼  frontend/src/features/request-editor/RequestEditor.tsx
          31 commits   2425 churn  0.30 density   2 fan-in  ▼ cooling

   34.1 ─  frontend/src/stores/useWorkspaceStore.ts
          24 commits   1483 churn  0.25 density  27 fan-in  ─ stable

   29.4 ▲  frontend/src/app/App.tsx
          34 commits   1682 churn  0.15 density   2 fan-in  ▲ accelerating

   28.7 ▼  frontend/src/features/response-viewer/ResponseViewer.tsx
          18 commits   1815 churn  0.28 density   2 fan-in  ▼ cooling

   23.8 ▲  frontend/src/components/WorkspaceSidebar.tsx
          19 commits    504 churn  0.22 density   1 fan-in  ▲ accelerating

   15.2 ▼  frontend/src/lib/response.ts
           8 commits    429 churn  0.34 density  10 fan-in  ▼ cooling

   15.1 ─  frontend/src/lib/specTree.ts
           7 commits    633 churn  0.36 density   4 fan-in  ─ stable

   14.7 ▼  frontend/src/components/RequestTabs.tsx
          11 commits    655 churn  0.23 density   2 fan-in  ▼ cooling

   13.7 ▼  frontend/src/lib/request.ts
          16 commits    394 churn  0.15 density  12 fan-in  ▼ cooling

   12.5 ─  frontend/src/components/shell/BottomPanel.tsx
           8 commits    347 churn  0.26 density   1 fan-in  ─ stable

   12.5 ▼  frontend/src/stores/useRequestStore.ts
          13 commits    789 churn  0.17 density  14 fan-in  ▼ cooling

   12.4 ─  frontend/src/components/shell/ContextSidebar.tsx
           9 commits    623 churn  0.23 density   2 fan-in  ─ stable

   12.4 ▼  frontend/src/features/environments-view/EnvironmentsView.tsx
          12 commits    977 churn  0.18 density   2 fan-in  ▼ cooling

   11.9 ▼  frontend/src/lib/body.ts
           7 commits    218 churn  0.30 density   5 fan-in  ▼ cooling

   11.8 ▲  frontend/src/components/CollectionTree.tsx
           8 commits    877 churn  0.25 density   3 fan-in  ▲ accelerating

   11.5 ▼  frontend/src/features/import-dialog/ImportDialog.tsx
           7 commits    467 churn  0.28 density   4 fan-in  ▼ cooling

   11.5 ▼  frontend/src/features/environments-view/EnvironmentEditor.tsx
           8 commits    565 churn  0.25 density   2 fan-in  ▼ cooling

   11.1 ▼  frontend/src/components/shell/TopBar.tsx
          11 commits    508 churn  0.17 density   2 fan-in  ▼ cooling

   10.7 ▼  frontend/src/features/spec-editor/SpecEditorView.tsx
           6 commits    261 churn  0.30 density   1 fan-in  ▼ cooling

    9.5 ▼  frontend/src/hooks/useKeyboardMap.ts
           5 commits    146 churn  0.32 density   1 fan-in  ▼ cooling

    9.0 ▲  frontend/src/features/command-palette/CommandPalette.tsx
           6 commits    199 churn  0.25 density   1 fan-in  ▲ accelerating

    8.9 ▼  frontend/src/features/realtime-pages/RealtimePage.tsx
           3 commits     67 churn  0.50 density   1 fan-in  ▼ cooling

    8.7 ▼  frontend/src/stores/useCommandPaletteStore.ts
           3 commits    140 churn  0.49 density   6 fan-in  ▼ cooling

    8.4 ─  frontend/src/features/settings-view/SettingsView.tsx
          10 commits    464 churn  0.14 density   1 fan-in  ─ stable

    8.2 ▼  frontend/src/features/auth-editor/AuthEditor.tsx
           8 commits    463 churn  0.18 density   2 fan-in  ▼ cooling

    7.8 ▼  frontend/src/features/auth-panel/AuthPanel.tsx
           9 commits    477 churn  0.15 density   3 fan-in  ▼ cooling

    7.6 ▼  frontend/src/features/realtime-view/RealtimeTab.tsx
           5 commits    284 churn  0.26 density   3 fan-in  ▼ cooling

    7.6 ▼  frontend/src/features/environments-view/SecretsEditor.tsx
           6 commits    577 churn  0.22 density   2 fan-in  ▼ cooling

    7.5 ─  frontend/src/features/runners-panel/RunnersPanel.tsx
           8 commits    863 churn  0.16 density   2 fan-in  ─ stable

    7.2 ─  frontend/src/lib/schemaGraph.ts
           4 commits     61 churn  0.30 density   3 fan-in  ─ stable

    7.1 ▼  frontend/src/features/mock-view/MocksView.tsx
           6 commits    921 churn  0.20 density   2 fan-in  ▼ cooling

    7.0 ▼  frontend/src/features/history-view/HistoryView.tsx
           6 commits    529 churn  0.20 density   1 fan-in  ▼ cooling

    6.7 ▼  frontend/src/components/RunView.tsx
           5 commits    336 churn  0.23 density   2 fan-in  ▼ cooling

    6.7 ▼  frontend/src/features/git-view/GitView.tsx
           3 commits     98 churn  0.37 density   0 fan-in  ▼ cooling

    6.3 ▼  frontend/src/components/KeyValueEditor.tsx
           6 commits    204 churn  0.18 density   3 fan-in  ▼ cooling

    6.3 ▼  frontend/src/stores/useHistoryStore.ts
           5 commits    172 churn  0.22 density   8 fan-in  ▼ cooling

    6.1 ▼  frontend/src/lib/authSchemes.ts
           9 commits    395 churn  0.12 density   2 fan-in  ▼ cooling

    6.0 ▼  frontend/src/features/workspace-home/HomeView.tsx
           5 commits    243 churn  0.20 density   2 fan-in  ▼ cooling

    5.9 ▲  frontend/src/lib/themes.ts
           3 commits    185 churn  0.33 density   4 fan-in  ▲ accelerating

    5.8 ▼  frontend/src/features/diff-view/DiffView.tsx
           4 commits    373 churn  0.25 density   2 fan-in  ▼ cooling

    5.5 ─  frontend/src/stores/useWorkspaceBootstrap.ts
           4 commits    297 churn  0.23 density   8 fan-in  ─ stable

    5.4 ▼  frontend/src/components/shell/ToolRail.tsx
           7 commits    351 churn  0.13 density   2 fan-in  ▼ cooling

    5.1 ─  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
           4 commits    414 churn  0.22 density   2 fan-in  ─ stable

    5.0 ▼  frontend/src/lib/paletteProviders.ts
           3 commits    126 churn  0.28 density   1 fan-in  ▼ cooling

    4.8 ▼  frontend/src/features/settings-view/ProxyTlsPanel.tsx
           4 commits    302 churn  0.20 density   1 fan-in  ▼ cooling

    4.8 ─  frontend/src/stores/useCollectionRunStore.ts
           4 commits    206 churn  0.21 density   3 fan-in  ─ stable

    4.8 ▼  frontend/src/components/shell/StatusBar.tsx
           5 commits    129 churn  0.16 density   2 fan-in  ▼ cooling

    4.4 ▼  frontend/src/lib/bottomPanel.ts
           3 commits     19 churn  0.25 density   2 fan-in  ▼ cooling

    4.3 ▲  frontend/src/stores/useThemeStore.ts
           4 commits    146 churn  0.19 density   5 fan-in  ▲ accelerating

    4.2 ▲  frontend/src/components/JsonTree.tsx
           3 commits    131 churn  0.25 density   1 fan-in  ▲ accelerating

    4.1 ▼  frontend/src/features/docs-view/DocsView.tsx
           3 commits    434 churn  0.23 density   2 fan-in  ▼ cooling

    4.0 ▼  frontend/src/lib/jsonpath.ts
           3 commits    163 churn  0.24 density   2 fan-in  ▼ cooling

    3.8 ▲  frontend/src/stores/useDatasetStore.ts
           3 commits     91 churn  0.21 density   3 fan-in  ▲ accelerating

    3.7 ▲  frontend/src/features/graphql-browser/GraphqlBrowser.tsx
           3 commits    287 churn  0.21 density   2 fan-in  ▲ accelerating

    3.7 ▼  frontend/src/lib/env.ts
           6 commits     67 churn  0.11 density   2 fan-in  ▼ cooling

    3.6 ▲  frontend/src/features/env-tools/EnvToolsPanel.tsx
           3 commits    283 churn  0.21 density   2 fan-in  ▲ accelerating

    3.6 ─  frontend/src/features/settings-view/CicdPanel.tsx
           4 commits    199 churn  0.15 density   1 fan-in  ─ stable

    3.5 ─  frontend/src/lib/realtime.ts
           4 commits     67 churn  0.15 density   3 fan-in  ─ stable

    3.4 ▲  frontend/src/features/dep-graph/DepGraphView.tsx
           3 commits    175 churn  0.19 density   0 fan-in  ▲ accelerating

    3.3 ▼  frontend/src/lib/crash.ts
           3 commits    140 churn  0.19 density   5 fan-in  ▼ cooling

    3.3 ▼  frontend/src/features/monitor-view/MonitorView.tsx
           4 commits    322 churn  0.14 density   0 fan-in  ▼ cooling

    3.3 ▼  frontend/src/features/workspace-bootstrap/WorkspaceEmptyState.tsx
           3 commits    105 churn  0.19 density   1 fan-in  ▼ cooling

    3.1 ▲  frontend/src/features/spec-editor/EndpointEditor.tsx
           3 commits    144 churn  0.17 density   1 fan-in  ▲ accelerating

    3.1 ▲  frontend/src/lib/mock.ts
           4 commits    176 churn  0.13 density   4 fan-in  ▲ accelerating

    2.9 ▼  frontend/src/features/perf-view/PerfView.tsx
           3 commits    252 churn  0.16 density   0 fan-in  ▼ cooling

    2.9 ▲  frontend/src/editors/CodeMirrorEditor.tsx
           4 commits    226 churn  0.13 density   7 fan-in  ▲ accelerating

    2.6 ▼  frontend/src/lib/history.ts
           3 commits     60 churn  0.15 density   5 fan-in  ▼ cooling

    1.6 ▼  frontend/src/components/shell/AppShell.tsx
           3 commits    324 churn  0.09 density   1 fan-in  ▼ cooling

    1.4 ▼  frontend/src/lib/collections.ts
           8 commits    239 churn  0.03 density  10 fan-in  ▼ cooling

  123 files excluded (< 3 commits)

  Files with high churn and high complexity: https://docs.fallow.tools/explanations/health#hotspot-metrics

● Refactoring targets (39)
  32 medium · 7 high
    score = quick-win ROI (higher = better) · pri = absolute priority

   17.8  pri:35.5    frontend/src/features/git-view/GitView.tsx
         untested risk · effort:medium · confidence:high  2 complex functions lack test coverage path, add tests before modifying

   15.6  pri:46.9    apps/desktop/frontend/src/bridge.ts
         dead code · effort:high · confidence:high  Remove 18 unused exports to reduce surface area (95% dead)
         importers: apps/desktop/frontend/src/main.tsx (initRequestBridge)

   15.5  pri:31.0    frontend/src/lib/typeGuards.ts
         dead code · effort:medium · confidence:high  Remove 3 unused exports to reduce surface area (50% dead)
         importers: frontend/src/components/JsonTree.tsx (JsonObject, JsonValue, isRecord, isString); frontend/src/features/auth-panel/AuthPanel.tsx (JsonObject, JsonValue, isRecord, isString); frontend/src/features/response-viewer/ResponseViewer.tsx (JsonValue, isRecord); frontend/src/lib/body.ts (JsonValue); frontend/src/lib/jsonpath.ts (JsonObject, JsonValue, isRecord)

   14.3  pri:28.5    frontend/src/lib/response.ts
         high impact · effort:medium · confidence:medium  Split high-impact file (382 LOC), 10 dependents amplify every change
         importers: frontend/src/components/JsonTree.tsx (jsonText); frontend/src/components/shell/BottomPanel.tsx (ResponseCookie, parseSetCookies); frontend/src/features/docs-view/DocsView.tsx (copyText); frontend/src/features/realtime-view/RealtimeTab.tsx (bytesToBase64); frontend/src/features/request-editor/RequestEditor.tsx (copyText)

   13.4  pri:26.7    frontend/src/features/import-dialog/ImportDialog.tsx
         complexity · effort:medium · confidence:high  Extract ImportDialog (cognitive: 45) in 356-LOC file into smaller functions
         importers: frontend/src/components/WorkspaceSidebar.tsx (ImportDialog); frontend/src/components/shell/TopBar.tsx (ImportDialog); frontend/src/features/index.ts (side effect); frontend/src/features/workspace-home/HomeView.tsx (ImportDialog)

   12.6  pri:25.1    frontend/src/stores/useCommandPaletteStore.ts
         high impact · effort:medium · confidence:medium  Split high-impact file (117 LOC), 6 dependents amplify every change
         importers: frontend/src/components/shell/TopBar.tsx (useCommandPaletteStore); frontend/src/features/command-palette/CommandPalette.tsx (getFilteredResults, groupByHint, useCommandPaletteStore); frontend/src/hooks/useKeyboardMap.ts (useCommandPaletteStore); frontend/src/lib/paletteProviders.ts (useCommandPaletteStore); frontend/src/stores/index.ts (side effect)

   12.5  pri:37.4    frontend/src/stores/useWorkspaceStore.ts
         circular dependency · effort:high · confidence:high  Break import cycle, 27 files depend on this, changes cascade through the cycle
         importers: frontend/src/app/App.tsx (useWorkspaceStore); frontend/src/components/RequestTabs.tsx (RequestTab); frontend/src/components/WorkspaceSidebar.tsx (WorkspaceView, useWorkspaceStore); frontend/src/components/shell/BottomPanel.tsx (useWorkspaceStore); frontend/src/components/shell/ContextSidebar.tsx (useWorkspaceStore)
         clones: frontend/src/stores/useWorkspaceStore.ts:366-371 dup:6f87acd9; frontend/src/stores/useWorkspaceStore.ts:393-398 dup:6f87acd9

   12.4  pri:24.7    frontend/src/lib/bottomPanel.ts
         dead code · effort:medium · confidence:high  Remove 2 unused exports to reduce surface area (67% dead)
         importers: frontend/src/components/shell/BottomPanel.tsx (BOTTOM_PANELS, BottomPanelId); frontend/src/stores/useBottomPanelStore.ts (BottomPanelId)

   12.3  pri:24.6    frontend/src/app/App.tsx
         complexity · effort:medium · confidence:high  Extract App (cognitive: 211) in 325-LOC file into smaller functions
         importers: frontend/src/app/index.ts (side effect); frontend/src/app/main.tsx (App)

   12.2  pri:24.3    frontend/src/lib/specTree.ts
         high impact · effort:medium · confidence:medium  Split high-impact file (384 LOC), 4 dependents amplify every change
         importers: frontend/src/features/spec-editor/EndpointEditor.tsx (EndpointInput, validateEndpoint); frontend/src/features/spec-editor/SpecEditorView.tsx (EndpointInput, nodesForContent, patchEndpointInContent); frontend/src/lib/specTree.test.ts (diagnosticsForSpec, flattenSpecTree, nodesForContent, patchEndpointInContent, validateEndpoint); frontend/src/stores/useSpecEditorStore.ts (SpecDiagnostic, diagnosticsForSpec)

  ... and 29 more targets (--format json for full list)

  Prioritized refactoring recommendations based on complexity, churn, and coupling signals: https://docs.fallow.tools/explanations/health#refactoring-targets

✗ 255 above threshold · 2600 analyzed · maintainability 89.1 (good) (0.07s)
```

</details>

---

### 12. JS/TS Advisory Review (fallow review)

- **Command:** `nubx -y fallow review`
- **Status:** 🟢 PASS
- **Duration:** 7s

<details>
<summary>Click to expand full output</summary>

```text
nub: pnpm-workspace.yaml is not read under nub identity — migrate it (`nub pm use nub`), delete it, or return to pnpm (`nub pm use pnpm`).
nub 0.7.5
░░░░░░░░░░░░░░░   13/68 pkgs · resolving
███████████████    7/7 pkgs
✓ resolved 7 · reused 7 in 3.0s
dependencies:
+ fallow@3.21.0

loaded config: /home/satyajit/Documents/GitHub/OSS/api-client/reqly-main-brancn/.fallowrc.json
   0.679176491s  WARN Skipped 9 package.json entry points outside project root or containing parent directory traversal: /usr/local/bin/reqly (3x), /usr/local/bin/reqly-desktop (3x), /tmp/reqly (2x), ../frontend/bindings

Decisions to make (4):
  1. [public-api-contract] `frontend/src/lib/crash.ts` changes exports (APP_VERSION, addBreadcrumb, addGoLog, formatReport, getBreadcrumbs, getGoLogs, platformLabel) imported by 4 files outside this PR. Does this change break or alter what those callers expect?
     trade-off: 4 modules outside the diff consume this contract; changing its shape requires coordinating them.
  2. [public-api-contract] `frontend/src/lib/export.ts` changes exports (EXPORT_FORMAT_OPTIONS, exportFormatLabel, fallbackExportAdapter) imported by 3 files outside this PR. Does this change break or alter what those callers expect?
     trade-off: 3 modules outside the diff consume this contract; changing its shape requires coordinating them.
  3. [public-api-contract] `frontend/src/lib/typeGuards.ts` changes exports (isRecord, isString) imported by 3 files outside this PR. Does this change break or alter what those callers expect?
     trade-off: 3 modules outside the diff consume this contract; changing its shape requires coordinating them.
  4. [public-api-contract] `frontend/src/app/App.tsx` changes export (App) imported by 2 files outside this PR. Does this change break or alter what those callers expect?
     trade-off: 2 modules outside the diff consume this contract; changing its shape requires coordinating them.
  ... 39 more structural decisions collapsed below the cap of 4

Review brief (drill-down): 493 changed files vs d438b742cb9aae1b0db98b6ae351d3f78228b8a8 · risk high · effort deep-dive
  partition: 36 units (by module)
  review order: apps/desktop/frontend/src → frontend/src/components/ui → frontend/src/styles → frontend/src → tools/oxlint/anti-slop/rules → frontend/src/app → frontend/src/components → frontend/src/components/shell → frontend/src/features/auth-editor → frontend/src/features/auth-panel → frontend/src/features/command-palette → frontend/src/features/dep-graph → frontend/src/features/diff-view → frontend/src/features/docs-view → frontend/src/features/environments-view → frontend/src/features/git-view → frontend/src/features/graphql-browser → frontend/src/features/grpc-view → frontend/src/features/history-view → frontend/src/features/import-dialog → frontend/src/features/jwt-inspector → frontend/src/features/mock-view → frontend/src/features/monitor-view → frontend/src/features/openapi-explorer → frontend/src/features/perf-view → frontend/src/features/realtime-pages → frontend/src/features/realtime-view → frontend/src/features/request-editor → frontend/src/features/response-viewer → frontend/src/features/runners-panel → frontend/src/features/settings-view → frontend/src/features/spec-editor → frontend/src/features/workspace-home → frontend/src/hooks → frontend/src/lib → frontend/src/stores
  independent slices: 4 (no import edge between them) [apps/desktop/frontend/src] [frontend/src, frontend/src/styles] [frontend/src/app, frontend/src/components, frontend/src/components/shell, frontend/src/components/ui, frontend/src/features/auth-editor, frontend/src/features/auth-panel, frontend/src/features/command-palette, frontend/src/features/dep-graph, frontend/src/features/diff-view, frontend/src/features/docs-view, frontend/src/features/environments-view, frontend/src/features/git-view, frontend/src/features/graphql-browser, frontend/src/features/grpc-view, frontend/src/features/history-view, frontend/src/features/import-dialog, frontend/src/features/jwt-inspector, frontend/src/features/mock-view, frontend/src/features/monitor-view, frontend/src/features/openapi-explorer, frontend/src/features/perf-view, frontend/src/features/realtime-pages, frontend/src/features/realtime-view, frontend/src/features/request-editor, frontend/src/features/response-viewer, frontend/src/features/runners-panel, frontend/src/features/settings-view, frontend/src/features/spec-editor, frontend/src/features/workspace-home, frontend/src/hooks, frontend/src/lib, frontend/src/stores] [tools/oxlint/anti-slop/rules]
  impact closure: 37 files affected beyond the diff
  coordination gap: apps/desktop/frontend/src/main.tsx consumes initRequestBridge from apps/desktop/frontend/src/bridge.ts (not in this diff)
  coordination gap: frontend/src/app/main.tsx consumes App from frontend/src/app/App.tsx (not in this diff)
  coordination gap: frontend/src/index.ts consumes App from frontend/src/app/App.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes CollectionTree from frontend/src/components/CollectionTree.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes RequestTabs from frontend/src/components/RequestTabs.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes RunView from frontend/src/components/RunView.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes WorkspaceSidebar from frontend/src/components/WorkspaceSidebar.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes ContextSidebar from frontend/src/components/shell/ContextSidebar.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes StatusBar from frontend/src/components/shell/StatusBar.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes ToolRail from frontend/src/components/shell/ToolRail.tsx (not in this diff)
  coordination gap: frontend/src/components/index.ts consumes TopBar from frontend/src/components/shell/TopBar.tsx (not in this diff)
  coordination gap: frontend/src/lib/notify.ts consumes toast from frontend/src/components/ui/toast.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes AuthEditor from frontend/src/features/auth-editor/AuthEditor.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes AuthPanel from frontend/src/features/auth-panel/AuthPanel.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes DiffView from frontend/src/features/diff-view/DiffView.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes DocsView from frontend/src/features/docs-view/DocsView.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes EnvironmentEditor from frontend/src/features/environments-view/EnvironmentEditor.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes EnvironmentsView from frontend/src/features/environments-view/EnvironmentsView.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes SecretsEditor from frontend/src/features/environments-view/SecretsEditor.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes GraphqlBrowser from frontend/src/features/graphql-browser/GraphqlBrowser.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes GrpcTab from frontend/src/features/grpc-view/GrpcTab.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes ImportDialog from frontend/src/features/import-dialog/ImportDialog.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes JwtInspector from frontend/src/features/jwt-inspector/JwtInspector.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes MocksView from frontend/src/features/mock-view/MocksView.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes OpenapiExplorer from frontend/src/features/openapi-explorer/OpenapiExplorer.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes RealtimeTab from frontend/src/features/realtime-view/RealtimeTab.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes RequestEditor from frontend/src/features/request-editor/RequestEditor.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes ResponseViewer from frontend/src/features/response-viewer/ResponseViewer.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes RunnersPanel from frontend/src/features/runners-panel/RunnersPanel.tsx (not in this diff)
  coordination gap: frontend/src/features/index.ts consumes HomeView from frontend/src/features/workspace-home/HomeView.tsx (not in this diff)
  coordination gap: frontend/src/index.ts consumes bodyTypes, boundaryFor, contentTypeFor, encodeFormData, encodeUrlEncoded, serializeBody from frontend/src/lib/body.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes fallbackCollectionsAdapter from frontend/src/lib/collections.ts (not in this diff)
  coordination gap: frontend/src/components/ErrorBoundary.tsx consumes addBreadcrumb, formatReport from frontend/src/lib/crash.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes APP_VERSION, addBreadcrumb, addGoLog, formatReport, getBreadcrumbs, getGoLogs, platformLabel from frontend/src/lib/crash.ts (not in this diff)
  coordination gap: frontend/src/lib/crashReporter.ts consumes addBreadcrumb, formatReport from frontend/src/lib/crash.ts (not in this diff)
  coordination gap: frontend/src/stores/useCollectionRunStore.ts consumes addBreadcrumb from frontend/src/lib/crash.ts (not in this diff)
  coordination gap: frontend/src/features/export-dialog/ExportDialog.tsx consumes EXPORT_FORMAT_OPTIONS, exportFormatLabel from frontend/src/lib/export.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes EXPORT_FORMAT_OPTIONS, exportFormatLabel, fallbackExportAdapter from frontend/src/lib/export.ts (not in this diff)
  coordination gap: frontend/src/stores/useExportStore.ts consumes fallbackExportAdapter from frontend/src/lib/export.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes fallbackHistoryAdapter from frontend/src/lib/history.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes MOCK_METHOD_OPTIONS, createMockScenario, fallbackMockAdapter, headerLinesFrom, matchRoute, parseHeaderLines, pruneExpiredState from frontend/src/lib/mock.ts (not in this diff)
  coordination gap: frontend/src/stores/useMockStore.ts consumes fallbackMockAdapter, headerLinesFrom, parseHeaderLines, pruneExpiredState from frontend/src/lib/mock.ts (not in this diff)
  coordination gap: frontend/src/stores/useProxyTlsStore.ts consumes DEFAULT_PROXY, DEFAULT_TLS, validateProxy, validateTls from frontend/src/lib/proxyTls.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes MESSAGE_BUFFER_CAP, fallbackRealtimeAdapter, formatFrameTime, isRealtimeKind from frontend/src/lib/realtime.ts (not in this diff)
  coordination gap: frontend/src/stores/useRealtimeStore.ts consumes MESSAGE_BUFFER_CAP, fallbackRealtimeAdapter from frontend/src/lib/realtime.ts (not in this diff)
  coordination gap: frontend/src/index.ts consumes appendParams, fetchSender, sentRows from frontend/src/lib/request.ts (not in this diff)
  coordination gap: frontend/src/components/JsonTree.tsx consumes isRecord, isString from frontend/src/lib/typeGuards.ts (not in this diff)
  coordination gap: frontend/src/lib/jsonpath.ts consumes isRecord from frontend/src/lib/typeGuards.ts (not in this diff)
  coordination gap: frontend/src/lib/response.ts consumes isRecord from frontend/src/lib/typeGuards.ts (not in this diff)
  coordination gap: frontend/src/features/workspace-bootstrap/WorkspaceEmptyState.tsx consumes useThemeStore from frontend/src/stores/useThemeStore.ts (not in this diff)
  coordination gap: frontend/src/features/workspace-bootstrap/WorkspaceEmptyState.tsx consumes useWorkspaceBootstrapStore from frontend/src/stores/useWorkspaceBootstrap.ts (not in this diff)
  coordination gap: frontend/src/features/env-tools/EnvToolsPanel.tsx consumes useWorkspaceStore from frontend/src/stores/useWorkspaceStore.ts (not in this diff)
  coordination gap: frontend/src/stores/useCollectionRunStore.ts consumes useWorkspaceStore from frontend/src/stores/useWorkspaceStore.ts (not in this diff)
  coordination gap: tools/oxlint/anti-slop/index.ts consumes noRuntimeTypeofRule from tools/oxlint/anti-slop/rules/no-runtime-typeof.ts (not in this diff)
  focus: 87 units to review here (of 97 changed)
    [review-here] frontend/src/stores/useWorkspaceStore.ts: high fan-in (27 importers), fan-out 7, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/import-dialog/ImportDialog.tsx: high fan-in (4 importers), fan-out 14, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/stores/index.ts: high fan-in (7 importers), fan-out 20
      confidence low: re-export indirection
    [review-here] frontend/src/stores/useWorkspaceBootstrap.ts: high fan-in (8 importers), fan-out 3, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/body.ts: high fan-in (5 importers), fan-out 2, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/request.ts: high fan-in (12 importers), fan-out 2, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/stores/useThemeStore.ts: high fan-in (5 importers), fan-out 2, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/realtime-view/RealtimeTab.tsx: high fan-in (3 importers), fan-out 11, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/collections.ts: high fan-in (10 importers), fan-out 1, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/history.ts: high fan-in (5 importers), fan-out 1, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/stores/useRequestStore.ts: high fan-in (14 importers), fan-out 3
      confidence low: re-export indirection
    [review-here] frontend/src/components/CollectionTree.tsx: high fan-in (3 importers), fan-out 4, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/auth-panel/AuthPanel.tsx: high fan-in (3 importers), fan-out 4, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/crash.ts: high fan-in (5 importers), changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/typeGuards.ts: high fan-in (9 importers), changes a contract consumed outside the diff
    [review-here] frontend/src/app/App.tsx: high fan-in (2 importers), fan-out 40, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/PageHeader.tsx: high fan-in (10 importers), fan-out 1
    [review-here] frontend/src/components/RequestTabs.tsx: high fan-in (2 importers), fan-out 9, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/RunView.tsx: high fan-in (2 importers), fan-out 6, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/shell/ContextSidebar.tsx: high fan-in (2 importers), fan-out 13, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/shell/TopBar.tsx: high fan-in (2 importers), fan-out 12, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/auth-editor/AuthEditor.tsx: high fan-in (2 importers), fan-out 5, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/diff-view/DiffView.tsx: high fan-in (2 importers), fan-out 11, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/docs-view/DocsView.tsx: high fan-in (2 importers), fan-out 10, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/environments-view/EnvironmentsView.tsx: high fan-in (2 importers), fan-out 9, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/graphql-browser/GraphqlBrowser.tsx: high fan-in (2 importers), fan-out 8, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/grpc-view/GrpcTab.tsx: high fan-in (2 importers), fan-out 8, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/jwt-inspector/JwtInspector.tsx: high fan-in (2 importers), fan-out 10, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/mock-view/MocksView.tsx: high fan-in (2 importers), fan-out 11, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/openapi-explorer/OpenapiExplorer.tsx: high fan-in (2 importers), fan-out 10, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/request-editor/RequestEditor.tsx: high fan-in (2 importers), fan-out 20, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/response-viewer/ResponseViewer.tsx: high fan-in (2 importers), fan-out 12, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/runners-panel/RunnersPanel.tsx: high fan-in (2 importers), fan-out 13, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/workspace-home/HomeView.tsx: high fan-in (2 importers), fan-out 11, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/stores/useHistoryStore.ts: high fan-in (8 importers), fan-out 1
      confidence low: re-export indirection
    [review-here] frontend/src/stores/useSpecEditorStore.ts: high fan-in (5 importers), fan-out 1
      confidence low: re-export indirection
    [review-here] frontend/src/lib/mock.ts: high fan-in (4 importers), changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/proxyTls.ts: high fan-in (4 importers), changes a contract consumed outside the diff
    [review-here] frontend/src/stores/useCommandPaletteStore.ts: high fan-in (6 importers)
      confidence low: re-export indirection
    [review-here] frontend/src/components/KeyValueEditor.tsx: high fan-in (3 importers), fan-out 3
    [review-here] frontend/src/components/WorkspaceSidebar.tsx: high fan-in (1 importer), fan-out 12, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/shell/ToolRail.tsx: high fan-in (2 importers), fan-out 3, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/environments-view/EnvironmentEditor.tsx: high fan-in (2 importers), fan-out 3, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/features/environments-view/SecretsEditor.tsx: high fan-in (2 importers), fan-out 3, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/specTree.ts: high fan-in (4 importers), fan-out 1
    [review-here] frontend/src/components/shell/StatusBar.tsx: high fan-in (2 importers), fan-out 2, changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/components/ui/toast.tsx: high fan-in (2 importers), fan-out 2, changes a contract consumed outside the diff
    [review-here] frontend/src/lib/export.ts: high fan-in (3 importers), changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/realtime.ts: high fan-in (3 importers), changes a contract consumed outside the diff
      confidence low: re-export indirection
    [review-here] frontend/src/lib/themes.ts: high fan-in (4 importers)
    [review-here] frontend/src/components/shell/BottomPanel.tsx: high fan-in (1 importer), fan-out 9
    [review-here] frontend/src/features/history-view/HistoryView.tsx: high fan-in (1 importer), fan-out 11
    [review-here] frontend/src/features/request-editor/TemplatePickerSheet.tsx: high fan-in (1 importer), fan-out 5
    [review-here] frontend/src/features/settings-view/CicdPanel.tsx: high fan-in (1 importer), fan-out 6
    [review-here] frontend/src/features/settings-view/ProxyTlsPanel.tsx: high fan-in (1 importer), fan-out 5
    [review-here] frontend/src/features/settings-view/SettingsView.tsx: high fan-in (1 importer), fan-out 9
    [review-here] frontend/src/features/spec-editor/SpecEditorView.tsx: high fan-in (1 importer), fan-out 8
    [review-here] frontend/src/hooks/useKeyboardMap.ts: high fan-in (1 importer), fan-out 5
    [review-here] frontend/src/index.css: high fan-in (3 importers), fan-out 1
    [review-here] frontend/src/lib/paletteProviders.ts: high fan-in (1 importer), fan-out 7
    [review-here] frontend/src/stores/useDatasetStore.ts: high fan-in (3 importers), fan-out 1
    [review-here] frontend/src/components/CreateWorkspaceModal.tsx: high fan-in (1 importer), fan-out 4
    [review-here] frontend/src/components/shell/AppShell.tsx: high fan-in (1 importer), fan-out 4
    [review-here] frontend/src/features/request-editor/RequestSettingsDialog.tsx: high fan-in (1 importer), fan-out 4
    [review-here] frontend/src/features/runners-panel/DatasetPicker.tsx: high fan-in (1 importer), fan-out 4
    [review-here] frontend/src/lib/schemaGraph.ts: high fan-in (3 importers)
    [review-here] frontend/src/stores/useShellStore.ts: high fan-in (3 importers)
      confidence low: re-export indirection
    [review-here] tools/oxlint/anti-slop/rules/no-runtime-typeof.ts: high fan-in (2 importers), changes a contract consumed outside the diff
    [review-here] apps/desktop/frontend/src/bridge.ts: high fan-in (1 importer), fan-out 1, changes a contract consumed outside the diff
    [review-here] frontend/src/components/shell/EnvironmentSelector.tsx: high fan-in (1 importer), fan-out 3
    [review-here] frontend/src/components/status.tsx: high fan-in (2 importers), fan-out 1
    [review-here] frontend/src/features/realtime-pages/RealtimePage.tsx: high fan-in (1 importer), fan-out 3
    [review-here] frontend/src/features/spec-editor/EndpointEditor.tsx: high fan-in (1 importer), fan-out 3
    [review-here] frontend/src/lib/authSchemes.ts: high fan-in (2 importers), fan-out 1
    [review-here] frontend/src/components/ContextMenu.tsx: high fan-in (2 importers)
    [review-here] frontend/src/components/shell/storage.ts: high fan-in (2 importers)
    [review-here] frontend/src/features/monitor-view/MonitorView.tsx: fan-out 4
    [review-here] frontend/src/lib/bottomPanel.ts: high fan-in (2 importers)
    [review-here] frontend/src/lib/datasets.ts: high fan-in (2 importers)
    [review-here] frontend/src/lib/methodTint.ts: high fan-in (2 importers)
    [review-here] frontend/src/stores/useSettingsStore.ts: high fan-in (2 importers)
      confidence low: re-export indirection
    [review-here] frontend/src/components/ui/dropdown-menu.tsx: high fan-in (1 importer), fan-out 1
    [review-here] frontend/src/components/ui/menubar.tsx: high fan-in (1 importer), fan-out 1
    [review-here] frontend/src/features/command-palette/CommandPalette.tsx: high fan-in (1 importer), fan-out 1
    [review-here] frontend/src/features/dep-graph/DepGraphView.tsx: fan-out 3
    [review-here] frontend/src/features/git-view/GitView.tsx: fan-out 3
    [review-here] frontend/src/features/perf-view/PerfView.tsx: fan-out 3
  de-prioritized: 10 units (run with --show-deprioritized to list)
  weakening signals (56, reviewer-private, advisory):
    suppression added: fallow-ignore added (0 -> 6) in LINT_REPORT.md
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/encoding/json/jsontext/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/encoding/json/jsontext/models.js
    suppression added: eslint-disable added (0 -> 2) in apps/desktop/backend/frontend/bindings/encoding/json/models.js
    suppression added: @ts-ignore added (0 -> 2) in apps/desktop/backend/frontend/bindings/encoding/json/models.js
    suppression added: eslint-disable added (0 -> 16) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/appservice.js
    suppression added: @ts-ignore added (0 -> 16) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/appservice.js
    suppression added: eslint-disable added (0 -> 12) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/models.js
    suppression added: @ts-ignore added (0 -> 12) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/audit/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/audit/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/auth/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/auth/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/collab/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/collab/models.js
    suppression added: eslint-disable added (0 -> 2) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/core/models.js
    suppression added: @ts-ignore added (0 -> 2) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/core/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/diffing/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/diffing/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/environments/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/environments/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/graphql/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/graphql/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/grpc/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/grpc/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/history/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/history/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/importer/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/importer/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/jwt/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/jwt/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/openapi/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/openapi/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/perf/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/perf/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/policy/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/policy/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/rbac/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/rbac/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/request/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/request/models.js
    suppression added: eslint-disable added (0 -> 2) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/response/models.js
    suppression added: @ts-ignore added (0 -> 2) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/response/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/scim/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/scim/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/testing/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/testing/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/theme/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/theme/models.js
    suppression added: eslint-disable added (0 -> 3) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/workflow/models.js
    suppression added: @ts-ignore added (0 -> 3) in apps/desktop/backend/frontend/bindings/github.com/Its-Satyajit/reqly/internal/workflow/models.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/wailsapp/wails/v3/internal/eventcreate.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/github.com/wailsapp/wails/v3/internal/eventcreate.js
    suppression added: eslint-disable added (0 -> 1) in apps/desktop/backend/frontend/bindings/time/models.js
    suppression added: @ts-ignore added (0 -> 1) in apps/desktop/backend/frontend/bindings/time/models.js
    suppression added: eslint-disable added (0 -> 1) in frontend/src/features/dep-graph/DepGraphView.tsx

── Unused Code ─────────────────────────────────────

● Unused files (7)
  frontend/src/features/dep-graph/DepGraphView.tsx
  frontend/src/features/git-view/GitView.tsx
  frontend/src/features/monitor-view/MonitorView.tsx
  frontend/src/features/perf-view/PerfView.tsx
  frontend/src/lib/git.ts
  frontend/src/lib/perf.ts
  tools/oxlint/anti-slop/rules/no-runtime-typeof.test.ts
  Files not reachable from any entry point — https://docs.fallow.tools/explanations/dead-code#unused-files
  To suppress: // fallow-ignore-file unused-file

● Unused exports (58)
  apps/desktop/frontend/src/bridge.ts (18)
    :71 wailsSender
    :129 wailsAuthAdapter
    :171 wailsEnvAdapter
    :370 wailsCollectionsAdapter
    :494 wailsHistoryAdapter
    ... and 13 more (--format json for full list)
  frontend/src/components/ui/toast.tsx (11)
    :222 Toast
    :223 ToastAction
    :224 ToastClose
    :225 ToastContent
    :226 ToastDescription
    ... and 6 more (--format json for full list)
  frontend/src/components/ui/dropdown-menu.tsx (9)
    :252 DropdownMenuPortal
    :255 DropdownMenuGroup
    :258 DropdownMenuCheckboxItem
    :259 DropdownMenuRadioGroup
    :260 DropdownMenuRadioItem
    ... and 4 more (--format json for full list)
  frontend/src/lib/schemaGraph.ts (6)
    :12 SCHEMA_EDGES
    :35 ZOOM_MIN
    :36 ZOOM_MAX
    :37 NODE_WIDTH
    :38 NODE_HEIGHT
    ... and 1 more (--format json for full list)
  frontend/src/components/ui/menubar.tsx (4)
    :184 MenubarGroup
    :188 MenubarSub
    :189 MenubarSubTrigger
    :190 MenubarSubContent
  frontend/src/lib/typeGuards.ts (3)
    :13 isNumber
    :21 isObject
    :25 isDefinedString
  frontend/src/lib/bottomPanel.ts (2)
    :5 isBottomPanelId
    :10 nextPanel
  frontend/src/components/status.tsx
    :22 statusTier
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :259 ProxyTlsPanel
  frontend/src/lib/authSchemes.ts
    :237 DEFAULT_OAUTH2_GRANT
  ... and 2 more in 2 files (--format json for full list)
  Exported symbols with no known consumers — https://docs.fallow.tools/explanations/dead-code#unused-exports
  To auto-fix: fallow fix --dry-run
  To suppress: // fallow-ignore-next-line unused-export
  (4 more in files already reported as unused)

── Dependencies ─────────────────────────────────────

● Unused dependencies (2)
  frontend
  next-themes
  Listed in dependencies but never imported — https://docs.fallow.tools/explanations/dead-code#unused-dependencies

── Structure ─────────────────────────────────────

● Circular dependencies (1)
  frontend/src/stores/useCollectionRunStore.ts
    → useWorkspaceStore.ts → useCollectionRunStore.ts

  Import cycles that can cause initialization failures and prevent tree-shaking — https://docs.fallow.tools/explanations/dead-code#circular-dependencies

✗ 7 files · 58 exports · 2 unused dependencies · 1 circular dependency (0.26s)
note: skipped 34 files matching default duplicates ignores (use --explain-skipped for the list)
note: module wiring excluded from clone detection (--no-ignore-imports to include it)

● Duplicates (12 clone groups)

     12 lines  2 instances  spread 2  dup:c77b3abb6f87acd9-5
    frontend/src/features/dep-graph/DepGraphView.tsx:75-86
    frontend/src/features/spec-editor/SpecEditorView.tsx:18-29

     23 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-11
    frontend/src/stores/useWorkspaceBootstrap.ts:143-165
    frontend/src/stores/useWorkspaceBootstrap.ts:173-195

     19 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-8
    frontend/src/features/command-palette/CommandPalette.tsx:57-74
    frontend/src/features/command-palette/CommandPalette.tsx:79-97

      5 lines  2 instances  spread 1  dup:e82d7ce7
    frontend/src/components/WorkspaceSidebar.tsx:175-179
    frontend/src/components/shell/ContextSidebar.tsx:34-38

     13 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-7
    frontend/src/features/monitor-view/MonitorView.tsx:138-143
    frontend/src/features/monitor-view/MonitorView.tsx:147-159

     10 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-10
    frontend/src/features/diff-view/DiffView.tsx:209-218
    frontend/src/features/diff-view/DiffView.tsx:223-232

      5 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-4
    frontend/src/components/WorkspaceSidebar.tsx:36-40
    frontend/src/components/shell/TopBar.tsx:39-43

      8 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-9
    frontend/src/lib/authSchemes.ts:59-66
    frontend/src/lib/authSchemes.ts:86-93

      8 lines  2 instances  spread 1  dup:c77b3abb6f87acd9-1
    frontend/src/stores/useHistoryStore.ts:97-104
    frontend/src/stores/useHistoryStore.ts:115-122

     15 lines  2 instances  dup:c77b3abb6f87acd9-3
    frontend/src/features/environments-view/EnvironmentEditor.tsx:23-29
    frontend/src/features/environments-view/SecretsEditor.tsx:28-42

  ... and 2 more clone groups
  Duplicate code blocks - https://docs.fallow.tools/explanations/duplication#clone-groups

● Clone families (1 with multiple groups)

  2 groups, 27 lines across frontend/src/features/environments-view/EnvironmentEditor.tsx, frontend/src/features/environments-view/SecretsEditor.tsx
    → Extract shared function (15 lines) from EnvironmentEditor.tsx, SecretsEditor.tsx
    → Extract shared function (12 lines) from EnvironmentEditor.tsx, SecretsEditor.tsx

  Groups of related clones across the same files — https://docs.fallow.tools/explanations/duplication#clone-families

✗ 256 lines (1.0%) duplicated across 14 files (0.07s)

■ Metrics: 18,262 LOC · dead files 7.2% · dead exports 18.0% · avg cyclomatic 2.3 · p90 cyclomatic 4 · maintainability 86.3 (good) · 1 circular dep · 2 unused deps

  Function size: 85% low · 7% medium · 3% high · 4% very high  (1-15 / 16-30 / 31-60 / >60 LOC)

  Render fan-in: <Button> 44 parents (124 incl. repeats) · <Input> 20 parents (47 incl. repeats) · <Alert> 17 parents (19 incl. repeats) · <AlertDescription> 17 parents (19 incl. repeats) · <Spinner> 15 parents (24 incl. repeats)

● Large functions (10 shown, 69 total)
  frontend/src/features/request-editor/RequestEditor.tsx
    :113 RequestEditor  532 lines
  frontend/src/features/mock-view/MocksView.tsx
    :30 MocksView  514 lines
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :42 ResponseViewer  510 lines
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :40 RunnersPanel  326 lines
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :32 EnvironmentsView  302 lines
  frontend/src/features/history-view/HistoryView.tsx
    :37 HistoryView  295 lines
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :27 OpenapiExplorer  278 lines
  frontend/src/app/App.tsx
    :57 App  268 lines
  frontend/src/stores/useWorkspaceStore.ts
    :220 useWorkspaceStore  266 lines
  frontend/src/features/import-dialog/ImportDialog.tsx
    :112 ImportDialog  244 lines
  Functions exceeding 60 lines of code (very high risk): https://docs.fallow.tools/explanations/health#unit-size
  use --top 69 to see all

● High complexity functions (153)
  CRAP scores are estimated from export references; run `fallow health --coverage <coverage-final.json>` for exact scores.
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :42 ResponseViewer CRITICAL
          78 ! cyclomatic  271 ! cognitive  510 lines
         react: 16 hooks (4 state, 10 memo, 2 custom), JSX depth 7
         6162.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :113 RequestEditor CRITICAL
          61 ! cyclomatic   95 ! cognitive  532 lines
         react: 18 hooks (5 state, 1 effect, 12 custom), max effect deps 2, JSX depth 6
         3782.0 ! CRAP
  frontend/src/lib/specTree.ts
    :49 tryParseYaml CRITICAL
          47 ! cyclomatic   89 ! cognitive  103 lines
         524.1 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :112 ImportDialog CRITICAL
          32 ! cyclomatic   45 ! cognitive  244 lines
         react: 1 props, 19 hooks (1 state, 18 custom), JSX depth 6
         blast radius: <ImportDialog> rendered in 3 places
         1056.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :278 normalizeOpenedRequest CRITICAL
          31 ! cyclomatic   15   cognitive   38 lines
         992.0 ! CRAP
  frontend/src/features/mock-view/MocksView.tsx
    :30 MocksView CRITICAL
          29 ! cyclomatic   64 ! cognitive  514 lines
         react: 29 hooks (2 state, 1 effect, 26 custom), max effect deps 0, JSX depth 8
         870.0 ! CRAP
  frontend/src/hooks/useKeyboardMap.ts
    :29 handler CRITICAL
          27 ! cyclomatic   36 ! cognitive   78 lines
         756.0 ! CRAP
  frontend/src/lib/body.ts
    :118 serializeBody CRITICAL
          26 ! cyclomatic   32 ! cognitive   59 lines
         702.0 ! CRAP
  frontend/src/app/App.tsx
    :57 App CRITICAL
          25 ! cyclomatic  211 ! cognitive  268 lines
         react: 19 hooks (1 state, 4 effect, 14 custom), max effect deps 2, JSX depth 10
         650.0 ! CRAP
  frontend/src/components/shell/BottomPanel.tsx
    :20 PanelContent CRITICAL
          25 ! cyclomatic   38 ! cognitive  222 lines
         react: 1 props, 9 hooks (9 custom), JSX depth 6
         650.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :105 DiffView CRITICAL
          24 ! cyclomatic   25 ! cognitive  181 lines
         react: 1 props, 11 hooks (9 state, 1 effect, 1 custom), max effect deps 1, JSX depth 6
         blast radius: <ChangesList> rendered in 2 places
         600.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :61 saveWarnings CRITICAL
          23 ! cyclomatic   28 ! cognitive   51 lines
         552.0 ! CRAP
  frontend/src/lib/datasets.ts
    :56 parseCsv CRITICAL
          23 ! cyclomatic   45 ! cognitive   65 lines
  frontend/src/lib/specTree.ts
    :268 patchEndpointInContent HIGH
          23 ! cyclomatic   39 ! cognitive   90 lines
  frontend/src/features/grpc-view/GrpcTab.tsx
    :39 GrpcTab CRITICAL
          22 ! cyclomatic   25 ! cognitive  155 lines
         react: 1 props, 9 hooks (2 effect, 7 custom), max effect deps 1, JSX depth 5
         506.0 ! CRAP
  frontend/src/lib/specTree.ts
    :233 yamlEscape CRITICAL
          22 ! cyclomatic    3   cognitive   28 lines
         126.5 ! CRAP
  frontend/src/components/RunView.tsx
    :109 RunView CRITICAL
          21 ! cyclomatic   39 ! cognitive  182 lines
         react: 15 hooks (4 state, 11 custom), JSX depth 7
         462.0 ! CRAP
  frontend/src/stores/useCommandPaletteStore.ts
    :99 groupByHint CRITICAL
          21 ! cyclomatic   28 ! cognitive   18 lines
         116.3 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :44 RealtimeTab CRITICAL
          20   cyclomatic   30 ! cognitive  209 lines
         react: 1 props, 11 hooks (3 state, 1 effect, 7 custom), max effect deps 1, JSX depth 4
         blast radius: <RealtimeTab> rendered in 3 places
         420.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :27 OpenapiExplorer CRITICAL
          18   cyclomatic   18 ! cognitive  278 lines
         react: 6 hooks (2 state, 4 custom), JSX depth 6
         342.0 ! CRAP
  frontend/src/components/CollectionTree.tsx
    :17 treeKeyDown CRITICAL
          17   cyclomatic   18 ! cognitive   38 lines
         blast radius: <CollectionBranch> rendered in 2 places
         306.0 ! CRAP
  frontend/src/lib/themes.ts
    :80 parseSimpleYaml HIGH
          17   cyclomatic   23 ! cognitive   51 lines
          79.4 ! CRAP
  frontend/src/features/graphql-browser/GraphqlBrowser.tsx
    :92 GraphqlBrowser CRITICAL
          16   cyclomatic   18 ! cognitive  180 lines
         react: 7 hooks (7 state), JSX depth 6
         blast radius: <FieldRow> rendered in 2 places
         272.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :71 wailsSender CRITICAL
          15   cyclomatic   16 ! cognitive   52 lines
         240.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :67 displayEntries CRITICAL
          15   cyclomatic   18 ! cognitive   11 lines
         240.0 ! CRAP
  frontend/src/features/settings-view/SettingsView.tsx
    :47 SettingsView CRITICAL
          15   cyclomatic   19 ! cognitive  191 lines
         react: 7 hooks (1 state, 6 custom), JSX depth 8
         240.0 ! CRAP
  frontend/src/lib/request.ts
    :128 fetchSender CRITICAL
          15   cyclomatic   14   cognitive   51 lines
         240.0 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :164 <arrow> CRITICAL
          14   cyclomatic   12   cognitive   37 lines
         react: JSX depth 2
         blast radius: <RealtimeTab> rendered in 3 places
         210.0 ! CRAP
  frontend/src/features/docs-view/DocsView.tsx
    :17 DocsView CRITICAL
          13   cyclomatic   27 ! cognitive  168 lines
         react: 14 hooks (1 state, 1 effect, 12 custom), max effect deps 1, JSX depth 8
         182.0 ! CRAP
  frontend/src/features/spec-editor/SpecEditorView.tsx
    :43 SpecEditorView CRITICAL
          13   cyclomatic   26 ! cognitive  159 lines
         react: 15 hooks (1 state, 1 memo, 13 custom), JSX depth 6
         182.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :779 decode CRITICAL
          13   cyclomatic    7   cognitive   20 lines
         182.0 ! CRAP
  frontend/src/components/shell/ContextSidebar.tsx
    :450 ContextSidebar CRITICAL
          13   cyclomatic    3   cognitive   62 lines
         react: 1 props, 1 hooks (1 custom), JSX depth 2
         blast radius: <SectionLabel> rendered in 13 places
         182.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :672 RetrySection CRITICAL
          13   cyclomatic    8   cognitive   97 lines
         react: 2 props, 1 hooks (1 state), JSX depth 4
         182.0 ! CRAP
  tools/oxlint/anti-slop/rules/no-runtime-typeof.ts
    :51 UnaryExpression CRITICAL
          13   cyclomatic    6   cognitive   22 lines
         182.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :40 RunnersPanel CRITICAL
          12   cyclomatic   18 ! cognitive  326 lines
         react: 3 hooks (1 state, 1 effect, 1 custom), max effect deps 1, JSX depth 7
         156.0 ! CRAP
  frontend/src/components/RunView.tsx
    :18 StepRow CRITICAL
          12   cyclomatic   11   cognitive   87 lines
         react: 2 props, 1 hooks (1 state), JSX depth 5
         156.0 ! CRAP
  frontend/src/lib/authSchemes.ts
    :306 authWarnings CRITICAL
          12   cyclomatic   11   cognitive   18 lines
         156.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :364 <arrow> CRITICAL
          12   cyclomatic   15   cognitive   33 lines
         react: JSX depth 2
         156.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :216 InheritedAuth CRITICAL
          11   cyclomatic   11   cognitive   40 lines
         react: 1 props, JSX depth 2
         blast radius: <AuthFieldRow> rendered in 2 places
         132.0 ! CRAP
  frontend/src/features/graphql-browser/GraphqlBrowser.tsx
    :18 FieldRow CRITICAL
          11   cyclomatic    8   cognitive   47 lines
         react: 2 props, 1 hooks (1 state), JSX depth 3
         blast radius: <FieldRow> rendered in 2 places
         132.0 ! CRAP
  frontend/src/features/git-view/GitView.tsx
    :6 GitView CRITICAL
          10   cyclomatic   17 ! cognitive   65 lines
         react: 8 hooks (6 state, 1 effect, 1 callback), max effect deps 1, JSX depth 4
         110.0 ! CRAP
  frontend/src/features/jwt-inspector/JwtInspector.tsx
    :54 <arrow> CRITICAL
          10   cyclomatic    4   cognitive   16 lines
         react: JSX depth 3
         blast radius: <ClaimsTable> rendered in 2 places
         110.0 ! CRAP
  frontend/src/components/KeyValueEditor.tsx
    :33 <arrow> CRITICAL
          10   cyclomatic    9   cognitive   96 lines
         react: JSX depth 4
         blast radius: <KeyValueEditor> rendered in 5 places
         110.0 ! CRAP
  frontend/src/features/request-editor/TemplatePickerSheet.tsx
    :26 TemplatePickerSheet CRITICAL
          10   cyclomatic   14   cognitive  130 lines
         react: 3 props, 3 hooks (2 state, 1 custom), JSX depth 5
         110.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :290 <arrow> CRITICAL
          10   cyclomatic    9   cognitive   31 lines
         blast radius: <ImportDialog> rendered in 3 places
         110.0 ! CRAP
  frontend/src/features/auth-panel/AuthPanel.tsx
    :37 AuthPanel CRITICAL
          10   cyclomatic   13   cognitive  139 lines
         react: 5 hooks (3 state, 1 effect, 1 custom), max effect deps 1, JSX depth 4
         blast radius: <AuthPanel> rendered in 2 places
         110.0 ! CRAP
  frontend/src/lib/body.ts
    :83 mimeForFile CRITICAL
          10   cyclomatic    1   cognitive   24 lines
         110.0 ! CRAP
  frontend/src/features/spec-editor/EndpointEditor.tsx
    :17 EndpointEditor CRITICAL
          10   cyclomatic   13   cognitive  122 lines
         react: 3 props, 7 hooks (5 state, 2 memo), JSX depth 4
         110.0 ! CRAP
  frontend/src/features/workspace-home/HomeView.tsx
    :32 HomeView HIGH
           9   cyclomatic   18 ! cognitive  162 lines
         react: 12 hooks (1 effect, 11 custom), max effect deps 2, JSX depth 6
         blast radius: <Stat> rendered in 4 places
          90.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :20 ProxyPanel HIGH
           9   cyclomatic   10   cognitive  134 lines
         react: 4 hooks (4 custom), JSX depth 6
         blast radius: <ProxyPanel> rendered in 2 places
          90.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :274 closeTab HIGH
           9   cyclomatic    6   cognitive   26 lines
          90.0 ! CRAP
    :366 saveRequest HIGH
           9   cyclomatic    9   cognitive   26 lines
          90.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :178 <arrow> HIGH
           9   cyclomatic    8   cognitive   82 lines
         react: JSX depth 4
          90.0 ! CRAP
  frontend/src/lib/body.ts
    :61 encodeFormData HIGH
           9   cyclomatic   10   cognitive   20 lines
          90.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :65 onSave HIGH
           9   cyclomatic   12   cognitive   31 lines
          90.0 ! CRAP
  frontend/src/components/shell/ToolRail.tsx
    :63 ToolRail HIGH
           9   cyclomatic   10   cognitive  103 lines
         react: 2 props, 2 hooks (2 custom), JSX depth 5
          90.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :421 <arrow> HIGH
           9   cyclomatic    8   cognitive   31 lines
         react: JSX depth 2
          90.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :94 start HIGH
           9   cyclomatic    9   cognitive   38 lines
          90.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :232 <arrow> HIGH
           9   cyclomatic    9   cognitive   67 lines
         react: JSX depth 4
          90.0 ! CRAP
  frontend/src/components/shell/TopBar.tsx
    :36 TopBar HIGH
           8   cyclomatic   18 ! cognitive  165 lines
         react: 12 hooks (12 custom), JSX depth 8
          72.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :18 EnvironmentEditor HIGH
           8   cyclomatic   17 ! cognitive  190 lines
         react: 2 props, 11 hooks (4 state, 1 effect, 3 memo, 3 custom), max effect deps 3, JSX depth 4
          72.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :37 HistoryView HIGH
           8   cyclomatic   27 ! cognitive  295 lines
         react: 19 hooks (7 state, 1 effect, 11 custom), max effect deps 1, JSX depth 6
          72.0 ! CRAP
  frontend/src/lib/datasets.ts
    :24 parseCsvLine
           8   cyclomatic   19 ! cognitive   31 lines
  apps/desktop/frontend/src/bridge.ts
    :349 normalizeRunReport HIGH
           8   cyclomatic    7   cognitive   20 lines
          72.0 ! CRAP
    :717 endpoints HIGH
           8   cyclomatic   10   cognitive   16 lines
          72.0 ! CRAP
  frontend/src/features/runners-panel/DatasetPicker.tsx
    :10 DatasetPicker HIGH
           8   cyclomatic   15   cognitive  134 lines
         react: 9 hooks (2 state, 3 callback, 4 custom), JSX depth 5
          72.0 ! CRAP
  frontend/src/lib/crash.ts
    :98 formatReport HIGH
           8   cyclomatic   11   cognitive   41 lines
          72.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :134 openFolder HIGH
           8   cyclomatic   10   cognitive   33 lines
          72.0 ! CRAP
    :168 openDirect HIGH
           8   cyclomatic   10   cognitive   29 lines
          72.0 ! CRAP
  frontend/src/features/perf-view/PerfView.tsx
    :6 PerfView HIGH
           8   cyclomatic   10   cognitive   98 lines
         react: 5 hooks (5 state), JSX depth 5
          72.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :115 <arrow> HIGH
           8   cyclomatic    7   cognitive   55 lines
         react: JSX depth 2
          72.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :203 <arrow> HIGH
           8   cyclomatic    7   cognitive   21 lines
         react: JSX depth 1
          72.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :128 run HIGH
           8   cyclomatic    7   cognitive   19 lines
         blast radius: <ChangesList> rendered in 2 places
          72.0 ! CRAP
  frontend/src/stores/useHistoryStore.ts
    :55 load HIGH
           8   cyclomatic    8   cognitive   16 lines
          72.0 ! CRAP
  frontend/src/components/RequestTabs.tsx
    :23 TabItem HIGH
           8   cyclomatic   15   cognitive  133 lines
         react: 3 props, 8 hooks (3 state, 5 custom), JSX depth 5
          72.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentsView.tsx
    :32 EnvironmentsView HIGH
           7   cyclomatic   18 ! cognitive  302 lines
         react: 11 hooks (4 state, 1 effect, 6 custom), max effect deps 1, JSX depth 5
          56.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :331 normalizeRunStep HIGH
           7   cyclomatic    6   cognitive   17 lines
          56.0 ! CRAP
    :394 save HIGH
           7   cyclomatic    6   cognitive   25 lines
          56.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :46 AuthEditor HIGH
           7   cyclomatic    8   cognitive   61 lines
         react: 3 props, JSX depth 3
         blast radius: <AuthFieldRow> rendered in 2 places
          56.0 ! CRAP
  frontend/src/components/status.tsx
    :22 statusTier HIGH
           7   cyclomatic    6   cognitive    8 lines
         blast radius: <StatusPill> rendered in 3 places
          56.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :59 visibleGroups HIGH
           7   cyclomatic   10   cognitive   14 lines
         blast radius: <ImportDialog> rendered in 3 places
          56.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :236 finishSwitch HIGH
           7   cyclomatic    4   cognitive   23 lines
          56.0 ! CRAP
  frontend/src/features/openapi-explorer/OpenapiExplorer.tsx
    :63 filteredEndpoints HIGH
           7   cyclomatic    4   cognitive    7 lines
          56.0 ! CRAP
  frontend/src/features/settings-view/CicdPanel.tsx
    :10 CicdPanel HIGH
           7   cyclomatic   14   cognitive  170 lines
         react: 8 hooks (3 state, 5 custom), JSX depth 5
          56.0 ! CRAP
  frontend/src/components/shell/ToolRail.tsx
    :67 railButton HIGH
           7   cyclomatic    6   cognitive   35 lines
         react: JSX depth 2
          56.0 ! CRAP
  frontend/src/components/ui/toast.tsx
    :135 ToastIcon HIGH
           7   cyclomatic    6   cognitive   46 lines
         react: 1 props, JSX depth 1
          56.0 ! CRAP
  frontend/src/features/runners-panel/RunnersPanel.tsx
    :343 <arrow> HIGH
           7   cyclomatic    6   cognitive   17 lines
         react: JSX depth 3
          56.0 ! CRAP
  frontend/src/components/CollectionTree.tsx
    :251 CollectionTree HIGH
           7   cyclomatic   13   cognitive  107 lines
         react: 7 hooks (1 state, 1 memo, 5 custom), JSX depth 4
         blast radius: <CollectionBranch> rendered in 2 places
          56.0 ! CRAP
  frontend/src/features/monitor-view/MonitorView.tsx
    :15 MonitorView HIGH
           7   cyclomatic   10   cognitive  184 lines
         react: 4 hooks (3 state, 1 effect), max effect deps 1, JSX depth 8
          56.0 ! CRAP
  frontend/src/features/jwt-inspector/JwtInspector.tsx
    :75 JwtInspector
           6   cyclomatic    9   cognitive  119 lines
         react: 4 hooks (4 state), JSX depth 6
         blast radius: <ClaimsTable> rendered in 2 places
          42.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :452 exportReport
           6   cyclomatic    5   cognitive   18 lines
          42.0 ! CRAP
    :499 show
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
    :895 invoke
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :115 <arrow>
           6   cyclomatic    3   cognitive    6 lines
         blast radius: <ProxyPanel> rendered in 2 places
          42.0 ! CRAP
    :128 <arrow>
           6   cyclomatic    3   cognitive    6 lines
         blast radius: <ProxyPanel> rendered in 2 places
          42.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :83 onSave
           6   cyclomatic    7   cognitive   22 lines
          42.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :145 bodyTypeFor
           6   cyclomatic    5   cognitive   14 lines
          42.0 ! CRAP
    :338 duplicateTab
           6   cyclomatic    5   cognitive   11 lines
          42.0 ! CRAP
    :393 overwriteRequest
           6   cyclomatic    5   cognitive   26 lines
          42.0 ! CRAP
  frontend/src/components/CreateWorkspaceModal.tsx
    :26 handlePickFolder
           6   cyclomatic    7   cognitive   14 lines
          42.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :770 VariablesView
           6   cyclomatic    5   cognitive   42 lines
         react: 3 props, JSX depth 2
          42.0 ! CRAP
  frontend/src/components/shell/StatusBar.tsx
    :5 StatusBar
           6   cyclomatic    7   cognitive   53 lines
         react: 2 hooks (2 custom), JSX depth 4
          42.0 ! CRAP
  frontend/src/components/WorkspaceSidebar.tsx
    :31 WorkspaceSidebar
           6   cyclomatic   15   cognitive  142 lines
         react: 10 hooks (1 state, 9 custom), JSX depth 6
          42.0 ! CRAP
  frontend/src/features/environments-view/SecretsEditor.tsx
    :21 SecretsEditor
           6   cyclomatic   15   cognitive  199 lines
         react: 3 props, 10 hooks (4 state, 1 effect, 2 memo, 3 custom), max effect deps 3, JSX depth 4
          42.0 ! CRAP
  frontend/src/features/diff-view/DiffView.tsx
    :71 ChangeRow
           6   cyclomatic    6   cognitive   33 lines
         react: 2 props, 1 hooks (1 state), JSX depth 4
         blast radius: <ChangesList> rendered in 2 places
          42.0 ! CRAP
  frontend/src/stores/useHistoryStore.ts
    :107 replayWithVars
           6   cyclomatic    7   cognitive   17 lines
          42.0 ! CRAP
  frontend/src/features/spec-editor/SpecEditorView.tsx
    :69 handleGenerate
           6   cyclomatic    6   cognitive   22 lines
          42.0 ! CRAP
  frontend/src/features/grpc-view/GrpcTab.tsx
    :12 statusBadge
           6   cyclomatic    1   cognitive   23 lines
         react: JSX depth 2
          42.0 ! CRAP
  frontend/src/components/RequestTabs.tsx
    :173 <arrow>
           6   cyclomatic    6   cognitive   12 lines
          42.0 ! CRAP
  apps/desktop/frontend/src/bridge.ts
    :371 load
           5   cyclomatic    4   cognitive   16 lines
          30.0 ! CRAP
    :458 steps
           5   cyclomatic    4   cognitive    8 lines
          30.0 ! CRAP
    :600 toDiffResultView
           5   cyclomatic    2   cognitive   17 lines
          30.0 ! CRAP
    :629 toGqlType
           5   cyclomatic    4   cognitive   21 lines
          30.0 ! CRAP
    :619 introspect
           5   cyclomatic    4   cognitive   40 lines
         react: 3 props
          30.0 ! CRAP
    :807 responses
           5   cyclomatic    4   cognitive    9 lines
          30.0 ! CRAP
    :872 generate
           5   cyclomatic    4   cognitive   15 lines
          30.0 ! CRAP
  frontend/src/features/settings-view/ProxyTlsPanel.tsx
    :155 TlsSecurityPanel
           5   cyclomatic    8   cognitive  103 lines
         react: 4 hooks (4 custom), JSX depth 5
         blast radius: <ProxyPanel> rendered in 2 places
          30.0 ! CRAP
  frontend/src/features/environments-view/EnvironmentEditor.tsx
    :36 dirty
           5   cyclomatic    5   cognitive   10 lines
          30.0 ! CRAP
    :52 duplicateKey
           5   cyclomatic    6   cognitive   11 lines
          30.0 ! CRAP
    :67 secretLikeWarnings
           5   cyclomatic    6   cognitive   12 lines
          30.0 ! CRAP
  frontend/src/lib/paletteProviders.ts
    :67 getItems
           5   cyclomatic    6   cognitive   32 lines
          30.0 ! CRAP
  frontend/src/features/auth-editor/AuthEditor.tsx
    :111 AuthFieldRow
           5   cyclomatic    4   cognitive   41 lines
         react: 4 props, JSX depth 3
         blast radius: <AuthFieldRow> rendered in 2 places
          30.0 ! CRAP
  frontend/src/hooks/useKeyboardMap.ts
    :20 isTypingTarget
           5   cyclomatic    3   cognitive    6 lines
          30.0 ! CRAP
  frontend/src/components/status.tsx
    :35 StatusPill
           5   cyclomatic    4   cognitive   28 lines
         react: 2 props, JSX depth 2
         blast radius: <StatusPill> rendered in 3 places
          30.0 ! CRAP
  frontend/src/features/git-view/GitView.tsx
    :13 load
           5   cyclomatic    4   cognitive   13 lines
          30.0 ! CRAP
  frontend/src/stores/useWorkspaceStore.ts
    :184 baseUrlFor
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
    :420 reloadRequest
           5   cyclomatic    4   cognitive   23 lines
          30.0 ! CRAP
  frontend/src/features/import-dialog/ImportDialog.tsx
    :45 groups
           5   cyclomatic    3   cognitive   10 lines
         blast radius: <ImportDialog> rendered in 3 places
          30.0 ! CRAP
  frontend/src/components/shell/ContextSidebar.tsx
    :40 openTestTab
           5   cyclomatic    5   cognitive    7 lines
         blast radius: <SectionLabel> rendered in 13 places
          30.0 ! CRAP
    :160 MocksContext
           5   cyclomatic   11   cognitive   81 lines
         react: 7 hooks (7 custom), JSX depth 4
         blast radius: <SectionLabel> rendered in 13 places
          30.0 ! CRAP
  frontend/src/stores/useWorkspaceBootstrap.ts
    :44 getStoredRecentWorkspaces
           5   cyclomatic    4   cognitive   11 lines
          30.0 ! CRAP
    :110 init
           5   cyclomatic    5   cognitive   23 lines
          30.0 ! CRAP
  frontend/src/features/history-view/HistoryView.tsx
    :92 onReplayWithVars
           5   cyclomatic    5   cognitive    9 lines
          30.0 ! CRAP
  frontend/src/components/CreateWorkspaceModal.tsx
    :41 handleCreate
           5   cyclomatic    5   cognitive   16 lines
          30.0 ! CRAP
  frontend/src/lib/authSchemes.ts
    :274 authForScheme
           5   cyclomatic    3   cognitive   11 lines
          30.0 ! CRAP
  frontend/src/features/request-editor/RequestEditor.tsx
    :133 onKeyDown
           5   cyclomatic    4   cognitive    6 lines
          30.0 ! CRAP
    :233 <arrow>
           5   cyclomatic    5   cognitive    6 lines
          30.0 ! CRAP
    :293 <arrow>
           5   cyclomatic    2   cognitive    3 lines
          30.0 ! CRAP
    :416 <arrow>
           5   cyclomatic    4   cognitive   21 lines
         react: JSX depth 1
          30.0 ! CRAP
    :653 settingsSummary
           5   cyclomatic    4   cognitive    7 lines
          30.0 ! CRAP
    :661 retrySummary
           5   cyclomatic    4   cognitive    5 lines
          30.0 ! CRAP
  frontend/src/components/shell/BottomPanel.tsx
    :190 <arrow>
           5   cyclomatic    3   cognitive   11 lines
         react: JSX depth 2
          30.0 ! CRAP
  frontend/src/components/WorkspaceSidebar.tsx
    :181 openTestTab
           5   cyclomatic    5   cognitive    7 lines
          30.0 ! CRAP
  frontend/src/features/response-viewer/ResponseViewer.tsx
    :83 cookiesText
           5   cyclomatic    4   cognitive    9 lines
          30.0 ! CRAP
    :106 imageDataUrl
           5   cyclomatic    3   cognitive    4 lines
          30.0 ! CRAP
  frontend/src/features/command-palette/CommandPalette.tsx
    :4 CommandPalette
           5   cyclomatic   14   cognitive  102 lines
         react: 9 hooks (1 effect, 8 custom), max effect deps 1, JSX depth 5
          30.0 ! CRAP
  frontend/src/stores/useRequestStore.ts
    :129 tabIsDirty
           5   cyclomatic    2   cognitive    4 lines
          30.0 ! CRAP
    :196 send
           5   cyclomatic    5   cognitive   43 lines
          30.0 ! CRAP
  frontend/src/features/realtime-view/RealtimeTab.tsx
    :25 statusBadge
           5   cyclomatic    1   cognitive   18 lines
         react: JSX depth 2
         blast radius: <RealtimeTab> rendered in 3 places
          30.0 ! CRAP
    :64 getPlaceholder
           5   cyclomatic    1   cognitive   14 lines
         blast radius: <RealtimeTab> rendered in 3 places
          30.0 ! CRAP
  frontend/src/features/request-editor/RequestSettingsDialog.tsx
    :43 RequestSettingsDialog
           5   cyclomatic    6   cognitive   82 lines
         react: 3 props, 2 hooks (2 state), JSX depth 5
          30.0 ! CRAP
  frontend/src/features/mock-view/MocksView.tsx
    :227 <arrow>
           5   cyclomatic    4   cognitive   95 lines
         react: JSX depth 4
          30.0 ! CRAP
    :351 <arrow>
           5   cyclomatic    4   cognitive   44 lines
         react: JSX depth 5
          30.0 ! CRAP
  Functions exceeding cyclomatic, cognitive, or CRAP thresholds; ! marks the dimension that breached (https://docs.fallow.tools/explanations/health#complexity-metrics)
  To suppress: // fallow-ignore-next-line complexity


CSS health
  Styling health: 100 A (CSS quality, scored separately from the code health score)
  2 stylesheets · 14 rules · 3.7% !important · 0 empty · max nesting 0
  value sprawl: 34 distinct colors · 1 font size · 0 z-index values
  value sprawl (cont.): 1 radius value · 1 line-height (candidates; tokenize repeated values via custom properties)
  custom properties: 27 defined, 2 unreferenced in CSS, 4 undefined (candidates; may be set from JS)
  frontend/src/index.css:69  specificity (1,0,0) · complexity 1 · 0 !important · nesting 0
  frontend/src/index.css:89  specificity (0,0,1) · complexity 3 · 3 !important · nesting 0
✗ 153 above threshold · 1733 analyzed (0.01s)
  Fix confidently
    frontend/src/index.css:69  css-selector-complexity  specificity 1-0-0  (advisory: rules.css-selector-complexity=warn, introduced design-system drift since d438b742cb9a)
  Verify first
    frontend/src/index.css:89  css-selector-complexity  3 !important declarations across 3 declarations  (advisory: rules.css-selector-complexity=warn, introduced design-system drift since d438b742cb9a)
  (run `fallow audit --format json` for full styling detail)
```

</details>

---

### 13. React Architecture & Performance (react-doctor)

- **Command:** `CI=1 printf '\n' | nubx -y react-doctor@latest`
- **Status:** 🟢 PASS
- **Duration:** 23s

<details>
<summary>Click to expand full output</summary>

```text
nub: pnpm-workspace.yaml is not read under nub identity — migrate it (`nub pm use nub`), delete it, or return to pnpm (`nub pm use pnpm`).
nub 0.7.5
███████████░░░░  207/271 pkgs · ~118.0 MB
███████████████  201/201 pkgs
✓ resolved 201 · reused 201 in 4.3s
dependencies:
+ react-doctor@0.9.12

✔ Select projects › @reqly/frontend, @reqly/desktop
✔ Scanned 190 files in 13.1s

React Doctor — reqly-main-brancn
Score: 72 / 100 Needs work

74 issues
Maintainability: 58 warnings
Bugs: 11 warnings
Performance: 3 warnings
Accessibility: 2 warnings

⚠ Manual memoization in compiler-managed code ×33
  react-doctor/react-compiler-no-manual-memoization
  src/components/CollectionTree.tsx:259
  src/features/dep-graph/DepGraphView.tsx:18
  src/features/dep-graph/DepGraphView.tsx:19
  src/features/dep-graph/DepGraphView.tsx:20
  src/features/dep-graph/DepGraphView.tsx:26
  src/features/dep-graph/DepGraphView.tsx:33
  src/features/dep-graph/DepGraphView.tsx:40
  src/features/environments-view/EnvironmentEditor.tsx:36
  src/features/environments-view/EnvironmentEditor.tsx:52
  src/features/environments-view/EnvironmentEditor.tsx:67
  src/features/environments-view/SecretsEditor.tsx:47
  src/features/environments-view/SecretsEditor.tsx:57
  src/features/import-dialog/ImportDialog.tsx:45
  src/features/import-dialog/ImportDialog.tsx:59
  src/features/import-dialog/ImportReportView.tsx:46
  src/features/response-viewer/ResponseViewer.tsx:57
  src/features/response-viewer/ResponseViewer.tsx:62
  src/features/response-viewer/ResponseViewer.tsx:77
  src/features/response-viewer/ResponseViewer.tsx:94
  src/features/response-viewer/ResponseViewer.tsx:95
  src/features/response-viewer/ResponseViewer.tsx:96
  src/features/response-viewer/ResponseViewer.tsx:97
  src/features/response-viewer/ResponseViewer.tsx:101
  src/features/response-viewer/ResponseViewer.tsx:105
  src/features/response-viewer/ResponseViewer.tsx:123
  src/features/response-viewer/ResponseViewer.tsx:554
  src/features/runners-panel/DatasetPicker.tsx:20
  src/features/runners-panel/DatasetPicker.tsx:33
  src/features/runners-panel/DatasetPicker.tsx:47
  src/features/spec-editor/EndpointEditor.tsx:24
  src/features/spec-editor/EndpointEditor.tsx:34
  src/features/spec-editor/SpecEditorView.tsx:58
  src/hooks/useFuseSearch.ts:12
⚠ Array index used as a key ×10
  react-doctor/no-array-index-as-key
  src/components/KeyValueEditor.tsx:34
  src/components/RunView.tsx:279
  src/components/shell/BottomPanel.tsx:162
  src/components/shell/BottomPanel.tsx:191
  src/components/shell/BottomPanel.tsx:219
  src/features/environments-view/EnvironmentEditor.tsx:145
  src/features/mock-view/MocksView.tsx:516
  src/features/monitor-view/MonitorView.tsx:180
  src/features/response-viewer/ResponseViewer.tsx:383
  src/features/spec-editor/SpecEditorView.tsx:186
⚠ Non-component export in component file ×5
  react-doctor/only-export-components
  src/components/status.tsx:22
  src/components/ui/button.tsx:58
  src/components/ui/tabs.tsx:80
  src/components/ui/toast.tsx:231
  src/components/ui/toast.tsx:232
⚠ deslop/unused-file ×8
  deslop/unused-file
  src/components/ui/tabs.tsx
  src/components/ui/tooltip.tsx
  src/features/dep-graph/DepGraphView.tsx
  src/features/git-view/GitView.tsx
  src/features/monitor-view/MonitorView.tsx
  src/features/perf-view/PerfView.tsx
  src/lib/git.ts
  src/lib/perf.ts
⚠ Heavy library loaded eagerly ×2
  react-doctor/prefer-dynamic-import
  src/editors/CodeMirrorEditor.tsx:6
  src/editors/CodeMirrorEditor.tsx:8
⚠ Custom modal instead of dialog
  react-doctor/prefer-html-dialog
  src/features/command-palette/CommandPalette.tsx:23
⚠ Interaction on static element
  react-doctor/no-static-element-interactions
  src/features/dep-graph/DepGraphView.tsx:66
⚠ Large component is hard to read and change ×5
  react-doctor/no-giant-component
  src/features/environments-view/EnvironmentsView.tsx:32
  src/features/mock-view/MocksView.tsx:30
  src/features/request-editor/RequestEditor.tsx:113
  src/features/response-viewer/ResponseViewer.tsx:42
  src/features/runners-panel/RunnersPanel.tsx:40
⚠ React Compiler can't optimize this
  react-hooks-js/set-state-in-effect
  src/features/git-view/GitView.tsx:28
⚠ deslop/unused-export ×7
  deslop/unused-export
  src/features/settings-view/ProxyTlsPanel.tsx:259
  src/lib/bottomPanel.ts:5
  src/lib/bottomPanel.ts:10
  src/lib/themes.ts:18
  src/lib/typeGuards.ts:13
  src/lib/typeGuards.ts:21
  src/lib/typeGuards.ts:25
⚠ fetch Response consumed without status check
  react-doctor/no-fetch-response-used-without-status-check
  src/lib/request.ts:147
```

</details>

---

