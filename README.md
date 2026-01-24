# aerospace-utils

A CLI tool to dynamically adjust [Aerospace](https://github.com/nikitabobko/AeroSpace) workspace gaps, allowing you to center your workspace with adjustable margins.

## Features

- **Dynamic Resizing**: Set workspace width as a percentage of your monitor width.
- **Auto-Centering**: Automatically calculates equal left and right outer gaps.
- **State Management**: Remembers your current settings and default preferences.
- **Aerospace Integration**: Automatically updates `aerospace.toml` and reloads the configuration.

## Installation

### Prerequisites

- macOS (required for monitor width detection)
- [Aerospace](https://github.com/nikitabobko/AeroSpace) installed and in your `PATH`.
- Rust toolchain (cargo).

### Install via Cargo

```bash
cargo install --path .
```

## Usage

### Set Workspace Size

Set the workspace to use a specific percentage of the monitor width. The remaining space is divided equally as gaps on the left and right.

```bash
# Set workspace to 80% of monitor width (10% gap on each side)
aerospace-utils workspace use 80

# Set size and save as default for future adjustments
aerospace-utils workspace use 80 --set-default
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
```

### View Configuration

Display the current resolved paths, calculated gaps, and saved state.

```bash
aerospace-utils workspace current
```

### Workspace Options

These options are available under `aerospace-utils workspace`.

- `--dry-run`: Print actions without modifying files or reloading Aerospace.
- `--verbose`: Show detailed processing information.
- `--no-reload`: Skip the `aerospace reload-config` command after updating configuration.
- `--config-path <PATH>`: Manually specify `aerospace.toml` path.
- `--state-path <PATH>`: Manually specify `aerospace-utils-state.toml` path.

#### Advanced / Debug Options
- `--monitor-width <PX>`: Override automatic monitor width detection (useful for testing or non-macOS).

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

It updates the `[gaps.outer.left]` and `[gaps.outer.right]` settings for `monitor.main` in your `aerospace.toml`.

## Configuration Files

1.  **`aerospace.toml`**: The tool modifies this file to apply the gaps.
    *   It expects `[gaps.outer.left]` and `[gaps.outer.right]` to be arrays.
    *   It specifically targets the entry for `monitor.main` (conventionally the second item or an explicit table).

2.  **`aerospace-utils-state.toml`**: Stores the current percentage and default preference.
    *   Default location: `~/.config/aerospace/aerospace-utils-state.toml`
