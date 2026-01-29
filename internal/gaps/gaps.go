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

// ErrInvalidShift indicates a shift that would result in negative gaps.
var ErrInvalidShift = errors.New("shift would result in negative gaps")

// ShiftedGaps holds the calculated left and right gap sizes.
type ShiftedGaps struct {
	LeftGapPercent  int64
	RightGapPercent int64
	LeftGapPixels   int64
	RightGapPixels  int64
}

// ValidateShift checks if a shift is valid for the given workspace percentage.
// The shift is expressed as a percentage of monitor width and cannot exceed the
// centered per-side gap (in pixels), otherwise one side would go negative.
func ValidateShift(monitorWidth, percentage, shiftPercent int64) error {
	baseGapPixels := CalculateGapSize(monitorWidth, percentage)
	shiftPixels := int64(math.Round(float64(monitorWidth) * math.Abs(float64(shiftPercent)) / 100.0))
	if shiftPixels > baseGapPixels {
		return ErrInvalidShift
	}
	return nil
}

// CalculateShiftedGaps computes the left and right gaps with a shift applied.
//
// A positive shift moves the workspace to the right (increasing left gap, decreasing right gap).
// A negative shift moves the workspace to the left (decreasing left gap, increasing right gap).
//
// The calculation is done in pixels so the workspace width remains constant (in pixels)
// for a given percentage.
func CalculateShiftedGaps(monitorWidth, percentage, shiftPercent int64) ShiftedGaps {
	baseGapPixels := CalculateGapSize(monitorWidth, percentage)
	shiftPixels := int64(math.Round(float64(monitorWidth) * float64(shiftPercent) / 100.0))

	leftGapPixels := baseGapPixels + shiftPixels
	rightGapPixels := baseGapPixels - shiftPixels

	leftPercent := int64(math.Round(float64(leftGapPixels) * 100.0 / float64(monitorWidth)))
	rightPercent := int64(math.Round(float64(rightGapPixels) * 100.0 / float64(monitorWidth)))

	return ShiftedGaps{
		LeftGapPercent:  leftPercent,
		RightGapPercent: rightPercent,
		LeftGapPixels:   leftGapPixels,
		RightGapPixels:  rightGapPixels,
	}
}
