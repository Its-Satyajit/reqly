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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// HARResult is the in-memory result of importing a HAR file.
type HARResult struct {
	Title      string
	Collection string
	Requests   []*requestfile.File
	Warnings   []string
}

// harFile mirrors HAR 1.2 log structure.
type harFile struct {
	Log harLog `json:"log"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Pages   []harPage  `json:"pages"`
	Entries []harEntry `json:"entries"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harPage struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type harEntry struct {
	Pageref         string      `json:"pageref"`
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         harRequest  `json:"request"`
	Response        harResponse `json:"response"`
	Cache           any         `json:"cache"`
	Timings         any         `json:"timings"`
	ServerIPAddress string      `json:"serverIPAddress"`
	Connection      string      `json:"connection"`
	Comment         string      `json:"comment"`
}

type harRequest struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []harNVP     `json:"headers"`
	Cookies     []harCookie  `json:"cookies"`
	QueryString []harNVP     `json:"queryString"`
	PostData    *harPostData `json:"postData"`
	HeadersSize int          `json:"headersSize"`
	BodySize    int          `json:"bodySize"`
	Comment     string       `json:"comment"`
}

type harResponse struct {
	Status      int        `json:"status"`
	StatusText  string     `json:"statusText"`
	Headers     []harNVP   `json:"headers"`
	Content     harContent `json:"content"`
	RedirectURL string     `json:"redirectURL"`
	HeadersSize int        `json:"headersSize"`
	BodySize    int        `json:"bodySize"`
}

type harContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
	Encoding string `json:"encoding"`
	Comment  string `json:"comment"`
}

type harNVP struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	Comment  string `json:"comment"`
}

type harPostData struct {
	MimeType string         `json:"mimeType"`
	Text     string         `json:"text"`
	Encoding string         `json:"encoding"`
	Params   []harPostParam `json:"params"`
	Comment  string         `json:"comment"`
}

type harPostParam struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Comment     string `json:"comment"`
}

// ParseHAR parses HAR JSON 1.2. It returns a HARResult plus warnings for dropped fields.
func ParseHAR(data []byte) (*HARResult, []string, error) {
	var f harFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, nil, fmt.Errorf("parse HAR: %w", err)
	}
	if f.Log.Entries == nil {
		return nil, nil, fmt.Errorf("not a HAR file (missing log.entries)")
	}
	var warnings []string
	if f.Log.Version != "" && f.Log.Version != "1.2" {
		warnings = append(warnings, fmt.Sprintf("unsupported HAR version %q: expected 1.2, parsing as 1.2", f.Log.Version))
	}
	if len(f.Log.Pages) > 0 {
		warnings = append(warnings, fmt.Sprintf("dropped %d pageref pages (grouping deferred to M28b)", len(f.Log.Pages)))
	}
	result := &HARResult{
		Title:      "har-import",
		Collection: "har-import",
	}
	seen := map[string]int{}
	for i, e := range f.Log.Entries {
		if e.Pageref != "" {
			// warn once per entry with pageref
			warnings = append(warnings, fmt.Sprintf("entry %d: pageref %q dropped (M28b)", i, e.Pageref))
		}
		if e.Cache != nil {
			warnings = append(warnings, fmt.Sprintf("entry %d: cache dropped", i))
		}
		if e.Timings != nil {
			warnings = append(warnings, fmt.Sprintf("entry %d: timings dropped (re-synthesized on export)", i))
		}
		if e.ServerIPAddress != "" || e.Connection != "" {
			warnings = append(warnings, fmt.Sprintf("entry %d: serverIPAddress/connection dropped", i))
		}
		// response is intentionally discarded for M28
		req, w, err := harEntryToRequest(e, i, seen)
		warnings = append(warnings, w...)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("entry %d skipped: %v", i, err))
			continue
		}
		if req == nil {
			continue
		}
		result.Requests = append(result.Requests, req)
	}
	if len(result.Requests) == 0 && len(f.Log.Entries) > 0 {
		warnings = append(warnings, "no requests imported (all entries skipped)")
	}
	result.Warnings = warnings
	return result, warnings, nil
}

func harEntryToRequest(e harEntry, idx int, seen map[string]int) (*requestfile.File, []string, error) {
	var warnings []string
	if e.Request.URL == "" {
		return nil, warnings, fmt.Errorf("missing request.url")
	}
	// headers + cookies → Headers
	var headers []request.Header
	headerMap := map[string]bool{}
	for _, h := range e.Request.Headers {
		if h.Name == "" {
			continue
		}
		headers = append(headers, request.Header{Key: h.Name, Value: h.Value})
		headerMap[strings.ToLower(h.Name)] = true
	}
	// cookies merged as Cookie header
	if len(e.Request.Cookies) > 0 {
		var parts []string
		for _, c := range e.Request.Cookies {
			if c.Name == "" {
				continue
			}
			parts = append(parts, c.Name+"="+c.Value)
		}
		if len(parts) > 0 {
			cookieVal := strings.Join(parts, "; ")
			headers = append(headers, request.Header{Key: "Cookie", Value: cookieVal})
		}
	}
	// queryString → Query
	var query []request.Parameter
	for _, q := range e.Request.QueryString {
		if q.Name == "" {
			continue
		}
		query = append(query, request.Parameter{Key: q.Name, Value: q.Value})
	}
	// fallback: parse URL query if queryString empty but URL has query
	if len(query) == 0 && strings.Contains(e.Request.URL, "?") {
		if u, err := url.Parse(e.Request.URL); err == nil {
			for k, vals := range u.Query() {
				for _, v := range vals {
					query = append(query, request.Parameter{Key: k, Value: v})
				}
			}
		}
	}
	// postData → Body
	var body string
	if e.Request.PostData != nil {
		pd := e.Request.PostData
		if len(pd.Params) > 0 {
			warnings = append(warnings, fmt.Sprintf("entry %d: postData.params dropped (%d params, M28 handles text only)", idx, len(pd.Params)))
		}
		text := pd.Text
		if pd.Encoding == "base64" && text != "" {
			decoded, err := base64.StdEncoding.DecodeString(text)
			if err != nil {
				// try StdEncoding with padding handling via RawStdEncoding?
				decoded2, err2 := base64.RawStdEncoding.DecodeString(text)
				if err2 != nil {
					return nil, warnings, fmt.Errorf("postData base64 decode failed: %w", err)
				}
				body = string(decoded2)
			} else {
				body = string(decoded)
			}
		} else {
			body = text
		}
		// binary spill >1MB
		if len(body) > 1024*1024 {
			// will be handled as file ref in Write; for now keep body as is but warn
			warnings = append(warnings, fmt.Sprintf("entry %d: body >1MB will be written as file reference", idx))
		}
		// mimeType → Content-Type header if not already present
		if pd.MimeType != "" && !headerMap["content-type"] {
			headers = append(headers, request.Header{Key: "Content-Type", Value: pd.MimeType})
		}
	}
	method := e.Request.Method
	if method == "" {
		method = "GET"
	}
	// name/filename from method+host+path, deduped
	filename := harFilename(method, e.Request.URL, seen)
	name := harDisplayName(method, e.Request.URL)
	req := request.Request{
		Name:    name,
		Method:  request.Method(strings.ToUpper(method)),
		URL:     e.Request.URL,
		Headers: headers,
		Query:   query,
		Body:    body,
	}
	file := &requestfile.File{
		Name:    filename,
		Request: req,
	}
	return file, warnings, nil
}

func harFilename(method, rawURL string, seen map[string]int) string {
	u, err := url.Parse(rawURL)
	host := ""
	path := rawURL
	if err == nil && u.Host != "" {
		host = u.Host
		path = u.Path
		if path == "" {
			path = "/"
		}
	}
	// sanitize host+path
	s := strings.ToLower(method) + "-" + host + "-" + path
	s = strings.Trim(s, "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, "?", "-")
	s = strings.ReplaceAll(s, "&", "-")
	s = strings.ReplaceAll(s, "=", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = strings.ToLower(method) + "-request"
	}
	s = sanitizeHARName(s)
	base := s
	if n, ok := seen[base]; ok {
		seen[base] = n + 1
		s = fmt.Sprintf("%s-%d", base, n+1)
	} else {
		seen[base] = 1
	}
	return s
}

func harDisplayName(method, rawURL string) string {
	return strings.ToUpper(method) + " " + rawURL
}

func sanitizeHARName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, " ", "-")
	// collapse duplicate dashes
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}

// Write writes the HARResult as a Git-native workspace: reqly.yaml + collections/<collection>/<request>.yaml
func (r *HARResult) Write(dir string, collectionName string) error {
	if collectionName == "" {
		collectionName = r.Collection
		if collectionName == "" {
			collectionName = "har-import"
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}
	wsCfg := map[string]any{"name": r.Title}
	if err := writeHARFile(filepath.Join(dir, "reqly.yaml"), wsCfg); err != nil {
		return err
	}
	collDir := filepath.Join(dir, "collections", sanitizeHARName(collectionName))
	if err := os.MkdirAll(collDir, 0o755); err != nil {
		return fmt.Errorf("create collection dir: %w", err)
	}
	collCfg := map[string]any{"name": collectionName}
	if err := writeHARFile(filepath.Join(collDir, "reqly.yaml"), collCfg); err != nil {
		return err
	}
	for _, f := range r.Requests {
		path := filepath.Join(collDir, sanitizeHARName(f.Name)+".yaml")
		if len(f.Request.Body) > 1024*1024 {
			blobDir := filepath.Join(dir, ".reqly", "blobs")
			_ = os.MkdirAll(blobDir, 0o755)
			blobPath := filepath.Join(blobDir, sanitizeHARName(f.Name)+".bin")
			_ = os.WriteFile(blobPath, []byte(f.Request.Body), 0o600)
		}
		if err := writeHARFile(path, f); err != nil {
			return err
		}
	}
	return nil
}

func writeHARFile(path string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
