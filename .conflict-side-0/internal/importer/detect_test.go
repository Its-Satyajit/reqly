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
	"os"
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Format
		ok   bool
	}{
		// cURL — content sniffing only, no filename involved.
		{"curl simple", []byte("curl https://example.com"), FormatCurl, true},
		{"curl with flags and leading whitespace", []byte("\n  curl -X POST -H 'a: b' -d '{}' https://example.com"), FormatCurl, true},
		{"curl multiline continuation", []byte("curl https://example.com \\\n  -H 'x: y'"), FormatCurl, true},
		{"curl.exe windows", []byte("curl.exe https://example.com"), FormatCurl, true},
		{"bare curl word is not a command", []byte("curl"), "", false},

		// OpenAPI 3.x — JSON and YAML shapes.
		{"openapi json", mustRead(t, filepath.Join("testdata", "import-suite", "openapi", "cli", "fixtures", "openapi.json")), FormatOpenAPI, true},
		{"openapi yaml", mustRead(t, filepath.Join("testdata", "import-suite", "openapi", "fixtures", "openapi-comprehensive.yaml")), FormatOpenAPI, true},
		{"openapi minimal yaml", []byte("openapi: 3.0.0\ninfo:\n  title: t\npaths: {}\n"), FormatOpenAPI, true},

		// HAR — log.entries structure.
		{"har minimal", []byte(`{"log":{"version":"1.2","entries":[{"request":{"method":"GET","url":"https://example.com"}}]}}`), FormatHAR, true},
		{"har pretty-printed", []byte("{\n  \"log\": {\n    \"entries\": []\n  }\n}"), FormatHAR, true},
		{"log without entries is not HAR", []byte(`{"log":{"version":"1.2"}}`), "", false},

		// Postman v2.1 — info.schema marker and envelope shape.
		{"postman collection", mustRead(t, filepath.Join("testdata", "import-suite", "postman", "fixtures", "collection-v2.json")), FormatPostman, true},
		{"postman nested", mustRead(t, filepath.Join("testdata", "import-suite", "postman", "fixtures", "nested-v2-collection.json")), FormatPostman, true},
		{"postman envelope", []byte(`{"collection":{"info":{"schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"}}}`), FormatPostman, true},
		{"info without postman schema is not postman", []byte(`{"info":{"title":"t"},"openapi":"3.0.0"}`), FormatOpenAPI, true},

		// Insomnia — v4 JSON export and v5 YAML collection.
		{"insomnia v4", mustRead(t, filepath.Join("testdata", "import-suite", "insomnia", "fixtures", "insomnia-v4.json")), FormatInsomnia, true},
		{"insomnia v5 yaml", mustRead(t, filepath.Join("testdata", "import-suite", "insomnia", "fixtures", "insomnia-v5.yaml")), FormatInsomnia, true},
		{"insomnia v4 with envs", mustRead(t, filepath.Join("testdata", "import-suite", "insomnia", "fixtures", "insomnia-v4-with-envs.json")), FormatInsomnia, true},

		// Bruno — items tree with name/version.
		{"bruno testbench", mustRead(t, filepath.Join("testdata", "import-suite", "bruno", "fixtures", "bruno-testbench.json")), FormatBruno, true},
		{"bruno descriptions json", mustRead(t, filepath.Join("testdata", "import-suite", "bruno", "fixtures", "descriptions-collection-bru.json")), FormatBruno, true},
		{"bruno minimal", []byte(`{"name":"c","version":"1","items":[]}`), FormatBruno, true},

		// Unknown / malformed — never guess.
		{"empty", nil, "", false},
		{"whitespace only", []byte("  \n\t "), "", false},
		{"plain text", []byte("hello world"), "", false},
		{"garbage bytes", []byte{0x00, 0xff, 0xfe, 0x01}, "", false},
		{"valid json matching no format", []byte(`{"foo":1,"bar":[2,3]}`), "", false},
		{"valid yaml matching no format", []byte("name: thing\nvalue: 42\n"), "", false},
		{"xml is unknown", []byte(`<?xml version="1.0"?><root/>`), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Detect(tt.data)
			if ok != tt.ok {
				t.Fatalf("Detect() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("Detect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return data
}
