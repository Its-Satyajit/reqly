// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage local request history (SQLite per-workspace)",
	Long: `Inspect and replay local request history stored in <workspace>/.reqly/history.db.

History is local metadata (0600, .gitignore'd), not Git-native. Each send records the fully-resolved request + response; display masks Authorization/Cookie secrets while the DB stores exact bytes for faithful replay.`,
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent history entries",
	RunE: func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		envFilter, _ := cmd.Flags().GetString("env")
		jsonOut, _ := cmd.Flags().GetBool("json")
		status, _ := cmd.Flags().GetString("status")
		svc, err := historyService()
		if err != nil {
			return err
		}
		defer closeHistoryStore(svc)
		var statusFilter *int
		if status != "" {
			v := statusToFilter(status)
			if v == nil {
				return fmt.Errorf("unknown --status %q (use 2xx, 4xx, 5xx, or 200)", status)
			}
			statusFilter = v
		}
		entries, err := svc.List(context.Background(), limit, 0, statusFilter)
		if err != nil {
			return err
		}
		// env filter post (service List doesn't yet filter by env range; simple post-filter)
		if envFilter != "" {
			var filtered []history.Entry
			for _, e := range entries {
				if e.Env == envFilter {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(entries)
		}
		for _, e := range entries {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s %d %dms %s\n", e.ID, e.Method, e.URL, e.Status, e.DurationMS, e.Env)
		}
		return nil
	},
}

var historyShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show one history entry (masked)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := historyService()
		if err != nil {
			return err
		}
		defer closeHistoryStore(svc)
		jsonOut, _ := cmd.Flags().GetBool("json")
		e, err := svc.Show(context.Background(), args[0])
		if err != nil {
			return err
		}
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(e)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\nURL: %s\nMethod: %s\nStatus: %d\nEnv: %s\nReqBody: %s\nRespBody: %s\n", e.ID, e.URL, e.Method, e.Status, e.Env, string(e.ReqBody), string(e.RespBody))
		return nil
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search history via FTS (url, request_path)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOut, _ := cmd.Flags().GetBool("json")
		svc, err := historyService()
		if err != nil {
			return err
		}
		defer closeHistoryStore(svc)
		entries, err := svc.Search(context.Background(), args[0], limit)
		if err != nil {
			return err
		}
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(entries)
		}
		for _, e := range entries {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", e.ID, e.RequestPath, e.URL)
		}
		return nil
	},
}

var historyReplayCmd = &cobra.Command{
	Use:   "replay <id>",
	Short: "Replay a history entry's stored request verbatim",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := historyService()
		if err != nil {
			return err
		}
		defer closeHistoryStore(svc)
		jsonOut, _ := cmd.Flags().GetBool("json")
		// Replay loads stored request and re-sends via request.Client
		// For now, history.Entry stores URL/method; replay rebuilds a minimal Request.
		e, err := svc.Show(context.Background(), args[0])
		if err != nil {
			return err
		}
		client := request.NewClient()
		// Load raw entry to get exact bodies (Show returns masked headers but bodies exact)
		storeEntry, err := rawHistoryEntry(args[0])
		if err != nil {
			return err
		}
		req := request.Request{Method: request.Method(storeEntry.Method), URL: storeEntry.URL, Headers: headersFromMap(storeEntry.ReqHeaders), Body: string(storeEntry.ReqBody)}
		resp, err := client.Execute(context.Background(), &req, nil)
		if err != nil {
			return err
		}
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(resp)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d %s %s\n", resp.StatusCode, resp.Proto, string(resp.Body))
		_ = e
		return nil
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear history (per env or workspace)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		envFlag, _ := cmd.Flags().GetString("env")
		all, _ := cmd.Flags().GetBool("all")
		force, _ := cmd.Flags().GetBool("force")
		var env *string
		if envFlag != "" {
			env = &envFlag
		} else if !all {
			return fmt.Errorf("specify --env <name> or --all")
		}
		if !force {
			fmt.Fprintf(cmd.OutOrStdout(), "clear history %s? use --force to confirm\n", describeClear(env))
			return nil
		}
		svc, err := historyService()
		if err != nil {
			return err
		}
		defer closeHistoryStore(svc)
		if err := svc.Clear(context.Background(), env); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "cleared")
		return nil
	},
}

func historyService() (*core.HistoryService, error) {
	root := workspaceRoot()
	dbPath := filepath.Join(root, ".reqly", "history.db")
	store, err := history.NewStore(dbPath)
	if err != nil {
		return nil, err
	}
	return core.NewHistoryService(store, request.NewClient()), nil
}

func rawHistoryEntry(id string) (history.Entry, error) {
	root := workspaceRoot()
	dbPath := filepath.Join(root, ".reqly", "history.db")
	store, err := history.NewStore(dbPath)
	if err != nil {
		return history.Entry{}, err
	}
	defer store.Close()
	return store.Show(context.Background(), id)
}

func closeHistoryStore(svc *core.HistoryService) {
	_ = svc.Close()
}

func workspaceRoot() string {
	if ws, err := collections.LoadWorkspace("."); err == nil && ws.Root != "" {
		return ws.Root
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func headersFromMap(m map[string][]string) []request.Header {
	var out []request.Header
	for k, vals := range m {
		for _, v := range vals {
			out = append(out, request.Header{Key: k, Value: v})
		}
	}
	return out
}

func statusToFilter(s string) *int {
	var v int
	switch s {
	case "2xx":
		v = 200
		return &v
	case "4xx":
		v = 400
		return &v
	case "5xx":
		v = 500
		return &v
	default:
		var code int
		if _, err := fmt.Sscan(s, &code); err == nil {
			return &code
		}
		return nil
	}
}

func describeClear(env *string) string {
	if env == nil {
		return "workspace"
	}
	return fmt.Sprintf("env %q", *env)
}

func init() {
	historyListCmd.Flags().Int("limit", 50, "max entries")
	historyListCmd.Flags().String("env", "", "filter by env")
	historyListCmd.Flags().Bool("json", false, "JSON output")
	historyListCmd.Flags().String("status", "", "filter by status (2xx,4xx,5xx or code)")
	historyShowCmd.Flags().Bool("json", false, "JSON output")
	historySearchCmd.Flags().Int("limit", 50, "max entries")
	historySearchCmd.Flags().Bool("json", false, "JSON output")
	historyReplayCmd.Flags().Bool("json", false, "JSON output")
	historyClearCmd.Flags().String("env", "", "clear only this env")
	historyClearCmd.Flags().Bool("all", false, "clear entire workspace history")
	historyClearCmd.Flags().Bool("force", false, "confirm without prompt")
	historyCmd.AddCommand(historyListCmd, historyShowCmd, historySearchCmd, historyReplayCmd, historyClearCmd)
	rootCmd.AddCommand(historyCmd)
}
