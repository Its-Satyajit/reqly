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
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Its-Satyajit/reqly/internal/openapi"
)

// schemaRef aliases the kin-openapi schema for helper signatures above.
type schemaRef = openapi3.SchemaRef

// OpenapiEndpoint is one explorer entry with inline schema views.
type OpenapiEndpoint struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationID string   `json:"operationId,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Summary     string   `json:"summary,omitempty"`

	// RequestSchema is the JSON-encoded request-body schema ("" when none).
	RequestSchema string `json:"requestSchema,omitempty"`
	// ResponseSchemas maps status code → JSON-encoded response schema.
	ResponseSchemas map[string]string `json:"responseSchemas,omitempty"`
}

// OpenapiExploreResult is the whole spec summary for the browser panel.
type OpenapiExploreResult struct {
	Title     string            `json:"title"`
	Version   string            `json:"version,omitempty"`
	Endpoints []OpenapiEndpoint `json:"endpoints"`
	Warnings  []string          `json:"warnings,omitempty"`
}

// resolveSpecPath loads an OpenAPI file by workspace-relative path.
func (s *AppService) resolveSpecPath(path string) (string, error) {
	if s == nil || s.root == "" {
		return "", fmt.Errorf("no workspace found: open a reqly workspace to explore specs")
	}
	abs, err := s.resolveTestPath(path)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// OpenapiExplore parses a spec and returns every operation with inline
// request/response schema views.
func (s *AppService) OpenapiExplore(specPath string) (*OpenapiExploreResult, error) {
	abs, err := s.resolveSpecPath(specPath)
	if err != nil {
		return nil, err
	}
	doc, err := openapi.LoadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}
	endpoints := openapi.Explore(doc)
	title := ""
	version := ""
	if doc.Info != nil {
		title = doc.Info.Title
		version = doc.Info.Version
	}
	out := &OpenapiExploreResult{
		Title:     title,
		Version:   version,
		Endpoints: make([]OpenapiEndpoint, 0, len(endpoints)),
	}
	for _, ep := range endpoints {
		view := OpenapiEndpoint{
			Method:      ep.Method,
			Path:        ep.Path,
			OperationID: ep.OperationID,
			Tags:        ep.Tags,
			Summary:     ep.Summary,
		}
		op := lookupOperation(doc, ep.Method, ep.Path)
		if op != nil {
			view.RequestSchema = schemaJSON(requestSchemaOf(op))
			view.ResponseSchemas = map[string]string{}
			if op.Responses != nil {
				for status, respRef := range op.Responses.Map() {
					resp := respRef.Value
					if resp == nil {
						continue
					}
					for _, mt := range resp.Content {
						if sj := schemaJSON(mt.Schema); sj != "" {
							view.ResponseSchemas[status] = sj
							break
						}
					}
				}
			}
		}
		out.Endpoints = append(out.Endpoints, view)
	}
	return out, nil
}

// OpenapiGenerateSelection is one operation to generate a request for.
type OpenapiGenerateSelection struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// OpenapiGenerateResult reports where generated request files landed and how
// many were created; Warnings carries per-operation skips.
type OpenapiGenerateResult struct {
	TargetDir string   `json:"targetDir"`
	Created   []string `json:"created"`
	Warnings  []string `json:"warnings,omitempty"`
}

// OpenapiGenerateRequests renders native request files for the selected
// operations into `<workspace>/collections/<dirName>` — identical rendering
// to `reqly openapi generate`.
func (s *AppService) OpenapiGenerateRequests(specPath string, selections []OpenapiGenerateSelection, dirName string) (*OpenapiGenerateResult, error) {
	if len(selections) == 0 {
		return nil, fmt.Errorf("select at least one operation")
	}
	dirName = strings.TrimSpace(dirName)
	if dirName == "" {
		return nil, fmt.Errorf("folder name is required")
	}
	abs, err := s.resolveSpecPath(specPath)
	if err != nil {
		return nil, err
	}
	doc, err := openapi.LoadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}

	endpoints := openapi.Explore(doc)
	opIDByKey := map[string]string{}
	for _, ep := range endpoints {
		if ep.OperationID != "" {
			opIDByKey[ep.Method+"|"+ep.Path] = ep.OperationID
		}
	}

	var files []openapi.GeneratedFile
	var warnings []string
	for _, sel := range selections {
		method := strings.ToUpper(sel.Method)
		key := method + "|" + sel.Path
		var opts openapi.GenerateOptions
		if id, ok := opIDByKey[key]; ok {
			opts = openapi.GenerateOptions{Operations: []string{id}}
		} else {
			opts = openapi.GenerateOptions{Method: method, Path: sel.Path}
		}
		gf, warn, gerr := openapi.Generate(doc, opts)
		warnings = append(warnings, warn...)
		if gerr != nil {
			return nil, fmt.Errorf("generate %s %s: %w", method, sel.Path, gerr)
		}
		files = append(files, gf...)
	}
	target := filepath.Join(s.root, "collections", filepath.Base(dirName))
	created, err := openapi.Write(target, files)
	if err != nil {
		return nil, fmt.Errorf("write requests: %w", err)
	}
	return &OpenapiGenerateResult{TargetDir: target, Created: created, Warnings: warnings}, nil
}

// lookupOperation finds an operation by conventional "METHOD /path" key.
func lookupOperation(doc *openapi3.T, method, path string) *openapi3.Operation {
	if doc.Paths == nil {
		return nil
	}
	item := doc.Paths.Find(path)
	if item == nil {
		return nil
	}
	return item.Operations()[strings.ToUpper(method)]
}

// requestSchemaOf returns the first application/json-ish request schema.
func requestSchemaOf(op *openapi3.Operation) *schemaRef {
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}
	for _, mt := range op.RequestBody.Value.Content {
		return mt.Schema
	}
	return nil
}

// schemaJSON renders a schema reference as indented JSON ("" when nil).
func schemaJSON(ref *schemaRef) string {
	if ref == nil {
		return ""
	}
	data, err := json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}
