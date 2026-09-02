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
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleRunExportInput() RunReportExportInput {
	return RunReportExportInput{
		Path:       "collections/users/coll.reqly.yaml",
		Started:    "2026-08-24T10:00:00Z",
		Finished:   "2026-08-24T10:00:02Z",
		DurationMS: 2000,
		Steps: []runReportStep{
			{Name: "list", RequestPath: "collections/users/list.json", Passed: true, DurationMS: 120},
			{
				Name:         "create",
				RequestPath:  "collections/users/create.json",
				Passed:       false,
				RequestError: "connect timeout",
				Tests:        []runReportTest{{Name: "status is 201", Passed: false}},
			},
		},
	}
}

func TestRunExportReportRequiresWorkspace(t *testing.T) {
	svc := &AppService{}
	if _, err := svc.RunExportReport("json", sampleRunExportInput()); err == nil {
		t.Fatal("expected error without a workspace, got nil")
	}
}

func TestRunExportReportRejectsUnknownFormat(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)
	if _, err := svc.RunExportReport("pdf", sampleRunExportInput()); err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}

func TestRunExportReportJSON(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.RunExportReport("json", sampleRunExportInput())
	if err != nil {
		t.Fatalf("RunExportReport json: %v", err)
	}
	if !strings.HasSuffix(res.Path, ".json") {
		t.Errorf("Path = %q, want .json suffix", res.Path)
	}
	if !strings.Contains(res.Content, `"failed": 1`) || !strings.Contains(res.Content, `"requestError": "connect timeout"`) {
		t.Errorf("JSON content missing expected fields:\n%s", res.Content)
	}
	if _, err := os.Stat(res.Path); err != nil {
		t.Errorf("report not on disk: %v", err)
	}
}

func TestRunExportReportJUnit(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.RunExportReport("junit", sampleRunExportInput())
	if err != nil {
		t.Fatalf("RunExportReport junit: %v", err)
	}
	if !strings.HasSuffix(res.Path, ".xml") {
		t.Errorf("Path = %q, want .xml suffix", res.Path)
	}
	var suite struct {
		XMLName  xml.Name `xml:"testsuite"`
		Tests    int      `xml:"tests,attr"`
		Failures int      `xml:"failures,attr"`
		Cases    []struct {
			Name     string `xml:"name,attr"`
			Failures []struct {
				Body string `xml:",chardata"`
			} `xml:"failure"`
		} `xml:"testcase"`
	}
	if err := xml.Unmarshal([]byte(res.Content), &suite); err != nil {
		t.Fatalf("JUnit XML invalid: %v", err)
	}
	if suite.Tests != 2 || suite.Failures != 1 || len(suite.Cases) != 2 {
		t.Fatalf("suite = tests:%d failures:%d cases:%d", suite.Tests, suite.Failures, len(suite.Cases))
	}
	var createCase *struct {
		Name     string `xml:"name,attr"`
		Failures []struct {
			Body string `xml:",chardata"`
		} `xml:"failure"`
	}
	for i := range suite.Cases {
		if suite.Cases[i].Name == "create" {
			createCase = &suite.Cases[i]
		}
	}
	if createCase == nil {
		t.Fatal("no testcase named create")
	}
	var bodies []string
	for _, f := range createCase.Failures {
		bodies = append(bodies, f.Body)
	}
	if !strings.Contains(strings.Join(bodies, "\n"), "connect timeout") {
		t.Errorf("create case failures missing request error: %v", bodies)
	}
}

func TestRunExportReportWritesUnderExports(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedExportWorkspace(t, wsDir)

	res, err := svc.RunExportReport("junit", sampleRunExportInput())
	if err != nil {
		t.Fatalf("RunExportReport: %v", err)
	}
	want := filepath.Join(wsDir, ".reqly", "exports")
	if filepath.Dir(res.Path) != want {
		t.Errorf("export dir = %q, want under %q", filepath.Dir(res.Path), want)
	}
}
