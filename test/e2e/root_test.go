package e2e

import (
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHelp(t *testing.T) {
	result := testutil.RunCLI("--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "aerospace-utils")
	assert.Contains(t, result.Stdout, "gaps")
	assert.Contains(t, result.Stdout, "Flags:")
}

func TestVersion(t *testing.T) {
	result := testutil.RunCLI("--version")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "aerospace-utils")
}

func TestUnknownCommand(t *testing.T) {
	result := testutil.RunCLI("unknown")

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "unknown command")
}

func TestGapsHelp(t *testing.T) {
	result := testutil.RunCLI("gaps", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "gaps")
	assert.Contains(t, result.Stdout, "use")
	assert.Contains(t, result.Stdout, "adjust")
	assert.Contains(t, result.Stdout, "current")
}

func TestGapsUnknownSubcommand(t *testing.T) {
	result := testutil.RunCLI("gaps", "unknown")

	assert.Equal(t, 0, result.ExitCode) // Cobra shows help for unknown subcommands
	assert.Contains(t, result.Stdout, "Available Commands")
}

func TestNoColorFlag(t *testing.T) {
	configPath := testdataPath(t, "aerospace.toml")
	statePath := testdataPath(t, "state.toml")

	result := testutil.RunCLI("gaps", "current",
		"--config-path", configPath,
		"--state-path", statePath,
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.NotContains(t, result.Stdout, "\x1b[")
}

func TestNoColorEnvVar(t *testing.T) {
	configPath := testdataPath(t, "aerospace.toml")
	statePath := testdataPath(t, "state.toml")

	result := testutil.RunCLIWithEnv(
		map[string]string{"NO_COLOR": "1"},
		"gaps", "current",
		"--config-path", configPath,
		"--state-path", statePath,
	)

	assert.Equal(t, 0, result.ExitCode)
	assert.NotContains(t, result.Stdout, "\x1b[")
}

func TestVerboseFlag(t *testing.T) {
	configPath := testdataPath(t, "aerospace.toml")
	statePath := testdataPath(t, "state.toml")

	result := testutil.RunCLI("gaps", "current",
		"--config-path", configPath,
		"--state-path", statePath,
		"--verbose",
		"--no-color",
	)

	assert.Equal(t, 0, result.ExitCode)
}
