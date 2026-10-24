# PRODUCT.md — Hydrapotion

> Generado vía flujo `init` de impeccable (equivalente ejecutado por agente en sandbox).
> Campos inferidos del repositorio están marcados `[inferido]` — confirmar con el dueño.

## Platform

`linux-desktop` (Go + Fyne v2, OpenGL; systray + ventanas). No es web ni móvil.
Complementado por modo headless (CLI + systemd user timer) en el mismo binario.

## Users

- Usuario primario `[inferido]`: el propio dueño del proyecto — persona que quiere
  llevar registro de su hidratación diaria en su PC Linux de trabajo.
- Job to be done `[inferido]`: "ayúdame a beber suficiente agua durante la jornada
  sin tener que pensar en ello, y déjame mostrarle a un médico mis datos si hace falta".

## Purpose & Positioning

- Tracker de hidratación de escritorio, "smart, personal, medical-grade logging" (README).
- Diferenciador `[inferido]`: recordatorios dinámicos que se adaptan a hora del día y
  clima (Open-Meteo), meta calculada con peso/estatura/ánimo/clima, y export
  médico (CSV/JSON). Sin Electron: binario único Go (~26 MB, `-s -w`).
- Arbitraje de RAM: GUI opcional; los recordatorios viven en un systemd user timer
  (~10 MB durante <1 s cada 15 min) — la GUI solo se abre para visualizar.

## Capabilities (confirmadas en código)

- Quick logging: 150/250/500 ml (botones + systray + CLI `hydrapotion add <ml>`).
- Meta dinámica: base 35 ml/kg + bonus estatura (>150 cm ×5) + bonus ánimo
  (Tenso +400, Bajo +200) + bonus clima (≥25° +200, ≥30° +400, ≥35° +600, humedad>60% +100).
- Recordatorios: intervalo base configurable (default 1800 s) modulado por hora del día
  (×0.8 mañ, ×0.7 tarde, ×1.0 noche temprana, ×2.0 madrugada) y clima (×0.5–0.9);
  piso 30 min, techo 6 h; `last_notified` compartido entre GUI y headless (state.json).
- Weekly chart con línea de meta y promedio; ánimo por día (Bien/Neutral/Bajo/Tenso).
- Export CSV (`Fecha,ml,Estado de animo`) / JSON; sonidos ASMR (paplay); tema oscuro.
- CLI headless: `add <ml>`, `status`, `--remind`, `--version`.
- systemd: `hydrapotion-reminder.timer` (15 min, Persistent) + `hydrapotion.service` opcional.

## Data & Storage

- `~/.hydrapotion/`: `settings.json` (peso, estatura, idioma, ubicación, intervalo),
  `history.json` (por día), `state.json` (last_notified, last_weather_fetch),
  `weather.json` (cache clima ≤1 h), `sounds/` (wav copiados desde ~/Descargas).

## Constraints (durables)

- Binario único, sin Node/Electron ni deps de runtime; GUI requiere libnotify (notify-send)
  y PulseAudio (paplay).
- i18n: traducciones inline es/en/zh; backend devuelve keys (`T(key)`).
- Single instance asumido en GUI; datos 100% locales (privacidad).
- Memoria: `GOGC=50` + `GOMEMLIMIT=192MiB` en GUI, `48MiB` en headless; CI compila con
  `-ldflags="-s -w"`.
- waybar eliminado (status.json ya no existe) — no reintroducir.

## Open decisions (no decidido por el dueño aún)

- Público más allá de uso personal: ¿publicar en Flathub/AUR? (afecta packaging y branding).
- ¿Mantener español por defecto o detectar del locale del sistema?
- Dirección visual para el rediseño solicitado: la identidad actual es tema oscuro
  estándar de Fyne con azul `#40a9ff` de acento `[inferido del código]`. El mundo
  visual nuevo se decide en flujo `new-work`, no aquí.
