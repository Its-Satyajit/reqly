# API Client — Technology Stack

> **Status:** Proposed / Selected
> **Architecture Goal:** Small footprint, fast startup, low memory usage, native desktop performance, and a shared core for Desktop, CLI, automation, and MCP.

---

## 1. Technology Overview

The API client will use a **Go + Wails + Goja** architecture, with React and TypeScript powering the user interface.

### Core Stack

| Layer             | Technology                   | Purpose                                                                                 |
| ----------------- | ---------------------------- | --------------------------------------------------------------------------------------- |
| Desktop Framework | **Wails**                    | Native desktop application using the system WebView                                     |
| Backend / Core    | **Go**                       | API engine, networking, storage, authentication, testing, Git, and application services |
| Scripting         | **Goja**                     | Embedded JavaScript runtime for pre/post-request scripts and automation                 |
| Frontend          | **React + TypeScript**       | Interactive desktop UI                                                                  |
| Build Tool        | **Vite**                     | Fast frontend development and production builds                                         |
| Styling           | **Tailwind CSS**             | Lightweight utility-based styling                                                       |
| UI Components     | **shadcn/ui + Base UI**     | Accessible, reusable interface components                                               |
| State Management  | **Zustand**                  | Lightweight application state management                                                |
| Code Editor       | **CodeMirror 6**             | Editors for JSON, JavaScript, GraphQL, XML, YAML, and other formats                     |
| Project Storage   | **Plain-text files**         | Human-readable, Git-friendly project data                                               |
| Local Metadata    | **SQLite**                   | History, indexing, cache, and other application metadata                                |
| Git               | **Go Git library**           | Git-native project management and collaboration                                         |
| Secrets           | **OS Keychain**              | Secure storage of sensitive credentials                                                 |
| CLI               | **Go**                       | Command-line access using the same core engine                                          |
| MCP               | **Go**                       | Built-in Model Context Protocol server                                                  |
| Testing           | **Go + Vitest + Playwright** | Backend, frontend, and end-to-end testing                                               |

---

# 2. Desktop Application

## Wails

**Wails** will be used as the desktop application framework.

It provides a native application shell while allowing the UI to be built with modern web technologies. The application uses the operating system's WebView rather than bundling a complete Chromium runtime.

### Why Wails?

* Small application footprint
* Fast startup
* Native desktop integration
* Cross-platform support
* React/TypeScript compatibility
* Direct Go ↔ JavaScript communication
* No Node.js runtime required in production
* No bundled Chromium browser

Target platforms:

* Linux
* Windows
* macOS

---

# 3. Backend & Core Engine

## Go

Go will be the primary implementation language for the application core.

The Go layer will contain the actual API-client functionality rather than putting business logic inside the React application.

### Responsibilities

* HTTP requests
* WebSocket communication
* SSE
* gRPC
* Authentication
* Variables
* Collections
* Environments
* Request chaining
* Tests
* Collection runner
* Mock server
* OpenAPI
* Git integration
* Secret management
* Import/export
* Request history
* CLI
* MCP
* Application services

### Why Go?

Go is well suited to a networking-heavy application because it provides:

* Fast startup
* Low runtime overhead
* Excellent concurrency
* Strong networking libraries
* Simple deployment
* Native filesystem access
* Strong HTTP support
* Straightforward cross-platform compilation

---

# 4. JavaScript Runtime

## Goja

**Goja** will provide the embedded JavaScript runtime.

This allows the API client to support JavaScript-based scripting without requiring Node.js or a separate JavaScript runtime.

### Use Cases

* Pre-request scripts
* Post-request scripts
* Test scripts
* Dynamic variables
* Request transformation
* Response processing
* Plugin functionality

Example architecture:

```text
API Request
     │
     ▼
Request Pipeline
     │
     ├── Variables
     ├── Pre-request Script
     ├── Authentication
     └── Request
             │
             ▼
          Response
             │
     ├── Post-request Script
     ├── Assertions
     └── Variable Extraction
```

### Performance Consideration

Goja should be **initialized lazily**.

The basic API client should not pay the runtime cost of the JavaScript engine when no scripting functionality is being used.

---

# 5. Frontend

## React + TypeScript

The application interface will be built with React and TypeScript.

React is well suited for the highly interactive nature of an API client.

### UI Areas

* Workspace explorer
* Collection tree
* Request editor
* Response viewer
* Environment manager
* Authentication configuration
* Test editor
* Script editor
* Git interface
* API documentation
* Settings
* Command palette
* Request history

TypeScript will provide type safety across the frontend and the Wails-generated Go bindings.

---

# 6. Frontend Build System

## Vite

Vite will be used for frontend development and production builds.

### Benefits

* Fast development server
* Fast Hot Module Replacement
* Efficient production builds
* Simple React integration
* Minimal configuration

The desktop application does not require server-side rendering, so a traditional SPA architecture is preferred.

---

# 7. Styling

## Tailwind CSS

Tailwind CSS will provide the application's styling system.

It will be used for:

* Layout
* Spacing
* Typography
* Responsive behavior
* Themes
* Component styling
* Design tokens

The goal is to maintain a compact and consistent UI without introducing a large runtime styling system.

---

# 8. UI Components

## shadcn/ui + Base UI

Use shadcn/ui and Base UI primitives for common application components.

### Components

* Dialogs
* Dropdowns
* Menus
* Tabs
* Tooltips
* Popovers
* Selects
* Command palette
* Context menus
* Forms

Components should remain composable and customizable rather than locking the application into a large visual framework.

---

# 9. State Management

## Zustand

Zustand will handle client-side application state.

### Example State

* Current workspace
* Selected collection
* Open requests
* Active environment
* UI layout
* Tabs
* User preferences
* Request editor state

State should remain lightweight and localized where possible.

---

# 10. Code Editor

## CodeMirror 6

CodeMirror 6 will be used for editable API-related content.

### Supported Editors

* JSON
* JavaScript
* GraphQL
* XML
* YAML
* Markdown
* Request bodies
* OpenAPI definitions

### Features

* Syntax highlighting
* Formatting
* Autocomplete
* Search
* Code folding
* Validation
* Custom language extensions

CodeMirror is preferred over a full IDE-style editor because the application needs several lightweight specialized editors rather than a complete development environment.

---

# 11. Project Storage

## Plain-Text Files

Project data should primarily be stored as human-readable files.

Example:

```text
api-project/
├── project.yaml
├── collections/
│   ├── users/
│   └── products/
├── environments/
│   ├── development.yaml
│   └── production.yaml
├── schemas/
├── tests/
├── scripts/
├── mocks/
└── docs/
```

### Benefits

* Git-friendly
* Human-readable
* Easy to backup
* Easy to diff
* Easy to inspect
* No proprietary database format
* Works naturally with CLI tools

The filesystem should be the **source of truth for API projects**.

---

# 12. Local Metadata

## SQLite

SQLite will be used only where structured local metadata provides a meaningful advantage.

### Potential Uses

* Request history
* Search indexes
* Cached data
* Execution results
* Local analytics
* UI metadata

SQLite should not become the primary storage format for Git-managed API projects.

---

# 13. Git Integration

## Git

Git will be a first-class part of the application.

The project structure should be designed around version-controlled files.

### Features

* Repository initialization
* Commit
* Branches
* Diff
* History
* Pull
* Push
* Merge
* Conflict handling
* Git provider integration

### Goal

A complete API project should be reproducible from a Git repository.

```text
Git Repository
      │
      ▼
API Project
      │
      ├── Collections
      ├── Environments
      ├── Tests
      ├── Scripts
      ├── Schemas
      └── Documentation
```

---

# 14. Secret Management

## OS Keychain

Sensitive values should be stored using the operating system's secure credential storage where possible.

### Examples

* API keys
* OAuth tokens
* Passwords
* Client secrets
* Private credentials

### Principles

* Never store sensitive values unnecessarily
* Mask secrets in the UI
* Avoid exposing secrets in logs
* Separate secrets from Git-managed project data

---

# 15. Networking

## Go Networking Stack

The Go networking stack will provide the foundation for API communication.

### Initial Support

* HTTP/1.1
* HTTP/2
* TLS
* WebSocket
* Server-Sent Events
* gRPC
* Multipart requests
* Streaming
* Proxy support
* Custom certificates

The networking layer should be abstracted behind a common request interface so additional protocols can be added without changing the application architecture.

```text
Request Engine
      │
      ├── HTTP
      ├── WebSocket
      ├── SSE
      ├── gRPC
      └── Future Protocols
```

---

# 16. API Protocol Modules

Protocol implementations should remain modular.

```text
protocols/
├── http/
├── graphql/
├── websocket/
├── sse/
├── grpc/
├── soap/
└── mqtt/
```

Advanced protocol modules can be initialized only when required.

This reduces unnecessary startup work and keeps the default runtime lightweight.

---

# 17. OpenAPI & Schema Support

OpenAPI and schema functionality will be implemented in Go.

### Responsibilities

* OpenAPI parsing
* Specification validation
* Endpoint generation
* Schema validation
* Request generation
* Documentation generation
* Mock generation
* Contract testing
* API diffing

This keeps API-definition functionality available to both the desktop application and CLI.

---

# 18. Testing Engine

The testing engine will live in the Go core.

### Features

* Assertions
* Collection tests
* Request tests
* Data-driven testing
* Contract testing
* Response validation
* Automated test execution
* Test reporting

The same test engine should work from:

```text
Desktop
CLI
CI/CD
MCP
```

---

# 19. Mock Server

The mock server will be implemented in Go.

### Features

* OpenAPI-based mocks
* Example-based mocks
* Dynamic responses
* Request matching
* Response templates
* Delays
* Error simulation
* Stateful mocks
* Local server
* CLI operation

Go's networking and concurrency model make it well suited for running a lightweight local mock server.

---

# 20. CLI

The CLI will use the same Go core as the desktop application.

Example:

```bash
api-client run users.get
api-client test
api-client collection run users
api-client mock
api-client validate
api-client diff
api-client docs
```

### Important Principle

The CLI should not implement a second API engine.

```text
              Go Core
             /       \
            /         \
       Desktop         CLI
        Wails           Go
```

This ensures that requests behave consistently in the GUI, CLI, and CI environments.

---

# 21. MCP Server

The API client will include a built-in MCP server implemented in Go.

Potential tools include:

* List workspaces
* List collections
* Search requests
* Get request
* Run request
* Run collection
* Run tests
* Inspect schemas
* Retrieve responses
* Generate documentation

The MCP layer should directly consume the same Go core used by the desktop application and CLI.

---

# 22. Testing the Application

Use different testing tools for different layers.

### Go

For:

* Core logic
* Networking
* Authentication
* Variables
* Scripts
* Testing engine
* Importers
* Exporters
* Git
* Mock server

### Vitest

For:

* React components
* UI logic
* Frontend utilities
* State management

### Playwright

For:

* End-to-end desktop workflows
* Request creation
* Collection management
* Environment switching
* Test execution
* UI interactions

---

# 23. Performance & Footprint Principles

Small footprint and fast startup are product requirements, not merely implementation details.

### Principles

1. **Use the system WebView**
2. **Do not bundle Chromium**
3. **Do not bundle Node.js**
4. **Use Go as the primary runtime**
5. **Load Goja lazily**
6. **Load advanced protocol modules when required**
7. **Use filesystem storage for project data**
8. **Use SQLite only for metadata**
9. **Avoid unnecessary background processes**
10. **Virtualize large collections and response lists**
11. **Avoid loading entire large responses into the UI**
12. **Keep startup initialization minimal**

---

# 24. Application Architecture

```text
┌───────────────────────────────────────────────┐
│                  Wails Desktop                │
│                                               │
│          React + TypeScript + Vite            │
│          Tailwind + shadcn/ui                 │
│          CodeMirror                           │
└───────────────────────┬───────────────────────┘
                        │
                    Wails Bridge
                        │
┌───────────────────────▼───────────────────────┐
│                    Go Core                    │
│                                               │
│  Request Engine       Authentication          │
│  Variables            Collections             │
│  Environments         Testing                 │
│  OpenAPI              Mock Server             │
│  Git                  Secrets                 │
│  Import / Export      History                 │
│  MCP                  CLI                     │
│                                               │
│  ┌─────────────────────────────────────────┐  │
│  │              Goja Runtime               │  │
│  │       Loaded only when required         │  │
│  └─────────────────────────────────────────┘  │
└───────────────────────────────────────────────┘
```

---

# 25. Shared Core Architecture

The Go core should be independent of Wails.

```text
                    Go Core
                       │
          ┌────────────┼────────────┐
          │            │            │
          ▼            ▼            ▼
       Wails          CLI          MCP
       Desktop        Go           Server
```

This provides:

* One request implementation
* One authentication system
* One variable system
* One testing engine
* One scripting engine
* One Git implementation
* One OpenAPI implementation
* Consistent behavior everywhere

---

# 26. Recommended Project Structure

```text
api-client/
│
├── apps/
│   ├── desktop/
│   │   ├── frontend/
│   │   └── backend/
│   │
│   └── cli/
│
├── internal/
│   ├── core/
│   ├── request/
│   ├── response/
│   ├── auth/
│   ├── variables/
│   ├── collections/
│   ├── environments/
│   ├── scripting/
│   ├── testing/
│   ├── mocking/
│   ├── openapi/
│   ├── graphql/
│   ├── grpc/
│   ├── websocket/
│   ├── git/
│   ├── secrets/
│   ├── history/
│   ├── importer/
│   ├── exporter/
│   └── mcp/
│
├── frontend/
│   ├── components/
│   ├── features/
│   ├── editors/
│   ├── stores/
│   └── app/
│
├── go.mod
├── package.json
└── pnpm-workspace.yaml
```

---

# 27. Technology Decisions

## Selected

* **Go** for the application core
* **Wails** for desktop
* **Goja** for JavaScript execution
* **React + TypeScript** for UI
* **Vite** for frontend builds
* **Tailwind CSS** for styling
* **shadcn/ui + Base UI** for components
* **Zustand** for UI state
* **CodeMirror 6** for editors
* **Plain-text files** for project storage
* **SQLite** for local metadata
* **OS keychain** for secrets
* **Go CLI** for automation
* **Go MCP** for AI/tool integration

## Core Design Goal

> **A native-feeling API client with the convenience of a modern TypeScript frontend and the small footprint, concurrency, and networking performance of a Go application.**

The architecture should keep the default application lightweight while allowing advanced features such as scripting, gRPC, mocking, contract testing, MCP, monitoring, and plugins to be loaded or initialized only when needed.
