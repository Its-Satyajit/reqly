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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	grpcclient "github.com/Its-Satyajit/reqly/internal/grpc"
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
	grpcCmd.AddCommand(grpcServicesCmd)
}
