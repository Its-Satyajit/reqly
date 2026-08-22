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
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// tempStoreDir returns a fresh directory to hold a token store for a test.
func tempStoreDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestFileStoreSetGetRoundTrip(t *testing.T) {
	dir := tempStoreDir(t)
	s, err := NewFileStore(filepath.Join(dir, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	if err := s.Set("ws:abc123", "access-token-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := s.Get("ws:abc123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "access-token-value" {
		t.Fatalf("Get = %q, want %q", got, "access-token-value")
	}
}

func TestFileStoreGetUnknownKeyReturnsNotExist(t *testing.T) {
	dir := tempStoreDir(t)
	s, err := NewFileStore(filepath.Join(dir, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	_, err = s.Get("ws:never-written")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get unknown key err = %v, want fs.ErrNotExist", err)
	}
}

func TestFileStoreSetCreatesFileWith0600(t *testing.T) {
	dir := tempStoreDir(t)
	path := filepath.Join(dir, ".reqly", "tokens.json")
	s, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	if err := s.Set("k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("tokens.json perm = %o, want 600", perm)
	}
}

func TestFileStorePersistsAcrossInstances(t *testing.T) {
	dir := tempStoreDir(t)
	path := filepath.Join(dir, ".reqly", "tokens.json")

	s1, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore 1: %v", err)
	}
	if err := s1.Set("ws:k", "persisted-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	s2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore 2: %v", err)
	}
	got, err := s2.Get("ws:k")
	if err != nil {
		t.Fatalf("Get on reloaded store: %v", err)
	}
	if got != "persisted-value" {
		t.Fatalf("Get = %q, want %q", got, "persisted-value")
	}
}

func TestFileStoreDeleteRemovesKey(t *testing.T) {
	dir := tempStoreDir(t)
	s, err := NewFileStore(filepath.Join(dir, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	if err := s.Set("k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := s.Delete("k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get("k"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get after delete err = %v, want fs.ErrNotExist", err)
	}

	// Deleting an absent key is not an error.
	if err := s.Delete("k"); err != nil {
		t.Fatalf("Delete absent key: %v", err)
	}
}

func TestFileStoreScopesKeysIndependently(t *testing.T) {
	dir := tempStoreDir(t)
	s, err := NewFileStore(filepath.Join(dir, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	if err := s.Set("ws:one", "token-a"); err != nil {
		t.Fatalf("Set 1: %v", err)
	}
	if err := s.Set("ws:two", "token-b"); err != nil {
		t.Fatalf("Set 2: %v", err)
	}

	a, err := s.Get("ws:one")
	if err != nil {
		t.Fatalf("Get 1: %v", err)
	}
	if a != "token-a" {
		t.Fatalf("Get 1 = %q, want %q", a, "token-a")
	}
	b, err := s.Get("ws:two")
	if err != nil {
		t.Fatalf("Get 2: %v", err)
	}
	if b != "token-b" {
		t.Fatalf("Get 2 = %q, want %q", b, "token-b")
	}
}

func TestFileStoreKeys(t *testing.T) {
	dir := tempStoreDir(t)
	s, err := NewFileStore(filepath.Join(dir, ".reqly", "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	keys, err := s.Keys()
	if err != nil {
		t.Fatalf("Keys on empty store: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("Keys on empty store = %v, want none", keys)
	}

	if err := s.Set("ws:one", "a"); err != nil {
		t.Fatalf("Set 1: %v", err)
	}
	if err := s.Set("ws:two", "b"); err != nil {
		t.Fatalf("Set 2: %v", err)
	}

	keys, err = s.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	want := map[string]bool{"ws:one": true, "ws:two": true}
	if len(keys) != 2 {
		t.Fatalf("Keys = %v, want 2 entries", keys)
	}
	for _, k := range keys {
		if !want[k] {
			t.Fatalf("unexpected key %q", k)
		}
	}

	if err := s.Delete("ws:one"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	keys, err = s.Keys()
	if err != nil {
		t.Fatalf("Keys after delete: %v", err)
	}
	if len(keys) != 1 || keys[0] != "ws:two" {
		t.Fatalf("Keys after delete = %v, want [ws:two]", keys)
	}
}
