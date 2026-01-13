package gaps

import (
	"fmt"
	"os"

	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/output"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current gap configuration and state",
	Long: `Display the current aerospace gap configuration and workspace state.

Shows:
- Config file path and gap values
- State file path and per-monitor percentages`,
	RunE: runCurrent,
}

func init() {
	Cmd.AddCommand(currentCmd)
}

func runCurrent(c *cobra.Command, args []string) error {
	opts := cli.GetOptions()
	out := output.New(opts.NoColor)

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

	// Print config info
	out.PrintHeader("Config")
	out.PrintPath("path", configPath)

	if _, err := os.Stat(configPath); err == nil {
		cfg, err := config.LoadAerospaceConfig(configPath)
		if err != nil {
			out.Error("  Error loading config: %v\n", err)
		} else {
			summary := cfg.Summary()
			printConfigSummary(out, summary)
		}
	} else {
		out.Unset("  (file not found)\n")
	}

	fmt.Println()

	// Print state info
	out.PrintHeader("State")
	out.PrintPath("path", statePath)

	if _, err := os.Stat(statePath); err == nil {
		state, err := config.LoadState(statePath)
		if err != nil {
			out.Error("  Error loading state: %v\n", err)
		} else {
			printStateSummary(out, state)
		}
	} else {
		out.Unset("  (file not found)\n")
	}

	return nil
}

func printConfigSummary(out *output.Printer, s config.Summary) {
	out.Label("  Inner gaps:\n")
	out.PrintKeyValue("horizontal", formatOptionalInt(s.InnerHorizontal))
	out.PrintKeyValue("vertical", formatOptionalInt(s.InnerVertical))

	out.Label("  Outer gaps:\n")
	out.PrintKeyValue("top", formatOptionalInt(s.OuterTop))
	out.PrintKeyValue("bottom", formatOptionalInt(s.OuterBottom))

	if len(s.LeftGaps) > 0 {
		out.Label("  Left (per-monitor):\n")
		for _, g := range s.LeftGaps {
			out.Printf("    ")
			out.Label("%s: ", g.Name)
			out.Value("%d\n", g.Value)
		}
	}

	if len(s.RightGaps) > 0 {
		out.Label("  Right (per-monitor):\n")
		for _, g := range s.RightGaps {
			out.Printf("    ")
			out.Label("%s: ", g.Name)
			out.Value("%d\n", g.Value)
		}
	}
}

func printStateSummary(out *output.Printer, s *config.WorkspaceState) {
	if len(s.Monitors) == 0 {
		out.Unset("  (no monitors configured)\n")
		return
	}

	for name, mon := range s.Monitors {
		out.Label("  %s:\n", name)
		out.Printf("    ")
		out.PrintKeyValue("current", formatOptionalInt(mon.Current))
		out.Printf("    ")
		out.PrintKeyValue("default", formatOptionalInt(mon.Default))
	}
}

func formatOptionalInt(v *int64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
