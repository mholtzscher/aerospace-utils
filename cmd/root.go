// Package cmd implements the CLI commands for aerospace-utils.
package cmd

import (
	"context"

	"github.com/mholtzscher/aerospace-utils/cmd/workspace"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	ufcli "github.com/urfave/cli/v3"
)

// Version is set at build time.
var Version = "0.2.0"

// Run is the entry point for the CLI.
func Run(ctx context.Context, args []string) error {
	opts := cli.GetOptions()

	app := &ufcli.Command{
		Name:    "aerospace-utils",
		Usage:   "CLI for managing Aerospace workspace sizing",
		Version: Version,
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:        "config-path",
				Destination: &opts.ConfigPath,
				Usage:       "Path to aerospace.toml (default: ~/.config/aerospace/aerospace.toml)",
			},
			&ufcli.StringFlag{
				Name:        "state-path",
				Destination: &opts.StatePath,
				Usage:       "Path to aerospace-utils-state.toml (default: ~/.config/aerospace/aerospace-utils-state.toml)",
			},
			&ufcli.StringFlag{
				Name:        "monitor",
				Destination: &opts.Monitor,
				Value:       "main",
				Usage:       "Target monitor name",
			},
			&ufcli.IntFlag{
				Name:  "monitor-width",
				Value: 0,
				Action: func(_ context.Context, cmd *ufcli.Command, val int64) error {
					opts.MonitorWidth = val
					return nil
				},
				Hidden: true,
				Usage:  "Override detected monitor width in pixels",
			},
			&ufcli.BoolFlag{
				Name:        "no-reload",
				Destination: &opts.NoReload,
				Usage:       "Skip aerospace reload-config after changes",
			},
			&ufcli.BoolFlag{
				Name:        "dry-run",
				Destination: &opts.DryRun,
				Usage:       "Print actions without writing changes",
			},
			&ufcli.BoolFlag{
				Name:        "verbose",
				Destination: &opts.Verbose,
				Usage:       "Print verbose output",
			},
			&ufcli.BoolFlag{
				Name:        "no-color",
				Destination: &opts.NoColor,
				Usage:       "Disable colored output",
			},
		},
		Commands: []*ufcli.Command{
			workspace.NewCommand(),
		},
	}

	return app.Run(ctx, args)
}
