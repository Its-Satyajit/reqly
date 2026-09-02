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

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleTestYAML = `name: users
request:
  method: GET
  url: https://api.example.com/users
tests:
  - name: ok
    assertions:
      - kind: status
        expected: 200
      - kind: body_contains
        value: users
`

func TestTestsListFindsReqlyTestFiles(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	collDir := filepath.Join(wsDir, "collections", "users")
	if err := os.MkdirAll(collDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(collDir, "users.reqly-test.yaml"), []byte(sampleTestYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	skipDir := filepath.Join(wsDir, ".reqly")
	if err := os.MkdirAll(skipDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skipDir, "hidden.reqly-test.yaml"), []byte(sampleTestYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	refs, err := svc.TestsList()
	if err != nil {
		t.Fatalf("TestsList: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("len(refs) = %d, want 1: %+v", len(refs), refs)
	}
	if refs[0].Name != "users" || refs[0].Path != "collections/users/users.reqly-test.yaml" {
		t.Errorf("ref = %+v", refs[0])
	}
}

func TestTestFileWriteRejectsInvalidSuite(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	target := filepath.Join(wsDir, "bad.reqly-test.yaml")
	err := svc.TestFileWrite("bad.reqly-test.yaml", "::: not yaml {{{")
	if err == nil {
		t.Fatal("write of invalid suite succeeded, want error")
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Error("invalid content was written to disk")
	}
}

func TestTestFileWriteAndReadRoundTrip(t *testing.T) {
	svc, _ := newServiceInWorkspace(t)
	if err := svc.TestFileWrite("t/users.reqly-test.yaml", sampleTestYAML); err != nil {
		t.Fatalf("TestFileWrite: %v", err)
	}
	got, err := svc.TestFileRead("t/users.reqly-test.yaml")
	if err != nil {
		t.Fatalf("TestFileRead: %v", err)
	}
	if got.Content != sampleTestYAML {
		t.Errorf("content mismatch:\n%q", got.Content)
	}
	if got.Format != "yaml" {
		t.Errorf("Format = %q, want yaml", got.Format)
	}
}

func TestTestRunReportsAssertionResults(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	content := strings.Replace(sampleTestYAML,
		"https://api.example.com/users",
		"scheme://nope.invalid/users", 1)
	res, err := svc.TestRun(TestRunRequest{Path: "x.reqly-test.yaml", Content: content})
	if err != nil {
		t.Fatalf("TestRun: %v", err)
	}
	if res.Passed {
		t.Fatal("run against unreachable host passed")
	}
	if res.Error == "" {
		t.Error("Error empty, want transport failure note")
	}
	if res.Results != nil {
		t.Errorf("Results = %+v, want nil when request failed", res.Results)
	}
	_ = wsDir
}

func TestTestRunWithSuccessfulAssertions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"users":[{"id":1}]}`))
	}))
	defer srv.Close()
	content := strings.Replace(sampleTestYAML, "https://api.example.com/users", srv.URL+"/users", 1)
	svc, _ := newServiceInWorkspace(t)
	res, err := svc.TestRun(TestRunRequest{Content: content})
	if err != nil {
		t.Fatalf("TestRun: %v", err)
	}
	if !res.Passed {
		t.Fatalf("suite failed: %+v", res)
	}
	if res.PassCount != 1 || res.Total != 1 {
		t.Errorf("PassCount/Total = %d/%d, want 1/1", res.PassCount, res.Total)
	}
	if len(res.Results) == 0 || len(res.Results[0].Results) == 0 {
		t.Fatal("no assertion results returned")
	}
	for _, r := range res.Results[0].Results {
		if !r.Passed {
			t.Errorf("assertion failed: %s", r.Message)
		}
	}
}
