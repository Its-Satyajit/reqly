// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"net"
	"testing"
)

func TestMQTTCmd_Pub(t *testing.T) {
	// Need a dummy listener so dial succeeds (R2).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	broker := ln.Addr().String()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"mqtt", "pub", broker, "--topic", "sensors/temp", "--message", `{"temp": 22.5}`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing mqtt pub: %v", err)
	}
}
