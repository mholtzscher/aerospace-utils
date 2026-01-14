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

func TestGapsCurrentHelp(t *testing.T) {
	result := testutil.RunCLI("gaps", "current", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "current")
}

func TestGapsCurrent(t *testing.T) {
	configPath := testdataPath(t, "aerospace.toml")
	statePath := testdataPath(t, "state.toml")

	result := testutil.RunCLI("gaps", "current",
		"--config-path", configPath,
		"--state-path", statePath,
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Config")
	assert.Contains(t, result.Stdout, "State")
	assert.Contains(t, result.Stdout, "path: "+configPath)
	assert.Contains(t, result.Stdout, "path: "+statePath)
	assert.Contains(t, result.Stdout, "horizontal: 10")
	assert.Contains(t, result.Stdout, "vertical: 10")
	assert.Contains(t, result.Stdout, "top: 10")
	assert.Contains(t, result.Stdout, "bottom: 10")
	assert.Contains(t, result.Stdout, "main: 100")
	assert.Contains(t, result.Stdout, "current: 50")
	assert.Contains(t, result.Stdout, "default: 50")
}

func TestGapsCurrentMissingConfig(t *testing.T) {
	result := testutil.RunCLI("gaps", "current",
		"--config-path", "/nonexistent/path/aerospace.toml",
		"--state-path", "/nonexistent/path/state.toml",
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "(file not found)")
	assert.GreaterOrEqual(t, strings.Count(result.Stdout, "(file not found)"), 2)
}

func TestGapsCurrentInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	invalidConfig, err := os.ReadFile(testdataPath(t, "invalid.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), invalidConfig, 0644))

	result := testutil.RunCLI("gaps", "current",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "nonexistent.toml"),
		"--no-color",
	)

	// current should succeed but show error loading config
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Error loading config")
	assert.Contains(t, result.Stdout, "(file not found)")
}

func TestGapsCurrentMultiMonitor(t *testing.T) {
	tmpDir := t.TempDir()

	configData, err := os.ReadFile(testdataPath(t, "aerospace-multi-monitor.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "aerospace.toml"), configData, 0644))

	stateData, err := os.ReadFile(testdataPath(t, "state.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "state.toml"), stateData, 0644))

	result := testutil.RunCLI("gaps", "current",
		"--config-path", filepath.Join(tmpDir, "aerospace.toml"),
		"--state-path", filepath.Join(tmpDir, "state.toml"),
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	// Should show multiple monitors in config
	assert.Contains(t, result.Stdout, "Built-in Retina Display")
	assert.Contains(t, result.Stdout, "LG UltraFine")
	assert.Contains(t, result.Stdout, "Built-in Retina Display: 200")
	assert.Contains(t, result.Stdout, "LG UltraFine: 150")
	// Should show state for multiple monitors
	assert.Contains(t, result.Stdout, "main")
	assert.Contains(t, result.Stdout, "current: 75")
	assert.Contains(t, result.Stdout, "default: 60")
}
