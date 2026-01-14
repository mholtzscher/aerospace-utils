// Package testutil provides test helpers for E2E CLI testing.
package testutil

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	binaryPath string
	buildOnce  sync.Once
	buildErr   error
)

// Result holds the output from running the CLI.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// BuildCLI compiles the CLI binary once for all tests.
// Call this in TestMain to build early and fail fast.
func BuildCLI(t *testing.T) {
	t.Helper()

	buildOnce.Do(func() {
		root, err := findModuleRoot()
		if err != nil {
			buildErr = err
			return
		}

		tmpDir, err := os.MkdirTemp("", "aerospace-utils-test-*")
		if err != nil {
			buildErr = err
			return
		}

		binaryPath = filepath.Join(tmpDir, "aerospace-utils")

		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = &BuildError{Output: string(out), Err: err}
			return
		}
	})

	if buildErr != nil {
		t.Fatalf("failed to build CLI: %v", buildErr)
	}
}

// BuildError wraps a build failure with its output.
type BuildError struct {
	Output string
	Err    error
}

func (e *BuildError) Error() string {
	return e.Err.Error() + ": " + e.Output
}

// Cleanup removes the built binary. Call in TestMain after tests run.
func Cleanup() {
	if binaryPath != "" {
		_ = os.RemoveAll(filepath.Dir(binaryPath))
	}
}

// RunCLI executes the CLI with the given arguments.
func RunCLI(args ...string) *Result {
	return RunCLIWithEnv(nil, args...)
}

// RunCLIWithEnv executes the CLI with custom environment variables.
func RunCLIWithEnv(env map[string]string, args ...string) *Result {
	if binaryPath == "" {
		return &Result{
			Stderr:   "CLI binary not built - call BuildCLI first",
			ExitCode: 1,
		}
	}

	cmd := exec.Command(binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	return result
}

// findModuleRoot walks up from current dir to find go.mod.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// HasAerospace checks if the aerospace binary is available.
func HasAerospace() bool {
	_, err := exec.LookPath("aerospace")
	return err == nil
}

// SkipIfNoAerospace skips the test if aerospace is not available.
func SkipIfNoAerospace(t *testing.T) {
	t.Helper()
	if !HasAerospace() {
		t.Skip("aerospace binary not found in PATH")
	}
}
