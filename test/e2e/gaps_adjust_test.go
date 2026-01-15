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

func TestGapsAdjustCases(t *testing.T) {
	tests := []e2eCase{
		{
			name:       "help",
			run:        func(t *testing.T) *testutil.Result { return testutil.RunCLI("gaps", "adjust", "--help") },
			expectExit: 0,
			stdoutContains: []string{
				"adjust",
			},
		},
		func() e2eCase {
			var configPath string
			var statePath string
			var configData []byte
			return e2eCase{
				name: "dry run",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, configData = setupConfigAndState(t, "aerospace.toml", "state.toml")
					result := testutil.RunCLI("gaps", "adjust",
						"--dry-run",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"--by=10",
					)
					return result
				},
				expectExit: 0,
				stdoutContains: []string{
					"dry",
					"60%",
					"384px",
				},
				assert: func(t *testing.T, _ *testutil.Result) {
					afterConfig, err := os.ReadFile(configPath)
					require.NoError(t, err)
					assert.Equal(t, string(configData), string(afterConfig))
				},
			}
		}(),
		{
			name: "by flag positive",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "adjust",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=10",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"60%",
			},
		},
		{
			name: "by flag negative",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "adjust",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=-10",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"40%",
			},
		},
		{
			name: "by flag short",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "adjust",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"-b", "5",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"55%",
			},
		},
		{
			name: "default adjustment",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "adjust",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"55%",
				"432px",
			},
		},
		{
			name: "default state path",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "aerospace")
				require.NoError(t, os.MkdirAll(configDir, 0755))

				writeFixture(t, configDir, "workspace-size.toml", "state.toml")

				return testutil.RunCLIWithEnv(
					map[string]string{"HOME": tmpDir},
					"gaps", "adjust",
					"--dry-run",
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"55%",
				"432px",
			},
		},
		{
			name: "no state",
			run: func(t *testing.T) *testutil.Result {
				configPath, _, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "adjust",
					"--config-path", configPath,
					"--state-path", filepath.Join(filepath.Dir(configPath), "nonexistent.toml"),
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"no current percentage",
			},
		},
		{
			name: "invalid state format",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte("not=toml"), 0644))

				return testutil.RunCLI("gaps", "adjust",
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"load state",
			},
		},
		{
			name: "overflow",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				stateContent := "[monitors.main]\ncurrent = 95\ndefault = 50\n"
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte(stateContent), 0644))

				return testutil.RunCLI("gaps", "adjust",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=10",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"invalid",
			},
		},
		{
			name: "underflow",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				stateContent := "[monitors.main]\ncurrent = 5\ndefault = 50\n"
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte(stateContent), 0644))

				return testutil.RunCLI("gaps", "adjust",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=-10",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"invalid",
			},
		},
		{
			name: "with monitor",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace-multi-monitor.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLI("gaps", "adjust",
					"--dry-run",
					"--monitor", "Built-in Retina Display",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "2560",
					"--no-color",
					"--by=10",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"Built-in Retina Display",
				"85%",
				"192px",
			},
		},
		{
			name: "unknown monitor without state",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLI("gaps", "adjust",
					"--monitor", "NonExistent Monitor",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"no current percentage",
			},
		},
		{
			name: "reload missing aerospace",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLIWithEnv(
					map[string]string{"PATH": ""},
					"gaps", "adjust",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=10",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"aerospace not found",
			},
		},
		{
			name: "reload failure",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				binDir := filepath.Join(tmpDir, "bin")
				require.NoError(t, os.MkdirAll(binDir, 0755))

				fakeBinary := filepath.Join(binDir, "aerospace")
				fakeScript := "#!/bin/sh\necho boom\nexit 1\n"
				require.NoError(t, os.WriteFile(fakeBinary, []byte(fakeScript), 0755))

				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLIWithEnv(
					map[string]string{"PATH": binDir},
					"gaps", "adjust",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"--by=10",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"reload failed",
				"boom",
			},
		},
		func() e2eCase {
			var configPath string
			var configData []byte
			var statePath string
			return e2eCase{
				name: "update keeps other config",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, configData = setupConfigAndState(t, "aerospace-extra-config.toml", "state.toml")
					return testutil.RunCLI("gaps", "adjust",
						"--no-reload",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"--by=10",
					)
				},
				expectExit: 0,
				assert: func(t *testing.T, _ *testutil.Result) {
					afterConfig, err := os.ReadFile(configPath)
					require.NoError(t, err)
					expectedConfig := strings.ReplaceAll(string(configData), "monitor.main = 100", "monitor.main = 384")
					assert.Equal(t, expectedConfig, string(afterConfig))
				},
			}
		}(),
		func() e2eCase {
			var configPath string
			var configData []byte
			var statePath string
			return e2eCase{
				name: "actual write",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, configData = setupConfigAndState(t, "aerospace.toml", "state.toml")
					return testutil.RunCLI("gaps", "adjust",
						"--no-reload",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"--by=10",
					)
				},
				expectExit: 0,
				assert: func(t *testing.T, _ *testutil.Result) {
					afterConfig, err := os.ReadFile(configPath)
					require.NoError(t, err)
					assert.NotEqual(t, string(configData), string(afterConfig))
					assert.Contains(t, string(afterConfig), "monitor.main = 384")
					assert.GreaterOrEqual(t, strings.Count(string(afterConfig), "monitor.main = 384"), 2)

					afterState, err := os.ReadFile(statePath)
					require.NoError(t, err)
					assert.Contains(t, string(afterState), "current = 60")
					assert.Contains(t, string(afterState), "default = 50")
				},
			}
		}(),
	}

	runCases(t, tests)
}
