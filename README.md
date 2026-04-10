# Hydrapotion

Desktop hydration tracker with smart reminders and mood-based recommendations.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=flat&logo=react&logoColor=61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat&logo=typescript&logoColor=white)
![Wails](https://img.shields.io/badge/Wails_v3-21B9D3?style=flat)

## Stack

`Go` `Wails v3` `React` `TypeScript` `SQLite` `Open-Meteo API`

## Features

- Daily water tracking with progress visualization
- **Dynamic goal based on mood + climate**:
  - Base: 35ml/kg + height adjustment
  - Mood history bonus (last 7 days stress)
  - Current mood bonus (low: +100ml, tense: +150ml)
  - Climate bonus (temperature + humidity)
- Mood-based recommendations (Well/Neutral/Low/Tense)
- Smart reminders with snooze
- Weather integration (Open-Meteo)
- Weekly/monthly stats
- System tray integration
- Dark theme, ES/EN support

## Run

```bash
cd frontend && npm install && npm run build
cd .. && go run .
```

## Build

```bash
go build -o hydrapotion .
```

## License

MIT

---

Made with ❤️ by Mori
