// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"net/http"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestCustomApplySetsHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/", nil)
	cfg := map[string]string{"header": "X-Custom-Auth", "value": "secret123"}
	if err := Apply(req, "custom", cfg, variables.NewSet()); err != nil {
		t.Fatalf("Apply custom: %v", err)
	}
	if got := req.Header.Get("X-Custom-Auth"); got != "secret123" {
		t.Fatalf("header = %q, want secret123", got)
	}
}

func TestCustomApplyRequiresHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/", nil)
	if err := Apply(req, "custom", map[string]string{"value": "v"}, variables.NewSet()); err == nil {
		t.Fatal("expected error for missing header")
	}
	if err := Apply(req, "custom", map[string]string{"header": "X-H", "value": ""}, variables.NewSet()); err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestCustomApplyInterpolation(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeGlobal, "tok", "myval")
	cfg := map[string]string{"header": "X-Token", "value": "{{tok}}"}
	if err := Apply(req, "custom", cfg, vars); err != nil {
		t.Fatalf("interpolated Apply: %v", err)
	}
	if got := req.Header.Get("X-Token"); got != "myval" {
		t.Fatalf("interpolated header = %q, want myval", got)
	}
}
