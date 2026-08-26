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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/docs"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate API documentation",
	Long:  `Generate Markdown docs for a workspace.`,
}

var docsGenerateCmd = &cobra.Command{
	Use:   "generate [src] --out <dir> [--env <name>]",
	Short: "Generate Markdown docs to a directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := "."
		if len(args) > 0 {
			src = args[0]
		}
		out, _ := cmd.Flags().GetString("out")
		if out == "" {
			return fmt.Errorf("--out <dir> is required")
		}
		env, _ := cmd.Flags().GetString("env")
		ws, err := collections.LoadWorkspace(src)
		if err != nil {
			return err
		}
		if err := docs.Generate(out, ws, env); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated docs to %s (%d collections)\n", out, len(ws.Collections))
		return nil
	},
}

var docsOut string
var docsEnv string

func init() {
	docsCmd.AddCommand(docsGenerateCmd)
	docsGenerateCmd.Flags().StringVar(&docsOut, "out", "", "output directory")
	docsGenerateCmd.Flags().StringVar(&docsEnv, "env", "", "environment for resolved curl examples")
}
