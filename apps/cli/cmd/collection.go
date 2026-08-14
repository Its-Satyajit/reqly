package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Work with collections",
}

var collectionRunCmd = &cobra.Command{
	Use:   "run <collection>",
	Short: "Run a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "collection run %s: not implemented yet\n", args[0])
		return nil
	},
}

func init() {
	collectionCmd.AddCommand(collectionRunCmd)
}
