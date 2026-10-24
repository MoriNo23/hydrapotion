# Hydrapotion

Desktop hydration tracker — smart, personal, medical-grade logging.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne_v2-2B2836?style=flat)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Quick logging** — 150ml / 250ml / 500ml one-click buttons
- **Smart goal** — base 35ml/kg + mood bonus + weather bonus (temp, humidity)
- **Weekly chart** — bar chart with goal line + weekly average, medical-grade visual
- **Mood tracking** — log mood per day (😊 Bien, 😐 Neutral, 😞 Bajo, 😰 Tenso)
- **Dark theme** — clean dark UI, easy on the eyes
- **System tray** — minimize to tray, quick-add from context menu
- **Smart reminders** — dynamic interval adapting to time-of-day (×0.7 afternoon, ×2.0 night) and weather (×0.5 if ≥35°C). Configurable base: 1h–3h.
- **systemd timer mode (Linux)** — reminders run headless via `hydrapotion --remind` + systemd user timer: the GUI can stay closed and the reminder still fires (~10 MB, <1s per tick instead of a full GUI process resident in RAM)
- **CLI** — `hydrapotion add 250` y `hydrapotion status` para loguear desde la terminal
- **Low RAM** — `GOMEMLIMIT`/`GOGC` tuned, `-s -w` binary, headless reminders without OpenGL
- **Weather** — auto-fetches temperature & humidity via Open-Meteo
- **Export** — CSV (`Fecha,ml,Estado de animo`) & JSON for doctor visits
- **Sound effects** — ASMR keyboard sounds on interaction
- **Single binary** — no Node.js, no Electron, no runtime deps

## Installation

### Dependencies

```bash
# Debian / Ubuntu
sudo apt install golang-go libgl1-mesa-dev xorg-dev pulseaudio libnotify-bin
```

### From source

```bash
git clone https://github.com/MoriNo23/hydrapotion.git
cd hydrapotion
go build -o ~/bin/hydrapotion -ldflags="-s -w" .
~/bin/hydrapotion
```

Or run directly:

```bash
go run .
```

### Pre-built binary

Download from [Releases](https://github.com/MoriNo23/hydrapotion/releases):

```bash
chmod +x hydrapotion-linux-amd64
./hydrapotion-linux-amd64
```

Optional: copy to PATH and add `.desktop` entry for app launcher.

## Usage

1. Open app → set **weight** and **height** in settings (gear icon)
2. Click **150ml / 250ml / 500ml** to log water
3. Select your **mood** — affects daily goal
4. Check **weekly chart** below the cards
5. **Export data** (CSV / JSON) from settings for medical records

Data is stored in `~/.hydrapotion/` (settings.json + history.json + state.json + weather.json).

## CLI (headless)

La GUI es opcional. El mismo binario expone subcomandos sin abrir ventanas:

```bash
hydrapotion add 250     # registra 250 ml
hydrapotion status      # progreso del día + clima cacheado
hydrapotion --remind    # evalúa si toca recordatorio y sale (lo llama systemd)
hydrapotion --version
```

## Recordatorios con systemd timer (Linux)

Los recordatorios ya no requieren la GUI residente. Un timer de usuario ejecuta
`hydrapotion --remind` cada 15 min; el binario decide con la lógica dinámica
(hora del día + clima + última notificación) si envía la notificación y termina.

```bash
# instalar binario + timer (~1 min):
./packaging/systemd/install-timers.sh

# opcional: autostart de la GUI al login:
./packaging/systemd/install-timers.sh --gui

# verificar:
systemctl --user list-timers | grep hydrapotion
journalctl --user -u hydrapotion-reminder.service -f

# desinstalar:
./packaging/systemd/uninstall-timers.sh
```

Consumo típico: GUI ~150–250 MB residente; modo timer ~10 MB durante <1 s
cada 15 min. La GUI (y su RAM) solo la abres cuando quieres ver el gráfico.

## Presupuesto de RAM

| Modo | RAM residente | Notas |
|------|---------------|-------|
| GUI (ventana abierta) | ~150–250 MB | `GOMEMLIMIT=192MiB`, `GOGC=50` dentro del binario |
| GUI en systray | ~80–150 MB | cierra la ventana con la X: queda en tray |
| Solo timer (GUI cerrada) | ~0 MB | tick headless de ~10 MB durante <1 s |

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| GUI | Fyne v2 (develop) |
| Storage | JSON in `~/.hydrapotion/` |
| Reminders | systemd user timer (Linux) + timer interno en GUI |
| Weather | Open-Meteo API (free) |
| Notifications | notify-send |
| Sound | paplay (PulseAudio) |

## CI

Every push builds the binary via GitHub Actions. Releases include a pre-compiled `hydrapotion-linux-amd64` asset.

## License

MIT
