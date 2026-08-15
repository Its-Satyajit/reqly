// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package scripting

import (
	"strings"
	"testing"
)

func TestSandboxGetSetVariable(t *testing.T) {
	store := map[string]string{"token": "abc"}
	s := NewSandbox(SandboxOptions{
		GetVariable: func(name string) (string, bool) { v, ok := store[name]; return v, ok },
		SetVariable: func(name, value string) { store[name] = value },
	})
	if err := s.Run(`reqly.setVariable("fresh", reqly.getVariable("token") + "!")`); err != nil {
		t.Fatal(err)
	}
	if store["fresh"] != "abc!" {
		t.Fatalf("fresh = %q", store["fresh"])
	}
}

func TestSandboxMissingVariableReturnsUndefined(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`var v = reqly.getVariable("nope"); if (v !== undefined) { throw "expected undefined" }`); err != nil {
		t.Fatal(err)
	}
	if err := s.Run(`reqly.hasVariable("nope") === false`); err != nil {
		t.Fatalf("expected run to succeed: %v", err)
	}
}

func TestSandboxSetVariableDisabled(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`reqly.setVariable("x", "1")`); err == nil {
		t.Fatal("expected error: setVariable requires a variable store")
	}
}

func TestSandboxRequestMutate(t *testing.T) {
	req := &requestView{
		Method:  "GET",
		URL:     "https://api.example.com/users",
		Headers: map[string]string{"Authorization": "Bearer old"},
		Body:    "",
	}
	s := NewSandbox(SandboxOptions{})
	s.BindRequest(req)

	src := `
		reqly.request.method = "POST";
		reqly.request.url = "https://api.example.com/items";
		reqly.request.headers.set("Authorization", "Bearer new");
		reqly.request.body = "{}";
	`
	if err := s.Run(src); err != nil {
		t.Fatal(err)
	}
	if req.Method != "POST" {
		t.Fatalf("method = %q", req.Method)
	}
	if req.URL != "https://api.example.com/items" {
		t.Fatalf("url = %q", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer new" {
		t.Fatalf("authorization = %q", req.Headers["Authorization"])
	}
	if req.Body != "{}" {
		t.Fatalf("body = %q", req.Body)
	}
}

func TestSandboxResponseRead(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	s.BindResponse(&responseView{
		Status:     200,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"ok":true}`,
	})
	if err := s.Run(`
		if (reqly.response.status !== 200) { throw "status" }
		if (reqly.response.headers["Content-Type"] !== "application/json") { throw "header" }
		if (reqly.response.body.indexOf("ok") === -1) { throw "body" }
	`); err != nil {
		t.Fatal(err)
	}
}

func TestSandboxRegistersTests(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	s.BindResponse(&responseView{Status: 200, Body: "hello"})
	if err := s.Run(`
		reqly.test("status is 200", function() { return reqly.response.status === 200; });
		reqly.test("body has hello", function() { return reqly.response.body.indexOf("hello") !== -1; });
	`); err != nil {
		t.Fatal(err)
	}
	tests := s.Tests()
	if len(tests) != 2 {
		t.Fatalf("expected 2 tests, got %d", len(tests))
	}
	if !tests[0].Fn() {
		t.Fatalf("test %q should pass", tests[0].Name)
	}
	if !tests[1].Fn() {
		t.Fatalf("test %q should pass", tests[1].Name)
	}
}

func TestSandboxTestFalse(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	s.BindResponse(&responseView{Status: 404})
	if err := s.Run(`reqly.test("fail", function() { return false; })`); err != nil {
		t.Fatal(err)
	}
	if s.Tests()[0].Fn() {
		t.Fatal("test should fail")
	}
}

func TestSandboxConsoleLogs(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`console.log("hello", 42)`); err != nil {
		t.Fatal(err)
	}
	logs := s.Logs()
	if len(logs) != 1 || !strings.Contains(logs[0], "hello") || !strings.Contains(logs[0], "42") {
		t.Fatalf("logs = %v", logs)
	}
}

func TestSandboxSyntaxError(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`const =`); err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestSandboxThrownError(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`throw new Error("boom")`); err == nil {
		t.Fatal("expected thrown error")
	}
}

func TestSandboxTestWithoutArgsPanics(t *testing.T) {
	s := NewSandbox(SandboxOptions{})
	if err := s.Run(`reqly.test("only name")`); err == nil {
		t.Fatal("expected error for missing fn argument")
	}
}
