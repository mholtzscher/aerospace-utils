# aerospace-utils extras: Raycast Script Commands

This folder contains [Raycast](https://www.raycast.com/) Script Commands that call `aerospace-utils` for quick workspace size/position tweaks.

Note: the `extras/` folder is not installed by Homebrew; grab these scripts from the repo.

## What they do

- `aerospace-workspace-size.sh`: Set workspace size via `aerospace-utils workspace use <PERCENT>` (default: 40)
- `aerospace-workspace-size-increment.sh`: Increase workspace size via `aerospace-utils workspace adjust --by <AMOUNT>` (default: 5)
- `aerospace-workspace-size-decrement.sh`: Decrease workspace size via `aerospace-utils workspace adjust --by -<AMOUNT>` (default: 5)
- `aerospace-workspace-shift-left.sh`: Shift workspace left via `aerospace-utils workspace shift --by -<AMOUNT>` (default: 5)
- `aerospace-workspace-shift-right.sh`: Shift workspace right via `aerospace-utils workspace shift --by <AMOUNT>` (default: 5)
- `aerospace-workspace-shift-reset.sh`: Reset shift (re-center) via `aerospace-utils workspace shift`

All scripts accept an optional first argument (entered in Raycast) and use a sensible default when omitted.

## Install

1) Ensure `aerospace-utils` is installed and available in `PATH`.

2) Copy the scripts into your Raycast scripts directory:

```bash
mkdir -p ~/.raycast/scripts/aerospace-utils
cp ./extras/aerospace-workspace-*.sh ~/.raycast/scripts/aerospace-utils/
chmod +x ~/.raycast/scripts/aerospace-utils/aerospace-workspace-*.sh
```

3) In Raycast: Preferences -> Extensions -> Script Commands

- Add the directory `~/.raycast/scripts/aerospace-utils`
- Search for commands like "Aerospace Workspace Size" and assign hotkeys if you want

## Requirements

- Raycast (Script Commands)
- `aerospace-utils` installed
- Aerospace installed
