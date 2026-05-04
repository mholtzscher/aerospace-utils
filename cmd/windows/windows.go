// Package windows implements window-management subcommands.
package windows

import ufcli "github.com/urfave/cli/v3"

// NewCommand creates the parent windows command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:        "windows",
		Usage:       "Recover and reposition app windows",
		Description: `Commands for recovering app windows when they are off-screen.`,
		Commands: []*ufcli.Command{
			newResetCommand(),
		},
	}
}
