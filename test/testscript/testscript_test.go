// Package testscript exercises the CLI using go-internal/testscript.
package testscript

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mholtzscher/aerospace-utils/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func runCmd() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"aerospace-utils": runCmd,
	})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 "scripts",
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			// Extract testscript's bin directory from PATH so scripts can
			// construct custom PATHs that still include aerospace-utils.
			path := env.Getenv("PATH")
			if parts := strings.SplitN(path, string(os.PathListSeparator), 2); len(parts) > 0 {
				env.Setenv("TESTSCRIPT_BIN", parts[0])
			}

			return nil
		},
	})
}
