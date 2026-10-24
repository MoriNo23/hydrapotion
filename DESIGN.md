---
name: Hydrapotion
description: Tracker de hidratación de escritorio con mundo visual "Hidrógrafo" — informe hidrológico tipo USGS en Go/Fyne.
colors:
  paper: "#F5F1E6"
  panel: "#ECE7D8"
  ink: "#22313A"
  ink-muted: "#5C6A64"
  hydro: "#1F5F7A"
  hydro-wash: "#E0E8EA"
  hydro-area: "rgba(31, 95, 122, 0.15)"
  vermilion: "#A63E28"
  moss: "#4A7C59"
  amber: "#7E5A18"
  hairline: "#D9D2BD"
  rule: "#8A8271"
typography:
  reading:
    fontFamily: "Fyne monospace (tabular)"
    fontSize: "15px"
    fontWeight: 700
  label:
    fontFamily: "Fyne monospace (tabular)"
    fontSize: "10px"
    fontWeight: 700
  section:
    fontFamily: "Fyne default sans"
    fontSize: "14px"
    fontWeight: 700
  body:
    fontFamily: "Fyne default sans"
    fontSize: "14px"
    fontWeight: 400
rounded:
  cell: "0"
  panel: "0"
spacing:
  plot-margin-x: "12px"
  panel-pad: "4px"
  window: "360x580 fijo"
components:
  report-cell:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.cell}"
  report-cell-hover:
    backgroundColor: "{colors.hydro-wash}"
    textColor: "{colors.ink}"
  hydrograph-panel:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.panel}"
---

# DESIGN.md — Hydrapotion

## Overview

El sistema visual de hydrapotion es un **informe hidrológico de estación de aforo**
(mundo "Hidrógrafo", asignado por concept-seed, seed `1b72503cee27e3f1`): la ventana
trata cada día como una hoja de datos de USGS impresa en papel hueso. No es un
dashboard: no hay anillos de progreso, tarjetas redondeadas ni gotitas. Los datos de
agua van siempre en azul-hidro sobre papel; el vermellón queda reservado para sellos
y la línea de capacidad; el musgo marca los días cumplidos. La jerarquía se lleva con
peso y caja (versalitas para secciones), nunca con un zoo de tamaños. Toda cifra se
compone en monoespaciada tabular. El widget firma es el **hidrógrafo**: la curva
escalonada de ingesta acumulada del día contra la línea de capacidad (meta).

## Colors

- **Papel** (`paper #F5F1E6`): fondo de ventana. **Panel** (`panel #ECE7D8`):
  paneles de datos (hidrógrafo, resumen semanal) y reposo de controles estándar.
- **Tinta** (`ink #22313A`): texto y glifos sobre papel. **Tinta desvanecida**
  (`ink-muted #5C6A64`): secundario, etiquetas de eje; ≥4.5:1 sobre papel y panel.
- **Azul-hidro** (`hydro #1F5F7A`): TODO dato de agua (curva, barras, glifos de
  vaso, foco, selección). Lavado de área `rgba(31,95,122,0.15)` bajo la curva;
  lavado de hover `#E0E8EA` sobre celdas.
- **Vermellón** (`vermilion #A63E28`): solo sellos (CLIMA/WEATHER +N), línea de
  capacidad punteada y marca de sobrellenado. Nunca decorativo. ≥4.5:1 sobre panel.
- **Musgo** (`moss #4A7C59`): día con meta cumplida (barra del resumen semanal).
- **Ámbar** (`amber #7E5A18`): promedio (línea punteada + etiqueta). ≥4.5:1 sobre panel.
- **Filete** (`hairline #D9D2BD`) y **regla media** (`rule #8A8271`): retículas,
  subrayados de sección y bordes de control a 1 px exacto.

## Typography

Una sola familia sans (la de Fyne por defecto, workhorse) para toda la interfaz;
monoespaciada tabular para **todas** las cifras. Escala corta: lectura del
hidrógrafo 15 mono bold; cifra de celda 13 mono bold; etiquetas de eje y sellos
10 mono (bold en sellos); encabezados de sección y UI 14. Los encabezados de
sección van en MAYÚSCULAS con peso bold y un filete debajo — sin kickers ni
numeración de secciones.

## Layout

Ventana única fija de **360×580** (compacta, sin tabs, sin paneles que crezcan).
Estructura: cabecera de informe arriba (título versalitas + fecha ISO como id de
estación + controles a la derecha sobre filete), panel héroe (hidrógrafo, alto
mínimo 180), sección REGISTRO RÁPIDO con rejilla de 3 celdas, sección RESUMEN
SEMANAL (panel de 118 de alto mínimo), fila de estado de ánimo al borde inferior.
Márgenes internos de trazado: hidrógrafo 12/12/40/18; resumen 46/12/16/28.

## Elevation & Depth

Sistema plano de tinta sobre papel: la profundidad la dan los filetes y el contraste
papel/panel, no las sombras. La sombra del tema (estándar Fyne) se usa solo en
overlays y menús; sin halos de color ni sombras de offset duro.

## Shapes

Radio **0** en todas las superficies dibujadas por el informe: celdas de registro,
paneles de datos, sellos. Los controles estándar de Fyne (botones, selects, entries)
conservan su radio pequeño por defecto — se leen como controles del sistema, no como
celdas del informe. Bordes: 1 px en `rule` para controles, `hairline` para retícula.

## Components

### Cabecera de informe
- Título "HYDRAPOTION" bold + fecha `2006-01-02` mono en tinta desvanecida;
  botones de texto (código de idioma "ES/EN/ZH", "—", engranaje) alineados a la
  derecha; filete debajo.

### Hidrógrafo (widget firma)
- Panel `panel` con borde `rule` 1 px; lectura mono bold arriba a la izquierda
  (`1250 / 2750 ml · 45%`); sello de clima arriba a la derecha (texto vermellón
  mono 10 en caja con borde, solo cuando hay bonus real); retícula de filetes a
  cuartos; línea de capacidad punteada vermellón; curva escalonada `hydro` de 2 px
  con lavado de área; tramos por encima de la capacidad en vermellón; la curva
  termina en "ahora" — el papel quieto restante es deliberado.

### Celda de registro (report-cell)
- Celda 1/3 del ancho, fondo transparente, borde `rule` 1 px, radio 0; glifo de
  vaso dibujado en `hydro`; cifra mono bold 13 (`150 ml`). **Hover**: fondo
  `hydro-wash` + borde `hydro`. Un toque registra el volumen (flujo intacto).

### Resumen semanal
- Panel `panel` con borde `rule`; 7 barras planas `hydro` (musgo si cumplió meta);
  meta punteada vermellón con etiqueta mono; promedio punteado ámbar (etiqueta
  oculta si cae sobre la base); días en mono 10 tinta desvanecida.

## Do's and Don'ts

### Do:
- **Do** componer toda cifra en monoespaciada tabular; los números son datos de aforo.
- **Do** reservar vermellón para sellos/capacidad/sobrellenado y musgo para meta cumplida.
- **Do** dibujar filetes a 1 px exacto y mantener radio 0 en celdas y paneles.
- **Do** dejar papel quieto alrededor de la curva: un solo trazado por panel.

### Don't:
- **Don't** anillos de progreso, tarjetas redondeadas tipo SaaS ni layouts de dashboard.
- **Don't** gotitas de caricatura, emojis como iconos o gamificación cursi.
- **Don't** monoespaciada como disfraz "técnico" en prosa; solo datos y medidas.
- **Don't** introducir un segundo color de datos: el agua es azul-hidro, siempre.
