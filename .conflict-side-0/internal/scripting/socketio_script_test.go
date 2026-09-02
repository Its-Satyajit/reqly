// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"testing"
)

func TestSandbox_SocketIO(t *testing.T) {
	sb := NewSandbox(SandboxOptions{})
	script := `
reqly.test("socketio sandbox binding check", function() {
    return typeof reqly.socketio === "object" && typeof reqly.socketio.emit === "function";
});
`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 || !tests[0].Fn() {
		t.Errorf("expected socketio sandbox binding test to pass")
	}
}
