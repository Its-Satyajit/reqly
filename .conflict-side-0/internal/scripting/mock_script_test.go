package scripting

import (
	"testing"
)

func TestScripting_MockStateMachine(t *testing.T) {
	var status float64
	var body string

	sb := NewSandbox(SandboxOptions{
		SetVariable: func(key, val string) {
			if key == "body" {
				body = val
			}
		},
	})

	script := `
		const scYaml = 'initial_state: init\nstates:\n  init:\n    transitions:\n      - method: GET\n        path: /ping\n        response:\n          status: 200\n          body: "pong"\n';
		const sm = reqly.mock.createStateMachine(scYaml);
		if (sm) {
			const res = reqly.mock.handle(sm, "GET", "/ping");
			if (res) {
				reqly.setVariable("body", res.body);
			}
		}
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if body != "pong" {
		t.Fatalf("want pong, got %q (status %f)", body, status)
	}
}
