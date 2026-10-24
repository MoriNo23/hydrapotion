#!/usr/bin/env bash
# uninstall-timers.sh — revierte install-timers.sh por completo.
set -e

UNIT_DST="$HOME/.config/systemd/user"

echo "==> Desactivando hydrapotion timers/servicios"
systemctl --user disable --now hydrapotion-reminder.timer 2>/dev/null || true
systemctl --user disable --now hydrapotion.service 2>/dev/null || true
rm -f "$UNIT_DST/hydrapotion-reminder.service" \
      "$UNIT_DST/hydrapotion-reminder.timer" \
      "$UNIT_DST/hydrapotion.service"
systemctl --user daemon-reload
echo "✔ Unidades eliminadas. El binario se conserva en ~/.local/bin/hydrapotion"
echo "  (bórralo con: rm ~/.local/bin/hydrapotion)"
