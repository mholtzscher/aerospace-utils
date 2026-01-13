package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// WorkspaceState holds per-monitor workspace percentages.
type WorkspaceState struct {
	Path     string
	Monitors map[string]*MonitorState
}

// MonitorState holds the current and default percentage for a monitor.
type MonitorState struct {
	Current *int64 `toml:"current,omitempty"`
	Default *int64 `toml:"default,omitempty"`
}

// stateFile is the TOML structure for the state file.
type stateFile struct {
	Monitors map[string]*MonitorState `toml:"monitors"`
}

// legacyStateFile is the old TOML structure for migration.
type legacyStateFile struct {
	Workspace struct {
		Current *int64 `toml:"current"`
		Default *int64 `toml:"default"`
	} `toml:"workspace"`
}

// LoadState loads the workspace state from disk.
// Supports legacy format migration.
func LoadState(path string) (*WorkspaceState, error) {
	state := &WorkspaceState{
		Path:     path,
		Monitors: make(map[string]*MonitorState),
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state file: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return state, nil
	}

	// Try new format first: [monitors.main]
	var newFormat stateFile
	if err := toml.Unmarshal(data, &newFormat); err == nil && len(newFormat.Monitors) > 0 {
		state.Monitors = newFormat.Monitors
		return state, nil
	}

	// Try legacy TOML format: [workspace]
	var legacy legacyStateFile
	if err := toml.Unmarshal(data, &legacy); err == nil {
		if legacy.Workspace.Current != nil || legacy.Workspace.Default != nil {
			state.Monitors["main"] = &MonitorState{
				Current: legacy.Workspace.Current,
				Default: legacy.Workspace.Default,
			}
			return state, nil
		}
	}

	// Try plain integer (oldest format)
	if val, err := strconv.ParseInt(content, 10, 64); err == nil {
		state.Monitors["main"] = &MonitorState{Current: &val}
		return state, nil
	}

	return nil, fmt.Errorf("unrecognized state file format")
}

// GetMonitor returns the state for a specific monitor, creating it if needed.
func (s *WorkspaceState) GetMonitor(name string) *MonitorState {
	if s.Monitors == nil {
		s.Monitors = make(map[string]*MonitorState)
	}
	if s.Monitors[name] == nil {
		s.Monitors[name] = &MonitorState{}
	}
	return s.Monitors[name]
}

// ResolvePercentage returns the percentage to use for a monitor.
// Priority: explicit > current > default
func (s *WorkspaceState) ResolvePercentage(monitorName string, explicit *int64) *int64 {
	if explicit != nil {
		return explicit
	}

	mon := s.Monitors[monitorName]
	if mon == nil {
		return nil
	}

	if mon.Current != nil {
		return mon.Current
	}
	return mon.Default
}

// Update updates the percentage for a monitor.
func (s *WorkspaceState) Update(monitorName string, percentage int64, setDefault bool) {
	mon := s.GetMonitor(monitorName)
	mon.Current = &percentage

	if setDefault || mon.Default == nil {
		mon.Default = &percentage
	}
}

// Write writes the state to disk.
func (s *WorkspaceState) Write() error {
	file := stateFile{Monitors: s.Monitors}

	data, err := toml.Marshal(file)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return WriteAtomic(s.Path, string(data))
}

// DefaultStatePath returns the default path to the state file.
func DefaultStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "workspace-size.toml")
}
