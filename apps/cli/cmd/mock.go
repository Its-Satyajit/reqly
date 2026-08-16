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

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/mocking"
	"github.com/Its-Satyajit/reqly/internal/openapi"
)

var (
	mockHost  string
	mockPort  int
	mockDelay time.Duration
	mockEvery int
)

var mockCmd = &cobra.Command{
	Use:   "mock <spec>",
	Short: "Serve a mock API from an OpenAPI spec",
	Long: `Start a local mock server generated from an OpenAPI 3.x spec (JSON or YAML).

  reqly mock openapi.yaml
  reqly mock openapi.yaml --port 4010 --delay 200ms

Requests are matched against the spec paths and answered with generated example
bodies from the operation responses. Use --delay to add artificial latency and
--fail-every to simulate intermittent 500 errors. Ctrl-C stops the server.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := openapi.LoadFile(args[0])
		if err != nil {
			return err
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
}

// serveMock serves handler on the bound listener and blocks until the context
// is cancelled or the listener fails. Split out for testability.
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
