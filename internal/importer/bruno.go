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
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// BrunoResult is a parsed Bruno collection export, ready to be written as a
// Git-native workspace.
type BrunoResult struct {
	Title        string
	Collection   string
	Auth         request.Auth
	Headers      []request.Header
	Root         *PostmanFolder
	Environments []InsomniaEnvironment
}

// RequestCount reports the total number of imported requests.
func (r *BrunoResult) RequestCount() int { return countRequests(r.Root) }

// ---- shapes ----

type brKV struct {
	Name    string `json:"name"`
	Value   any    `json:"value"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	File    string `json:"file"`
}

func (kv brKV) str() string { return anyToString(kv.Value) }

type brBody struct {
	Mode    string `json:"mode"`
	JSON    string `json:"json"`
	XML     string `json:"xml"`
	Text    string `json:"text"`
	Sparql  string `json:"sparql"`
	GraphQL *struct {
		Query     string `json:"query"`
		Variables string `json:"variables"`
	} `json:"graphql"`
	FormUrlEncoded []brKV `json:"formUrlEncoded"`
	MultipartForm  []brKV `json:"multipartForm"`
	File           []brKV `json:"file"`
}

type brAuth struct {
	Mode   string         `json:"mode"`
	Config map[string]any `json:"-"`
}

func (a *brAuth) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	_ = json.Unmarshal(raw["mode"], &a.Mode)
	if a.Mode != "" && a.Mode != "none" {
		if cfg, ok := raw[a.Mode]; ok {
			return json.Unmarshal(cfg, &a.Config)
		}
	}
	a.Config = nil
	return nil
}

type brItem struct {
	Type  string   `json:"type"`
	Name  string   `json:"name"`
	Seq   int      `json:"seq"`
	Items []brItem `json:"items"`
	// Request is decoded lazily so items with unknown/garbage types can be
	// skipped with a warning instead of failing the whole import.
	Request json.RawMessage `json:"request"`
}

// brKVList keeps parsing when individual header/param rows are malformed.
type brKVList []brKV

func (l *brKVList) UnmarshalJSON(data []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil
	}
	for _, r := range raws {
		var kv brKV
		if json.Unmarshal(r, &kv) == nil {
			*l = append(*l, kv)
		}
	}
	return nil
}

type brRequest struct {
	URL        string          `json:"url"`
	Method     string          `json:"method"`
	Headers    brKVList        `json:"headers"`
	Params     brKVList        `json:"params"`
	Body       *brBody         `json:"body"`
	Auth       *brAuth         `json:"auth"`
	Script     json.RawMessage `json:"script"`
	Assertions json.RawMessage `json:"assertions"`
	Tests      json.RawMessage `json:"tests"`
	Docs       json.RawMessage `json:"docs"`
}

type brEnvVar struct {
	Name    string `json:"name"`
	Value   any    `json:"value"`
	Enabled bool   `json:"enabled"`
	Secret  bool   `json:"secret"`
}

type brEnvironment struct {
	Name      string     `json:"name"`
	Variables []brEnvVar `json:"variables"`
}

type brCollection struct {
	Name         string         `json:"name"`
	Items        []brItem       `json:"items"`
	Environments brTolerantEnvs `json:"environments"`
	Root         *struct {
		Request *struct {
			Auth    *brAuth  `json:"auth"`
			Headers brKVList `json:"headers"`
		} `json:"request"`
	} `json:"root"`
}

// brTolerantEnvs keeps parsing when an environment entry is malformed: good
// entries are kept, broken ones are dropped silently.
type brTolerantEnvs []brEnvironment

func (e *brTolerantEnvs) UnmarshalJSON(data []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil
	}
	for _, r := range raws {
		var env brEnvironment
		if json.Unmarshal(r, &env) == nil && (env.Name != "" || len(env.Variables) > 0) {
			*e = append(*e, env)
		}
	}
	return nil
}

// ParseBruno parses a Bruno collection export JSON into a BrunoResult.
func ParseBruno(data []byte) (*BrunoResult, *ImportReport, error) {
	var col brCollection
	if err := json.Unmarshal(data, &col); err != nil {
		return nil, nil, fmt.Errorf("parse Bruno collection: %w", err)
	}
	res := &BrunoResult{
		Title:      strings.TrimSpace(col.Name),
		Collection: "bruno-import",
		Root:       &PostmanFolder{Name: ""},
	}

	rep := NewReport("bruno")
	if col.Root != nil && col.Root.Request != nil {
		if col.Root.Request.Auth != nil {
			auth, warn := convertBrunoAuth(col.Root.Request.Auth)
			res.Auth = auth
			if warn != "" {
				rep.Add("", CategoryAuth, SeverityWarned, "%s", warn)
			}
		}
		for _, h := range col.Root.Request.Headers {
			if h.Enabled && h.Name != "" {
				res.Headers = append(res.Headers, request.Header{Key: h.Name, Value: h.str()})
			}
		}
	}
	collectBrunoItems(col.Items, res.Root, res, rep)
	for _, env := range col.Environments {
		res.Environments = append(res.Environments, buildBrunoEnvironment(env))
	}
	return res, rep, nil
}

// collectBrunoItems walks the items tree; unknown item types warn+skip.
func collectBrunoItems(items []brItem, dst *PostmanFolder, res *BrunoResult, rep *ImportReport) {
	for _, it := range items {
		switch it.Type {
		case "folder":
			folder := &PostmanFolder{Name: it.Name}
			collectBrunoItems(it.Items, folder, res, rep)
			dst.Folders = append(dst.Folders, folder)
		case "http", "graphql":
			file := brunoItemToFile(&it, res, rep)
			if file != nil {
				dst.Requests = append(dst.Requests, file)
			}
		default:
			name := it.Name
			if name == "" {
				name = "(unnamed)"
			}
			rep.Add(name, CategoryOther, SeverityDropped, "item %q has unsupported type %q; skipped", name, it.Type)
		}
	}
}

// brunoItemToFile converts one item into a request file, recording
// degradations on rep.
func brunoItemToFile(it *brItem, res *BrunoResult, rep *ImportReport) *requestfile.File {
	var req *brRequest
	if len(it.Request) > 0 && string(it.Request) != "null" {
		if err := json.Unmarshal(it.Request, &req); err != nil {
			rep.Add(it.Name, CategoryOther, SeverityDropped, "%s: request block unreadable (%v); skipped", it.Name, err)
			return nil
		}
	}
	if req == nil {
		rep.Add(it.Name, CategoryOther, SeverityDropped, "%s: no request block; skipped", it.Name)
		return nil
	}
	name := it.Name

	for _, kind := range []struct {
		key   string
		label string
	}{
		{"script", "script"}, {"assertions", "assertions"},
		{"tests", "tests"}, {"docs", "docs"},
	} {
		if len(nonEmptyMember(req, kind.key)) > 0 {
			rep.Add(name, CategoryScript, SeverityDropped, "%s: %s not imported", name, kind.label)
		}
	}

	headers := make([]request.Header, 0, len(req.Headers)+len(res.Headers))
	headerSet := map[string]bool{}
	appendHeader := func(key, value string) {
		headers = append(headers, request.Header{Key: key, Value: value})
		headerSet[strings.ToLower(key)] = true
	}
	for _, r := range req.Headers {
		if !r.Enabled || r.Name == "" {
			continue
		}
		appendHeader(r.Name, r.str())
	}
	for _, h := range res.Headers {
		if !headerSet[strings.ToLower(h.Key)] {
			appendHeader(h.Key, h.Value)
		}
	}
	query := make([]request.Parameter, 0, len(req.Params))
	for _, p := range req.Params {
		if !p.Enabled || p.Name == "" {
			continue
		}
		query = append(query, request.Parameter{Key: p.Name, Value: p.str()})
	}

	body, ct, bodyWarns := convertBrunoBody(req.Body, headerSet)
	rep.AddAll(name, CategoryBody, SeverityWarned, bodyWarns)
	if ct != "" && !headerSet["content-type"] {
		appendHeader("Content-Type", ct)
	}

	auth := res.Auth
	if req.Auth != nil && req.Auth.Mode != "" && req.Auth.Mode != "none" {
		converted, warn := convertBrunoAuth(req.Auth)
		if warn != "" {
			rep.Add(name, CategoryAuth, SeverityWarned, "%s: %s", name, warn)
			auth = request.Auth{}
		} else {
			auth = converted
		}
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}
	reqURL := req.URL
	if strings.TrimSpace(reqURL) == "" {
		rep.Add(name, CategoryOther, SeverityDropped, "%s: request has no URL; skipped", name)
		return nil
	}
	display := name
	if display == "" {
		display = method + " " + reqURL
	}
	file := &requestfile.File{
		Name: display,
		Request: request.Request{
			Name:    display,
			Method:  request.Method(method),
			URL:     reqURL,
			Headers: headers,
			Query:   query,
			Body:    body,
			Auth:    auth,
		},
	}
	return file
}

// nonEmptyMember reports whether a request member carries content worth
// warning about (empty objects/strings are silent).
func nonEmptyMember(req *brRequest, key string) json.RawMessage {
	switch key {
	case "script":
		var m map[string]json.RawMessage
		if err := json.Unmarshal(req.Script, &m); err != nil {
			return nil
		}
		for _, v := range m {
			s := strings.TrimSpace(string(v))
			if s != "" && s != `""` && s != "{}" && s != "[]" && s != "null" {
				return v
			}
		}
		return nil
	case "assertions":
		var list []json.RawMessage
		if err := json.Unmarshal(req.Assertions, &list); err == nil && len(list) > 0 {
			return req.Assertions
		}
		return nil
	default:
		var s string
		raw := map[string]json.RawMessage{"tests": req.Tests, "docs": req.Docs}[key]
		if len(raw) == 0 {
			return nil
		}
		if err := json.Unmarshal(raw, &s); err == nil && strings.TrimSpace(s) != "" {
			return raw
		}
		return nil
	}
}

// convertBrunoBody maps a Bruno body onto Reqly's wire-body model.
func convertBrunoBody(b *brBody, headerSet map[string]bool) (body, contentType string, warnings []string) {
	if b == nil {
		return "", "", nil
	}
	hasCT := headerSet["content-type"]
	textWithCT := func(text, ct string) (string, string, []string) {
		if !hasCT {
			return text, ct, nil
		}
		return text, "", nil
	}
	switch b.Mode {
	case "":
		return "", "", nil
	case "json":
		if b.JSON == "" {
			return "", "", nil
		}
		return textWithCT(b.JSON, "application/json")
	case "xml":
		if b.XML == "" {
			return "", "", nil
		}
		return textWithCT(b.XML, "application/xml")
	case "text", "plaintext":
		return b.Text, "", nil
	case "sparql":
		return b.Sparql, "", nil
	case "graphql":
		if b.GraphQL == nil || b.GraphQL.Query == "" {
			return "", "", nil
		}
		encoded, err := json.Marshal(struct {
			Query     string          `json:"query"`
			Variables json.RawMessage `json:"variables,omitempty"`
		}{Query: b.GraphQL.Query, Variables: json.RawMessage(strings.TrimSpace(b.GraphQL.Variables))})
		if err != nil {
			return b.GraphQL.Query, "", append(warnings, "graphql variables were not valid JSON; imported the query only")
		}
		return textWithCT(string(encoded), "application/json")
	case "formUrlEncoded":
		form := url.Values{}
		for _, kv := range b.FormUrlEncoded {
			if !kv.Enabled || kv.Name == "" {
				continue
			}
			form.Set(kv.Name, kv.str())
		}
		ct := ""
		if !hasCT && len(form) > 0 {
			ct = "application/x-www-form-urlencoded"
		}
		return form.Encode(), ct, nil
	case "multipartForm":
		boundary := fmt.Sprintf("reqlyboundary%d", len(b.MultipartForm)*7919+17)
		var sb strings.Builder
		for _, kv := range b.MultipartForm {
			if !kv.Enabled || kv.Name == "" {
				continue
			}
			if kv.Type == "file" {
				src := kv.File
				if src == "" {
					src = kv.str()
				}
				warnings = append(warnings, fmt.Sprintf("multipart file field %q references local path %q and cannot be inlined; re-attach the file after import", kv.Name, src))
				continue
			}
			sb.WriteString("--" + boundary + "\r\n")
			sb.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=%q\r\n\r\n", kv.Name))
			sb.WriteString(kv.str() + "\r\n")
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
	case "file":
		names := make([]string, 0, len(b.File))
		for _, f := range b.File {
			names = append(names, f.File)
		}
		return "", "", append(warnings, fmt.Sprintf("file-mode body (%s) not imported; re-attach the files after import", strings.Join(names, ", ")))
	default:
		return "", "", append(warnings, fmt.Sprintf("unsupported body mode %q; skipped", b.Mode))
	}
}

// convertBrunoAuth maps a Bruno auth block onto Reqly's scheme registry.
func convertBrunoAuth(a *brAuth) (request.Auth, string) {
	cfgOf := func(keys ...string) map[string]string {
		out := map[string]string{}
		for _, k := range keys {
			if v, ok := a.Config[k]; ok {
				if s := anyToString(v); s != "" {
					out[k] = s
				}
			}
		}
		return out
	}
	switch a.Mode {
	case "", "none":
		return request.Auth{}, ""
	case "basic":
		return request.Auth{Type: "basic", Config: cfgOf("username", "password")}, ""
	case "bearer":
		return request.Auth{Type: "bearer", Config: cfgOf("token")}, ""
	case "apikey":
		c := cfgOf("key", "value")
		loc := anyToString(a.Config["placement"])
		switch loc {
		case "queryparams":
			c["in"] = "query"
		default:
			c["in"] = "header"
		}
		return request.Auth{Type: "apikey", Config: c}, ""
	case "digest":
		return request.Auth{Type: "digest", Config: cfgOf("username", "password", "realm", "algorithm")}, ""
	default:
		return request.Auth{}, fmt.Sprintf("auth type %q not supported for import; skipped", a.Mode)
	}
}

// buildBrunoEnvironment splits env variables into variables and secrets.
func buildBrunoEnvironment(env brEnvironment) InsomniaEnvironment {
	out := InsomniaEnvironment{Name: env.Name, Variables: map[string]string{}, Secrets: map[string]string{}}
	for _, v := range env.Variables {
		if !v.Enabled || v.Name == "" {
			continue
		}
		if v.Secret {
			out.Secrets[v.Name] = anyToString(v.Value)
		} else {
			out.Variables[v.Name] = anyToString(v.Value)
		}
	}
	return out
}

// Write writes the result as a Git-native workspace with environments.
func (r *BrunoResult) Write(dir string) error {
	collection := r.Collection
	if collection == "" {
		collection = "bruno-import"
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
	cfg := map[string]any{"name": collection}
	if r.Auth.Type != "" {
		cfg["auth"] = r.Auth
	}
	if len(r.Headers) > 0 {
		cfg["headers"] = r.Headers
	}
	if err := writeYAMLFile(filepath.Join(collDir, "reqly.yaml"), cfg); err != nil {
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
