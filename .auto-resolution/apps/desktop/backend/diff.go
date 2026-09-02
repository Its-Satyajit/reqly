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
	"context"
	"encoding/json"
	"fmt"

	"github.com/Its-Satyajit/reqly/internal/diffing"
	"github.com/Its-Satyajit/reqly/internal/history"
	"gopkg.in/yaml.v3"
)

// SpecDiffResult is the payload for the Specs tab of the diff view.
type SpecDiffResult struct {
	Result   *diffing.DiffResult `json:"result"`
	Breaking int                 `json:"breaking"`
	Addition int                 `json:"addition"`
}

// ResponseDiffMeta describes one history entry compared in the diff view.
type ResponseDiffMeta struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Method  string `json:"method"`
	Status  int    `json:"status"`
	Env     string `json:"env,omitempty"`
	Preview string `json:"preview"` // truncated response body for side-by-side view
}

// ResponseDiffResult pairs structural changes with both entries' metadata.
type ResponseDiffResult struct {
	MetaA  *ResponseDiffMeta   `json:"metaA"`
	MetaB  *ResponseDiffMeta   `json:"metaB"`
	Result *diffing.DiffResult `json:"result"`
	ErrorA string              `json:"errorA,omitempty"`
	ErrorB string              `json:"errorB,omitempty"`
}

// DiffSpecs diffs two OpenAPI specs (workspace-relative or absolute paths)
// and classifies every change by severity.
func (s *AppService) DiffSpecs(pathA string, pathB string) (*SpecDiffResult, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to diff specs")
	}
	absA, err := s.resolveTestPath(pathA)
	if err != nil {
		return nil, err
	}
	absB, err := s.resolveTestPath(pathB)
	if err != nil {
		return nil, err
	}
	res, err := diffing.OpenAPIFiles(absA, absB)
	if err != nil {
		return nil, err
	}
	classified := diffing.WithSeverity(res)
	out := &SpecDiffResult{Result: classified}
	for _, c := range classified.Changes {
		switch c.Severity {
		case diffing.SeverityBreaking:
			out.Breaking++
		case diffing.SeverityNonBreaking:
			out.Addition++
		}
	}
	return out, nil
}

// DiffResponses compares two history entries' response bodies structurally
// and returns both entries' metadata for a side-by-side header.
func (s *AppService) DiffResponses(idA string, idB string) (*ResponseDiffResult, error) {
	h := s.hist()
	if h == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to diff responses")
	}
	entryA, err := h.Show(context.Background(), idA)
	if err != nil {
		return nil, fmt.Errorf("entry A: %w", err)
	}
	entryB, err := h.Show(context.Background(), idB)
	if err != nil {
		return nil, fmt.Errorf("entry B: %w", err)
	}
	res, err := diffing.JSON(entryA.RespBody, entryB.RespBody)
	if err != nil {
		return nil, fmt.Errorf("diff bodies: %w", err)
	}
	return &ResponseDiffResult{
		MetaA:  metaFromEntry(&entryA),
		MetaB:  metaFromEntry(&entryB),
		Result: res,
	}, nil
}

// DiffJSONText diffs two raw JSON/YAML documents pasted in the dialog —
// YAML input is converted before the structural comparison.
func (s *AppService) DiffJSONText(a string, b string) (*diffing.DiffResult, error) {
	jsonA, err := toJSONForDiff([]byte(a))
	if err != nil {
		return nil, fmt.Errorf("first document: %w", err)
	}
	jsonB, err := toJSONForDiff([]byte(b))
	if err != nil {
		return nil, fmt.Errorf("second document: %w", err)
	}
	return diffing.JSON(jsonA, jsonB)
}

func toJSONForDiff(data []byte) ([]byte, error) {
	if json.Valid(data) {
		return data, nil
	}
	var obj any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("not valid JSON or YAML")
	}
	return json.Marshal(obj)
}

func metaFromEntry(e *history.Entry) *ResponseDiffMeta {
	body := ""
	if len(e.RespBody) > 2000 {
		body = string(e.RespBody[:2000])
	} else {
		body = string(e.RespBody)
	}
	return &ResponseDiffMeta{
		ID:      e.ID,
		URL:     e.URL,
		Method:  e.Method,
		Status:  e.Status,
		Env:     e.Env,
		Preview: body,
	}
}
