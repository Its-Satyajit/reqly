package scripting

import (
	"os/exec"
	"testing"
)

func TestScripting_DetectGitProvider(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "remote", "add", "origin", "https://gitlab.com/reqly/reqly.git").Run(); err != nil {
		t.Fatalf("remote add: %v", err)
	}

	var detected string
	sb := NewSandbox(SandboxOptions{
		SetVariable: func(key, val string) {
			if key == "provider" {
				detected = val
			}
		},
	})

	script := `
		const p = reqly.git.detectProvider("` + dir + `");
		if (p && p.name) {
			reqly.setVariable("provider", p.name);
		}
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if detected != "gitlab" {
		t.Fatalf("want provider 'gitlab', got %q", detected)
	}
}
