package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/mholtzscher/aerospace-utils/internal/aerospace"
	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/gaps"
	"github.com/mholtzscher/aerospace-utils/internal/output"
	ufcli "github.com/urfave/cli/v3"
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

Running shift without --by resets shift to 0 (centered).

Examples:
  aerospace-utils workspace shift           # reset to centered
  aerospace-utils workspace shift -b -5     # shift 5% left
  aerospace-utils workspace shift -b 5      # shift 5% right`,
		Flags: []ufcli.Flag{
			&ufcli.IntFlag{
				Name:    flagShiftBy,
				Aliases: []string{"b"},
				Value:   0,
				Usage:   "Amount to shift workspace (positive = right, negative = left)",
			},
		},
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			return runShift(cmd)
		},
	}
}

func runShift(cmd *ufcli.Command) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)

	shift := cmd.Int(flagShiftBy)

	// Create services
	configSvc := config.NewAerospaceService(opts.ConfigPath)
	stateSvc := config.NewWorkspaceService(opts.StatePath)

	// Get current percentage for this monitor
	monState, err := stateSvc.GetMonitorState(opts.Monitor)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if monState.Current == nil {
		return errors.New("no current percentage set; use 'workspace use' first")
	}

	percentage := *monState.Current
	if err := gaps.ValidatePercentage(percentage); err != nil {
		return err
	}

	// Get monitor width
	monitorWidth, err := resolveMonitorWidth(opts)
	if err != nil {
		return err
	}

	// Validate shift is within bounds
	if err := gaps.ValidateShift(monitorWidth, percentage, int64(shift)); err != nil {
		return fmt.Errorf("invalid shift: %w", err)
	}

	// Calculate shifted gaps
	shiftedGaps := gaps.CalculateShiftedGaps(monitorWidth, percentage, int64(shift))

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
		return fmt.Errorf("config file not found: %s\nCreate it manually or run 'aerospace' to generate a default config", configSvc.ConfigPath())
	}

	// Update config with asymmetric gaps
	if err := configSvc.SetMonitorAsymmetricGaps(opts.Monitor, shiftedGaps.LeftGapPixels, shiftedGaps.RightGapPixels); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	if err := configSvc.Write(); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Update state with shift
	if err := stateSvc.SetShift(opts.Monitor, shift); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	// Reload aerospace config
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

	// Build success message
	shiftMsg := ""
	if shift == 0 {
		shiftMsg = " (centered)"
	} else if shift > 0 {
		shiftMsg = fmt.Sprintf(" (shifted %d%% right)", shift)
	} else {
		shiftMsg = fmt.Sprintf(" (shifted %d%% left)", -shift)
	}

	out.Success("Set %s to %d%% (left: %dpx (%d%%), right: %dpx (%d%%))%s%s\n",
		opts.Monitor, percentage,
		shiftedGaps.LeftGapPixels, shiftedGaps.LeftGapPercent,
		shiftedGaps.RightGapPixels, shiftedGaps.RightGapPercent,
		shiftMsg, reloadStatus)

	return nil
}
