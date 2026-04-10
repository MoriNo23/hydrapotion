# Hydrapotion

Desktop hydration tracker with smart reminders and mood-based recommendations.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne_v2-1A1A1A?style=flat)

## Stack

`Go` `Fyne v2` `SQLite` `Open-Meteo API`

## Features

- Daily water tracking with progress visualization
- **Dynamic goal based on mood + climate**:
  - Base: 35ml/kg
  - Mood bonus (low: +200ml, tense: +400ml)
  - Climate bonus (temperature + humidity)
- Mood-based recommendations (Bien/Neutral/Bajo/Tenso)
- **System tray integration** - Quick access from taskbar
- **Reminder popup** - Notifications with snooze options (5/15/30 min)
- **Weekly chart** - Visual progress for last 7 days
- Dark theme, Spanish language
- Lightweight single binary

## Screenshots

The app features:
- Main window with progress bar and quick-add buttons
- System tray menu for quick water logging
- Popup reminders with snooze functionality
- Weekly progress chart

## Run

```bash
go run .
```

## Build

```bash
go build -o hydrapotion .
```

## Tech Notes

Migrated from Wails v3 to Fyne v2 for:
- Simpler build process (no frontend toolchain needed)
- Native Go UI toolkit
- Smaller binary size
- Better system tray support on Linux

## License

MIT

---

Made with heart by Mori
