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

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

func sampleResponse() *response.Response {
	return &response.Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"user":{"name":"reqly","age":30},"tags":["a","b"]}`),
		Duration:   50 * time.Millisecond,
	}
}

func TestStatusAssertion(t *testing.T) {
	resp := sampleResponse()

	if r := evaluate(Assertion{Kind: AssertStatus, Expected: 200}, resp); !r.Passed {
		t.Fatalf("expected 200 to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertStatus, Expected: 404}, resp); r.Passed {
		t.Fatal("expected 404 to fail")
	}
	if r := evaluate(Assertion{Kind: AssertStatus, Expected: 200, Not: true}, resp); r.Passed {
		t.Fatal("expected inverted 200 to fail")
	}
}

func TestHeaderAssertion(t *testing.T) {
	resp := sampleResponse()

	if r := evaluate(Assertion{Kind: AssertHeader, Path: "Content-Type"}, resp); !r.Passed {
		t.Fatalf("expected present header to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertHeader, Path: "Content-Type", Value: "json"}, resp); !r.Passed {
		t.Fatalf("expected matching header value to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertHeader, Path: "Content-Type", Value: "xml"}, resp); r.Passed {
		t.Fatal("expected non-matching header value to fail")
	}
	if r := evaluate(Assertion{Kind: AssertHeader, Path: "X-Missing"}, resp); r.Passed {
		t.Fatal("expected missing header to fail")
	}
}

func TestBodyAssertions(t *testing.T) {
	resp := sampleResponse()

	if r := evaluate(Assertion{Kind: AssertBodyContains, Value: "reqly"}, resp); !r.Passed {
		t.Fatalf("expected body contains to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertBodyContains, Value: "nope"}, resp); r.Passed {
		t.Fatal("expected body contains to fail")
	}
	if r := evaluate(Assertion{Kind: AssertBodyEquals, Value: string(resp.Body)}, resp); !r.Passed {
		t.Fatalf("expected body equals to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertBodyEquals, Value: "other"}, resp); r.Passed {
		t.Fatal("expected body equals to fail")
	}
}

func TestJSONAssertion(t *testing.T) {
	resp := sampleResponse()

	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.user.name"}, resp); !r.Passed {
		t.Fatalf("expected existing json path to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.missing"}, resp); r.Passed {
		t.Fatal("expected missing json path to fail")
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.user.name", Exact: true, Value: "reqly"}, resp); !r.Passed {
		t.Fatalf("expected exact json match to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.user.age", Exact: true, Value: "30"}, resp); !r.Passed {
		t.Fatalf("expected numeric json match to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.user.name", Exact: true, Value: "other"}, resp); r.Passed {
		t.Fatal("expected exact json mismatch to fail")
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.tags[1]"}, resp); !r.Passed {
		t.Fatalf("expected array index json path to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertJSONPath, Path: "$.tags[9]"}, resp); r.Passed {
		t.Fatal("expected out-of-range array index to fail")
	}
}

func TestResponseTimeAssertion(t *testing.T) {
	resp := sampleResponse()

	if r := evaluate(Assertion{Kind: AssertResponseTime, Expected: 100}, resp); !r.Passed {
		t.Fatalf("expected 50ms <= 100ms to pass, got %q", r.Message)
	}
	if r := evaluate(Assertion{Kind: AssertResponseTime, Expected: 10}, resp); r.Passed {
		t.Fatal("expected 50ms <= 10ms to fail")
	}
}

func TestUnknownAssertion(t *testing.T) {
	resp := sampleResponse()
	if r := evaluate(Assertion{Kind: "bogus"}, resp); r.Passed {
		t.Fatal("expected unknown assertion kind to fail")
	}
}

func TestSuiteRunAndPassed(t *testing.T) {
	resp := sampleResponse()

	suite := Suite{Name: "smoke", Tests: []Test{
		{
			Name: "status and body",
			Assertions: []Assertion{
				{Kind: AssertStatus, Expected: 200},
				{Kind: AssertBodyContains, Value: "reqly"},
			},
		},
		{
			Name: "json shape",
			Assertions: []Assertion{
				{Kind: AssertJSONPath, Path: "$.user.name"},
			},
		},
	}}

	results := suite.Run(resp)
	if len(results) != 2 {
		t.Fatalf("expected 2 test results, got %d", len(results))
	}
	for _, tr := range results {
		if !tr.Passed {
			t.Fatalf("test %q unexpectedly failed", tr.Name)
		}
	}
	if !suite.Passed(resp) {
		t.Fatal("expected suite to pass")
	}
}

func TestSuiteRunWithFailure(t *testing.T) {
	resp := sampleResponse()

	suite := Suite{Tests: []Test{
		{Name: "fail", Assertions: []Assertion{
			{Kind: AssertStatus, Expected: 500},
		}},
	}}

	results := suite.Run(resp)
	if results[0].Passed {
		t.Fatal("expected test to fail")
	}
	if suite.Passed(resp) {
		t.Fatal("expected suite to not pass")
	}
}

func TestSuiteRunNilResponse(t *testing.T) {
	suite := Suite{Tests: []Test{{Name: "t", Assertions: []Assertion{
		{Kind: AssertStatus, Expected: 200},
	}}}}

	results := suite.Run(nil)
	if len(results) != 1 || results[0].Passed {
		t.Fatalf("expected nil response to fail, got %+v", results)
	}
}

func TestSuiteRunEmpty(t *testing.T) {
	suite := Suite{}
	if !suite.Passed(sampleResponse()) {
		t.Fatal("expected empty suite to pass")
	}
}

func TestSuiteRunAgainstRealHTTP(t *testing.T) {
	// Integration: execute via the shared request engine and assert on it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write([]byte(`{"id":42}`))
	}))
	defer srv.Close()

	client := request.NewClient()
	ctx := context.Background()
	resp, err := client.Execute(ctx, &request.Request{
		Method: request.MethodGet,
		URL:    srv.URL + "/users",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}

	suite := Suite{Tests: []Test{
		{Name: "created", Assertions: []Assertion{
			{Kind: AssertStatus, Expected: 201},
			{Kind: AssertHeader, Path: "Content-Type", Value: "json"},
			{Kind: AssertJSONPath, Path: "$.id", Exact: true, Value: "42"},
		}},
	}}

	if !suite.Passed(resp) {
		results := suite.Run(resp)
		t.Fatalf("expected suite to pass, got %+v", results)
	}
}

func TestJSONSchemaAssertion(t *testing.T) {
	tmp := t.TempDir()
	schemaFile := filepath.Join(tmp, "user-schema.json")
	schemaContent := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["user", "tags"],
		"properties": {
			"user": {
				"type": "object",
				"required": ["name", "age"],
				"properties": {
					"name": { "type": "string" },
					"age": { "type": "integer" }
				}
			},
			"tags": {
				"type": "array",
				"items": { "type": "string" }
			}
		}
	}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	resp := sampleResponse()
	res := evaluate(Assertion{Kind: AssertJSONSchema, Path: schemaFile}, resp)
	if !res.Passed {
		t.Fatalf("expected valid JSON schema assertion to pass, got: %s", res.Message)
	}
}

func TestJSONSchemaAssertionViolations(t *testing.T) {
	tmp := t.TempDir()
	schemaFile := filepath.Join(tmp, "strict-schema.json")
	schemaContent := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["email"]
	}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	resp := sampleResponse()
	res := evaluate(Assertion{Kind: AssertJSONSchema, Path: schemaFile}, resp)
	if res.Passed {
		t.Fatal("expected schema validation to fail on missing email property")
	}
	if !strings.Contains(res.Message, "missing") && !strings.Contains(res.Message, "required") && !strings.Contains(res.Message, "email") {
		t.Fatalf("expected violation details in message, got: %s", res.Message)
	}
}
