# 08: Tool Pages

**What to build:** All standalone tool pages in the Tool Rail — Diff, JWT Inspector, GraphQL, gRPC, Runners, OpenAPI Explorer. Each page is a complete feature with its own layout and interactions.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

### Diff Page
- [ ] Two dropdowns (A and B) with same source list: Current Request, History Items, Saved Requests, Clipboard
- [ ] Diff modes: Side by Side, Unified, Structural JSON, Headers
- [ ] Diff output shows additions/deletions with highlighting

### JWT Inspector
- [ ] Token input area (textarea)
- [ ] "Use token from request" button pulls token from current request's auth config
- [ ] Decode button parses JWT
- [ ] Decoded output: Header (JSON), Payload (JSON), Signature (Present/Absent)
- [ ] Token status: Valid/Expired, Expires in X minutes

### GraphQL Page
- [ ] Endpoint input with Connect button
- [ ] Schema sidebar shows: Query, Mutation, Subscription, Types, Enums, Interfaces, Input Types, Scalars, Directives
- [ ] Automatic introspection on first connect (if server supports it)
- [ ] Fallback to local .graphql/.gql schema file
- [ ] Query editor with syntax highlighting
- [ ] Variables editor
- [ ] Response viewer
- [ ] Run button executes query

### gRPC Page
- [ ] Server address input with Connect button
- [ ] Services tree shows: Services → Methods
- [ ] Server reflection as primary schema source
- [ ] Fallback to local .proto files
- [ ] Method details panel: Request type, Response type, Metadata
- [ ] Request editor (JSON)
- [ ] Response viewer
- [ ] Invoke button executes method

### Runners Page
- [ ] Tabs: Collection, Pagination, Bulk, Dataset
- [ ] Collection Runner: Select collection, environment, delay, concurrency
- [ ] Pagination Runner: Select request, strategy, parameter, start, max pages
- [ ] Bulk Runner: Select source, dataset, environment, concurrency, delay, retries
- [ ] Dataset Runner: Select dataset file or inline, iterations, concurrency
- [ ] Each tab shows configuration and results table
- [ ] Run button executes the runner
- [ ] Progress indicator during execution

### OpenAPI Explorer
- [ ] API tree sidebar shows endpoints grouped by tags
- [ ] Endpoint details panel: summary, parameters, request schema, response schema, security
- [ ] One-spec dropdown to switch between imported specs
- [ ] "Open in Request Builder" button sends endpoint to Request Builder
- [ ] Import Spec button opens import dialog

- [ ] All components use theme tokens
