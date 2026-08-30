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
	"strings"

	"go.yaml.in/yaml/v3"
)

// Format identifies which import format a payload represents, as determined
// by Detect from content alone.
type Format string

const (
	FormatCurl     Format = "curl"
	FormatOpenAPI  Format = "openapi"
	FormatHAR      Format = "har"
	FormatPostman  Format = "postman"
	FormatInsomnia Format = "insomnia"
	FormatBruno    Format = "bruno"
)

// Detect sniffs the content of data and reports which supported import format
// it represents. Detection is advisory: callers surface an unknown payload as
// "not detected" rather than guessing, and a format hint from the user takes
// precedence over detection. Filename and extension are deliberately not
// inputs — the same bytes must detect identically regardless of how they
// arrived (file picker, drag-and-drop, or pasted text).
func Detect(data []byte) (Format, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", false
	}
	if f, ok := detectCurl(trimmed); ok {
		return f, true
	}
	if f, ok := detectJSON(trimmed); ok {
		return f, true
	}
	return detectYAML(trimmed)
}

// detectCurl recognizes a curl invocation: the trimmed payload starts with
// the curl (or curl.exe) command word followed by an argument — a bare
// mention of "curl" with no arguments is not a command.
func detectCurl(s []byte) (Format, bool) {
	for _, exe := range []string{"curl", "curl.exe"} {
		if !bytes.HasPrefix(s, []byte(exe)) {
			continue
		}
		rest := s[len(exe):]
		if len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\n' || rest[0] == '\r') {
			return FormatCurl, true
		}
	}
	return "", false
}

// detectJSON classifies JSON payloads by their discriminating top-level keys.
// Checks run most-specific first so overlapping shapes resolve correctly:
// an OpenAPI document also carries "info", so the Postman check requires the
// Postman schema marker rather than info alone.
func detectJSON(data []byte) (Format, bool) {
	var top map[string]json.RawMessage
	if json.Unmarshal(data, &top) != nil {
		return "", false
	}

	if _, ok := top["__export_format"]; ok {
		return FormatInsomnia, true
	}
	if _, ok := top["collection"]; ok {
		return FormatPostman, true
	}
	if logRaw, ok := top["log"]; ok {
		var logObj map[string]json.RawMessage
		if json.Unmarshal(logRaw, &logObj) == nil {
			if _, ok := logObj["entries"]; ok {
				return FormatHAR, true
			}
		}
	}
	var probe struct {
		Info struct {
			Schema        string `json:"schema"`
			PostmanSchema string `json:"_postman_schema"`
		} `json:"info"`
		OpenAPI any    `json:"openapi"`
		Items   []any  `json:"items"`
		Name    string `json:"name"`
		Version any    `json:"version"`
	}
	if json.Unmarshal(data, &probe) != nil {
		return "", false
	}
	if strings.Contains(probe.Info.Schema, "postman") || strings.Contains(probe.Info.PostmanSchema, "postman") {
		return FormatPostman, true
	}
	if probe.OpenAPI != nil {
		return FormatOpenAPI, true
	}
	if probe.Items != nil && (probe.Name != "" || probe.Version != nil) {
		return FormatBruno, true
	}
	return "", false
}

// detectYAML classifies YAML payloads by the keys only the v5 Insomnia export
// and OpenAPI documents carry.
func detectYAML(data []byte) (Format, bool) {
	var probe struct {
		Type    string `yaml:"type"`
		OpenAPI string `yaml:"openapi"`
	}
	if yaml.Unmarshal(data, &probe) != nil {
		return "", false
	}
	if strings.HasPrefix(probe.Type, "collection.insomnia.rest/") {
		return FormatInsomnia, true
	}
	if strings.TrimSpace(probe.OpenAPI) != "" {
		return FormatOpenAPI, true
	}
	return "", false
}
