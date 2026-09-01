// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Its-Satyajit/reqly/internal/workflow"
)

var workflowVerbose bool

var workflowCmd = &cobra.Command{
	Use:   "workflow <workflow.yaml>",
	Short: "Execute a visual/programmatic multi-step API workflow",
	Long: `Execute a multi-step API workflow with variable extraction, condition evaluation,
and step reporting.

  reqly workflow auth-flow.yaml
  reqly workflow auth-flow.yaml --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read workflow file: %w", err)
		}

		var wf workflow.Workflow
		if err := yaml.Unmarshal(data, &wf); err != nil {
			return fmt.Errorf("parse workflow file: %w", err)
		}

		exec := workflow.NewWorkflowExecutor(nil)
		report, err := exec.Execute(context.Background(), &wf, workflow.WorkflowOptions{})
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Workflow: %s (%s, %d steps)\n", report.WorkflowName, report.Duration.Round(100), len(report.Steps))
		for _, s := range report.Steps {
			status := "PASSED"
			if !s.Passed {
				status = "FAILED"
			}
			fmt.Fprintf(out, "  [%s] %s: %s\n", status, s.Name, s.RequestPath)
			if !s.Passed {
				if s.RequestError != "" {
					fmt.Fprintf(out, "    Error: %s\n", s.RequestError)
				} else if s.Response != nil {
					statusInfo := fmt.Sprintf("%d", s.Response.StatusCode)
					if s.Response.StatusText != "" {
						statusInfo += " " + s.Response.StatusText
					}
					bodySnippet := strings.TrimSpace(string(s.Response.Body))
					if len(bodySnippet) > 200 {
						bodySnippet = bodySnippet[:200] + "..."
					}
					if bodySnippet != "" {
						fmt.Fprintf(out, "    Status: %s (Body: %s)\n", statusInfo, bodySnippet)
					} else {
						fmt.Fprintf(out, "    Status: %s\n", statusInfo)
					}
				}
			}
			if workflowVerbose && len(s.Logs) > 0 {
				for _, l := range s.Logs {
					fmt.Fprintf(out, "    Log: %s\n", l)
				}
			}
		}
		if len(report.ExtractedVars) > 0 {
			fmt.Fprintln(out, "Extracted Variables:")
			for k, v := range report.ExtractedVars {
				fmt.Fprintf(out, "  - %s = %s\n", k, v)
			}
		}

		if !report.Passed {
			return fmt.Errorf("workflow execution failed")
		}
		return nil
	},
}

func init() {
	workflowCmd.Flags().BoolVarP(&workflowVerbose, "verbose", "v", false, "print verbose execution logs for workflow steps")
	rootCmd.AddCommand(workflowCmd)
}
