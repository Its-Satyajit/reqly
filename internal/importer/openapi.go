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

package importer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// openapiDoc is the subset of an OpenAPI 3.x document the importer understands.
type openapiDoc struct {
	OpenAPI string                 `json:"openapi" yaml:"openapi"`
	Info    openapiInfo            `json:"info" yaml:"info"`
	Servers []openapiServer        `json:"servers" yaml:"servers"`
	Paths   map[string]openapiPath `json:"paths" yaml:"paths"`
}

type openapiInfo struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type openapiServer struct {
	URL string `json:"url" yaml:"url"`
}

type openapiPath struct {
	Get        *openapiOperation  `json:"get" yaml:"get"`
	Put        *openapiOperation  `json:"put" yaml:"put"`
	Post       *openapiOperation  `json:"post" yaml:"post"`
	Delete     *openapiOperation  `json:"delete" yaml:"delete"`
	Patch      *openapiOperation  `json:"patch" yaml:"patch"`
	Head       *openapiOperation  `json:"head" yaml:"head"`
	Options    *openapiOperation  `json:"options" yaml:"options"`
	Parameters []openapiParameter `json:"parameters" yaml:"parameters"`
}

type openapiOperation struct {
	OperationID string              `json:"operationId" yaml:"operationId"`
	Summary     string              `json:"summary" yaml:"summary"`
	Tags        []string            `json:"tags" yaml:"tags"`
	Parameters  []openapiParameter  `json:"parameters" yaml:"parameters"`
	RequestBody *openapiRequestBody `json:"requestBody" yaml:"requestBody"`
}

type openapiParameter struct {
	Name string `json:"name" yaml:"name"`
	In   string `json:"in" yaml:"in"`
}

type openapiRequestBody struct {
	Content map[string]any `json:"content" yaml:"content"`
}

// ParseOpenAPI parses an OpenAPI 3.x document (JSON or YAML).
func ParseOpenAPI(data []byte) (*openapiDoc, error) {
	var doc openapiDoc
	if err := json.Unmarshal(data, &doc); err == nil && doc.OpenAPI != "" {
		return &doc, nil
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse OpenAPI document: %w", err)
	}
	if doc.OpenAPI == "" {
		return nil, fmt.Errorf("not an OpenAPI document (missing openapi field)")
	}
	return &doc, nil
}

// OpenAPIResult is the in-memory result of importing an OpenAPI document.
type OpenAPIResult struct {
	Title       string
	BaseURL     string
	Collections []*OpenAPICollection
}

// OpenAPICollection groups imported requests by OpenAPI tag.
type OpenAPICollection struct {
	Name    string
	Request []*requestfile.File
}

// ToOpenAPIResult converts a parsed OpenAPI document into a request tree.
func (d *openapiDoc) ToOpenAPIResult() *OpenAPIResult {
	base := ""
	if len(d.Servers) > 0 {
		base = d.Servers[0].URL
	}

	groups := map[string]*OpenAPICollection{}
	for path, p := range d.Paths {
		for method, op := range p.operations() {
			tag := "default"
			if len(op.Tags) > 0 && op.Tags[0] != "" {
				tag = op.Tags[0]
			}
			coll := groups[tag]
			if coll == nil {
				coll = &OpenAPICollection{Name: tag}
				groups[tag] = coll
			}

			f := &requestfile.File{
				Name: operationFilename(method, op.OperationID, path),
				Request: request.Request{
					Name:   operationName(op, method, path),
					Method: request.Method(strings.ToUpper(method)),
					URL:    path,
				},
			}
			fillParameters(f, append(append([]openapiParameter{}, p.Parameters...), op.Parameters...))
			fillRequestBody(f, op.RequestBody)
			coll.Request = append(coll.Request, f)
		}
	}

	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)

	result := &OpenAPIResult{Title: d.Info.Title, BaseURL: base}
	for _, name := range names {
		result.Collections = append(result.Collections, groups[name])
	}
	return result
}

// operations returns the defined methods on a path.
func (p *openapiPath) operations() map[string]*openapiOperation {
	ops := map[string]*openapiOperation{}
	if p.Get != nil {
		ops["get"] = p.Get
	}
	if p.Put != nil {
		ops["put"] = p.Put
	}
	if p.Post != nil {
		ops["post"] = p.Post
	}
	if p.Delete != nil {
		ops["delete"] = p.Delete
	}
	if p.Patch != nil {
		ops["patch"] = p.Patch
	}
	if p.Head != nil {
		ops["head"] = p.Head
	}
	if p.Options != nil {
		ops["options"] = p.Options
	}
	return ops
}

// fillParameters adds query and header parameters to the request.
func fillParameters(f *requestfile.File, params []openapiParameter) {
	for _, p := range params {
		switch p.In {
		case "query":
			f.Request.Query = append(f.Request.Query, request.Parameter{Key: p.Name})
		case "header":
			f.Request.Headers = append(f.Request.Headers, request.Header{Key: p.Name})
		}
	}
}

// fillRequestBody maps the first JSON content type to a JSON body.
func fillRequestBody(f *requestfile.File, body *openapiRequestBody) {
	if body == nil {
		return
	}
	for contentType := range body.Content {
		if strings.Contains(contentType, "json") {
			f.Request.Headers = append(f.Request.Headers, request.Header{Key: "Content-Type", Value: contentType})
			f.Request.Body = "{}"
			return
		}
	}
}

// operationFilename builds a request file name from operationId or method+path.
func operationFilename(method, operationID, path string) string {
	if operationID != "" {
		return operationID
	}
	base := strings.Trim(path, "/")
	base = strings.ReplaceAll(base, "/", "-")
	base = strings.ReplaceAll(base, "{", "")
	base = strings.ReplaceAll(base, "}", "")
	if base == "" {
		base = "root"
	}
	return method + "-" + base
}

// operationName returns the display name for an operation.
func operationName(op *openapiOperation, method, path string) string {
	if op.Summary != "" {
		return op.Summary
	}
	if op.OperationID != "" {
		return op.OperationID
	}
	return strings.ToUpper(method) + " " + path
}

// Write writes the imported result to disk as a Git-native workspace:
// reqly.yaml descriptor + collections/<tag>/<operation>.yaml request files.
func (r *OpenAPIResult) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}

	wsCfg := map[string]any{
		"name":    r.Title,
		"baseURL": r.BaseURL,
	}
	if err := writeYAMLFile(filepath.Join(dir, "reqly.yaml"), wsCfg); err != nil {
		return err
	}

	for _, coll := range r.Collections {
		collDir := filepath.Join(dir, "collections", sanitizeName(coll.Name))
		if err := os.MkdirAll(collDir, 0o755); err != nil {
			return fmt.Errorf("create collection dir: %w", err)
		}
		collCfg := map[string]any{"name": coll.Name}
		if err := writeYAMLFile(filepath.Join(collDir, "reqly.yaml"), collCfg); err != nil {
			return err
		}
		for _, f := range coll.Request {
			path := filepath.Join(collDir, sanitizeName(f.Name)+".yaml")
			if err := writeYAMLFile(path, f); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeYAMLFile marshals v to a YAML file.
func writeYAMLFile(path string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// sanitizeName makes a string safe to use as a file or directory name.
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

// URL joins a path onto the result's base URL.
func (r *OpenAPIResult) URL(path string) (string, error) {
	if r.BaseURL == "" {
		return path, nil
	}
	u, err := url.Parse(r.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", r.BaseURL, err)
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")
	return u.String(), nil
}
