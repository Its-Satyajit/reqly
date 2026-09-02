package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		policy  Policy
		wantErr bool
	}{
		{name: "default", policy: DefaultPolicy(), wantErr: false},
		{name: "max steps valid", policy: Policy{MaxWorkflowSteps: 5}, wantErr: false},
		{name: "max steps negative", policy: Policy{MaxWorkflowSteps: -1}, wantErr: true},
		{name: "empty allowed action", policy: Policy{AllowedActions: []string{""}}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.policy)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestEnforce(t *testing.T) {
	p := Policy{AllowedActions: []string{"request.send", "workflow.run"}}
	if err := Enforce(p, "request.send", "users/list"); err != nil {
		t.Fatalf("unexpected deny: %v", err)
	}
	if err := Enforce(p, "theme.import", "my-theme"); err == nil {
		t.Fatalf("expected deny for theme.import")
	}
	// wildcard
	p2 := Policy{AllowedActions: []string{"*"}}
	if err := Enforce(p2, "any.action", "x"); err != nil {
		t.Fatalf("wildcard should allow: %v", err)
	}
	// empty allowed means allow all
	p3 := DefaultPolicy()
	if err := Enforce(p3, "any", "x"); err != nil {
		t.Fatalf("empty allowed should allow all: %v", err)
	}
}

func TestEnforceWorkflow(t *testing.T) {
	p := Policy{MaxWorkflowSteps: 2}
	if err := EnforceWorkflow(p, 2); err != nil {
		t.Fatalf("2 steps should pass: %v", err)
	}
	if err := EnforceWorkflow(p, 3); err == nil {
		t.Fatalf("3 steps should fail")
	}
	// 0 means no limit
	p2 := DefaultPolicy()
	if err := EnforceWorkflow(p2, 100); err != nil {
		t.Fatalf("no limit should pass: %v", err)
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".reqly", "policy.yaml")
	p := Policy{RequireAudit: true, MaxWorkflowSteps: 10, AllowedActions: []string{"request.send"}, RequireAuth: true}
	if err := Save(path, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Check perm 0600
	if fi, err := os.Stat(path); err == nil && fi.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", fi.Mode().Perm())
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.RequireAudit != true || loaded.MaxWorkflowSteps != 10 || loaded.RequireAuth != true {
		t.Fatalf("unexpected loaded %+v", loaded)
	}
	// Missing file returns default
	missing, err := Load(filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if missing.AllowCustomThemes != true {
		t.Fatalf("default should allow custom themes, got %+v", missing)
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	// Write invalid yaml with negative max steps
	if err := os.WriteFile(path, []byte("maxWorkflowSteps: -1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatalf("expected error for invalid policy")
	}
}
