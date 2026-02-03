#!/usr/bin/env zsh

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Aerospace Workspace Size
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🖥️
# @raycast.argument1 { "type": "text", "placeholder": "Percentage (1-100)", "percentBarLabel": "Usable Workspace Size", "optional": true }

# Documentation:
# @raycast.description Set Aerospace workspace size by percentage of monitor width (defaults to 40%)
# @raycast.author mholtzscher
# @raycast.authorURL https://raycast.com/mholtzscher

# Use 40% as default if no argument provided
PERCENTAGE="${1:-40}"

aerospace-utils workspace use $PERCENTAGE
