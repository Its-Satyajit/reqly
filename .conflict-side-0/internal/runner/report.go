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
	"fmt"
	"strings"
	"time"
)

// jsonStep is the serialization-safe projection of a StepResult: errors and
// responses become plain scalars so encoding/json never emits empty objects.
type jsonStep struct {
	Name         string       `json:"name"`
	RequestPath  string       `json:"requestPath,omitempty"`
	Passed       bool         `json:"passed"`
	RequestError string       `json:"requestError,omitempty"`
	StatusCode   int          `json:"statusCode,omitempty"`
	StatusText   string       `json:"statusText,omitempty"`
	DurationMS   int64        `json:"durationMs,omitempty"`
	Tests        []TestResult `json:"tests,omitempty"`
	Logs         []string     `json:"logs,omitempty"`
}

type jsonReport struct {
	Started    time.Time  `json:"started"`
	Finished   time.Time  `json:"finished"`
	Total      int        `json:"total"`
	Passed     int        `json:"passed"`
	Failed     int        `json:"failed"`
	DurationMS int64      `json:"durationMs"`
	Steps      []jsonStep `json:"steps"`
}

// JSONReport serializes a run report as indented JSON. mask, when non-nil,
// is applied to error text and script logs so secrets never reach disk.
func JSONReport(r *Report, mask func(string) string) ([]byte, error) {
	out := jsonReport{
		Started:    r.Started,
		Finished:   r.Finished,
		Total:      r.Total,
		Passed:     r.Passed,
		Failed:     r.Failed,
		DurationMS: r.Duration.Milliseconds(),
		Steps:      make([]jsonStep, 0, len(r.Steps)),
	}
	for _, s := range r.Steps {
		step := jsonStep{
			Name:        s.Name,
			RequestPath: s.RequestPath,
			Passed:      s.Passed,
			Tests:       s.Tests,
			Logs:        maskAll(s.Logs, mask),
		}
		if s.RequestError != nil {
			step.RequestError = applyMask(mask, s.RequestError.Error())
		}
		if s.Response != nil {
			step.StatusCode = s.Response.StatusCode
			step.StatusText = s.Response.StatusText
			step.DurationMS = s.Response.Duration.Milliseconds()
		}
		out.Steps = append(out.Steps, step)
	}
	return json.MarshalIndent(out, "", "  ")
}

// junitTestCase mirrors JUnit XML <testcase>.
type junitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	ClassName string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Time      string        `xml:"time,attr"`
	Failures  []junitFailur `xml:"failure"`
}

type junitFailur struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type junitSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      string          `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr"`
	Cases     []junitTestCase `xml:"testcase"`
}

// JUnitReport renders a run report as JUnit XML (one testsuite, one testcase
// per step). mask behaves as in JSONReport.
func JUnitReport(r *Report, suiteName string, mask func(string) string) ([]byte, error) {
	suite := junitSuite{
		Name:      suiteName,
		Tests:     r.Total,
		Failures:  r.Failed,
		Time:      seconds(r.Duration),
		Timestamp: r.Started.UTC().Format("2006-01-02T15:04:05"),
	}
	for _, s := range r.Steps {
		tc := junitTestCase{
			ClassName: stepClassName(s.RequestPath),
			Name:      s.Name,
			Time:      seconds(stepDuration(s)),
		}
		if !s.Passed {
			for _, tr := range s.Tests {
				if tr.Passed {
					continue
				}
				tc.Failures = append(tc.Failures, junitFailur{
					Message: "assertion failed",
					Type:    "assertion",
					Body:    applyMask(mask, tr.Name),
				})
			}
			if s.RequestError != nil {
				tc.Failures = append(tc.Failures, junitFailur{
					Message: "request failed",
					Type:    "request",
					Body:    applyMask(mask, s.RequestError.Error()),
				})
			}
			if len(tc.Failures) == 0 {
				tc.Failures = append(tc.Failures, junitFailur{Message: "step failed", Type: "unknown"})
			}
		}
		suite.Cases = append(suite.Cases, tc)
	}
	out, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("render JUnit report: %w", err)
	}
	return append([]byte(xml.Header+"\n"), out...), nil
}

func stepDuration(s StepResult) time.Duration {
	if s.Response != nil {
		return s.Response.Duration
	}
	return 0
}

// stepClassName uses the request's folder path as the JUnit classname.
func stepClassName(requestPath string) string {
	i := strings.LastIndex(requestPath, "/")
	if i <= 0 {
		return "collection"
	}
	return strings.ReplaceAll(requestPath[:i], "/", ".")
}

func maskAll(logs []string, mask func(string) string) []string {
	if mask == nil || len(logs) == 0 {
		return logs
	}
	out := make([]string, len(logs))
	for i, l := range logs {
		out[i] = mask(l)
	}
	return out
}

func applyMask(mask func(string) string, s string) string {
	if mask == nil {
		return s
	}
	return mask(s)
}

func seconds(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}
