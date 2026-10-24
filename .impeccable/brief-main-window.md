# Surface brief — hydrapotion main window (main.go)

## Scope and mode

- Target: main window (`CreateMainUI` in main.go) + custom widgets it renders
  (progress widget, quick-log cards, weekly chart). Settings/reminder windows
  inherit the theme, no separate brief.
- Visitor mode: **Operate** — the user logs water fast and reads their day.
- Platform note: desktop Linux, Go + Fyne v2 (PRODUCT.md "Platform: web" is a
  schema mapping; all implementation is Go/Fyne canvas drawing + widgets).

## Audience, job, task

- Audience: the owner (single user), working at his Linux desk during a workday.
- Job: log water in <2 s (150/250/500 buttons, systray, CLI) and glance at the
  day's real state; export medical-grade history.
- Task on this surface: read today's progress vs goal → tap one volume → done.
- Frequency: many glances per hour, several logs per day. Windows shows on
  demand, hides to tray.

## Constraints (pinned)

- Workflow untouched: one-tap 150/250/500, systray, CLI, reminder timers.
- Compact single fixed window (360x540), no tabs, no growing panels.
- RAM budget: GOGC=50, GOMEMLIMIT=192MiB — raster widgets must stay cheap
  (redraw only on data change, cache the image like the incumbent widgets do).
- i18n inline es/en/zh via T(key); waybar/status.json stays dead.
- Avoid list (user-declared): cartoon droplets, SaaS-gray dashboard defaults,
  hacker-terminal costume, visual noise.

## Chosen direction — "Hidrógrafo" (USGS water-report world)

Assigned candidate #3 of the grounded list; concept-seed direction round
resolved with the assigned card. Challengers `saville-catalog` and
`centre-rail` stayed competitive (full alternates); `timetable-rack`,
`video-feed`, `cassette-j-card`, `ikeda-datamatics` declined, each donating
one discipline (raised lines below).

- World: a USGS hydrological survey report sheet — printed station data,
  gauged with monospace numerals, ruled hairlines, stamped annotations.
- Palette (Restrained strategy, light ground): paper bone `#F5F1E6`, panel
  paper `#EDE8D8`, iron ink `#22313A`, hydro data-blue `#1F5F7A`, vermilion
  stamp `#B5442D`, moss met-green `#4A7C59`, amber avg `#C08A2D`, hairline
  `#D9D2BD`.
- Type voices: Fyne sans for labels; tabular monospace for every numeral.
- Memorable moment: the day drawn as a **hydrograph** — cumulative intake
  curve against a dashed goal line, weather bonus as a stamp badge.
- Light ground forced by scene: personal utility read under office/day light,
  paper-report world; dark would costume it.

### Direction contract

- **THESIS**: the window is a daily gauging-station report, not a water-tracker
  dashboard; it refuses the category's progress ring + rounded-card SaaS shell
  and refuses cartoon droplets.
- **OWN-WORLD**: bone paper grounds with hairline rules; iron-ink text;
  hydro-blue for all water data; vermilion reserved for stamps/warnings, moss
  for "goal met", amber for averages; every numeral in tabular mono; section
  headers small-caps weight-ranked over one ruled underline; controls read as
  ruled report cells, not floating rounded cards.
- **STORY**: "this is my station log" — precise, honest, medical-grade; the
  user reads their day the way a hydrologist reads a chart, logs with one tap,
  trusts the export.
- **FIRST VIEWPORT** (360x540 fixed): top report header (small-caps title,
  date as station id, lang/settings/minimize at right) over one hairline; hero
  panel = the day's hydrograph (cumulative curve vs dashed goal capacity line,
  quiet paper margins, mono reading "1240 / 2450 ml" + %, weather stamp when a
  climate bonus is active); ruled REGISTRO RÁPIDO section with the three
  volume cells (mono numerals, ink glyphs, same one-tap behavior); dense 7-day
  gauging strip (hairline grid, met-days in moss); mood selector row at the
  bottom edge.
- **FORM**: candidate 3 of 7 grounded directions (informe hidrológico USGS);
  seed: seed:1b72503cee27e3f1.
- **RAISES**: from ikeda-datamatics — mono numeric density at pixel-precise
  hairlines; from saville-catalog — courage for vast quiet paper around the
  one curve; from timetable-rack — rank by weight/case only, no size zoo; from
  video-feed — the day's status owns the viewport undivided; from
  cassette-j-card — the day as fixed-capacity vessel (goal = hard capacity
  rule, overflow shown as over-fill mark).
- **FINISH**: unreviewed and undocumented is unfinished; this build ends with
  the finish review, the verdict, DESIGN.md, and every shipping raster
  carrying its provenance.

## Unresolved decisions

- None blocking. Weather stamp shows only when a climate bonus is active
  (honest data, no invented states).
