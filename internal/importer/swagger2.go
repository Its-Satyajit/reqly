// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

type swagger2Doc struct {
	Swagger  string                 `json:"swagger" yaml:"swagger"`
	Info     openapiInfo            `json:"info" yaml:"info"`
	Host     string                 `json:"host" yaml:"host"`
	BasePath string                 `json:"basePath" yaml:"basePath"`
	Schemes  []string               `json:"schemes" yaml:"schemes"`
	Paths    map[string]openapiPath `json:"paths" yaml:"paths"`
}

// ParseSwagger2 parses a Swagger 2.0 / OpenAPI 2.0 document.
func ParseSwagger2(data []byte) (*swagger2Doc, error) {
	var doc swagger2Doc
	if err := json.Unmarshal(data, &doc); err == nil && doc.Swagger == "2.0" {
		return &doc, nil
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse Swagger 2.0 document: %w", err)
	}
	if doc.Swagger != "2.0" {
		return nil, fmt.Errorf("not a Swagger 2.0 document")
	}
	return &doc, nil
}

func (d *swagger2Doc) ToOpenAPIResult() *OpenAPIResult {
	scheme := "https"
	if len(d.Schemes) > 0 && d.Schemes[0] != "" {
		scheme = d.Schemes[0]
	}
	base := ""
	if d.Host != "" {
		base = fmt.Sprintf("%s://%s%s", scheme, d.Host, d.BasePath)
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
