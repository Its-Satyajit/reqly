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

package runner

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
)

func sampleReport() *Report {
	return &Report{
		Started:  time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC),
		Finished: time.Date(2026, 8, 23, 10, 0, 2, 0, time.UTC),
		Total:    3, Passed: 2, Failed: 1,
		Duration: 2 * time.Second,
		Steps: []StepResult{
			{
				Name:        "login",
				RequestPath: "users/auth/login",
				Passed:      true,
				Response:    &response.Response{StatusCode: 200, StatusText: "OK", Duration: 120 * time.Millisecond},
				Logs:        []string{"token=sekret"},
			},
			{
				Name:         "get-user",
				RequestPath:  "users/get-user",
				Passed:       false,
				RequestError: errors.New("connect refused host=secret-host"),
				Tests:        []TestResult{{Name: "status is 200", Passed: false}},
			},
			{
				Name:   "logout",
				Passed: true,
			},
		},
	}
}

func TestJSONReport(t *testing.T) {
	mask := func(s string) string { return strings.ReplaceAll(s, "sekret", "[MASKED]") }
	data, err := JSONReport(sampleReport(), mask)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	for _, want := range []string{`"total": 3`, `"failed": 1`, `"name": "login"`, `token=[MASKED]`, `"statusCode": 200`} {
		if !strings.Contains(doc, want) {
			t.Errorf("JSON report missing %q:\n%s", want, doc)
		}
	}
	if strings.Contains(doc, "sekret") {
		t.Errorf("JSON report leaks unmasked secret:\n%s", doc)
	}
}

func TestJUnitReport(t *testing.T) {
	mask := func(s string) string { return strings.ReplaceAll(s, "secret-host", "[MASKED]") }
	data, err := JUnitReport(sampleReport(), "users", mask)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	for _, want := range []string{
		`<testsuite name="users"`,
		`tests="3"`,
		`failures="1"`,
		`<testcase classname="users.auth" name="login"`,
		`classname="collection" name="logout"`,
		`type="request"`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("JUnit report missing %q:\n%s", want, doc)
		}
	}
	if !strings.Contains(doc, "[MASKED]") || strings.Contains(doc, "secret-host") {
		t.Errorf("JUnit report leak:\n%s", doc)
	}
}

func TestJUnitReportIsValidXML(t *testing.T) {
	data, err := JUnitReport(sampleReport(), "suite", nil)
	if err != nil {
		t.Fatal(err)
	}
	var suite junitSuite
	if err := xml.Unmarshal(data, &suite); err != nil {
		t.Fatalf("invalid XML: %v\n%s", err, data)
	}
	if len(suite.Cases) != 3 {
		t.Fatalf("cases = %d, want 3", len(suite.Cases))
	}
	failed := 0
	for _, c := range suite.Cases {
		failed += len(c.Failures)
	}
	if failed != 2 { // one request error + one assertion failure on get-user
		t.Fatalf("failures = %d, want 2", failed)
	}
}

func TestReportsHandleEmptyRun(t *testing.T) {
	r := &Report{}
	if _, err := JSONReport(r, nil); err != nil {
		t.Fatal(err)
	}
	data, err := JUnitReport(r, "empty", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `tests="0"`) {
		t.Fatalf("empty suite =\n%s", data)
	}
}
