// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestEdgeGridApplyMissingKeys(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	s := edgeGridScheme{}
	if err := s.Apply(req, nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when edgegrid keys missing")
	}
	if err := s.Apply(req, map[string]string{"clientToken": "ct"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when clientSecret missing")
	} else if !strings.Contains(err.Error(), "clientSecret") {
		t.Fatalf("expected error to mention clientSecret, got %v", err)
	}
	if err := s.Apply(req, map[string]string{"clientToken": "ct", "clientSecret": "cs", "accessToken": "at"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when host missing")
	}
}

func TestEdgeGridApplyInterpolated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://akab-xxxx.luna.akamaiapis.net/test", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeRequest, "ct", "ctoken")
	vars.Set(variables.ScopeRequest, "cs", "csecret")
	vars.Set(variables.ScopeRequest, "at", "atoken")
	vars.Set(variables.ScopeRequest, "host", "akab-xxxx.luna.akamaiapis.net")
	err := edgeGridScheme{}.Apply(req, map[string]string{
		"clientToken":  "{{ct}}",
		"clientSecret": "{{cs}}",
		"accessToken":  "{{at}}",
		"host":         "{{host}}",
	}, vars)
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); !strings.HasPrefix(got, "EG1-HMAC-SHA256 ") {
		t.Fatalf("expected EG1-HMAC-SHA256 header, got %q", got)
	}
}

func TestEdgeGridApplyHeaderStructure(t *testing.T) {
	oldNow := edgeGridNow
	oldNonce := edgeGridNonce
	edgeGridNow = func() string { return "20200101T00:00:00+0000" }
	edgeGridNonce = func() string { return "nonce-123" }
	defer func() {
		edgeGridNow = oldNow
		edgeGridNonce = oldNonce
	}()

	req := httptest.NewRequest(http.MethodGet, "https://akab-xxxx.luna.akamaiapis.net/config", nil)
	err := edgeGridScheme{}.Apply(req, map[string]string{
		"clientToken":  "ct",
		"clientSecret": "cs",
		"accessToken":  "at",
		"host":         "akab-xxxx.luna.akamaiapis.net",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	auth := req.Header.Get("Authorization")
	for _, want := range []string{
		"client_token=ct",
		"access_token=at",
		"timestamp=20200101T00:00:00+0000",
		"nonce=nonce-123",
		"signature=",
	} {
		if !strings.Contains(auth, want) {
			t.Fatalf("header missing %q in %q", want, auth)
		}
	}
	if !strings.HasPrefix(auth, "EG1-HMAC-SHA256 ") {
		t.Fatalf("expected EG1 prefix, got %q", auth)
	}
}

func TestEdgeGridSecretKeys(t *testing.T) {
	s := edgeGridScheme{}
	keys := s.SecretKeys()
	if len(keys) != 2 || keys[0] != "clientSecret" || keys[1] != "accessToken" {
		t.Fatalf("SecretKeys: got %v", keys)
	}
	if got := MaskValues("edgegrid", map[string]string{"clientSecret": "cs", "accessToken": "at"}, variables.NewSet()); len(got) != 2 {
		t.Fatalf("MaskValues edgegrid: got %v", got)
	}
}
