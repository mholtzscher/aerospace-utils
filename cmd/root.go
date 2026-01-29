// Package cmd implements the CLI commands for aerospace-utils.
package cmd

import (
	"context"

	"github.com/mholtzscher/aerospace-utils/cmd/workspace"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	ufcli "github.com/urfave/cli/v3"
)

// Version is set at build time.
var Version = "0.3.3" // x-release-please-version

// Run is the entry point for the CLI.
func Run(ctx context.Context, args []string) error {
	app := &ufcli.Command{
		Name:    "aerospace-utils",
		Usage:   "CLI for managing Aerospace workspace sizing",
		Version: Version,
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:  cli.FlagConfigPath,
				Usage: "Path to aerospace.toml (default: ~/.config/aerospace/aerospace.toml)",
			},
			&ufcli.StringFlag{
				Name:  cli.FlagStatePath,
				Usage: "Path to aerospace-utils-state.toml (default: ~/.config/aerospace/aerospace-utils-state.toml)",
			},
			&ufcli.StringFlag{
				Name:  cli.FlagMonitor,
				Value: "main",
				Usage: "Target monitor name",
			},
			&ufcli.IntFlag{
				Name:   cli.FlagMonitorWidth,
				Value:  0,
				Hidden: true,
				Usage:  "Override detected monitor width in pixels",
			},
			&ufcli.BoolFlag{
				Name:  cli.FlagNoReload,
				Usage: "Skip aerospace reload-config after changes",
			},
			&ufcli.BoolFlag{
				Name:  cli.FlagDryRun,
				Usage: "Print actions without writing changes",
			},
			&ufcli.BoolFlag{
				Name:  cli.FlagVerbose,
				Usage: "Print verbose output",
			},
			&ufcli.BoolFlag{
				Name:  cli.FlagNoColor,
				Usage: "Disable colored output",
			},
		},
		Commands: []*ufcli.Command{
			workspace.NewCommand(),
		},
	}

	return app.Run(ctx, args)
}
