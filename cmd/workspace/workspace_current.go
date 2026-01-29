package workspace

import (
	"context"
	"fmt"

	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/config"
	"github.com/mholtzscher/aerospace-utils/internal/output"
	ufcli "github.com/urfave/cli/v3"
)

func newCurrentCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "current",
		Usage: "Show current gap configuration and state",
		Description: `Display the current aerospace gap configuration and workspace state.

Shows:
- Config file path and gap values
- State file path and per-monitor percentages`,
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			return runCurrent(ctx, cmd)
		},
	}
}

func runCurrent(ctx context.Context, cmd *ufcli.Command) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)

	configSvc := config.NewAerospaceService(opts.ConfigPath)
	stateSvc := config.NewWorkspaceService(opts.StatePath)

	// Print config info
	out.PrintHeader("Config")
	out.PrintPath("path", configSvc.ConfigPath())

	exists, err := configSvc.Exists()
	if err != nil {
		out.Error("  Error checking config: %v\n", err)
	} else if exists {
		summary, err := configSvc.Summary()
		if err != nil {
			out.Error("  Error loading config: %v\n", err)
		} else {
			printConfigSummary(out, summary)
		}
	} else {
		out.Unset("  (file not found)\n")
	}

	fmt.Println()

	// Print state info
	out.PrintHeader("State")
	out.PrintPath("path", stateSvc.StatePath())

	exists, err = stateSvc.Exists()
	if err != nil {
		out.Error("  Error checking state: %v\n", err)
	} else if exists {
		monitors, err := stateSvc.Monitors()
		if err != nil {
			out.Error("  Error loading state: %v\n", err)
		} else {
			printMonitorsSummary(out, monitors)
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

func printMonitorsSummary(out *output.Printer, monitors map[string]*config.MonitorState) {
	if len(monitors) == 0 {
		out.Unset("  (no monitors configured)\n")
		return
	}

	for name, mon := range monitors {
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
