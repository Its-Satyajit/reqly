package collections

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

func TestSaveWorkspace_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ws := &Workspace{
		Root:   dir,
		Config: Config{Name: "test-ws"},
		Collections: []*Collection{
			{
				Name:   "users",
				Dir:    filepath.Join(dir, "collections", "users"),
				Config: Config{Name: "users"},
				Requests: []*RequestEntry{
					{Name: "list", Path: filepath.Join(dir, "collections", "users", "list.yaml"), File: &requestfile.File{Request: request.Request{Method: "GET", URL: "https://api.example.com/users"}}},
				},
				Folders: []*Folder{
					{
						Name:     "auth",
						Dir:      filepath.Join(dir, "collections", "users", "auth"),
						Config:   Config{Name: "auth"},
						Requests: []*RequestEntry{{Name: "login", Path: filepath.Join(dir, "collections", "users", "auth", "login.json"), File: &requestfile.File{Request: request.Request{Method: "POST", URL: "https://api.example.com/login"}}}},
					},
				},
			},
		},
	}
	if err := SaveWorkspace(dir, ws); err != nil {
		t.Fatalf("SaveWorkspace: %v", err)
	}
	loaded, err := LoadWorkspace(dir)
	if err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
	}
	if len(loaded.Collections) != 1 || loaded.Collections[0].Name != "users" {
		t.Fatalf("collections: %v", loaded.Collections)
	}
	if len(loaded.Collections[0].Requests) != 1 || loaded.Collections[0].Requests[0].Name != "list" {
		t.Fatalf("requests: %v", loaded.Collections[0].Requests)
	}
	if len(loaded.Collections[0].Folders) != 1 || loaded.Collections[0].Folders[0].Name != "auth" {
		t.Fatalf("folders: %v", loaded.Collections[0].Folders)
	}
	// check .json preserved
	if _, err := os.Stat(filepath.Join(dir, "collections", "users", "auth", "login.json")); err != nil {
		t.Fatalf("json not preserved: %v", err)
	}
}

func TestSaveWorkspace_PruneDeleted(t *testing.T) {
	dir := t.TempDir()
	ws := &Workspace{Root: dir, Config: Config{Name: "ws"}, Collections: []*Collection{{Name: "a", Dir: filepath.Join(dir, "collections", "a"), Config: Config{Name: "a"}, Requests: []*RequestEntry{{Name: "one", Path: filepath.Join(dir, "collections", "a", "one.yaml"), File: &requestfile.File{Request: request.Request{Method: "GET", URL: "https://a.com"}}}}}}}
	_ = SaveWorkspace(dir, ws)
	// now remove request and save again
	ws.Collections[0].Requests = nil
	if err := SaveWorkspace(dir, ws); err != nil {
		t.Fatalf("SaveWorkspace prune: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "collections", "a", "one.yaml")); !os.IsNotExist(err) {
		t.Fatalf("prune failed, file still exists")
	}
	// remove collection
	ws.Collections = nil
	_ = SaveWorkspace(dir, ws)
	if _, err := os.Stat(filepath.Join(dir, "collections", "a")); !os.IsNotExist(err) {
		t.Fatalf("prune collection failed")
	}
}
