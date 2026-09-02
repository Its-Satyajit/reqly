// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry is one audit trail record. Stored as JSONL in .reqly/audit.log (0600).
type Entry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details,omitempty"`
}

var validActions = map[string]bool{
	"request.send":   true,
	"workflow.run":   true,
	"automation.run": true,
	"collection.run": true,
	"theme.import":   true,
	"theme.export":   true,
	"auth.login":     true,
	"auth.logout":    true,
	"env.update":     true,
	"mock.start":     true,
	"mock.stop":      true,
}

// Validate checks an entry for required fields.
func Validate(e Entry) error {
	if strings.TrimSpace(e.Action) == "" {
		return fmt.Errorf("action is required")
	}
	if !validActions[e.Action] {
		return fmt.Errorf("unknown action %q", e.Action)
	}
	if strings.TrimSpace(e.Resource) == "" {
		return fmt.Errorf("resource is required")
	}
	return nil
}

// Store is a file-backed audit log. Each workspace has its own .reqly/audit.log.
type Store struct {
	mu   sync.Mutex
	path string
}

// NewStore returns a store rooted at workspace dir (e.g. "."). It ensures
// .reqly exists with 0700 and audit.log with 0600.
func NewStore(root string) (*Store, error) {
	dir := filepath.Join(root, ".reqly")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("mkdir .reqly: %w", err)
	}
	path := filepath.Join(dir, "audit.log")
	// Ensure file exists with 0600
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			return nil, fmt.Errorf("create audit.log: %w", err)
		}
	} else if err == nil {
		_ = os.Chmod(path, 0o600)
	}
	return &Store{path: path}, nil
}

// Add appends an entry. ID and Timestamp are set if empty.
func (s *Store) Add(e Entry) (Entry, error) {
	if err := Validate(e); err != nil {
		return Entry{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if e.ID == "" {
		e.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	if e.Actor == "" {
		e.Actor = "local"
	}
	data, err := json.Marshal(e)
	if err != nil {
		return Entry{}, err
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return Entry{}, err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// List returns all entries sorted by timestamp ascending.
func (s *Store) List() ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var out []Entry
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // skip corrupt lines
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp.Before(out[j].Timestamp) })
	return out, nil
}

// Clear truncates the log.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.WriteFile(s.path, nil, 0o600)
}
