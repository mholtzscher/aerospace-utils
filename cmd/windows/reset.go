package windows

import (
	"context"

	ufcli "github.com/urfave/cli/v3"

	"github.com/mholtzscher/aerospace-utils/internal/cli"
	"github.com/mholtzscher/aerospace-utils/internal/output"
	resetwindows "github.com/mholtzscher/aerospace-utils/internal/windows"
)

func newResetCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "reset",
		Usage: "Center windows on the main monitor",
		Description: `Resize visible app windows to 95% of the main monitor and center them.

Useful after disabling Aerospace when windows drift off-screen.`,
		Action: runReset,
	}
}

func runReset(ctx context.Context, cmd *ufcli.Command) error {
	opts := cli.GetOptions(cmd)
	out := output.New(opts.NoColor)

	if opts.DryRun {
		out.DryRun()
		out.Printf("Would reset windows to 95%% size centered on main monitor\n")
		return nil
	}

	result, err := resetwindows.Reset(ctx)
	if err != nil {
		return err
	}

	if opts.Verbose {
		out.Printf("attempted=%d moved=%d failed=%d\n", result.Attempted, result.Moved, result.Failed)
	}

	if result.Moved == 0 {
		if result.Attempted > 0 {
			out.Warning("No windows moved (attempted=%d failed=%d)\n", result.Attempted, result.Failed)
			return nil
		}

		out.Warning("No windows to reset\n")
		return nil
	}

	out.Success("Reset %d windows (moved+centered on main monitor)\n", result.Moved)
	return nil
}
