# UI Architecture: Shell & Request Workspace (§1–§25)


## GUI-5 P1 Data Layer (spec §56.3–56.8) — PRESERVED

> These data layer items (lib + stores + tests) are preserved from the previous implementation.

- [x] **G-5.1** Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [x] **G-5.2** Proxy/TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [x] **G-5.3** Data-driven testing — zustand store + pure lib (CSV/JSON parse, row vars, validate) + 23 tests — 2026-08-26
- [x] **G-5.4** CI/CD integration — zustand store + pure lib (CLI gen, GitHub Action YAML, report parse) + 13 tests — 2026-08-26
- [x] **G-5.5** Mock server GUI data — extended zustand store + pure lib (scenarios, fault injection, matchers, logs) + 20 tests — 2026-08-26
- [x] **G-5.6** GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26


---

# Lower-precedence UI architecture reference

This appendix preserves the complete `Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification` as a reference inventory. It is intentionally subordinate to the development roadmaps and the GUI execution roadmap.

Rules for using this appendix:
- Its page/panel/navigation details are implementation guidance.
- Its presence does not create a new product commitment.
- Its layout or naming proposals do not override roadmap priority.
- When it contains a UI idea that is absent from the development roadmap, treat that idea as a candidate/reference until a roadmap milestone adopts it.
- When roadmap status and UI reference status differ, roadmap status wins.

# Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification

## 1. Purpose

This document is the consolidated UI specification for Reqly.

It includes the full set of pages, panels, sub-panels, editors, inspectors, dialogs, navigation surfaces, workspace controls, debugging surfaces, and roadmap features discussed previously.

The architecture is based on four persistent UI layers:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ TOP BAR                                                                      │
├──────┬───────────────────────┬───────────────────────────────────────────────┤
│      │                       │                                               │
│ TOOL │ CONTEXT SIDEBAR       │ MAIN WORKSPACE                               │
│ RAIL │                       │                                               │
│      │                       │                                               │
│      │                       │                                               │
├──────┴───────────────────────┴───────────────────────────────────────────────┤
│ BOTTOM UTILITY PANEL                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

The model is:

```text
Workspace
   ↓
Tool Rail
   ↓
Context Sidebar
   ↓
Page / Main Workspace
   ↓
Page-specific Panels
   ↓
Bottom Utility Panels
```

---

# 2. Global Application Shell

## 2.1 Top Bar

The top bar is always available.

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ ◉ REQLY     Workspace ▾     ⌕ Search      Import  Export      Development ▾     ● Sync   ⚙  │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Components

```text
Top Bar
├── Reqly Logo
├── Workspace Switcher
├── Global Search
├── Import
├── Export
├── Active Environment
├── Sync / Connection Status
├── Notifications
├── Settings
└── User / Account Menu
```

## 2.2 Workspace Switcher

```text
┌─────────────────────────────────────────┐
│ WORKSPACE                               │
├─────────────────────────────────────────┤
│ 🔍 Search workspaces...                 │
│                                         │
│ PERSONAL                                │
│ ● My Workspace                          │
│                                         │
│ PROJECTS                                │
│ ◈ Reqly API                             │
│ ◈ Payments                              │
│ ◈ Internal Services                     │
│                                         │
│ SHARED                                  │
│ 👥 Backend Team                         │
│ 👥 Engineering                          │
│                                         │
│ ──────────────────────────────────────  │
│ + Create Workspace                      │
│ ⚙ Manage Workspaces                     │
└─────────────────────────────────────────┘
```

## 2.3 Global Search

Global search should search across:

```text
Requests
Collections
Folders
Environments
Variables
History
Mocks
OpenAPI
GraphQL operations
gRPC methods
Tests
Documentation
Workspaces
Commands
```

Example:

```text
┌─────────────────────────────────────────────────────────────┐
│ ⌕ Search Reqly...                                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Requests                                                    │
│   GET /users                                                │
│   POST /users                                               │
│                                                             │
│ Collections                                                │
│   Users                                                     │
│                                                             │
│ History                                                     │
│   GET /users?page=2                                        │
│                                                             │
│ Commands                                                    │
│   Open Environments                                        │
│   New Request                                               │
└─────────────────────────────────────────────────────────────┘
```

---

# 3. Tool Rail

The tool rail is approximately 48–56px wide.

```text
┌──────┐
│  ◉   │ Workspace
├──────┤
│  ⚡  │ Requests
│  ◈   │ Environments
│  ≋   │ History
├──────┤
│  ◫   │ Mocks
│  ⇄   │ Diff
│  ♢   │ JWT
│  ◎   │ GraphQL
│  ⌁   │ gRPC
│  ▶   │ Runners
│  ◇   │ Explorer
│  ▤   │ Docs
├──────┤
│  ◌   │ WebSocket
│  ◌   │ SSE
├──────┤
│  ⚙   │ Settings
└──────┘
```

## Tool groups

```text
WORKSPACE
├── Workspace
├── Requests
├── Environments
└── History

API TOOLS
├── Mocks
├── Diff
├── JWT Inspector
├── GraphQL
├── gRPC
├── Runners
├── Explorer
└── Docs

REALTIME
├── WebSocket
└── SSE

SYSTEM
└── Settings
```

---

# 4. Context Sidebar

The second sidebar changes according to the active tool.

It answers:

> What am I working with inside this tool?

Typical width:

```text
220–280px
```

It supports:

* collapse
* resize
* search
* tree navigation
* contextual actions
* recent items
* pinned items

Shortcut:

```text
Ctrl/Cmd + B
```

---

# 5. Workspace Home / Overview

The workspace itself should have a landing page.

```text
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ WORKSPACE: Reqly API                                                                  │
├──────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│ Welcome back                                                                          │
│                                                                                      │
│ ┌────────────────┐ ┌────────────────┐ ┌────────────────┐ ┌────────────────────────┐ │
│ │ Requests       │ │ Environments   │ │ Collections    │ │ Recent Activity        │ │
│ │ 128            │ │ 4              │ │ 12             │ │ 18 requests today     │ │
│ └────────────────┘ └────────────────┘ └────────────────┘ └────────────────────────┘ │
│                                                                                      │
│ QUICK ACTIONS                                                                         │
│ [ New Request ] [ Import API ] [ Open Collection ] [ New Environment ]               │
│                                                                                      │
│ RECENT REQUESTS                                                                       │
│ GET /users                                      200     142 ms                        │
│ POST /users                                     201     183 ms                        │
│ GET /products                                   200      91 ms                        │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

# 6. Requests Page

Requests are Reqly's primary workspace.

The page contains:

```text
Requests
├── Request Tabs
├── Request URL Bar
├── Request Builder
├── Request Metadata
├── Response Viewer
└── Request Actions
```

---

# 7. Request Tabs

Requests should behave similarly to editor tabs.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ● GET Users    ● Create User    ● Get Product    +                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

Tab features:

* open
* close
* pin
* duplicate
* reopen closed tab
* close others
* close all
* unsaved indicator
* drag reorder
* context menu
* restore previous session

---

# 8. Request URL Bar

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ GET ▾ │ https://api.example.com/v1/users/{{user_id}}           Send ▶ ▾       │
└──────────────────────────────────────────────────────────────────────────────┘
```

Features:

* method selector
* URL editor
* autocomplete
* path variable detection
* environment variables
* request history
* Send
* Send with options
* cancel request
* save request
* duplicate request

Methods:

```text
GET
POST
PUT
PATCH
DELETE
HEAD
OPTIONS
CONNECT
TRACE
```

---

# 9. Request Builder

## 9.1 Request Builder Navigation

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

These are individual request panels.

---

# 10. Params Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ QUERY PARAMETERS                                                             │
├────┬──────────────┬──────────────────────────────┬──────────────────────────┤
│ ☑  │ Key          │ Value                        │ Description              │
├────┼──────────────┼──────────────────────────────┼──────────────────────────┤
│ ☑  │ page         │ 1                            │ Current page             │
│ ☑  │ limit        │ 20                           │ Items per page            │
│ ☑  │ search       │ {{search}}                   │ Search term              │
└────┴──────────────┴──────────────────────────────┴──────────────────────────┘

+ Add parameter
```

Supports:

* query parameters
* path parameters
* parameter enable/disable
* descriptions
* variable interpolation
* generated values
* encoding preview

---

# 11. Headers Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ HEADERS                                                                      │
├────┬────────────────┬───────────────────────────────────────────────────────┤
│ ☑  │ Content-Type   │ application/json                                      │
│ ☑  │ Accept         │ application/json                                      │
│ ☑  │ Authorization  │ Bearer {{access_token}}                              │
│ ☐  │ X-Debug        │ true                                                  │
└────┴────────────────┴───────────────────────────────────────────────────────┘

+ Add header
Import headers
```

Header sources:

```text
Request
Collection
Folder
Environment
Auth
Generated
```

---

# 12. Body Panel

Supported formats:

```text
None
JSON
Raw
Text
XML
HTML
Form URL Encoded
Multipart Form
Binary
GraphQL
```

JSON editor:

```text
1 │ {
2 │   "name": "John Doe",
3 │   "email": "john@example.com",
4 │   "role": "user"
5 │ }
```

Editor capabilities:

* syntax highlighting
* autocomplete
* formatting
* schema validation
* folding
* line numbers
* error markers
* search
* replace
* generated examples

---

# 13. Auth Panel

Authentication options:

```text
Inherit
No Auth
Basic Auth
Bearer Token
API Key
OAuth 2.0
Digest Auth
AWS Signature
Custom
```

Example:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ AUTHENTICATION                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Type: Bearer Token ▾                                                         │
│                                                                              │
│ Token                                                                        │
│ {{access_token}}                                                             │
│                                                                              │
│ [ Generate ] [ Select Saved Token ]                                          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 14. Pre-request Script Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PRE-REQUEST SCRIPT                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ 1 │ const token = await reqly.auth.getToken();                              │
│ 2 │ reqly.variables.set("timestamp", Date.now());                           │
│ 3 │                                                                          │
└──────────────────────────────────────────────────────────────────────────────┘
```

Capabilities:

* JavaScript/TypeScript-style scripting
* autocomplete
* variables
* request mutation
* auth helpers
* console output
* execution errors
* script files

---

# 15. Tests Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ TESTS                                                                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ 1 │ expect(response.status).toBe(200);                                     │
│ 2 │ expect(response.body.data).toHaveLength(20);                            │
│ 3 │                                                                          │
└──────────────────────────────────────────────────────────────────────────────┘

[ Run Tests ]
```

Results:

```text
● status is 200
● response has users
● content type is JSON
○ pagination metadata
```

---

# 16. Request Docs Panel

The request-level docs panel shows:

* endpoint description
* generated API documentation
* OpenAPI metadata
* parameters
* schemas
* examples
* authentication requirements

---

# 17. Request Settings Panel

Request-specific settings:

```text
Timeout
Redirect handling
SSL verification
Proxy
Retry
Cookies
Compression
HTTP version
Streaming
```

---

# 18. Request Metadata Panel

A request metadata inspector should expose:

```text
Request
├── Collection
├── Folder
├── Created
├── Updated
├── Owner
├── Environment
├── Tags
├── Description
└── Source
```

Actions:

* rename
* move
* duplicate
* tag
* pin
* archive
* delete

---

# 19. Response Viewer

The response viewer is part of the Requests page.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ RESPONSE                                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ ● 200 OK    142 ms    2.4 KB    HTTP/2                        Copy  Save  ⋮ │
├──────────────────────────────────────────────────────────────────────────────┤
│ Body │ Headers │ Cookies │ Test Results │ Timeline                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 20. Response Body Panel

Modes:

```text
Pretty
Raw
Tree
Preview
```

Format modes:

```text
JSON
XML
HTML
Text
Binary
Image
```

JSON tree:

```text
▼ data
  ├── 0
  │   ├── id
  │   ├── name
  │   ├── email
  │   └── role
  └── 1
      └── ...
```

Actions:

* copy
* download
* expand all
* collapse all
* search
* generate schema
* generate tests
* send to diff
* replay

---

# 21. Response Headers Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ RESPONSE HEADERS                                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ content-type      application/json                                           │
│ cache-control     no-cache                                                   │
│ content-length   2401                                                        │
│ server            nginx                                                      │
│ x-request-id      abc-123                                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 22. Response Cookies Panel

Shows:

```text
Name
Value
Domain
Path
Expires
Secure
HttpOnly
SameSite
```

Cookie actions:

* inspect
* copy
* delete
* replay

---

# 23. Test Results Panel

Displays:

```text
● 8 passed
● 1 skipped
● 0 failed
```

Includes:

* assertion details
* execution duration
* stack trace
* console output
* failed value diff

---

# 24. Timeline Panel

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

Breakdown:

```text
DNS
Connection
TLS
Upload
Server processing
Download
```

---

# 25. Response Actions
