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

func TestGapsAdjustHelp(t *testing.T) {
	result := testutil.RunCLI("gaps", "adjust", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "adjust")
}

func TestGapsAdjustDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--dry-run",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"--by=10",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "dry")
	assert.Contains(t, result.Stdout, "60%")
	assert.Contains(t, result.Stdout, "384px")

	afterConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, string(configData), string(afterConfig))
}

func TestGapsAdjustByFlag(t *testing.T) {
	tests := []struct {
		name           string
		flag           string
		expect         int
		expectedOutput string
	}{
		{"positive", "--by=10", 0, "60%"},
		{"negative", "--by=-10", 0, "40%"},
		{"short flag", "-b", 0, "55%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

			stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

			args := []string{"gaps", "adjust",
				"--dry-run",
				"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
				"--state-path", filepath.Join(tmpDir, "state.toml"),
				"--monitor-width", "1920",
				"--no-color",
			}
			if tt.flag == "-b" {
				args = append(args, "-b", "5")
			} else {
				args = append(args, tt.flag)
			}

			result := testutil.RunCLI(args...)
			assert.Equal(t, tt.expect, result.ExitCode)
			assert.Contains(t, result.Stdout, tt.expectedOutput)
		})
	}
}

func TestGapsAdjustDefault(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--dry-run",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "55%") // 50 + default 5
	assert.Contains(t, result.Stdout, "432px")
}

func TestGapsAdjustNoState(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "nonexistent.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "no current percentage")
}

func TestGapsAdjustOverflow(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	// Create state with current at 95%
	stateContent := "[monitors.main]\ncurrent = 95\ndefault = 50\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), []byte(stateContent), 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"--by=10",
	)

	// 95 + 10 = 105, which exceeds 100
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "invalid")
}

func TestGapsAdjustUnderflow(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	// Create state with current at 5%
	stateContent := "[monitors.main]\ncurrent = 5\ndefault = 50\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), []byte(stateContent), 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"--by=-10",
	)

	// 5 - 10 = -5, which is below 1
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "invalid")
}

func TestGapsAdjustWithMonitor(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace-multi-monitor.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--dry-run",
		"--monitor", "Built-in Retina Display",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "2560",
		"--no-color",
		"--by=10",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Built-in Retina Display")
	// State has current=75 for Built-in Retina Display, so 75+10=85%
	assert.Contains(t, result.Stdout, "85%")
	assert.Contains(t, result.Stdout, "192px")
}

func TestGapsAdjustUnknownMonitorNoState(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "adjust",
		"--monitor", "NonExistent Monitor",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	// Should fail because there's no state for this monitor
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "no current percentage")
}

func TestGapsAdjustReloadMissingAerospace(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	result := testutil.RunCLIWithEnv(
		map[string]string{"PATH": ""},
		"gaps", "adjust",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"--by=10",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "aerospace not found")
}

func TestGapsAdjustReloadFailure(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	fakeBinary := filepath.Join(binDir, "aerospace")
	fakeScript := "#!/bin/sh\necho boom\nexit 1\n"
	require.NoError(t, os.WriteFile(fakeBinary, []byte(fakeScript), 0755))

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	result := testutil.RunCLIWithEnv(
		map[string]string{"PATH": binDir},
		"gaps", "adjust",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"--by=10",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "reload failed")
	assert.Contains(t, result.Stdout, "boom")
}

func TestGapsAdjustActualWrite(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	// Run without --dry-run but with --no-reload
	result := testutil.RunCLI("gaps", "adjust",
		"--no-reload",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"--by=10",
	)

	assert.Equal(t, 0, result.ExitCode)

	// Verify config was actually modified
	afterConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotEqual(t, string(configData), string(afterConfig))
	assert.Contains(t, string(afterConfig), "monitor.main = 384")
	assert.GreaterOrEqual(t, strings.Count(string(afterConfig), "monitor.main = 384"), 2)

	// Verify state was updated to 60 (50 + 10)
	afterState, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.Contains(t, string(afterState), "current = 60")
	assert.Contains(t, string(afterState), "default = 50")
}
