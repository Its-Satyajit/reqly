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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/importer"
	reqlyopenapi "github.com/Its-Satyajit/reqly/internal/openapi"
	"github.com/Its-Satyajit/reqly/internal/request"
)

// ImportKindWorkspace marks an import that materializes as a workspace
// folder (HAR, Postman, Insomnia, Bruno).
const ImportKindWorkspace = "workspace"

// ImportKindRequest marks an import that yields a single parsed request
// (cURL) for the frontend to open as an unsaved tab — nothing is written.
const ImportKindRequest = "request"

// ImportRequest is the generic import ask. Content carries the raw payload
// (file bytes or a pasted command); FormatHint overrides content-based
// detection when non-empty; DryRun parses and previews without touching the
// filesystem; TargetDir names the workspace folder created on commit
// (empty → sanitized collection title).
type ImportRequest struct {
	Content    string `json:"content"`
	Filename   string `json:"filename,omitempty"`
	FormatHint string `json:"formatHint,omitempty"`
	DryRun     bool   `json:"dryRun"`
	TargetDir  string `json:"targetDir,omitempty"`
}

// ImportResult reports what an import produced or would produce. Report is
// the structured M42 degradation record, rendered by the frontend grouped by
// category with severity tallies.
type ImportResult struct {
	Kind             string                 `json:"kind"`
	Format           importer.Format        `json:"format"`
	Title            string                 `json:"title,omitempty"`
	RequestCount     int                    `json:"requestCount"`
	EnvironmentCount int                    `json:"environmentCount,omitempty"`
	TargetDir        string                 `json:"targetDir,omitempty"`
	Report           *importer.ImportReport `json:"report,omitempty"`
	// Operations carries the flattened OpenAPI operation list on dry-run so
	// the dialog can render a tag-grouped preview before commit.
	Operations []reqlyopenapi.Endpoint `json:"operations,omitempty"`
	// Request carries the parsed request for cURL imports (Kind "request").
	Request *request.Request `json:"request,omitempty"`

	// write materializes the parsed result at dir on commit; set per format
	// by the dispatch above and never serialized to the frontend.
	write func(dir string) error
}

// Import parses and optionally commits an import through the shared importer
// core — identical fidelity to `reqly import`. Detection runs on content
// alone; a valid FormatHint wins over detection. Dry-run never touches the
// filesystem. Commit writes a fresh folder under the workspace root named by
// TargetDir (sanitized) or the collection title, and fails fast when it
// already exists — imports never silently merge into an existing workspace.
func (s *AppService) Import(req ImportRequest) (*ImportResult, error) {
	if s == nil {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to import")
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("nothing to import: drop a file or paste a command first")
	}
	format, err := resolveFormat(req.FormatHint, []byte(req.Content))
	if err != nil {
		return nil, err
	}
	return s.importFormat(format, req)
}

// Detect sniffs content and reports which import format it represents, so
// the dialog can badge the detected format as the user types or drops a
// file. Advisory only — Import's FormatHint overrides it.
func (s *AppService) Detect(content string) (string, bool, error) {
	format, ok := importer.Detect([]byte(content))
	return string(format), ok, nil
}

// resolveFormat applies the hint override or falls back to content sniffing.
func resolveFormat(hint string, content []byte) (importer.Format, error) {
	if hint != "" {
		f := importer.Format(hint)
		switch f {
		case importer.FormatCurl, importer.FormatOpenAPI, importer.FormatHAR,
			importer.FormatPostman, importer.FormatInsomnia, importer.FormatBruno:
			return f, nil
		default:
			return "", fmt.Errorf("unknown format %q: pick one of curl, openapi, har, postman, insomnia, bruno", hint)
		}
	}
	if f, ok := importer.Detect(content); ok {
		return f, nil
	}
	return "", fmt.Errorf("could not detect the format: choose one manually from the list")
}

// importFormat dispatches to the matching parser and, for commit requests,
// writes the result under the workspace root.
func (s *AppService) importFormat(format importer.Format, req ImportRequest) (*ImportResult, error) {
	var res *ImportResult

	switch format {
	case importer.FormatPostman:
		parsed, report, err := importer.ParsePostman([]byte(req.Content))
		if err != nil {
			return nil, err
		}
		res = &ImportResult{
			Title:        parsed.Title,
			RequestCount: parsed.RequestCount(),
			Report:       report,
		}
		res.write = func(dir string) error { return parsed.Write(dir) }

	case importer.FormatInsomnia:
		parsed, report, err := importer.ParseInsomnia([]byte(req.Content))
		if err != nil {
			return nil, err
		}
		res = &ImportResult{
			Title:            parsed.Title,
			RequestCount:     parsed.RequestCount(),
			EnvironmentCount: len(parsed.Environments),
			Report:           report,
		}
		res.write = func(dir string) error { return parsed.Write(dir) }

	case importer.FormatBruno:
		parsed, report, err := importer.ParseBruno([]byte(req.Content))
		if err != nil {
			return nil, err
		}
		res = &ImportResult{
			Title:            parsed.Title,
			RequestCount:     parsed.RequestCount(),
			EnvironmentCount: len(parsed.Environments),
			Report:           report,
		}
		res.write = func(dir string) error { return parsed.Write(dir) }

	case importer.FormatHAR:
		parsed, report, err := importer.ParseHAR([]byte(req.Content))
		if err != nil {
			return nil, err
		}
		res = &ImportResult{
			Title:        parsed.Title,
			RequestCount: len(parsed.Requests),
			Report:       report,
		}
		res.write = func(dir string) error { return parsed.Write(dir, "") }

	case importer.FormatOpenAPI:
		parsed, err := importer.ParseOpenAPI([]byte(req.Content))
		if err != nil {
			return nil, err
		}
		result := parsed.ToOpenAPIResult()
		requestCount := 0
		for _, c := range result.Collections {
			requestCount += len(c.Request)
		}
		res = &ImportResult{
			Title:        result.Title,
			RequestCount: requestCount,
		}
		if req.DryRun {
			// Preview fidelity comes from the explorer's kin-openapi parse;
			// a spec the strict loader rejects still imports via the
			// tolerant importer, just without operation metadata.
			if doc, err := reqlyopenapi.Load([]byte(req.Content)); err == nil {
				res.Operations = reqlyopenapi.Explore(doc)
			}
		}
		res.write = func(dir string) error { return result.Write(dir) }

	case importer.FormatCurl:
		parsed, err := importer.ParseCurl(req.Content)
		if err != nil {
			return nil, err
		}
		return &ImportResult{
			Kind:    ImportKindRequest,
			Format:  format,
			Request: parsed,
		}, nil
	default:
		return nil, fmt.Errorf("format %q is not supported in the import dialog yet", format)
	}

	res.Kind = ImportKindWorkspace
	res.Format = format
	res.TargetDir = req.TargetDir
	if res.TargetDir == "" {
		res.TargetDir = importer.SanitizeDirName(res.Title)
	} else {
		res.TargetDir = importer.SanitizeDirName(res.TargetDir)
	}
	if !req.DryRun {
		dir, err := s.commitDir(res.TargetDir)
		if err != nil {
			return nil, err
		}
		if err := res.write(dir); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// commitDir resolves and validates the absolute target folder for a commit,
// refusing to touch anything that already exists.
func (s *AppService) commitDir(target string) (string, error) {
	root := s.importRoot()
	if root == "" {
		return "", fmt.Errorf("no workspace found: open a reqly workspace to import into")
	}
	dir := filepath.Join(root, target)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return "", fmt.Errorf("folder %q already exists: pick another name before importing", target)
	} else if err == nil {
		return "", fmt.Errorf("%q already exists and is not a folder", target)
	}
	return dir, nil
}

func (s *AppService) importRoot() string {
	if s.workspace == nil {
		return ""
	}
	return s.root
}
