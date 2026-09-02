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
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(path string) error {
	return os.WriteFile(path, []byte("x"), 0o644)
}

func okOpener(t *testing.T) keychainOpener {
	return func(service, indexPath string) (Store, error) {
		if service != "reqly" {
			t.Errorf("unexpected service %q", service)
		}
		return NewFileStore(filepath.Join(filepath.Dir(indexPath), "opened-via-keychain.json"))
	}
}

func failingOpener() keychainOpener {
	return func(string, string) (Store, error) { return nil, errors.New("no keyring daemon") }
}

func TestOpenForWorkspace_EmptyRoot(t *testing.T) {
	got := openForWorkspace("", "file", "", okOpener(t))
	if got.Store != nil || got.Backend != "" || got.Warning != "" {
		t.Fatalf("expected empty result for empty root, got %+v", got)
	}
}

func TestOpenForWorkspace_DefaultsToFile(t *testing.T) {
	root := t.TempDir()
	got := openForWorkspace(root, "file", "", okOpener(t))
	if got.Store == nil || got.Backend != "file" || got.Warning != "" {
		t.Fatalf("expected file store without warning, got %+v", got)
	}
}

func TestOpenForWorkspace_KeychainPreferred(t *testing.T) {
	root := t.TempDir()
	got := openForWorkspace(root, "keychain", "", okOpener(t))
	if got.Backend != "keychain" || got.Store == nil || got.Warning != "" {
		t.Fatalf("expected keychain store, got %+v", got)
	}
}

func TestOpenForWorkspace_EnvOverridesDefault(t *testing.T) {
	root := t.TempDir()
	got := openForWorkspace(root, "keychain", "file", okOpener(t))
	if got.Backend != "file" || got.Warning != "" {
		t.Fatalf("expected REQLY_TOKEN_STORE to override default to file, got %+v", got)
	}

	got = openForWorkspace(root, "file", "keychain", okOpener(t))
	if got.Backend != "keychain" || got.Warning != "" {
		t.Fatalf("expected env override to select keychain, got %+v", got)
	}
}

func TestOpenForWorkspace_KeychainFallsBackToFile(t *testing.T) {
	root := t.TempDir()
	got := openForWorkspace(root, "keychain", "", failingOpener())
	if got.Store == nil || got.Backend != "file" {
		t.Fatalf("expected fallback to file store, got %+v", got)
	}
	if got.Warning == "" {
		t.Fatal("expected a non-fatal warning on keychain fallback")
	}
}

func TestOpenForWorkspace_UnknownBackendTreatedAsFile(t *testing.T) {
	root := t.TempDir()
	got := openForWorkspace(root, "file", "memcached", okOpener(t))
	if got.Store == nil || got.Backend != "file" {
		t.Fatalf("expected unknown backend to fall back to file, got %+v", got)
	}
	if got.Warning == "" {
		t.Fatal("expected warning naming the unknown backend")
	}
}

func TestOpenForWorkspace_FileErrorDisablesCaching(t *testing.T) {
	// root points at a path that cannot host .reqly/tokens.json: a file, not a dir.
	root := filepath.Join(t.TempDir(), "not-a-dir")
	if err := writeFile(root); err != nil {
		t.Fatal(err)
	}
	got := openForWorkspace(root, "file", "", failingOpener())
	if got.Store != nil || got.Backend != "" || got.Warning == "" {
		t.Fatalf("expected nil store with warning, got %+v", got)
	}
}
