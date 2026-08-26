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
	"path/filepath"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/environments"
)

func seedEnv(t *testing.T, wsDir, name, yaml string) {
	t.Helper()
	dir := filepath.Join(wsDir, "environments")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
}

const envDevYAML = `description: dev
variables:
  baseUrl: https://dev.example.com
  region: us-east-1
secrets:
  apiToken: dev-secret
`

const envProdYAML = `description: prod
variables:
  baseUrl: https://api.example.com
secrets:
  apiToken: prod-secret
`

func TestEnvDiffMasksSecretsAndReportsKeys(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedEnv(t, wsDir, "dev", envDevYAML)
	seedEnv(t, wsDir, "prod", envProdYAML)

	res, err := svc.EnvDiff("dev", "prod")
	if err != nil {
		t.Fatalf("EnvDiff: %v", err)
	}
	found := map[string]environments.KeyDiff{}
	for _, d := range res.Diffs {
		found[d.Name] = d
	}
	d, ok := found["baseUrl"]
	if !ok || d.Status != "changed" {
		t.Errorf("baseUrl diff = %+v", d)
	}
	sec, ok := found["apiToken"]
	if !ok {
		t.Fatal("apiToken missing from diff")
	}
	if sec.From != "[SECRET]" || sec.To != "[SECRET]" {
		t.Errorf("secret values leaked: %q / %q", sec.From, sec.To)
	}
}

func TestEnvDiffUnknownEnvErrors(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedEnv(t, wsDir, "dev", envDevYAML)
	if _, err := svc.EnvDiff("dev", "nope"); err == nil {
		t.Fatal("diff with unknown env succeeded")
	}
}

func TestEnvValidateReportsUndefinedVariables(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedEnv(t, wsDir, "dev", envDevYAML)
	collDir := filepath.Join(wsDir, "collections", "users")
	if err := os.MkdirAll(collDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(collDir, "reqly.yaml"), []byte("name: users\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reqFile := `{"version":"1","request":{"name":"list","method":"GET","url":"{{baseUrl}}/x?t={{region}}&z={{notDefined}}"}}`
	if err := os.WriteFile(filepath.Join(collDir, "list.json"), []byte(reqFile), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := svc.EnvValidate("dev")
	if err != nil {
		t.Fatalf("EnvValidate: %v", err)
	}
	sawUndefined := false
	for _, issue := range res.Issues {
		if issue.Message == `undefined variable "notDefined" referenced by a workspace request or test` {
			sawUndefined = true
		}
	}
	if !sawUndefined {
		t.Errorf("undefined-variable issue for notDefined not reported: %+v", res.Issues)
	}
}

func TestEnvCrossValidateFindsGaps(t *testing.T) {
	svc, wsDir := newServiceInWorkspace(t)
	seedEnv(t, wsDir, "dev", envDevYAML)
	seedEnv(t, wsDir, "prod", envProdYAML)

	gaps, err := svc.EnvCrossValidate()
	if err != nil {
		t.Fatalf("EnvCrossValidate: %v", err)
	}
	byKey := map[string]CrossEnvGap{}
	for _, g := range gaps {
		byKey[g.Key] = g
	}
	gap, ok := byKey["region"]
	if !ok {
		t.Fatalf("region gap missing: %+v", gaps)
	}
	if len(gap.MissingIn) != 1 || gap.MissingIn[0] != "prod" {
		t.Errorf("region MissingIn = %v, want [prod]", gap.MissingIn)
	}
}
