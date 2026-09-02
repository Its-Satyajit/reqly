# ADR 0033: Workflow Engine — Sequential Execution with Variable Extraction and Conditional Steps (M65)

## Status
Accepted — grill Q1 (core + CLI + Goja, UI pending)

## Context
`ROADMAP.md` P2 §57 and `Milestones/04-phase-3-power-user-features.md` list *Visual workflow builder* and *Self-hosted automation* as the remaining P2 gaps. Users need to chain requests (login → extract token → profile) with variable propagation and conditional branching without a cloud runner. The `collection run` runner already does sequential execution with `reqly.test()` but has no declarative YAML format, no per-step `condition`, and no JSON-key `extract`. The visual builder will need a pure-Go execution seam that desktop, CLI, and scripting can share, with the same `variables` precedence and `request.Client` transport as single sends.

## Decision
Ship `internal/workflow` (M65) as the workflow execution seam, with dual parity: Go core + CLI + Goja.

* **Core:** `internal/workflow/workflow.go` — `Workflow {Name, Description, Variables, Steps}`, `WorkflowStep {ID, Name, Request, Condition, Extract}`, `WorkflowReport {WorkflowName, Passed, Duration, Steps: StepResult[], ExtractedVars}`, `StepResult {Name, RequestPath, Passed, RequestError string, Response, Logs}`, `WorkflowOptions {EnvironmentVars}`, `WorkflowExecutor` with `Execute(ctx, *Workflow, WorkflowOptions)`. Variable store is `variables.NewSet()` seeded from `Workflow.variables` (`ScopeGlobal`) and `opts.EnvironmentVars` (`ScopeEnvironment`); `extract` writes back to `ScopeGlobal`. Condition evaluated via one-shot `goja.New()` with `reqly.getVariable` bound to `varStore.Resolve` (full precedence). Request interpolation clones `Headers` and `Query` slices, then `Interpolate` for URL/headers/query/body (ignores interpolate errors — leaves raw). Execution via `request.Client.Execute(ctx, &reqCopy, varStore)` so auth, retry, proxy, TLS, and cookie jar apply. Extract reads `resp.Body` as `map[string]any` (top-level keys only) and stringifies with `fmt.Sprintf("%v")`.
* **CLI:** `apps/cli/cmd/workflow.go` — `reqly workflow <workflow.yaml>` reads YAML (or JSON, as YAML superset) via `gopkg.in/yaml.v3`, `NewWorkflowExecutor(nil)`, `Execute(context.Background(), &wf, WorkflowOptions{})`, prints `Workflow: NAME (duration, N steps)` plus per-step `[PASSED/FAILED] name: id` and `Extracted Variables:` block, returns error when `!report.Passed` so CI fails.
* **Scripting:** `internal/scripting/sandbox.go` — `reqly.workflow.run(yamlString)` bound in `bindReqly()`: YAML-unmarshals the string, executes with background context, returns a Goja object `{workflowName, passed, duration (ms), extractedVars}` or `undefined` on parse/execute error. `internal/scripting/workflow_script_test.go` covers Goja round-trip.
* **Tests:** `internal/workflow/workflow_test.go` — table-driven `httptest` covering auth-flow chaining, `ConditionSkip`, `ExtractMissingKey`, `QueryInterpolation`, `NilWorkflow`, `FailureStatus` (500 → failed), `EnvironmentVars`. CLI `TestWorkflowCLICommand` and scripting `TestScripting_Workflow` cover entry points. No history pollution — workflow reports are episodic; history recording stays in the per-request engine path.
* **UI:** Deferred. The visual builder (drag-drop DAG, JSONPath extract, full variable scopes) is tracked as P2 UI follow-up — core shipped / UI pending, consistent with `ROADMAP.md` precedence (core shipped ≠ UI shipped).

## Consequences
Q1: Single-flight sequential execution only; DAG parallelism, `while` loops, and nested workflows are deferred to M65b (keeps the engine testable).
Q2: `Extract` is top-level keys only (`parsed[jsonKey]`); JSONPath/nested dotted paths (`$.data.token`) are M65b — current callers use flat token responses (login → `{"token":"…"}`).
Q3: `Condition` is a pure boolean JS expression with only `reqly.getVariable`; access to `request`/`response` is deferred (keeps the one-shot VM cheap, no state leaking between steps).
Q4: No persistence: workflows are plain YAML files, Git-native like request files; self-hosted automation (cron, webhook) reuses the same file format with a scheduler shim in M65b.
