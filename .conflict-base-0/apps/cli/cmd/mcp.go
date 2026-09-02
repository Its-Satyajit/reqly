package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol server",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve MCP over stdio",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcp.Serve(os.Stdin, os.Stdout)
	},
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd)
}
