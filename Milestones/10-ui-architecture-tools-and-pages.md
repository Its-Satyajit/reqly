# UI Architecture: Pages, Tools & Protocols (§26–§55)


```text
Copy
Save
Download
Replay
Diff
Generate Schema
Generate Test
Generate Documentation
Open in Explorer
```

---

# 26. Collections Explorer

The Requests context sidebar contains the collection tree.

```text
┌──────────────────────────────┐
│ COLLECTIONS                  │
├──────────────────────────────┤
│ 🔍 Search                   │
│                              │
│ ▼ Reqly API                 │
│   ▼ Authentication          │
│     POST Login              │
│     POST Refresh            │
│                              │
│   ▼ Users                   │
│     GET List Users          │
│     GET Get User            │
│     POST Create User        │
│                              │
│   ▼ Products                │
│     GET Products            │
│                              │
│ + Collection                │
│ + Folder                    │
│ + Request                   │
└──────────────────────────────┘
```

Collection actions:

```text
Rename
Move
Duplicate
Delete
Run
Import
Export
Generate Docs
Generate Tests
Generate Mock
```

---

# 27. Collections Page

A collection can also open as a full page.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ COLLECTION: Users                                      Run Collection ▶      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Description                                                                    │
│ User management API                                                           │
│                                                                              │
│ Requests                                                                       │
│ GET     /users                                                                │
│ GET     /users/:id                                                            │
│ POST    /users                                                                │
│ PATCH   /users/:id                                                            │
│ DELETE  /users/:id                                                            │
│                                                                              │
│ Tests: 18     Requests: 5     Docs: Generated                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 28. Environments Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ENVIRONMENTS                                           + New Environment      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Local   Development ●   Staging   Production                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│ Variables                                                                     │
│                                                                              │
│ Name             Value                         Secret       Description       │
│ base_url         https://api.dev.com                       API URL            │
│ api_version      v1                                                    │
│ access_token     •••••••••••••                ✓            Auth              │
│ user_id          42                                             Test user    │
│                                                                              │
│ + Add variable                                                               │
│                                                                              │
│ [ Save ] [ Validate ] [ Diff ] [ Cross-check ]                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 29. Environment Diff Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ENVIRONMENT DIFF                                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ Development ▾                          Production ▾                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Variable         Development              Production            Result      │
│ base_url         api.dev.example.com      api.example.com       ≠           │
│ api_version      v1                       v1                     =           │
│ timeout          5000                     10000                  ≠           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 30. Environment Validate Panel

Validation categories:

```text
Required variables
Malformed URLs
Unresolved variables
Duplicate variables
Unused variables
Invalid values
Circular references
Missing secrets
```

Results:

```text
● 24 checks passed
● 2 warnings
● 1 error
```

---

# 31. Environment Cross-check Panel

Cross-checks variables against:

```text
Requests
Collections
Scripts
Tests
OpenAPI
Mocks
Documentation
```

Example:

```text
Variable           Used By                     Result
base_url           42 requests                ✓
access_token       18 requests                ✓
legacy_host        0 requests                 ⚠ unused
missing_token      POST /orders               ✕ missing
```

---

# 32. History Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ HISTORY                                                Search   Clear         │
├──────────────────────────────────────────────────────────────────────────────┤
│ All ▾   Method ▾   Status ▾   Environment ▾   Date ▾                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ TODAY                                                                        │
│                                                                              │
│ 10:42 GET     /v1/users              200   142ms   Development               │
│ 10:39 POST    /v1/users              201   183ms   Development               │
│ 10:32 GET     /v1/products           200    91ms   Development               │
│ 10:27 DELETE  /v1/users/42           204   112ms   Development               │
└──────────────────────────────────────────────────────────────────────────────┘
```

History detail panel:

```text
Request
Response
Headers
Body
Timing
Environment
Variables
```

Actions:

```text
Reopen
Duplicate
Save to Collection
Replay
Diff
Export
Delete
```

---

# 33. Mocks Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ MOCK SERVERS                                          + New Mock Server       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Demo API                                                                     │
│ ● Running    http://localhost:4010                                          │
├──────────────────────────────────────────────────────────────────────────────┤
│ Routes                                                                        │
│ GET    /users            200      users.json                                │
│ POST   /users            201      create-user.json                          │
│ GET    /users/:id        200      user.json                                 │
│ DELETE /users/:id        204      empty                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Mock route editor:

```text
Method
Path
Matching rules
Status
Headers
Body
Latency
Scenario
```

---

# 34. Diff Page

The Diff tool compares:

```text
Requests
Responses
Headers
JSON
Text
Environments
Saved versions
```

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ DIFF                                                                         │
├──────────────────────────────┬───────────────────────────────────────────────┤
│ A                            │ B                                             │
├──────────────────────────────┼───────────────────────────────────────────────┤
│ GET /users?page=1            │ GET /users?page=2                             │
│                              │                                               │
│ "total": 42                  │ "total": 44                                  │
│ "page": 1                    │ "page": 2                                    │
└──────────────────────────────┴───────────────────────────────────────────────┘
```

Modes:

```text
Side by Side
Unified
Structural JSON
Headers
```

---

# 35. JWT Inspector Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ JWT INSPECTOR                                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Paste JWT                                                                     │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ eyJhbGciOiJIUzI1NiIs...                                                  │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ [ Decode ] [ Clear ]                                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ HEADER                                                                       │
│ { "alg": "HS256", "typ": "JWT" }                                             │
│                                                                              │
│ PAYLOAD                                                                      │
│ { "sub": "123", "name": "John", "iat": 123, "exp": 456 }                     │
│                                                                              │
│ SIGNATURE                                                                    │
│ Present                                                                      │
│                                                                              │
│ TOKEN STATUS                                                                 │
│ ● Valid                                                                      │
│ Expires in 42 minutes                                                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

Inspect:

```text
Header
Payload
Signature
Claims
Expiration
Issued At
Not Before
Issuer
Audience
Subject
Algorithm
```

---

# 36. GraphQL Browser Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ GRAPHQL                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Endpoint: https://api.example.com/graphql                         Run ▶     │
├───────────────────────┬──────────────────────────────────────────────────────┤
│ SCHEMA                │ QUERY                                                │
│                       │                                                      │
│ Query                 │ query Users {                                        │
│ ├── users             │   users {                                            │
│ ├── user              │     id                                               │
│ └── search            │     name                                             │
│                       │     email                                            │
│ Mutation              │   }                                                  │
│ ├── createUser        │ }                                                    │
│ └── deleteUser        │                                                      │
├───────────────────────┴──────────────────────────────────────────────────────┤
│ RESPONSE                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Panels:

```text
Schema
Query
Mutation
Variables
Headers
Fragments
Response
Documentation
```

---

# 37. GraphQL Schema Browser

The schema sidebar should allow:

```text
Query
Mutation
Subscription
Types
Enums
Interfaces
Input Types
Scalars
Directives
```

---

# 38. Runners Page

```text
Runners
├── Collection Runner
├── Pagination Runner
├── Bulk Runner
├── Dataset Runner
└── Run Results
```

---

# 39. Pagination Runner

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PAGINATION RUNNER                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ Request: GET /users                                                          │
│ Strategy: Offset ▾                                                           │
│ Parameter: page                                                              │
│ Start: 1                                                                     │
│ Max pages: 100                                                               │
│                                                                              │
│ Stop condition: Empty response                                               │
│                                                                              │
│ [ Run ]                                                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Page │ Status │ Duration │ Items                                             │
│ 1    │ 200    │ 124ms   │ 20                                                │
│ 2    │ 200    │ 118ms   │ 20                                                │
│ 3    │ 200    │ 121ms   │ 20                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 40. Bulk Runner

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ BULK RUNNER                                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│ Source: Users Collection                                                      │
│ Dataset: users.csv                                                            │
│ Environment: Development                                                      │
│ Concurrency: 4                                                               │
│ Delay: 100 ms                                                                 │
│ Retries: 2                                                                    │
│                                                                              │
│ [ Run ]                                                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Progress: ███████████████████░░░░ 78%                                        │
│                                                                              │
│ Passed  82     Failed  3     Skipped 1                                      │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 41. Explorer Page

OpenAPI Explorer:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ OPENAPI EXPLORER                                        Import Spec           │
├───────────────────┬──────────────────────────────────────────────────────────┤
│ API               │ GET /users                                               │
│                   │                                                          │
│ ▼ Users           │ List users                                               │
│   GET /users      │                                                          │
│   POST /users     │ Parameters                                               │
│   GET /users/{id} │ page     query      integer                             │
│                   │ limit    query      integer                             │
│ ▼ Products        │                                                          │
│ ▼ Orders          │ Responses                                                │
│                   │ 200 UserList                                             │
│                   │ 400 Error                                                │
│                   │                                                          │
│                   │ [ Open in Request Builder ]                             │
└───────────────────┴──────────────────────────────────────────────────────────┘
```

Panels:

```text
API Tree
Endpoint Details
Parameters
Request Schema
Response Schema
Responses
Security
Examples
```

---

# 42. REST Documentation Page

```text
┌──────────────────────┬───────────────────────────────────────────────────────┐
│ CONTENTS             │ GET /users                                            │
├──────────────────────┼───────────────────────────────────────────────────────┤
│ Overview             │ Returns a list of users.                             │
│ Authentication      │                                                       │
│ Users                │ Parameters                                            │
│ Products             │ page     integer                                     │
│ Orders               │ limit    integer                                     │
│ Errors               │                                                       │
│                      │ Example                                               │
│                      │ curl ...                                              │
│                      │                                                       │
│                      │ Response                                              │
└──────────────────────┴───────────────────────────────────────────────────────┘
```

---

# 43. gRPC Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ gRPC                                                                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ grpc.example.com:443                                            Connect ●    │
├───────────────────────┬──────────────────────────────────────────────────────┤
│ SERVICES              │ METHOD                                               │
│ UserService           │ GetUser                                              │
│ ├── GetUser           │                                                      │
│ ├── ListUsers         │ Request                                              │
│ └── CreateUser        │ { "id": 42 }                                         │
│                       │                                                      │
│ OrderService          │ Metadata                                             │
│ ├── GetOrder          │ authorization: Bearer ...                           │
│ └── CreateOrder       │                                                      │
│                       │ [ Invoke ]                                           │
├───────────────────────┴──────────────────────────────────────────────────────┤
│ RESPONSE                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Panels:

```text
Servers
Services
Methods
Request
Metadata
Response
Status
Streaming
Timeline
```

---

# 44. Import Dialog

```text
┌────────────────────────────────────────────────────────────────────┐
│ IMPORT                                                             │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│ ○ OpenAPI / Swagger                                               │
│ ○ Postman Collection                                              │
│ ○ Insomnia                                                        │
│ ○ cURL                                                            │
│ ○ HAR                                                              │
│ ○ Reqly                                                            │
│                                                                    │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ Drop file here or Browse                                      │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                    │
│ Destination: My Workspace ▾                                       │
│                                                                    │
│ [ Cancel ]                                      [ Import ]         │
└────────────────────────────────────────────────────────────────────┘
```

Import preview should show:

```text
Collections found
Requests found
Environments found
Variables found
Conflicts
Warnings
```

---

# 45. Export Dialog

```text
Export
├── Collection
├── Workspace
├── OpenAPI
├── cURL
├── HAR
├── Environment
└── Documentation
```

Export options:

```text
Include secrets
Include tests
Include scripts
Include docs
Normalize variables
```

---

# 46. Auth / OAuth Tokens Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ AUTHENTICATION                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ SAVED CREDENTIALS                                                            │
│ ● Development OAuth                                                          │
│ ● GitHub Token                                                               │
│ ● API Key                                                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│ OAuth 2.0                                                                    │
│ Authorization URL                                                            │
│ Token URL                                                                    │
│ Client ID                                                                     │
│ Client Secret                                                                │
│ Scopes                                                                       │
│                                                                              │
│ [ Authorize ] [ Refresh ] [ Revoke ]                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ ● Valid                                                                      │
│ Expires in 52 minutes                                                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

Token management:

```text
Access Tokens
Refresh Tokens
API Keys
OAuth Clients
Saved Credentials
Token Expiry
Token Revocation
```

---

# 47. Test Files Page

```text
┌───────────────────┬──────────────────────────────────────────────────────────┐
│ TEST FILES        │ users.test.ts                                            │
├───────────────────┼──────────────────────────────────────────────────────────┤
│ users.test.ts     │ describe("Users API", () => {                            │
│ auth.test.ts      │   it("returns users", async () => {                      │
│ products.test.ts  │     const response = await reqly.send(...);             │
│                   │     expect(response.status).toBe(200);                  │
│                   │   });                                                    │
└───────────────────┴──────────────────────────────────────────────────────────┘
```

Features:

```text
File browser
Editor
Syntax highlighting
Run test
Run file
Debug
Test output
Test history
Failures
```

---

# 48. WebSocket Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ WEBSOCKET                                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ ws://localhost:8080/socket                                    Connected ●    │
├───────────────────────────┬──────────────────────────────────────────────────┤
│ CONNECTION                │ MESSAGES                                         │
│ Headers                   │                                                  │
│ Authorization: Bearer...  │ → {"type":"subscribe"}                          │
│                           │ ← {"type":"connected"}                          │
│ Protocols                 │ ← {"type":"message"}                            │
│ json                      │                                                  │
│                           │ ┌──────────────────────────────────────────────┐ │
│                           │ │ {"type":"ping"}                              │ │
│                           │ └──────────────────────────────────────────────┘ │
│                           │ [ Send ]                                         │
└───────────────────────────┴──────────────────────────────────────────────────┘
```

Additional tabs:

```text
Connection
Headers
Messages
Events
Timeline
```

---

# 49. SSE Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ SERVER-SENT EVENTS                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ https://api.example.com/events                                 Connected ●    │
├──────────────────────────────────────────────────────────────────────────────┤
│ 10:42:31  event: connected                                                   │
│           data: {"client":"123"}                                             │
│                                                                              │
│ 10:42:34  event: update                                                      │
│           data: {"status":"processing"}                                      │
│                                                                              │
│ 10:42:39  event: complete                                                    │
│           data: {"status":"done"}                                            │
│                                                                              │
│ [ Clear ] [ Pause ] [ Save Stream ]                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

Tabs:

```text
Connection
Headers
Event Stream
Timeline
```

---

# 50. Global Console Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ CONSOLE                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ 10:42:31 INFO Sending GET /v1/users                                         │
│ 10:42:31 INFO DNS lookup                   4 ms                             │
│ 10:42:31 INFO TCP connection               8 ms                             │
│ 10:42:31 INFO TLS handshake               21 ms                             │
│ 10:42:31 INFO Response received           106 ms                            │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 51. Global Network Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ NETWORK                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Time      Method  URL                  Status   Duration                     │
│ 10:42     GET     /users               200      142ms                        │
│ 10:39     POST    /users               201      183ms                        │
│ 10:32     GET     /products            200       91ms                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 52. Global Tests Panel

Shows test activity across the current workspace:

```text
Passed    84
Failed     3
Skipped    2
```

Filters:

```text
Request
Collection
File
Environment
Date
```

---

# 53. Global Variables Panel

Shows effective variables.

```text
Variable
├── Global
├── Workspace
├── Environment
├── Collection
├── Request
└── Runtime
```

Example:

```text
base_url       https://api.dev.com
access_token   •••••••••
user_id        42
timestamp      1720000000
```

The UI should show which scope produced each value.

---

# 54. Global Cookies Panel

Displays:

```text
Domain
Path
Name
Value
Secure
HttpOnly
SameSite
Expires
```

Supports:

* inspect
* search
* delete
* clear
* export

---

# 55. Settings

Settings should be a full-page utility rather than a miscellaneous modal.

```text
Settings
├── General
├── Appearance

---

## Code Review Gate (`/code-review` — two-axis)

- [ ] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [ ] Spec: this milestone (`Milestones/` + Phase) vs implementation (`ROADMAP.md` DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`
