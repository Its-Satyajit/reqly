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

package secrets

import (
	"os"
	"path/filepath"
)

// Opened is the result of opening a workspace token store. Store is nil when
// no store could be opened (callers proceed without token caching); Warning
// carries a non-fatal fallback message the caller chooses whether to print.
type Opened struct {
	Store   Store
	Backend string // "keychain", "file", or "" when no store was opened
	Warning string
}

// OpenForWorkspace opens the token store for a workspace root, applying the
// shared backend policy every front-end uses: REQLY_TOKEN_STORE overrides
// defaultBackend ("keychain" or "file"); a keychain that cannot be opened
// falls back to the file store; an unknown backend name is treated as "file"
// with a warning. This is the single home for that decision — front-ends must
// not maintain their own copies.
func OpenForWorkspace(root, defaultBackend string) Opened {
	return openForWorkspace(root, defaultBackend, os.Getenv("REQLY_TOKEN_STORE"), openOSKeychain)
}

// keychainOpener abstracts the OS keychain so tests can simulate an
// unavailable credential store.
type keychainOpener func(service, indexPath string) (Store, error)

// keychainOpenerOverride is a test-only replacement for the OS keychain
// opener; nil in production.
var keychainOpenerOverride keychainOpener

// SetKeychainOpenerForTest replaces the OS-keychain opener and returns a
// restore function, mirroring variables.SetTagGeneratorForTest. Production
// code must not call it.
func SetKeychainOpenerForTest(opener keychainOpener) (restore func()) {
	prev := keychainOpenerOverride
	keychainOpenerOverride = opener
	return func() { keychainOpenerOverride = prev }
}

func openOSKeychain(service, indexPath string) (Store, error) {
	if keychainOpenerOverride != nil {
		return keychainOpenerOverride(service, indexPath)
	}
	return NewKeychainStore(service, indexPath)
}

func openForWorkspace(root, defaultBackend, envBackend string, openKeychain keychainOpener) Opened {
	if root == "" {
		return Opened{}
	}
	backend := envBackend
	if backend == "" {
		backend = defaultBackend
	}
	switch backend {
	case "keychain":
		store, err := openKeychain("reqly", filepath.Join(root, ".reqly", "keychain.index"))
		if err != nil {
			return fileStore(root, "keychain unavailable: "+err.Error()+"; falling back to the file store")
		}
		return Opened{Store: store, Backend: "keychain"}
	case "file":
	default:
		backend = "file"
		return fileStore(root, "unknown token store \""+envBackend+"\"; using the file store")
	}
	return fileStore(root, "")
}

func fileStore(root, warning string) Opened {
	store, err := NewFileStore(filepath.Join(root, ".reqly", "tokens.json"))
	if err != nil {
		return Opened{Warning: warning + "; token caching disabled"}
	}
	return Opened{Store: store, Backend: "file", Warning: warning}
}
