package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGapsUseHelp(t *testing.T) {
	result := testutil.RunCLI("gaps", "use", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "use")
	assert.Contains(t, result.Stdout, "percent")
}

func TestGapsUseInvalidPercentage(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{"negative", "-10"},
		{"zero", "0"},
		{"over 100", "150"},
		{"non-numeric", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutil.RunCLI("gaps", "use", "--monitor-width", "1920", tt.arg)
			assert.NotEqual(t, 0, result.ExitCode)
		})
	}
}

func TestGapsUseBoundaryPercentages(t *testing.T) {
	tests := []struct {
		name    string
		percent string
	}{
		{"minimum valid (1%)", "1"},
		{"maximum valid (100%)", "100"},
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

			result := testutil.RunCLI("gaps", "use",
				"--dry-run",
				"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
				"--state-path", filepath.Join(tmpDir, "state.toml"),
				"--monitor-width", "1920",
				"--no-color",
				tt.percent,
			)

			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, tt.percent+"%")
		})
	}
}

func TestGapsUseDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"50",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "dry")

	afterConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, string(configData), string(afterConfig))
}

func TestGapsUseFromState(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "50%") // uses current from state
}

func TestGapsUseNoPercentageNoState(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "nonexistent.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "no percentage specified")
}

func TestGapsUseDefaultFallback(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	// Create state with only default, no current
	stateContent := "[monitors.main]\ndefault = 75\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), []byte(stateContent), 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	// Should use the default percentage (75%)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "75%")
}

func TestGapsUseEmptyState(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	// Create empty state file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), []byte(""), 0644))

	// With explicit percentage, should succeed
	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"50",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "50%")
}

func TestGapsUseEmptyStateNoPercentage(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	// Create empty state file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), []byte(""), 0644))

	// Without explicit percentage and empty state, should fail
	result := testutil.RunCLI("gaps", "use",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
	)

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "no percentage")
}

func TestGapsUseSetDefault(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--set-default",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"75",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "75%")
}

func TestGapsUseNoReload(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--no-reload",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"50",
	)

	assert.Equal(t, 0, result.ExitCode)
}

func TestGapsUseWithMonitor(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace-multi-monitor.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--monitor", "Built-in Retina Display",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "2560",
		"--no-color",
		"60",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Built-in Retina Display")
	assert.Contains(t, result.Stdout, "60%")
}

func TestGapsUseUnknownMonitor(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--dry-run",
		"--monitor", "NonExistent Monitor",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"50",
	)

	// Should succeed in dry-run - the monitor name is just used for state tracking
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "NonExistent Monitor")
}

func TestGapsUseInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	invalidConfig, err := os.ReadFile(testdataPath(t, "invalid.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), invalidConfig, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "use",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--monitor-width", "1920",
		"--no-color",
		"50",
	)

	assert.NotEqual(t, 0, result.ExitCode)
}

func TestGapsUseActualWrite(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace.toml"))
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, "aerospace.toml")
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	statePath := filepath.Join(tmpDir, "state.toml")
	require.NoError(t, os.WriteFile(statePath, stateData, 0644))

	// Run without --dry-run but with --no-reload (to avoid needing aerospace binary)
	result := testutil.RunCLI("gaps", "use",
		"--no-reload",
		"--config-path", configPath,
		"--state-path", statePath,
		"--monitor-width", "1920",
		"--no-color",
		"60",
	)

	assert.Equal(t, 0, result.ExitCode)

	// Verify config was actually modified
	afterConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotEqual(t, string(configData), string(afterConfig))

	// Verify state was actually modified
	afterState, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.Contains(t, string(afterState), "60")
}

func TestGapsUseWithAerospace(t *testing.T) {
	testutil.SkipIfNoAerospace(t)
	t.Skip("requires controlled environment")
}
