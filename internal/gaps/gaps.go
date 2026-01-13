// Package gaps provides core gap calculation and validation logic.
package gaps

import (
	"errors"
	"math"
)

// ErrInvalidPercentage indicates a percentage outside the valid range.
var ErrInvalidPercentage = errors.New("percentage must be between 1 and 100")

// ValidatePercentage ensures the percentage is within the valid range (1-100).
func ValidatePercentage(percentage int64) error {
	if percentage < 1 || percentage > 100 {
		return ErrInvalidPercentage
	}
	return nil
}

// CalculateGapSize computes the gap size in pixels from monitor width and percentage.
// Formula: gap = monitor_width * ((100 - percentage) / 100) / 2
// This gives the gap on each side (left and right) to achieve the desired workspace percentage.
func CalculateGapSize(monitorWidth, percentage int64) int64 {
	fraction := float64(100-percentage) / 100.0
	gap := float64(monitorWidth) * fraction / 2.0
	return int64(math.Round(gap))
}
