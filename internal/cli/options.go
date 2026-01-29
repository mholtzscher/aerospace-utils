// Package cli provides shared CLI types and utilities.
package cli

import ufcli "github.com/urfave/cli/v3"

// Flag names for global options.
const (
	FlagConfigPath   = "config-path"
	FlagStatePath    = "state-path"
	FlagMonitor      = "monitor"
	FlagMonitorWidth = "monitor-width"
	FlagNoReload     = "no-reload"
	FlagDryRun       = "dry-run"
	FlagVerbose      = "verbose"
	FlagNoColor      = "no-color"
)

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
		ConfigPath:   root.String(FlagConfigPath),
		StatePath:    root.String(FlagStatePath),
		Monitor:      root.String(FlagMonitor),
		MonitorWidth: int64(root.Int(FlagMonitorWidth)),
		NoReload:     root.Bool(FlagNoReload),
		DryRun:       root.Bool(FlagDryRun),
		Verbose:      root.Bool(FlagVerbose),
		NoColor:      root.Bool(FlagNoColor),
	}
}
