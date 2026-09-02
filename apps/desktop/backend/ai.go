// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/ai"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// AiExplain returns a templated explanation for a response JSON payload.
func (s *AppService) AiExplain(responseJSON string) (string, error) {
	if strings.TrimSpace(responseJSON) == "" {
		return ai.ExplainResponse(nil), nil
	}
	var resp response.Response
	if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
		// Fallback: treat as raw body with 200
		resp = response.Response{StatusCode: 200, StatusText: "OK", Body: []byte(responseJSON)}
	}
	return ai.ExplainResponse(&resp), nil
}

// AiGenerateTests synthesizes Goja test assertions for a response.
func (s *AppService) AiGenerateTests(responseJSON string) (string, error) {
	var resp response.Response
	if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
		resp = response.Response{StatusCode: 200, Body: []byte(responseJSON)}
	}
	return ai.GenerateTests(&resp), nil
}

// AiGenerateDocs synthesizes Markdown docs for a request/response pair.
func (s *AppService) AiGenerateDocs(requestJSON, responseJSON string) (string, error) {
	var resp response.Response
	if responseJSON != "" {
		_ = json.Unmarshal([]byte(responseJSON), &resp)
		if resp.Body == nil {
			resp.Body = []byte(responseJSON)
			if resp.StatusCode == 0 {
				resp.StatusCode = 200
			}
		}
	}
	return ai.GenerateDocs(nil, &resp), nil
}

// AiDiagnose analyzes an error or response status.
func (s *AppService) AiDiagnose(responseJSON, errMsg string) (string, error) {
	var resp *response.Response
	if responseJSON != "" {
		var r response.Response
		if err := json.Unmarshal([]byte(responseJSON), &r); err == nil {
			resp = &r
		} else {
			resp = &response.Response{StatusCode: 200, Body: []byte(responseJSON)}
		}
	}
	var err error
	if errMsg != "" {
		err = fmt.Errorf("%s", errMsg)
	}
	return ai.Diagnose(resp, err), nil
}

// AiExplainSchema returns a text explanation for a schema JSON.
func (s *AppService) AiExplainSchema(schemaJSON string) (string, error) {
	if strings.TrimSpace(schemaJSON) == "" {
		return "no schema", nil
	}
	trim := strings.TrimSpace(schemaJSON)
	if len(trim) > 100 {
		trim = trim[:100]
	}
	return fmt.Sprintf("Schema (%d bytes): %s", len(schemaJSON), trim), nil
}
