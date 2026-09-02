package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/monitor"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

var (
	monitorInterval time.Duration
	monitorJSON     bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Scheduled health checks",
}

var monitorRunCmd = &cobra.Command{
	Use:   "run <request-file> [--interval 5m] [--json]",
	Short: "Run scheduled health checks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		f, err := requestfile.LoadFile(path)
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(path)
		svc := core.NewRunService(findWorkspaceRoot(baseDir))
		defer svc.Close()

		fileVars := f.VariablesSet()
		noRecord := false
		send := func(ctx context.Context) (time.Duration, int, error) {
			res, err := svc.Run(ctx, f.Request, core.RunRequestOptions{
				EnvFlag:       envFlag,
				FileEnv:       f.Environment,
				FileVars:      fileVars,
				RecordHistory: &noRecord,
			})
			if err != nil {
				return 0, 0, err
			}
			return res.Response.Duration, res.Response.StatusCode, nil
		}
		threshold := monitor.Threshold{Status: 200}
		onResult := func(r monitor.Result) {
			if monitorJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				_ = enc.Encode(r)
			} else {
				status := "ok"
				if !r.OK {
					status = "fail"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] status=%d latency=%dms\n", status, r.Status, r.LatencyMs)
			}
		}
		interval := monitorInterval
		return monitor.Run(cmd.Context(), interval, threshold, send, onResult)
	},
}

func init() {
	monitorRunCmd.Flags().DurationVar(&monitorInterval, "interval", 0, "check interval (default 5m)")
	monitorRunCmd.Flags().BoolVar(&monitorJSON, "json", false, "JSON per-tick output")
	monitorCmd.AddCommand(monitorRunCmd)
}
