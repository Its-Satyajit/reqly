package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/secrets"
)

func initRepoWithRemote(t *testing.T, remote string) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if remote != "" {
		if err := exec.Command("git", "-C", dir, "remote", "add", "origin", remote).Run(); err != nil {
			t.Fatalf("remote add: %v", err)
		}
	}
	return dir
}

func TestGitHubRemote(t *testing.T) {
	dir := initRepoWithRemote(t, "https://github.com/owner/repo.git")
	gh := NewGitHub(nil)
	url, err := gh.Remote(dir)
	if err != nil {
		t.Fatalf("Remote: %v", err)
	}
	if url == "" {
		t.Fatalf("want url")
	}
}

func TestGitHubTokenEnv(t *testing.T) {
	t.Setenv("REQLY_GITHUB_TOKEN", "tok123")
	gh := NewGitHub(nil)
	token, err := gh.Token()
	if err != nil || token != "tok123" {
		t.Fatalf("Token %q %v", token, err)
	}
}

func TestGitHubTokenStore(t *testing.T) {
	dir := t.TempDir()
	store, err := secrets.NewFileStore(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	_ = store.Set("github", "storeTok")
	gh := NewGitHub(store)
	t.Setenv("REQLY_GITHUB_TOKEN", "")
	token, err := gh.Token()
	if err != nil || token != "storeTok" {
		t.Fatalf("Token %q %v", token, err)
	}
	_ = os.Remove(filepath.Join(dir, "tokens.json"))
}
