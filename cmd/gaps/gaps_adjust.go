package gaps

import (
	"errors"
	"fmt"

	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/gaps"
	"github.com/spf13/cobra"
)

var adjustAmount int64

var adjustCmd = &cobra.Command{
	Use:   "adjust",
	Short: "Adjust workspace size by amount",
	Long: `Adjust the workspace size percentage by a given amount.

Positive values increase the workspace size (smaller gaps).
Negative values decrease the workspace size (larger gaps).
Default adjustment is +5.

Examples:
  aerospace-utils gaps adjust           # +5%
  aerospace-utils gaps adjust -b 10     # +10%
  aerospace-utils gaps adjust -b -5     # -5%
  aerospace-utils gaps adjust --by=-10  # -10%
  aerospace-utils gaps adjust -b -10 --monitor "Dell U2722D"`,
	Args: cobra.NoArgs,
	RunE: runAdjust,
}

func init() {
	Cmd.AddCommand(adjustCmd)

	adjustCmd.Flags().Int64VarP(&adjustAmount, "by", "b", 5,
		"Amount to adjust workspace size percentage (positive or negative)")
}

func runAdjust(c *cobra.Command, args []string) error {
	opts := cli.GetOptions()

	amount := adjustAmount

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
