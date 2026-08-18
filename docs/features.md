# Reqly — Feature Set

> **Status:** Draft
> **Product Name:** TBD
> **Positioning:** A local-first, developer-focused API client built around privacy, Git-native workflows, extensibility, API testing, API contracts, and support for modern API protocols.

---

# 1. Product Principles

## 1.1 Local-First

Store projects, collections, environments, scripts, tests, and API materials locally by default.

## 1.2 Offline Support

Allow users to work with local projects, collections, documentation, tests, and previously available data without an active internet connection.

## 1.3 Plain-Text & Git-Native

Store project data in human-readable files that can be inspected, edited, diffed, reviewed, and versioned with Git.

## 1.4 No Account Required

The application should be usable without creating an online account.

## 1.5 Free

Core functionality should be available without requiring a paid subscription.

## 1.6 Open Source

The application and its core functionality should be openly available for inspection and contribution.

## 1.7 Privacy by Default

API requests, responses, credentials, collections, and sensitive project information should remain local unless the user explicitly enables a feature that requires external communication.

## 1.8 No Traffic Telemetry

API request and response traffic should never be collected or transmitted as telemetry.

## 1.9 Cross-Platform

Provide native applications for:

* macOS
* Windows
* Linux

## 1.10 Work With Any API

Support multiple API styles and communication protocols rather than being limited to REST.

---

# 2. API Protocol Support

## 2.1 REST

Full REST API request building, execution, and inspection.

### Features

* HTTP methods
* URL
* Headers
* Query parameters
* Path parameters
* Request body
* Authentication
* Cookies
* Redirect handling
* Request/response history
* Code generation
* File upload
* File download
* Streaming
* Server-Sent Events

---

## 2.2 GraphQL

Dedicated tooling for GraphQL APIs.

### Features

* GraphQL request interface
* Query editor
* Variables
* Query builder
* Schema introspection
* Query autocomplete
* Mutation autocomplete
* Fragment support
* Query validation
* Type navigation
* Schema documentation

---

## 2.3 gRPC

Native gRPC request and service tooling.

### Features

* Create requests
* Protocol Buffer (`.proto`) support
* Server reflection
* Service discovery
* Method discovery
* Message schema browser
* Metadata editor
* Unary requests
* Client streaming
* Server streaming
* Bidirectional streaming

---

## 2.4 WebSocket

Tools for persistent bidirectional connections.

### Features

* Connection management
* Request interface
* Message composer
* Message types
* Incoming/outgoing message inspection
* Connection status
* Message history

---

## 2.5 Server-Sent Events

Support APIs that continuously stream server-generated events over HTTP.

### Features

* SSE request configuration
* Live event stream
* Event inspection
* Connection status
* Event history

---

## 2.6 SOAP

Support SOAP-based enterprise and legacy services.

### Features

* SOAP requests
* WSDL import
* Service discovery
* Operation discovery
* XML request builder
* XML response inspection

---

## 2.7 MQTT

Support lightweight publish/subscribe messaging.

### Features

* MQTT connections
* Publish
* Subscribe
* Topics
* QoS
* Retained messages
* Will messages
* Authentication
* TLS

---

## 2.8 Socket.IO

Support Socket.IO-based realtime APIs.

### Features

* Connection management
* Events
* Rooms
* Namespaces
* Event payloads
* Connection debugging

---

# 3. HTTP & Network Support

## 3.1 HTTP Versions

Support:

* HTTP/1.1
* HTTP/2
* HTTP/3 where supported by the underlying networking stack

## 3.2 Request Configuration

Support:

* Connection timeout
* Request timeout
* Redirect handling
* Compression
* Streaming
* Chunked transfer
* Keep-alive
* Custom request settings

## 3.3 Proxy Support

Support:

* System proxy
* HTTP proxy
* HTTPS proxy
* SOCKS5 proxy
* Proxy per environment
* Proxy per request

## 3.4 TLS & Certificates

Support:

* TLS configuration
* Certificate inspection
* Certificate validation
* Client certificates
* Mutual TLS
* Custom certificate authorities

---

# 4. Import & Export

Allow users to migrate existing API projects and share API definitions.

## Supported Import Sources

* Postman
* Insomnia
* OpenAPI
* Swagger
* cURL
* SOAP / WSDL
* HAR

## Export

Support exporting:

* Collections
* OpenAPI specifications
* Requests
* Responses
* HAR
* Documentation
* Test results

## Import Goals

* Preserve request structure
* Preserve environments and variables where possible
* Preserve authentication configuration
* Preserve scripts where possible
* Convert collections into native project structures
* Report unsupported or incompatible features

---

# 5. OpenAPI & API Specification

OpenAPI should be treated as a first-class API definition rather than only an import format.

## 5.1 OpenAPI Support

Support:

* OpenAPI 2.x
* OpenAPI 3.x
* OpenAPI 3.1

## 5.2 OpenAPI Editor

Provide an editor for creating and modifying OpenAPI specifications.

## 5.3 Specification Validation

Validate OpenAPI documents and report:

* Invalid schemas
* Missing fields
* Invalid references
* Invalid endpoint definitions
* Specification inconsistencies

## 5.4 Endpoint Explorer

Browse an API directly from its OpenAPI definition.

## 5.5 Generate Requests

Generate request templates from OpenAPI definitions.

## 5.6 Generate Documentation

Generate API documentation directly from specifications.

## 5.7 Generate Mocks

Generate mock endpoints and responses from API specifications.

---

# 6. Schema & Contract Management

## 6.1 JSON Schema

Support:

* Schema editing
* Schema validation
* Schema inspection
* Schema generation

## 6.2 GraphQL Schema

Provide schema browsing and validation.

## 6.3 Protobuf Schema

Provide service and message schema inspection.

## 6.4 XML / XSD

Support XML schema validation where applicable.

## 6.5 Schema Visualization

Visualize relationships between types, objects, endpoints, and schemas.

## 6.6 Contract Testing

Validate API responses against an API contract.

Example:

```text
OpenAPI Schema
      ↓
Send Request
      ↓
Receive Response
      ↓
Validate Response
      ↓
PASS / FAIL
```

---

# 7. Request Builder

## 7.1 Request Interface

A unified interface for creating and configuring API requests.

### Request Components

* URL
* HTTP method
* Path parameters
* Query parameters
* Headers
* Body
* Authentication
* Variables
* Scripts
* Certificates
* Proxy
* Request settings

---

# 8. Request Body Builder

Provide specialized editors for common request body formats.

### Supported Types

* JSON
* XML
* Form Data
* URL-encoded form
* Raw text
* Binary
* File upload
* GraphQL
* Multipart

### Features

* Syntax highlighting
* Formatting
* Validation
* Schema-aware editing
* Drag-and-drop file upload

---

# 9. Request Inheritance

Allow configuration to be inherited through the project hierarchy.

Example:

```text
Workspace
   ↓
Collection
   ↓
Folder
   ↓
Request
```

Possible inherited properties:

* Base URL
* Headers
* Authentication
* Variables
* Scripts
* Certificates
* Proxy
* Request settings

Individual requests can override inherited values.

---

# 10. Request Templates

Provide reusable request templates for common API workflows.

Examples:

* Authenticated REST request
* JSON POST
* Paginated GET
* File upload
* GraphQL query
* OAuth request
* Multipart request

Templates should support variables and inherited configuration.

---

# 11. Response Inspection

## 11.1 Response Overview

Display:

* Status code — shipped
* Status text — shipped
* Response time — shipped
* Response size — shipped
* Headers — shipped
* Cookies — shipped (view from `Set-Cookie` response headers; persistence is separate, see §12)
* Request metadata — shipped (proto)

## 11.2 Response Data

Support:

* JSON
* XML
* HTML
* Plain text
* CSV
* Images
* PDF
* Binary data

## 11.3 Response Views

Provide:

* Raw view — shipped
* Pretty-printed view — shipped (JSON pretty-print + XML indentation)
* Tree view — shipped (interactive expand/collapse JSON tree)
* Table view — pending
* Hex/binary view
* Preview view

## 11.4 Response Search

Search and navigate through large response bodies — shipped (case-insensitive substring search across the current view, with match highlighting).

## 11.5 JSONPath

Filter and query JSON responses — shipped (dependency-free evaluator: `$` root, dot/bracket segments, wildcard `*`, array indexes; match list with canonical paths; specific per-segment errors).

## 11.6 XPath

Query XML responses using XPath — pending.

## 11.7 Response Actions

Allow users to:

* Copy value — pending
* Copy JSONPath — pending
* Copy response — shipped (body or headers)
* Download response — shipped (Content-Disposition filename, else derived from Content-Type)
* Format response — shipped (JSON/XML pretty)
* Save response as example — pending

---

# 12. Cookies

Provide a complete cookie management system.

### Features

* Cookie jar — pending (separate roadmap item §1.6 Cookies persistence)
* View cookies — shipped (parse `Set-Cookie` response headers in the desktop response viewer: name, value, domain, path, expiry, Secure/HttpOnly/SameSite flags)
* Edit cookies — pending
* Delete cookies — pending
* Domain/path matching — pending
* Secure cookies — shipped (flag rendered)
* HttpOnly — shipped (flag rendered)
* SameSite — shipped (flag rendered)
* Cookie persistence — pending (out of scope for milestone 14; display only)
* Clear cookies per environment — pending
* Clear cookies per workspace — pending

---

# 13. Request & Response Examples

Allow users to save representative request and response examples.

Useful for:

* Documentation
* Mock generation
* Testing
* API design
* Debugging
* Team collaboration

---

# 14. Variables & Environments

Provide a flexible variable hierarchy.

## Variable Types

* Global environment variables
* Environment variables
* Collection variables
* Folder variables
* Request variables
* Runtime variables
* Prompt variables
* Process environment variables

## 14.1 Variable Interpolation

Resolve variables inside:

* URLs
* Headers
* Parameters
* Request bodies
* Scripts
* Authentication
* Other supported project configuration

## 14.2 Environment Management

Support:

* Development
* Staging
* Production
* Custom environments

## 14.3 Environment Validation

Detect:

* Missing variables
* Invalid variables
* Unused variables
* Secret exposure

## 14.4 Environment Diff

Compare two environments and identify:

* Added variables
* Removed variables
* Changed values
* Missing variables
* Secret differences

---

# 15. Dynamic Values & Template Tags

Allow dynamically generated values to be inserted into requests.

Examples:

* UUID
* Timestamp
* Random values
* Runtime values
* Environment values
* Generated tokens

Users should also be able to create custom template tags through plugins.

---

# 16. Authentication & Authorization

Support common authentication mechanisms.

### Built-in Authentication

* Basic
* Bearer
* JWT
* Digest
* NTLM
* OAuth 1.0
* OAuth 2.0
* AWS Signature
* Akamai EdgeGrid

### Additional Support

* API keys
* Custom authentication
* Authentication inheritance
* Authentication plugins

---

# 17. OAuth 2.0

Support common OAuth 2.0 flows.

### Flows

* Authorization Code + PKCE — shipped (RFC 6749 §4.1 + RFC 7636, ADR 0007)
* Client Credentials — shipped
* Device flow — shipped (RFC 8628, ADR 0008): headless approval via a printed verification URI + code, RFC poll semantics (`interval`, `authorization_pending`, `slow_down`, `expired_token`/`access_denied`)
* Password Credentials (deprecated in OAuth 2.1 — deferred)

### Features

* System browser authentication (`reqly auth login`, one-shot loopback callback on an ephemeral `127.0.0.1` port)
* Device-flow login (`reqly auth login --flow device` — verification URI + code, waits for approval, no browser)
* Custom-scheme callbacks (`reqly://` deep links for the desktop app; loopback remains the CLI default)
* First-request auto-login (a request with `grant_type: authorization_code` and no cached token opens the browser; device flows print the verification URI to stderr)
* Token storage behind `secrets.Store`: per-workspace `.reqly/tokens.json` (0600, default) or the OS keychain (`--store keychain` / `REQLY_TOKEN_STORE`; keychain default on desktop, file fallback with a warning)
* Token refresh (expiry-skewed proactive + reactive 401 retry-once)
* Automatic token refresh via the refresh-token grant (RFC 6749 §6) — the browser is never re-opened while a refresh token exists
* Refresh-token handling (rotation when the server returns a new refresh token; kept otherwise)
* Token expiration detection
* Desktop auth panel (login/status/logout in the sidebar; masked token list, flow picker, JSON config editor)
* OAuth configuration (`grant_type`, `token_url`, `authorization_url`, `device_authorization_url`, `client_id`, `client_secret`, `redirect_uri`, `scope`, `audience`, `token_name`)
* Certificate management

---

# 18. JWT Tooling

Provide dedicated JWT inspection and debugging tools.

### Features

* JWT decoder
* Header inspection
* Claims viewer
* Expiration detection
* Signature information
* JWT signing/verification where supported

---

# 19. Secret Management

## 19.1 Encrypted Secrets

Encrypt sensitive values at rest.

## 19.2 OS Keychain

Use the operating system's secure credential storage — shipped (KeychainStore behind `secrets.Store` via go-keyring: Secret Service / Keychain / WinCred; 0600 key index for enumeration; `--store keychain` / `REQLY_TOKEN_STORE` selection with file-store fallback).

## 19.3 Secret Variables

Allow requests and environments to reference secrets without exposing their values.

## 19.4 Secret Masking

Prevent secrets from appearing in:

* UI output
* Logs
* Test output
* Debugging information
* Generated documentation

## 19.5 `.env` Support

Read environment values from `.env` files.

---

# 20. External Secret Managers

Support external secret-management systems.

### Initial Integrations

* HashiCorp Vault
* AWS Secrets Manager
* Azure Key Vault

### Capabilities

* Secret retrieval
* Secret references
* Secure variable resolution
* Authentication
* Migration support

---

# 21. Scripts & Automation

Use JavaScript for request customization and automation.

## 21.1 Pre-Request Scripts

Run JavaScript before sending a request.

Use cases:

* Generate signatures
* Generate dynamic values
* Modify request data
* Prepare variables
* Authentication preparation

## 21.2 Post-Request Scripts

Run JavaScript after receiving a response.

Use cases:

* Extract tokens
* Validate responses
* Set variables
* Transform data

## 21.3 JavaScript Runtime

Provide:

* JavaScript reference
* Built-in libraries
* External libraries
* JavaScript files
* Dynamic variables
* Request object
* Response object
* Response queries
* Variable access

---

# 22. Request Chaining

Allow requests to consume values generated by previous requests.

Example:

```text
Login
  ↓
Extract access token
  ↓
Create project
  ↓
Extract project ID
  ↓
Upload file
```

Support chaining through:

* Variables
* Scripts
* Response queries
* Runtime values

---

# 23. Testing

## 23.1 Assertions

Support assertions against:

* Status code
* Headers
* Response body
* JSON values
* XML values
* Response time
* Schema validity

## 23.2 Manual Testing

Run tests interactively during API development.

## 23.3 Automated Testing

Execute tests automatically through runners and CLI.

## 23.4 Data-Driven Testing

Execute the same test suite against multiple datasets.

---

# 24. Collection Runner

Execute multiple requests and tests sequentially.

### Use Cases

* Regression testing
* Smoke testing
* API workflows
* Environment validation
* Integration testing

---

# 25. Chain Runner

Execute dependent API workflows where later requests consume values from earlier requests.

Support:

* Sequential execution
* Conditional execution
* Variable passing
* Assertions
* Script execution
* Failure handling

---

# 26. Bulk Request Execution

Execute the same request against many inputs.

### Input Sources

* CSV
* JSON
* Variables
* Generated datasets

### Execution Modes

* Sequential
* Parallel
* Configurable concurrency

Useful for:

* Batch API testing
* Data validation
* Large-scale request testing
* Reproducing issues across many resources

---

# 27. Pagination Runner

Automatically execute paginated APIs.

### Supported Patterns

* Page / pageSize
* Offset / limit
* Cursor
* Link headers

### Features

* Automatic page traversal
* Stop conditions
* Response aggregation
* Export aggregated results
* Configurable maximum pages

---

# 28. Retry & Resilience

Provide configurable retry behavior.

### Features

* Retry count
* Retry delay
* Exponential backoff
* Retry on status codes
* Retry on network errors
* Timeout handling
* Rate-limit handling
* 429 handling

---

# 29. Performance Testing

Provide lightweight API performance testing.

### Configuration

* Request count
* Concurrency
* Duration
* Rate limits

### Metrics

* Requests per second
* Average latency
* Median latency
* P95
* P99
* Minimum latency
* Maximum latency
* Error rate
* Status distribution

---

# 30. Mock Server

Provide local API mocking capabilities.

## Features

* Generate mocks from OpenAPI
* Generate mocks from examples
* Dynamic responses
* Request matching
* Response templates
* Multiple response scenarios
* Custom status codes
* Delay simulation
* Error simulation
* Stateful mocks
* Local mock server
* CLI mock server

---

# 31. API Monitoring

Allow requests and tests to be executed on a schedule.

### Features

* Scheduled requests
* Scheduled collections
* Health checks
* Response assertions
* Latency monitoring
* Availability monitoring
* Failure detection
* Execution history
* Alerts

Monitoring should be compatible with the local-first architecture and optionally run through the CLI or a self-hosted service.

---

# 32. API Diff & Contract Diff

Compare API definitions and identify changes.

### Detect

* Added endpoints
* Removed endpoints
* Changed parameters
* Changed schemas
* Changed authentication
* Changed response types
* Breaking changes

Support comparing:

* OpenAPI versions
* Schemas
* Requests
* Responses
* Environments

---

# 33. Breaking Change Detection

Automatically identify potentially breaking API changes.

Examples:

```text
❌ Endpoint removed
❌ Required parameter added
❌ Response field removed
❌ Response type changed
❌ Authentication requirement changed

⚠️ Optional parameter changed
```

This should be usable locally and in CI.

---

# 34. API Dependency Graph

Visualize dependencies between requests and API workflows.

Example:

```text
Login
  ↓
Create User
  ↓
Create Project
  ↓
Upload File
  ↓
Process File
```

Dependencies can be derived from request chaining, variables, and scripts.

---

# 35. Request & Response Diff

Compare requests or responses.

### Examples

* Request A vs Request B
* Development vs Production
* Before vs After
* Expected vs Actual

Support:

* JSON structural diff
* Header diff
* Body diff
* Status diff

---

# 36. Request Replay

Replay historical requests from the request history.

### Features

* Replay exact request
* Replay with modified variables
* Replay against another environment
* Replay multiple times
* Replay from captured network traffic

---

# 37. HAR Support

Import and export HTTP Archive files.

### Use Cases

* Browser traffic capture
* Debugging
* Request reproduction
* Sharing network sessions
* Migrating browser requests into collections

---

# 38. Browser Integration

Allow requests to be imported from browser developer tools.

Support workflows such as:

```text
Browser DevTools
      ↓
Copy as cURL
      ↓
API Client
      ↓
Request
```

Potential direct integrations:

* Chrome DevTools
* Firefox DevTools
* Safari Web Inspector

---

# 39. Workspaces & Collections

## 39.1 Workspaces

Top-level containers for API projects.

## 39.2 Collections

Group related API requests.

## 39.3 Nested Folders

Organize collections hierarchically.

Example:

```text
Workspace
├── Authentication
│   ├── Login
│   └── Refresh Token
├── Users
│   ├── Get User
│   └── Update User
└── Products
    ├── List Products
    └── Create Product
```

---

# 40. Filesystem Mirroring

Mirror workspaces to the filesystem.

This enables:

* Git version control
* Local backups
* File synchronization
* Dropbox synchronization
* Direct file inspection
* Direct editing
* CLI workflows

---

# 41. Git Integration & Collaboration

## 41.1 Git-Native Projects

Store collections, environments, scripts, tests, schemas, mocks, and documentation as version-controlled files.

## 41.2 Git Strategies

Provide configurable strategies for project synchronization, branching, merging, and collaboration.

## 41.3 Collaboration Through Git

Allow teams to collaborate using normal Git workflows:

* Branches
* Commits
* Pull requests
* Reviews
* Merges
* History

## 41.4 GUI Git Integration

Provide Git operations directly inside the application.

## 41.5 CLI Collaboration

Support Git-based API project workflows through the command line.

---

# 42. Git Provider Integrations

Support:

* GitHub
* GitLab
* Bitbucket
* Azure DevOps

Provider integrations may support authentication, repository discovery, cloning, synchronization, and pull-request workflows.

---

# 43. API Documentation

## 43.1 Documentation Generation

Generate human-readable API documentation from collections and API definitions.

## 43.2 RESTful API Documentation

Document:

* Endpoints
* Parameters
* Headers
* Authentication
* Request bodies
* Responses
* Examples

## 43.3 GraphQL API Documentation

Document:

* Schema
* Queries
* Mutations
* Variables
* Types

## 43.4 Realtime API Documentation

Document:

* WebSocket events
* SSE events
* Socket.IO events
* Messages
* Payloads

---

# 44. API Changelog

Generate changelogs from API specification and Git changes.

Example:

```text
v1.4.0

Added
+ POST /users/avatar

Changed
~ GET /users/{id}

Breaking
- DELETE /users/{id}/legacy
```

---

# 45. Debugging & Diagnostics

## 45.1 Request Timeline

Provide a detailed request lifecycle timeline.

Potential stages:

* DNS
* Connection
* TLS
* Request
* Server processing
* Response
* Transfer

## 45.2 Developer Tools

Provide application-level debugging tools.

## 45.3 Inspections

Inspect:

* Request
* Response
* Authentication
* Variables
* Scripts
* Runtime state
* Network behavior

---

# 46. Network Interceptor

Capture or intercept supported network traffic for debugging and request reproduction.

Potential capabilities:

* Capture requests
* Inspect requests
* Import captured requests
* Modify requests
* Replay requests

---

# 47. CLI

Provide a complete command-line interface.

Example:

```bash
api-client run users.list
api-client test
api-client collection run users
api-client mock
api-client validate
api-client diff
api-client docs
```

### CLI Use Cases

* Local development
* CI/CD
* Automated testing
* Git hooks
* API validation
* Mock servers
* Documentation generation

---

# 48. CI/CD & Automation

Support API workflows outside the desktop application.

### Use Cases

* API regression testing
* Integration testing
* Contract testing
* Breaking-change detection
* Collection execution
* Environment validation
* Documentation generation
* Mock server deployment

---

# 49. API Project as Code

Treat the entire API project as a version-controlled development artifact.

Example:

```text
api/
├── collections/
├── environments/
├── schemas/
├── mocks/
├── tests/
├── scripts/
├── docs/
└── project.yaml
```

The complete project should be reproducible from the filesystem and Git repository.

---

# 50. Extensibility & Plugins

Provide a plugin system for extending the client.

Plugins may extend:

* Authentication
* Template tags
* Request processing
* Response processing
* Protocols
* UI
* Integrations
* Developer workflows

---

# 51. UI Customization

## Built-in Themes

Provide multiple built-in themes.

## Custom Themes

Allow users to create and install custom themes.

## UI Extensions

Allow plugins to add functionality to the application interface.

---

# 52. Developer Experience

## Keyboard Shortcuts

Provide shortcuts for frequently used operations.

## Command Spotlight

Quickly search for and execute:

* Commands
* Requests
* Collections
* Workspaces
* Settings

## Context Menu

Provide context-aware actions throughout the application.

## Widgets

Allow customizable widgets for frequently used tools and information.

## Code Snippets

Provide reusable API and code snippets.

---

# 53. AI Features

Provide optional AI-assisted API development.

Potential capabilities:

* Request generation
* Response explanation
* Test generation
* Documentation generation
* Error analysis
* Request transformation
* Schema assistance
* API specification generation
* Breaking-change explanation

AI features should remain optional and should not be required for core API-client functionality.

---

# 54. Personal Access Tokens

Support personal access tokens for authenticated integrations and Git workflows.

---

# 55. Collaboration & Access

Potential future collaboration capabilities:

* Shared projects
* Team workspaces
* Access control
* Project permissions
* Personal access tokens
* Self-hosted collaboration server
* Organization policies
* Audit logs
* SSO / SCIM

These should remain compatible with the local-first philosophy.

---

# 56. Core Feature Priority

## P0 — Core API Client

The minimum set required to make the product a serious API client.

* REST
* GraphQL
* gRPC
* WebSocket
* SSE
* SOAP
* Request builder
* Response viewer
* JSON/XML support
* Headers
* Parameters
* Request bodies
* File upload/download
* Variables
* Environments
* Authentication
* Secrets
* Cookies
* Collections
* Workspaces
* Request history
* Request chaining
* Scripts
* Tests
* Collection runner
* Git-native storage
* Import/export
* OpenAPI
* JSON Schema
* CLI
* Cross-platform desktop application

---

# 57. P1 — Differentiating Features

Features that can make the product significantly more capable than a basic API client.

* OpenAPI editor
* Schema validation
* Contract testing
* Mock server
* API diff
* Breaking-change detection
* Request/response diff
* Environment diff
* Request inheritance
* HAR import/export
* JWT tooling
* GraphQL introspection
* gRPC reflection
* Advanced HTTP support
* Proxy and TLS controls
* Pagination runner
* Bulk request execution
* Retry/resilience
* API documentation
* CI/CD support

---

# 58. P2 — Power-User Features

Advanced functionality for experienced developers and teams.

* API monitoring
* Performance testing
* MQTT
* Socket.IO
* API dependency graph
* Request replay
* API changelog
* Browser integrations
* Advanced mock state
* Visual workflow builder
* Advanced network interception
* Self-hosted automation

---

# 59. P3 — Ecosystem & Enterprise

Long-term ecosystem and organization features.

* Plugin marketplace
* Theme marketplace
* Git provider integrations
* Shared workspaces
* Team collaboration
* Self-hosted collaboration server
* Enterprise SSO
* SCIM
* Audit logs
* Organization policies
* Enterprise secret management
* Advanced access control

---

# 60. Feature Summary

| Area                | Capabilities                                                               |
| ------------------- | -------------------------------------------------------------------------- |
| **Platform**        | Local-first, offline, cross-platform, open source, no account              |
| **Privacy**         | Local storage, encrypted secrets, OS keychain, no traffic telemetry        |
| **Protocols**       | REST, GraphQL, gRPC, WebSocket, SSE, SOAP, MQTT, Socket.IO                 |
| **HTTP**            | HTTP/1.1, HTTP/2, HTTP/3, proxies, TLS, certificates                       |
| **Import**          | Postman, Insomnia, OpenAPI, Swagger, cURL, WSDL, HAR                       |
| **Request Builder** | URL, methods, headers, parameters, body, files, auth                       |
| **Response**        | JSON, XML, HTML, CSV, images, PDF, binary, tree/table/raw views            |
| **Variables**       | Global, environment, collection, folder, request, runtime, prompt, process |
| **Authentication**  | Basic, Bearer, JWT, Digest, NTLM, OAuth, AWS, Akamai, API keys             |
| **Secrets**         | Encryption, keychain, secret variables, masking, `.env`, Vault, AWS, Azure |
| **Scripting**       | Pre-request, post-request, JavaScript, dynamic variables                   |
| **Testing**         | Assertions, manual, automated, data-driven, contract testing               |
| **Automation**      | Collection runner, chain runner, pagination, bulk execution, retry         |
| **Mocking**         | OpenAPI mocks, dynamic responses, stateful mocks, local mock server        |
| **Monitoring**      | Scheduled requests, health checks, latency, availability, alerts           |
| **API Contracts**   | OpenAPI, JSON Schema, validation, diff, breaking-change detection          |
| **GraphQL**         | Introspection, schema browser, autocomplete, query builder                 |
| **gRPC**            | Proto files, reflection, service discovery, streaming                      |
| **Git**             | Git-native projects, diff, branches, collaboration, providers              |
| **Documentation**   | REST, GraphQL, realtime API docs, generated documentation                  |
| **Debugging**       | Timeline, DevTools, inspection, interceptor, replay                        |
| **CLI**             | Run, test, mock, validate, diff, docs, CI/CD                               |
| **Extensibility**   | Plugins, custom auth, template tags, UI extensions                         |
| **Customization**   | Themes, widgets, shortcuts, context menus, snippets                        |
| **AI**              | Request generation, testing, documentation, analysis, schema assistance    |
| **Enterprise**      | Teams, SSO, SCIM, audit logs, policies, access control                     |

---

# 61. Product Vision

The product should not simply be positioned as another HTTP request tool.

The broader vision is:

> **A local-first API development environment where requests, tests, schemas, mocks, documentation, environments, and API contracts live together as version-controlled project files.**

The architecture can be centered around:

```text
                    API Project
                         │
        ┌────────────────┼────────────────┐
        │                │                │
   Collections       Schemas          Environments
        │                │                │
   Requests           OpenAPI         Variables
        │                │                │
   Scripts            Contracts        Secrets
        │                │
     Tests ─────────── Mocks
        │
     Runner
        │
       CLI
        │
   CI / Automation
        │
       Git
```

The strongest product differentiators are therefore:

1. **Local-first architecture**
2. **Git-native API projects**
3. **No account requirement**
4. **Privacy-first operation**
5. **Broad protocol support**
6. **First-class OpenAPI and schema tooling**
7. **Integrated API testing**
8. **Contract testing and breaking-change detection**
9. **Local mocking**
10. **CLI + desktop workflows**
11. **Plugin architecture**
12. **Reproducible API projects**
13. **Strong developer experience**
14. **Open-source foundation**

The end goal is an **API development platform that happens to include an excellent API client**, rather than an API client that accumulated 900 features after someone discovered checkboxes.
