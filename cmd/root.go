// Package cmd implements the CLI commands for aerospace-utils.
package cmd

import (
	"fmt"
	"os"

	"github.com/mholtzscher/aerospace-utils/cmd/gaps"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/spf13/cobra"
)

// Version is set at build time.
var Version = "0.2.0"

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:     "aerospace-utils",
	Short:   "CLI for managing Aerospace workspace gaps",
	Version: Version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	opts := cli.Options()

	// Global persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&opts.ConfigPath, "config-path", "",
		"Path to aerospace.toml (default: ~/.config/aerospace/aerospace.toml)")
	rootCmd.PersistentFlags().StringVar(&opts.StatePath, "state-path", "",
		"Path to workspace-size.toml (default: ~/.config/aerospace/workspace-size.toml)")
	rootCmd.PersistentFlags().StringVar(&opts.Monitor, "monitor", "main",
		"Target monitor name")
	rootCmd.PersistentFlags().Int64Var(&opts.MonitorWidth, "monitor-width", 0,
		"Override detected monitor width in pixels")
	rootCmd.PersistentFlags().BoolVar(&opts.NoReload, "no-reload", false,
		"Skip aerospace reload-config after changes")
	rootCmd.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false,
		"Print actions without writing changes")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false,
		"Print verbose output")
	rootCmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false,
		"Disable colored output")

	// Hide the monitor-width flag (for testing/advanced use)
	if err := rootCmd.PersistentFlags().MarkHidden("monitor-width"); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to hide monitor-width flag:", err)
	}

	// Add subcommands
	rootCmd.AddCommand(gaps.Cmd)
}
