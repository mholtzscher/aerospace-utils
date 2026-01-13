//go:build linux

package display

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// xrandr output pattern: "DP-1 connected primary 2560x1440+0+0 ..."
var connectedPattern = regexp.MustCompile(`^(\S+)\s+connected\s+(primary\s+)?(\d+)x(\d+)`)

// Enumerate returns information about all active displays using xrandr.
func Enumerate() ([]Info, error) {
	cmd := exec.Command("xrandr", "--query")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("xrandr failed: %w (is xrandr installed?)", err)
	}

	var displays []Info
	var primaryFound bool

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		matches := connectedPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		name := matches[1]
		isPrimary := strings.TrimSpace(matches[2]) == "primary"
		width, _ := strconv.ParseInt(matches[3], 10, 64)

		if isPrimary {
			primaryFound = true
		}

		displays = append(displays, Info{
			ID:    uint32(len(displays)),
			Name:  name,
			Width: width,
			Main:  isPrimary,
		})
	}

	// If no primary found, mark the first display as main
	if !primaryFound && len(displays) > 0 {
		displays[0].Main = true
	}

	if len(displays) == 0 {
		return nil, errors.New("no displays found via xrandr")
	}

	return displays, nil
}

// MainWidth returns the width of the primary display.
func MainWidth() (int64, error) {
	displays, err := Enumerate()
	if err != nil {
		return 0, err
	}

	for _, d := range displays {
		if d.Main {
			return d.Width, nil
		}
	}

	return 0, errors.New("no primary display found")
}

// Available returns true since xrandr-based detection is available on Linux.
func Available() bool {
	_, err := exec.LookPath("xrandr")
	return err == nil
}
