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

package environments

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
	reqtesting "github.com/Its-Satyajit/reqly/internal/testing"
)

// Severity classifies a validation finding.
type Severity string

// Validation severities.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Issue is a single validation finding.
type Issue struct {
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// secretNamePattern matches variable names that look like they hold secrets:
// key, token, secret, password, credential.
var secretNamePattern = regexp.MustCompile(`(?i)(key|token|secret|password|credential)`)

// varReferencePattern matches {{name}} placeholders in request and test files.
var varReferencePattern = regexp.MustCompile(`\{\{\s*([\w.-]+)\s*\}\}`)

// Validate checks an environment and returns its issues. When workspaceDir is
// non-empty and points at a workspace, it also scans the workspace's request
// and test files for variables the environment does not provide. The workspace
// scan is skipped (without error) when workspaceDir is not a workspace, so a
// standalone environment file remains valid outside a workspace.
func Validate(env *Environment, workspaceDir string) ([]Issue, error) {
	var issues []Issue

	if workspaceDir != "" && collections.IsWorkspace(workspaceDir) {
		undefined, err := validateUndefinedVariables(env, workspaceDir)
		if err != nil {
			return nil, err
		}
		issues = append(issues, undefined...)
	}

	for key := range env.Variables {
		if _, dup := env.Secrets[key]; dup {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("key %q defined in both variables and secrets", key),
			})
			continue
		}
		if secretNamePattern.MatchString(key) {
			issues = append(issues, Issue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("variable %q looks like a secret but lives in variables (move to secrets)", key),
			})
		}
	}

	return issues, nil
}

// validateUndefinedVariables scans request and test files under the workspace
// for {{var}} references the environment does not provide. A variable counts
// as defined when it is provided by any scope in the precedence chain
// (process-env → global → environment → collection → folder → request), so a
// reference backed by the workspace or a request file's own variables is not
// flagged.
func validateUndefinedVariables(env *Environment, workspaceDir string) ([]Issue, error) {
	ws, err := collections.LoadWorkspace(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("scan workspace: %w", err)
	}

	// Vars defined by the environment plus every scope the environment cannot
	// see past: process env, workspace/collection/folder globals, and each
	// request file's own variables.
	defined := make(map[string]struct{})
	addKeys(defined, env.Variables)
	addKeys(defined, env.Secrets)
	addKeys(defined, ws.Config.Variables)

	missing := make(map[string]struct{})
	for _, coll := range ws.Collections {
		collDefined := cloneDefined(defined)
		addKeys(collDefined, coll.Config.Variables)
		collectUndefined(missing, collDefined, coll.Requests)
		walkFolders(missing, collDefined, coll.Config.Variables, coll.Folders)
	}
	if err := scanTestFiles(missing, defined, workspaceDir); err != nil {
		return nil, err
	}

	var issues []Issue
	for name := range missing {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("undefined variable %q referenced by a workspace request or test", name),
		})
	}
	return issues, nil
}

// addKeys merges the keys of a variable map into a defined set.
func addKeys(defined map[string]struct{}, vars map[string]string) {
	for key := range vars {
		defined[key] = struct{}{}
	}
}

// scanTestFiles walks workspaceDir for test files and collects their variable
// references. Test files live outside the collection tree, so they need a
// directory scan of their own.
func scanTestFiles(missing map[string]struct{}, defined map[string]struct{}, workspaceDir string) error {
	return filepath.WalkDir(workspaceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip .git and the environments directory: no test files there.
			switch d.Name() {
			case ".git", "environments", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".json", ".yaml", ".yml":
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tf, err := reqtesting.ParseTestFile(data)
		if err != nil {
			return nil // not a test file; skip
		}
		for _, name := range referencedVariables(&tf.Request) {
			if _, ok := defined[name]; ok {
				continue
			}
			if _, self := tf.Variables[name]; self {
				continue
			}
			missing[name] = struct{}{}
		}
		return nil
	})
}

// walkFolders recursively collects undefined variable references. parentVars
// are the variables inherited from the enclosing container (collection or
// parent folder); they apply to every request in this subtree.
func walkFolders(missing map[string]struct{}, defined map[string]struct{}, parentVars map[string]string, folders []*collections.Folder) {
	for _, folder := range folders {
		scope := cloneDefined(defined)
		addKeys(scope, parentVars)
		addKeys(scope, folder.Config.Variables)
		collectUndefined(missing, scope, folder.Requests)
		walkFolders(missing, scope, folder.Config.Variables, folder.Folders)
	}
}

// collectUndefined scans request entries, adding referenced-but-undefined
// variables to missing. defined already carries the full scope chain for these
// entries (process-env, global, environment, collection, folder). A variable
// defined in the request file itself is treated as defined.
func collectUndefined(missing map[string]struct{}, defined map[string]struct{}, entries []*collections.RequestEntry) {
	for _, entry := range entries {
		for _, name := range referencedVariables(&entry.File.Request) {
			if _, ok := defined[name]; ok {
				continue
			}
			if _, self := entry.File.Variables[name]; self {
				continue
			}
			missing[name] = struct{}{}
		}
	}
}

// cloneDefined returns a copy of a defined set so subtree scope additions do
// not leak across siblings.
func cloneDefined(defined map[string]struct{}) map[string]struct{} {
	copy := make(map[string]struct{}, len(defined))
	for key := range defined {
		copy[key] = struct{}{}
	}
	return copy
}

// referencedVariables extracts {{name}} references from a request across its
// URL, body, headers, query parameters, and auth configuration.
func referencedVariables(r *request.Request) []string {
	var haystack []string
	haystack = append(haystack, r.URL, r.Body)
	for _, h := range r.Headers {
		haystack = append(haystack, h.Key, h.Value)
	}
	for _, q := range r.Query {
		haystack = append(haystack, q.Key, q.Value)
	}
	if r.Auth.Type != "" {
		for _, v := range r.Auth.Config {
			haystack = append(haystack, v)
		}
	}

	var names []string
	for _, text := range haystack {
		matches := varReferencePattern.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			names = append(names, m[1])
		}
	}
	return names
}
