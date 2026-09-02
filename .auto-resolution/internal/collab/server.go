// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package collab

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
)

// Server is a self-hosted collaboration server for shared workspaces.
// It is local-only, zero telemetry, and serves workspace metadata over HTTP
// for team sharing. Auth is via RBAC (when configured) or open when no RBAC.
type Server struct {
	root string
	mux  *http.ServeMux
}

// NewServer returns a Server rooted at workspace root.
func NewServer(root string) *Server {
	s := &Server{root: root, mux: http.NewServeMux()}
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/collab", s.handleCollab)
	s.mux.HandleFunc("/workspace", s.handleWorkspace)
	return s
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCollab(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := DefaultPath(s.root)
	ws, err := Load(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ws)
}

func (s *Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Simple: return workspace path and collaborator count
	path := DefaultPath(s.root)
	ws, err := Load(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Also try to list workspace collections via filepath
	collectionsDir := filepath.Join(s.root, "collections")
	// Count collections (directories)
	var collections []string
	if entries, err := filepath.Glob(filepath.Join(collectionsDir, "*")); err == nil {
		for _, e := range entries {
			collections = append(collections, filepath.Base(e))
		}
	}
	resp := map[string]any{
		"path":          ws.Path,
		"collaborators": ws.Collaborators,
		"collections":   collections,
		"health":        "ok",
	}
	// Sanitize path to avoid leaking absolute
	if s.root != "" && strings.HasPrefix(ws.Path, s.root) {
		resp["path"] = strings.TrimPrefix(ws.Path, s.root)
		if resp["path"] == "" {
			resp["path"] = "/"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
