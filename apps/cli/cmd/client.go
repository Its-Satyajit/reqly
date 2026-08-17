// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// newRequestClient returns a request engine wired with store-backed OAuth
// token caching. The store lives at <workspace root>/.reqly/tokens.json and
// scopes cache keys to that workspace. When no workspace descriptor can be
// found, a plain client without caching is returned.
func newRequestClient(startDir string) *request.Client {
	root := findWorkspaceRoot(startDir)
	if root == "" {
		return request.NewClient()
	}
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		return request.NewClient()
	}
	return request.NewClient(request.WithTokenCache(store, root))
}

// findWorkspaceRoot walks up from dir to the nearest directory containing a
// reqly.yaml descriptor, returning its absolute path, or "" when none exists.
func findWorkspaceRoot(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		descriptor := filepath.Join(abs, "reqly.yaml")
		if info, err := os.Stat(descriptor); err == nil && !info.IsDir() {
			return abs
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}
