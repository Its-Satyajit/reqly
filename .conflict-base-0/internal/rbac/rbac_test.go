package rbac

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		rbac    RBAC
		wantErr bool
	}{
		{name: "default", rbac: DefaultRBAC(), wantErr: false},
		{name: "no roles", rbac: RBAC{}, wantErr: true},
		{name: "empty role name", rbac: RBAC{Roles: map[string]Role{"": {Name: "", Permissions: []string{"a"}}}}, wantErr: true},
		{name: "mismatched key", rbac: RBAC{Roles: map[string]Role{"admin": {Name: "other", Permissions: []string{"*"}}}}, wantErr: true},
		{name: "no permissions", rbac: RBAC{Roles: map[string]Role{"admin": {Name: "admin", Permissions: []string{}}}}, wantErr: true},
		{name: "unknown user role", rbac: RBAC{Roles: map[string]Role{"admin": {Name: "admin", Permissions: []string{"*"}}}, UserRoles: map[string]string{"bob": "missing"}}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.rbac)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestCan(t *testing.T) {
	r := DefaultRBAC()
	r.UserRoles["alice"] = "admin"
	r.UserRoles["bob"] = "viewer"
	if !Can(r, "alice", "any.action") {
		t.Fatalf("admin should allow any")
	}
	if !Can(r, "bob", "request.send") {
		t.Fatalf("viewer should allow request.send")
	}
	if Can(r, "bob", "workflow.run") {
		t.Fatalf("viewer should deny workflow.run")
	}
	// unknown user defaults to viewer
	if !Can(r, "unknown", "request.send") {
		t.Fatalf("unknown should default to viewer allow")
	}
	if Can(r, "unknown", "workflow.run") {
		t.Fatalf("unknown should deny workflow")
	}
}

func TestEnforce(t *testing.T) {
	r := DefaultRBAC()
	r.UserRoles["bob"] = "viewer"
	if err := Enforce(r, "bob", "request.send", "x"); err != nil {
		t.Fatalf("should allow: %v", err)
	}
	if err := Enforce(r, "bob", "workflow.run", "x"); err == nil {
		t.Fatalf("should deny")
	}
}

func TestListRoles(t *testing.T) {
	r := DefaultRBAC()
	roles := ListRoles(r)
	if len(roles) != 3 || roles[0] != "admin" {
		t.Fatalf("unexpected roles %v", roles)
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".reqly", "rbac.yaml")
	r := DefaultRBAC()
	r.UserRoles["alice"] = "admin"
	if err := Save(path, r); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if fi, err := os.Stat(path); err == nil && fi.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", fi.Mode().Perm())
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.UserRoles["alice"] != "admin" {
		t.Fatalf("unexpected %+v", loaded)
	}
	// Missing returns default
	missing, err := Load(filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if len(missing.Roles) != 3 {
		t.Fatalf("default should have 3 roles, got %+v", missing)
	}
}
