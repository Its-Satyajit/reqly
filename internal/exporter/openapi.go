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

package exporter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"go.yaml.in/yaml/v3"
)

// ---- OpenAPI 3.0 document shapes (only what the exporter emits) ----

type oaOperation struct {
	Summary     string                `yaml:"summary,omitempty"`
	OperationID string                `yaml:"operationId,omitempty"`
	Security    []map[string][]string `yaml:"security,omitempty"`
	Parameters  []oaParameter         `yaml:"parameters,omitempty"`
	RequestBody *oaRequestBody        `yaml:"requestBody,omitempty"`
	Responses   map[string]oaResponse `yaml:"responses"`
}

type oaParameter struct {
	Name   string            `yaml:"name"`
	In     string            `yaml:"in"`
	Schema map[string]string `yaml:"schema"`
}

type oaRequestBody struct {
	Required bool                   `yaml:"required"`
	Content  map[string]oaMediaType `yaml:"content"`
}

type oaMediaType struct {
	Schema map[string]string `yaml:"schema"`
}

type oaResponse struct {
	Description string `yaml:"description"`
}

type oaPathItem map[string]oaOperation

type oaSecurityScheme struct {
	Type   string `yaml:"type"`
	Scheme string `yaml:"scheme,omitempty"`
	In     string `yaml:"in,omitempty"`
	Name   string `yaml:"name,omitempty"`
}

type oaComponents struct {
	SecuritySchemes map[string]oaSecurityScheme `yaml:"securitySchemes,omitempty"`
}

type oaDoc struct {
	OpenAPI    string                `yaml:"openapi"`
	Info       oaInfo                `yaml:"info"`
	Servers    []oaServer            `yaml:"servers,omitempty"`
	Components *oaComponents         `yaml:"components,omitempty"`
	Paths      map[string]oaPathItem `yaml:"paths"`
}

type oaInfo struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type oaServer struct {
	URL string `yaml:"url"`
}

// ExportOpenAPI generates an OpenAPI 3.0 YAML document from requests.
// Response schemas are intentionally not invented: every operation declares a
// minimal default response only (a documented best-effort limit).
func ExportOpenAPI(title, baseURL string, requests []request.Request) ([]byte, error) {
	doc := oaDoc{
		OpenAPI: "3.0.3",
		Info:    oaInfo{Title: title, Version: "1.0.0"},
		Paths:   map[string]oaPathItem{},
	}
	if doc.Info.Title == "" {
		doc.Info.Title = "Reqly Collection"
	}
	if baseURL != "" {
		doc.Servers = []oaServer{{URL: baseURL}}
	}

	used := map[string]bool{}
	schemeNames := map[string]string{} // scheme key -> apikey header/query name
	for _, r := range requests {
		rawURL := strings.TrimSpace(r.URL)
		if rawURL == "" {
			continue
		}
		path, host := splitOAPath(rawURL)
		if path == "" {
			continue
		}
		item, ok := doc.Paths[path]
		if !ok {
			item = oaPathItem{}
			doc.Paths[path] = item
		}
		method := strings.ToLower(strings.TrimSpace(string(r.Method)))
		if method == "" {
			method = "get"
		}
		name := strings.TrimSpace(r.Name)
		if name == "" {
			name = strings.ToUpper(method) + " " + path
		}
		op := oaOperation{
			Summary:     name,
			OperationID: sanitizeOperationID(name),
			Responses: map[string]oaResponse{
				"default": {Description: "Response"},
			},
		}
		for _, q := range r.Query {
			op.Parameters = append(op.Parameters, oaParameter{
				Name: q.Key, In: "query", Schema: map[string]string{"type": "string"},
			})
		}
		ct := contentTypeOf(r.Headers)
		if r.Body != "" && ct == "" {
			ct = "application/octet-stream"
		}
		if r.Body != "" {
			op.RequestBody = &oaRequestBody{
				Required: true,
				Content:  map[string]oaMediaType{ct: {Schema: map[string]string{"type": "object"}}},
			}
		}
		switch strings.ToLower(r.Auth.Type) {
		case "basic":
			op.Security = []map[string][]string{{"basicAuth": {}}}
			used["basicAuth"] = true
		case "bearer":
			op.Security = []map[string][]string{{"bearerAuth": {}}}
			used["bearerAuth"] = true
		case "apikey":
			name := r.Auth.Config["key"]
			if name == "" {
				name = "X-API-Key"
			}
			in := strings.ToLower(r.Auth.Config["in"])
			if in == "" {
				in = "header"
			}
			key := "apiKeyHeader_" + name
			op.Security = []map[string][]string{{key: {}}}
			used[key] = true
			schemeNames[key] = in + "\x00" + name
		}
		item[method] = op
		if host != "" && len(doc.Servers) == 0 {
			doc.Servers = []oaServer{{URL: host}}
		}
	}

	if len(used) > 0 {
		comp := &oaComponents{SecuritySchemes: map[string]oaSecurityScheme{}}
		for key := range used {
			switch key {
			case "basicAuth":
				comp.SecuritySchemes[key] = oaSecurityScheme{Type: "http", Scheme: "basic"}
			case "bearerAuth":
				comp.SecuritySchemes[key] = oaSecurityScheme{Type: "http", Scheme: "bearer"}
			default:
				parts := strings.SplitN(schemeNames[key], "\x00", 2)
				comp.SecuritySchemes[key] = oaSecurityScheme{Type: "apiKey", In: parts[0], Name: parts[1]}
			}
		}
		doc.Components = comp
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return nil, fmt.Errorf("render OpenAPI document: %w", err)
	}
	return out, nil
}

// contentTypeOf returns the Content-Type header value, if any.
func contentTypeOf(headers []request.Header) string {
	for _, h := range headers {
		if strings.EqualFold(h.Key, "Content-Type") && h.Value != "" {
			return h.Value
		}
	}
	return ""
}

// splitOAPath extracts the URL path and scheme://host from a raw URL.
// Templated hosts ({{baseUrl}}/users) yield just their path portion.
func splitOAPath(rawURL string) (path, host string) {
	// Templated base: everything after the last closing brace is the path
	// (query string dropped — parameters come from r.Query).
	if strings.HasPrefix(rawURL, "{{") || strings.Contains(rawURL, "}}/") {
		if i := strings.LastIndex(rawURL, "}}"); i >= 0 {
			rest := rawURL[i+2:]
			if q := strings.Index(rest, "?"); q >= 0 {
				rest = rest[:q]
			}
			rest = strings.TrimSuffix(rest, "/")
			if rest == "" {
				return "", ""
			}
			return "/" + strings.TrimPrefix(rest, "/"), ""
		}
		return "", ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Path == "" {
		return "", ""
	}
	if u.Scheme != "" && u.Host != "" {
		host = u.Scheme + "://" + u.Host
	}
	return "/" + strings.TrimPrefix(u.Path, "/"), host
}

func sanitizeOperationID(name string) string {
	var sb strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		case r == ' ', r == '-', r == '_', r == '/':
			sb.WriteRune('_')
		}
	}
	out := sb.String()
	if out == "" {
		return "operation"
	}
	return out
}
