// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"testing"
)

func TestSandbox_ReflectGRPC(t *testing.T) {
	sb := NewSandbox(SandboxOptions{})
	script := `
reqly.test("grpc reflect binding check", function() {
    return typeof reqly.reflectGRPC === "function";
});
`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 || !tests[0].Fn() {
		t.Errorf("expected reflectGRPC sandbox binding test to pass")
	}
}
