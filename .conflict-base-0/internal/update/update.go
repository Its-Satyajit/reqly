// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultRepo    = "Its-Satyajit/reqly"
	DefaultBaseURL = "https://api.github.com"
)

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

// AssetInfo describes a downloadable binary asset.
type AssetInfo struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
	Size        int64  `json:"size"`
}

// ReleaseInfo holds the latest release metadata and update status.
type ReleaseInfo struct {
	CurrentVersion string      `json:"currentVersion"`
	LatestVersion  string      `json:"latestVersion"`
	HasUpdate      bool        `json:"hasUpdate"`
	ReleaseNotes   string      `json:"releaseNotes"`
	ReleaseURL     string      `json:"releaseUrl"`
	PublishedAt    time.Time   `json:"publishedAt"`
	Assets         []AssetInfo `json:"assets"`
}

// Checker handles querying release versions and downloading/applying updates.
type Checker struct {
	Repo    string
	BaseURL string
	Client  *http.Client
}

// NewChecker returns a Checker configured with defaults.
func NewChecker() *Checker {
	return &Checker{
		Repo:    DefaultRepo,
		BaseURL: DefaultBaseURL,
		Client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Check queries GitHub for the latest release and checks against currentVersion.
func (c *Checker) Check(ctx context.Context, currentVersion string) (*ReleaseInfo, error) {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	repo := c.Repo
	if repo == "" {
		repo = DefaultRepo
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", baseURL, repo)
	if strings.HasPrefix(baseURL, "http://") || strings.HasPrefix(baseURL, "https://") && !strings.Contains(baseURL, "api.github.com") {
		// When testing against custom mock server without full GitHub route
		if !strings.HasSuffix(baseURL, "/releases/latest") {
			url = baseURL
		}
	}

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create update check request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "reqly-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("release check failed with HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release json: %w", err)
	}

	assets := make([]AssetInfo, 0, len(rel.Assets))
	for _, a := range rel.Assets {
		assets = append(assets, AssetInfo{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
			Size:        a.Size,
		})
	}

	hasUpdate := IsNewer(currentVersion, rel.TagName)

	return &ReleaseInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  rel.TagName,
		HasUpdate:      hasUpdate,
		ReleaseNotes:   rel.Body,
		ReleaseURL:     rel.HTMLURL,
		PublishedAt:    rel.PublishedAt,
		Assets:         assets,
	}, nil
}

// FindAssetForPlatform finds the matching release asset for the given OS and architecture.
func FindAssetForPlatform(assets []AssetInfo, targetOS, targetArch string) (AssetInfo, bool) {
	ext := ""
	if targetOS == "windows" {
		ext = ".exe"
	}
	expectedName := fmt.Sprintf("reqly-%s-%s%s", targetOS, targetArch, ext)

	for _, a := range assets {
		if strings.EqualFold(a.Name, expectedName) {
			return a, true
		}
	}
	return AssetInfo{}, false
}

// IsNewer reports whether latest is newer than current using semver comparison.
func IsNewer(current, latest string) bool {
	cur := parseSemver(current)
	lat := parseSemver(latest)

	if cur == nil {
		return lat != nil
	}
	if lat == nil {
		return false
	}

	for i := 0; i < 3; i++ {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) []int {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 1 {
		return nil
	}

	res := make([]int, 3)
	for i := 0; i < 3 && i < len(parts); i++ {
		p := parts[i]
		if idx := strings.IndexAny(p, "-+"); idx != -1 {
			p = p[:idx]
		}
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		res[i] = num
	}
	return res
}

// ApplyBinaryUpdate downloads the new executable and replaces the target path atomically.
func (c *Checker) ApplyBinaryUpdate(ctx context.Context, downloadURL, targetPath string) error {
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download new binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download binary failed with status: %s", resp.Status)
	}

	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, "reqly-update-*")
	if err != nil {
		return fmt.Errorf("create temp binary: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write binary content: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("set executable permissions: %w", err)
	}

	// On Windows, you cannot replace a running executable directly; rename first to .old
	if runtime.GOOS == "windows" {
		oldPath := targetPath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(targetPath, oldPath); err != nil {
			return fmt.Errorf("rename existing binary: %w", err)
		}
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}
