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
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/sse"
)

var (
	sseHeaderFlags []string
	sseTimeout     time.Duration
	sseCount       int
)

var sseCmd = &cobra.Command{
	Use:   "sse <url>",
	Short: "Stream Server-Sent Events from an endpoint",
	Long: `Connect to a Server-Sent Events endpoint and stream events to stdout.

  reqly sse https://reqly-test-api.vercel.app/events

Each event prints as "event <name> (<id>)" followed by its data. Use --count
to stop after a fixed number of events, or Ctrl-C to disconnect.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headers, err := parseHeaders(sseHeaderFlags)
		if err != nil {
			return err
		}

		httpHeader := map[string][]string{}
		for _, h := range headers {
			httpHeader[h.Key] = append(httpHeader[h.Key], h.Value)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		if sseTimeout > 0 {
			var c context.CancelFunc
			ctx, c = context.WithTimeout(ctx, sseTimeout)
			defer c()
		}

		client := sse.NewClient(args[0], sse.WithHeaders(httpHeader))
		if err := client.Connect(ctx); err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "streaming from %s (Ctrl-C to stop)\n", args[0])

		count := 0
		for {
			ev, err := client.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				return fmt.Errorf("stream ended: %w", err)
			}
			count++
			if ev.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "event %s", ev.Name)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), "event message")
			}
			if ev.ID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), " (%s)", ev.ID)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), ev.Data)
			if ev.Retry > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "retry %v\n", ev.Retry)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if sseCount > 0 && count >= sseCount {
				break
			}
		}

		_ = client.Close()
		fmt.Fprintf(cmd.OutOrStdout(), "received %d event(s)\n", count)
		return nil
	},
}

func init() {
	sseCmd.Flags().StringArrayVarP(&sseHeaderFlags, "header", "H", nil, "request header in 'Key: Value' form (repeatable)")
	sseCmd.Flags().DurationVarP(&sseTimeout, "timeout", "t", 0, "maximum stream duration (0 = unlimited)")
	sseCmd.Flags().IntVarP(&sseCount, "count", "c", 0, "stop after N events (0 = unlimited)")
}
