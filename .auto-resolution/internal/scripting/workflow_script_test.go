package scripting

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScripting_Workflow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 42}`))
	}))
	defer srv.Close()

	var passed bool
	sb := NewSandbox(SandboxOptions{
		SetVariable: func(key, val string) {
			if key == "passed" && val == "true" {
				passed = true
			}
		},
	})

	script := `
		const wfYaml = 'name: Script Flow\nsteps:\n  - id: s1\n    name: S1\n    request:\n      method: GET\n      url: ` + srv.URL + `\n';
		const report = reqly.workflow.run(wfYaml);
		if (report && report.passed) {
			reqly.setVariable("passed", "true");
		}
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !passed {
		t.Fatalf("expected workflow to pass in scripting")
	}
}
