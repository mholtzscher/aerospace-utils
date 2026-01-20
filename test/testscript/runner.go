package testscript

import (
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

// CLIPath returns the built CLI path. Call BuildCLI first.
func CLIPath(t *testing.T) string {
	t.Helper()
	if binaryPath == "" {
		t.Fatal("CLI binary not built - call BuildCLI first")
	}
	return binaryPath
}

// Cleanup removes the built binary. Call in TestMain after tests run.
func Cleanup() {
	if binaryPath != "" {
		_ = os.RemoveAll(filepath.Dir(binaryPath))
	}
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
