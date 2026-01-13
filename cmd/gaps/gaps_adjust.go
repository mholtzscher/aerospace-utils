package gaps

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/gaps"
	"github.com/spf13/cobra"
)

var adjustCmd = &cobra.Command{
	Use:   "adjust [amount]",
	Short: "Adjust workspace size by amount",
	Long: `Adjust the workspace size percentage by a given amount.

Positive values increase the workspace size (smaller gaps).
Negative values decrease the workspace size (larger gaps).
Default adjustment is +5.

Examples:
  aerospace-utils gaps adjust      # +5%
  aerospace-utils gaps adjust 10   # +10%
  aerospace-utils gaps adjust -5   # -5%
  aerospace-utils gaps adjust -10 --monitor "Dell U2722D"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdjust,
}

func init() {
	Cmd.AddCommand(adjustCmd)

	// Allow negative numbers as arguments (e.g., -5)
	adjustCmd.Flags().SetInterspersed(false)
}

func runAdjust(c *cobra.Command, args []string) error {
	opts := cli.GetOptions()

	// Parse adjustment amount (default: 5)
	amount := int64(5)
	if len(args) > 0 {
		a, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", args[0], err)
		}
		amount = a
	}

	// Resolve state path
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

	// Get current percentage for this monitor
	monState := state.Monitors[opts.Monitor]
	if monState == nil || monState.Current == nil {
		return errors.New("no current percentage set; use 'gaps use' first")
	}

	// Calculate new percentage
	newPercent := *monState.Current + amount

	// Validate new percentage
	if err := gaps.ValidatePercentage(newPercent); err != nil {
		return fmt.Errorf("adjusted percentage %d is invalid: %w", newPercent, err)
	}

	// Delegate to gaps use
	return RunWithPercent(newPercent)
}
