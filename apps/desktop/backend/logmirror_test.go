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
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestLogMirrorEmitsWarningsAndErrorsToFrontend(t *testing.T) {
	var mirrored []map[string]any
	orig := emitRunEvent
	emitRunEvent = func(name string, data any) {
		if name != goLogEventName {
			t.Errorf("event name = %q, want %q", name, goLogEventName)
		}
		mirrored = append(mirrored, data.(map[string]any))
	}
	t.Cleanup(func() { emitRunEvent = orig })

	logger := slog.New(newLogMirrorHandler(nil, slog.LevelWarn))
	logger.Info("chatty info") // below threshold — must not mirror
	logger.Warn("cache miss", "path", "users")
	logger.Error("send failed")

	if len(mirrored) != 2 {
		t.Fatalf("mirrored %d events, want 2", len(mirrored))
	}
	if mirrored[0]["level"] != "WARN" || mirrored[0]["message"] != "cache miss" {
		t.Errorf("first event = %v, want WARN cache miss", mirrored[0])
	}
	if mirrored[0]["path"] != "users" {
		t.Errorf("attrs missing from mirrored event: %v", mirrored[0])
	}
	if mirrored[1]["level"] != "ERROR" {
		t.Errorf("second event level = %v, want ERROR", mirrored[1])
	}
}

func TestLogMirrorKeepsTerminalOutput(t *testing.T) {
	var terminal bytes.Buffer
	inner := slog.NewTextHandler(&terminal, &slog.HandlerOptions{Level: slog.LevelDebug})

	logger := slog.New(newLogMirrorHandler(inner, slog.LevelWarn))
	emitRunEvent = func(string, any) {}
	t.Cleanup(func() { emitRunEvent = func(string, any) {} })

	logger.Warn("both surfaces see this")
	if !strings.Contains(terminal.String(), "both surfaces see this") {
		t.Errorf("terminal output lost: %q", terminal.String())
	}
}
