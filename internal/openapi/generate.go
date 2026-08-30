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

package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// GenerateOptions selects which operations to generate request files for.
// Exactly one selector form must be used: All, Operations (operationIds),
// Tags, or the Method+Path pair.
type GenerateOptions struct {
	All        bool
	Operations []string
	Tags       []string
	Method     string
	Path       string
}

// GeneratedFile is one operation rendered as a native request file that has
// not yet been written to disk.
type GeneratedFile struct {
	Filename string            // base name without extension
	File     *requestfile.File // nil variables are declared empty placeholders
}

// Generate resolves the selectors against doc and renders each matched
// operation as a GeneratedFile. Warnings describe skipped bodies, unmappable
// security schemes, unresolved parameters, and filename collisions; they do
// not fail generation.
func Generate(doc *openapi3.T, opts GenerateOptions) ([]GeneratedFile, []string, error) {
	endpoints, err := selectEndpoints(doc, opts)
	if err != nil {
		return nil, nil, err
	}

	var warnings []string
	files := make([]GeneratedFile, 0, len(endpoints))
	for _, ep := range endpoints {
		item := doc.Paths.Value(ep.Path)
		op := item.GetOperation(ep.Method)

		f, warns := renderOperation(doc, ep, item.Parameters, op)
		filename, collides := uniqueFilename(operationFilename(ep.Method, ep.OperationID, ep.Path), files)
		if collides {
			warns = append(warns, fmt.Sprintf("%s: duplicate filename, written as %s.yaml", displayOp(ep), filename))
		}
		files = append(files, GeneratedFile{Filename: filename, File: f})
		warnings = append(warnings, warns...)
	}
	return files, warnings, nil
}

// Write writes generated files into dir as <filename>.yaml and returns the
// written paths.
func Write(dir string, files []GeneratedFile) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	paths := make([]string, 0, len(files))
	for _, gf := range files {
		path := filepath.Join(dir, gf.Filename+".yaml")
		data, err := yaml.Marshal(gf.File)
		if err != nil {
			return nil, fmt.Errorf("marshal %s: %w", path, err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

// renderOperation builds the request file for one operation.
func renderOperation(doc *openapi3.T, ep Endpoint, pathItemParams openapi3.Parameters, op *openapi3.Operation) (*requestfile.File, []string) {
	var warnings []string

	f := &requestfile.File{
		Name: displayName(ep, op),
		Request: request.Request{
			Name:   displayName(ep, op),
			Method: request.Method(ep.Method),
		},
		Variables: map[string]string{},
	}

	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = doc.Servers[0].URL
		f.Variables = map[string]string{"baseUrl": baseURL}
	}

	params := mergeParameters(pathItemParams, op.Parameters)
	url := ep.Path
	queryParams := []request.Parameter{}
	headers := []request.Header{}
	unfilled := []string{}

	for _, p := range params {
		if p == nil {
			continue
		}
		value, resolved := resolveParamValue(p)
		switch p.In {
		case "path":
			if !resolved {
				f.Variables[p.Name] = ""
				unfilled = append(unfilled, p.Name)
				continue
			}
			url = strings.ReplaceAll(url, "{"+p.Name+"}", value)
		case "query":
			if p.Required {
				if resolved {
					queryParams = append(queryParams, request.Parameter{Key: p.Name, Value: value})
				} else {
					queryParams = append(queryParams, request.Parameter{Key: p.Name})
					f.Variables[p.Name] = ""
					unfilled = append(unfilled, p.Name)
				}
			} else if resolved {
				queryParams = append(queryParams, request.Parameter{Key: p.Name, Value: value})
			}
		case "header":
			if p.Required {
				headers = append(headers, request.Header{Key: p.Name, Value: value})
				if !resolved {
					f.Variables[p.Name] = ""
				}
			}
		}
	}

	if baseURL != "" {
		url = "{{baseUrl}}" + url
	}
	f.Request.URL = url
	f.Request.Query = queryParams
	f.Request.Headers = headers

	if len(unfilled) > 0 {
		sort.Strings(unfilled)
		warnings = append(warnings, fmt.Sprintf("%s: unfilled required parameters %s — left as placeholders", displayOp(ep), strings.Join(unfilled, ", ")))
	}

	if op.Security != nil {
		warnings = append(warnings, applySecurity(doc, f, *op.Security, ep)...)
	} else if len(doc.Security) > 0 {
		warnings = append(warnings, applySecurity(doc, f, doc.Security, ep)...)
	}

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		bodyWarns := applyRequestBody(f, op.RequestBody.Value)
		for _, w := range bodyWarns {
			warnings = append(warnings, displayOp(ep)+": "+w)
		}
	}
	if len(f.Variables) == 0 {
		f.Variables = nil
	}
	return f, warnings
}

// applySecurity maps the first requirement's schemes onto native auth blocks;
// unmappable schemes are warned about and skipped.
func applySecurity(doc *openapi3.T, f *requestfile.File, reqs openapi3.SecurityRequirements, ep Endpoint) []string {
	var warnings []string
	for _, req := range reqs {
		for name := range req {
			schemeRef := doc.Components.SecuritySchemes[name]
			if schemeRef == nil || schemeRef.Value == nil {
				warnings = append(warnings, fmt.Sprintf("%s: security scheme %q not found in components; skipped", displayOp(ep), name))
				continue
			}
			scheme := schemeRef.Value
			switch {
			case scheme.Type == "http" && scheme.Scheme == "bearer":
				f.Request.Auth = request.Auth{Type: "bearer", Config: map[string]string{"token": "{{token}}"}}
				if f.Variables == nil {
					f.Variables = map[string]string{}
				}
				f.Variables["token"] = ""
				return warnings
			case scheme.Type == "http" && scheme.Scheme == "basic":
				f.Request.Auth = request.Auth{
					Type:   "basic",
					Config: map[string]string{"username": "{{username}}", "password": "{{password}}"},
				}
				if f.Variables == nil {
					f.Variables = map[string]string{}
				}
				f.Variables["username"] = ""
				f.Variables["password"] = ""
				return warnings
			case scheme.Type == "apiKey" && scheme.In == "header":
				f.Request.Auth = request.Auth{
					Type: "apikey",
					Config: map[string]string{
						"key":   scheme.Name,
						"value": "{{apiKey}}",
						"in":    "header",
					},
				}
				if f.Variables == nil {
					f.Variables = map[string]string{}
				}
				f.Variables["apiKey"] = ""
				return warnings
			default:
				warnings = append(warnings, fmt.Sprintf("%s: security scheme %q (%s%s%s) cannot be mapped to a request file; skipped",
					displayOp(ep), name, scheme.Type, schemeDesc(scheme), scheme.In))
			}
		}
	}
	return warnings
}

func schemeDesc(s *openapi3.SecurityScheme) string {
	switch s.Type {
	case "http":
		return "/" + s.Scheme + " "
	case "apiKey":
		return " in " + s.In + " "
	}
	return " "
}

// applyRequestBody renders an inline JSON literal from example precedence or
// warns when the content type cannot be mapped.
func applyRequestBody(f *requestfile.File, rb *openapi3.RequestBody) []string {
	var warnings []string
	content := rb.Content
	names := make([]string, 0, len(content))
	for name := range content {
		names = append(names, name)
	}
	sort.Strings(names)

	jsonType := ""
	for _, name := range names {
		if strings.Contains(name, "json") {
			jsonType = name
			break
		}
	}
	if jsonType == "" {
		if len(names) > 0 {
			warnings = append(warnings, fmt.Sprintf("request body %s not supported; omitted", strings.Join(names, ", ")))
		}
		return warnings
	}

	mt := content[jsonType]
	example := pickExample(mt)
	if example == nil {
		warnings = append(warnings, fmt.Sprintf("%s body has no example or default; empty object used", jsonType))
		example = map[string]any{}
	}
	data, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s body example is not JSON-serializable; omitted", jsonType))
		return warnings
	}

	f.Request.Body = string(data)
	hasCT := false
	for _, h := range f.Request.Headers {
		if strings.EqualFold(h.Key, "Content-Type") {
			hasCT = true
		}
	}
	if !hasCT {
		f.Request.Headers = append(f.Request.Headers, request.Header{Key: "Content-Type", Value: jsonType})
	}
	return warnings
}

// pickExample applies the example precedence: media-type example, first named
// example, schema example, schema default, schema exampleFromSchemas.
func pickExample(mt *openapi3.MediaType) any {
	if mt.Example != nil {
		return mt.Example
	}
	if len(mt.Examples) > 0 {
		names := make([]string, 0, len(mt.Examples))
		for name := range mt.Examples {
			names = append(names, name)
		}
		sort.Strings(names)
		if ex := mt.Examples[names[0]].Value; ex != nil && ex.Value != nil {
			return ex.Value
		}
	}
	if mt.Schema == nil || mt.Schema.Value == nil {
		return nil
	}
	schema := mt.Schema.Value
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Default != nil {
		return schema.Default
	}
	return nil
}

// resolveParamValue returns a concrete value for a parameter using the
// example → default precedence; ok reports whether one was found.
func resolveParamValue(p *openapi3.Parameter) (value string, ok bool) {
	if p.Example != nil {
		return stringify(p.Example), true
	}
	if p.Schema != nil && p.Schema.Value != nil {
		if p.Schema.Value.Example != nil {
			return stringify(p.Schema.Value.Example), true
		}
		if p.Schema.Value.Default != nil {
			return stringify(p.Schema.Value.Default), true
		}
	}
	return "", false
}

// stringify renders an example/default scalar as a URL-safe literal.
func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		data, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(data)
	}
}

// mergeParameters combines path-item and operation parameters, with operation
// parameters winning on (name, in) conflicts. Returned values are dereferenced.
func mergeParameters(pathItemParams, opParams openapi3.Parameters) []*openapi3.Parameter {
	seen := map[string]bool{}
	var out []*openapi3.Parameter
	for _, p := range opParams {
		if p.Value != nil {
			seen[p.Value.In+"|"+p.Value.Name] = true
			out = append(out, p.Value)
		}
	}
	for _, p := range pathItemParams {
		if p.Value == nil {
			continue
		}
		key := p.Value.In + "|" + p.Value.Name
		if seen[key] {
			continue
		}
		out = append(out, p.Value)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a == nil || b == nil {
			return false
		}
		return paramRank(a.In) < paramRank(b.In)
	})
	return out
}

func paramRank(in string) int {
	switch in {
	case "path":
		return 0
	case "header":
		return 1
	case "query":
		return 2
	}
	return 3
}

// operationFilename mirrors the importer convention: operationId when present,
// otherwise method plus path segments with separators stripped.
func operationFilename(method, operationID, path string) string {
	if operationID != "" {
		return sanitizeFilename(operationID)
	}
	base := strings.Trim(path, "/")
	base = strings.ReplaceAll(base, "/", "-")
	base = strings.ReplaceAll(base, "{", "")
	base = strings.ReplaceAll(base, "}", "")
	if base == "" {
		base = "root"
	}
	return strings.ToLower(method) + "-" + base
}

// uniqueFilename appends -2/-3/... suffixes on collisions within one run.
func uniqueFilename(base string, existing []GeneratedFile) (string, bool) {
	taken := map[string]bool{}
	for _, f := range existing {
		taken[f.Filename] = true
	}
	name := base
	n := 2
	for taken[name] {
		name = fmt.Sprintf("%s-%d", base, n)
		n++
	}
	return name, name != base
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

func displayName(ep Endpoint, op *openapi3.Operation) string {
	if op.Summary != "" {
		return op.Summary
	}
	if ep.OperationID != "" {
		return ep.OperationID
	}
	return ep.Method + " " + ep.Path
}

func displayOp(ep Endpoint) string {
	if ep.OperationID != "" {
		return ep.OperationID
	}
	return ep.Method + " " + ep.Path
}
