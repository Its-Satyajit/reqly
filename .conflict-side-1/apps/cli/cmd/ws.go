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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/websocket"
)

// lockedWriter serializes writes to an underlying writer, safe for concurrent
// use from multiple goroutines (the incoming-frame goroutine and the stdin loop).
type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

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

		// Incoming frames are printed as they arrive. Both this goroutine and
		// the stdin loop below write to the same writer, so serialize them.
		out := &lockedWriter{w: cmd.OutOrStdout()}
		done := make(chan error, 1)
		go func() {
			for {
				msg, err := client.Receive(ctx)
				if err != nil {
					done <- err
					return
				}
				fmt.Fprintf(out, "[%s] %s\n", time.Now().Format("15:04:05"), msg.Data)
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
			fmt.Fprintf(out, "> %s\n", line)
		}

		select {
		case <-ctx.Done():
		case err := <-done:
			if err != nil && err.Error() != "EOF" && !strings.Contains(err.Error(), "closed") {
				fmt.Fprintf(cmd.ErrOrStderr(), "connection closed: %v\n", err)
			}
		}

		_ = client.Close(context.Background())
		fmt.Fprintf(out, "\nclosed")
		return nil
	},
}

func init() {
	wsCmd.Flags().StringArrayVarP(&wsHeaderFlags, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	wsCmd.Flags().DurationVarP(&wsTimeout, "timeout", "t", 30*time.Second, "maximum connection time")
}
