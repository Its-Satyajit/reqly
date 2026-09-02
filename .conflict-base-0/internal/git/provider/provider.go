// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

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

func getGitRemoteURL(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GitHub implements Provider for github.com via REQLY_GITHUB_TOKEN + secrets.Store.
type GitHub struct {
	store secrets.Store
}

func NewGitHub(store secrets.Store) *GitHub { return &GitHub{store: store} }

func (g *GitHub) Name() string { return "github" }

func (g *GitHub) Remote(dir string) (string, error) {
	url, err := getGitRemoteURL(dir)
	if err != nil {
		return "", err
	}
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

// GitLab implements Provider for gitlab.com via REQLY_GITLAB_TOKEN + secrets.Store.
type GitLab struct {
	store secrets.Store
}

func NewGitLab(store secrets.Store) *GitLab { return &GitLab{store: store} }

func (gl *GitLab) Name() string { return "gitlab" }

func (gl *GitLab) Remote(dir string) (string, error) {
	url, err := getGitRemoteURL(dir)
	if err != nil {
		return "", err
	}
	if !strings.Contains(url, "gitlab.com") {
		return "", fmt.Errorf("unsupported provider for %q", url)
	}
	return url, nil
}

func (gl *GitLab) Token() (string, error) {
	if token := os.Getenv("REQLY_GITLAB_TOKEN"); token != "" {
		return token, nil
	}
	if gl.store == nil {
		return "", fmt.Errorf("PAT required: set REQLY_GITLAB_TOKEN or login via reqly auth login --provider gitlab")
	}
	token, err := gl.store.Get("gitlab")
	if err != nil || token == "" {
		return "", fmt.Errorf("PAT required: set REQLY_GITLAB_TOKEN or login via reqly auth login --provider gitlab")
	}
	return token, nil
}

// Bitbucket implements Provider for bitbucket.org via REQLY_BITBUCKET_TOKEN + secrets.Store.
type Bitbucket struct {
	store secrets.Store
}

func NewBitbucket(store secrets.Store) *Bitbucket { return &Bitbucket{store: store} }

func (b *Bitbucket) Name() string { return "bitbucket" }

func (b *Bitbucket) Remote(dir string) (string, error) {
	url, err := getGitRemoteURL(dir)
	if err != nil {
		return "", err
	}
	if !strings.Contains(url, "bitbucket.org") {
		return "", fmt.Errorf("unsupported provider for %q", url)
	}
	return url, nil
}

func (b *Bitbucket) Token() (string, error) {
	if token := os.Getenv("REQLY_BITBUCKET_TOKEN"); token != "" {
		return token, nil
	}
	if b.store == nil {
		return "", fmt.Errorf("PAT required: set REQLY_BITBUCKET_TOKEN or login via reqly auth login --provider bitbucket")
	}
	token, err := b.store.Get("bitbucket")
	if err != nil || token == "" {
		return "", fmt.Errorf("PAT required: set REQLY_BITBUCKET_TOKEN or login via reqly auth login --provider bitbucket")
	}
	return token, nil
}

// AzureDevOps implements Provider for dev.azure.com / visualstudio.com via REQLY_AZURE_DEVOPS_TOKEN + secrets.Store.
type AzureDevOps struct {
	store secrets.Store
}

func NewAzureDevOps(store secrets.Store) *AzureDevOps { return &AzureDevOps{store: store} }

func (az *AzureDevOps) Name() string { return "azure-devops" }

func (az *AzureDevOps) Remote(dir string) (string, error) {
	url, err := getGitRemoteURL(dir)
	if err != nil {
		return "", err
	}
	if !strings.Contains(url, "dev.azure.com") && !strings.Contains(url, "visualstudio.com") {
		return "", fmt.Errorf("unsupported provider for %q", url)
	}
	return url, nil
}

func (az *AzureDevOps) Token() (string, error) {
	if token := os.Getenv("REQLY_AZURE_DEVOPS_TOKEN"); token != "" {
		return token, nil
	}
	if az.store == nil {
		return "", fmt.Errorf("PAT required: set REQLY_AZURE_DEVOPS_TOKEN or login via reqly auth login --provider azure-devops")
	}
	token, err := az.store.Get("azure-devops")
	if err != nil || token == "" {
		return "", fmt.Errorf("PAT required: set REQLY_AZURE_DEVOPS_TOKEN or login via reqly auth login --provider azure-devops")
	}
	return token, nil
}

// Detect returns the matching Git Provider (GitHub, GitLab, Bitbucket, Azure DevOps) based on remote URL.
func Detect(dir string, store secrets.Store) (Provider, error) {
	url, err := getGitRemoteURL(dir)
	if err != nil {
		return nil, err
	}

	if strings.Contains(url, "github.com") {
		return NewGitHub(store), nil
	}
	if strings.Contains(url, "gitlab.com") {
		return NewGitLab(store), nil
	}
	if strings.Contains(url, "bitbucket.org") {
		return NewBitbucket(store), nil
	}
	if strings.Contains(url, "dev.azure.com") || strings.Contains(url, "visualstudio.com") {
		return NewAzureDevOps(store), nil
	}

	return nil, fmt.Errorf("unrecognized git provider for remote url %q", url)
}

// ListSupportedProviders returns all supported provider identifiers.
func ListSupportedProviders() []string {
	return []string{"github", "gitlab", "bitbucket", "azure-devops"}
}
