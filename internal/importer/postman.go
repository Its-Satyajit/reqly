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
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// PostmanFolder is one level of the imported collection tree: a named folder
// holding nested folders and request files.
type PostmanFolder struct {
	Name     string
	Folders  []*PostmanFolder
	Requests []*requestfile.File
}

// PostmanResult is the parsed Postman collection, ready to be written as a
// Git-native workspace.
type PostmanResult struct {
	Title      string
	Collection string
	Variables  map[string]string
	// Auth is the collection-level auth applied to every request without its
	// own auth block (zero value when the source had none).
	Auth request.Auth
	Root *PostmanFolder
}

// pmCollection mirrors the Postman v2.1 collection JSON.
type pmCollection struct {
	Info *struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Variable []pmVariable      `json:"variable"`
	Auth     *json.RawMessage  `json:"auth"`
	Item     []json.RawMessage `json:"item"`
}

type pmVariable struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

type pmItem struct {
	Name     string            `json:"name"`
	Item     []json.RawMessage `json:"item"`
	Request  json.RawMessage   `json:"request"`
	Event    []json.RawMessage `json:"event"`
	Variable []pmVariable      `json:"variable"`
	Auth     *json.RawMessage  `json:"auth"`
}

type pmRequest struct {
	Method string           `json:"method"`
	URL    json.RawMessage  `json:"url"`
	Header pmHeaders        `json:"header"`
	Body   *pmBody          `json:"body"`
	Auth   *json.RawMessage `json:"auth"`
}

// pmHeaders accepts both header forms: an array of {key, value} objects or
// a raw string of newline-separated "Key: Value" lines.
type pmHeaders []struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

func (h *pmHeaders) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*h = nil
		for _, line := range strings.Split(raw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			k, v, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			*h = append(*h, struct {
				Key      string `json:"key"`
				Value    string `json:"value"`
				Disabled bool   `json:"disabled"`
			}{Key: strings.TrimSpace(k), Value: strings.TrimSpace(v)})
		}
		return nil
	}
	type alias pmHeaders
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*h = pmHeaders(a)
	return nil
}

// pmURL accepts both URL forms: a plain string or a URL object with raw,
// host, path, and query members.
type pmURL struct {
	Raw   string `json:"raw"`
	Query []struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Disabled bool   `json:"disabled"`
	} `json:"query"`
}

func (u *pmURL) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &u.Raw)
	}
	type alias pmURL
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*u = pmURL(a)
	return nil
}

type pmBody struct {
	Mode    string `json:"mode"`
	Raw     string `json:"raw"`
	Options struct {
		Raw struct {
			Language string `json:"language"`
		} `json:"raw"`
	} `json:"options"`
	URL_encoded []pmKeyValue `json:"urlencoded"`
	Formdata    []pmFormData `json:"formdata"`
	GraphQL     struct {
		Query     string `json:"query"`
		Variables string `json:"variables"`
	} `json:"graphql"`
	File struct {
		Src string `json:"src"`
	} `json:"file"`
}

type pmKeyValue struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

type pmFormData struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Type     string `json:"type"` // text | file
	Src      string `json:"src"`
	Disabled bool   `json:"disabled"`
}

type pmAuth struct {
	Type   string       `json:"type"`
	Basic  pmAuthParams `json:"basic"`
	Bearer pmAuthParams `json:"bearer"`
	APIKey pmAuthParams `json:"apikey"`
}

// pmAuthParams accepts both auth parameter forms: a plain object or an array
// of {key, value} rows. Values may be any JSON scalar; non-strings are
// converted to their literal text.
type pmAuthParams map[string]string

func (p *pmAuthParams) UnmarshalJSON(data []byte) error {
	scalar := func(v any) string {
		switch t := v.(type) {
		case string:
			return t
		case float64:
			if t == math.Trunc(t) && math.Abs(t) < 1<<53 {
				return strconv.FormatInt(int64(t), 10)
			}
			return strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(t)
		default:
			return ""
		}
	}
	if len(data) > 0 && data[0] == '[' {
		var rows []struct {
			Key   string `json:"key"`
			Value any    `json:"value"`
		}
		if err := json.Unmarshal(data, &rows); err != nil {
			return err
		}
		*p = map[string]string{}
		for _, r := range rows {
			(*p)[r.Key] = scalar(r.Value)
		}
		return nil
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*p = map[string]string{}
	for k, v := range obj {
		(*p)[k] = scalar(v)
	}
	return nil
}

// ParsePostman parses a Postman collection (v2.x JSON) into a PostmanResult.
// Both export shapes are accepted: the bare collection object and the
// wrapped envelope {"collection": {...}}. Unsupported features (scripts,
// file bodies, unmappable auth) are reported as warnings rather than errors.
func ParsePostman(data []byte) (*PostmanResult, []string, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' {
		return nil, nil, fmt.Errorf("parse Postman collection: not a JSON object")
	}
	// Unwrap the {"collection": {...}} export envelope when present.
	var probe struct {
		Collection *json.RawMessage `json:"collection"`
		Info       *json.RawMessage `json:"info"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, nil, fmt.Errorf("parse Postman collection: %w", err)
	}
	if probe.Collection != nil && probe.Info == nil {
		return parsePostmanCollection(*probe.Collection)
	}
	return parsePostmanCollection(data)
}

func parsePostmanCollection(data []byte) (*PostmanResult, []string, error) {
	var col pmCollection
	if err := json.Unmarshal(data, &col); err != nil {
		return nil, nil, fmt.Errorf("parse Postman collection: %w", err)
	}
	if col.Info == nil {
		return nil, nil, fmt.Errorf("not a valid Postman collection: missing info block")
	}
	res := &PostmanResult{
		Title:      strings.TrimSpace(col.Info.Name),
		Collection: "postman-import",
		Variables:  map[string]string{},
		Root:       &PostmanFolder{Name: ""},
	}
	for _, v := range col.Variable {
		if !v.Disabled && v.Key != "" {
			res.Variables[v.Key] = v.Value
		}
	}
	var warnings []string
	if col.Info.Schema != "" && !strings.Contains(col.Info.Schema, "v2.1") {
		warnings = append(warnings, fmt.Sprintf("unsupported schema %q (want a v2.1 collection)", col.Info.Schema))
	}
	if col.Auth != nil {
		auth, warn := convertAuth(*col.Auth)
		res.Auth = auth
		if warn != "" {
			warnings = append(warnings, warn)
		}
	}
	walkItems(col.Item, res.Root, res, &warnings)
	return res, warnings, nil
}

// walkItems converts an item list into dst, recursing into nested folders.
// An item may carry a request, sub-items, or both (Postman allows a request
// with children); both facets are preserved when present.
func walkItems(items []json.RawMessage, dst *PostmanFolder, res *PostmanResult, warnings *[]string) {
	for _, raw := range items {
		var it pmItem
		if err := json.Unmarshal(raw, &it); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("item skipped: %v", err))
			continue
		}
		hasRequest := len(it.Request) > 0 && string(it.Request) != "null"
		switch {
		case hasRequest && len(it.Item) == 0:
			file, warns := postmanItemToFile(&it, res)
			*warnings = append(*warnings, warns...)
			if file != nil {
				dst.Requests = append(dst.Requests, file)
			}
		case hasRequest || len(it.Item) > 0:
			folder := &PostmanFolder{Name: it.Name}
			if hasRequest {
				file, warns := postmanItemToFile(&it, res)
				*warnings = append(*warnings, warns...)
				if file != nil {
					folder.Requests = append(folder.Requests, file)
				}
			}
			walkItems(it.Item, folder, res, warnings)
			dst.Folders = append(dst.Folders, folder)
		default:
			*warnings = append(*warnings, fmt.Sprintf("item %q has neither requests nor sub-folders; skipped", it.Name))
		}
	}
}

// mustQuote JSON-escapes a plain string into a JSON string literal.
func mustQuote(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

// postmanItemToFile converts one Postman request item into a request file.
func postmanItemToFile(it *pmItem, res *PostmanResult) (*requestfile.File, []string) {
	var warnings []string
	for _, e := range it.Event {
		var ev struct {
			Listen string `json:"listen"`
		}
		_ = json.Unmarshal(e, &ev)
		listen := ev.Listen
		if listen == "" {
			listen = "script"
		}
		warnings = append(warnings, fmt.Sprintf("%s: %s script not imported", it.Name, listen))
	}

	var pr pmRequest
	// v2 shorthand: "request" may be a plain string holding just the URL.
	if err := json.Unmarshal(it.Request, &pr); err != nil {
		var rawURL string
		if err2 := json.Unmarshal(it.Request, &rawURL); err2 == nil {
			pr = pmRequest{Method: "GET", URL: mustQuote(rawURL)}
		} else {
			warnings = append(warnings, fmt.Sprintf("%s: bad request JSON (%v); skipped", it.Name, err))
			return nil, warnings
		}
	}

	method := strings.ToUpper(strings.TrimSpace(pr.Method))
	if method == "" {
		method = "GET"
	}
	reqURL, query := convertURL(pr.URL)
	if reqURL == "" {
		warnings = append(warnings, fmt.Sprintf("%s: request has no URL; skipped", it.Name))
		return nil, warnings
	}
	headers := make([]request.Header, 0, len(pr.Header))
	headerSet := map[string]bool{}
	for _, h := range pr.Header {
		if h.Disabled || h.Key == "" {
			continue
		}
		headers = append(headers, request.Header{Key: h.Key, Value: h.Value})
		headerSet[strings.ToLower(h.Key)] = true
	}

	body, ct, bodyWarns := convertBody(pr.Body, headerSet)
	warnings = append(warnings, bodyWarns...)
	if ct != "" && !headerSet["content-type"] {
		headers = append(headers, request.Header{Key: "Content-Type", Value: ct})
	}

	auth := res.Auth
	if pr.Auth != nil {
		converted, warn := convertAuth(*pr.Auth)
		auth = converted
		if warn != "" {
			warnings = append(warnings, fmt.Sprintf("%s: %s", it.Name, warn))
		}
	}

	vars := map[string]string{}
	for _, v := range it.Variable {
		if !v.Disabled && v.Key != "" {
			vars[v.Key] = v.Value
		}
	}

	name := it.Name
	if name == "" {
		name = method + " " + reqURL
	}
	file := &requestfile.File{
		Name:      name,
		Variables: vars,
		Request: request.Request{
			Name:    name,
			Method:  request.Method(method),
			URL:     reqURL,
			Headers: headers,
			Query:   query,
			Body:    body,
			Auth:    auth,
		},
	}
	return file, warnings
}

// convertURL normalizes both Postman URL forms into Reqly's raw-URL + query
// parameter pair. Disabled query parameters are dropped.
func convertURL(raw json.RawMessage) (string, []request.Parameter) {
	if len(raw) == 0 {
		return "", nil
	}
	var u pmURL
	if err := json.Unmarshal(raw, &u); err != nil {
		return string(raw), nil
	}
	query := make([]request.Parameter, 0, len(u.Query))
	rawQuery := false
	for _, q := range u.Query {
		if q.Disabled || q.Key == "" {
			continue
		}
		query = append(query, request.Parameter{Key: q.Key, Value: q.Value})
		rawQuery = true
	}
	out := u.Raw
	if out == "" {
		// Object form without raw: nothing better than an empty string.
		return "", query
	}
	// Keep {{var}} templates intact: only strip disabled params from a real
	// query string when the object carried explicit query rows.
	if rawQuery || u.Query != nil {
		if i := strings.Index(out, "?"); i >= 0 && len(query) == 0 {
			out = out[:i]
		}
	}
	return out, query
}

// convertBody maps a Postman body onto Reqly's wire-body model. It returns
// the body text and implied Content-Type ("" when none applies).
func convertBody(b *pmBody, headerSet map[string]bool) (body, contentType string, warnings []string) {
	if b == nil {
		return "", "", nil
	}
	hasCT := headerSet["content-type"]
	switch b.Mode {
	case "", "raw":
		ct := ""
		switch b.Options.Raw.Language {
		case "json":
			ct = "application/json"
		case "xml":
			ct = "application/xml"
		case "html":
			ct = "text/html"
		case "javascript":
			ct = "text/javascript"
		}
		return b.Raw, ct, nil
	case "urlencoded":
		form := url.Values{}
		for _, kv := range b.URL_encoded {
			if kv.Disabled || kv.Key == "" {
				continue
			}
			form.Set(kv.Key, kv.Value)
		}
		ct := ""
		if !hasCT {
			ct = "application/x-www-form-urlencoded"
		}
		return form.Encode(), ct, nil
	case "formdata":
		boundary := fmt.Sprintf("reqlyboundary%d", len(b.Formdata)*7919+17)
		var sb strings.Builder
		for _, f := range b.Formdata {
			if f.Disabled || f.Key == "" {
				continue
			}
			if f.Type == "file" {
				warnings = append(warnings, fmt.Sprintf("form-data file field %q references local path %q and cannot be inlined; re-attach the file after import", f.Key, f.Src))
				continue
			}
			sb.WriteString("--" + boundary + "\r\n")
			sb.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=%q\r\n\r\n", f.Key))
			sb.WriteString(f.Value + "\r\n")
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
	case "graphql":
		variables := json.RawMessage(strings.TrimSpace(b.GraphQL.Variables))
		encoded, err := json.Marshal(struct {
			Query     string          `json:"query"`
			Variables json.RawMessage `json:"variables,omitempty"`
		}{Query: b.GraphQL.Query, Variables: variables})
		if err != nil {
			// Invalid variables JSON: send the query alone.
			fallback, _ := json.Marshal(map[string]string{"query": b.GraphQL.Query})
			ct := ""
			if !hasCT {
				ct = "application/json"
			}
			return string(fallback), ct, append(warnings, "graphql variables were not valid JSON; imported the query only")
		}
		ct := ""
		if !hasCT {
			ct = "application/json"
		}
		return string(encoded), ct, warnings
	case "file":
		return "", "", append(warnings, fmt.Sprintf("file mode body (%s) not imported; re-attach the file after import", b.File.Src))
	default:
		return "", "", append(warnings, fmt.Sprintf("unsupported body mode %q; skipped", b.Mode))
	}
}

// convertAuth maps a Postman auth block onto Reqly's scheme registry. Only
// basic, bearer, and apikey are mappable; anything else returns a warning.
func convertAuth(raw json.RawMessage) (request.Auth, string) {
	var a pmAuth
	if err := json.Unmarshal(raw, &a); err != nil {
		return request.Auth{}, fmt.Sprintf("auth block unreadable (%v); skipped", err)
	}
	switch a.Type {
	case "basic":
		return request.Auth{Type: "basic", Config: a.Basic}, ""
	case "bearer":
		return request.Auth{Type: "bearer", Config: a.Bearer}, ""
	case "apikey":
		if a.APIKey["in"] == "" {
			a.APIKey["in"] = "header"
		}
		return request.Auth{Type: "apikey", Config: a.APIKey}, ""
	case "":
		return request.Auth{}, ""
	case "noauth":
		// Postman's explicit "no auth" — maps to Reqly's zero auth silently.
		return request.Auth{}, ""
	default:
		return request.Auth{}, fmt.Sprintf("auth type %q not supported for import; skipped", a.Type)
	}
}

// RequestCount reports the total number of imported requests across all
// folder levels.
func (r *PostmanResult) RequestCount() int { return countRequests(r.Root) }

func countRequests(f *PostmanFolder) int {
	if f == nil {
		return 0
	}
	n := len(f.Requests)
	for _, sub := range f.Folders {
		n += countRequests(sub)
	}
	return n
}

// SanitizeDirName makes a collection title safe for use as an output
// directory name.
func SanitizeDirName(name string) string {
	out := sanitizeName(name)
	if strings.TrimSpace(out) == "" || out == "-" {
		return "postman-import"
	}
	return out
}

// Write writes the result as a Git-native workspace: reqly.yaml +
// collections/<collection>/ with nested folder descriptors.
func (r *PostmanResult) Write(dir string) error {
	collection := r.Collection
	if collection == "" {
		collection = "postman-import"
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
	if len(r.Variables) > 0 {
		cfg["variables"] = r.Variables
	}
	if r.Auth.Type != "" {
		cfg["auth"] = r.Auth
	}
	if err := writeYAMLFile(filepath.Join(collDir, "reqly.yaml"), cfg); err != nil {
		return err
	}
	return writePostmanFolder(r.Root, collDir)
}

// writePostmanFolder recurses a folder tree to disk: each child folder gets
// its own directory + descriptor; requests are written as YAML files with
// per-directory filename dedupe.
func writePostmanFolder(folder *PostmanFolder, dir string) error {
	seen := map[string]int{}
	for _, f := range folder.Requests {
		name := dedupeName(sanitizeName(f.Name), seen)
		path := filepath.Join(dir, name+".yaml")
		if err := writeYAMLFile(path, f); err != nil {
			return err
		}
	}
	for _, sub := range folder.Folders {
		subName := sub.Name
		if strings.TrimSpace(subName) == "" {
			subName = "folder"
		}
		subDir := filepath.Join(dir, dedupeName(sanitizeName(subName), seen))
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			return fmt.Errorf("create folder dir: %w", err)
		}
		if err := writeYAMLFile(filepath.Join(subDir, "reqly.yaml"), map[string]any{"name": sub.Name}); err != nil {
			return err
		}
		if err := writePostmanFolder(sub, subDir); err != nil {
			return err
		}
	}
	return nil
}

// dedupeName appends -2, -3, ... on collision within a directory scope.
func dedupeName(name string, seen map[string]int) string {
	if name == "" {
		name = "request"
	}
	seen[name]++
	if n := seen[name]; n > 1 {
		return fmt.Sprintf("%s-%d", name, n)
	}
	return name
}
