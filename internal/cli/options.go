// Package cli provides shared CLI types and utilities.
package cli

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

var globalOpts GlobalOptions

// GetOptions returns the current global options.
func GetOptions() *GlobalOptions {
	return &globalOpts
}

// Options returns a pointer to the options struct for flag binding.
func Options() *GlobalOptions {
	return &globalOpts
}
