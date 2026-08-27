package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/ai"
	"github.com/Its-Satyajit/reqly/internal/jsonschema"
	"github.com/Its-Satyajit/reqly/internal/response"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI assistant (local heuristic)",
}

var aiExplainCmd = &cobra.Command{
	Use:   "explain <response.json>",
	Short: "Explain a response",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var resp response.Response
		if err := json.Unmarshal(data, &resp); err != nil {
			// Fallback: treat body as text response
			resp = response.Response{StatusCode: 200, StatusText: "OK", Body: data}
		}
		fmt.Fprintln(cmd.OutOrStdout(), ai.ExplainResponse(&resp))
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
	aiCmd.AddCommand(aiExplainCmd, aiSchemaCmd)
}
