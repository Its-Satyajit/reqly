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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// maskAcquiredToken adds a runtime-acquired auth token (e.g. an OAuth access
// token resolved during Execute) to the output masker so it never renders in
// responses, headers, errors, or logs. It is a no-op when no token was
// acquired.
func maskAcquiredToken(masker *environments.Masker, token string) {
	if token != "" {
		masker.Add(token)
	}
}

// newRequestClient returns a request engine wired with store-backed OAuth
// token caching. The store (file by default, keychain when selected and
// available) lives under <workspace root>/.reqly and scopes cache keys to
// that workspace. When no workspace descriptor can be found, a plain client
// without caching is returned. Extra options are appended for callers that
// need engine callbacks (e.g. retry observers).
func newRequestClient(startDir string, opts ...request.Option) *request.Client {
	root := findWorkspaceRoot(startDir)
	if root == "" {
		return request.NewClient(opts...)
	}
	store, _, err := openTokenStore(root)
	if err != nil {
		warnf("warning: %v; requests will not use cached tokens\n", err)
		return request.NewClient(opts...)
	}
	return request.NewClient(append([]request.Option{request.WithTokenCache(store, root)}, opts...)...)
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
