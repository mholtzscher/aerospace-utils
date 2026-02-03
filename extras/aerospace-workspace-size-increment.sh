#!/usr/bin/env zsh

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Aerospace Workspace Increment
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 📈
# @raycast.argument1 { "type": "text", "placeholder": "Increment amount (default: 5)", "optional": true }

# Documentation:
# @raycast.description Increment Aerospace workspace size by percentage (default 5%)
# @raycast.author mholtzscher
# @raycast.authorURL https://raycast.com/mholtzscher

# Use 5 as default if no argument provided
AMOUNT="${1:-5}"
AMOUNT="${AMOUNT#-}"

aerospace-utils workspace adjust --by $AMOUNT
