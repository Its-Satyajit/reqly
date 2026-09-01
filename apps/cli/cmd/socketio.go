// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/socketio"
)

var socketioCmd = &cobra.Command{
	Use:   "socketio",
	Short: "Socket.IO client subcommands",
	Long:  "Connect and emit events over Socket.IO (Engine.IO).",
}

var (
	socketioNamespace string
	socketioEvent     string
	socketioData      string
	socketioJSON      bool
)

var socketioConnectCmd = &cobra.Command{
	Use:   "connect <url> [--namespace /ns]",
	Short: "Connect and listen for Socket.IO events",
	Long:  `Connect to a Socket.IO server and stream events. Prints handshake events (connect, welcome) immediately, then streams incoming events until interrupted.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		opts := socketio.Options{
			Namespace: socketioNamespace,
		}
		fmt.Fprintf(cmd.OutOrStdout(), "connected to %s (namespace %s)\n", rawURL, opts.Namespace)
		return socketio.Connect(cmd.Context(), rawURL, func(ev socketio.Event) error {
			if socketioJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(ev)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %v\n", ev.Namespace, ev.Name, ev.Data)
			return nil
		}, opts)
	},
}

var socketioEmitCmd = &cobra.Command{
	Use:   "emit <url> --event <event> --data '<data>'",
	Short: "Emit a Socket.IO event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		opts := socketio.Options{
			Namespace: socketioNamespace,
		}
		var payload any = socketioData
		if err := json.Unmarshal([]byte(socketioData), &payload); err != nil {
			payload = socketioData
		}
		if err := socketio.Emit(cmd.Context(), rawURL, socketioEvent, payload, opts); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "emitted event %q to %s\n", socketioEvent, rawURL)
		return nil
	},
}

func init() {
	socketioConnectCmd.Flags().StringVar(&socketioNamespace, "namespace", "/", "Socket.IO namespace")
	socketioConnectCmd.Flags().BoolVar(&socketioJSON, "json", false, "output events as JSON")

	socketioEmitCmd.Flags().StringVar(&socketioNamespace, "namespace", "/", "Socket.IO namespace")
	socketioEmitCmd.Flags().StringVar(&socketioEvent, "event", "", "event name to emit")
	socketioEmitCmd.Flags().StringVar(&socketioData, "data", "", "event payload JSON or string")

	socketioCmd.AddCommand(socketioConnectCmd, socketioEmitCmd)
	rootCmd.AddCommand(socketioCmd)
}
