// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"net"
	"net/http"
	"testing"
)

func TestSocketIOCmd_Emit(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	go srv.Serve(ln)
	defer srv.Close()
	url := "http://" + ln.Addr().String()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"socketio", "emit", url, "--event", "ping", "--data", `{"msg":"hello"}`})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing socketio emit: %v", err)
	}
}

func TestSocketIOCmd_Connect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	go srv.Serve(ln)
	defer srv.Close()
	url := "http://" + ln.Addr().String()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"socketio", "connect", url})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing socketio connect: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("connected to")) {
		t.Fatalf("expected handshake output, got %q", buf.String())
	}
}
