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

// Package runner executes collections and request chains. Requests run in
// order with a shared variable store, so a post-request script can extract a
// token and a later request can interpolate it. Pre/post scripts and
// reqly.test() assertions are evaluated through the scripting sandbox.
package runner

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/scripting"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// TestResult is the outcome of one reqly.test() assertion.
type TestResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

// StepResult records the outcome of running one request.
type StepResult struct {
	// Name is the request file name.
	Name string `json:"name"`
	// Request is the request file path.
	RequestPath string `json:"requestPath"`
	// Passed is true when the request succeeded and every test passed.
	Passed bool `json:"passed"`
	// RequestError is the transport/pre-script error, if any.
	RequestError error `json:"requestError,omitempty"`
	// Response is the received response (nil on failure).
	Response *response.Response `json:"response,omitempty"`
	// Tests are the results of reqly.test() assertions.
	Tests []TestResult `json:"tests"`
	// Logs are console output from pre/post scripts.
	Logs []string `json:"logs"`
}

// Report is the aggregate result of a run.
type Report struct {
	// Steps in execution order.
	Steps []StepResult `json:"steps"`
	// Started/Finished bound the run.
	Started  time.Time     `json:"started"`
	Finished time.Time     `json:"finished"`
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Duration time.Duration `json:"duration"`
}

// OK reports whether every step passed.
func (r *Report) OK() bool { return r.Failed == 0 }

// Options configures a run.
type Options struct {
	// FailFast stops the run after the first failing step.
	FailFast bool
	// Client executes requests (nil uses request.NewClient()).
	Client *request.Client
	// ClientOptions are applied when Client is nil.
	ClientOptions []request.Option
}

// RunCollection runs every request in the collection (and nested folders) in
// deterministic order. Vars are the starting variable store shared across all
// steps.
func RunCollection(ctx context.Context, ws *collections.Workspace, coll *collections.Collection, vars *variables.Set, opts Options) (*Report, error) {
	if vars == nil {
		vars = ws.VariablesSet()
	}
	r := &Runner{vars: vars, opts: opts}
	if opts.Client != nil {
		r.client = opts.Client
	} else {
		r.client = request.NewClient(opts.ClientOptions...)
	}

	steps := []step{}
	collectSteps(&steps, coll.Requests, coll.Folders, nil)
	report := &Report{Started: time.Now(), Total: len(steps)}
	for _, s := range steps {
		result := r.runStep(ctx, ws, coll, s.chain, s.entry)
		report.Steps = append(report.Steps, result)
		if result.Passed {
			report.Passed++
		} else {
			report.Failed++
			if opts.FailFast {
				break
			}
		}
	}
	report.Finished = time.Now()
	report.Duration = report.Finished.Sub(report.Started)
	return report, nil
}

// step pairs a request entry with its folder ancestor chain.
type step struct {
	chain []*collections.Folder
	entry *collections.RequestEntry
}

// collectSteps gathers every request in a container (collection or folder) and
// its subfolders in deterministic (name-sorted) order.
func collectSteps(steps *[]step, requests []*collections.RequestEntry, folders []*collections.Folder, chain []*collections.Folder) {
	sortedRequests := append([]*collections.RequestEntry{}, requests...)
	sort.Slice(sortedRequests, func(i, j int) bool { return sortedRequests[i].Name < sortedRequests[j].Name })
	for _, entry := range sortedRequests {
		*steps = append(*steps, step{chain: chain, entry: entry})
	}

	sortedFolders := append([]*collections.Folder{}, folders...)
	sort.Slice(sortedFolders, func(i, j int) bool { return sortedFolders[i].Name < sortedFolders[j].Name })
	for _, folder := range sortedFolders {
		collectSteps(steps, folder.Requests, folder.Folders, append(chain, folder))
	}
}

// Runner executes steps against a shared variable store.
type Runner struct {
	vars   *variables.Set
	client *request.Client
	opts   Options
}

// runStep executes one request end to end: pre-script, transport, post-script.
func (r *Runner) runStep(ctx context.Context, ws *collections.Workspace, coll *collections.Collection, chain []*collections.Folder, entry *collections.RequestEntry) StepResult {
	result := StepResult{
		Name:        entry.Name,
		RequestPath: entry.Path,
		Passed:      true,
	}

	resolved, err := ws.ResolveRequest(coll, chain, entry)
	if err != nil {
		result.Passed = false
		result.RequestError = err
		return result
	}
	req := resolved.Request

	// The execution variable set is the full scope chain plus the shared
	// process-env, environment, and runtime variables provided to the run or
	// set by earlier steps (chaining).
	execVars := resolved.Vars.Clone()
	for _, scope := range []variables.Scope{variables.ScopeProcessEnv, variables.ScopeEnvironment, variables.ScopeRuntime} {
		r.vars.Range(scope, func(key, value string) {
			execVars.Set(scope, key, value)
		})
	}
	r.vars = execVars

	// Pre-request script mutates the request and variables.
	if entry.File.PreRequest != "" {
		logs, err := r.runPreScript(&req, entry.File.PreRequest)
		result.Logs = append(result.Logs, logs...)
		if err != nil {
			result.Passed = false
			result.RequestError = fmt.Errorf("pre-request script: %w", err)
			return result
		}
	}

	resp, err := r.client.Execute(ctx, &req, r.vars)
	if err != nil {
		result.Passed = false
		result.RequestError = err
		return result
	}
	result.Response = resp

	// Post-request script inspects the response and registers tests.
	if entry.File.PostRequest != "" {
		logs, tests, err := r.runPostScript(&req, resp, entry.File.PostRequest)
		result.Logs = append(result.Logs, logs...)
		if err != nil {
			result.Passed = false
			result.RequestError = fmt.Errorf("post-request script: %w", err)
			return result
		}
		for _, t := range tests {
			result.Tests = append(result.Tests, TestResult{Name: t.Name, Passed: t.Fn()})
			if !result.Tests[len(result.Tests)-1].Passed {
				result.Passed = false
			}
		}
	}

	if !resp.OK() {
		result.Passed = false
	}
	return result
}

// runPreScript evaluates the pre-request script with the request bound.
func (r *Runner) runPreScript(req *request.Request, source string) ([]string, error) {
	view := scripting.NewRequestView(req)
	s := scripting.NewSandbox(scripting.SandboxOptions{
		GetVariable: func(name string) (string, bool) { return r.vars.Resolve(name) },
		SetVariable: func(name, value string) { r.vars.Set(variables.ScopeRuntime, name, value) },
	})
	s.BindRequest(view)
	if err := s.Run(source); err != nil {
		return s.Logs(), err
	}
	view.ApplyTo(req)
	return s.Logs(), nil
}

// runPostScript evaluates the post-request script with the response bound.
func (r *Runner) runPostScript(req *request.Request, resp *response.Response, source string) ([]string, []scripting.Test, error) {
	s := scripting.NewSandbox(scripting.SandboxOptions{
		GetVariable: func(name string) (string, bool) { return r.vars.Resolve(name) },
		SetVariable: func(name, value string) { r.vars.Set(variables.ScopeRuntime, name, value) },
	})
	s.BindResponse(scripting.NewResponseView(resp))
	if err := s.Run(source); err != nil {
		return s.Logs(), nil, err
	}
	return s.Logs(), s.Tests(), nil
}
