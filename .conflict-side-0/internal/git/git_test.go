package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	// Configure user for commit
	_ = exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", dir, "config", "user.name", "test").Run()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "add", "a.txt").Run(); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "init").Run(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return dir
}

func TestStatusClean(t *testing.T) {
	dir := initRepo(t)
	lines, err := Status(dir)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("want clean got %v", lines)
	}
}

func TestLog(t *testing.T) {
	dir := initRepo(t)
	lines, err := Log(dir, 10, 0)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("want log lines")
	}
}

func TestCommit(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := Commit(dir, "add b", []string{"b.txt"}); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	lines, _ := Log(dir, 10, 0)
	if len(lines) != 2 {
		t.Fatalf("want 2 commits got %v", lines)
	}
}
