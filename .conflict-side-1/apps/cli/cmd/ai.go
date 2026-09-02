// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/ai"
	"github.com/Its-Satyajit/reqly/internal/jsonschema"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
	"github.com/Its-Satyajit/reqly/internal/response"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI assistant (local heuristics & generators)",
}

func loadResponseFile(path string) (*response.Response, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var resp response.Response
	if err := json.Unmarshal(data, &resp); err == nil && resp.StatusCode > 0 {
		return &resp, nil
	}
	return &response.Response{StatusCode: 200, StatusText: "OK", Body: data}, nil
}

var aiExplainCmd = &cobra.Command{
	Use:   "explain <response.json>",
	Short: "Explain a response summary and latency breakdown",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := loadResponseFile(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), ai.ExplainResponse(resp))
		return nil
	},
}

var aiTestCmd = &cobra.Command{
	Use:   "test <response.json>",
	Short: "Synthesize Goja/JavaScript test assertions from a response",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := loadResponseFile(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), ai.GenerateTests(resp))
		return nil
	},
}

var aiDocsCmd = &cobra.Command{
	Use:   "docs <request.yaml|json> [response.json]",
	Short: "Generate Markdown API documentation from request & response",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		rf, err := requestfile.LoadFile(args[0])
		var req *request.Request
		if err == nil && rf != nil {
			req = &rf.Request
		} else {
			// Try raw request JSON
			data, readErr := os.ReadFile(args[0])
			if readErr != nil {
				return fmt.Errorf("read %s: %w", args[0], readErr)
			}
			var r request.Request
			if unmarshalErr := json.Unmarshal(data, &r); unmarshalErr == nil {
				req = &r
			}
		}

		var resp *response.Response
		if len(args) > 1 {
			resp, _ = loadResponseFile(args[1])
		}

		fmt.Fprintln(cmd.OutOrStdout(), ai.GenerateDocs(req, resp))
		return nil
	},
}

var aiDiagnoseCmd = &cobra.Command{
	Use:   "diagnose <response.json>",
	Short: "Diagnose response failure codes or network errors with actionable tips",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := loadResponseFile(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), ai.Diagnose(resp, nil))
		return nil
	},
}

var aiSchemaCmd = &cobra.Command{
	Use:   "schema <schema> <instance>",
	Short: "Explain schema validation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaData, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		instanceData, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		sch, err := jsonschema.Compile(schemaData, "")
		if err != nil {
			return err
		}
		violations, err := jsonschema.Validate(sch, instanceData)
		if err != nil {
			return err
		}
		if len(violations) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "valid")
			return nil
		}
		for _, v := range violations {
			fmt.Fprintln(cmd.OutOrStdout(), v.String())
		}
		return nil
	},
}

func init() {
	aiCmd.AddCommand(
		aiExplainCmd,
		aiTestCmd,
		aiDocsCmd,
		aiDiagnoseCmd,
		aiSchemaCmd,
	)
}
