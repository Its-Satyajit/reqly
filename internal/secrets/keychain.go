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

package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/zalando/go-keyring"
)

// KeychainStore is a Store backed by the OS credential store (Secret
// Service on Linux, Keychain on macOS, WinCred on Windows) via go-keyring.
// Secret values live in the keychain; the set of stored keys is tracked in
// a 0600 index file because keychain APIs have no portable list operation.
// The account names are hash-derived cache keys (sha256 of workspace root +
// config), not credentials, so the index holds no secret material.
type KeychainStore struct {
	service string
	index   string
	ops     keychainOps
	mu      sync.Mutex
}

// keychainOps is the minimal keyring surface KeychainStore needs. The real
// implementation adapts go-keyring; tests use an in-memory fake, so unit
// tests never touch the OS credential store.
//
// Contract: Get returns an error wrapping fs.ErrNotExist when the account is
// absent; any other error means the keychain itself is unavailable.
type keychainOps interface {
	Set(service, account, value string) error
	Get(service, account string) (string, error)
	Delete(service, account string) error
}

// keyringAdapter adapts the go-keyring package to keychainOps, mapping its
// not-found sentinel to fs.ErrNotExist and treating a delete of an absent
// secret as a no-op.
type keyringAdapter struct{}

func (keyringAdapter) Set(service, account, value string) error {
	return keyring.Set(service, account, value)
}

func (keyringAdapter) Get(service, account string) (string, error) {
	value, err := keyring.Get(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, account)
	}
	if err != nil {
		return "", fmt.Errorf("secrets: keychain get: %w", err)
	}
	return value, nil
}

func (keyringAdapter) Delete(service, account string) error {
	if err := keyring.Delete(service, account); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("secrets: keychain delete: %w", err)
	}
	return nil
}

// availabilityProbeAccount is a key that can never be a real cache key. A
// Get on it returning fs.ErrNotExist proves the keychain is reachable;
// any other error means it is not (no Secret Service daemon, locked, …).
const availabilityProbeAccount = "reqly:availability-probe"

// NewKeychainStore returns a KeychainStore using the OS keychain under
// service name, persisting its key index to indexPath. It probes keychain
// availability at open; when the keychain is unreachable it returns an
// error so callers can fall back to another store.
func NewKeychainStore(service, indexPath string) (*KeychainStore, error) {
	if dir := filepath.Dir(indexPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("secrets: create keychain index dir: %w", err)
		}
	}
	s := &KeychainStore{service: service, index: indexPath, ops: keyringAdapter{}}
	if err := probeKeychain(s.ops, service); err != nil {
		return nil, fmt.Errorf("secrets: keychain unavailable: %w", err)
	}
	return s, nil
}

// probeKeychain checks that ops is reachable by Get-ing an account that can
// never exist: fs.ErrNotExist means the keychain answered (reachable), any
// other error means it is unavailable.
func probeKeychain(ops keychainOps, service string) error {
	if _, err := ops.Get(service, availabilityProbeAccount); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

// Get returns the value stored under key, or an error wrapping
// fs.ErrNotExist when absent.
func (s *KeychainStore) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.readIndex()
	if err != nil {
		return "", err
	}
	if !containsKey(index, key) {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, key)
	}
	value, err := s.ops.Get(s.service, key)
	if err != nil {
		return "", err // fs.ErrNotExist passes through
	}
	return value, nil
}

// Set stores value under key, overwriting any previous value.
func (s *KeychainStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ops.Set(s.service, key, value); err != nil {
		return fmt.Errorf("secrets: keychain set: %w", err)
	}
	index, err := s.readIndex()
	if err != nil {
		return err
	}
	if !containsKey(index, key) {
		index = append(index, key)
		sort.Strings(index)
	}
	return s.writeIndex(index)
}

// Delete removes key. Deleting an absent key is not an error.
func (s *KeychainStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.readIndex()
	if err != nil {
		return err
	}
	if !containsKey(index, key) {
		return nil
	}
	if err := s.ops.Delete(s.service, key); err != nil {
		return err
	}
	return s.writeIndex(removeKey(index, key))
}

// Keys returns all stored keys, or nil when the store is empty. Entries
// whose keychain value was deleted out-of-band are dropped (self-healing
// the index).
func (s *KeychainStore) Keys() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.readIndex()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(index))
	drifted := false
	for _, key := range index {
		if _, err := s.ops.Get(s.service, key); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				drifted = true
				continue
			}
			return nil, err
		}
		keys = append(keys, key)
	}
	if drifted {
		_ = s.writeIndex(keys) // best-effort self-heal
	}
	return keys, nil
}

// readIndex loads the stored keys, or nil when the index does not exist.
func (s *KeychainStore) readIndex() ([]string, error) {
	data, err := os.ReadFile(s.index)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("secrets: read keychain index: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("secrets: parse keychain index: %w", err)
	}
	return keys, nil
}

// writeIndex atomically persists keys at 0600.
func (s *KeychainStore) writeIndex(keys []string) error {
	data, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("secrets: encode keychain index: %w", err)
	}
	return writeFileAtomic(s.index, data, 0o600)
}

func containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

func removeKey(keys []string, key string) []string {
	out := keys[:0]
	for _, k := range keys {
		if k != key {
			out = append(out, k)
		}
	}
	return out
}
