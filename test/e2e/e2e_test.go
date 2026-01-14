// Package e2e contains end-to-end tests for the CLI.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type e2eCase struct {
	name           string
	run            func(t *testing.T) *testutil.Result
	expectExit     int
	stdoutContains []string
	stderrContains []string
	assert         func(t *testing.T, result *testutil.Result)
}

func TestMain(m *testing.M) {
	testutil.BuildCLI(&testing.T{})
	code := m.Run()
	testutil.Cleanup()
	os.Exit(code)
}

func runCases(t *testing.T, tests []e2eCase) {
	t.Helper()
	for _, tt := range tests {
		testCase := tt
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := testCase.run(t)
			assert.Equal(t, testCase.expectExit, result.ExitCode)
			for _, snippet := range testCase.stdoutContains {
				assert.Contains(t, result.Stdout, snippet)
			}
			for _, snippet := range testCase.stderrContains {
				assert.Contains(t, result.Stderr, snippet)
			}
			if testCase.assert != nil {
				testCase.assert(t, result)
			}
		})
	}
}

func testdataPath(t *testing.T, filename string) string {
	t.Helper()
	absPath, err := filepath.Abs(filepath.Join("testdata", filename))
	require.NoError(t, err)
	return absPath
}

func writeFixture(t *testing.T, dir, name, fixture string) string {
	t.Helper()
	data, err := os.ReadFile(testdataPath(t, fixture))
	require.NoError(t, err)
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func setupConfigAndState(t *testing.T, configFixture, stateFixture string) (string, string, []byte) {
	t.Helper()
	tmpDir := t.TempDir()
	configData, err := os.ReadFile(testdataPath(t, configFixture))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))
	statePath := writeFixture(t, tmpDir, "state.toml", stateFixture)
	return configPath, statePath, configData
}
