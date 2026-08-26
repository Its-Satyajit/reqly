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

package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func TestGrpcServicesListsMethods(t *testing.T) {
	srv := testsrv.Start(t)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "services", srv.Addr})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc services: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "reqly.test.v1.EchoService") {
		t.Errorf("output missing service:\n%s", text)
	}
	if !strings.Contains(text, "/reqly.test.v1.EchoService/Echo (unary)") {
		t.Errorf("output missing unary method:\n%s", text)
	}
	if !strings.Contains(text, "/reqly.test.v1.EchoService/StreamEcho (server-streaming)") {
		t.Errorf("output missing streaming method:\n%s", text)
	}
}

func TestGrpcServicesJSONOutput(t *testing.T) {
	srv := testsrv.Start(t)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "services", srv.Addr, "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc services --json: %v", err)
	}
	var services []struct {
		Name    string `json:"name"`
		Methods []struct {
			FullName        string `json:"fullName"`
			ServerStreaming bool   `json:"serverStreaming"`
		} `json:"methods"`
	}
	if err := json.Unmarshal(out.Bytes(), &services); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	found := false
	for _, s := range services {
		if s.Name == "reqly.test.v1.EchoService" && len(s.Methods) == 2 {
			found = true
		}
	}
	if !found {
		t.Errorf("EchoService with 2 methods missing from JSON: %s", out.String())
	}
}

func TestGrpcServicesReflectionDisabled(t *testing.T) {
	srv := testsrv.StartPlain(t)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "services", srv.Addr})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error against reflection-disabled server")
	}
}

func TestGrpcServicesUnreachable(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "services", "127.0.0.1:1"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for unreachable target")
	}
}
