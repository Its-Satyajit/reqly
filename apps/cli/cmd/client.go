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
	"fmt"
	"os"
	"path/filepath"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// newRequestClient returns a request engine wired with store-backed OAuth
// token caching. The store (file by default, keychain when selected and
// available) lives under <workspace root>/.reqly and scopes cache keys to
// that workspace. When no workspace descriptor can be found, a plain client
// without caching is returned.
func newRequestClient(startDir string) *request.Client {
	root := findWorkspaceRoot(startDir)
	if root == "" {
		return request.NewClient()
	}
	store, _, err := openTokenStore(root)
	if err != nil {
		warnf("warning: %v; requests will not use cached tokens\n", err)
		return request.NewClient()
	}
	return request.NewClient(request.WithTokenCache(store, root))
}

// newKeychainStore opens the OS-keychain store for a workspace root. It is a
// variable so tests can force the unavailable path deterministically instead
// of depending on whether the host has a Secret Service.
var newKeychainStore = secrets.NewKeychainStore

// openTokenStore opens the token store for a workspace root per
// storeBackendFor, returning the store and the active backend name. A
// keychain backend that cannot be opened falls back to the file store with a
// warning.
func openTokenStore(root string) (secrets.Store, string, error) {
	backend := storeBackendFor()
	switch backend {
	case "keychain":
		store, err := newKeychainStore("reqly", filepath.Join(root, ".reqly", "keychain.index"))
		if err != nil {
			warnf("warning: %v; falling back to the file store\n", err)
			backend = "file"
			break
		}
		return store, "keychain", nil
	case "file":
	default:
		return nil, "", fmt.Errorf("unknown token store %q (want file or keychain)", backend)
	}
	store, err := secrets.NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		return nil, "", err
	}
	return store, backend, nil
}

// warnf prints a warning to stderr. It is a variable so tests can capture
// (and silence) it.
var warnf = func(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// findWorkspaceRoot walks up from dir to the nearest workspace descriptor. It
// delegates to collections.FindWorkspaceRoot — the shared home for workspace
// discovery used by the CLI and the desktop app.
func findWorkspaceRoot(dir string) string {
	return collections.FindWorkspaceRoot(dir)
}
