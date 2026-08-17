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
	"sync"
)

// Store persists secret values (tokens) keyed by a stable identifier, with
// the value never allowed to surface in logs. Implementations must write
// atomically and lock on read-modify-write so concurrent callers cannot
// corrupt state.
type Store interface {
	// Get returns the value stored under key. It returns an error wrapping
	// fs.ErrNotExist when no value is stored under key.
	Get(key string) (string, error)
	// Set stores value under key, overwriting any previous value.
	Set(key, value string) error
	// Delete removes key. Deleting an absent key is not an error.
	Delete(key string) error
	// Keys returns all stored keys, or nil when the store is empty.
	Keys() ([]string, error)
}

// FileStore is a Store backed by a single JSON file. Values are keyed by
// string and written with 0600 permissions via temp-file + rename so a crash
// never leaves a half-written store.
type FileStore struct {
	path string
	mu   sync.Mutex
}

// NewFileStore returns a FileStore persisting to path. The parent directory
// is created if absent.
func NewFileStore(path string) (*FileStore, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("secrets: create store dir: %w", err)
		}
	}
	return &FileStore{path: path}, nil
}

// Get returns the value stored under key, or fs.ErrNotExist when absent.
func (s *FileStore) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, key)
	}
	if err != nil {
		return "", fmt.Errorf("secrets: read store: %w", err)
	}

	var entries map[string]string
	if len(data) > 0 {
		if err := json.Unmarshal(data, &entries); err != nil {
			return "", fmt.Errorf("secrets: parse store: %w", err)
		}
	}
	value, ok := entries[key]
	if !ok {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, key)
	}
	return value, nil
}

// Set stores value under key, creating the store file if needed.
func (s *FileStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := map[string]string{}
	if data, err := os.ReadFile(s.path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("secrets: parse store: %w", err)
		}
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("secrets: read store: %w", err)
	}
	entries[key] = value

	return s.write(entries)
}

// Delete removes key from the store. Deleting an absent key is not an error.
func (s *FileStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("secrets: read store: %w", err)
	}

	entries := map[string]string{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("secrets: parse store: %w", err)
		}
	}
	if _, ok := entries[key]; !ok {
		return nil
	}
	delete(entries, key)

	return s.write(entries)
}

// Keys returns all stored keys, or nil when the store is empty.
func (s *FileStore) Keys() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("secrets: read store: %w", err)
	}

	var entries map[string]string
	if len(data) > 0 {
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("secrets: parse store: %w", err)
		}
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	return keys, nil
}

// write atomically replaces the store file with entries.
func (s *FileStore) write(entries map[string]string) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("secrets: encode store: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".tokens-*.tmp")
	if err != nil {
		return fmt.Errorf("secrets: create temp store: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("secrets: chmod temp store: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("secrets: write temp store: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("secrets: sync temp store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("secrets: close temp store: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("secrets: replace store: %w", err)
	}
	return nil
}
