// Package display provides monitor/display detection functionality.
package display

// Info contains information about a display.
type Info struct {
	ID    uint32 // CoreGraphics display ID (macOS)
	Name  string // Human-readable display name
	Width int64  // Display width in pixels
	Main  bool   // Whether this is the main/primary display
}
