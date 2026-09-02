package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	ws := &collections.Workspace{
		Root:   dir,
		Config: collections.Config{Name: "test-ws"},
		Collections: []*collections.Collection{
			{
				Name:   "users",
				Config: collections.Config{Name: "users"},
				Requests: []*collections.RequestEntry{
					{Name: "list", File: &requestfile.File{Request: request.Request{Method: "GET", URL: "https://api.example.com/users"}}},
					{Name: "create", File: &requestfile.File{Request: request.Request{Method: "POST", URL: "https://api.example.com/users", Headers: []request.Header{{Key: "Content-Type", Value: "application/json"}}, Body: `{"name":"Alice"}`}}},
				},
			},
		},
	}
	if err := Generate(out, ws, ""); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// index.md exists and lists collection
	data, err := os.ReadFile(filepath.Join(out, "index.md"))
	if err != nil {
		t.Fatalf("index.md: %v", err)
	}
	if !strings.Contains(string(data), "users") {
		t.Fatalf("index.md missing collection: %q", string(data))
	}
	// users.md exists and contains requests
	data, err = os.ReadFile(filepath.Join(out, "users.md"))
	if err != nil {
		t.Fatalf("users.md: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "GET") || !strings.Contains(s, "https://api.example.com/users") {
		t.Fatalf("users.md missing GET: %q", s)
	}
	if !strings.Contains(s, "curl") {
		t.Fatalf("users.md missing curl: %q", s)
	}
	if !strings.Contains(s, "Alice") {
		t.Fatalf("users.md missing body: %q", s)
	}
}
