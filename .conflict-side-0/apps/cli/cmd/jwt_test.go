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
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	jwtpkg "github.com/Its-Satyajit/reqly/internal/jwt"
)

func encodeSeg(v any) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}

func makeJWTForCLI(header, payload map[string]any, sig string) string {
	return encodeSeg(header) + "." + encodeSeg(payload) + "." + sig
}

func TestJWTDecode_Pretty(t *testing.T) {
	jwtpkg.Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { jwtpkg.Now = time.Now })
	token := makeJWTForCLI(map[string]any{"alg": "HS256"}, map[string]any{"sub": "u1", "exp": float64(2000)}, "sig")
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"jwt", "decode", token})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Header:") || !strings.Contains(got, "HS256") {
		t.Fatalf("pretty header: got %q", got)
	}
	if !strings.Contains(got, "Payload:") || !strings.Contains(got, "u1") {
		t.Fatalf("pretty payload: got %q", got)
	}
	if !strings.Contains(got, "Expiry:") {
		t.Fatalf("expiry line: got %q", got)
	}
}

func TestJWTDecode_JSON(t *testing.T) {
	jwtpkg.Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { jwtpkg.Now = time.Now })
	token := makeJWTForCLI(map[string]any{"alg": "HS256"}, map[string]any{"sub": "u1"}, "sig")
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"jwt", "decode", token, "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		t.Fatalf("json: %v\nout=%q", err, out.String())
	}
	if _, ok := m["header"]; !ok {
		t.Fatalf("header missing: %v", m)
	}
	if _, ok := m["payload"]; !ok {
		t.Fatalf("payload missing: %v", m)
	}
}

func TestJWTDecode_Stdin(t *testing.T) {
	jwtpkg.Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { jwtpkg.Now = time.Now })
	token := makeJWTForCLI(map[string]any{"alg": "HS256"}, map[string]any{"sub": "u1"}, "sig")
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetIn(strings.NewReader(token))
	rootCmd.SetArgs([]string{"jwt", "decode", "-"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute stdin: %v", err)
	}
	if !strings.Contains(out.String(), "u1") {
		t.Fatalf("stdin payload: got %q", out.String())
	}
}

func TestJWTDecode_BearerPrefix(t *testing.T) {
	jwtpkg.Now = func() time.Time { return time.Unix(1000, 0).UTC() }
	t.Cleanup(func() { jwtpkg.Now = time.Now })
	token := makeJWTForCLI(map[string]any{"alg": "HS256"}, map[string]any{"sub": "u1"}, "sig")
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"jwt", "decode", "Bearer " + token})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("bearer: %v", err)
	}
	if !strings.Contains(out.String(), "u1") {
		t.Fatalf("bearer payload: got %q", out.String())
	}
}

func TestJWTDecode_Malformed(t *testing.T) {
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"jwt", "decode", "bad.token"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error for malformed token")
	}
	combined := errBuf.String() + out.String() + err.Error()
	if !strings.Contains(combined, "expected 3 segments") {
		t.Fatalf("error message: got %q errBuf %q err %q", out.String(), errBuf.String(), err.Error())
	}
}

func TestJWTDecode_Help(t *testing.T) {
	var out, errBuf bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"jwt", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out.String(), "decode") {
		t.Fatalf("help should list decode: got %q", out.String())
	}
}
