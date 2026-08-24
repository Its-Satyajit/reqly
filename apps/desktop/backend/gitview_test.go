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

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitBridgeRepoService(t *testing.T) (*AppService, string) {
	t.Helper()
	svc, wsDir := newServiceInWorkspace(t)
	if err := os.WriteFile(filepath.Join(wsDir, "base.txt"), []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds := [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "t@t"},
		{"config", "user.name", "t"},
		{"add", "reqly.yaml", "base.txt"},
		{"commit", "-m", "init"},
	}
	for _, args := range cmds {
		c := exec.Command("git", args...)
		c.Dir = wsDir
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	return svc, wsDir
}

func TestGitStatusBridgeReportsBranch(t *testing.T) {
	svc, _ := gitBridgeRepoService(t)
	st, err := svc.GitStatus()
	if err != nil {
		t.Fatalf("GitStatus: %v", err)
	}
	if !st.RepoFound || st.Branch != "main" || !st.Clean {
		t.Fatalf("status = %+v", st)
	}
}

func TestGitStageCommitBridge(t *testing.T) {
	svc, wsDir := gitBridgeRepoService(t)
	file := filepath.Join(wsDir, "env.request.json")
	if err := os.WriteFile(file, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.GitStage([]string{"env.request.json"}); err != nil {
		t.Fatalf("GitStage: %v", err)
	}
	if err := svc.GitCommit("feat: add env request"); err != nil {
		t.Fatalf("GitCommit: %v", err)
	}

	st, err := svc.GitStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !st.Clean {
		t.Fatalf("expected clean repo after commit, got %+v", st.Files)
	}
}

func TestGitStatusWithoutWorkspace(t *testing.T) {
	t.Chdir(t.TempDir())
	svc := NewAppService()
	st, err := svc.GitStatus()
	if err != nil {
		t.Fatalf("GitStatus: %v", err)
	}
	if st.RepoFound {
		t.Fatal("RepoFound = true without workspace, want false")
	}
	if err := svc.GitCommit("x"); err == nil {
		t.Fatal("GitCommit without workspace should error")
	}
	if err := svc.GitStage([]string{"a"}); err == nil {
		t.Fatal("GitStage without workspace should error")
	}
}
