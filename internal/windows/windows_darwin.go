//go:build darwin

package windows

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const resetScript = `
set attemptedCount to 0
set movedCount to 0
set failedCount to 0

tell application "Finder"
	set desktopBounds to bounds of window of desktop
end tell

set screenLeft to item 1 of desktopBounds
set screenTop to item 2 of desktopBounds
set screenRight to item 3 of desktopBounds
set screenBottom to item 4 of desktopBounds

set screenW to screenRight - screenLeft
set screenH to screenBottom - screenTop

set targetW to (screenW * 95) div 100
set targetH to (screenH * 95) div 100

set targetX to screenLeft + ((screenW - targetW) div 2)
set targetY to screenTop + ((screenH - targetH) div 2)

if targetX < screenLeft then set targetX to screenLeft
if targetY < screenTop then set targetY to screenTop

tell application "System Events"
	set visibleProcesses to every process whose visible is true and background only is false
	repeat with theProcess in visibleProcesses
		try
			set theWindows to every window of theProcess
			repeat with theWindow in theWindows
				set shouldSkip to false
				try
					set shouldSkip to minimized of theWindow
				on error
					set shouldSkip to false
				end try

				if shouldSkip is false then
					set attemptedCount to attemptedCount + 1
					set didMove to false

					try
						set zoomed of theWindow to false
					end try

					try
						set bounds of theWindow to {targetX, targetY, targetX + targetW, targetY + targetH}
						set didMove to true
					on error
						try
							set size of theWindow to {targetW, targetH}
						end try

						try
							set winSize to size of theWindow
							set winW to item 1 of winSize
							set winH to item 2 of winSize

							set centeredX to screenLeft + ((screenW - winW) div 2)
							set centeredY to screenTop + ((screenH - winH) div 2)

							if winW <= screenW then
								if centeredX < screenLeft then set centeredX to screenLeft
								set maxX to screenRight - winW
								if centeredX > maxX then set centeredX to maxX
							else
								set centeredX to screenLeft
							end if

							if winH <= screenH then
								if centeredY < screenTop then set centeredY to screenTop
								set maxY to screenBottom - winH
								if centeredY > maxY then set centeredY to maxY
							else
								set centeredY to screenTop
							end if

							set position of theWindow to {centeredX, centeredY}
							set didMove to true
						on error
							set didMove to false
						end try
					end try

					if didMove is true then
						set movedCount to movedCount + 1
					else
						set failedCount to failedCount + 1
					end if
				end if
			end repeat
		end try
	end repeat
end tell

return "attempted=" & attemptedCount & " moved=" & movedCount & " failed=" & failedCount
`

// Reset moves all visible app windows to the main display center.
func Reset(ctx context.Context) (ResetResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	binaryPath, err := exec.LookPath("osascript")
	if err != nil {
		return ResetResult{}, errors.New("osascript not found in PATH")
	}

	cmd := exec.CommandContext(ctx, binaryPath, "-e", resetScript)
	combined, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(combined))
		if msg == "" {
			msg = err.Error()
		}

		if looksLikePermissionError(msg) {
			return ResetResult{}, fmt.Errorf(
				"windows reset failed: %s (enable Accessibility + Automation for your terminal)",
				msg,
			)
		}

		return ResetResult{}, fmt.Errorf("windows reset failed: %s", msg)
	}

	result, err := parseResult(string(combined))
	if err != nil {
		return ResetResult{}, err
	}

	return result, nil
}

func parseResult(raw string) (ResetResult, error) {
	line := strings.TrimSpace(raw)
	result := ResetResult{}

	if _, err := fmt.Sscanf(
		line,
		"attempted=%d moved=%d failed=%d",
		&result.Attempted,
		&result.Moved,
		&result.Failed,
	); err != nil {
		return ResetResult{}, fmt.Errorf("parse windows reset result %q: %w", line, err)
	}

	return result, nil
}

func looksLikePermissionError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "not authorized") ||
		strings.Contains(lower, "assistive") ||
		strings.Contains(lower, "accessibility") ||
		strings.Contains(lower, "system events got an error")
}
