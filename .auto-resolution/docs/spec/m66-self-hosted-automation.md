# Spec M66: Self-Hosted Automation — Local Workflow Scheduler

## Summary
Run a `Workflow` on a local schedule without any cloud. Git-native YAML, interval-based, context-cancellable, shared with desktop via `AppService.AutomationRun`.

## Motivation
- Schedule nightly contract tests, hourly health workflows, or token-refresh chains locally.
- Keep automation config in Git alongside workflows and request files.
- Reuse the workflow engine's variable extraction and condition semantics.

## Scope
- **In:** `Automation` YAML schema, `Scheduler.Run` ticker, `IsEnabled`/`IntervalDuration`/`Validate`, `internal/automation` + CLI `reqly automation run <file> [--once --interval --max-runs]` + desktop `AutomationRun` binding, 6 core tests + CLI + desktop tests.
- **Out:** Cron syntax (`* * * * *`), Git-ops trigger (push/PR), persistence of automation reports to history, desktop scheduling UI (interval picker + start/stop), `workflowFile` indirection (M66b).

## Data Model
```yaml
name: nightly-health
workflow:
  name: health
  steps:
    - id: s1
      name: S1
      request: {method: GET, url: https://api.example.com/health}
interval: 10s      # Go duration; "0" or "" → run once
enabled: true      # nil → true
maxRuns: 3         # 0 → infinite until ctx cancel
```

```go
type Automation struct {
    Name     string             `yaml:"name"`
    Workflow workflow.Workflow `yaml:"workflow"`
    Interval string             `yaml:"interval,omitempty"`
    Enabled  *bool              `yaml:"enabled,omitempty"`
    MaxRuns  int                `yaml:"maxRuns,omitempty"`
}
func (a *Automation) IsEnabled() bool
func (a *Automation) IntervalDuration() (time.Duration, error)
func (a *Automation) Validate() error

type Scheduler struct { client *request.Client }
func NewScheduler(client *request.Client) *Scheduler
func (s *Scheduler) Run(ctx context.Context, a *Automation, onReport func(*workflow.WorkflowReport)) error
```

## Execution Semantics
1. Validate (`name` + `workflow.steps` required, `interval` parseable, `maxRuns >=0`).
2. If `!IsEnabled()` → error.
3. `interval, _ := IntervalDuration()`
4. `exec := workflow.NewWorkflowExecutor(s.client)`
5. `runOnce := func() { report, _ := exec.Execute(ctx, &a.Workflow, WorkflowOptions{}); onReport(report) }`
6. If `interval == 0` → `runOnce()` once and return.
7. Else `runOnce()` immediately, then `ticker := time.NewTicker(interval)`; loop `select { case <-ctx.Done(): return ctx.Err(); case <-ticker.C: runOnce(); count++; if MaxRuns>0 && count>=MaxRuns { return nil } }`

## CLI
```
reqly automation run <automation.yaml> [--once] [--interval 30s] [--max-runs 5]
```
`--once` forces `interval="0"`; `--interval`/`--max-runs` override file. Prints per-run header and steps; `context.Canceled` is not an error (Ctrl+C).

## Desktop
```
AutomationRun(yamlStr string) (*workflow.WorkflowReport, error)
```
Forces `interval="0"` → single run, captures report via `onReport` closure.

## Testing
- `TestAutomation_Validate` (4 cases), `TestAutomation_IsEnabled`, `TestScheduler_RunOnce`, `TestScheduler_RunInterval` (10ms×3, timing assertion), `TestScheduler_Disabled`, `TestScheduler_NilAutomation` — `httptest`.
- `TestAutomationCLIOnce` — temp YAML + `rootCmd.Execute(["automation","run",file,"--once"])` → output contains `test-auto` + `PASSED`.
- `TestAutomationRun_Desktop` — `AppService.AutomationRun` → `Passed` + 1 step.

## Acceptance Criteria
- [x] `Validate` rejects missing name/steps/bad interval.
- [x] `Run` with interval 0 runs once and calls `onReport` once.
- [x] `Run` with interval 10ms + MaxRuns 3 runs 3 times and takes ≥15ms.
- [x] Disabled automation errors.
- [x] CLI `--once` runs once and prints `PASSED`.
- [x] Desktop `AutomationRun` runs once and returns passed report.
