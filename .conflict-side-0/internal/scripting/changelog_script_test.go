package scripting

import (
	"testing"
)

func TestScripting_GenerateChangelog(t *testing.T) {
	var semver string
	sb := NewSandbox(SandboxOptions{
		SetVariable: func(key, val string) {
			if key == "semver" {
				semver = val
			}
		},
	})
	script := `
		const oldSpec = JSON.stringify({openapi: "3.0.0", paths: {"/users": {get: {responses: {"200": {description: "OK"}}}}}});
		const newSpec = JSON.stringify({openapi: "3.0.0", paths: {"/users": {get: {responses: {"200": {description: "OK"}}}, post: {responses: {"201": {description: "Created"}}}}}});
		const res = reqly.generateChangelog(oldSpec, newSpec);
		reqly.setVariable("semver", res.suggested_semver);
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if semver != "minor" {
		t.Fatalf("want semver 'minor', got %q", semver)
	}
}
