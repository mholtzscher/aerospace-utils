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

// AerospaceConfig represents the aerospace.toml configuration file.
type AerospaceConfig struct {
	Path    string
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

// LoadAerospaceConfig loads and parses the aerospace.toml file.
func LoadAerospaceConfig(path string) (*AerospaceConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var parsed aerospaceConfigData
	if err := toml.Unmarshal(content, &parsed); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &AerospaceConfig{
		Path:    path,
		content: string(content),
		parsed:  parsed,
	}, nil
}

// Summary returns a summary of the gap configuration.
func (c *AerospaceConfig) Summary() Summary {
	s := Summary{
		InnerHorizontal: c.parsed.Gaps.Inner.Horizontal,
		InnerVertical:   c.parsed.Gaps.Inner.Vertical,
	}

	s.OuterTop = extractScalarGap(c.parsed.Gaps.Outer.Top)
	s.OuterBottom = extractScalarGap(c.parsed.Gaps.Outer.Bottom)
	s.LeftGaps = extractMonitorGaps(c.parsed.Gaps.Outer.Left)
	s.RightGaps = extractMonitorGaps(c.parsed.Gaps.Outer.Right)

	return s
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

// MonitorNames returns all monitor names found in the outer gap configuration.
func (c *AerospaceConfig) MonitorNames() []string {
	seen := make(map[string]bool)
	var names []string

	for _, g := range c.Summary().LeftGaps {
		if !seen[g.Name] {
			seen[g.Name] = true
			names = append(names, g.Name)
		}
	}
	for _, g := range c.Summary().RightGaps {
		if !seen[g.Name] {
			seen[g.Name] = true
			names = append(names, g.Name)
		}
	}

	return names
}

// SetMonitorGaps updates the gap value for a specific monitor.
// Uses regex-based replacement to preserve file formatting.
func (c *AerospaceConfig) SetMonitorGaps(monitorName string, gapSize int64) error {
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
		if pattern.MatchString(c.content) {
			c.content = pattern.ReplaceAllString(c.content, "${1}"+newValue)
			replaced = true
		}
	}

	if !replaced {
		return fmt.Errorf("monitor %q not found in config", monitorName)
	}

	return nil
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

// Write writes the config back to disk atomically.
func (c *AerospaceConfig) Write() error {
	return WriteAtomic(c.Path, c.content)
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

// ErrMonitorNotFound indicates the specified monitor was not found in config.
var ErrMonitorNotFound = errors.New("monitor not found in config")
