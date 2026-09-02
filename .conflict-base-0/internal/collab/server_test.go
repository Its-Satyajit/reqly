package collab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServer_Health(t *testing.T) {
	srv := NewServer(t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("unexpected %v", out)
	}
}

func TestServer_Collab(t *testing.T) {
	dir := t.TempDir()
	// Create a collab file
	ws := SharedWorkspace{Path: dir, Collaborators: []Collaborator{{User: "alice", Role: "admin"}}}
	path := DefaultPath(dir)
	if err := Save(path, ws); err != nil {
		t.Fatalf("Save: %v", err)
	}
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/collab", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var out SharedWorkspace
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Collaborators) != 1 || out.Collaborators[0].User != "alice" {
		t.Fatalf("unexpected %+v", out)
	}
}

func TestServer_Workspace(t *testing.T) {
	dir := t.TempDir()
	// Create collections dir
	if err := os.MkdirAll(filepath.Join(dir, "collections", "users"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws := SharedWorkspace{Path: dir, Collaborators: []Collaborator{{User: "bob", Role: "editor"}}}
	if err := Save(DefaultPath(dir), ws); err != nil {
		t.Fatal(err)
	}
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/workspace", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var out map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out["health"] != "ok" {
		t.Fatalf("unexpected %v", out)
	}
	colls, ok := out["collections"].([]any)
	if !ok || len(colls) != 1 {
		t.Fatalf("expected 1 collection, got %v", out["collections"])
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	srv := NewServer(t.TempDir())
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	// health allows only GET via our handler? Actually we check method for collab/workspace but health currently allows any, but we test collab
	req2 := httptest.NewRequest(http.MethodPost, "/collab", nil)
	w2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w2, req2)
	if w2.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", w2.Code)
	}
}
