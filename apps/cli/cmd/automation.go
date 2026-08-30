// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/automation"
	"github.com/Its-Satyajit/reqly/internal/workflow"
)

var automationCmd = &cobra.Command{
	Use:   "automation",
	Short: "Self-hosted workflow automation (local scheduler)",
	Long:  `Run a workflow on a local schedule without any cloud.`,
}

var automationRunCmd = &cobra.Command{
	Use:   "run <automation.yaml>",
	Short: "Run an automation file (workflow on interval)",
	Long: `Run an automation defined in YAML.

  reqly automation run my-automation.yaml
  reqly automation run my-automation.yaml --once
  reqly automation run my-automation.yaml --interval 30s --max-runs 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		once, _ := cmd.Flags().GetBool("once")
		intervalFlag, _ := cmd.Flags().GetString("interval")
		maxRuns, _ := cmd.Flags().GetInt("max-runs")

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read automation file: %w", err)
		}
		var auto automation.Automation
		if err := yaml.Unmarshal(data, &auto); err != nil {
			return fmt.Errorf("parse automation file: %w", err)
		}
		if intervalFlag != "" {
			auto.Interval = intervalFlag
		}
		if once {
			auto.Interval = "0"
		}
		if maxRuns > 0 {
			auto.MaxRuns = maxRuns
		}
		if err := auto.Validate(); err != nil {
			return err
		}
		sched := automation.NewScheduler(nil)
		ctx := context.Background()
		out := cmd.OutOrStdout()
		count := 0
		err = sched.Run(ctx, &auto, func(report *workflow.WorkflowReport) {
			count++
			status := "PASSED"
			if !report.Passed {
				status = "FAILED"
			}
			fmt.Fprintf(out, "[%d] Automation %q workflow %q: %s (%d steps, %s)\n", count, auto.Name, report.WorkflowName, status, len(report.Steps), report.Duration.Round(time.Millisecond))
			for _, s := range report.Steps {
				st := "PASSED"
				if !s.Passed {
					st = "FAILED"
				}
				fmt.Fprintf(out, "  [%s] %s: %s\n", st, s.Name, s.RequestPath)
			}
		})
		if err != nil && err != context.Canceled {
			return err
		}
		return nil
	},
}

func init() {
	automationRunCmd.Flags().Bool("once", false, "run once regardless of interval")
	automationRunCmd.Flags().String("interval", "", "override interval (e.g. 10s, 5m)")
	automationRunCmd.Flags().Int("max-runs", 0, "override maxRuns (0 = infinite)")
	automationCmd.AddCommand(automationRunCmd)
	rootCmd.AddCommand(automationCmd)
}
