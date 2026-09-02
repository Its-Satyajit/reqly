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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"1.2.0", "1.3.0", true},
		{"v1.2.0", "v1.3.0", true},
		{"1.2.0", "1.2.1", true},
		{"1.2.0", "2.0.0", true},
		{"1.2.0", "1.2.0", false},
		{"v1.2.0", "1.2.0", false},
		{"1.3.0", "1.2.0", false},
		{"2.0.0", "1.9.9", false},
		{"dev", "1.0.0", true},
		{"", "1.0.0", true},
	}

	for _, tt := range tests {
		got := IsNewer(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestCheck(t *testing.T) {
	release := githubRelease{
		TagName: "v1.5.0",
		Name:    "Reqly v1.5.0",
		Body:    "## Changes\n- Cool new feature",
		HTMLURL: "https://github.com/Its-Satyajit/reqly/releases/tag/v1.5.0",
		Assets: []githubAsset{
			{
				Name:               "reqly-linux-amd64",
				BrowserDownloadURL: "https://example.com/reqly-linux-amd64",
				Size:               1234567,
			},
			{
				Name:               "reqly-darwin-arm64",
				BrowserDownloadURL: "https://example.com/reqly-darwin-arm64",
				Size:               1234567,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	ctx := context.Background()
	checker := &Checker{
		BaseURL: srv.URL,
		Client:  srv.Client(),
	}

	info, err := checker.Check(ctx, "1.2.0")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !info.HasUpdate {
		t.Errorf("expected HasUpdate=true")
	}
	if info.LatestVersion != "v1.5.0" {
		t.Errorf("expected LatestVersion=v1.5.0, got %s", info.LatestVersion)
	}
	if info.ReleaseNotes != release.Body {
		t.Errorf("expected ReleaseNotes=%q, got %q", release.Body, info.ReleaseNotes)
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	assets := []AssetInfo{
		{Name: "reqly-linux-amd64", DownloadURL: "https://example.com/linux-amd64"},
		{Name: "reqly-linux-arm64", DownloadURL: "https://example.com/linux-arm64"},
		{Name: "reqly-darwin-amd64", DownloadURL: "https://example.com/darwin-amd64"},
		{Name: "reqly-darwin-arm64", DownloadURL: "https://example.com/darwin-arm64"},
		{Name: "reqly-windows-amd64.exe", DownloadURL: "https://example.com/windows-amd64"},
	}

	asset, found := FindAssetForPlatform(assets, "linux", "amd64")
	if !found || asset.DownloadURL != "https://example.com/linux-amd64" {
		t.Errorf("failed to match linux/amd64: %+v", asset)
	}

	asset, found = FindAssetForPlatform(assets, "darwin", "arm64")
	if !found || asset.DownloadURL != "https://example.com/darwin-arm64" {
		t.Errorf("failed to match darwin/arm64: %+v", asset)
	}

	asset, found = FindAssetForPlatform(assets, "windows", "amd64")
	if !found || asset.DownloadURL != "https://example.com/windows-amd64" {
		t.Errorf("failed to match windows/amd64: %+v", asset)
	}

	_, found = FindAssetForPlatform(assets, "freebsd", "386")
	if found {
		t.Errorf("expected not found for freebsd/386")
	}
}

func TestApplyBinaryUpdate(t *testing.T) {
	dir := t.TempDir()
	origPath := filepath.Join(dir, "mybin")
	if err := os.WriteFile(origPath, []byte("old-binary-content"), 0o755); err != nil {
		t.Fatal(err)
	}

	newContent := []byte("new-binary-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(newContent)
	}))
	defer srv.Close()

	ctx := context.Background()
	checker := &Checker{Client: srv.Client()}

	if err := checker.ApplyBinaryUpdate(ctx, srv.URL, origPath); err != nil {
		t.Fatalf("ApplyBinaryUpdate failed: %v", err)
	}

	data, err := os.ReadFile(origPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(newContent) {
		t.Fatalf("expected binary content %q, got %q", string(newContent), string(data))
	}

	info, err := os.Stat(origPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected binary to be executable, mode: %v", info.Mode())
	}
}
