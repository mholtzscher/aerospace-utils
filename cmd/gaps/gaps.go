// Package gaps implements the gaps subcommands.
package gaps

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for gap-related subcommands.
var Cmd = &cobra.Command{
	Use:   "gaps",
	Short: "Manage Aerospace workspace gaps",
	Long:  `Commands for viewing and adjusting workspace gaps based on monitor size.`,
}
