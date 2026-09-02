// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"testing"
)

func TestVerifyToken_HS256(t *testing.T) {
	secret := []byte("secret-key")
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{"sub": "1234567890", "name": "Satyajit"}

	tokenStr, err := Sign(header, payload, secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	res, err := VerifyToken(tokenStr, secret, VerifyOptions{Algorithm: "HS256"})
	if err != nil {
		t.Fatalf("unexpected error during verification: %v", err)
	}
	if !res.Valid || !res.SignatureValid {
		t.Errorf("expected valid signature, got errors: %v", res.Errors)
	}

	// Test invalid secret
	resBad, _ := VerifyToken(tokenStr, []byte("wrong-secret"), VerifyOptions{Algorithm: "HS256"})
	if resBad.Valid || resBad.SignatureValid {
		t.Errorf("expected invalid signature with wrong secret")
	}
}
