#!/usr/bin/env bash
# install-timers.sh — instala hydrapotion + systemd user timer en ~1 minuto.
#
#   ./install-timers.sh           → instala binario + timer de recordatorios
#   ./install-timers.sh --gui     → además autostart de la GUI al login
#
# Requiere: systemd con sesión user (cualquier distro desktop moderna).
set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN_SRC="$REPO_DIR/hydrapotion"
BIN_DST="$HOME/.local/bin/hydrapotion"
UNIT_DST="$HOME/.config/systemd/user"

echo "==> Hydrapotion: instalando timer de recordatorios (systemd user)"

# 1. Binario
if [ ! -x "$BIN_SRC" ]; then
  echo "    binario no encontrado en $BIN_SRC — compilando..."
  (cd "$REPO_DIR" && go build -o hydrapotion -ldflags="-s -w" .)
fi
mkdir -p "$HOME/.local/bin"
install -m 0755 "$BIN_SRC" "$BIN_DST"
echo "    binario: $BIN_DST"

# 2. Unidades
mkdir -p "$UNIT_DST"
install -m 0644 "$REPO_DIR/packaging/systemd/hydrapotion-reminder.service" "$UNIT_DST/"
install -m 0644 "$REPO_DIR/packaging/systemd/hydrapotion-reminder.timer"    "$UNIT_DST/"
echo "    unidades: $UNIT_DST/hydrapotion-reminder.{service,timer}"

# 3. Timer activo
systemctl --user daemon-reload
systemctl --user enable --now hydrapotion-reminder.timer

# 4. GUI autostart (opcional)
if [ "$1" = "--gui" ]; then
  install -m 0644 "$REPO_DIR/packaging/systemd/hydrapotion.service" "$UNIT_DST/"
  systemctl --user enable hydrapotion.service
  echo "    GUI autostart: activado (arranca en el próximo login)"
fi

echo
echo "✔ Listo. Verifica con:"
echo "    systemctl --user status hydrapotion-reminder.timer"
echo "    systemctl --user list-timers | grep hydrapotion"
echo "    journalctl --user -u hydrapotion-reminder.service -f"
echo
echo "Recuerda habilitar lingering NO es necesario (timer de usuario corre con tu sesión)."
