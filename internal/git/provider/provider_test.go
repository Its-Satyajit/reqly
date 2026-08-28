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

func TestGitLabProvider(t *testing.T) {
	dir := initRepoWithRemote(t, "https://gitlab.com/group/project.git")
	gl := NewGitLab(nil)
	url, err := gl.Remote(dir)
	if err != nil {
		t.Fatalf("GitLab Remote: %v", err)
	}
	if url == "" {
		t.Fatalf("want gitlab url")
	}

	t.Setenv("REQLY_GITLAB_TOKEN", "gl_tok_123")
	tok, err := gl.Token()
	if err != nil || tok != "gl_tok_123" {
		t.Fatalf("want gl_tok_123, got %q (%v)", tok, err)
	}
}

func TestBitbucketProvider(t *testing.T) {
	dir := initRepoWithRemote(t, "https://bitbucket.org/team/repo.git")
	bb := NewBitbucket(nil)
	url, err := bb.Remote(dir)
	if err != nil {
		t.Fatalf("Bitbucket Remote: %v", err)
	}
	if url == "" {
		t.Fatalf("want bitbucket url")
	}

	t.Setenv("REQLY_BITBUCKET_TOKEN", "bb_tok_456")
	tok, err := bb.Token()
	if err != nil || tok != "bb_tok_456" {
		t.Fatalf("want bb_tok_456, got %q (%v)", tok, err)
	}
}

func TestAzureDevOpsProvider(t *testing.T) {
	dir := initRepoWithRemote(t, "https://dev.azure.com/org/proj/_git/repo")
	az := NewAzureDevOps(nil)
	url, err := az.Remote(dir)
	if err != nil {
		t.Fatalf("Azure DevOps Remote: %v", err)
	}
	if url == "" {
		t.Fatalf("want azure devops url")
	}

	t.Setenv("REQLY_AZURE_DEVOPS_TOKEN", "az_tok_789")
	tok, err := az.Token()
	if err != nil || tok != "az_tok_789" {
		t.Fatalf("want az_tok_789, got %q (%v)", tok, err)
	}
}

func TestDetectAllProviders(t *testing.T) {
	cases := []struct {
		url      string
		expected string
	}{
		{"https://github.com/foo/bar.git", "github"},
		{"git@gitlab.com:foo/bar.git", "gitlab"},
		{"https://bitbucket.org/foo/bar.git", "bitbucket"},
		{"https://dev.azure.com/foo/bar/_git/baz", "azure-devops"},
		{"https://foo.visualstudio.com/bar/_git/baz", "azure-devops"},
	}

	for _, tc := range cases {
		dir := initRepoWithRemote(t, tc.url)
		p, err := Detect(dir, nil)
		if err != nil {
			t.Fatalf("Detect %s error: %v", tc.url, err)
		}
		if p.Name() != tc.expected {
			t.Errorf("Detect %s: expected %q, got %q", tc.url, tc.expected, p.Name())
		}
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
