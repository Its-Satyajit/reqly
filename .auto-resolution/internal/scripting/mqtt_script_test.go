// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"testing"
)

func TestSandbox_MQTT(t *testing.T) {
	sb := NewSandbox(SandboxOptions{})
	script := `
reqly.test("mqtt sandbox binding check", function() {
    return typeof reqly.mqtt === "object" && typeof reqly.mqtt.publish === "function";
});
`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 || !tests[0].Fn() {
		t.Errorf("expected mqtt sandbox binding test to pass")
	}
}
