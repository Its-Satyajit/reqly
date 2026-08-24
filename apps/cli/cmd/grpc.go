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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	grpcclient "github.com/Its-Satyajit/reqly/internal/grpc"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// grpcCmd is the gRPC client command group (M43).
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "gRPC client",
	Long:  `gRPC calls with server-reflection service discovery; unary and server-streaming.`,
}

var grpcServicesJSON bool

var grpcServicesCmd = &cobra.Command{
	Use:   "services <host:port>",
	Short: "Discover services and methods via server reflection",
	Long: `List every service and method a gRPC server exposes via server reflection.

  reqly grpc services localhost:50051
  reqly grpc services localhost:50051 --json`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		services, err := grpcclient.Discover(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if grpcServicesJSON {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(services)
		}
		for _, svc := range services {
			fmt.Fprintf(out, "%s\n", svc.Name)
			for _, m := range svc.Methods {
				kind := "unary"
				if m.ServerStreaming {
					kind = "server-streaming"
				}
				fmt.Fprintf(out, "  %s (%s) — %s\n", m.FullName, kind, m.InputType)
			}
		}
		return nil
	},
}

func init() {
	grpcServicesCmd.Flags().BoolVar(&grpcServicesJSON, "json", false, "output machine-readable JSON")
	grpcInvokeCmd.Flags().StringVar(&grpcInvokeTimeout, "timeout", "", "override the call deadline (Go duration, e.g. 5s)")
	grpcCmd.AddCommand(grpcServicesCmd)
	grpcCmd.AddCommand(grpcInvokeCmd)
}

var grpcInvokeTimeout string

var grpcInvokeCmd = &cobra.Command{
	Use:   "invoke <request-file>",
	Short: "Invoke the gRPC call described by a request file",
	Long: `Send the unary gRPC call described by a request file's grpc: block.

  reqly grpc invoke requests/rpc.reqly.json
  reqly grpc invoke rpc.yaml --timeout 5s

Prints the response message as JSON. Exits non-zero on transport errors and
non-OK gRPC statuses.`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := requestfile.LoadFile(args[0])
		if err != nil {
			return err
		}
		g := f.Request.GRPC
		if g == nil || g.Service == "" || g.Method == "" {
			return fmt.Errorf("request file has no grpc block (grpc.service / grpc.method)")
		}
		if f.Request.URL == "" {
			return fmt.Errorf("request file requires a request.url (host:port of the gRPC server)")
		}

		timeout := 30 * time.Second
		if strings.TrimSpace(g.Timeout) != "" {
			parsed, perr := time.ParseDuration(g.Timeout)
			if perr != nil {
				return fmt.Errorf("invalid grpc.timeout %q: %w", g.Timeout, perr)
			}
			timeout = parsed
		}
		if grpcInvokeTimeout != "" {
			parsed, perr := time.ParseDuration(grpcInvokeTimeout)
			if perr != nil {
				return fmt.Errorf("invalid --timeout %q: %w", grpcInvokeTimeout, perr)
			}
			timeout = parsed
		}

		message := []byte{}
		if g.Message != nil {
			message, err = json.Marshal(g.Message)
			if err != nil {
				return fmt.Errorf("encode grpc.message: %w", err)
			}
		}

		md := map[string]string{}
		for _, h := range f.Request.Headers {
			if h.Key == "" {
				continue
			}
			md[h.Key] = h.Value
		}

		out := cmd.OutOrStdout()
		res, err := grpcclient.Invoke(cmd.Context(),
			grpcclient.Call{
				Target:     f.Request.URL,
				Service:    g.Service,
				Method:     g.Method,
				ProtoFiles: g.ProtoFiles,
			},
			message,
			grpcclient.InvokeOptions{Metadata: md, Timeout: timeout},
		)
		if err != nil {
			return err
		}
		if !res.OK {
			return fmt.Errorf("gRPC status %s (%d): %s", res.CodeName, res.Code, res.StatusMessage)
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, res.MessageJSON, "", "  "); err != nil {
			return fmt.Errorf("format response: %w", err)
		}
		fmt.Fprintln(out, pretty.String())
		return nil
	},
}
