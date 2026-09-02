// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"

	jwtpkg "github.com/Its-Satyajit/reqly/internal/jwt"
)

func TestJWTVerifyCmd(t *testing.T) {
	secret := []byte("secret-key")
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{"sub": "1234567890", "name": "Satyajit"}

	tokenStr, err := jwtpkg.Sign(header, payload, secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	resetJWTFlags()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"jwt", "verify", tokenStr, "--secret", "secret-key"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected failure executing reqly jwt verify: %v", err)
	}
}

func resetJWTFlags() {
	jwtVerifySecret = ""
	jwtSignSecret = ""
	jwtSignAlg = "HS256"
}
