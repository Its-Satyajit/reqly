package provider

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/secrets"
)

// Provider is a git host integration.
type Provider interface {
	Name() string
	Remote(dir string) (string, error)
	Token() (string, error)
}

// GitHub implements Provider for github.com via REQLY_GITHUB_TOKEN + secrets.Store.
type GitHub struct {
	store secrets.Store
}

func NewGitHub(store secrets.Store) *GitHub { return &GitHub{store: store} }

func (g *GitHub) Name() string { return "github" }

func (g *GitHub) Remote(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote: %w", err)
	}
	url := strings.TrimSpace(string(out))
	if !strings.Contains(url, "github.com") {
		return "", fmt.Errorf("unsupported provider for %q", url)
	}
	return url, nil
}

func (g *GitHub) Token() (string, error) {
	if token := os.Getenv("REQLY_GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if g.store == nil {
		return "", fmt.Errorf("PAT required: set REQLY_GITHUB_TOKEN or login via reqly auth login --provider github")
	}
	token, err := g.store.Get("github")
	if err != nil || token == "" {
		return "", fmt.Errorf("PAT required: set REQLY_GITHUB_TOKEN or login via reqly auth login --provider github")
	}
	return token, nil
}

// Detect returns GitHub provider when remote is github.com.
func Detect(dir string, store secrets.Store) (Provider, error) {
	gh := NewGitHub(store)
	if _, err := gh.Remote(dir); err != nil {
		return nil, err
	}
	return gh, nil
}
