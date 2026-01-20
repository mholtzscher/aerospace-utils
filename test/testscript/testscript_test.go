// Package testscript exercises the CLI using go-internal/testscript.
package testscript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	gotestscript "github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testutil.BuildCLI(&testing.T{})
	code := m.Run()
	testutil.Cleanup()
	os.Exit(code)
}

func TestScripts(t *testing.T) {
	cliPath := testutil.CLIPath(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	testdataDir := filepath.Clean(filepath.Join(cwd, "testdata"))

	gotestscript.Run(t, gotestscript.Params{
		Dir: "scripts",
		Setup: func(env *gotestscript.Env) error {
			cliDir := filepath.Dir(cliPath)
			env.Setenv("PATH", cliDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			env.Setenv("AEROSPACE_TESTDATA", testdataDir)
			env.Setenv("AEROSPACE_CLI", cliPath)
			return nil
		},
	})
}
