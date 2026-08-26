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

// Package testing implements the test engine: assertions, request/collection
// tests, data-driven testing, contract testing, and test reporting.
package testing

import (
	"fmt"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
)

// AssertionKind identifies what aspect of a response an assertion checks.
type AssertionKind string

const (
	// AssertStatus checks the HTTP status code.
	AssertStatus AssertionKind = "status"
	// AssertHeader checks a response header exists (and optionally its value).
	AssertHeader AssertionKind = "header"
	// AssertBodyContains checks the raw body contains a substring.
	AssertBodyContains AssertionKind = "body_contains"
	// AssertBodyEquals checks the raw body equals a string exactly.
	AssertBodyEquals AssertionKind = "body_equals"
	// AssertJSONPath checks a value extracted from a JSON body at a JSONPath.
	AssertJSONPath AssertionKind = "json"
	// AssertResponseTime checks the request duration against a threshold (ms).
	AssertResponseTime AssertionKind = "response_time"
)

// Assertion is a single check evaluated against a response.
type Assertion struct {
	Kind AssertionKind `json:"kind"`

	// Path is a JSONPath for AssertJSONPath, or a header name for AssertHeader.
	Path string `json:"path,omitempty"`

	// Value is the expected string value (header value, body substring, or the
	// JSON value rendered as text when StringEquals is set).
	Value string `json:"value,omitempty"`

	// Expected is the expected number for status / response_time assertions,
	// and the expected JSON value (when JSON is used as a number).
	Expected int64 `json:"expected,omitempty"`

	// Exact makes AssertJSONPath compare the extracted value as a string
	// instead of only checking presence.
	Exact bool `json:"exact,omitempty"`

	// Not inverts the assertion result.
	Not bool `json:"not,omitempty"`
}

// Result records the outcome of a single assertion.
type Result struct {
	Assertion Assertion `json:"assertion"`
	Passed    bool      `json:"passed"`
	Message   string    `json:"message,omitempty"`
}

// Test is a named collection of assertions.
type Test struct {
	Name       string      `json:"name"`
	Assertions []Assertion `json:"assertions"`
}

// TestResult records the outcome of a single test.
type TestResult struct {
	Name    string   `json:"name"`
	Passed  bool     `json:"passed"`
	Results []Result `json:"results"`
}

// Suite is a group of tests evaluated against one response.
type Suite struct {
	Name  string `json:"name"`
	Tests []Test `json:"tests"`
}

// Run evaluates every test in the suite against resp and returns the results.
// resp may be nil, in which case all assertions fail with a clear message.
func (s Suite) Run(resp *response.Response) []TestResult {
	results := make([]TestResult, 0, len(s.Tests))
	for _, test := range s.Tests {
		results = append(results, test.run(resp))
	}
	return results
}

// Passed reports whether every test in the suite passed.
func (s Suite) Passed(resp *response.Response) bool {
	for _, tr := range s.Run(resp) {
		if !tr.Passed {
			return false
		}
	}
	return true
}

func (t Test) run(resp *response.Response) TestResult {
	tr := TestResult{Name: t.Name, Passed: true}
	if resp == nil {
		tr.Passed = false
		tr.Results = append(tr.Results, Result{
			Assertion: Assertion{},
			Passed:    false,
			Message:   "no response available",
		})
		return tr
	}

	for _, a := range t.Assertions {
		r := evaluate(a, resp)
		if !r.Passed {
			tr.Passed = false
		}
		tr.Results = append(tr.Results, r)
	}
	return tr
}

// evaluate runs a single assertion against the response.
func evaluate(a Assertion, resp *response.Response) Result {
	passed := true
	var message string

	switch a.Kind {
	case AssertStatus:
		passed = int64(resp.StatusCode) == a.Expected
		message = fmt.Sprintf("status %d == %d", resp.StatusCode, a.Expected)

	case AssertHeader:
		values := resp.Headers[a.Path]
		if len(values) == 0 {
			passed = false
			message = fmt.Sprintf("header %q not present", a.Path)
		} else if a.Value != "" {
			joined := strings.Join(values, ", ")
			passed = strings.Contains(joined, a.Value)
			message = fmt.Sprintf("header %q contains %q (got %q)", a.Path, a.Value, joined)
		} else {
			message = fmt.Sprintf("header %q present", a.Path)
		}

	case AssertBodyContains:
		passed = strings.Contains(resp.Text(), a.Value)
		message = fmt.Sprintf("body contains %q", a.Value)

	case AssertBodyEquals:
		passed = resp.Text() == a.Value
		message = fmt.Sprintf("body equals %q", a.Value)

	case AssertJSONPath:
		if resp.JSONValue(a.Path) == nil {
			passed = false
			message = fmt.Sprintf("no value at %q", a.Path)
		} else if a.Exact {
			got := fmt.Sprintf("%v", resp.JSONValue(a.Path))
			passed = got == a.Value
			message = fmt.Sprintf("json %q == %q (got %q)", a.Path, a.Value, got)
		} else {
			message = fmt.Sprintf("value exists at %q", a.Path)
		}

	case AssertResponseTime:
		limit := time.Duration(a.Expected) * time.Millisecond
		passed = resp.Duration <= limit
		message = fmt.Sprintf("response time %s <= %s", resp.Duration.Round(time.Millisecond), limit)

	default:
		passed = false
		message = fmt.Sprintf("unknown assertion kind %q", a.Kind)
	}

	if a.Not {
		passed = !passed
		message = "NOT(" + message + ")"
	}

	return Result{Assertion: a, Passed: passed, Message: message}
}
