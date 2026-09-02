package collab

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		ws      SharedWorkspace
		wantErr bool
	}{
		{name: "valid", ws: SharedWorkspace{Path: "/tmp/ws", Collaborators: []Collaborator{{User: "alice", Role: "admin"}}}, wantErr: false},
		{name: "missing path", ws: SharedWorkspace{Collaborators: []Collaborator{{User: "a", Role: "viewer"}}}, wantErr: true},
		{name: "empty user", ws: SharedWorkspace{Path: "/tmp", Collaborators: []Collaborator{{User: "", Role: "viewer"}}}, wantErr: true},
		{name: "bad role", ws: SharedWorkspace{Path: "/tmp", Collaborators: []Collaborator{{User: "bob", Role: "bad"}}}, wantErr: true},
		{name: "duplicate", ws: SharedWorkspace{Path: "/tmp", Collaborators: []Collaborator{{User: "a", Role: "viewer"}, {User: "a", Role: "editor"}}}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.ws)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestAddRemove(t *testing.T) {
	ws := SharedWorkspace{Path: "/tmp/ws"}
	if err := AddCollaborator(&ws, "alice", "admin"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !IsCollaborator(ws, "alice") {
		t.Fatalf("should be collaborator")
	}
	// Update role
	if err := AddCollaborator(&ws, "alice", "viewer"); err != nil {
		t.Fatalf("Add update: %v", err)
	}
	if ws.Collaborators[0].Role != "viewer" {
		t.Fatalf("want viewer, got %s", ws.Collaborators[0].Role)
	}
	if err := RemoveCollaborator(&ws, "alice"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if IsCollaborator(ws, "alice") {
		t.Fatalf("should not be collaborator after remove")
	}
	if err := RemoveCollaborator(&ws, "missing"); err == nil {
		t.Fatalf("expected error for missing")
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".reqly", "collab.yaml")
	ws := SharedWorkspace{Path: dir, Collaborators: []Collaborator{{User: "bob", Role: "editor"}}}
	if err := Save(path, ws); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if fi, err := os.Stat(path); err == nil && fi.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", fi.Mode().Perm())
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Collaborators) != 1 || loaded.Collaborators[0].User != "bob" {
		t.Fatalf("unexpected %+v", loaded)
	}
	// Missing returns empty
	missing, err := Load(filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if missing.Path == "" {
		t.Fatalf("expected path for missing")
	}
}
