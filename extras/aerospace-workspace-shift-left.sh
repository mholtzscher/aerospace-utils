#!/usr/bin/env zsh

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Aerospace Workspace Shift Left
# @raycast.mode silent

# Optional parameters:
# @raycast.icon ⬅️
# @raycast.argument1 { "type": "text", "placeholder": "Shift amount (default: 5)", "optional": true }

# Documentation:
# @raycast.description Shift Aerospace workspace left by percentage (default 5%, negative value)
# @raycast.author mholtzscher
# @raycast.authorURL https://raycast.com/mholtzscher

# Use 5 as default and coerce to negative
AMOUNT="${1:-5}"
AMOUNT="-${AMOUNT#-}"

aerospace-utils workspace shift --by $AMOUNT
