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
	"strconv"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"gopkg.in/yaml.v3"
)

// InsomniaEnvironment is one imported environment, destined for
// environments/<name>.yaml in the output workspace (shared by the Postman,
// Insomnia, and Bruno importers).
type InsomniaEnvironment struct {
	Name      string
	Variables map[string]string
	Secrets   map[string]string
	Warnings  []string
}

// InsomniaResult is a parsed Insomnia export, ready to be written as a
// Git-native workspace.
type InsomniaResult struct {
	Title        string
	Collection   string
	Root         *PostmanFolder
	Environments []InsomniaEnvironment
}

// RequestCount reports the total number of imported requests.
func (r *InsomniaResult) RequestCount() int { return countRequests(r.Root) }

// ---- shared shapes (JSON tags for v4, YAML tags for v5) ----

type inHeader struct {
	Name     string `json:"name" yaml:"name"`
	Value    string `json:"value" yaml:"value"`
	Disabled bool   `json:"disabled" yaml:"disabled"`
}

type inParam struct {
	Name     string `json:"name" yaml:"name"`
	Value    string `json:"value" yaml:"value"`
	Disabled bool   `json:"disabled" yaml:"disabled"`
	Type     string `json:"type" yaml:"type"`
	FileName string `json:"fileName" yaml:"fileName"`
}

type inBody struct {
	MimeType string    `json:"mimeType" yaml:"mimeType"`
	Text     string    `json:"text" yaml:"text"`
	Params   []inParam `json:"params" yaml:"params"`
}

type inRequestCore struct {
	Name           string         `json:"name" yaml:"name"`
	Method         string         `json:"method" yaml:"method"`
	URL            string         `json:"url" yaml:"url"`
	Headers        []inHeader     `json:"headers" yaml:"headers"`
	Parameters     []inParam      `json:"parameters" yaml:"parameters"`
	Body           *inBody        `json:"body" yaml:"body"`
	Authentication map[string]any `json:"authentication" yaml:"authentication"`
}

// ParseInsomnia parses an Insomnia export (v4 JSON or v5 YAML) into an
// InsomniaResult. Structural errors hard-error; version mismatches warn.
func ParseInsomnia(data []byte) (*InsomniaResult, []string, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, nil, fmt.Errorf("parse Insomnia export: empty input")
	}
	if strings.HasPrefix(trimmed, "{") {
		return parseInsomniaV4([]byte(trimmed))
	}
	return parseInsomniaV5([]byte(trimmed))
}

// ---- v4: flat JSON resource list linked by parentId ----

type inV4Resource struct {
	Type     string `json:"_type"`
	ID       string `json:"_id"`
	ParentID string `json:"parentId"`
	Name     string `json:"name"`
	inRequestCore
	Data map[string]any `json:"data"`
}

type inV4Export struct {
	ExportFormat int            `json:"__export_format"`
	Resources    []inV4Resource `json:"resources"`
}

func parseInsomniaV4(data []byte) (*InsomniaResult, []string, error) {
	var exp inV4Export
	if err := json.Unmarshal(data, &exp); err != nil {
		return nil, nil, fmt.Errorf("parse Insomnia v4 export: %w", err)
	}
	var warnings []string
	if exp.ExportFormat != 0 && exp.ExportFormat != 4 {
		warnings = append(warnings, fmt.Sprintf("__export_format is %d, expected 4; attempting best-effort import", exp.ExportFormat))
	}
	res := &InsomniaResult{Collection: "insomnia-import", Root: &PostmanFolder{Name: ""}}

	// Pass 1: one folder per request_group, keyed by its resource id.
	folders := map[string]*PostmanFolder{}
	for _, r := range exp.Resources {
		if r.Type == "request_group" {
			folders[r.ID] = &PostmanFolder{Name: r.Name}
		} else if r.Type == "workspace" && res.Title == "" {
			res.Title = r.Name
		}
	}
	// Attach folders to their parent folder (root when unknown).
	for _, r := range exp.Resources {
		if r.Type != "request_group" {
			continue
		}
		parent := res.Root
		if f, ok := folders[r.ParentID]; ok && f != folders[r.ID] {
			parent = f
		}
		parent.Folders = append(parent.Folders, folders[r.ID])
	}
	folderOf := func(r inV4Resource) *PostmanFolder {
		if f, ok := folders[r.ParentID]; ok {
			return f
		}
		return res.Root
	}

	// Pass 2: requests and environments in resource order.
	for _, r := range exp.Resources {
		switch r.Type {
		case "request":
			file, warns := insomniaRequestToFile(&r.inRequestCore, r.Name)
			warnings = append(warnings, warns...)
			if file != nil {
				folderOf(r).Requests = append(folderOf(r).Requests, file)
			}
		case "environment":
			env := buildInsomniaEnvironment(r.Name, r.Data)
			res.Environments = append(res.Environments, env)
			warnings = append(warnings, env.Warnings...)
		}
	}
	return res, warnings, nil
}

// ---- v5: hierarchical YAML collection ----

type inV5Item struct {
	Name           string         `yaml:"name"`
	Method         string         `yaml:"method"`
	URL            string         `yaml:"url"`
	Description    string         `yaml:"description"`
	Headers        []inHeader     `yaml:"headers"`
	Parameters     []inParam      `yaml:"parameters"`
	Body           *inBody        `yaml:"body"`
	Authentication map[string]any `yaml:"authentication"`
	Meta           struct {
		ID string `yaml:"id"`
	} `yaml:"meta"`
	Children []inV5Item `yaml:"children"`
}

type inV5Environment struct {
	Name            string                   `yaml:"name"`
	Data            map[string]any           `yaml:"data"`
	SubEnvironments []map[string]interface{} `yaml:"subEnvironments"`
}

type inV5Doc struct {
	Type          string           `yaml:"type"`
	SchemaVersion string           `yaml:"schema_version"`
	Name          string           `yaml:"name"`
	Collection    *[]inV5Item      `yaml:"collection"`
	Environments  *inV5Environment `yaml:"environments"`
	CookieJar     map[string]any   `yaml:"cookieJar"`
}

func parseInsomniaV5(data []byte) (*InsomniaResult, []string, error) {
	var doc inV5Doc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse Insomnia v5 export: %w", err)
	}
	if doc.Collection == nil {
		return nil, nil, fmt.Errorf("not a valid Insomnia v5 export: missing collection block")
	}
	var warnings []string
	const v5Prefix = "collection.insomnia.rest/"
	if doc.Type != "" && !strings.HasPrefix(doc.Type, v5Prefix) {
		return nil, nil, fmt.Errorf("not a valid Insomnia v5 export: type %q", doc.Type)
	}
	if strings.TrimPrefix(doc.Type, v5Prefix) != "5.0" && doc.Type != "" {
		warnings = append(warnings, fmt.Sprintf("collection type %q differs from %s5.0; attempting best-effort import", doc.Type, v5Prefix))
	}

	res := &InsomniaResult{
		Title:      doc.Name,
		Collection: "insomnia-import",
		Root:       &PostmanFolder{Name: ""},
	}
	collectV5Items(*doc.Collection, res.Root, &warnings)

	// Environments: single object with subEnvironments (v5) or a plain list.
	if doc.Environments != nil {
		env := buildInsomniaEnvironment(doc.Environments.Name, doc.Environments.Data)
		res.Environments = append(res.Environments, env)
		warnings = append(warnings, env.Warnings...)
		for _, sub := range doc.Environments.SubEnvironments {
			name, _ := sub["name"].(string)
			data, _ := sub["data"].(map[string]any)
			subEnv := buildInsomniaEnvironment(name, data)
			res.Environments = append(res.Environments, subEnv)
			warnings = append(warnings, subEnv.Warnings...)
		}
	}
	return res, warnings, nil
}

// collectV5Items walks the v5 tree: items with children become folders,
// items with a URL become requests; both facets are preserved when present.
func collectV5Items(items []inV5Item, dst *PostmanFolder, warnings *[]string) {
	for _, it := range items {
		hasURL := strings.TrimSpace(it.URL) != ""
		switch {
		case hasURL && len(it.Children) == 0:
			file, warns := insomniaRequestToFile(&inRequestCore{
				Name: it.Name, Method: it.Method, URL: it.URL,
				Headers: it.Headers, Parameters: it.Parameters,
				Body: it.Body, Authentication: it.Authentication,
			}, it.Name)
			*warnings = append(*warnings, warns...)
			if file != nil {
				dst.Requests = append(dst.Requests, file)
			}
		case hasURL || len(it.Children) > 0:
			folder := &PostmanFolder{Name: it.Name}
			if hasURL {
				file, warns := insomniaRequestToFile(&inRequestCore{
					Name: it.Name, Method: it.Method, URL: it.URL,
					Headers: it.Headers, Parameters: it.Parameters,
					Body: it.Body, Authentication: it.Authentication,
				}, it.Name)
				*warnings = append(*warnings, warns...)
				if file != nil {
					folder.Requests = append(folder.Requests, file)
				}
			}
			collectV5Items(it.Children, folder, warnings)
			dst.Folders = append(dst.Folders, folder)
		default:
			*warnings = append(*warnings, fmt.Sprintf("item %q has neither requests nor children; skipped", it.Name))
		}
	}
}

// ---- shared request/auth/body/environment conversion ----

// insomniaRequestToFile converts one Insomnia request into a request file.
func insomniaRequestToFile(core *inRequestCore, fallbackName string) (*requestfile.File, []string) {
	var warnings []string
	headers := make([]request.Header, 0, len(core.Headers))
	headerSet := map[string]bool{}
	for _, h := range core.Headers {
		if h.Disabled || h.Name == "" {
			continue
		}
		headers = append(headers, request.Header{Key: h.Name, Value: h.Value})
		headerSet[strings.ToLower(h.Name)] = true
	}
	query := make([]request.Parameter, 0, len(core.Parameters))
	for _, p := range core.Parameters {
		if p.Disabled || p.Name == "" {
			continue
		}
		query = append(query, request.Parameter{Key: p.Name, Value: p.Value})
	}

	body, ct, bodyWarns := convertInsomniaBody(core.Body, headerSet)
	warnings = append(warnings, bodyWarns...)
	if ct != "" && !headerSet["content-type"] {
		headers = append(headers, request.Header{Key: "Content-Type", Value: ct})
	}

	auth, authWarn := convertInsomniaAuth(core.Authentication)
	if authWarn != "" {
		warnings = append(warnings, fmt.Sprintf("%s: %s", fallbackName, authWarn))
	}

	method := strings.ToUpper(strings.TrimSpace(core.Method))
	if method == "" {
		method = "GET"
	}
	name := core.Name
	if name == "" {
		name = fallbackName
	}
	if name == "" {
		name = method + " " + core.URL
	}
	file := &requestfile.File{
		Name: name,
		Request: request.Request{
			Name:    name,
			Method:  request.Method(method),
			URL:     core.URL,
			Headers: headers,
			Query:   query,
			Body:    body,
			Auth:    auth,
		},
	}
	return file, warnings
}

// convertInsomniaBody maps an Insomnia body onto Reqly's wire-body model.
func convertInsomniaBody(b *inBody, headerSet map[string]bool) (body, contentType string, warnings []string) {
	if b == nil {
		return "", "", nil
	}
	hasCT := headerSet["content-type"]
	mt := strings.ToLower(strings.TrimSpace(b.MimeType))
	switch mt {
	case "":
		if b.Text != "" {
			return b.Text, "", nil
		}
		return "", "", nil
	case "application/x-www-form-urlencoded":
		form := url.Values{}
		for _, p := range b.Params {
			if p.Disabled || p.Name == "" {
				continue
			}
			form.Set(p.Name, p.Value)
		}
		ct := ""
		if !hasCT && len(form) > 0 {
			ct = mt
		}
		return form.Encode(), ct, nil
	case "multipart/form-data":
		boundary := fmt.Sprintf("reqlyboundary%d", len(b.Params)*7919+17)
		var sb strings.Builder
		for _, p := range b.Params {
			if p.Disabled || p.Name == "" {
				continue
			}
			if p.Type == "file" {
				src := p.FileName
				if src == "" {
					src = p.Value
				}
				warnings = append(warnings, fmt.Sprintf("form-data file field %q references local path %q and cannot be inlined; re-attach the file after import", p.Name, src))
				continue
			}
			sb.WriteString("--" + boundary + "\r\n")
			sb.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=%q\r\n\r\n", p.Name))
			sb.WriteString(p.Value + "\r\n")
		}
		if sb.Len() == 0 {
			return "", "", warnings
		}
		sb.WriteString("--" + boundary + "--\r\n")
		ct := ""
		if !hasCT {
			ct = "multipart/form-data; boundary=" + boundary
		}
		return sb.String(), ct, warnings
	default:
		ct := ""
		if !hasCT {
			ct = b.MimeType
		}
		return b.Text, ct, nil
	}
}

// convertInsomniaAuth maps an Insomnia authentication block onto Reqly's
// scheme registry. Unmappable types return a warning.
func convertInsomniaAuth(m map[string]any) (request.Auth, string) {
	if len(m) == 0 {
		return request.Auth{}, ""
	}
	cfg := func(keys ...string) map[string]string {
		out := map[string]string{}
		for _, k := range keys {
			if v, ok := m[k]; ok {
				if s := anyToString(v); s != "" {
					out[k] = s
				}
			}
		}
		return out
	}
	t := anyToString(m["type"])
	switch t {
	case "", "none":
		return request.Auth{}, ""
	case "basic":
		return request.Auth{Type: "basic", Config: cfg("username", "password")}, ""
	case "bearer":
		return request.Auth{Type: "bearer", Config: cfg("token")}, ""
	case "apikey":
		c := cfg("key", "value")
		loc := anyToString(m["location"])
		if loc == "" {
			loc = "header"
		}
		c["in"] = loc
		return request.Auth{Type: "apikey", Config: c}, ""
	case "digest":
		return request.Auth{Type: "digest", Config: cfg("username", "password", "realm", "algorithm")}, ""
	default:
		return request.Auth{}, fmt.Sprintf("auth type %q not supported for import; skipped", t)
	}
}

// buildInsomniaEnvironment flattens environment data into variables.
// Nested values are JSON-encoded under dotted keys, warned per occurrence.
func buildInsomniaEnvironment(name string, data map[string]any) InsomniaEnvironment {
	env := InsomniaEnvironment{Name: name, Variables: map[string]string{}}
	var flatten func(prefix string, m map[string]any)
	flatten = func(prefix string, m map[string]any) {
		for k, v := range m {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			switch val := v.(type) {
			case string:
				env.Variables[key] = val
			case float64:
				env.Variables[key] = strconv.FormatFloat(val, 'f', -1, 64)
			case bool:
				env.Variables[key] = strconv.FormatBool(val)
			case map[string]any:
				env.Warnings = append(env.Warnings, fmt.Sprintf("environment %q: nested variable %q flattened to dotted key", name, key))
				flatten(key, val)
			default:
				encoded, _ := json.Marshal(val)
				env.Variables[key] = string(encoded)
			}
		}
	}
	if data != nil {
		flatten("", data)
	}
	return env
}

// anyToString renders a JSON/YAML scalar as a string.
func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		return ""
	}
}

// Write writes the result as a Git-native workspace: reqly.yaml +
// collections/<name>/ plus environments/<env>.yaml files.
func (r *InsomniaResult) Write(dir string) error {
	collection := r.Collection
	if collection == "" {
		collection = "insomnia-import"
	}
	title := r.Title
	if title == "" {
		title = collection
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}
	if err := writeYAMLFile(filepath.Join(dir, "reqly.yaml"), map[string]any{"name": title}); err != nil {
		return err
	}
	collDir := filepath.Join(dir, "collections", sanitizeName(collection))
	if err := os.MkdirAll(collDir, 0o755); err != nil {
		return fmt.Errorf("create collection dir: %w", err)
	}
	if err := writeYAMLFile(filepath.Join(collDir, "reqly.yaml"), map[string]any{"name": collection}); err != nil {
		return err
	}
	if err := writePostmanFolder(r.Root, collDir); err != nil {
		return err
	}
	if len(r.Environments) > 0 {
		envDir := filepath.Join(dir, "environments")
		if err := os.MkdirAll(envDir, 0o755); err != nil {
			return fmt.Errorf("create environments dir: %w", err)
		}
		seen := map[string]int{}
		for _, env := range r.Environments {
			name := env.Name
			if strings.TrimSpace(name) == "" {
				name = "environment"
			}
			payload := struct {
				Variables map[string]string `yaml:"variables,omitempty"`
				Secrets   map[string]string `yaml:"secrets,omitempty"`
			}{Variables: env.Variables, Secrets: env.Secrets}
			fileName := dedupeName(sanitizeName(name), seen)
			if err := writeYAMLFile(filepath.Join(envDir, fileName+".yaml"), payload); err != nil {
				return err
			}
		}
	}
	return nil
}
