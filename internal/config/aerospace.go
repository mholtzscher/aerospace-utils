// Package config handles aerospace.toml and state file operations.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
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

	var parsed aerospaceConfigData
	if err := toml.Unmarshal(content, &parsed); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigParse, err)
	}

	as.config = &aerospaceConfig{
		path:    as.configPath,
		content: string(content),
		parsed:  parsed,
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

// Summary returns a summary of the gap configuration.
func (as *AerospaceService) Summary() (Summary, error) {
	if err := as.loadConfig(); err != nil {
		return Summary{}, err
	}

	s := Summary{
		InnerHorizontal: as.config.parsed.Gaps.Inner.Horizontal,
		InnerVertical:   as.config.parsed.Gaps.Inner.Vertical,
	}

	s.OuterTop = extractScalarGap(as.config.parsed.Gaps.Outer.Top)
	s.OuterBottom = extractScalarGap(as.config.parsed.Gaps.Outer.Bottom)
	s.LeftGaps = extractMonitorGaps(as.config.parsed.Gaps.Outer.Left)
	s.RightGaps = extractMonitorGaps(as.config.parsed.Gaps.Outer.Right)

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
// Uses regex-based replacement to preserve file formatting.
func (as *AerospaceService) SetMonitorGaps(monitorName string, gapSize int64) error {
	if err := as.loadConfig(); err != nil {
		return err
	}

	// Build regex patterns for the monitor name
	// Handle both quoted and unquoted monitor names
	var patterns []*regexp.Regexp

	// Pattern for: monitor.main = 123 or monitor.'name' = 123 or monitor."name" = 123
	if monitorName == "main" || isSimpleKey(monitorName) {
		// Unquoted key: monitor.main = 123
		patterns = append(patterns, regexp.MustCompile(
			fmt.Sprintf(`(monitor\.%s\s*=\s*)(\d+)`, regexp.QuoteMeta(monitorName)),
		))
	}

	// Single-quoted key: monitor.'name' = 123
	patterns = append(patterns, regexp.MustCompile(
		fmt.Sprintf(`(monitor\.'%s'\s*=\s*)(\d+)`, regexp.QuoteMeta(monitorName)),
	))

	// Double-quoted key: monitor."name" = 123
	patterns = append(patterns, regexp.MustCompile(
		fmt.Sprintf(`(monitor\."%s"\s*=\s*)(\d+)`, regexp.QuoteMeta(monitorName)),
	))

	// Try each pattern
	newValue := strconv.FormatInt(gapSize, 10)
	replaced := false

	for _, pattern := range patterns {
		if pattern.MatchString(as.config.content) {
			as.config.content = pattern.ReplaceAllString(as.config.content, "${1}"+newValue)
			replaced = true
		}
	}

	if !replaced {
		return fmt.Errorf("%w: %s", ErrMonitorNotFound, monitorName)
	}

	return nil
}

// Write writes the config back to disk atomically.
func (as *AerospaceService) Write() error {
	if as.config == nil {
		return errors.New("no config loaded")
	}

	if err := WriteAtomic(as.config.path, as.config.content); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigWrite, err)
	}
	return nil
}

// aerospaceConfig holds the loaded config state.
type aerospaceConfig struct {
	path    string
	content string // Raw file content for format-preserving edits
	parsed  aerospaceConfigData
}

// aerospaceConfigData is the parsed structure of aerospace.toml.
type aerospaceConfigData struct {
	Gaps struct {
		Inner struct {
			Horizontal *int64 `toml:"horizontal"`
			Vertical   *int64 `toml:"vertical"`
		} `toml:"inner"`
		Outer struct {
			Top    interface{} `toml:"top"`
			Bottom interface{} `toml:"bottom"`
			Left   interface{} `toml:"left"`
			Right  interface{} `toml:"right"`
		} `toml:"outer"`
	} `toml:"gaps"`
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

// extractScalarGap extracts a scalar gap value from various possible types.
func extractScalarGap(v interface{}) *int64 {
	switch val := v.(type) {
	case int64:
		return &val
	case float64:
		i := int64(val)
		return &i
	case []interface{}:
		// Array form - look for scalar default at the end
		for i := len(val) - 1; i >= 0; i-- {
			if scalar := extractScalarGap(val[i]); scalar != nil {
				return scalar
			}
		}
	}
	return nil
}

// extractMonitorGaps extracts per-monitor gap values from an array.
func extractMonitorGaps(v interface{}) []MonitorGap {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}

	var gaps []MonitorGap
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		monitor, ok := m["monitor"].(map[string]interface{})
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

// isSimpleKey returns true if the key doesn't need quoting in TOML.
func isSimpleKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
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

	// Create temp file in same directory for atomic rename
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up temp file on error
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

	// Atomic rename
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
