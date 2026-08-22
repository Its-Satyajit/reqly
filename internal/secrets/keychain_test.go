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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
)

// fakeKeyring is an in-memory keychainOps implementing the adapter contract:
// Get on an absent account returns fs.ErrNotExist; a non-nil fail error
// simulates an unavailable keychain.
type fakeKeyring struct {
	mu     sync.Mutex
	values map[string]string
	fail   error
}

func (f *fakeKeyring) key(service, account string) string { return service + "\x00" + account }

func (f *fakeKeyring) Set(service, account, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return f.fail
	}
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[f.key(service, account)] = value
	return nil
}

func (f *fakeKeyring) Get(service, account string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return "", f.fail
	}
	value, ok := f.values[f.key(service, account)]
	if !ok {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, account)
	}
	return value, nil
}

func (f *fakeKeyring) Delete(service, account string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return f.fail
	}
	delete(f.values, f.key(service, account))
	return nil
}

// newTestKeychainStore builds a KeychainStore over a fresh index path with
// the given driver, bypassing the availability probe. It creates the index
// parent dir like NewKeychainStore does.
func newTestKeychainStore(t *testing.T, ops keychainOps) *KeychainStore {
	t.Helper()
	index := filepath.Join(t.TempDir(), ".reqly", "keychain.index")
	if err := os.MkdirAll(filepath.Dir(index), 0o700); err != nil {
		t.Fatal(err)
	}
	return &KeychainStore{service: "test-service", index: index, ops: ops}
}

func TestKeychainStoreRoundTrip(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
	if err := s.Set("ws:abc", "access-token-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := s.Get("ws:abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "access-token-value" {
		t.Fatalf("Get = %q, want %q", got, "access-token-value")
	}
}

func TestKeychainStoreGetUnknownKeyReturnsNotExist(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
	_, err := s.Get("ws:never-written")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get unknown key err = %v, want fs.ErrNotExist", err)
	}
}

func TestKeychainStoreDeleteRemovesKey(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
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

func TestKeychainStoreSetOverwritesWithoutDuplicating(t *testing.T) {
	kr := &fakeKeyring{}
	s := newTestKeychainStore(t, kr)
	if err := s.Set("k", "v1"); err != nil {
		t.Fatalf("Set 1: %v", err)
	}
	if err := s.Set("k", "v2"); err != nil {
		t.Fatalf("Set 2: %v", err)
	}
	got, err := s.Get("k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "v2" {
		t.Fatalf("Get = %q, want v2", got)
	}
	keys, err := s.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != "k" {
		t.Fatalf("Keys = %v, want [k]", keys)
	}
}

func TestKeychainStoreKeys(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
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
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "ws:one" || keys[1] != "ws:two" {
		t.Fatalf("Keys = %v, want [ws:one ws:two]", keys)
	}
}

func TestKeychainStoreKeysFiltersDriftedEntries(t *testing.T) {
	kr := &fakeKeyring{}
	s := newTestKeychainStore(t, kr)
	if err := s.Set("ws:kept", "a"); err != nil {
		t.Fatalf("Set 1: %v", err)
	}
	if err := s.Set("ws:gone", "b"); err != nil {
		t.Fatalf("Set 2: %v", err)
	}
	// Out-of-band deletion (e.g. via the OS keychain app).
	if err := kr.Delete("test-service", "ws:gone"); err != nil {
		t.Fatalf("out-of-band delete: %v", err)
	}

	keys, err := s.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != "ws:kept" {
		t.Fatalf("Keys = %v, want [ws:kept] (drifted entry dropped)", keys)
	}
	// The index was self-healed: a fresh instance sees only ws:kept.
	s2 := &KeychainStore{service: "test-service", index: s.index, ops: kr}
	keys, err = s2.Keys()
	if err != nil {
		t.Fatalf("Keys on reloaded store: %v", err)
	}
	if len(keys) != 1 || keys[0] != "ws:kept" {
		t.Fatalf("Keys on reloaded store = %v, want [ws:kept]", keys)
	}
}

func TestKeychainStorePersistsIndexAcrossInstances(t *testing.T) {
	kr := &fakeKeyring{}
	s1 := newTestKeychainStore(t, kr)
	if err := s1.Set("ws:k", "persisted-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	s2 := &KeychainStore{service: "test-service", index: s1.index, ops: kr}
	got, err := s2.Get("ws:k")
	if err != nil {
		t.Fatalf("Get on reloaded store: %v", err)
	}
	if got != "persisted-value" {
		t.Fatalf("Get = %q, want %q", got, "persisted-value")
	}
}

func TestKeychainStoreScopesKeysIndependently(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
	if err := s.Set("ws:one", "token-a"); err != nil {
		t.Fatalf("Set 1: %v", err)
	}
	if err := s.Set("ws:two", "token-b"); err != nil {
		t.Fatalf("Set 2: %v", err)
	}
	a, err := s.Get("ws:one")
	if err != nil || a != "token-a" {
		t.Fatalf("Get 1 = %q, %v", a, err)
	}
	b, err := s.Get("ws:two")
	if err != nil || b != "token-b" {
		t.Fatalf("Get 2 = %q, %v", b, err)
	}
}

func TestKeychainStoreConcurrentSets(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.Set(fmt.Sprintf("ws:%d", i), "v"); err != nil {
				t.Errorf("Set %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	keys, err := s.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 20 {
		t.Fatalf("Keys = %d entries, want 20", len(keys))
	}
}

func TestProbeKeychainReachable(t *testing.T) {
	if err := probeKeychain(&fakeKeyring{}, "probe-service"); err != nil {
		t.Fatalf("probe on reachable keychain: %v", err)
	}
}

func TestProbeKeychainUnavailable(t *testing.T) {
	kr := &fakeKeyring{fail: errors.New("no secret service")}
	if err := probeKeychain(kr, "probe-service"); err == nil {
		t.Fatal("probe on unavailable keychain: err = nil, want error")
	}
}

func TestKeychainStoreIndexPerm(t *testing.T) {
	s := newTestKeychainStore(t, &fakeKeyring{})
	if err := s.Set("k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	info, err := os.Stat(s.index)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("index perm = %o, want 600", perm)
	}
}

// TestKeychainStoreRealDriver exercises the real go-keyring adapter against
// the OS credential store. It is skipped unless REQLY_TEST_KEYCHAIN=1 so CI
// (which has no Secret Service) never touches the keychain.
func TestKeychainStoreRealDriver(t *testing.T) {
	if os.Getenv("REQLY_TEST_KEYCHAIN") != "1" {
		t.Skip("REQLY_TEST_KEYCHAIN=1 to exercise the real OS keychain")
	}
	index := filepath.Join(t.TempDir(), ".reqly", "keychain.index")
	s, err := NewKeychainStore("reqly-test", index)
	if err != nil {
		t.Fatalf("NewKeychainStore: %v (is a keychain available?)", err)
	}
	key := "test-key-" + t.Name()
	if err := s.Set(key, "smoke-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	defer s.Delete(key)
	got, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "smoke-value" {
		t.Fatalf("Get = %q, want smoke-value", got)
	}
	keys, err := s.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	found := false
	for _, k := range keys {
		if k == key {
			found = true
		}
	}
	if !found {
		t.Fatalf("Keys = %v, missing %q", keys, key)
	}
}
