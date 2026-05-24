#!/usr/bin/env bash
# Waybar custom module — reads ~/.hydrapotion/status.json
# Usage: add to ~/.config/waybar/config.jsonc:
#   "custom/hydrapotion": {
#       "exec": "$HOME/.config/waybar/scripts/hydrapotion.sh",
#       "return-type": "json",
#       "interval": 5
#   }
#
# CSS classes: hydrapotion (default), hydrapotion-good (>80%), hydrapotion-done (100%)

STATUS="$HOME/.hydrapotion/status.json"
if [ ! -f "$STATUS" ]; then
  echo '{"text":"💧 --","class":"hydrapotion","percentage":0}'
  exit 0
fi

PCT=$(jq -r '.percent // 0' "$STATUS")
CONSUMED=$(jq -r '.consumed // 0' "$STATUS")
GOAL=$(jq -r '.goal // 0' "$STATUS")

# Round to integer before arithmetic
PCT_INT=$(jq -r '.percent // 0 | round' "$STATUS")

# Build 10-segment bar
FILLED=$(( PCT_INT / 10 ))
[ "$FILLED" -gt 10 ] && FILLED=10
EMPTY=$(( 10 - FILLED ))

# Blue filled segments + dark empty segments
FILLED_BAR=$(printf ' %.0s' $(seq 1 $FILLED))
EMPTY_BAR=$(printf ' %.0s' $(seq 1 $EMPTY))

BAR="<span background='#40a9ff' foreground='#40a9ff'>${FILLED_BAR}</span>"
BAR+="<span background='#2a2a3d' foreground='#2a2a3d'>${EMPTY_BAR}</span>"

# Determine CSS class
CLASS="hydrapotion"
if [ "$PCT_INT" -ge 100 ]; then
  CLASS="hydrapotion-done"
elif [ "$PCT_INT" -ge 80 ]; then
  CLASS="hydrapotion-good"
fi

TEXT="💧 ${PCT_INT}% ${BAR}"
jq -n --arg text "$TEXT" --arg cls "$CLASS" --arg pct "$PCT_INT" \
  '{text: $text, class: $cls, percentage: $pct}'
