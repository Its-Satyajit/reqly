// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

func TestSSO_Desktop(t *testing.T) {
	secret := []byte("test-secret-32-bytes-long-for-hs256")
	issuer := "https://auth.example.com"
	clientID := "reqly"
	header := map[string]any{"alg": "HS256"}
	claims := map[string]any{"iss": issuer, "sub": "bob"}
	token, err := jwt.Sign(header, claims, secret)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	svc := NewAppService()
	if err := svc.SSOValidate(issuer, clientID, token, string(secret)); err != nil {
		t.Fatalf("SSOValidate: %v", err)
	}
	if err := svc.SSOValidate(issuer, clientID, token, "wrong"); err == nil {
		t.Fatalf("should fail with wrong secret")
	}
}

func TestSCIM_Desktop(t *testing.T) {
	svc := NewAppService()
	u, err := svc.SCIMCreateUser("alice", "alice@example.com")
	if err != nil {
		t.Fatalf("SCIMCreateUser: %v", err)
	}
	if u.UserName != "alice" {
		t.Fatalf("unexpected %+v", u)
	}
	list := svc.SCIMListUsers()
	// New store per call, so list is empty (in-memory)
	if len(list) != 0 {
		t.Fatalf("expected 0, got %d", len(list))
	}
}
