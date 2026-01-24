// Package workspace implements the workspace subcommands.
package workspace

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for workspace-related subcommands.
var Cmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage Aerospace workspace sizing",
	Long:  `Commands for viewing and adjusting workspace sizing based on monitor gaps.`,
}
