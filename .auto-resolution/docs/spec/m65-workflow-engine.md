# Spec M65: Workflow Engine — Sequential Multi-Step Execution

## Summary
Provide a Git-native, sequential workflow engine for chaining API requests with variable extraction and conditional steps, exposed via Go core (`internal/workflow`), CLI (`reqly workflow`), and Goja scripting (`reqly.workflow.run`). Visual workflow builder UI is deferred — M65 ships core + CLI + scripting (core shipped / UI pending).

## Motivation
- Chain login → profile flows without leaving Reqly: extract `token` from one response and interpolate `{{token}}` in the next request.
- Declarative YAML format versionable in Git, like request files.
- Single-flight, deterministic execution that reuses the existing `request.Client` pipeline (auth, variables, cookies, proxy, TLS) and `variables` precedence.
- Unblock the visual builder (DAG editor) with a pure-Go seam that desktop and CLI share.

## Scope
- **In:** `Workflow` YAML schema, `WorkflowExecutor` sequential loop, `condition` (Goja boolean with `reqly.getVariable`), `extract` (top-level JSON key → variable), URL/headers/query/body interpolation, `WorkflowReport`/`StepResult`, CLI `reqly workflow <file>` (YAML or JSON), Goja `reqly.workflow.run(yamlString)` → `{workflowName, passed, duration, extractedVars}`, 7 unit tests + CLI + scripting tests, `ROADMAP.md`/`Milestones/04` ledger updates.
- **Out:** Visual DAG editor, parallel/DAG execution, `while`/loop constructs, nested workflows, JSONPath/dotted-key extract (`$.data.token`), cron/scheduler, webhook triggers, persistence/history of workflow runs (deferred to M65b).

## Data Model
```yaml
name: Authentication Flow
description: optional
variables:
  baseUrl: https://api.example.com
steps:
  - id: login
    name: Login Step
    request:
      method: POST
      url: "{{baseUrl}}/login"
      body: '{"user":"alice"}'
      headers: [{key: Content-Type, value: application/json}]
      query: [{key: version, value: "1"}]
    condition: "reqly.getVariable('token') !== ''" # optional Goja JS boolean
    extract:
      token: token  # response JSON top-level key → varName
  - id: profile
    name: Profile Step
    request:
      method: GET
      url: "{{baseUrl}}/profile"
      headers: [{key: Authorization, value: "Bearer {{token}}"}]
```

```go
type Workflow struct {
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description,omitempty" yaml:"description,omitempty"`
    Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
    Steps       []WorkflowStep    `json:"steps" yaml:"steps"`
}
type WorkflowStep struct {
    ID        string            `json:"id" yaml:"id"`
    Name      string            `json:"name" yaml:"name"`
    Request   request.Request   `json:"request" yaml:"request"`
    Condition string            `json:"condition,omitempty" yaml:"condition,omitempty"`
    Extract   map[string]string `json:"extract,omitempty" yaml:"extract,omitempty"`
}
type WorkflowReport struct {
    WorkflowName  string            `json:"workflowName"`
    Passed        bool              `json:"passed"`
    Duration      time.Duration     `json:"duration"`
    Steps         []StepResult      `json:"steps"`
    ExtractedVars map[string]string `json:"extractedVars"`
}
type StepResult struct {
    Name         string             `json:"name"`
    RequestPath  string             `json:"requestPath"`
    Passed       bool               `json:"passed"`
    RequestError string             `json:"requestError,omitempty"`
    Response     *response.Response `json:"response,omitempty"`
    Logs         []string           `json:"logs"`
}
type WorkflowOptions struct {
    EnvironmentVars map[string]string
}
```

## Execution Semantics
1. `varStore := variables.NewSet()` seeded from `Workflow.variables` (`ScopeGlobal`) then `opts.EnvironmentVars` (`ScopeEnvironment`).
2. For each step in order:
   a. If `condition != ""`, create fresh `goja.New()`, bind `reqly.getVariable(name)` → `varStore.Resolve(name)`, `RunString(condition)`. Skip (`continue`) only when `err == nil && !val.ToBoolean()`. Errors or true → run.
   b. Clone `Headers` and `Query` slices, then `Interpolate` URL, Body, each header value and query value (ignore interpolation errors).
   c. `resp, err := client.Execute(ctx, &reqCopy, varStore)` — reuses full engine (auth, cookies, proxy, TLS, retry).
   d. `passed := err == nil && resp != nil && resp.StatusCode < 400`; overall `report.Passed` fails on any such step.
   e. If `resp.Body` is valid JSON object and `Extract` non-empty, for each `varName: jsonKey` when `parsed[jsonKey]` exists, `varStore.Set(ScopeGlobal, varName, fmt.Sprintf("%v", val))` and `report.ExtractedVars[varName]=str`.
   f. Append `StepResult` with log line `METHOD URL -> status (ms)` or `-> error: msg`.
3. `report.Duration = time.Since(start)`, return.

## CLI
```
reqly workflow <workflow.yaml>  # YAML; JSON accepted as YAML superset
```
Reads file, `yaml.Unmarshal`, `NewWorkflowExecutor(nil).Execute(context.Background(), &wf, WorkflowOptions{})`, prints to stdout:
```
Workflow: NAME (duration, N steps)
  [PASSED] Login Step: login
  [FAILED] Profile Step: profile
Extracted Variables:
  - token = secret123
```
Exit non-zero when `!report.Passed` (for CI). Flags: none for M65 (`--env` deferred; use `Workflow.variables` + `EnvironmentVars` seam).

## Scripting
```js
const yaml = 'name: Script Flow\nsteps:\n  - id: s1\n    name: S1\n    request:\n      method: GET\n      url: https://api.example.com\n';
const report = reqly.workflow.run(yaml);
// report.workflowName, report.passed (bool), report.duration (ms), report.extractedVars (object)
if (report && report.passed) reqly.setVariable("done", "true");
```
Returns `undefined` on YAML parse error or nil report. Synchronous execution on `context.Background()`.

## Testing Strategy
- **Core:** `TestExecuteWorkflow` (auth chaining), `TestExecuteWorkflow_ConditionSkip` (false → skipped), `TestExecuteWorkflow_ExtractMissingKey` (no var), `TestExecuteWorkflow_QueryInterpolation` ({{token}} in query), `TestExecuteWorkflow_NilWorkflow` (error), `TestExecuteWorkflow_FailureStatus` (500 → failed), `TestExecuteWorkflow_EnvironmentVars` (opts injection). All use `httptest.NewServer`.
- **CLI:** `TestWorkflowCLICommand` — writes temp YAML, `rootCmd.SetArgs(["workflow", file])`, asserts stdout contains `Workflow: Test Flow` and `token = xyz`.
- **Scripting:** `TestScripting_Workflow` — Goja round-trip `reqly.workflow.run` → `reqly.setVariable("passed","true")` via `NewSandbox`.
- **Gates:** `go test ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l`, `go build -o reqly ./apps/cli`, `nub run typecheck`, `nub run lint`.

## Desktop
Deferred. The engine already satisfies dual parity (core + CLI + scripting); desktop will add `RunWorkflow` binding (like `runners.go`) and a visual DAG editor (drag-drop steps, condition editor, JSONPath extract, live variable inspector) in M65b — tracked as `core shipped / UI pending`.

## Risks
- Top-level extract breaks on nested tokens (`{"data":{"token":"x"}}`) — M65 warns in ADR, callers flatten to `{"token":…}`.
- Condition has no request/response access — mitigated by keeping it simple and documenting extension point.

## Acceptance Criteria
- [x] `internal/workflow` executes the auth-flow fixture end-to-end (login extract → profile with `{{token}}` interpolation) with `Passed` true and `ExtractedVars[token]=secret123`.
- [x] Skipped steps (condition false) do not appear in `report.Steps`.
- [x] Query interpolation (`{{token}}` in `query.value`) is sent as query param.
- [x] `nil` workflow returns error.
- [x] HTTP 500 step marks report `Passed=false` and `StepResult.Passed=false`.
- [x] `EnvironmentVars` injected via `WorkflowOptions` interpolates in headers.
- [x] CLI prints workflow name, per-step PASS/FAIL, and extracted vars, and exits non-zero on failure.
- [x] Goja `reqly.workflow.run` returns object with `passed` and `extractedVars` consumable by scripts.
