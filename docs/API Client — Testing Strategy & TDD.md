# API Client — Testing Strategy & TDD

> **Status:** Draft
> **Development Methodology:** Test-Driven Development (TDD)
> **Primary Goal:** Build a reliable, deterministic, and regression-resistant API client where critical behavior is defined by executable tests before implementation.

---

# 1. Testing Philosophy

The project will follow **Test-Driven Development (TDD)** as the primary development methodology.

The fundamental development cycle is:

```text
Write Test
    ↓
RED
    ↓
Implement Minimum Code
    ↓
GREEN
    ↓
Refactor
    ↓
Repeat
```

The objective is not to maximize the coverage percentage at all costs.

Tests should provide confidence that the application behaves correctly under:

* Normal conditions
* Edge cases
* Invalid input
* Network failures
* Authentication failures
* Concurrency
* Security-sensitive situations
* Regression scenarios

**Coverage is a quality indicator, not the definition of test quality.**

---

# 2. TDD Workflow

## 2.1 Red

Write a failing test describing the desired behavior.

```text
Requirement
    ↓
Test
    ↓
FAIL
```

The test should fail for the expected reason.

---

## 2.2 Green

Implement the minimum production code necessary to make the test pass.

```text
Failing Test
    ↓
Minimum Implementation
    ↓
PASS
```

Avoid implementing functionality that is not required by the current behavior being tested.

---

## 2.3 Refactor

After the test passes:

* Improve structure
* Remove duplication
* Improve naming
* Simplify implementation
* Improve performance where appropriate
* Improve error handling

All existing tests must remain passing during refactoring.

---

# 3. Testing Pyramid

The project will use three primary levels of automated testing.

```text
                 ┌─────────────────┐
                 │    E2E Tests    │
                 │ Critical flows  │
                 └────────┬────────┘
                          │
                ┌─────────▼─────────┐
                │ Integration Tests │
                │ Important systems │
                └─────────┬─────────┘
                          │
              ┌───────────▼───────────┐
              │      Unit Tests       │
              │ Majority of coverage │
              └───────────────────────┘
```

## Unit Tests

Fast tests covering isolated business logic.

## Integration Tests

Tests covering interactions between multiple components.

## E2E Tests

Tests covering complete user workflows through the desktop application.

---

# 4. Testing Tools

| Area            | Tool                    | Purpose                                   |
| --------------- | ----------------------- | ----------------------------------------- |
| Go Core         | **Go `testing`**        | Unit and package tests                    |
| Go Concurrency  | **Go race detector**    | Detect data races                         |
| Go Performance  | **Go benchmarks**       | Performance regression testing            |
| Goja            | **Go `testing` + Goja** | JavaScript runtime and scripting behavior |
| Frontend        | **Vitest**              | UI components and frontend logic          |
| Desktop E2E     | **Playwright**          | Critical application workflows            |
| API Integration | **Go test servers**     | Deterministic protocol testing            |
| CI              | **GitHub Actions**      | Automated validation                      |

---

# 5. Coverage Policy

The project does **not** enforce 100% test coverage.

Coverage is used as a quality indicator rather than the primary measure of test quality.

### Coverage Targets

| Area                            |   Target Coverage | Priority    |
| ------------------------------- | ----------------: | ----------- |
| Core HTTP / Request Engine      |        **90–95%** | 🔴 Critical |
| Authentication & Token Handling |        **90–95%** | 🔴 Critical |
| Request / Response Parsing      |          **90%+** | 🔴 Critical |
| Secret Management               |          **90%+** | 🔴 Critical |
| Environment / Variables         |        **85–90%** | 🟠 High     |
| Collections / Workspace Logic   |        **85–90%** | 🟠 High     |
| Goja JavaScript Scripting       |        **80–90%** | 🟠 High     |
| Import / Export                 |        **80–90%** | 🟠 High     |
| Persistence / Local Storage     |        **80–85%** | 🟡 Medium   |
| Wails UI Bindings               |        **70–80%** | 🟡 Medium   |
| UI Components                   |        **60–75%** | 🟢 Lower    |
| Integration / E2E               | **Key workflows** | 🔴 Critical |

---

# 6. Overall Coverage Target

The project-level coverage policy is:

> **Minimum overall coverage: 80%**
> **Target overall coverage: 85%**
> **Critical business logic: 90%+**
> **100% coverage is not required.**

Coverage should be evaluated alongside:

* Test quality
* Edge-case coverage
* Failure handling
* Regression coverage
* Security coverage
* Integration coverage
* E2E coverage

A lower-coverage module may be acceptable when the uncovered code is genuinely low-risk, while critical code should maintain significantly higher coverage.

---

# 7. Coverage vs Confidence

The project should prioritize **confidence over coverage percentage**.

For example, reaching 100% line coverage does not guarantee that a request engine is correctly tested.

A test suite should verify meaningful behavior such as:

```text
Request
  ↓
Authentication
  ↓
Variables
  ↓
Network
  ↓
Response
  ↓
Parsing
  ↓
Assertions
```

rather than simply executing every line.

Coverage should therefore be used to identify potentially untested areas, not as a reason to create meaningless tests.

---

# 8. Unit Test Strategy

Unit tests should make up the majority of the test suite.

They should be:

* Fast
* Deterministic
* Isolated
* Repeatable
* Easy to debug

The primary focus should be business logic rather than UI implementation details.

---

# 9. Core HTTP / Request Engine

**Target: 90–95%**

This is one of the most critical components of the application.

Tests should cover:

* GET
* POST
* PUT
* PATCH
* DELETE
* Headers
* Query parameters
* Path parameters
* Request bodies
* JSON
* Form data
* Multipart
* File uploads
* File downloads
* Redirects
* Timeouts
* Connection failures
* DNS failures
* TLS failures
* Non-2xx responses
* Streaming responses
* Request cancellation
* Authentication
* Variable substitution
* Script execution
* Retry behavior

Tests should also verify correct behavior when multiple features interact.

Example:

```text
Environment
    ↓
Variables
    ↓
Authentication
    ↓
Pre-request Script
    ↓
HTTP Request
    ↓
Response
    ↓
Post-request Script
```

---

# 10. Authentication & Token Handling

**Target: 90–95%**

Authentication is security-critical and should have extensive behavioral coverage.

### Authentication Methods

* Basic
* Bearer
* API Key
* JWT
* Digest
* NTLM
* OAuth 1.0
* OAuth 2.0
* AWS Signature
* Akamai EdgeGrid
* Custom authentication

### OAuth Testing

Test:

* Authorization Code
* Client Credentials
* Password Credentials
* Token acquisition
* Token refresh
* Expired tokens
* Invalid tokens
* Missing credentials
* Invalid configuration
* Browser-based authorization

Sensitive values must never appear in test output.

---

# 11. Secret Management

**Target: 90%+**

Secret management should receive critical-level coverage.

Test:

* Secret storage
* Secret retrieval
* Encryption/decryption behavior
* OS keychain integration
* Secret masking
* Secret interpolation
* Secret deletion
* Invalid credentials
* Access failures
* Secret isolation

Tests must verify that secrets do not leak through:

* Logs
* Errors
* Debug output
* Test output
* Request history
* Generated documentation

---

# 12. Request & Response Parsing

**Target: 90%+**

Test supported response types:

* JSON
* XML
* HTML
* Plain text
* CSV
* Binary
* Images
* Empty responses
* Malformed responses
* Large responses

### JSONPath

Test:

* Objects
* Arrays
* Nested values
* Missing values
* Invalid expressions

### XPath

Test:

* XML elements
* Attributes
* Nested nodes
* Missing nodes
* Invalid expressions

---

# 13. Environment & Variables

**Target: 85–90%**

Variable resolution is fundamental to request execution.

Test:

* Global variables
* Environment variables
* Collection variables
* Folder variables
* Request variables
* Runtime variables
* Prompt variables
* Process environment variables

### Variable Precedence

The project's variable precedence rules must be explicitly defined and tested.

Example:

```text
Global
  ↓
Environment
  ↓
Collection
  ↓
Folder
  ↓
Request
  ↓
Runtime
```

### Edge Cases

* Missing variables
* Empty values
* Nested interpolation
* Multiple variables
* Variable overrides
* Circular references
* Invalid syntax
* Secret variables
* Environment switching

---

# 14. Collections & Workspace Logic

**Target: 85–90%**

Test:

* Workspace creation
* Workspace loading
* Collection creation
* Folder creation
* Request creation
* Request movement
* Rename operations
* Delete operations
* Nested folders
* Request inheritance
* Collection inheritance

Persistence should be verified separately.

---

# 15. Goja JavaScript Scripting

**Target: 80–90%**

Test the scripting API and its integration with the request pipeline.

## Pre-request Scripts

Test:

* Variable creation
* Variable modification
* Request modification
* Authentication preparation
* Dynamic values
* Script failures

## Post-request Scripts

Test:

* Response access
* Response parsing
* Variable extraction
* Response transformation
* Assertions
* Script failures

## Script Context

Test exposed APIs such as:

```text
request
response
variables
environment
console
crypto
utilities
```

## Script Security

Test that scripts cannot access functionality outside their allowed capabilities.

---

# 16. Import & Export

**Target: 80–90%**

Each importer should have its own test suite.

### Import Sources

* Postman
* Insomnia
* OpenAPI
* Swagger
* cURL
* WSDL
* HAR

Test:

* Valid input
* Invalid input
* Missing fields
* Unsupported features
* Authentication
* Variables
* Scripts
* Request bodies
* Nested collections

---

# 17. Persistence & Local Storage

**Target: 80–85%**

Test:

* Project creation
* Project loading
* Project saving
* File serialization
* File parsing
* Invalid project files
* Missing files
* Corrupted files
* Version migration
* File permission failures

## Round-Trip Testing

A project should survive a serialization cycle without losing supported information.

```text
Project
   ↓
Serialize
   ↓
Filesystem
   ↓
Parse
   ↓
Project
```

---

# 18. Git Testing

Test Git operations independently from the UI.

### Areas

* Repository initialization
* Repository detection
* Status
* Changes
* Commit
* Branches
* Diff
* Merge
* Pull
* Push
* Conflict handling

Git failures must never silently cause loss of local API project data.

---

# 19. Integration Testing

Integration tests verify boundaries between major components.

Important combinations include:

```text
Request Engine
      +
Variables
      +
Authentication
      +
Scripts
```

and:

```text
Project Files
      +
Collections
      +
Environments
      +
Git
```

and:

```text
OpenAPI
      +
Request Generation
      +
Contract Testing
      +
Mock Server
```

Integration tests should use deterministic local test servers instead of public APIs.

For **manual / exploratory / demo testing** against a real endpoint, the project provides a companion **reqly-test-api** — a small ElysiaJS mock server with hardcoded data (users CRUD, echo, headers, delay, status, auth), hosted on Vercel. See the `reqly-test-api` repo and the "Mock API" section of the README.

---

# 20. Test Servers

Automated tests should not depend on third-party APIs.

Use local deterministic servers for:

* HTTP
* WebSocket
* SSE
* gRPC
* Mock APIs

Benefits:

* Offline testing
* Faster tests
* Deterministic results
* No rate limits
* No external outages
* Reproducible failures

---

# 21. Wails UI Bindings

**Target: 70–80%**

The Wails bridge should be tested for:

* Method calls
* Arguments
* Return values
* Serialization
* Error propagation
* Cancellation
* Long-running operations
* Event delivery

The bridge should remain thin.

Business logic should stay in the Go core and receive its higher coverage there.

---

# 22. UI Components

**Target: 60–75%**

UI testing should focus on user-visible behavior.

Prioritize:

* Request editor
* Response viewer
* Collection tree
* Environment selector
* Authentication configuration
* Variable editor
* Test interface
* Command palette
* Critical dialogs

Avoid writing tests solely to cover visual implementation details.

---

# 23. Integration & E2E Testing

E2E tests should cover **key workflows only**.

The goal is not to reproduce the entire unit-test suite through the GUI.

### Critical Workflow Example

```text
Launch Application
       ↓
Create Workspace
       ↓
Create Collection
       ↓
Create Request
       ↓
Configure Environment
       ↓
Send Request
       ↓
Inspect Response
       ↓
Save Project
       ↓
Restart Application
       ↓
Verify Project
```

Other important workflows:

* Import Postman collection
* Configure authentication
* Execute collection
* Run tests
* Switch environments
* Execute scripts
* Git workflow
* Generate mock server
* OpenAPI contract validation

---

# 24. Error & Failure Testing

Success cases are not enough.

Every important subsystem should test expected failures.

### Network

* DNS failure
* Connection refused
* Timeout
* TLS failure
* Connection cancellation
* Invalid URL

### Authentication

* Invalid token
* Expired token
* Missing credentials
* Invalid configuration

### Storage

* Permission denied
* Missing files
* Corrupted files
* Disk errors

### Git

* Invalid repository
* Merge conflict
* Authentication failure
* Network failure

### Scripts

* Syntax errors
* Runtime errors
* Timeout
* Invalid API usage

Errors should be predictable, actionable, and must not expose secrets.

---

# 25. Concurrency Testing

Go's concurrent execution model makes concurrency testing important.

Test:

* Parallel requests
* Collection runner
* Concurrent history writes
* WebSocket connections
* SSE streams
* Mock server concurrency
* Request cancellation
* Concurrent project operations

Run the race detector regularly:

```bash
go test -race ./...
```

No known data races should be accepted into the main branch.

---

# 26. Performance Testing

Performance is a product requirement.

The test strategy should therefore include performance regression tests.

### Measure

* Cold startup
* Warm startup
* Memory usage
* Request overhead
* Collection loading
* Large response handling
* JSON parsing
* History search
* Script execution
* Collection runner throughput

### Stress Cases

* 10,000+ requests
* Large collections
* Large JSON responses
* Large request/response bodies
* Long-running streams
* Large history databases
* Large Git repositories

---

# 27. Go Benchmarks

Use Go benchmarks for performance-sensitive code.

Potential benchmark targets:

* Variable interpolation
* JSON parsing
* JSONPath queries
* OpenAPI parsing
* Request construction
* Response processing
* History indexing
* Collection loading
* Goja script execution
* Mock routing

Example:

```bash
go test -bench=. ./...
```

Performance regressions should be investigated rather than automatically accepted.

---

# 28. Regression Testing

Every significant bug fix should include a regression test whenever practical.

```text
Bug
 ↓
Reproduce
 ↓
Write Regression Test
 ↓
Fix
 ↓
Test Passes
 ↓
Keep Test
```

This prevents previously fixed behavior from silently breaking again.

---

# 29. TDD Feature Workflow

Every feature should follow:

```text
1. Define behavior
       ↓
2. Write failing test
       ↓
3. Confirm RED
       ↓
4. Implement minimum solution
       ↓
5. Confirm GREEN
       ↓
6. Refactor
       ↓
7. Add edge cases
       ↓
8. Add integration coverage
       ↓
9. Add E2E coverage if required
       ↓
10. Update documentation
```

---

# 30. Example TDD Workflow

### Feature

> Environment variables override global variables.

### Test

```text
Given:

Global API_URL =
"https://global.example.com"

Environment API_URL =
"https://dev.example.com"

When:

The request resolves {{API_URL}}

Then:

The resolved URL should be:
"https://dev.example.com"
```

### TDD Cycle

```text
Test
 ↓
RED
 ↓
Implement precedence
 ↓
GREEN
 ↓
Refactor
 ↓
Add edge cases
```

Additional tests should cover:

* Missing environment
* Missing variable
* Empty value
* Multiple variables
* Request-level override

---

# 31. Test Naming

Tests should describe behavior.

Prefer:

```text
resolves_environment_variable_before_global_variable
```

instead of:

```text
testVariable2
```

Frontend:

```text
shows_authentication_error_when_token_is_invalid
```

E2E:

```text
creates_request_and_executes_it_against_environment
```

A test name should explain the behavior being guaranteed.

---

# 32. Test Fixtures

Maintain reusable fixtures:

```text
tests/
├── fixtures/
│   ├── openapi/
│   ├── postman/
│   ├── insomnia/
│   ├── har/
│   ├── grpc/
│   ├── soap/
│   ├── requests/
│   ├── responses/
│   └── projects/
│
├── integration/
├── e2e/
└── performance/
```

Fixtures should represent realistic API projects and edge cases.

---

# 33. CI Testing

Every pull request should run automated validation.

```text
Pull Request
     │
     ├── Formatting
     ├── Static Analysis
     ├── Unit Tests
     ├── Race Tests
     ├── Frontend Tests
     ├── Integration Tests
     ├── E2E Tests
     └── Build
             │
             ▼
           PASS
```

---

# 34. CI Test Levels

## Fast Checks

Run on every change:

* Formatting
* Linting
* Type checking
* Unit tests

## Pull Request CI

Run on every pull request:

* Unit tests
* Integration tests
* Race detection
* Frontend tests
* Build validation
* Coverage validation

## Release CI

Run before releases:

* Full E2E suite
* Performance tests
* Security tests
* Cross-platform builds
* Import/export compatibility tests
* Large-data tests
* Installation tests
* Upgrade tests

---

# 35. Coverage Enforcement

Coverage thresholds should be enforced at the appropriate level.

### Project Level

```text
Minimum: 80%
Target: 85%
```

### Critical Packages

Critical packages should maintain **90%+** coverage.

### Important Principle

Coverage should not be artificially increased with meaningless tests.

A pull request should not be considered higher quality simply because it increases the percentage.

Coverage decreases may be acceptable when justified, particularly when removing dead code or restructuring functionality, but critical business logic should remain within its defined target.

---

# 36. Definition of Done

A feature is complete when:

* [ ] Requirement is clearly defined
* [ ] TDD cycle has been followed
* [ ] Unit tests are implemented
* [ ] Edge cases are covered
* [ ] Error behavior is tested
* [ ] Integration tests are added where required
* [ ] E2E coverage exists for critical workflows
* [ ] Security implications are tested
* [ ] Performance impact is considered
* [ ] Regression tests are added where applicable
* [ ] Coverage remains within project targets
* [ ] Documentation is updated
* [ ] CI passes
* [ ] No known race conditions exist
* [ ] No secrets are exposed by tests or logs

---

# 37. Testing Principles

The project will follow these principles:

1. **TDD is the default development workflow.**
2. **Tests should describe behavior, not implementation details.**
3. **Critical business logic receives the highest coverage targets.**
4. **80% is the minimum overall coverage threshold.**
5. **85% is the project-wide target.**
6. **Critical business logic should maintain 90%+ coverage.**
7. **100% coverage is explicitly not a project goal.**
8. **Edge cases and failure paths are as important as happy paths.**
9. **Security-sensitive behavior must receive strong test coverage.**
10. **Unit tests should form the majority of the test suite.**
11. **Integration tests should validate important component boundaries.**
12. **E2E tests should cover critical user workflows only.**
13. **Automated tests should be deterministic and reproducible.**
14. **Public APIs should not be required for automated tests.**
15. **Performance regressions should be measured and investigated.**
16. **Concurrency should be validated with race detection.**
17. **Every significant bug should result in a regression test where practical.**
18. **Coverage must never be increased through meaningless tests.**
19. **Tests should run locally and in CI.**
20. **A green test suite should provide meaningful confidence in the product.**

---

# 38. Final Testing Strategy

The testing strategy for the **Go + Wails + Goja API Client** can be summarized as:

```text
                         TDD
                          │
                 ┌────────┴────────┐
                 │                 │
              Behavior          Requirements
                 │                 │
                 └────────┬────────┘
                          ▼
                    Unit Tests
                     80–90%
                          │
                          ▼
                 Integration Tests
                          │
                          ▼
                    E2E Tests
                Critical Workflows
                          │
              ┌───────────┴───────────┐
              │                       │
        Performance                Security
        Benchmarks                 Testing
              │                       │
              └───────────┬───────────┘
                          ▼
                         CI
                          │
                          ▼
                       Release
```

The project-wide goal is **~85% coverage with substantially higher coverage for critical business logic**, rather than chasing an arbitrary 100% number.

The guiding principle is:

> **Test behavior, protect critical paths, cover failure modes, and use coverage to find gaps rather than to manufacture confidence.**
