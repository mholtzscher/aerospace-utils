//go:build !darwin && !linux

package display

import "errors"

// ErrUnsupportedPlatform indicates display detection is not available.
var ErrUnsupportedPlatform = errors.New("display detection is only available on macOS")

// Enumerate returns an error on non-macOS platforms.
func Enumerate() ([]Info, error) {
	return nil, ErrUnsupportedPlatform
}

// MainWidth returns an error on non-macOS platforms.
func MainWidth() (int64, error) {
	return 0, ErrUnsupportedPlatform
}

// available indicates whether display detection is available on this platform.
const available = false

// Available returns true if display detection is supported on this platform.
func Available() bool {
	return available
}
