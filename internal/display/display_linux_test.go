//go:build linux

package display

import (
	"testing"
)

func TestEnumerate(t *testing.T) {
	if !Available() {
		t.Skip("xrandr not available")
	}

	displays, err := Enumerate()
	if err != nil {
		t.Fatalf("Enumerate() error: %v", err)
	}

	if len(displays) == 0 {
		t.Fatal("expected at least one display")
	}

	t.Logf("Found %d displays:", len(displays))
	for _, d := range displays {
		t.Logf("  - %s: %dpx wide, main=%v", d.Name, d.Width, d.Main)
	}

	// Verify at least one is marked as main
	var hasMain bool
	for _, d := range displays {
		if d.Main {
			hasMain = true
			break
		}
	}
	if !hasMain {
		t.Error("expected at least one display to be marked as main")
	}
}

func TestMainWidth(t *testing.T) {
	if !Available() {
		t.Skip("xrandr not available")
	}

	width, err := MainWidth()
	if err != nil {
		t.Fatalf("MainWidth() error: %v", err)
	}

	if width <= 0 {
		t.Errorf("MainWidth() = %d; want > 0", width)
	}

	t.Logf("Main display width: %d", width)
}
