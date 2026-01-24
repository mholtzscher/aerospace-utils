package workspace

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
  aerospace-utils workspace adjust           # +5%
  aerospace-utils workspace adjust -b 10     # +10%
  aerospace-utils workspace adjust -b -5     # -5%
  aerospace-utils workspace adjust --by=-10  # -10%
  aerospace-utils workspace adjust -b -10 --monitor "Dell U2722D"`,
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

	// Create workspace service
	stateSvc := config.NewWorkspaceService(opts.StatePath)

	// Get current percentage for this monitor
	monState, err := stateSvc.GetMonitorState(opts.Monitor)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if monState.Current == nil {
		return errors.New("no current percentage set; use 'workspace use' first")
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
