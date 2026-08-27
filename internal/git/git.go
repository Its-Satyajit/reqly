// Package git provides thin wrappers over git CLI for the Git GUI (status/diff/log/commit).
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Status returns porcelain status lines for dir (empty when clean, error when not a repo).
func Status(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git status: %w: %s", err, out.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

// Diff returns git diff (staged when staged true) for dir.
func Diff(dir string, staged bool) (string, error) {
	args := []string{"-C", dir, "diff"}
	if staged {
		args = []string{"-C", dir, "diff", "--staged"}
	}
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff: %w: %s", err, out.String())
	}
	return out.String(), nil
}

// Log returns recent oneline log lines (limit, offset) for dir.
func Log(dir string, limit, offset int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	args := []string{"-C", dir, "log", "--oneline", "-n", fmt.Sprintf("%d", limit), "--skip", fmt.Sprintf("%d", offset)}
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git log: %w: %s", err, out.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

// Commit stages files and commits with message in dir.
func Commit(dir, message string, files []string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message is required")
	}
	if len(files) > 0 {
		args := append([]string{"-C", dir, "add", "--"}, files...)
		cmd := exec.Command("git", args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git add: %w: %s", err, out.String())
		}
	}
	cmd := exec.Command("git", "-C", dir, "commit", "-m", message)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, out.String())
	}
	return nil
}
