package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		entry   Entry
		wantErr bool
	}{
		{name: "valid", entry: Entry{Action: "request.send", Resource: "users/list"}, wantErr: false},
		{name: "missing action", entry: Entry{Resource: "x"}, wantErr: true},
		{name: "unknown action", entry: Entry{Action: "unknown", Resource: "x"}, wantErr: true},
		{name: "missing resource", entry: Entry{Action: "request.send"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.entry)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestStore_AddAndList(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Ensure 0600
	if fi, err := os.Stat(filepath.Join(dir, ".reqly", "audit.log")); err == nil && fi.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", fi.Mode().Perm())
	}
	e1, err := store.Add(Entry{Action: "request.send", Resource: "users/list", Actor: "tester", Details: "GET /users"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if e1.ID == "" || e1.Timestamp.IsZero() {
		t.Fatalf("expected ID and timestamp set, got %+v", e1)
	}
	time.Sleep(2 * time.Millisecond)
	e2, err := store.Add(Entry{Action: "workflow.run", Resource: "my-workflow"})
	if err != nil {
		t.Fatalf("Add2: %v", err)
	}
	if e2.Actor != "local" {
		t.Fatalf("want local actor, got %q", e2.Actor)
	}
	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	if list[0].ID != e1.ID || list[1].ID != e2.ID {
		t.Fatalf("unexpected order %+v", list)
	}
}

func TestStore_Clear(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add(Entry{Action: "request.send", Resource: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("want 0 after clear, got %d", len(list))
	}
}

func TestStore_InvalidAction(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)
	if _, err := store.Add(Entry{Action: "bad", Resource: "x"}); err == nil {
		t.Fatalf("expected error for bad action")
	}
}
