package workspace

import (
	"context"
	"errors"
	"fmt"

	ufcli "github.com/urfave/cli/v3"

	"github.com/mholtzscher/aerospace-utils/internal/aerospace"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/gaps"
	"github.com/mholtzscher/aerospace-utils/internal/output"
)

const flagShiftBy = "by"

func newShiftCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "shift",
		Usage: "Shift workspace position left or right",
		Description: `Shift the workspace position left or right by adjusting side gap sizes.

With a 50% workspace, the side gaps are 25% each.
Shift left by 5% -> left gap 20%, right gap 30%
Shift right by 5% -> left gap 30%, right gap 20%

Shifts are cumulative - each command adds to the current shift.
Running shift without --by resets shift to 0 (centered).

Examples:
  aerospace-utils workspace shift           # reset to centered
  aerospace-utils workspace shift -b -5     # shift 5% left from current
  aerospace-utils workspace shift -b 5      # shift 5% right from current
  aerospace-utils workspace shift -b 3      # another 3% right (now 8% right total)`,
		Flags: []ufcli.Flag{
			&ufcli.IntFlag{
				Name:    flagShiftBy,
				Aliases: []string{"b"},
				Value:   0,
				Usage:   "Amount to shift workspace (positive = right, negative = left)",
			},
		},
		Action: func(_ context.Context, cmd *ufcli.Command) error {
			return runShift(cmd)
		},
	}
}

//nolint:funlen // Complex command with multiple steps
func runShift(cmd *ufcli.Command) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)

	amount := cmd.Int(flagShiftBy)

	// Create services
	configSvc := config.NewAerospaceService(opts.ConfigPath)
	stateSvc := config.NewWorkspaceService(opts.StatePath)

	// Get current state for this monitor
	monState, err := stateSvc.GetMonitorState(opts.Monitor)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if monState.Current == nil {
		return errors.New("no current percentage set; use 'workspace use' first")
	}

	percentage := *monState.Current
	if validationErr := gaps.ValidatePercentage(percentage); validationErr != nil {
		return validationErr
	}

	// Get current shift (0 if not set)
	currentShift := int64(0)
	if monState.Shift != nil {
		currentShift = *monState.Shift
	}

	// Calculate new shift
	var newShift int64
	if cmd.IsSet(flagShiftBy) {
		// Flag was explicitly set - add to current shift (cumulative)
		newShift = currentShift + int64(amount)
	} else {
		// No flag provided - reset to 0 (centered)
		newShift = 0
	}

	// Get monitor width
	monitorWidth, err := resolveMonitorWidth(opts)
	if err != nil {
		return err
	}

	// Validate shift is within bounds
	if validationErr := gaps.ValidateShift(monitorWidth, percentage, newShift); validationErr != nil {
		return fmt.Errorf("invalid shift: %w", validationErr)
	}

	shift := newShift

	// Calculate shifted gaps
	shiftedGaps := gaps.CalculateShiftedGaps(monitorWidth, percentage, shift)

	if opts.DryRun {
		out.DryRun()
		out.Printf("Would set %s to %d%% (left: %dpx (%d%%), right: %dpx (%d%%))\n",
			opts.Monitor, percentage,
			shiftedGaps.LeftGapPixels, shiftedGaps.LeftGapPercent,
			shiftedGaps.RightGapPixels, shiftedGaps.RightGapPercent)
		return nil
	}

	// Check if config exists
	exists, err := configSvc.Exists()
	if err != nil {
		return fmt.Errorf("check config: %w", err)
	}
	if !exists {
		return fmt.Errorf(
			"config file not found: %s\nCreate it manually or run 'aerospace' to generate a default config",
			configSvc.ConfigPath(),
		)
	}

	// Update config with asymmetric gaps
	if configErr := configSvc.SetMonitorAsymmetricGaps(
		opts.Monitor,
		shiftedGaps.LeftGapPixels,
		shiftedGaps.RightGapPixels,
	); configErr != nil {
		return fmt.Errorf("update config: %w", configErr)
	}

	if writeErr := configSvc.Write(); writeErr != nil {
		return fmt.Errorf("write config: %w", writeErr)
	}

	// Update state with shift
	if shiftErr := stateSvc.SetShift(opts.Monitor, shift); shiftErr != nil {
		return fmt.Errorf("write state: %w", shiftErr)
	}

	// Reload aerospace config
	reloadStatus := ""
	if !opts.NoReload {
		bin, findErr := aerospace.FindBinary()
		if findErr != nil {
			reloadStatus = " (aerospace not found)"
		} else if reloadErr := bin.ReloadConfig(); reloadErr != nil {
			reloadStatus = fmt.Sprintf(" (reload failed: %v)", reloadErr)
		}
	} else {
		reloadStatus = " (reload skipped)"
	}

	// Build success message
	var shiftMsg string
	switch {
	case shift == 0:
		shiftMsg = " (centered)"
	case shift > 0:
		shiftMsg = fmt.Sprintf(" (shifted %d%% right)", shift)
	default:
		shiftMsg = fmt.Sprintf(" (shifted %d%% left)", -shift)
	}

	out.Success("Set %s to %d%% (left: %dpx (%d%%), right: %dpx (%d%%))%s%s\n",
		opts.Monitor, percentage,
		shiftedGaps.LeftGapPixels, shiftedGaps.LeftGapPercent,
		shiftedGaps.RightGapPixels, shiftedGaps.RightGapPercent,
		shiftMsg, reloadStatus)

	return nil
}
