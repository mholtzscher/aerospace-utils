package gaps

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mholtzscher/aerospace-utils/internal/aerospace"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/display"
	"github.com/mholtzscher/aerospace-utils/internal/gaps"
	"github.com/mholtzscher/aerospace-utils/internal/output"
	"github.com/spf13/cobra"
)

var setDefault bool

var useCmd = &cobra.Command{
	Use:   "use [percent]",
	Short: "Set workspace size percentage",
	Long: `Set the workspace size as a percentage of the monitor width.

The gap size is calculated to achieve the desired percentage.
If no percentage is given, uses the current or default percentage.

Examples:
  aerospace-utils gaps use 40
  aerospace-utils gaps use 80 --monitor "Dell U2722D"
  aerospace-utils gaps use --set-default 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUse,
}

func init() {
	Cmd.AddCommand(useCmd)
	useCmd.Flags().BoolVar(&setDefault, "set-default", false,
		"Also set as the default percentage for this monitor")
}

func runUse(c *cobra.Command, args []string) error {
	opts := cli.GetOptions()
	out := output.New(opts.NoColor)

	// Parse optional percentage argument
	var explicitPercent *int64
	if len(args) > 0 {
		p, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid percentage %q: %w", args[0], err)
		}
		explicitPercent = &p
	}

	// Resolve paths
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}
	configPath = config.ExpandPath(configPath)

	statePath := opts.StatePath
	if statePath == "" {
		statePath = config.DefaultStatePath()
	}
	statePath = config.ExpandPath(statePath)

	// Load state
	state, err := config.LoadState(statePath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// Resolve percentage
	percentage := state.ResolvePercentage(opts.Monitor, explicitPercent)
	if percentage == nil {
		return errors.New("no percentage specified and no current/default set for this monitor")
	}

	// Validate percentage
	if err := gaps.ValidatePercentage(*percentage); err != nil {
		return err
	}

	// Get monitor width
	monitorWidth, err := resolveMonitorWidth(opts)
	if err != nil {
		return err
	}

	// Calculate gap size
	gapSize := gaps.CalculateGapSize(monitorWidth, *percentage)

	if opts.Verbose || opts.DryRun {
		if opts.DryRun {
			out.DryRun()
		}
		out.Printf("Setting %s to %d%% (%dpx gaps on %dpx monitor)\n",
			opts.Monitor, *percentage, gapSize, monitorWidth)
	}

	// Update config and state
	if !opts.DryRun {
		// Load and update config
		cfg, err := config.LoadAerospaceConfig(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if err := cfg.SetMonitorGaps(opts.Monitor, gapSize); err != nil {
			return fmt.Errorf("update config: %w", err)
		}

		if err := cfg.Write(); err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		// Update state
		state.Update(opts.Monitor, *percentage, setDefault)
		if err := state.Write(); err != nil {
			return fmt.Errorf("write state: %w", err)
		}

		// Reload aerospace config
		if !opts.NoReload {
			bin, err := aerospace.FindBinary()
			if err != nil {
				out.Warning("aerospace not found, skipping reload\n")
			} else if err := bin.ReloadConfig(); err != nil {
				out.ReloadFailed(err)
			} else {
				out.ReloadOK()
			}
		} else if opts.Verbose {
			out.ReloadSkipped()
		}
	}

	return nil
}

// resolveMonitorWidth determines the monitor width to use for gap calculation.
func resolveMonitorWidth(opts *cli.GlobalOptions) (int64, error) {
	// Use explicit override if provided
	if opts.MonitorWidth > 0 {
		return opts.MonitorWidth, nil
	}

	// Check if display detection is available
	if !display.Available() {
		return 0, errors.New("display detection not available; use --monitor-width")
	}

	// For "main" monitor, use the simple main display width
	if opts.Monitor == "main" {
		return display.MainWidth()
	}

	// Enumerate displays and match by name
	displays, err := display.Enumerate()
	if err != nil {
		return 0, fmt.Errorf("enumerate displays: %w", err)
	}

	// Try to match by name (case-insensitive)
	for _, d := range displays {
		if strings.EqualFold(d.Name, opts.Monitor) {
			return d.Width, nil
		}
	}

	// Build helpful error message
	var names []string
	for _, d := range displays {
		names = append(names, d.Name)
	}

	return 0, fmt.Errorf("monitor %q not found; available: %s (use --monitor-width to specify)",
		opts.Monitor, strings.Join(names, ", "))
}

// RunWithPercent is called by adjust to apply a calculated percentage.
func RunWithPercent(percentage int64) error {
	return runUse(useCmd, []string{strconv.FormatInt(percentage, 10)})
}
