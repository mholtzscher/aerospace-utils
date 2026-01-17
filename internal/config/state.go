package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var (
	ErrStateRead    = errors.New("failed to read state file")
	ErrStateFormat  = errors.New("unrecognized state file format")
	ErrStateMarshal = errors.New("failed to marshal state")
	ErrStateWrite   = errors.New("failed to write state file")
)

// WorkspaceService abstracts state file resolution, loading, and writing.
type WorkspaceService struct {
	statePath string
	state     *workspaceState // lazily loaded
}

// NewWorkspaceService creates a service. If explicitPath is empty, uses DefaultStatePath().
func NewWorkspaceService(explicitPath string) *WorkspaceService {
	path := explicitPath
	if path == "" {
		path = DefaultStatePath()
	}
	path = ExpandPath(path)

	return &WorkspaceService{
		statePath: path,
	}
}

// StatePath returns the resolved state file path.
func (ws *WorkspaceService) StatePath() string {
	return ws.statePath
}

// Exists returns true if the state file exists.
func (ws *WorkspaceService) Exists() (bool, error) {
	_, err := os.Stat(ws.statePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat state: %w", err)
}

// loadState loads the state from disk if not already loaded.
func (ws *WorkspaceService) loadState() error {
	if ws.state != nil {
		return nil
	}

	state := &workspaceState{
		path:     ws.statePath,
		monitors: make(map[string]*MonitorState),
	}

	data, err := os.ReadFile(ws.statePath)
	if os.IsNotExist(err) {
		ws.state = state
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStateRead, err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		ws.state = state
		return nil
	}

	var file stateFile
	if err := toml.Unmarshal(data, &file); err == nil && len(file.Monitors) > 0 {
		state.monitors = file.Monitors
		ws.state = state
		return nil
	}

	return ErrStateFormat
}

// GetMonitorState returns the state for a specific monitor.
func (ws *WorkspaceService) GetMonitorState(monitor string) (*MonitorState, error) {
	if err := ws.loadState(); err != nil {
		return nil, err
	}
	return ws.getOrCreateMonitor(monitor), nil
}

// getOrCreateMonitor returns the state for a specific monitor, creating it if needed.
func (ws *WorkspaceService) getOrCreateMonitor(name string) *MonitorState {
	if ws.state.monitors == nil {
		ws.state.monitors = make(map[string]*MonitorState)
	}
	if ws.state.monitors[name] == nil {
		ws.state.monitors[name] = &MonitorState{}
	}
	return ws.state.monitors[name]
}

// ResolvePercentage returns the percentage to use for a monitor.
// Priority: explicit > current > default.
func (ws *WorkspaceService) ResolvePercentage(monitor string, explicit *int64) (*int64, error) {
	if err := ws.loadState(); err != nil {
		return nil, err
	}

	if explicit != nil {
		return explicit, nil
	}

	mon := ws.state.monitors[monitor]
	if mon == nil {
		return nil, nil
	}

	if mon.Current != nil {
		return mon.Current, nil
	}
	return mon.Default, nil
}

// Update updates the percentage for a monitor and writes to disk.
func (ws *WorkspaceService) Update(monitor string, percentage int64, setDefault bool) error {
	if err := ws.loadState(); err != nil {
		return err
	}

	mon := ws.getOrCreateMonitor(monitor)
	mon.Current = &percentage

	if setDefault || mon.Default == nil {
		mon.Default = &percentage
	}

	return ws.write()
}

// write writes the state to disk.
func (ws *WorkspaceService) write() error {
	file := stateFile{Monitors: ws.state.monitors}

	data, err := toml.Marshal(file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStateMarshal, err)
	}

	if err := WriteAtomic(ws.state.path, string(data)); err != nil {
		return fmt.Errorf("%w: %w", ErrStateWrite, err)
	}
	return nil
}

// Monitors returns all monitor states. Useful for display/iteration.
func (ws *WorkspaceService) Monitors() (map[string]*MonitorState, error) {
	if err := ws.loadState(); err != nil {
		return nil, err
	}
	return ws.state.monitors, nil
}

// workspaceState holds per-monitor workspace percentages.
type workspaceState struct {
	path     string
	monitors map[string]*MonitorState
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

// DefaultStatePath returns the default path to the state file.
func DefaultStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "aerospace-utils-state.toml")
}
