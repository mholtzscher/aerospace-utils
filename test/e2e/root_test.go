package e2e

import (
	"testing"

	"github.com/mholtzscher/aerospace-utils/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRootCommandCases(t *testing.T) {
	type commandCase struct {
		name             string
		args             []string
		env              map[string]string
		expectExit       int
		stdoutContains   []string
		stderrContains   []string
		stdoutNotContain []string
	}

	configPath := testdataPath(t, "aerospace.toml")
	statePath := testdataPath(t, "state.toml")

	tests := []commandCase{
		{
			name:       "help",
			args:       []string{"--help"},
			expectExit: 0,
			stdoutContains: []string{
				"aerospace-utils",
				"gaps",
				"Flags:",
			},
		},
		{
			name:       "version",
			args:       []string{"--version"},
			expectExit: 0,
			stdoutContains: []string{
				"aerospace-utils",
			},
		},
		{
			name:       "unknown command",
			args:       []string{"unknown"},
			expectExit: 1,
			stderrContains: []string{
				"unknown command",
			},
		},
		{
			name:       "gaps help",
			args:       []string{"gaps", "--help"},
			expectExit: 0,
			stdoutContains: []string{
				"gaps",
				"use",
				"adjust",
				"current",
			},
		},
		{
			name:       "gaps unknown subcommand",
			args:       []string{"gaps", "unknown"},
			expectExit: 0,
			stdoutContains: []string{
				"Available Commands",
			},
		},
		{
			name:       "no-color flag",
			args:       []string{"gaps", "current", "--config-path", configPath, "--state-path", statePath, "--no-color"},
			expectExit: 0,
			stdoutNotContain: []string{
				"\x1b[",
			},
		},
		{
			name:       "no-color env var",
			args:       []string{"gaps", "current", "--config-path", configPath, "--state-path", statePath},
			env:        map[string]string{"NO_COLOR": "1"},
			expectExit: 0,
			stdoutNotContain: []string{
				"\x1b[",
			},
		},
		{
			name:       "verbose flag",
			args:       []string{"gaps", "current", "--config-path", configPath, "--state-path", statePath, "--verbose", "--no-color"},
			expectExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *testutil.Result
			if tt.env != nil {
				result = testutil.RunCLIWithEnv(tt.env, tt.args...)
			} else {
				result = testutil.RunCLI(tt.args...)
			}

			assert.Equal(t, tt.expectExit, result.ExitCode)
			for _, snippet := range tt.stdoutContains {
				assert.Contains(t, result.Stdout, snippet)
			}
			for _, snippet := range tt.stderrContains {
				assert.Contains(t, result.Stderr, snippet)
			}
			for _, snippet := range tt.stdoutNotContain {
				assert.NotContains(t, result.Stdout, snippet)
			}
		})
	}
}
