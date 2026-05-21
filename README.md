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
- **Reminders** — configurable (5–60 min) popup with snooze
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
go build -o ~/bin/hydrapotion .
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

Data is stored in `~/.hydrapotion/` (settings.json + history.json).

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| GUI | Fyne v2 (develop) |
| Storage | JSON in `~/.hydrapotion/` |
| Weather | Open-Meteo API (free) |
| Notifications | notify-send |
| Sound | paplay (PulseAudio) |

## CI

Every push builds the binary via GitHub Actions. Releases include a pre-compiled `hydrapotion-linux-amd64` asset.

## License

MIT
