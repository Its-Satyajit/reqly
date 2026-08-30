# UI Architecture: Phase Panels, Navigation & Layout (§56–§63)

├── Editor
├── Network
├── Proxy
├── TLS
├── Environments
├── Auth
├── Storage
├── Keyboard Shortcuts
├── Notifications
├── Extensions
└── Advanced
```

---

# 56. Phase 2: P1

## 56.1 OpenAPI Spec Editor

```text
┌──────────────────────┬───────────────────────────────────────────────────────┐
│ SPEC TREE            │ openapi.yaml                                          │
├──────────────────────┼───────────────────────────────────────────────────────┤
│ Info                 │ openapi: 3.1.0                                       │
│ Servers              │                                                       │
│ Paths                │ paths:                                                │
│ ├── /users           │   /users:                                             │
│ ├── /products        │     get:                                              │
│ └── /orders          │       summary: List users                            │
│ Components           │                                                       │
│ ├── Schemas          │ components:                                           │
│ └── Security         │   schemas:                                            │
└──────────────────────┴───────────────────────────────────────────────────────┘
```

## 56.2 Schema Visualization

```text
                    ┌─────────────┐
                    │    User     │
                    └──────┬──────┘
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
         Address         Order        Profile
                            │
                            ▼
                         Product
```

## 56.3 Request Templates

```text
Templates
├── REST
│   ├── CRUD
│   ├── Pagination
│   ├── Upload
│   └── Authentication
├── GraphQL
├── gRPC
└── Realtime
```

## 56.4 Proxy / TLS Controls

```text
Proxy
├── HTTP Proxy
├── HTTPS Proxy
├── SOCKS
└── Authentication

TLS
├── Verification
├── Custom CA
├── Client Certificates
└── TLS Version
```

## 56.5 Data-driven Testing

Dataset UI:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ DATASET                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ users.csv                                                                     │
│                                                                              │
│ id    name      role                                                          │
│ 1     John      user                                                          │
│ 2     Jane      admin                                                         │
│ 3     Mike      user                                                          │
│                                                                              │
│ Iterations: 3    Concurrency: 2                                             │
│ [ Run Dataset ]                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 56.6 CI/CD Integration

Provides:

```text
CLI command
Pipeline configuration
Environment selection
Secrets
Reports
Exit codes
Test artifacts
```

## 56.7 Full Mock Server GUI

Adds:

```text
Routes
Scenarios
State
Dynamic Data
Latency
Fault Injection
Request Matching
Logs
Server Configuration
```

## 56.8 GraphQL / gRPC Documentation

Generated documentation from:

```text
GraphQL Schema
Protobuf
gRPC Reflection
```

---

# 57. Phase 3: P2

## 57.1 API Monitoring Dashboard

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ API MONITORING                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Requests 48,291 │ Success 99.7% │ Avg 184ms │ Availability 99.99%           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Availability                  Latency                                        │
│ 100% ──────────────           400ms ┤       ╭──╮                             │
│  99% ───────╮──────           200ms ┤──╭────╯  ╰──                           │
│              ╰────              0ms └────────────────                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 57.2 Performance Testing

Panels:

```text
Load Configuration
Concurrency
Request Rate
Duration
Latency
Throughput
Errors
Scenarios
Results
Comparison
```

## 57.3 MQTT / Socket.IO

Shared realtime interface:

```text
Connection
Topics / Events
Messages
Headers
Payload
Timeline
```

## 57.4 Dependency Graph

```text
Frontend
   │
   ▼
Users API ─────► Auth API
   │
   ├───────────► Orders API
   │                 │
   │                 ▼
   └───────────► Payments API
```

## 57.5 Request Replay

Replay sources:

```text
History
Network
Monitoring
Timeline
Diff
Saved Request
```

Replay controls:

```text
Original Environment
Current Environment
Modified Headers
Modified Body
Modified Query
```

## 57.6 In-app Developer Tools / Debugger

Panels:

```text
Console
Network
Variables
Scripts
Breakpoints
Runtime Errors
Request Inspector
Response Inspector
```

## 57.7 Git GUI

```text
Repository
├── Status
├── Branches
├── Commits
├── Diff
├── Staging
├── Pull
├── Push
├── Merge
└── Conflict Resolution
```

## 57.8 Network Interception / Timeline Debugging

```text
DNS       ███
TCP          ████
TLS              ██████
Upload                  ██
Server                    █████████████
Download                                ███
──────────────────────────────────────────────►
0ms                                        142ms
```

---

# 58. Phase 4: P3

## 58.1 Plugin Marketplace

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PLUGIN MARKETPLACE                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Search plugins...                                                            │
│                                                                              │
│ ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐              │
│ │ GraphQL Tools    │ │ AWS Toolkit      │ │ Database Explorer│              │
│ │ ★ 4.9            │ │ ★ 4.8            │ │ ★ 4.7            │              │
│ │ [ Install ]      │ │ [ Install ]      │ │ [ Install ]      │              │
│ └──────────────────┘ └──────────────────┘ └──────────────────┘              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 58.2 Theme Marketplace

Controls:

```text
Color Tokens
Typography
Density
Editor Style
Icons
Panel Style
```

## 58.3 Git Provider Integrations

Providers:

```text
GitHub
GitLab
Bitbucket
```

Functions:

```text
Authenticate
Select Repository
Select Branch
Pull
Push
Sync
Commit
Review
```

## 58.4 Team / Shared Workspaces

```text
Workspace
├── Members
├── Roles
├── Collections
├── Environments
├── Shared Secrets
├── Activity
└── Settings
```

## 58.5 Enterprise

```text
Enterprise
├── SSO
├── SCIM
├── Audit Logs
├── Organization Policies
├── Permissions
├── Secret Policies
├── IP Restrictions
└── Compliance
```

---

# 59. Phase 5

## 59.1 MCP Server

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ MCP SERVER                                                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│ Status: ● Running                                                            │
│ Endpoint: localhost:xxxx                                                     │
│                                                                              │
│ TOOLS                                                                        │
│ send_request                                                                 │
│ inspect_response                                                             │
│ list_collections                                                             │
│ get_environment                                                              │
│ run_collection                                                               │
│ inspect_openapi                                                              │
│                                                                              │
│ RESOURCES                                                                    │
│ collections                                                                  │
│ environments                                                                 │
│ documentation                                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 59.2 Command Palette

```text
┌─────────────────────────────────────────────────────────────┐
│ > Search commands...                                        │
├─────────────────────────────────────────────────────────────┤
│ REQUEST                                                     │
│   New Request                                    Ctrl+N     │
│   Send Request                                  Ctrl+Enter  │
│   Duplicate Request                              Ctrl+D     │
│                                                             │
│ NAVIGATION                                                  │
│   Open History                                  Ctrl+H      │
│   Open Environments                             Ctrl+E      │
│   Open Collections                              Ctrl+K      │
│                                                             │
│ TOOLS                                                       │
│   Open JWT Inspector                                        │
│   Open GraphQL                                              │
│   Open gRPC                                                 │
└─────────────────────────────────────────────────────────────┘
```

## 59.3 Optional AI Assistant

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ REQUEST WORKSPACE                                      AI ASSISTANT           │
├─────────────────────────────────────────────────────┬────────────────────────┤
│ GET /users                                          │ CONTEXT                │
│                                                     │ Request                │
│ Params                                              │ Response               │
│ Headers                                             │ Environment            │
│ Body                                                │ OpenAPI                │
│ Response                                            │ Collection             │
│                                                     │ Tests                  │
│                                                     │                        │
│                                                     │ SUGGESTIONS            │
│                                                     │ Explain response       │
│                                                     │ Generate tests         │
│                                                     │ Find issue             │
│                                                     │ Generate schema        │
│                                                     │                        │
│                                                     │ Ask about request…    │
└─────────────────────────────────────────────────────┴────────────────────────┘
```

---

# 60. Complete Navigation Map

```text
REQLY
│
├── WORKSPACE
│   ├── Home / Overview
│   ├── Collections
│   │   ├── Collection Details
│   │   ├── Request Tree
│   │   └── Collection Runner
│   │
│   ├── Requests
│   │   ├── Request Tabs
│   │   ├── URL Bar
│   │   ├── Params
│   │   ├── Headers
│   │   ├── Body
│   │   ├── Auth
│   │   ├── Pre-request
│   │   ├── Tests
│   │   ├── Docs
│   │   ├── Settings
│   │   ├── Metadata
│   │   └── Response
│   │       ├── Body
│   │       ├── Headers
│   │       ├── Cookies
│   │       ├── Test Results
│   │       └── Timeline
│   │
│   ├── Environments
│   │   ├── Variables
│   │   ├── Diff
│   │   ├── Validate
│   │   └── Cross-check
│   │
│   └── History
│       ├── Request History
│       └── Request Detail
│
├── API TOOLS
│   ├── Mocks
│   │   ├── Servers
│   │   ├── Routes
│   │   ├── Route Editor
│   │   └── Logs
│   │
│   ├── Diff
│   │   ├── Request Diff
│   │   ├── Response Diff
│   │   ├── JSON Diff
│   │   ├── Headers Diff
│   │   └── Environment Diff
│   │
│   ├── JWT Inspector
│   │   ├── Header
│   │   ├── Payload
│   │   ├── Claims
│   │   └── Token Status
│   │
│   ├── GraphQL
│   │   ├── Schema
│   │   ├── Query
│   │   ├── Mutation
│   │   ├── Variables
│   │   ├── Headers
│   │   └── Response
│   │
│   ├── gRPC
│   │   ├── Servers
│   │   ├── Services
│   │   ├── Methods
│   │   ├── Request
│   │   ├── Metadata
│   │   ├── Response
│   │   └── Streaming
│   │
│   ├── Runners
│   │   ├── Collection
│   │   ├── Pagination
│   │   ├── Bulk
│   │   ├── Dataset
│   │   └── Results
│   │
│   ├── Explorer
│   │   ├── API Tree
│   │   ├── Endpoint Details
│   │   ├── Parameters
│   │   ├── Schemas
│   │   ├── Responses
│   │   └── Security
│   │
│   └── Docs
│       ├── Overview
│       ├── Authentication
│       ├── Endpoints
│       ├── Schemas
│       ├── Examples
│       └── Errors
│
├── REALTIME
│   ├── WebSocket
│   │   ├── Connection
│   │   ├── Headers
│   │   ├── Messages
│   │   ├── Events
│   │   └── Timeline
│   │
│   └── SSE
│       ├── Connection
│       ├── Headers
│       ├── Event Stream
│       └── Timeline
│
├── DEVELOPMENT
│   ├── Test Files
│   ├── Console
│   ├── Network
│   ├── Variables
│   ├── Cookies
│   └── Test Results
│
├── SETTINGS
│   ├── General
│   ├── Appearance
│   ├── Editor
│   ├── Network
│   ├── Proxy
│   ├── TLS
│   ├── Auth
│   ├── Storage
│   ├── Keyboard Shortcuts
│   ├── Notifications
│   └── Advanced
│
├── PHASE 2
│   ├── OpenAPI Editor
│   ├── Schema Visualization
│   ├── Request Templates
│   ├── Proxy / TLS Controls
│   ├── Data-driven Testing
│   ├── CI/CD
│   ├── Full Mock GUI
│   └── GraphQL / gRPC Docs
│
├── PHASE 3
│   ├── Monitoring
│   ├── Performance Testing
│   ├── MQTT
│   ├── Socket.IO
│   ├── Dependency Graph
│   ├── Request Replay
│   ├── Developer Tools
│   ├── Git GUI
│   └── Network Timeline
│
├── PHASE 4
│   ├── Plugin Marketplace
│   ├── Theme Marketplace
│   ├── Git Providers
│   ├── Team Workspaces
│   └── Enterprise
│
└── PHASE 5
    ├── MCP
    ├── Command Palette
    └── AI Assistant
```

---

# 61. Shared Interaction Patterns

All pages should reuse the same interaction vocabulary.

## Search

Every large collection should expose:

```text
⌕ Search
```

## Primary action

Usually appears top-right:

```text
[ New ]
[ Run ]
[ Send ]
[ Save ]
[ Connect ]
```

## Secondary actions

Use:

```text
⋮
```

for:

```text
Rename
Duplicate
Move
Export
Delete
Archive
```

## Status

Use consistent states:

```text
● Connected
● Running
● Valid
● Success
● Warning
● Error
```

## Tabs

Use tabs for different views of the same resource.

Examples:

```text
Body
Headers
Cookies
Tests
Timeline
```

## Panels

Use panels for related configuration.

Examples:

```text
Params
Headers
Auth
Body
Tests
```

---

# 62. Page vs Panel Rules

Reqly should distinguish between pages and panels.

### Full pages

Use for:

```text
Requests
Environments
History
Mocks
Diff
JWT
GraphQL
gRPC
Runners
Explorer
Docs
WebSocket
SSE
Settings
```

### Context panels

Use for:

```text
Collection Tree
Request Metadata
Environment Variables
Schema Tree
Service Tree
Saved Credentials
```

### Request panels

Use for:

```text
Params
Headers
Body
Auth
Pre-request
Tests
Docs
Settings
```

### Response panels

Use for:

```text
Body
Headers
Cookies
Test Results
Timeline
```

### Bottom utility panels

Use for:

```text
Console
Network
Tests
Variables
Cookies
```

### Dialogs

Use for:

```text
Import
Export
Create Workspace
Create Collection
Create Environment
Authentication
Connection Setup
Confirmation
```

This keeps the product from turning every small piece of functionality into a separate page, which is one of the faster ways to make an application feel like enterprise software from 2011.

---

# 63. Final Layout Model

The final Reqly application should read like this:

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ TOP BAR                                                                                      │
│ Reqly │ Workspace │ Search │ Import │ Export │ Environment │ Status │ Settings               │
├──────┬───────────────────────┬───────────────────────────────────────────────────────────────┤
│      │                       │                                                               │
│ TOOL │ CONTEXT               │ MAIN WORKSPACE                                                │
│ RAIL │ SIDEBAR               │                                                               │
│      │                       │ Tabs / Editor / Inspector / Results                           │
│ ⚡   │ Collections            │                                                               │
│ ◈    │ Requests               │                                                               │
│ ≋    │ History                │                                                               │
│ ◫    │ Mocks                  │                                                               │
│ ⇄    │ Diff                   │                                                               │
│ ♢    │ JWT                    │                                                               │
│ ◎    │ GraphQL                │                                                               │
│ ⌁    │ gRPC                   │                                                               │
│ ▶    │ Runners                │                                                               │
│ ◇    │ Explorer               │                                                               │
│ ▤    │ Docs                   │                                                               │
│ ◌    │ WebSocket              │                                                               │
│ ◌    │ SSE                    │                                                               │
│      │                       │                                                               │
├──────┴───────────────────────┴───────────────────────────────────────────────────────────────┤
│ Console │ Network │ Tests │ Variables │ Cookies                                  Ready ●    │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

The key architectural distinction is:

```text
TOP BAR
    = Workspace / global actions

TOOL RAIL
    = switch major tool

CONTEXT SIDEBAR
    = switch resource inside that tool

MAIN WORKSPACE
    = perform the actual API task

PAGE PANELS
    = configure or inspect that task

BOTTOM PANEL
    = observe execution and debugging state
```

This consolidated structure preserves the missing surfaces from the earlier specs while keeping the two-sidebar model intact.


# Source traceability

## Primary source files used
- `ROADMAP(2).md` — newest development-roadmap snapshot used as the canonical base.
- `ROADMAP(3).md` — older development roadmap used to preserve detailed milestone/ticket history and implementation notes that would otherwise be lost.
- `gui-roadmap.md` — desktop GUI execution tracker used for UI delivery state.
- `Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md` — full lower-precedence UI reference preserved in the appendix.

## Conflict policy
- Newer development-roadmap state beats older development-roadmap state.
- Development-roadmap scope beats GUI-roadmap scope.
- GUI roadmap delivery state can clarify desktop status but cannot delete product-roadmap work.
- UI architecture is reference material and cannot promote itself into product scope.
- A feature may legitimately have mixed status, such as core `[x]`, desktop `[~]`, and UI polish `[ ]`.
- Historical entries remain for traceability even when current status has moved on.

## Completeness rule
No source detail should be deleted merely because it is duplicated. Duplicate statements are consolidated into the current roadmap, while unique ticket-level or UI-reference detail is retained in the historical or lower-precedence sections.

---

## Code Review Gate (`/code-review` — two-axis)

- [x] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [x] Spec: this milestone (`Milestones/` + Phase) vs implementation (`ROADMAP.md` DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`
