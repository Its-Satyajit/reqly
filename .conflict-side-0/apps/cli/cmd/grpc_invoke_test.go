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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
)

func writeGrpcRequestFile(t *testing.T, dir, contents string) string {
	t.Helper()
	path := filepath.Join(dir, "rpc.reqly.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write request file: %v", err)
	}
	return path
}

func TestGrpcInvokeUnaryEcho(t *testing.T) {
	srv := testsrv.Start(t)
	dir := t.TempDir()
	path := writeGrpcRequestFile(t, dir, `{
		"name": "echo",
		"request": {
			"url": "`+srv.Addr+`",
			"headers": [{"key": "x-reqly", "value": "1"}],
			"grpc": {
				"service": "reqly.test.v1.EchoService",
				"method": "Echo",
				"message": {"text": "hello from file"},
				"timeout": "5s"
			}
		}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "invoke", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc invoke: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "hello from file") {
		t.Errorf("output missing echoed text:\n%s", out.String())
	}
}

func TestGrpcInvokeNonOKStatusFailsCommand(t *testing.T) {
	srv := testsrv.Start(t)
	path := writeGrpcRequestFile(t, t.TempDir(), `{
		"request": {
			"url": "`+srv.Addr+`",
			"grpc": {"service": "reqly.test.v1.FailingService", "method": "Boom"}
		}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "invoke", path})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit for non-OK status")
	}
	if !strings.Contains(err.Error(), "NotFound") || !strings.Contains(err.Error(), "thing missing") {
		t.Errorf("error should carry code and message: %v", err)
	}
}

func TestGrpcInvokeProtoFilesFallback(t *testing.T) {
	srv := testsrv.StartPlain(t)
	protoPath, _ := filepath.Abs("../../.." + "/internal/grpc/testsrv/testapi.proto")
	path := writeGrpcRequestFile(t, t.TempDir(), `{
		"request": {
			"url": "`+srv.Addr+`",
			"grpc": {
				"service": "reqly.test.v1.EchoService",
				"method": "Echo",
				"message": {"text": "via proto"},
				"protoFiles": ["`+protoPath+`"]
			}
		}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "invoke", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc invoke with protoFiles: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "via proto") {
		t.Errorf("output missing response:\n%s", out.String())
	}
}

func TestGrpcInvokeRequiresGrpcBlock(t *testing.T) {
	srv := testsrv.Start(t)
	path := writeGrpcRequestFile(t, t.TempDir(), `{
		"request": {"url": "`+srv.Addr+`", "method": "GET"}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "invoke", path})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for request file without grpc block")
	}
}

func TestGrpcInvokeStreamWithMaxMessages(t *testing.T) {
	srv := testsrv.Start(t)
	path := writeGrpcRequestFile(t, t.TempDir(), `{
		"request": {
			"url": "`+srv.Addr+`",
			"grpc": {"service": "reqly.test.v1.EchoService", "method": "StreamEcho"}
		}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	grpcMaxMessages = 2
	defer func() { grpcMaxMessages = 0 }()
	rootCmd.SetArgs([]string{"grpc", "invoke", path, "--max-messages", "2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc invoke --max-messages: %v\n%s", err, out.String())
	}
	lines := strings.Count(out.String(), "\n")
	if lines < 2 {
		t.Errorf("expected at least 2 streamed messages, got:\n%s", out.String())
	}
}

func TestGrpcInvokeStreamingMethodEndToEnd(t *testing.T) {
	srv := testsrv.Start(t)
	path := writeGrpcRequestFile(t, t.TempDir(), `{
		"request": {
			"url": "`+srv.Addr+`",
			"grpc": {"service": "reqly.test.v1.EchoService", "method": "StreamEcho"}
		}
	}`)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"grpc", "invoke", path})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("grpc invoke stream: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"sequence"`) {
		t.Errorf("output missing streamed messages:\n%s", out.String())
	}
}
