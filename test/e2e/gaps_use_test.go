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

func TestGapsUseCases(t *testing.T) {
	tests := []e2eCase{
		{
			name:       "help",
			run:        func(t *testing.T) *testutil.Result { return testutil.RunCLI("gaps", "use", "--help") },
			expectExit: 0,
			stdoutContains: []string{
				"use",
				"percent",
			},
		},
		{
			name: "invalid negative",
			run: func(t *testing.T) *testutil.Result {
				return testutil.RunCLI("gaps", "use", "--monitor-width", "1920", "-10")
			},
			expectExit: 1,
		},
		{
			name: "invalid zero",
			run: func(t *testing.T) *testutil.Result {
				return testutil.RunCLI("gaps", "use", "--monitor-width", "1920", "0")
			},
			expectExit: 1,
		},
		{
			name: "invalid over 100",
			run: func(t *testing.T) *testutil.Result {
				return testutil.RunCLI("gaps", "use", "--monitor-width", "1920", "150")
			},
			expectExit: 1,
		},
		{
			name: "invalid non-numeric",
			run: func(t *testing.T) *testutil.Result {
				return testutil.RunCLI("gaps", "use", "--monitor-width", "1920", "abc")
			},
			expectExit: 1,
		},
		{
			name: "display unavailable without width",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "state.toml")
				return testutil.RunCLIWithEnv(
					map[string]string{"PATH": ""},
					"gaps", "use",
					"--dry-run",
					"--state-path", statePath,
					"--no-color",
					"50",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"display detection not available",
			},
		},
		{
			name: "default paths",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "aerospace")
				require.NoError(t, os.MkdirAll(configDir, 0755))

				writeFixture(t, configDir, "aerospace.toml", "aerospace.toml")
				writeFixture(t, configDir, "workspace-size.toml", "state.toml")

				return testutil.RunCLIWithEnv(
					map[string]string{"HOME": tmpDir},
					"gaps", "use",
					"--dry-run",
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"50%",
				"dry",
			},
		},
		{
			name: "legacy state format",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "state.toml")
				stateContent := "[workspace]\ncurrent = 40\ndefault = 30\n"
				require.NoError(t, os.WriteFile(statePath, []byte(stateContent), 0644))

				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"40%",
			},
		},
		{
			name: "boundary minimum",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"1",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"1%",
			},
		},
		{
			name: "boundary maximum",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"100",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"100%",
			},
		},
		func() e2eCase {
			var configPath string
			var configData []byte
			var statePath string
			return e2eCase{
				name: "dry run",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, configData = setupConfigAndState(t, "aerospace.toml", "state.toml")
					return testutil.RunCLI("gaps", "use",
						"--dry-run",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"50",
					)
				},
				expectExit: 0,
				stdoutContains: []string{
					"dry",
					"50%",
					"480px",
				},
				assert: func(t *testing.T, _ *testutil.Result) {
					afterConfig, err := os.ReadFile(configPath)
					require.NoError(t, err)
					assert.Equal(t, string(configData), string(afterConfig))
				},
			}
		}(),
		{
			name: "from state",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"50%",
				"480px",
			},
		},
		{
			name: "no percentage without state",
			run: func(t *testing.T) *testutil.Result {
				configPath, _, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--config-path", configPath,
					"--state-path", filepath.Join(filepath.Dir(configPath), "nonexistent.toml"),
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"no percentage specified",
			},
		},
		{
			name: "default fallback",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				stateContent := "[monitors.main]\ndefault = 75\n"
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte(stateContent), 0644))

				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"75%",
				"240px",
			},
		},
		{
			name: "empty state with percentage",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte(""), 0644))

				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"50",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"50%",
				"480px",
			},
		},
		{
			name: "empty state without percentage",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace.toml")
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte(""), 0644))

				return testutil.RunCLI("gaps", "use",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"no percentage",
			},
		},
		func() e2eCase {
			var configPath string
			var statePath string
			return e2eCase{
				name: "set default",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, _ = setupConfigAndState(t, "aerospace.toml", "state.toml")
					return testutil.RunCLI("gaps", "use",
						"--no-reload",
						"--set-default",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"75",
					)
				},
				expectExit: 0,
				stdoutContains: []string{
					"75%",
				},
				assert: func(t *testing.T, _ *testutil.Result) {
					afterConfig, err := os.ReadFile(configPath)
					require.NoError(t, err)
					assert.Contains(t, string(afterConfig), "monitor.main = 240")
					assert.GreaterOrEqual(t, strings.Count(string(afterConfig), "monitor.main = 240"), 2)

					afterState, err := os.ReadFile(statePath)
					require.NoError(t, err)
					assert.Contains(t, string(afterState), "current = 75")
					assert.Contains(t, string(afterState), "default = 75")
				},
			}
		}(),
		{
			name: "no reload",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--no-reload",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"50",
				)
			},
			expectExit: 0,
		},
		{
			name: "with monitor",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "aerospace-multi-monitor.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--monitor", "Built-in Retina Display",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "2560",
					"--no-color",
					"60",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"Built-in Retina Display",
				"60%",
			},
		},
		{
			name: "unknown monitor",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLI("gaps", "use",
					"--dry-run",
					"--monitor", "NonExistent Monitor",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"50",
				)
			},
			expectExit: 0,
			stdoutContains: []string{
				"NonExistent Monitor",
			},
		},
		{
			name: "invalid config",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				configPath := writeFixture(t, tmpDir, "aerospace.toml", "invalid.toml")
				statePath := writeFixture(t, tmpDir, "state.toml", "state.toml")

				return testutil.RunCLI("gaps", "use",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"50",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"load config",
			},
		},
		{
			name: "invalid state format",
			run: func(t *testing.T) *testutil.Result {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "state.toml")
				require.NoError(t, os.WriteFile(statePath, []byte("not=toml"), 0644))

				return testutil.RunCLI("gaps", "use",
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"50",
				)
			},
			expectExit: 1,
			stderrContains: []string{
				"load state",
			},
		},
		{
			name: "reload missing aerospace",
			run: func(t *testing.T) *testutil.Result {
				configPath, statePath, _ := setupConfigAndState(t, "aerospace.toml", "state.toml")
				return testutil.RunCLIWithEnv(
					map[string]string{"PATH": ""},
					"gaps", "use",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"60",
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
					"gaps", "use",
					"--config-path", configPath,
					"--state-path", statePath,
					"--monitor-width", "1920",
					"--no-color",
					"60",
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
				name: "actual write",
				run: func(t *testing.T) *testutil.Result {
					configPath, statePath, configData = setupConfigAndState(t, "aerospace.toml", "state.toml")
					return testutil.RunCLI("gaps", "use",
						"--no-reload",
						"--config-path", configPath,
						"--state-path", statePath,
						"--monitor-width", "1920",
						"--no-color",
						"60",
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

func TestGapsUseWithAerospace(t *testing.T) {
	testutil.SkipIfNoAerospace(t)
	t.Skip("requires controlled environment")
}
