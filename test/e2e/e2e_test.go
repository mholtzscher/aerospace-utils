// Package e2e contains end-to-end tests for the CLI.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutil.BuildCLI(&testing.T{})
	code := m.Run()
	testutil.Cleanup()
	os.Exit(code)
}

func testdataPath(t *testing.T, filename string) string {
	t.Helper()
	absPath, err := filepath.Abs(filepath.Join("testdata", filename))
	require.NoError(t, err)
	return absPath
}
