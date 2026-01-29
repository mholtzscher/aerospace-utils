// Package config handles aerospace.toml and state file operations.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	ErrConfigRead      = errors.New("failed to read config file")
	ErrConfigParse     = errors.New("failed to parse config file")
	ErrConfigWrite     = errors.New("failed to write config file")
	ErrMonitorNotFound = errors.New("monitor not found in config")
)

// AerospaceService abstracts config file resolution, loading, and writing.
type AerospaceService struct {
	configPath string
	config     *aerospaceConfig // lazily loaded
}

// NewAerospaceService creates a service. If explicitPath is empty, uses DefaultConfigPath().
func NewAerospaceService(explicitPath string) *AerospaceService {
	path := explicitPath
	if path == "" {
		path = DefaultConfigPath()
	}
	path = ExpandPath(path)

	return &AerospaceService{
		configPath: path,
	}
}

// ConfigPath returns the resolved config file path.
func (as *AerospaceService) ConfigPath() string {
	return as.configPath
}

// loadConfig loads the config from disk if not already loaded.
func (as *AerospaceService) loadConfig() error {
	if as.config != nil {
		return nil
	}

	content, err := os.ReadFile(as.configPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfigRead, err)
	}

	var parsed map[string]any
	if _, err := toml.Decode(string(content), &parsed); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigParse, err)
	}

	as.config = &aerospaceConfig{
		path:   as.configPath,
		parsed: parsed,
	}

	return nil
}

// Exists returns true if the config file exists.
func (as *AerospaceService) Exists() (bool, error) {
	_, err := os.Stat(as.configPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat config: %w", err)
}

// MonitorGap represents a gap value for a specific monitor.
type MonitorGap struct {
	Name  string
	Value int64
}

// Summary contains extracted gap information from the config.
type Summary struct {
	InnerHorizontal *int64
	InnerVertical   *int64
	OuterTop        *int64
	OuterBottom     *int64
	LeftGaps        []MonitorGap
	RightGaps       []MonitorGap
}

// Summary returns a summary of the gap configuration.
func (as *AerospaceService) Summary() (Summary, error) {
	if err := as.loadConfig(); err != nil {
		return Summary{}, err
	}

	s := Summary{}

	gaps, ok := as.config.parsed["gaps"].(map[string]any)
	if !ok {
		return s, nil
	}

	if inner, ok := gaps["inner"].(map[string]any); ok {
		s.InnerHorizontal = extractInt64(inner["horizontal"])
		s.InnerVertical = extractInt64(inner["vertical"])
	}

	if outer, ok := gaps["outer"].(map[string]any); ok {
		s.OuterTop = extractScalarGap(outer["top"])
		s.OuterBottom = extractScalarGap(outer["bottom"])
		s.LeftGaps = extractMonitorGaps(outer["left"])
		s.RightGaps = extractMonitorGaps(outer["right"])
	}

	return s, nil
}

// MonitorNames returns all monitor names found in the outer gap configuration.
func (as *AerospaceService) MonitorNames() ([]string, error) {
	summary, err := as.Summary()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var names []string

	for _, g := range summary.LeftGaps {
		if !seen[g.Name] {
			seen[g.Name] = true
			names = append(names, g.Name)
		}
	}
	for _, g := range summary.RightGaps {
		if !seen[g.Name] {
			seen[g.Name] = true
			names = append(names, g.Name)
		}
	}

	return names, nil
}

// SetMonitorGaps updates the gap value for a specific monitor.
func (as *AerospaceService) SetMonitorGaps(monitorName string, gapSize int64) error {
	if err := as.loadConfig(); err != nil {
		return err
	}

	leftUpdated := updateMonitorGapInConfig(as.config.parsed, "left", monitorName, gapSize)
	rightUpdated := updateMonitorGapInConfig(as.config.parsed, "right", monitorName, gapSize)
	if !leftUpdated && !rightUpdated {
		return fmt.Errorf("%w: %s", ErrMonitorNotFound, monitorName)
	}

	return nil
}

// SetMonitorAsymmetricGaps updates both left and right gap values for a monitor.
func (as *AerospaceService) SetMonitorAsymmetricGaps(monitorName string, leftGap, rightGap int64) error {
	if err := as.loadConfig(); err != nil {
		return err
	}

	leftUpdated := updateMonitorGapInConfig(as.config.parsed, "left", monitorName, leftGap)
	rightUpdated := updateMonitorGapInConfig(as.config.parsed, "right", monitorName, rightGap)
	if !leftUpdated || !rightUpdated {
		return fmt.Errorf("%w: %s", ErrMonitorNotFound, monitorName)
	}

	return nil
}

func updateMonitorGapInConfig(config map[string]any, side, monitorName string, gapSize int64) bool {
	gaps, ok := config["gaps"].(map[string]any)
	if !ok {
		return false
	}
	outer, ok := gaps["outer"].(map[string]any)
	if !ok {
		return false
	}

	sideArray, ok := asAnySlice(outer[side])
	if !ok {
		return false
	}

	updated := false
	for _, item := range sideArray {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		monitor, ok := m["monitor"].(map[string]any)
		if !ok {
			continue
		}

		if _, exists := monitor[monitorName]; !exists {
			continue
		}

		monitor[monitorName] = gapSize
		updated = true
	}

	return updated
}

// Write writes the config back to disk atomically.
func (as *AerospaceService) Write() error {
	if as.config == nil {
		return errors.New("no config loaded")
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(as.config.parsed); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := WriteAtomic(as.config.path, buf.String()); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigWrite, err)
	}
	return nil
}

// aerospaceConfig holds the loaded config state.
type aerospaceConfig struct {
	path   string
	parsed map[string]any
}

// extractInt64 extracts an int64 from an interface{} value.
func extractInt64(v any) *int64 {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case int64:
		return &val
	case float64:
		i := int64(val)
		return &i
	}
	return nil
}

// extractScalarGap extracts a scalar gap value from various possible types.
func extractScalarGap(v any) *int64 {
	switch val := v.(type) {
	case int64:
		return &val
	case float64:
		i := int64(val)
		return &i
	case []any:
		// Array form - look for scalar default at the end.
		for i := len(val) - 1; i >= 0; i-- {
			if scalar := extractScalarGap(val[i]); scalar != nil {
				return scalar
			}
		}
	}
	return nil
}

// extractMonitorGaps extracts per-monitor gap values from an array.
func extractMonitorGaps(v any) []MonitorGap {
	arr, ok := asAnySlice(v)
	if !ok {
		return nil
	}

	var gaps []MonitorGap
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		monitor, ok := m["monitor"].(map[string]any)
		if !ok {
			continue
		}

		for name, val := range monitor {
			var value int64
			switch v := val.(type) {
			case int64:
				value = v
			case float64:
				value = int64(v)
			default:
				continue
			}
			gaps = append(gaps, MonitorGap{Name: name, Value: value})
		}
	}

	return gaps
}

func asAnySlice(v any) ([]any, bool) {
	switch vv := v.(type) {
	case []any:
		return vv, true
	case []map[string]any:
		out := make([]any, 0, len(vv))
		for i := range vv {
			out = append(out, vv[i])
		}
		return out, true
	default:
		return nil, false
	}
}

// DefaultConfigPath returns the default path to aerospace.toml.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "aerospace.toml")
}

// WriteAtomic writes content to a file atomically using a temporary file.
func WriteAtomic(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create temp file in same directory for atomic rename.
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up temp file on error.
	success := false
	defer func() {
		if !success {
			if err := os.Remove(tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintln(os.Stderr, "Failed to remove temp file:", err)
			}
		}
	}()

	if _, err := tmp.WriteString(content); err != nil {
		if closeErr := tmp.Close(); closeErr != nil {
			return fmt.Errorf("write temp file: %w; close temp file: %v", err, closeErr)
		}
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	success = true
	return nil
}

// ExpandPath expands ~ to the home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
