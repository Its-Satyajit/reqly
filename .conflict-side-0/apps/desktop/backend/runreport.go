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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/runner"
)

// Report export formats supported by the Run View — identical fidelity to
// `reqly collection test --junit/--json`.
const (
	reportFormatJSON  = "json"
	reportFormatJUnit = "junit"
)

// runReportTest is one assertion outcome as streamed to the frontend.
type runReportTest struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

// runReportStep mirrors the frontend RunStep for export purposes.
type runReportStep struct {
	Name         string          `json:"name"`
	RequestPath  string          `json:"requestPath,omitempty"`
	Passed       bool            `json:"passed"`
	RequestError string          `json:"requestError,omitempty"`
	DurationMS   int64           `json:"durationMs,omitempty"`
	Tests        []runReportTest `json:"tests,omitempty"`
}

// RunReportExportInput is the finished run's aggregate, sent back by the
// frontend when the user exports the report.
type RunReportExportInput struct {
	Path       string          `json:"path,omitempty"` // collection/folder path (suite name)
	Started    string          `json:"started,omitempty"`
	Finished   string          `json:"finished,omitempty"`
	DurationMS int64           `json:"durationMs"`
	Steps      []runReportStep `json:"steps"`
}

// RunExportResult reports where the export landed plus its content so the
// Run View can preview without re-reading the file.
type RunExportResult struct {
	Format  string `json:"format"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// RunExportReport serializes the last collection run as JSON or JUnit XML
// into `<root>/.reqly/exports/` and returns the rendered text.
func (s *AppService) RunExportReport(format string, in RunReportExportInput) (*RunExportResult, error) {
	if s == nil || s.workspace == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace before exporting")
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != reportFormatJSON && format != reportFormatJUnit {
		return nil, fmt.Errorf("unknown format %q: pick json or junit", format)
	}
	rep := runner.Report{
		Started:  parseOrNow(in.Started),
		Finished: parseOrNow(in.Finished),
		Duration: time.Duration(in.DurationMS) * time.Millisecond,
	}
	rep.Total = len(in.Steps)
	for _, st := range in.Steps {
		step := runner.StepResult{
			Name:        st.Name,
			RequestPath: st.RequestPath,
			Passed:      st.Passed,
		}
		if st.RequestError != "" {
			step.RequestError = errors.New(st.RequestError)
		}
		if st.DurationMS > 0 {
			// Only the duration matters for report rendering; status fields
			// stay zero so the export never fabricates response data.
			step.Response = &response.Response{Duration: time.Duration(st.DurationMS) * time.Millisecond}
		}
		for _, t := range st.Tests {
			step.Tests = append(step.Tests, runner.TestResult{Name: t.Name, Passed: t.Passed})
		}
		if step.Passed {
			rep.Passed++
		} else {
			rep.Failed++
		}
		rep.Steps = append(rep.Steps, step)
	}

	var data []byte
	var err error
	switch format {
	case reportFormatJSON:
		data, err = runner.JSONReport(&rep, nil)
	default:
		suite := strings.TrimSpace(in.Path)
		if suite == "" {
			suite = "reqly-run"
		}
		data, err = runner.JUnitReport(&rep, suite, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("render %s report: %w", format, err)
	}

	outDir := filepath.Join(s.root, ".reqly", "exports")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("create exports dir: %w", err)
	}
	base := strings.Trim(strings.TrimSpace(filepath.Base(strings.ReplaceAll(in.Path, "\\", "/"))), "/")
	if base == "" || base == "." {
		base = "run"
	}
	ext := ".json"
	if format == reportFormatJUnit {
		ext = ".xml"
	}
	name := fmt.Sprintf("%s-%s%s", base, time.Now().Format("20060102-150405"), ext)
	path := filepath.Join(outDir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, err
	}
	return &RunExportResult{Format: format, Path: path, Content: string(data)}, nil
}

func parseOrNow(ts string) time.Time {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}
	return time.Now()
}
