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
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/mocking"
	"github.com/Its-Satyajit/reqly/internal/openapi"
)

var (
	mockHost         string
	mockPort         int
	mockDelay        time.Duration
	mockEvery        int
	mockScenarioFile string
)

var mockCmd = &cobra.Command{
	Use:   "mock [spec] [--scenario <scenario.yaml>]",
	Short: "Serve a mock API from an OpenAPI spec or stateful scenario",
	Long: `Start a local mock server generated from an OpenAPI 3.x spec or a multi-scenario
state machine definition.

  reqly mock openapi.yaml
  reqly mock --scenario scenario.yaml
  reqly mock openapi.yaml --scenario scenario.yaml --port 4010 --delay 200ms

Requests are matched against state machine scenarios and spec paths. Use --delay to add
artificial latency and --fail-every to simulate intermittent 500 errors. Ctrl-C stops the server.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var doc *openapi3.T
		if len(args) > 0 {
			var err error
			doc, err = openapi.LoadFile(args[0])
			if err != nil {
				return err
			}
		}

		opts := []mocking.Option{
			mocking.WithLogger(log.New(cmd.OutOrStdout(), "", log.LstdFlags)),
		}
		if mockDelay > 0 {
			opts = append(opts, mocking.WithDelay(mockDelay))
		}
		if mockEvery > 1 {
			opts = append(opts, mocking.WithFailureRate(mockEvery))
		}

		if mockScenarioFile != "" {
			data, err := os.ReadFile(mockScenarioFile)
			if err != nil {
				return fmt.Errorf("read scenario file: %w", err)
			}
			var sc mocking.Scenario
			if err := yaml.Unmarshal(data, &sc); err != nil {
				return fmt.Errorf("parse scenario file: %w", err)
			}
			sm := mocking.NewStateMachine(&sc)
			opts = append(opts, mocking.WithStateMachine(sm))
		}

		if doc == nil && mockScenarioFile == "" {
			return fmt.Errorf("either an OpenAPI spec file or --scenario <file> is required")
		}

		handler, err := mocking.NewServer(doc, opts...)
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		addr := net.JoinHostPort(mockHost, fmt.Sprintf("%d", mockPort))
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("mock: cannot listen on %s: %w", addr, err)
		}
		defer listener.Close()

		return serveMock(ctx, cmd, handler, listener)
	},
}

func init() {
	mockCmd.Flags().StringVar(&mockHost, "host", "127.0.0.1", "interface to listen on")
	mockCmd.Flags().IntVarP(&mockPort, "port", "p", 4010, "port to listen on")
	mockCmd.Flags().DurationVar(&mockDelay, "delay", 0, "artificial latency before every response (e.g. 250ms)")
	mockCmd.Flags().IntVar(&mockEvery, "fail-every", 0, "simulate a 500 error on every Nth request (0 disables)")
	mockCmd.Flags().StringVar(&mockScenarioFile, "scenario", "", "path to stateful scenario YAML/JSON file")
}

// serveMock serves handler on the bound listener and blocks until the context
// serveMock serves handler on listener until the context is cancelled or serving fails.
// It gracefully shuts down the HTTP server when the context is cancelled. Returns a
// serving error, or nil after graceful shutdown.
func serveMock(ctx context.Context, cmd *cobra.Command, handler http.Handler, listener net.Listener) error {
	url := fmt.Sprintf("http://%s", listener.Addr().String())
	fmt.Fprintf(cmd.OutOrStdout(), "mock server listening on %s (Ctrl-C to stop)\n", url)

	server := &http.Server{Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		fmt.Fprintln(cmd.OutOrStdout(), "\nmock server stopped")
		return nil
	}
}
