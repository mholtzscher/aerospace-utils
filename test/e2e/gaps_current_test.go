package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGapsCurrentCases(t *testing.T) {
	tests := []e2eCase{
		{
			name:       "help",
			run:        func(t *testing.T) *testutil.Result { return testutil.RunCLI("gaps", "current", "--help") },
			expectExit: 0,
			stdoutContains: []string{
				"current",
			},
		},
		func() e2eCase {
			var configPath string
			var statePath string
			return e2eCase{
				name: "current output",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, _ = setupConfigAndState(t, "aerospace.toml", "state.toml")
					return testutil.RunCLI("gaps", "current",
						"--config-path", configPath,
						"--state-path", statePath,
						"--no-color",
					)
				},
				expectExit: 0,
				stdoutContains: []string{
					"Config",
					"State",
					"horizontal: 10",
					"vertical: 10",
					"top: 10",
					"bottom: 10",
					"main: 100",
					"current: 50",
					"default: 50",
				},
				assert: func(t *testing.T, result *testutil.Result) {
					assert.Contains(t, result.Stdout, "path: "+configPath)
					assert.Contains(t, result.Stdout, "path: "+statePath)
				},
			}
		}(),
		{
			name: "default paths",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "aerospace")
				require.NoError(t, os.MkdirAll(configDir, 0755))

				configPath := writeFixture(t, configDir, "aerospace.toml", "aerospace.toml")
				statePath := writeFixture(t, configDir, "aerospace-utils-state.toml", "state.toml")

				result := testutil.RunCLIWithEnv(
					map[string]string{"HOME": tmpDir},
					"gaps", "current",
					"--no-color",
				)

				assert.Contains(t, result.Stdout, "path: "+configPath)
				assert.Contains(t, result.Stdout, "path: "+statePath)
				return result
			},
			expectExit: 0,
			stdoutContains: []string{
				"current: 50",
			},
		},
		{
			name: "missing config",
			run: func(t *testing.T) *testutil.Result {
				return testutil.RunCLI("gaps", "current",
					"--config-path", "/nonexistent/path/aerospace.toml",
					"--state-path", "/nonexistent/path/state.toml",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"(file not found)",
			},
			assert: func(t *testing.T, result *testutil.Result) {
				assert.GreaterOrEqual(t, strings.Count(result.Stdout, "(file not found)"), 2)
			},
		},
		{
			name: "invalid config",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "invalid.toml")
				return testutil.RunCLI("gaps", "current",
					"--config-path", configPath,
					"--state-path", filepath.Join(tmpDir, "nonexistent.toml"),
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"Error loading config",
				"(file not found)",
			},
		},
		{
			name: "invalid state format",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte("not=toml"), 0644))

				return testutil.RunCLI("gaps", "current",
					"--config-path", testdataPath(t, "aerospace.toml"),
					"--state-path", statePath,
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"Error loading state",
			},
		},
		{
			name: "multi monitor",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace-multi-monitor.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLI("gaps", "current",
					"--config-path", configPath,
					"--state-path", statePath,
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"Built-in Retina Display",
				"LG UltraFine",
				"Built-in Retina Display: 200",
				"LG UltraFine: 150",
				"main",
				"current: 75",
				"default: 60",
			},
		},
	}

	runCases(t, tests)
}
