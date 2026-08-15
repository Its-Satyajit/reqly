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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/websocket"
)

var (
	wsHeaderFlags []string
	wsTimeout     time.Duration
	// stdinReader is the input source for interactive messages. It is a
	// variable so tests can substitute an in-memory stream.
	stdinReader io.Reader = os.Stdin
)

var wsCmd = &cobra.Command{
	Use:   "ws <url>",
	Short: "Interact with a WebSocket endpoint",
	Long: `Connect to a WebSocket endpoint and exchange messages.

  reqly ws wss://echo.websocket.org

Messages typed on stdin are sent as text frames; incoming frames are printed
with a timestamp. Use Ctrl-D to close the connection cleanly.

A message can be piped in instead:

  echo '{"action":"ping"}' | reqly ws wss://echo.websocket.org`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headers, err := parseHeaders(wsHeaderFlags)
		if err != nil {
			return err
		}

		httpHeader := map[string][]string{}
		for _, h := range headers {
			httpHeader[h.Key] = append(httpHeader[h.Key], h.Value)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if wsTimeout > 0 {
			var c context.CancelFunc
			ctx, c = context.WithTimeout(ctx, wsTimeout)
			defer c()
		}

		client := websocket.NewClient(args[0], websocket.WithHeaders(httpHeader))
		if err := client.Dial(ctx); err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "connected to %s\n", args[0])

		// Incoming frames are printed as they arrive.
		done := make(chan error, 1)
		go func() {
			for {
				msg, err := client.Receive(ctx)
				if err != nil {
					done <- err
					return
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", time.Now().Format("15:04:05"), msg.Data)
			}
		}()

		// Lines from stdin are sent as text frames; EOF closes the connection.
		scanner := bufio.NewScanner(stdinReader)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if err := client.SendText(ctx, line); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "send: %v\n", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "> %s\n", line)
		}

		select {
		case <-ctx.Done():
		case err := <-done:
			if err != nil && err.Error() != "EOF" && !strings.Contains(err.Error(), "closed") {
				fmt.Fprintf(cmd.ErrOrStderr(), "connection closed: %v\n", err)
			}
		}

		_ = client.Close(context.Background())
		fmt.Fprintln(cmd.OutOrStdout(), "\nclosed")
		return nil
	},
}

func init() {
	wsCmd.Flags().StringArrayVarP(&wsHeaderFlags, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	wsCmd.Flags().DurationVarP(&wsTimeout, "timeout", "t", 30*time.Second, "maximum connection time")
}
