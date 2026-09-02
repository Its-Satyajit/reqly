# ADR 0034: Self-Hosted Automation — Local Workflow Scheduler (M66)

## Status
Accepted — grill Q1 (core + CLI + desktop, UI pending)

## Context
M65 shipped the workflow engine; `Milestones/04` lists `Self-hosted automation` as the last P2 gap. Users want to run workflows periodically without a cloud (health-check workflows, token-refresh chains, nightly contract tests). The existing `monitor.Run` already does interval scheduling for single requests, but automation needs to schedule an entire `Workflow` (multi-step, conditional, variable extraction) with Git-native config, local-only execution, and zero telemetry.

## Decision
Ship `internal/automation` (M66) as a thin scheduler over `internal/workflow`.

* **Core:** `internal/automation/automation.go` — `Automation {Name, Workflow workflow.Workflow, Interval string, Enabled *bool, MaxRuns int}` (nil Enabled = true, Interval Go duration string), `IsEnabled()`, `IntervalDuration() time.Duration`, `Validate()`, and `Scheduler {client *request.Client}` with `NewScheduler` and `Run(ctx, *Automation, func(*workflow.WorkflowReport)) error`. `Run` validates, checks `IsEnabled`, parses interval, creates `workflow.NewWorkflowExecutor`, runs once immediately, then ticks at interval (or once if 0), counting `MaxRuns` and selecting on `ctx.Done()`.
* **CLI:** `apps/cli/cmd/automation.go` — `reqly automation run <automation.yaml> [--once] [--interval <dur>] [--max-runs N]` loads YAML, applies flag overrides, validates, then `Scheduler.Run(context.Background(), &auto, onReport)` where `onReport` prints `[#] Automation "name" workflow "wf": PASSED…` plus per-step lines. Captures `context.Canceled` as success (Ctrl+C).
* **Desktop:** `apps/desktop/backend/automation.go` — `AppService.AutomationRun(yamlStr string) (*workflow.WorkflowReport, error)` unmarshals, forces `Interval="0"` to run once (desktop does not run background ticks; scheduling UI will add interval controls later), validates, then `Scheduler.Run` with single-report capture. Exposed via Wails.
* **Tests:** `internal/automation/automation_test.go` (Validate, IsEnabled, RunOnce, RunInterval 10ms×3, Disabled, Nil) + `apps/cli/cmd/automation_test.go` (once run) + `apps/desktop/backend/automation_test.go` (desktop once) — all `httptest`, no network.

## Consequences
Q1: Interval is Go duration string (`10s`, `5m`), not cron — cron syntax deferred to M66b (keeps parsing trivial, `time.ParseDuration`).
Q2: Desktop runs once only — background scheduling (ticker in a Wails goroutine + stop button) is UI follow-up, not required for core/CLI gate.
Q3: No persistence of automation reports — reports are printed/streamed, not written to `history.db`; history stays per-request.
Q4: `Workflow` is embedded in `Automation`, not a file path — avoids extra I/O and keeps the automation file self-contained and Git-native; a `workflowFile` indirection is M66b.
