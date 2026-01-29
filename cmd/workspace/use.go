package workspace

import (
	"context"
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
	ufcli "github.com/urfave/cli/v3"
)

const flagSetDefault = "set-default"

func newUseCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "use",
		Usage: "Set workspace size percentage",
		Description: `Set the workspace size as a percentage of the monitor width.

The gap size is calculated to achieve the desired percentage.
If no percentage is given, uses the current or default percentage. If the
state file is missing or empty, defaults to 60%.

Examples:
  aerospace-utils workspace use 40
  aerospace-utils workspace use 80 --monitor "Dell U2722D"
  aerospace-utils workspace use --set-default 50`,
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:  flagSetDefault,
				Usage: "Also set as the default percentage for this monitor",
			},
		},
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			return runUse(cmd)
		},
	}
}

func runUse(cmd *ufcli.Command) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)

	// Parse optional percentage argument
	var explicitPercent *int64
	if cmd.Args().Len() > 0 {
		p, err := strconv.ParseInt(cmd.Args().Get(0), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid percentage %q: %w", cmd.Args().Get(0), err)
		}
		explicitPercent = &p
	}

	return applyPercentage(cmd, opts, out, explicitPercent)
}

func applyPercentage(cmd *ufcli.Command, opts *cli.GlobalOptions, out *output.Printer, explicitPercent *int64) error {
	configSvc := config.NewAerospaceService(opts.ConfigPath)
	stateSvc := config.NewWorkspaceService(opts.StatePath)

	// Resolve percentage
	percentage, err := stateSvc.ResolvePercentage(opts.Monitor, explicitPercent)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
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

	// Check for existing shift and apply asymmetric gaps if set
	originalShift, err := stateSvc.GetShift(opts.Monitor)
	if err != nil {
		return fmt.Errorf("load shift: %w", err)
	}
	shift := originalShift

	useAsymmetric := false
	var shiftedGaps gaps.ShiftedGaps
	var symmetricGapSize int64
	var gapMsg string

	if shift != 0 {
		if err := gaps.ValidateShift(monitorWidth, *percentage, shift); err != nil {
			// Shift is no longer valid for the current percentage.
			shift = 0
		} else {
			useAsymmetric = true
			shiftedGaps = gaps.CalculateShiftedGaps(monitorWidth, *percentage, shift)
			gapMsg = fmt.Sprintf("(left: %dpx (%d%%), right: %dpx (%d%%))",
				shiftedGaps.LeftGapPixels, shiftedGaps.LeftGapPercent,
				shiftedGaps.RightGapPixels, shiftedGaps.RightGapPercent)
		}
	}

	if !useAsymmetric {
		symmetricGapSize = gaps.CalculateGapSize(monitorWidth, *percentage)
		gapMsg = fmt.Sprintf("(%dpx gaps)", symmetricGapSize)
	}

	if opts.DryRun {
		out.DryRun()
		out.Printf("Would set %s to %d%% %s\n",
			opts.Monitor, *percentage, gapMsg)
		return nil
	}

	// Check if config exists
	exists, err := configSvc.Exists()
	if err != nil {
		return fmt.Errorf("check config: %w", err)
	}
	if !exists {
		return fmt.Errorf("config file not found: %s\nCreate it manually or run 'aerospace' to generate a default config", configSvc.ConfigPath())
	}

	if useAsymmetric {
		if err := configSvc.SetMonitorAsymmetricGaps(opts.Monitor, shiftedGaps.LeftGapPixels, shiftedGaps.RightGapPixels); err != nil {
			return fmt.Errorf("update config: %w", err)
		}
	} else {
		if err := configSvc.SetMonitorGaps(opts.Monitor, symmetricGapSize); err != nil {
			return fmt.Errorf("update config: %w", err)
		}
	}

	if err := configSvc.Write(); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Update state - preserves shift by calling Update then SetShift
	setDefaultFlag := cmd.Bool(flagSetDefault)
	if err := stateSvc.Update(opts.Monitor, *percentage, setDefaultFlag); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	// If shift was reset, write 0 to clear it
	if originalShift != shift {
		if err := stateSvc.SetShift(opts.Monitor, 0); err != nil {
			return fmt.Errorf("write state: %w", err)
		}
	}

	// Reload aerospace config and build single-line output
	reloadStatus := ""
	if !opts.NoReload {
		bin, err := aerospace.FindBinary()
		if err != nil {
			reloadStatus = " (aerospace not found)"
		} else if err := bin.ReloadConfig(); err != nil {
			reloadStatus = fmt.Sprintf(" (reload failed: %v)", err)
		}
	} else {
		reloadStatus = " (reload skipped)"
	}

	defaultSuffix := ""
	if setDefaultFlag {
		defaultSuffix = ", set as default"
	}
	out.Success("Set %s to %d%% %s%s%s\n",
		opts.Monitor, *percentage, gapMsg, defaultSuffix, reloadStatus)

	return nil
}

// resolveMonitorWidth determines the monitor width to use for gap calculation.
func resolveMonitorWidth(opts *cli.GlobalOptions) (int64, error) {
	// Use explicit override if provided
	if opts.MonitorWidth > 0 {
		return int64(opts.MonitorWidth), nil
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
func RunWithPercent(cmd *ufcli.Command, percentage int64) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)
	return applyPercentage(cmd, opts, out, &percentage)
}
