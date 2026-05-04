// Package windows provides macOS window reset operations.
package windows

import "errors"

// ErrUnsupportedPlatform indicates window reset is unavailable on this OS.
var ErrUnsupportedPlatform = errors.New("windows reset is only supported on macOS")

// ResetResult contains counters from a reset operation.
type ResetResult struct {
	Attempted int
	Moved     int
	Failed    int
}
