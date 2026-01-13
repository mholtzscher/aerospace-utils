// Package aerospace provides interaction with the aerospace binary.
package aerospace

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Binary represents the aerospace CLI binary.
type Binary struct {
	path string
}

// ErrNotFound indicates the aerospace binary was not found.
var ErrNotFound = errors.New("aerospace binary not found in PATH")

// FindBinary locates the aerospace binary in PATH.
func FindBinary() (*Binary, error) {
	path, err := exec.LookPath("aerospace")
	if err != nil {
		return nil, ErrNotFound
	}

	// Verify it's executable
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat aerospace: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("aerospace at %s is not executable", path)
	}

	return &Binary{path: path}, nil
}

// Path returns the path to the aerospace binary.
func (b *Binary) Path() string {
	return b.path
}

// ReloadConfig runs `aerospace reload-config`.
func (b *Binary) ReloadConfig() error {
	cmd := exec.Command(b.path, "reload-config")

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("aerospace reload-config failed: %s", string(output))
		}
		return fmt.Errorf("aerospace reload-config failed: %w", err)
	}

	return nil
}

// DefaultConfigPath returns the default aerospace config path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "aerospace.toml")
}
