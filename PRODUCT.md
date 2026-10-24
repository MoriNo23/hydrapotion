# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

> Nota de mapeo (2026-10-24): hydrapotion es una app de escritorio Linux (Go + Fyne v2,
> OpenGL; systray + ventana única + CLI headless). El schema de impeccable solo admite
> `web`/`ios`/`android`/`adaptive`; se registra `web` porque Fyne renderiza su propio
> lenguaje de diseño por widgets/contenedores (flujo de trabajo análogo a UI por
> componentes). Todo el trabajo visual se implementa en Go/Fyne, no en HTML/CSS.

## Users

- Usuario primario único: el dueño del proyecto — persona que trabaja frente a su PC
  Linux y quiere registrar su hidratación diaria sin fricción durante la jornada.
- Job to be done: "ayúdame a beber suficiente agua durante la jornada sin tener que
  pensar en ello, y déjame mostrarle a un médico mis datos si hace falta".
- Confirmado por el dueño (2026-10-24): uso estrictamente personal. Sin onboarding para
  terceros, sin multiusuario, sin cuentas.

## Product Purpose

Tracker de hidratación de escritorio que registra el consumo de agua, calcula una meta
dinámica (peso, estatura, ánimo, clima) y emite recordatorios adaptativos. Existe para
eliminar el esfuerzo mental de hidratarse bien durante la jornada de trabajo y para
dejar un historial presentable ante un médico. Éxito: el usuario registra agua casi sin
pensar, bebe más cerca de su meta, y puede exportar su historial real.

## Positioning

Recordatorios que se adaptan de verdad a la hora del día y al clima (Open-Meteo), meta
calculada con el contexto personal (peso/estatura/ ánimo/clima) y export médico
(CSV/JSON) — todo en un binario único de Go (~26 MB) sin Electron ni runtimes externos.
La RAM es criterio de diseño: los recordatorios no exigen una GUI residente (systemd
user timer, proceso de <1 s cada 15 min); la GUI solo se abre para visualizar.

## Operating Context

- Jornada laboral frente al PC Linux; notificaciones vía libnotify (`notify-send`),
  sonidos ASMR vía PulseAudio (`paplay`).
- Datos 100% locales en `~/.hydrapotion/` (settings, history, state, weather cache,
  sonidos). Privacidad por diseño; nada sale de la máquina salvo la consulta de clima.
- Idioma: detectar el locale del sistema al primer arranque (es/en/zh ya existen como
  traducciones inline), con override manual en settings — confirmado por el dueño.
- Distribución: GitHub Releases (binario precompilado por CI con `-ldflags="-s -w"`);
  sin Flathub/AUR por ahora — confirmado por el dueño.

## Capabilities and Constraints

- Quick logging: 150/250/500 ml (botones de GUI + systray + CLI `hydrapotion add <ml>`).
- Meta dinámica: base 35 ml/kg + bonus estatura (>150 cm ×5) + bonus ánimo (Tenso +400,
  Bajo +200) + bonus clima (≥25° +200, ≥30° +400, ≥35° +600, humedad >60% +100).
- Recordatorios: intervalo base configurable (default 1800 s) modulado por hora (×0.8
  mañanas, ×0.7 tarde, ×1.0 noche temprana, ×2.0 madrugada) y clima (×0.5–0.9); piso
  30 min, techo 6 h; `last_notified` compartido entre GUI y headless vía `state.json`.
- Weekly chart con línea de meta y promedio; ánimo por día (Bien/Neutral/Bajo/Tenso).
- Export CSV (`Fecha,ml,Estado de animo`) y JSON; CLI headless `add/status/--remind`.
- systemd: `hydrapotion-reminder.timer` (15 min, Persistent) + `hydrapotion.service`
  opcional para la GUI.
- Restricciones durables: binario único sin Node/Electron; GUI requiere libnotify y
  PulseAudio; single instance; `GOGC=50` + `GOMEMLIMIT=192MiB` (GUI) / `48MiB`
  (headless); integración con barras de estado externas eliminada a propósito — no
  reintroducir `status.json` ni consumers externos de estado.

## Brand Commitments

- Nombre: **hydrapotion**.
- Sin Electron / sin runtimes embebidos; binario Go único (binding técnico-identitario).
- Identidad visual incumbente: tema oscuro estándar de Fyne con acento azul `#40a9ff`
  (sustituible — el mundo visual nuevo se decide en `new-work`, no aquí).

## Evidence on Hand

- Código fuente completo y funcional en este repositorio; README con features y
  presupuesto de RAM; CI compila el binario en cada push.
- No hay: logo propio, capturas de pantalla, testimonios ni métricas de uso. No
  fabricar evidencia de marketing; el público es el propio dueño.

## Product Principles

1. **Cero fricción al registrar**: loguear agua toma menos de 2 segundos (botón,
   systray o CLI) — cualquier pantalla nueva no puede competir con ese camino.
2. **La GUI es un lujo, no un requisito**: todo lo esencial (recordar, loguear) debe
   funcionar sin GUI residente; la RAM es presupuesto de diseño.
3. **Datos locales, portables, médicos**: historial honesto y exportable; nada de
   cuentas ni nube.
4. **Contexto real sobre ajustes manuales**: la app responde a hora, clima y ánimo en
   vez de pedir al usuario que afine constantemente.

## Open decisions

- Dirección visual del rediseño solicitado ("arreglar el diseño de la app"): se
  establecerá en flujo `new-work` con el dueño (WORLD_DISCOVERY_REQUIRED del engine).
