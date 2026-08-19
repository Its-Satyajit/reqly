// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/runner"
)

// RunTestResult is one reqly.test() assertion result in a run.
type RunTestResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

// RunStep is the bridge-friendly outcome of one request in a collection run.
// Credential values are masked before serialization; the raw runner step
// (with its error interface and untagged response) is never exposed.
type RunStep struct {
	Name string `json:"name"`
	// RequestPath is the workspace-relative Request Path of the step.
	RequestPath string `json:"requestPath"`
	// Passed is true when the request succeeded and every test passed.
	Passed bool `json:"passed"`
	// RequestError is the transport/pre-script error text ("" when nil).
	RequestError string `json:"requestError,omitempty"`
	// Response is the received response (nil on failure).
	Response *SendResponse `json:"response,omitempty"`
	// Tests are the results of reqly.test() assertions.
	Tests []RunTestResult `json:"tests"`
	// Logs are console output from pre/post scripts.
	Logs []string `json:"logs"`
}

// RunReport is the aggregate result of a collection run.
type RunReport struct {
	Steps      []RunStep `json:"steps"`
	Started    time.Time `json:"started"`
	Finished   time.Time `json:"finished"`
	Total      int       `json:"total"`
	Passed     int       `json:"passed"`
	Failed     int       `json:"failed"`
	DurationMS int64     `json:"durationMs"`
	OK         bool      `json:"ok"`
}

// RunOptions configures a collection run.
type RunOptions struct {
	// Env is the environment name applied to the whole run; when empty, the
	// workspace descriptor's environment: field is used.
	Env string
	// FailFast stops the run after the first failing step.
	FailFast bool
	// OnStep, if non-nil, is invoked once per completed step, in execution
	// order, before Run returns.
	OnStep func(RunStep)
}

// CollectionRunService executes collections (or individual folders) on disk
// through the shared runner and returns bridge-safe reports. Runs read fresh
// from disk, apply one environment for the whole run, and are single-flight.
type CollectionRunService struct {
	root string

	mu      sync.Mutex
	running bool
}

// NewCollectionRunService returns a service rooted at the given workspace
// root ("" means no workspace; Run then errors).
func NewCollectionRunService(root string) *CollectionRunService {
	return &CollectionRunService{root: root}
}

// Run executes the collection or folder at path (a workspace-relative
// container path such as "users" or "users/auth"). Every request is run in
// deterministic order through the shared runner, streaming each completed
// step through opts.OnStep. Only one run may be in flight at a time.
func (s *CollectionRunService) Run(ctx context.Context, path string, opts RunOptions) (*RunReport, error) {
	if s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to run collections")
	}
	if !s.acquire() {
		return nil, fmt.Errorf("a collection run is already in progress")
	}
	defer s.release()

	ws, err := collections.LoadWorkspace(s.root)
	if err != nil {
		return nil, err
	}

	// Runs use one environment for the whole run and ignore per-file
	// environment: fields, mirroring the CLI.
	envSet, masker, err := environments.ResolveSet(s.root, environments.Selection{
		EnvFlag:   opts.Env,
		ConfigEnv: collections.WorkspaceEnvironment(s.root),
	})
	if err != nil {
		return nil, err
	}

	coll, folder, err := findContainer(ws, path)
	if err != nil {
		return nil, err
	}

	steps := make([]RunStep, 0)
	runOpts := runner.Options{
		FailFast: opts.FailFast,
		OnStep: func(step runner.StepResult) {
			masker.Add(step.AuthValues()...)
			dto := runStepDTO(s.root, step, masker)
			steps = append(steps, dto)
			if opts.OnStep != nil {
				opts.OnStep(dto)
			}
		},
	}

	var report *runner.Report
	if folder != nil {
		report, err = runner.RunFolder(ctx, ws, coll, folder, envSet, runOpts)
	} else {
		report, err = runner.RunCollection(ctx, ws, coll, envSet, runOpts)
	}
	if err != nil {
		return nil, err
	}
	return runReportDTO(report, steps), nil
}

// acquire is a no-op single-flight guard: only one run at a time.
func (s *CollectionRunService) acquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	return true
}

func (s *CollectionRunService) release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

// findContainer resolves a workspace-relative container path ("users" or
// "users/auth") to its collection and, for a folder target, the folder. A
// folder's owning collection is returned alongside so inherited configuration
// resolves correctly.
func findContainer(ws *collections.Workspace, path string) (*collections.Collection, *collections.Folder, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return nil, nil, fmt.Errorf("collection or folder %q not found", path)
	}

	var coll *collections.Collection
	for _, c := range ws.Collections {
		if c.Name == parts[0] {
			coll = c
			break
		}
	}
	if coll == nil {
		return nil, nil, fmt.Errorf("collection %q not found in the workspace", path)
	}
	if len(parts) == 1 {
		return coll, nil, nil
	}

	folder, ok := findFolder(coll, parts[1:])
	if !ok {
		return nil, nil, fmt.Errorf("folder %q not found in collection %q", path, coll.Name)
	}
	return coll, folder, nil
}

// findFolder walks the folder segments below a collection, returning the
// terminal folder.
func findFolder(coll *collections.Collection, parts []string) (*collections.Folder, bool) {
	var current []*collections.Folder = coll.Folders
	var found *collections.Folder
	for i, part := range parts {
		var next *collections.Folder
		for _, f := range current {
			if f.Name == part {
				next = f
				break
			}
		}
		if next == nil {
			return nil, false
		}
		found = next
		if i < len(parts)-1 {
			current = next.Folders
		}
	}
	return found, true
}

// runStepDTO maps a raw runner step to its bridge-safe form, masking
// credential values in the error, logs, response body, and header values. The
// Request Path is converted to its workspace-relative form.
func runStepDTO(root string, step runner.StepResult, masker *environments.Masker) RunStep {
	dto := RunStep{
		Name:        step.Name,
		RequestPath: containerPath(root, step.RequestPath),
		Passed:      step.Passed,
		Tests:       make([]RunTestResult, 0, len(step.Tests)),
		Logs:        make([]string, 0, len(step.Logs)),
	}
	if step.RequestError != nil {
		dto.RequestError = masker.Mask(step.RequestError.Error())
	}
	if step.Response != nil {
		dto.Response = &SendResponse{
			StatusCode: step.Response.StatusCode,
			StatusText: step.Response.StatusText,
			Proto:      step.Response.Proto,
			Headers:    maskHeaders(step.Response.Headers, masker),
			Body:       masker.Mask(step.Response.Text()),
			DurationMS: step.Response.Duration.Milliseconds(),
			Size:       step.Response.Size,
			OK:         step.Response.OK(),
		}
	}
	for _, t := range step.Tests {
		dto.Tests = append(dto.Tests, RunTestResult{Name: t.Name, Passed: t.Passed})
	}
	for _, l := range step.Logs {
		dto.Logs = append(dto.Logs, masker.Mask(l))
	}
	return dto
}

// maskHeaders returns the response headers with credential values masked in
// each header value.
func maskHeaders(headers map[string][]string, masker *environments.Masker) map[string][]string {
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		masked := make([]string, len(values))
		for i, v := range values {
			masked[i] = masker.Mask(v)
		}
		out[key] = masked
	}
	return out
}

// runReportDTO maps a runner report and its streamed bridge-safe steps to the
// aggregate form.
func runReportDTO(report *runner.Report, steps []RunStep) *RunReport {
	return &RunReport{
		Steps:      steps,
		Started:    report.Started,
		Finished:   report.Finished,
		Total:      report.Total,
		Passed:     report.Passed,
		Failed:     report.Failed,
		DurationMS: report.Duration.Milliseconds(),
		OK:         report.OK(),
	}
}
