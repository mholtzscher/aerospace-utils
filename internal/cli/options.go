// Package cli provides shared CLI types and utilities.
package cli

import ufcli "github.com/urfave/cli/v3"

// GlobalOptions holds flags available to all subcommands.
type GlobalOptions struct {
	ConfigPath   string
	StatePath    string
	Monitor      string
	MonitorWidth int64
	NoReload     bool
	DryRun       bool
	Verbose      bool
	NoColor      bool
}

// GetOptions reads GlobalOptions from the root command's flags.
// Call this in your command's Action to get the current values.
func GetOptions(cmd *ufcli.Command) *GlobalOptions {
	if cmd == nil {
		return &GlobalOptions{}
	}

	root := cmd.Root()
	if root == nil {
		root = cmd
	}

	return &GlobalOptions{
		ConfigPath:   root.String("config-path"),
		StatePath:    root.String("state-path"),
		Monitor:      root.String("monitor"),
		MonitorWidth: root.Int("monitor-width"),
		NoReload:     root.Bool("no-reload"),
		DryRun:       root.Bool("dry-run"),
		Verbose:      root.Bool("verbose"),
		NoColor:      root.Bool("no-color"),
	}
}
