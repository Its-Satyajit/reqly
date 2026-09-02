// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunCmd_Timeline(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer ts.Close()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"run", ts.URL, "--timeline"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error running reqly run --timeline: %v", err)
	}

	if !strings.Contains(buf.String(), "TIMELINE WATERFALL") {
		t.Errorf("expected TIMELINE WATERFALL header in output, got:\n%s", buf.String())
	}
}
