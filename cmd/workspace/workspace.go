// Package workspace implements the workspace subcommands.
package workspace

import (
	ufcli "github.com/urfave/cli/v3"
)

// NewCommand creates the parent workspace command with all subcommands.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:        "workspace",
		Usage:       "Manage Aerospace workspace sizing",
		Description: `Commands for viewing and adjusting workspace sizing based on monitor gaps.`,
		Commands: []*ufcli.Command{
			newUseCommand(),
			newAdjustCommand(),
			newCurrentCommand(),
		},
	}
}
