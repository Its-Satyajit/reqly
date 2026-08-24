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
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/core"
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
		services, err := grpcclient.Discover(cmd.Context(), args[0], grpcclient.Transport{
			TLS:           grpcTLSSkipVerify || grpcCAFile != "",
			TLSSkipVerify: grpcTLSSkipVerify,
			CAFile:        grpcCAFile,
		})
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
	grpcServicesCmd.Flags().BoolVar(&grpcTLSSkipVerify, "tls-skip-verify", false, "skip TLS certificate verification")
	grpcServicesCmd.Flags().StringVar(&grpcCAFile, "ca-file", "", "PEM CA bundle to trust")
	grpcInvokeCmd.Flags().StringVar(&grpcInvokeTimeout, "timeout", "", "override the call deadline (Go duration, e.g. 5s)")
	grpcInvokeCmd.Flags().StringVar(&grpcEnv, "env", "", "environment to use (REQLY_ENV wins)")
	grpcInvokeCmd.Flags().IntVar(&grpcMaxMessages, "max-messages", 0, "server-streaming: stop after n messages (0 = until the server closes)")
	grpcCmd.AddCommand(grpcServicesCmd)
	grpcCmd.AddCommand(grpcInvokeCmd)
}

var (
	grpcInvokeTimeout string
	grpcEnv           string
	grpcTLSSkipVerify bool
	grpcCAFile        string
	grpcMaxMessages   int
)

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
		baseDir := "."
		if requestfile.LooksLikeFile(args[0]) {
			baseDir = filepath.Dir(args[0])
		}
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

		if grpcInvokeTimeout != "" {
			if _, perr := time.ParseDuration(grpcInvokeTimeout); perr != nil {
				return fmt.Errorf("invalid --timeout %q: %w", grpcInvokeTimeout, perr)
			}
		}

		// One deep pipeline call (ADR 0025/0028): env precedence, variable
		// interpolation, masking, and history recording live in core.
		root := findWorkspaceRoot(baseDir)
		svc := core.NewRunService(root)
		defer svc.Close()

		out := cmd.OutOrStdout()
		streamed := false
		res, err := svc.RunGRPCStreamed(cmd.Context(), f.Request, core.RunRequestOptions{
			EnvFlag:  envFlag,
			FileEnv:  f.Environment,
			FileVars: f.VariablesSet(),
		}, func(ev grpcclient.StreamEvent) error {
			streamed = true
			fmt.Fprintln(out, string(ev.MessageJSON))
			if grpcMaxMessages > 0 && ev.Seq >= grpcMaxMessages {
				return core.ErrStopConsumption
			}
			return nil
		})
		if err != nil && !errors.Is(err, core.ErrStopConsumption) {
			return err
		}
		if errors.Is(err, core.ErrStopConsumption) {
			// Client-side cap reached; the streamed messages above are the
			// complete output by definition.
			return nil
		}
		r := res.Result
		if !r.OK {
			return fmt.Errorf("gRPC status %s (%d): %s", r.CodeName, r.Code, r.StatusMessage)
		}
		if streamed {
			return nil
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, r.MessageJSON, "", "  "); err != nil {
			return fmt.Errorf("format response: %w", err)
		}
		fmt.Fprintln(out, pretty.String())
		if res.Warning != "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning:", res.Warning)
		}
		return nil
	},
}
