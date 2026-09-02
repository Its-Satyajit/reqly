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

// Package testsupport holds test-only fixtures shared across packages:
// workspace layouts today; fakes for external systems (OAuth providers,
// clocks) as they are needed. Importing this package outside _test.go files
// of test builds is a bug — keep production dependency-free from it.
package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

// Workspace lays out a workspace fixture in a fresh temp directory and
// returns its root. Files maps workspace-relative paths to contents; parent
// directories are created automatically. When no reqly.yaml is supplied a
// minimal descriptor is written so the directory is a valid workspace.
//
//	testsupport.Workspace(t, map[string]string{
//		"environments/dev.yaml": "variables:\n  REGION: us-west-2\n",
//	})
func Workspace(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	Files(t, dir, files)
	return dir
}

// Files lays out files under an existing root (created if needed), applying
// the same default-descriptor rule as Workspace.
func Files(t *testing.T, root string, files map[string]string) {
	t.Helper()
	if _, ok := files["reqly.yaml"]; !ok {
		if _, err := os.Stat(filepath.Join(root, "reqly.yaml")); os.IsNotExist(err) {
			files["reqly.yaml"] = "name: test-ws\n"
		}
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("testsupport: mkdir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("testsupport: write %s: %v", name, err)
		}
	}
}
