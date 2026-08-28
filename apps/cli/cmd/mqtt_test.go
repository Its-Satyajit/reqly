// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"
)

func TestMQTTCmd_Pub(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"mqtt", "pub", "tcp://localhost:1883", "--topic", "sensors/temp", "--message", `{"temp": 22.5}`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing mqtt pub: %v", err)
	}
}
