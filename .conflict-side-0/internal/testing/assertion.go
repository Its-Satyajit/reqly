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
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/jsonschema"
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
	// AssertJSONSchema validates the response body against a JSON Schema file.
	AssertJSONSchema AssertionKind = "json_schema"
)

// Assertion is a single check evaluated against a response.
type Assertion struct {
	Kind AssertionKind `json:"kind" yaml:"kind"`

	// Path is a JSONPath for AssertJSONPath, or a header name for AssertHeader.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Value is the expected string value (header value, body substring, or the
	// JSON value rendered as text when StringEquals is set).
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	// Expected is the expected number for status / response_time assertions,
	// and the expected JSON value (when JSON is used as a number).
	Expected int64 `json:"expected,omitempty" yaml:"expected,omitempty"`

	// Exact makes AssertJSONPath compare the extracted value as a string
	// instead of only checking presence.
	Exact bool `json:"exact,omitempty" yaml:"exact,omitempty"`

	// Not inverts the assertion result.
	Not bool `json:"not,omitempty" yaml:"not,omitempty"`
}

type rawAssertion struct {
	Kind      AssertionKind `json:"kind" yaml:"kind"`
	Path      string        `json:"path,omitempty" yaml:"path,omitempty"`
	Header    string        `json:"header,omitempty" yaml:"header,omitempty"`
	Name      string        `json:"name,omitempty" yaml:"name,omitempty"`
	Key       string        `json:"key,omitempty" yaml:"key,omitempty"`
	Value     string        `json:"value,omitempty" yaml:"value,omitempty"`
	Expected  any           `json:"expected,omitempty" yaml:"expected,omitempty"`
	Max       any           `json:"max,omitempty" yaml:"max,omitempty"`
	Threshold any           `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	MaxMs     any           `json:"maxMs,omitempty" yaml:"maxMs,omitempty"`
	Status    any           `json:"status,omitempty" yaml:"status,omitempty"`
	Exact     bool          `json:"exact,omitempty" yaml:"exact,omitempty"`
	Not       bool          `json:"not,omitempty" yaml:"not,omitempty"`
}

func (a *Assertion) applyRaw(raw *rawAssertion) {
	a.Kind = raw.Kind
	a.Path = raw.Path
	if a.Path == "" {
		if raw.Header != "" {
			a.Path = raw.Header
		} else if raw.Name != "" {
			a.Path = raw.Name
		} else if raw.Key != "" {
			a.Path = raw.Key
		}
	}
	a.Value = raw.Value
	a.Exact = raw.Exact
	a.Not = raw.Not

	a.Expected = parseExpectedInt64(raw.Expected)
	if a.Expected == 0 {
		if v := parseExpectedInt64(raw.Max); v != 0 {
			a.Expected = v
		} else if v := parseExpectedInt64(raw.Threshold); v != 0 {
			a.Expected = v
		} else if v := parseExpectedInt64(raw.MaxMs); v != 0 {
			a.Expected = v
		} else if v := parseExpectedInt64(raw.Status); v != 0 {
			a.Expected = v
		}
	}
}

// UnmarshalJSON unmarshals an Assertion, resolving aliases like header/name for Path, and max/threshold/maxMs/status for Expected.
func (a *Assertion) UnmarshalJSON(data []byte) error {
	var raw rawAssertion
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	a.applyRaw(&raw)
	return nil
}

// UnmarshalYAML unmarshals an Assertion from YAML, resolving aliases like header/name for Path, and max/threshold/maxMs/status for Expected.
func (a *Assertion) UnmarshalYAML(node *yaml.Node) error {
	var raw rawAssertion
	if err := node.Decode(&raw); err != nil {
		return err
	}
	a.applyRaw(&raw)
	return nil
}

func parseExpectedInt64(val any) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return n
		}
	}
	return 0
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

	case AssertJSONSchema:
		if strings.TrimSpace(a.Path) == "" {
			passed = false
			message = "json_schema: missing schema path"
			break
		}
		data, err := os.ReadFile(a.Path)
		if err != nil {
			passed = false
			message = fmt.Sprintf("json_schema: read %q: %v", a.Path, err)
			break
		}
		sch, err := jsonschema.Compile(data, a.Value)
		if err != nil {
			passed = false
			message = fmt.Sprintf("json_schema: compile %q: %v", a.Path, err)
			break
		}
		violations, err := jsonschema.Validate(sch, resp.Body)
		if err != nil {
			passed = false
			message = fmt.Sprintf("json_schema: validate: %v", err)
			break
		}
		if len(violations) == 0 {
			passed = true
			message = fmt.Sprintf("json_schema %q valid", a.Path)
		} else {
			passed = false
			parts := make([]string, 0, len(violations))
			for _, v := range violations {
				parts = append(parts, v.String())
			}
			message = fmt.Sprintf("json_schema %q: %s", a.Path, strings.Join(parts, "; "))
		}

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
