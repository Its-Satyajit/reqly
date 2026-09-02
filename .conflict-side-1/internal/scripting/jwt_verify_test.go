// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

func TestSandbox_VerifyJWT(t *testing.T) {
	secret := []byte("test-secret")
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{"sub": "123"}
	token, err := jwt.Sign(header, payload, secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	sb := NewSandbox(SandboxOptions{})
	script := `reqly.test("verify jwt token", function() { return reqly.verifyJWT("` + token + `", "test-secret"); });`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 || !tests[0].Fn() {
		t.Errorf("expected verifyJWT sandbox test to pass")
	}
}
