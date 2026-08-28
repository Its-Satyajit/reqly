// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"
)

func TestSocketIOCmd_Emit(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"socketio", "emit", "http://localhost:3000", "--event", "ping", "--data", `{"msg":"hello"}`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing socketio emit: %v", err)
	}
}
