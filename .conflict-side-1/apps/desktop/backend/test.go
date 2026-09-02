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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/testing"
)

// TestFileRef identifies one .reqly-test file in the workspace.
type TestFileRef struct {
	Name string `json:"name"`
	Path string `json:"path"` // workspace-relative, forward slashes
}

// TestFileContent is a test file's raw text plus its detected format.
type TestFileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Format  string `json:"format"` // "yaml" | "json"
	Version string `json:"version"`
}

// TestRunRequest runs assertions from disk (Path) or from unsaved editor
// text (Content wins when non-empty). Env optionally names an environment,
// overridden by REQLY_ENV exactly like `reqly test --env`.
type TestRunRequest struct {
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	Env     string `json:"env,omitempty"`
}

// TestRunResult reports per-test assertion outcomes.
type TestRunResult struct {
	Passed     bool                 `json:"passed"`
	PassCount  int                  `json:"passCount"`
	Total      int                  `json:"total"`
	DurationMs int64                `json:"durationMs"`
	Results    []testing.TestResult `json:"results"`
	Error      string               `json:"error,omitempty"`
}

// TestsList walks the workspace for *.reqly-test.{yaml,yml,json} files —
// the same on-disk contract as `reqly test`.
func (s *AppService) TestsList() ([]TestFileRef, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to list tests")
	}
	var refs []TestFileRef
	err := filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".reqly" {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		if !strings.HasSuffix(stem, ".reqly-test") {
			return nil
		}
		rel, rerr := filepath.Rel(s.root, path)
		if rerr != nil {
			return nil
		}
		refs = append(refs, TestFileRef{
			Name: strings.TrimSuffix(stem, ".reqly-test"),
			Path: filepath.ToSlash(rel),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan tests: %w", err)
	}
	return refs, nil
}

// TestFileRead loads a test file's raw text for the editor. Paths may be
// workspace-relative or absolute.
func (s *AppService) TestFileRead(path string) (*TestFileContent, error) {
	abs, err := s.resolveTestPath(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read test file: %w", err)
	}
	format := "yaml"
	if strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		format = "json"
	}
	tf, perr := testing.ParseTestFile(data)
	name := ""
	env := ""
	if perr == nil && tf != nil {
		name = tf.Name
		env = tf.Environment
	}
	return &TestFileContent{
		Path:    path,
		Content: string(data),
		Format:  format,
		Version: fmt.Sprintf("%s|%s|%d", name, env, time.Now().UnixNano()),
	}, nil
}

// TestFileWrite validates content through the shared parser before writing —
// a broken suite never reaches disk.
func (s *AppService) TestFileWrite(path string, content string) error {
	if _, err := testing.ParseTestFile([]byte(content)); err != nil {
		return fmt.Errorf("invalid test file: %w", err)
	}
	abs, err := s.resolveTestPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write test file: %w", err)
	}
	return nil
}

// TestRun executes the request defined in the test file and evaluates every
// assertion — identical fidelity to `reqly test`, including secret masking
// via the shared pipeline.
func (s *AppService) TestRun(req TestRunRequest) (*TestRunResult, error) {
	started := time.Now()
	var data []byte
	var pathForHistory string
	if strings.TrimSpace(req.Content) != "" {
		data = []byte(req.Content)
		pathForHistory = req.Path
	} else {
		abs, err := s.resolveTestPath(req.Path)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(abs)
		if err != nil {
			return nil, fmt.Errorf("read test file: %w", err)
		}
		data = raw
		pathForHistory = req.Path
	}
	tf, err := testing.ParseTestFile(data)
	if err != nil {
		return nil, fmt.Errorf("parse test file: %w", err)
	}
	if s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to run tests")
	}
	svc := core.NewRunService(s.root)
	defer svc.Close()
	res, runErr := svc.Run(context.Background(), tf.Request, core.RunRequestOptions{
		EnvFlag:     req.Env,
		FileEnv:     tf.Environment,
		FileVars:    tf.VariablesSet(),
		RequestPath: pathForHistory,
	})
	out := &TestRunResult{Total: len(tf.Tests)}
	if runErr != nil {
		out.Error = runErr.Error()
		out.DurationMs = time.Since(started).Milliseconds()
		return out, nil
	}
	results := tf.Suite().Run(res.Response)
	for _, tr := range results {
		if tr.Passed {
			out.PassCount++
		}
	}
	out.Passed = out.PassCount == len(results)
	out.Results = results
	out.DurationMs = time.Since(started).Milliseconds()
	return out, nil
}

// resolveTestPath accepts workspace-relative or absolute paths and refuses
// anything escaping the workspace root.
func (s *AppService) resolveTestPath(path string) (string, error) {
	if s == nil || s.root == "" {
		return "", fmt.Errorf("no workspace found: open a reqly workspace to edit tests")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("no test file path given")
	}
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(s.root, filepath.FromSlash(path))
	}
	cleanRoot, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	cleanAbs, err := filepath.Abs(abs)
	if err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(cleanRoot, cleanAbs); err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("test file %q is outside the workspace", path)
	}
	return cleanAbs, nil
}
