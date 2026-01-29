# aerospace-utils

A CLI tool to dynamically adjust [Aerospace](https://github.com/nikitabobko/AeroSpace) workspace gaps, allowing you to center your workspace with adjustable margins.

## Features

- **Dynamic Resizing**: Set workspace width as a percentage of your monitor width.
- **Auto-Centering**: Automatically calculates equal left and right outer gaps.
- **Position Shifting**: Shift the workspace left/right by a percentage while keeping the same workspace width.
- **State Management**: Remembers your current settings and default preferences.
- **Aerospace Integration**: Automatically updates `aerospace.toml` and reloads the configuration.

## Installation

### Prerequisites

- macOS (primary support, uses CoreGraphics for monitor detection).
- Linux (experimental support via `xrandr` for development).
- [Aerospace](https://github.com/nikitabobko/AeroSpace) installed and in your `PATH`.
- Go 1.22+ (if building from source).

### Install via Nix (Recommended)

```bash
nix build
nix run
```

### Build from Source

```bash
go build -o aerospace-utils .
```

### Development Environment

This project uses [direnv](https://direnv.net/) and [Nix](https://nixos.org/) for automatic environment setup.
If you have `direnv` and `nix` installed:

```bash
direnv allow
```

This will provide a shell with all necessary dependencies (Go, golangci-lint, etc.) and development scripts (e.g., `dev-build`, `dev-test`).

## Usage

### Set Workspace Size

Set the workspace to use a specific percentage of the monitor width. The remaining space is divided equally as gaps on the left and right.

If no percentage is provided and the state file is missing or empty, the tool defaults to 60% on first run.

```bash
# Set workspace to 80% of monitor width (10% gap on each side)
aerospace-utils workspace use 80

# Set size and save as default for future adjustments
aerospace-utils workspace use 80 --set-default

# Target a specific monitor by name
aerospace-utils workspace use 70 --monitor "Dell U2722D"
```

### Adjust Size

Incrementally increase or decrease the current workspace size.

```bash
# Increase workspace width by 5% (default)
aerospace-utils workspace adjust

# Increase workspace width by 10%
aerospace-utils workspace adjust -b 10

# Decrease workspace width by 5%
aerospace-utils workspace adjust -b -5
aerospace-utils workspace adjust --by=-5

# Adjust on a specific monitor
aerospace-utils workspace adjust -b 5 --monitor "Dell U2722D"
```

### Shift Position

Shift the workspace left or right by a percentage while keeping the same workspace width.

```bash
# Shift workspace 5% left (decrease left gap, increase right gap)
aerospace-utils workspace shift -b -5

# Shift workspace 5% right (increase left gap, decrease right gap)
aerospace-utils workspace shift -b 5

# Reset shift back to centered
aerospace-utils workspace shift

# Shift on a specific monitor
aerospace-utils workspace shift -b 5 --monitor "Dell U2722D"
```

### View Configuration

Display the current resolved paths, calculated gaps, and saved state.

```bash
aerospace-utils workspace current
```

### Global Options

These options are available for all commands:

- `--monitor <NAME>`: Target specific monitor (default: "main").
- `--dry-run`: Print actions without modifying files or reloading Aerospace.
- `--verbose`: Show detailed processing information.
- `--no-reload`: Skip the `aerospace reload-config` command after updating configuration.
- `--no-color`: Disable colored output.
- `--config-path <PATH>`: Manually specify `aerospace.toml` path.
- `--state-path <PATH>`: Manually specify `aerospace-utils-state.toml` path.
- `--monitor-width <PX>`: Override automatic monitor width detection (advanced).

## How it Works

The tool detects your main monitor's width and calculates the outer gaps required to achieve the desired workspace percentage.

**Example:**
- Monitor Width: `3000px`
- Desired Workspace: `80%` (2400px)
- Total Gap: `20%` (600px)
- Gap per side: `300px`

```
           Monitor Width (100%)
┌──────────────────────────────────────────┐
│      Gap        Workspace        Gap     │
│    ┌─────┐┌──────────────────┐┌─────┐    │
│    │     ││                  ││     │    │
│    │ 300 ││       2400       ││ 300 │    │
│    │     ││                  ││     │    │
│    └─────┘└──────────────────┘└─────┘    │
└──────────────────────────────────────────┘
```

It updates the `[gaps.outer.left]` and `[gaps.outer.right]` settings for the target monitor (default: `monitor.main`) in your `aerospace.toml`.

Note: when writing, the tool re-encodes `aerospace.toml` (comments/formatting may change).

### Shifting Example

You can keep the same workspace width but shift it left/right by redistributing the side gaps.

**Example:**
- Monitor Width: `3000px`
- Desired Workspace: `50%` (1500px)
- Centered gaps: `25%` each (750px)
- Shift left by `5%`:
  - Left gap: `20%` (600px)
  - Right gap: `30%` (900px)

```
            Monitor Width (100%)
┌──────────────────────────────────────────┐
│   Gap        Workspace           Gap     │
│ ┌─────┐┌──────────────────┐┌────────┐   │
│ │     ││                  ││        │   │
│ │ 600 ││       1500       ││  900   │   │
│ │     ││                  ││        │   │
│ └─────┘└──────────────────┘└────────┘   │
└──────────────────────────────────────────┘
```

## Configuration Files

1.  **`aerospace.toml`**: The tool modifies this file to apply the gaps.
    *   It expects `[gaps.outer.left]` and `[gaps.outer.right]` to be arrays.
    *   It targets the entry matching the specified monitor name (default: "main").

2.  **`aerospace-utils-state.toml`**: Stores the current percentage and default preference.
    *   Default location: `~/.config/aerospace/aerospace-utils-state.toml`
