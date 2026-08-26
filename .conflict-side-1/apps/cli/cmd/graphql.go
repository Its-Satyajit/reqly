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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/graphql"
)

var graphqlIntrospectHeaders []string
var graphqlIntrospectType string
var graphqlIntrospectJSON bool

var graphqlCmd = &cobra.Command{
	Use:   "graphql",
	Short: "GraphQL schema tooling",
	Long:  "Inspect GraphQL schemas via the standard introspection query.",
}

var graphqlIntrospectCmd = &cobra.Command{
	Use:   "introspect <url> [--header \"k: v\"]... [--type <Name>] [--json]",
	Short: "Introspect a GraphQL endpoint and print its schema",
	Long: `POST the standard introspection query to a GraphQL endpoint and render
the schema: root query/mutation/subscription fields first, then remaining
types alphabetically.

  reqly graphql introspect https://api.test/graphql
  reqly graphql introspect https://api.test/graphql --header "Authorization: Bearer t" --type User
  reqly graphql introspect https://api.test/graphql --json > schema.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := graphql.IntrospectOptions{}
		for _, h := range graphqlIntrospectHeaders {
			i := strings.Index(h, ":")
			if i <= 0 {
				return fmt.Errorf("invalid --header %q (want \"Name: value\")", h)
			}
			opts.Headers = append(opts.Headers, [2]string{
				strings.TrimSpace(h[:i]), strings.TrimSpace(h[i+1:]),
			})
		}
		schema, raw, err := graphql.Introspect(context.Background(), args[0], opts)
		if err != nil {
			return err
		}
		if graphqlIntrospectJSON {
			var pretty map[string]any
			if err := json.Unmarshal(raw, &pretty); err == nil {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(pretty)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		}
		fmt.Fprint(cmd.OutOrStdout(), schema.Summary(graphqlIntrospectType))
		return nil
	},
}

func init() {
	graphqlCmd.AddCommand(graphqlIntrospectCmd)
	graphqlIntrospectCmd.Flags().StringArrayVar(&graphqlIntrospectHeaders, "header", nil, "extra request header, repeatable (\"Name: value\")")
	graphqlIntrospectCmd.Flags().StringVar(&graphqlIntrospectType, "type", "", "render only this type")
	graphqlIntrospectCmd.Flags().BoolVar(&graphqlIntrospectJSON, "json", false, "print the raw introspection result as JSON")

	rootCmd.AddCommand(graphqlCmd)
}
