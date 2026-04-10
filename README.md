# Hydrapotion

Desktop hydration tracker with smart reminders and mood-based recommendations.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne_v2-2B2836?style=flat)

## Features

- **Daily water tracking** - Quick add buttons (150ml, 250ml, 500ml)
- **Dynamic goal calculation**:
  - Base: 35ml per kg of body weight
  - Mood bonus: +200ml (low), +400ml (tense/stressed)
  - Adjusts recommendations based on how you feel
- **System tray integration** - Minimizes to tray, quick add from context menu
- **Smart reminders** - Popup notifications with snooze (5/15/30 min)
- **Weekly progress chart** - Visual bar graph of last 7 days
- **Dark theme** - Easy on the eyes
- **Lightweight** - Single binary, no external dependencies

## Screenshots

```
┌─────────────────────────────────┐
│          Hydrapotion            │
├─────────────────────────────────┤
│     1200 / 2450 ml (49%)        │
│  ████████████░░░░░░░░░░░░░░░░░  │
├─────────────────────────────────┤
│  [150ml]  [250ml]  [500ml]      │
├─────────────────────────────────┤
│  Estado: [Neutral ▼]            │
│  Peso: 70 kg                    │
├─────────────────────────────────┤
│  Progreso semanal:              │
│    ▓▓▓▓▓▓▓                      │
│    ▓▓▓▓▓▓▓▓▓▓                   │
│    ▓▓▓▓▓▓▓▓▓▓▓▓                 │
│   Mon Tue Wed Thu Fri Sat Sun   │
└─────────────────────────────────┘
```

## Installation

### Requirements
- Go 1.21+
- Linux (uses notify-send for notifications)

### Build

```bash
git clone https://github.com/MoriNo23/hydrapotion.git
cd hydrapotion
go build -o hydrapotion .
./hydrapotion
```

### Run (development)

```bash
go run .
```

## Usage

1. Set your weight to calculate daily goal
2. Click buttons to log water intake
3. Select your mood for adjusted recommendations
4. App minimizes to system tray on close
5. Reminders popup every 30 min (configurable)

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| GUI | Fyne v2 |
| Storage | JSON files (~/.hydrapotion/) |
| Notifications | notify-send (Linux) |

## Why Fyne?

Migrated from Wails v3 to Fyne v2:
- No Node.js/npm toolchain required
- Pure Go - single binary
- Better Linux system tray support
- Faster builds (~10s vs ~60s)

## Roadmap

- [ ] Open-Meteo weather integration
- [ ] Configurable reminder intervals
- [ ] English language support
- [ ] macOS/Windows builds

## License

MIT

---

Made with 💧 by [Mori](https://github.com/MoriNo23)
