//go:build !darwin

package windows

import "context"

// Reset returns an unsupported platform error outside macOS.
func Reset(_ context.Context) (ResetResult, error) {
	return ResetResult{}, ErrUnsupportedPlatform
}
