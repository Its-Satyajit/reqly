// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package bulk

import (
	"context"
	"strings"
	"sync"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// Options configures bulk execution.
type Options struct {
	Parallel        bool
	Concurrency     int
	ContinueOnError bool
}

// Step is one iteration of the bulk loop.
type Step struct {
	Index    int
	Row      map[string]string
	Request  request.Request
	Response *response.Response
	Err      error
}

// SendFunc sends a request (already interpolated for the row) and returns a response.
type SendFunc func(context.Context, request.Request) (*response.Response, error)

// Run executes req repeatedly for each row, interpolating row values.
func Run(ctx context.Context, req request.Request, rows []map[string]string, opts Options, sendFn SendFunc, onStep func(Step)) error {
	if len(rows) == 0 {
		return nil
	}
	concurrency := opts.Concurrency
	if !opts.Parallel {
		concurrency = 1
	} else {
		if concurrency <= 0 {
			concurrency = 5
		}
	}

	// Sequential path
	if concurrency == 1 {
		for i, row := range rows {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			reqCopy := interpolateRequest(req, row)
			resp, err := sendFn(ctx, reqCopy)
			step := Step{Index: i + 1, Row: row, Request: reqCopy, Response: resp, Err: err}
			if onStep != nil {
				onStep(step)
			}
			if err != nil {
				if !opts.ContinueOnError {
					return err
				}
				continue
			}
			if resp != nil && resp.StatusCode != 0 && (resp.StatusCode < 200 || resp.StatusCode >= 300) && !opts.ContinueOnError {
				break
			}
		}
		return nil
	}

	// Parallel path: ordered results
	type result struct {
		step Step
		err  error
	}
	results := make([]result, len(rows))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, row := range rows {
		i, row := i, row
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				mu.Lock()
				results[i] = result{step: Step{Index: i + 1, Row: row}, err: ctx.Err()}
				mu.Unlock()
				return
			}
			defer func() { <-sem }()

			// If context cancelled and not continue, skip
			select {
			case <-ctx.Done():
				mu.Lock()
				results[i] = result{step: Step{Index: i + 1, Row: row}, err: ctx.Err()}
				mu.Unlock()
				return
			default:
			}

			reqCopy := interpolateRequest(req, row)
			resp, err := sendFn(ctx, reqCopy)
			step := Step{Index: i + 1, Row: row, Request: reqCopy, Response: resp, Err: err}
			mu.Lock()
			results[i] = result{step: step, err: err}
			if err != nil && !opts.ContinueOnError && firstErr == nil {
				firstErr = err
				cancel()
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Report in order
	for _, r := range results {
		if r.step.Response != nil || r.step.Err != nil || r.step.Request.URL != "" {
			if onStep != nil {
				onStep(r.step)
			}
		} else if r.err != nil {
			if onStep != nil {
				onStep(Step{Index: 0, Err: r.err})
			}
		}
		// Stop on first non-2xx if not continue (but we already cancelled)
		if !opts.ContinueOnError && r.step.Response != nil && r.step.Response.StatusCode != 0 && (r.step.Response.StatusCode < 200 || r.step.Response.StatusCode >= 300) {
			break
		}
		if !opts.ContinueOnError && r.step.Err != nil {
			break
		}
	}
	if firstErr != nil && !opts.ContinueOnError {
		return firstErr
	}
	return nil
}

func interpolateRequest(req request.Request, row map[string]string) request.Request {
	out := req
	out.URL = interpolateString(req.URL, row)
	out.Body = interpolateString(req.Body, row)
	// headers
	if len(req.Headers) > 0 {
		out.Headers = make([]request.Header, len(req.Headers))
		copy(out.Headers, req.Headers)
		for i, h := range out.Headers {
			out.Headers[i].Key = interpolateString(h.Key, row)
			out.Headers[i].Value = interpolateString(h.Value, row)
		}
	}
	if len(req.Query) > 0 {
		out.Query = make([]request.Parameter, len(req.Query))
		copy(out.Query, req.Query)
		for i, q := range out.Query {
			out.Query[i].Key = interpolateString(q.Key, row)
			out.Query[i].Value = interpolateString(q.Value, row)
		}
	}
	// auth config
	if len(req.Auth.Config) > 0 {
		out.Auth.Config = make(map[string]string, len(req.Auth.Config))
		for k, v := range req.Auth.Config {
			out.Auth.Config[interpolateString(k, row)] = interpolateString(v, row)
		}
	}
	return out
}

func interpolateString(s string, row map[string]string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	for k, v := range row {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
		s = strings.ReplaceAll(s, "{{ "+k+" }}", v)
	}
	return s
}
