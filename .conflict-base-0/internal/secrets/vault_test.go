package secrets

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVaultStore_GetSetDelete(t *testing.T) {
	// In-memory Vault mock
	store := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check token
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		key := r.URL.Path
		// Simplified: extract key after /data/reqly/
		switch r.Method {
		case http.MethodGet:
			// Expect /v1/secret/data/reqly/<key>
			if v, ok := store[key]; ok {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"data": map[string]string{"value": v},
					},
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			data, _ := payload["data"].(map[string]any)
			val, _ := data["value"].(string)
			store[key] = val
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(store, key)
			w.WriteHeader(http.StatusNoContent)
		case "LIST":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"keys": []string{"a", "b"}},
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	vs, err := NewVaultStore(VaultConfig{Addr: srv.URL, Token: "test-token", Mount: "secret", Prefix: "reqly/"})
	if err != nil {
		t.Fatalf("NewVaultStore: %v", err)
	}
	// Get missing → NotExist
	if _, err := vs.Get("missing"); !isNotExist(err) {
		t.Fatalf("expected NotExist, got %v", err)
	}
	// Set and Get
	if err := vs.Set("mykey", "myval"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := vs.Get("mykey")
	if err != nil || got != "myval" {
		t.Fatalf("Get = %q, err %v, want myval", got, err)
	}
	// Keys
	keys, err := vs.Keys()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %v", keys)
	}
	// Delete
	if err := vs.Delete("mykey"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := vs.Get("mykey"); !isNotExist(err) {
		t.Fatalf("expected NotExist after delete, got %v", err)
	}
	// Delete absent is no error
	if err := vs.Delete("nope"); err != nil {
		t.Fatalf("Delete absent: %v", err)
	}
}

func TestVaultStore_Validation(t *testing.T) {
	if _, err := NewVaultStore(VaultConfig{Addr: "", Token: "t"}); err == nil {
		t.Fatalf("expected error for empty addr")
	}
	if _, err := NewVaultStore(VaultConfig{Addr: "http://x", Token: ""}); err == nil {
		t.Fatalf("expected error for empty token")
	}
	if _, err := NewVaultStore(VaultConfig{Addr: "http://x:8200", Token: "t"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func isNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
