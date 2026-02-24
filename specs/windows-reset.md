# Windows Reset (macOS)

## Problem
After disabling Aerospace, some app windows remain positioned off-screen (or on a now-missing monitor/space). This makes apps effectively unusable until each window is manually dragged back.

## Goal
Add a recovery command that:
- targets all windows (best-effort) across all visible apps
- moves them onto the **main monitor**
- resizes to a reasonable default so they fit
- centers them inside the main monitor viewport

Command name: `aerospace-utils windows reset`.

## Non-goals
- Perfectly restore prior layout/geometry.
- Move/resize windows of apps that disallow programmatic window management.
- Handle minimized windows in a special way (they are already reachable via Dock).
- Linux/Windows window management support.

## UX
### CLI
`aerospace-utils windows reset`

Uses existing global flags:
- `--dry-run`: print intent only; no window changes.
- `--no-color`: plain output.
- `--verbose`: print extra info (e.g., counts, osascript invocation failures).

### Output
Success (single line): `Reset <N> windows (moved+centered on main monitor)`

If nothing moved: `No windows to reset`.

Common failure message should include actionable hint:
- Missing `osascript`: `osascript not found in PATH`
- Accessibility / Automation denied: suggest enabling Accessibility for the invoking terminal and/or allowing Automation for `System Events`.

## Behavior
### Target set
- Enumerate `System Events` processes where `visible is true` and `background only is false`.
- For each process, iterate `every window`.
- Best-effort: per-window failures must not abort the whole run.

Default: skip minimized windows when the `minimized` property exists and is true (avoid disturbing Dock-only windows).

### Geometry
All positioning targets the **main monitor**.

Source of main monitor bounds:
- Prefer `tell application "Finder" to get bounds of window of desktop`.
- Treat returned bounds as `{screenLeft, screenTop, screenRight, screenBottom}`.

Sizing defaults:
- `targetW = 0.95 * screenW`
- `targetH = 0.95 * screenH`

Algorithm per window:
1. Try `set size of window to {targetW, targetH}`.
2. Read `size` after resizing attempt (some windows clamp or reject sizes).
3. Compute centered position:
   - `x = screenLeft + (screenW - winW) / 2`
   - `y = screenTop + (screenH - winH) / 2`
4. Clamp position so window stays within main bounds when possible:
   - if `winW <= screenW`: `x in [screenLeft, screenRight - winW]` else `x = screenLeft`
   - if `winH <= screenH`: `y in [screenTop, screenBottom - winH]` else `y = screenTop`
5. `set position of window to {x, y}`.

If resize fails for a window, still attempt to center+clamp using current `size`.

### macOS permissions
Requires macOS Accessibility permission for the caller (Terminal / iTerm / etc.) so `System Events` can set window geometry.

## Implementation
### Packages / files
Add internal implementation:
- `internal/windows/windows_darwin.go`
  - `Reset(ctx context.Context) (ResetResult, error)`
  - Executes `osascript` with a single inline AppleScript program.
  - Parses/returns counts for reporting.
- `internal/windows/windows_other.go`
  - `Reset(...)` returns a clear unsupported-platform error.

Add CLI wiring:
- `cmd/windows/windows.go` parent command (`windows`)
- `cmd/windows/reset.go` subcommand (`reset`)
- `cmd/root.go` registers `windows.NewCommand()` next to `workspace.NewCommand()`

### AppleScript payload
Single script run (avoid per-window `osascript` overhead).
- Accumulate counters: attempted windows, moved windows, failed windows.
- Return a compact string the Go code can parse, e.g. `attempted=12 moved=9 failed=3`.

### Error handling
- If `osascript` exits non-zero: wrap stderr; detect common permission-denied strings; append hint.
- If bounds query fails: abort with error (no safe default bounds).

## Testing (e2e only)
Add `test/testscript/scripts/windows-reset.txtar`.

Strategy:
- Provide a fake `osascript` in `$WORK/bin/osascript` that:
  - records invocation
  - prints a deterministic counter line
  - exits 0
- `env PATH=$WORK/bin:$PATH`
- `exec aerospace-utils windows reset --no-color`
- Assert stdout contains `Reset` and uses the mocked counters.

Add a second script (or extend same) for missing `osascript`:
- `env PATH=$TESTSCRIPT_BIN` (no fake; no system osascript assumed)
- `exec` expects non-zero exit and stderr contains `osascript not found`.

## Acceptance Criteria
- `aerospace-utils windows reset` exists and is discoverable via `--help`.
- On macOS with permissions, windows are moved to main monitor and centered; resize attempt made.
- Per-window failures do not abort the run.
- Non-darwin builds compile; running command returns unsupported error.
- `just check` passes.
